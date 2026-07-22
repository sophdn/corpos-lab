package battery

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"corpos-lab/internal/model"
)

// Characterization tests ported from lab-app-server sequences/mod.rs
// (8 tests); the Rust MockModel/NullTranslator pair becomes a fakeClient
// func type (mock-idiom translation per code-migration-discipline).

type fakeClient struct {
	text string
	err  error
	name string
	// prompts records every prompt received, for assertion.
	prompts []string
}

func (f *fakeClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	f.prompts = append(f.prompts, prompt)
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{Text: f.text}, nil
}

func (f *fakeClient) Name() string {
	if f.name != "" {
		return f.name
	}
	return "fake-model"
}

func (f *fakeClient) Version() string { return "0.0.0" }

func (f *fakeClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

// fakeProfiles is the in-memory ProfileReader for battery tests: docs maps a
// resolved path to its content; err, when set, is returned for every read
// (the unreadable-referent case). seen records the paths asked for, so a test
// can assert the step resolved the reference relative to the entry file.
type fakeProfiles struct {
	docs map[string]string
	err  error
	seen []string
}

func (f *fakeProfiles) ReadProfile(path string) (string, error) {
	f.seen = append(f.seen, path)
	if f.err != nil {
		return "", f.err
	}
	doc, ok := f.docs[path]
	if !ok {
		return "", errors.New("no such profile: " + path)
	}
	return doc, nil
}

func testInput() Input {
	return Input{
		ItemID:  "test-item",
		Content: "test content",
		Model:   &fakeClient{text: "mock"},
	}
}

func passStep() StepFn {
	return func(_ context.Context, _ *State) StepOutcome { return PassOutcome() }
}

func failStep(reason string) StepFn {
	return func(_ context.Context, _ *State) StepOutcome { return FailOutcome(reason) }
}

func errorStep(msg string) StepFn {
	return func(_ context.Context, _ *State) StepOutcome { return ErrorOutcome(msg) }
}

func observationStep(data string) StepFn {
	return func(_ context.Context, _ *State) StepOutcome {
		return ObservationOutcome(json.RawMessage(data))
	}
}

func TestAllPassSequence(t *testing.T) {
	seq := NewSequence("test-all-pass")
	seq.AddStep("step-1", passStep())
	seq.AddStep("step-2", passStep())
	seq.AddStep("step-3", passStep())

	result := RunSequence(context.Background(), seq, testInput())

	if !result.Passed {
		t.Fatal("expected pass")
	}
	if len(result.StepResults) != 3 {
		t.Fatalf("expected 3 step results, got %d", len(result.StepResults))
	}
	if result.FailureReason != "" {
		t.Fatalf("expected no failure reason, got %q", result.FailureReason)
	}
	if result.ExitIndex != 2 {
		t.Fatalf("expected exit index 2, got %d", result.ExitIndex)
	}
}

func TestVerdictFailExitsEarly(t *testing.T) {
	seq := NewSequence("test-fail")
	seq.AddStep("step-1", passStep())
	seq.AddStep("step-2", failStep("bad item"))
	seq.AddStep("step-3", passStep()) // should not run

	result := RunSequence(context.Background(), seq, testInput())

	if result.Passed {
		t.Fatal("expected failure")
	}
	if len(result.StepResults) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.StepResults))
	}
	if result.ExitIndex != 1 {
		t.Fatalf("expected exit index 1, got %d", result.ExitIndex)
	}
	if result.FailureReason != "bad item" {
		t.Fatalf("expected failure reason %q, got %q", "bad item", result.FailureReason)
	}
}

// ContinueOnFail: every step runs, the verdict is still a failure, and the
// FIRST failure is the reason — the later items get measured instead of being
// left unassessed. The recertification case this exists for fails at item 9
// and still needs item 15 measured.
func TestContinueOnFailRunsEveryStep(t *testing.T) {
	seq := NewSequence("test-continue")
	seq.AddStep("step-1", passStep())
	seq.AddStep("step-2", failStep("first bad item"))
	seq.AddStep("step-3", failStep("second bad item"))
	seq.AddStep("step-4", passStep())

	in := testInput()
	in.ContinueOnFail = true
	result := RunSequence(context.Background(), seq, in)

	if result.Passed {
		t.Fatal("a failing item must still fail the run under ContinueOnFail")
	}
	if len(result.StepResults) != 4 {
		t.Fatalf("expected all 4 steps to run, got %d", len(result.StepResults))
	}
	if result.ExitIndex != 3 {
		t.Fatalf("expected exit index 3 (the last step that ran), got %d", result.ExitIndex)
	}
	if result.FailureReason != "first bad item" {
		t.Fatalf("expected the FIRST failure as the reason, got %q", result.FailureReason)
	}
}

// ContinueOnFail suspends the Fail half of the stop rule only. An Error is a
// broken transport, not a verdict, so it still stops the sequence — repeating
// it down the remaining items would record noise, not measurements.
func TestContinueOnFailStillStopsOnError(t *testing.T) {
	seq := NewSequence("test-continue-error")
	seq.AddStep("step-1", errorStep("transport down"))
	seq.AddStep("step-2", passStep()) // should not run

	in := testInput()
	in.ContinueOnFail = true
	result := RunSequence(context.Background(), seq, in)

	if result.Passed {
		t.Fatal("expected failure")
	}
	if len(result.StepResults) != 1 {
		t.Fatalf("expected 1 step result, got %d", len(result.StepResults))
	}
	if result.FailureReason != "transport down" {
		t.Fatalf("got %q", result.FailureReason)
	}
}

// A clean run under ContinueOnFail is indistinguishable from a clean fail-fast
// run: nothing failed, so there is nothing to carry.
func TestContinueOnFailPassesWhenNothingFails(t *testing.T) {
	seq := NewSequence("test-continue-clean")
	seq.AddStep("step-1", passStep())
	seq.AddStep("step-2", passStep())

	in := testInput()
	in.ContinueOnFail = true
	result := RunSequence(context.Background(), seq, in)

	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
	if result.FailureReason != "" {
		t.Fatalf("expected no failure reason, got %q", result.FailureReason)
	}
}

func TestObservationThenVerdictReadsData(t *testing.T) {
	seq := NewSequence("test-obs-verdict")
	seq.AddStep("observe", observationStep(`{"score": 42}`))
	seq.AddStep("judge", func(_ context.Context, st *State) StepOutcome {
		prior := st.GetStepResult("observe")
		if prior == nil || prior.Outcome.Kind != OutcomeObservation {
			return ErrorOutcome("missing observation")
		}
		var payload struct {
			Score int `json:"score"`
		}
		if err := json.Unmarshal(prior.Outcome.Data, &payload); err != nil {
			return ErrorOutcome("bad observation payload: " + err.Error())
		}
		if payload.Score > 40 {
			return PassOutcome()
		}
		return FailOutcome("score below threshold")
	})

	result := RunSequence(context.Background(), seq, testInput())

	if !result.Passed {
		t.Fatalf("expected pass, got failure: %s", result.FailureReason)
	}
	if len(result.StepResults) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.StepResults))
	}
}

func TestErrorExitsEarly(t *testing.T) {
	seq := NewSequence("test-error")
	seq.AddStep("step-1", passStep())
	seq.AddStep("step-2", errorStep("connection lost"))
	seq.AddStep("step-3", passStep()) // should not run

	result := RunSequence(context.Background(), seq, testInput())

	if result.Passed {
		t.Fatal("expected failure")
	}
	if len(result.StepResults) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.StepResults))
	}
	if result.ExitIndex != 1 {
		t.Fatalf("expected exit index 1, got %d", result.ExitIndex)
	}
	if result.FailureReason != "connection lost" {
		t.Fatalf("expected failure reason %q, got %q", "connection lost", result.FailureReason)
	}
}

func TestEmptySequencePasses(t *testing.T) {
	seq := NewSequence("test-empty")

	result := RunSequence(context.Background(), seq, testInput())

	if !result.Passed {
		t.Fatal("expected pass")
	}
	if len(result.StepResults) != 0 {
		t.Fatalf("expected no step results, got %d", len(result.StepResults))
	}
	if result.FailureReason != "" {
		t.Fatalf("expected no failure reason, got %q", result.FailureReason)
	}
	if result.ExitIndex != 0 {
		t.Fatalf("empty sequence exit index should be 0 (Rust saturating_sub parity), got %d", result.ExitIndex)
	}
}

func TestDeferredAndFlagDoNotStopSequence(t *testing.T) {
	// The load-bearing fail-fast nuance: Deferred and Flag continue, so a
	// deferred item cannot hide a downstream Fail.
	seq := NewSequence("test-deferred-continues")
	seq.AddStep("deferred", func(_ context.Context, _ *State) StepOutcome {
		return VerdictOutcome(Deferred("corpus access"))
	})
	seq.AddStep("flagged", func(_ context.Context, _ *State) StepOutcome {
		return VerdictOutcome(Flag("advisory"))
	})
	seq.AddStep("hard-fail", failStep("downstream failure"))

	result := RunSequence(context.Background(), seq, testInput())

	if result.Passed {
		t.Fatal("expected the downstream fail to be reached")
	}
	if result.ExitIndex != 2 {
		t.Fatalf("expected exit at step 2 (deferred+flag continued), got %d", result.ExitIndex)
	}
	if result.FailureReason != "downstream failure" {
		t.Fatalf("got %q", result.FailureReason)
	}
}

func TestGetStepResultReturnsNilForMissingName(t *testing.T) {
	seq := NewSequence("test-miss")
	seq.AddStep("judge", func(_ context.Context, st *State) StepOutcome {
		if st.GetStepResult("nonexistent") == nil {
			return PassOutcome()
		}
		return FailOutcome("nonexistent step unexpectedly present")
	})

	result := RunSequence(context.Background(), seq, testInput())

	if !result.Passed {
		t.Fatalf("expected pass, got: %s", result.FailureReason)
	}
}

func TestStepCountReflectsAddedSteps(t *testing.T) {
	seq := NewSequence("test-count")
	if seq.StepCount() != 0 {
		t.Fatalf("expected 0, got %d", seq.StepCount())
	}
	seq.AddStep("a", passStep())
	if seq.StepCount() != 1 {
		t.Fatalf("expected 1, got %d", seq.StepCount())
	}
	seq.AddStep("b", passStep())
	seq.AddStep("c", passStep())
	if seq.StepCount() != 3 {
		t.Fatalf("expected 3, got %d", seq.StepCount())
	}
}

func TestStateStartsWithEmptyResults(t *testing.T) {
	input := testInput()
	st := &State{ItemID: input.ItemID, Content: input.Content, Model: input.Model}

	if st.ItemID != "test-item" || st.Content != "test content" {
		t.Fatalf("state fields not carried: %+v", st)
	}
	if len(st.StepResults) != 0 {
		t.Fatal("expected empty step results")
	}
}

func TestOutcomeConstructors(t *testing.T) {
	if o := FailItemOutcome(2, "r"); o.Verdict == nil || o.Verdict.Item == nil || *o.Verdict.Item != 2 {
		t.Fatalf("FailItemOutcome lost item: %+v", o)
	}
	if o := ErrorOutcome("m"); o.Kind != OutcomeError || o.Message != "m" {
		t.Fatalf("ErrorOutcome shape: %+v", o)
	}
	if o := ObservationOutcome(json.RawMessage(`{}`)); o.Kind != OutcomeObservation {
		t.Fatalf("ObservationOutcome shape: %+v", o)
	}
}

func TestFakeClientErrorPath(t *testing.T) {
	// Exercises the fake used by inference-step tests: an erroring client
	// must surface its error.
	f := &fakeClient{err: errors.New("boom")}
	_, err := f.Generate(context.Background(), "p", model.GenParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	if f.Name() != "fake-model" || f.Version() != "0.0.0" {
		t.Fatal("fake identity defaults")
	}
	if n := (&fakeClient{name: "custom"}).Name(); n != "custom" {
		t.Fatalf("got %q", n)
	}
	if model.Float64(0.5) == nil || model.Int(3) == nil {
		t.Fatal("param helpers")
	}
}
