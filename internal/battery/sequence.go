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
// content, the inference client for verdict steps, and — for Item 15 — the
// entry's filesystem path and the reader that resolves its fallout-profile
// reference. EntryPath and Profiles may be zero for sequences that never run
// Item 15; a run that does reach Item 15 without them fails that item closed.
type Input struct {
	ItemID    string
	Content   string
	Model     model.Client
	EntryPath string
	Profiles  ProfileReader
	// ContinueOnFail runs every remaining step after a Fail verdict instead of
	// stopping at it. The run verdict is unchanged — the sequence still does not
	// pass, and FailureReason still carries the FIRST failure — but the later
	// items get measured instead of being left silently unassessed. Fail-fast is
	// a cost optimization, not a claim that the downstream items are unknowable:
	// a recertification run whose subject is item 15 learns nothing about it when
	// an item-9 failure ends the sequence seven steps early. An Error outcome
	// still stops the sequence regardless — a broken transport is not a verdict,
	// and repeating it down the remaining items records noise, not measurements.
	ContinueOnFail bool
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
//
// Input.ContinueOnFail suspends the Fail half of that rule: every step runs,
// the first Fail is still the Result's FailureReason, and Passed is still
// false. ExitIndex keeps its documented meaning — the index of the last step
// that ran — so under ContinueOnFail it is the final step, and the first
// failure is located by scanning StepResults for the first Fail.
func RunSequence(ctx context.Context, seq *Sequence, input Input) Result {
	st := &State{
		ItemID:    input.ItemID,
		Content:   input.Content,
		Model:     input.Model,
		EntryPath: input.EntryPath,
		Profiles:  input.Profiles,
	}

	failed := false
	firstFailure := ""

	for index, step := range seq.steps {
		start := time.Now()
		outcome := step.fn(ctx, st)
		duration := time.Since(start)

		isFail := outcome.Kind == OutcomeVerdict && outcome.Verdict != nil && outcome.Verdict.Kind == KindFail
		isError := outcome.Kind == OutcomeError

		st.StepResults = append(st.StepResults, StepResult{
			Index:    index,
			StepName: step.name,
			Outcome:  outcome,
			Duration: duration,
		})

		if (isFail || isError) && !failed {
			failed = true
			if isError {
				firstFailure = outcome.Message
			} else {
				firstFailure = outcome.Verdict.Reason
			}
		}

		if isError || (isFail && !input.ContinueOnFail) {
			return Result{
				ItemID:        input.ItemID,
				StepResults:   st.StepResults,
				Passed:        false,
				ExitIndex:     index,
				FailureReason: firstFailure,
			}
		}
	}

	exitIndex := len(seq.steps) - 1
	if exitIndex < 0 {
		exitIndex = 0
	}
	return Result{
		ItemID:        input.ItemID,
		StepResults:   st.StepResults,
		Passed:        !failed,
		ExitIndex:     exitIndex,
		FailureReason: firstFailure,
	}
}
