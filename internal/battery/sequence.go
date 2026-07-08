package battery

import (
	"context"
	"time"

	"corpos-lab/internal/model"
)

// Sequence is a named, ordered list of steps.
type Sequence struct {
	// Name is the human-readable sequence name.
	Name  string
	steps []namedStep
}

type namedStep struct {
	name string
	fn   StepFn
}

// NewSequence creates an empty sequence with the given name.
func NewSequence(name string) *Sequence {
	return &Sequence{Name: name}
}

// AddStep registers a step at the end of the sequence.
func (s *Sequence) AddStep(name string, fn StepFn) {
	s.steps = append(s.steps, namedStep{name: name, fn: fn})
}

// StepCount returns the number of registered steps.
func (s *Sequence) StepCount() int { return len(s.steps) }

// Input is what a sequence runs against: the item under evaluation, its
// content, and the inference client for verdict steps.
type Input struct {
	ItemID  string
	Content string
	Model   model.Client
}

// Result is the outcome of running a sequence. ExitIndex is the index of
// the last step that ran (0 for an empty sequence). FailureReason is empty
// on pass; on failure it carries the stopping Fail's reason or the Error's
// message.
type Result struct {
	ItemID        string       `json:"item_id"`
	StepResults   []StepResult `json:"step_results"`
	Passed        bool         `json:"passed"`
	ExitIndex     int          `json:"exit_index"`
	FailureReason string       `json:"failure_reason,omitempty"`
}

// ItemVerdicts extracts the typed verdicts from the result's step outcomes
// in step order, skipping observations and errors — the input shape
// ComposeRunVerdict folds.
func (r Result) ItemVerdicts() []Verdict {
	var verdicts []Verdict
	for _, sr := range r.StepResults {
		if sr.Outcome.Kind == OutcomeVerdict && sr.Outcome.Verdict != nil {
			verdicts = append(verdicts, *sr.Outcome.Verdict)
		}
	}
	return verdicts
}

// RunSequence executes steps in order, failing fast on the first Fail
// verdict or Error outcome. Deferred and Flag verdicts do NOT stop the
// sequence: the battery registers 15 items of which several defer pending
// corpus access, and short-circuiting on Deferred would hide downstream
// Fails. Flag is advisory — recorded, not stopping.
func RunSequence(ctx context.Context, seq *Sequence, input Input) Result {
	st := &State{
		ItemID:  input.ItemID,
		Content: input.Content,
		Model:   input.Model,
	}

	for index, step := range seq.steps {
		start := time.Now()
		outcome := step.fn(ctx, st)
		duration := time.Since(start)

		stop := outcome.Kind == OutcomeError ||
			(outcome.Kind == OutcomeVerdict && outcome.Verdict != nil && outcome.Verdict.Kind == KindFail)

		st.StepResults = append(st.StepResults, StepResult{
			Index:    index,
			StepName: step.name,
			Outcome:  outcome,
			Duration: duration,
		})

		if stop {
			reason := outcome.Message
			if outcome.Kind == OutcomeVerdict {
				reason = outcome.Verdict.Reason
			}
			return Result{
				ItemID:        input.ItemID,
				StepResults:   st.StepResults,
				Passed:        false,
				ExitIndex:     index,
				FailureReason: reason,
			}
		}
	}

	exitIndex := len(seq.steps) - 1
	if exitIndex < 0 {
		exitIndex = 0
	}
	return Result{
		ItemID:      input.ItemID,
		StepResults: st.StepResults,
		Passed:      true,
		ExitIndex:   exitIndex,
	}
}
