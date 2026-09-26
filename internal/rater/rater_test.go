package rater

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"corpos-lab/internal/model"
)

// fakeSeam is an in-memory chat seam. handler returns the reply for the nth
// call (0-indexed) and the prompt it received. It is safe for concurrent use so
// the worker-pool path can be raced.
type fakeSeam struct {
	handler func(n int, prompt string) (model.RawChatResult, error)
	mu      sync.Mutex
	prompts []string
	n       int
}

func (f *fakeSeam) RawChat(_ context.Context, prompt string, _ model.GenParams) (model.RawChatResult, error) {
	f.mu.Lock()
	n := f.n
	f.n++
	f.prompts = append(f.prompts, prompt)
	f.mu.Unlock()
	return f.handler(n, prompt)
}

func fixedClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 9, 22, 8, 30, 0, 0, time.UTC) }
}

func TestRateSliceGroundedScoresProvenanceAndUnparsed(t *testing.T) {
	replies := []model.RawChatResult{
		{Content: "C", Model: "served-x", ResponseID: "id-1", SystemFingerprint: "b9",
			Usage: model.Usage{PromptTokens: 10, CompletionTokens: 1, TotalTokens: 11}},
		{Content: "", ReasoningContent: "weighing it, finally N",
			Usage: model.Usage{PromptTokens: 20, CompletionTokens: 5, TotalTokens: 25}},
		{Content: "no verdict here",
			Usage: model.Usage{PromptTokens: 5, CompletionTokens: 2, TotalTokens: 7}},
	}
	seam := &fakeSeam{handler: func(n int, _ string) (model.RawChatResult, error) { return replies[n], nil }}
	r := NewGrounded(Config{Seam: seam, RaterID: "devstral", Endpoint: "http://x/v1", RequestedModel: "devstral-small"}, "RUBRIC")
	r.now = fixedClock()

	got, err := r.RateSlice(context.Background(), []SliceLine{
		{ID: "a", Text: "resp a"}, {ID: "b", Text: "resp b"}, {ID: "c", Text: "resp c"},
	})
	if err != nil {
		t.Fatalf("RateSlice: %v", err)
	}
	want := map[string]string{"a": "C", "b": "N", "c": "N"}
	for id, code := range want {
		if got.Scores[id] != code {
			t.Fatalf("score[%s] = %q, want %q", id, got.Scores[id], code)
		}
	}
	// b came from a reasoning fallback (parseable); only c was unparseable.
	if got.Unparsed != 1 {
		t.Fatalf("Unparsed = %d, want 1", got.Unparsed)
	}
	if got.Provenance == nil {
		t.Fatal("provenance must be captured")
	}
	p := got.Provenance
	if p.ModelReported != "served-x" || p.ResponseID != "id-1" || p.SystemFingerprint != "b9" {
		t.Fatalf("provenance from first reply wrong: %+v", p)
	}
	if p.Endpoint != "http://x/v1" || p.RequestedModel != "devstral-small" {
		t.Fatalf("provenance config fields wrong: %+v", p)
	}
	if p.RunDate != "2026-09-22T08:30:00+0000" {
		t.Fatalf("RunDate = %q", p.RunDate)
	}
	if p.Calls != 3 || p.Usage.TotalTokens != 43 || p.Usage.PromptTokens != 35 {
		t.Fatalf("usage/calls wrong: calls=%d usage=%+v", p.Calls, p.Usage)
	}
	// The grounded prompt carries the rubric and the response text.
	if len(seam.prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(seam.prompts))
	}
}

func TestRateSliceActionFallsBackToReasoning(t *testing.T) {
	seam := &fakeSeam{handler: func(_ int, _ string) (model.RawChatResult, error) {
		return model.RawChatResult{Content: "   ", ReasoningContent: "it edits v2, so A_canon"}, nil
	}}
	r := NewAction(Config{Seam: seam, RaterID: "deepseek-flash", Hosted: true})

	got, err := r.RateSlice(context.Background(), []SliceLine{{ID: "x", Text: "resp", Scenario: "api-version"}})
	if err != nil {
		t.Fatalf("RateSlice: %v", err)
	}
	if got.Scores["x"] != "A_canon" {
		t.Fatalf("verdict = %q, want A_canon (reasoning fallback)", got.Scores["x"])
	}
	if got.Unparsed != 0 {
		t.Fatalf("Unparsed = %d, want 0 (action verdict parsed)", got.Unparsed)
	}
}

func TestRateSliceActionUnparseableBecomesUnscoreable(t *testing.T) {
	seam := &fakeSeam{handler: func(_ int, _ string) (model.RawChatResult, error) {
		return model.RawChatResult{Content: "", ReasoningContent: ""}, nil
	}}
	r := NewAction(Config{Seam: seam, RaterID: "deepseek-flash", Hosted: true})
	got, err := r.RateSlice(context.Background(), []SliceLine{{ID: "x", Text: "resp", Scenario: "api-version"}})
	if err != nil {
		t.Fatalf("RateSlice: %v", err)
	}
	if got.Scores["x"] != "unscoreable" {
		t.Fatalf("verdict = %q, want unscoreable", got.Scores["x"])
	}
	if got.Unparsed != 1 {
		t.Fatalf("Unparsed = %d, want 1", got.Unparsed)
	}
}

func TestRateSliceStopsOnFirstError(t *testing.T) {
	boom := errors.New("endpoint down")
	seam := &fakeSeam{handler: func(n int, _ string) (model.RawChatResult, error) {
		if n == 1 {
			return model.RawChatResult{}, boom
		}
		return model.RawChatResult{Content: "C"}, nil
	}}
	r := NewGrounded(Config{Seam: seam, RaterID: "r"}, "RUBRIC")

	_, err := r.RateSlice(context.Background(), []SliceLine{{ID: "a"}, {ID: "b"}, {ID: "c"}})
	if err == nil {
		t.Fatal("want an error when a line fails")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("error should wrap the seam error, got %v", err)
	}
}

func TestNewRaterSeedOnlyForLocal(t *testing.T) {
	local := NewGrounded(Config{RaterID: "l", Hosted: false}, "R")
	if local.params.Seed == nil || *local.params.Seed != 1 {
		t.Fatalf("local rater must pin seed 1, got %v", local.params.Seed)
	}
	if local.params.MaxTokens == nil || *local.params.MaxTokens != 16 {
		t.Fatalf("local default budget = %v, want 16", local.params.MaxTokens)
	}
	hosted := NewAction(Config{RaterID: "h", Hosted: true})
	if hosted.params.Seed != nil {
		t.Fatal("hosted rater must send no seed")
	}
	if hosted.params.MaxTokens == nil || *hosted.params.MaxTokens != 2048 {
		t.Fatalf("hosted default budget = %v, want 2048", hosted.params.MaxTokens)
	}
	explicit := 64
	custom := NewGrounded(Config{RaterID: "c", Hosted: true, MaxTokens: &explicit}, "R")
	if *custom.params.MaxTokens != 64 {
		t.Fatalf("explicit budget must win, got %v", custom.params.MaxTokens)
	}
}
