package battery

import (
	"context"
	"encoding/json"
	"time"

	"corpos-lab/internal/model"
)

// StepOutcomeKind discriminates StepOutcome variants.
type StepOutcomeKind string

// Step outcome kinds.
const (
	OutcomeObservation StepOutcomeKind = "observation"
	OutcomeVerdict     StepOutcomeKind = "verdict"
	OutcomeError       StepOutcomeKind = "error"
)

// StepOutcome is the result of a single step execution: an observation
// (data collected, no judgement), a typed verdict, or an error that
// prevented completion.
type StepOutcome struct {
	Kind StepOutcomeKind `json:"kind"`
	// Data carries an observation's payload.
	Data json.RawMessage `json:"data,omitempty"`
	// Verdict carries a verdict outcome's typed verdict.
	Verdict *Verdict `json:"verdict,omitempty"`
	// Message carries an error outcome's description.
	Message string `json:"message,omitempty"`
}

// PassOutcome returns a passing verdict outcome.
func PassOutcome() StepOutcome {
	v := Pass()
	return StepOutcome{Kind: OutcomeVerdict, Verdict: &v}
}

// FailOutcome returns a failing verdict outcome with no battery item number.
func FailOutcome(reason string) StepOutcome {
	v := Fail(reason)
	return StepOutcome{Kind: OutcomeVerdict, Verdict: &v}
}

// FailItemOutcome returns a failing verdict outcome at a specific battery item.
func FailItemOutcome(item int, reason string) StepOutcome {
	v := FailItem(item, reason)
	return StepOutcome{Kind: OutcomeVerdict, Verdict: &v}
}

// VerdictOutcome wraps a typed verdict as a step outcome.
func VerdictOutcome(v Verdict) StepOutcome {
	return StepOutcome{Kind: OutcomeVerdict, Verdict: &v}
}

// ErrorOutcome returns an error outcome.
func ErrorOutcome(message string) StepOutcome {
	return StepOutcome{Kind: OutcomeError, Message: message}
}

// ObservationOutcome returns an observation outcome carrying data.
func ObservationOutcome(data json.RawMessage) StepOutcome {
	return StepOutcome{Kind: OutcomeObservation, Data: data}
}

// StepResult is one executed step: its position, name, outcome, and
// wall-clock duration.
type StepResult struct {
	Index    int           `json:"index"`
	StepName string        `json:"step_name"`
	Outcome  StepOutcome   `json:"outcome"`
	Duration time.Duration `json:"duration_ns"`
}

// ProfileReader is the sans-IO seam for Item 15: it reads a referenced
// fallout-profile document by path. The step resolves the path the entry
// points at (pure string math) and hands it here; the injected reader is the
// only thing that touches the filesystem, so the battery package stays sans-IO
// and tests substitute an in-memory fake. The path is already resolved
// relative to the entry file — the reader opens it as given.
type ProfileReader interface {
	ReadProfile(path string) (string, error)
}

// State is the sequence execution context steps receive: the item under
// evaluation, its content, the inference client, and the results of every
// step run so far (observation → judge patterns read prior results).
//
// EntryPath is the filesystem path of the entry under assessment; Item 15
// resolves the entry's "**Fallout profile:**" reference relative to it.
// Profiles is the injected reader Item 15 uses to open that referent — nil
// unless the caller wired one, in which case Item 15 fails closed rather than
// manufacturing a pass it cannot verify.
type State struct {
	ItemID      string
	Content     string
	Model       model.Client
	EntryPath   string
	Profiles    ProfileReader
	StepResults []StepResult
}

// GetStepResult returns the result of a prior step by name, or nil if no
// step with that name has run.
func (s *State) GetStepResult(name string) *StepResult {
	for i := range s.StepResults {
		if s.StepResults[i].StepName == name {
			return &s.StepResults[i]
		}
	}
	return nil
}

// StepFn is a single sequence step: it receives the cancellation context
// and the mutable sequence state, and produces an outcome.
type StepFn func(ctx context.Context, st *State) StepOutcome
