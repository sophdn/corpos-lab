package battery

import (
	"context"
	"strings"
	"testing"
)

// --- Item 12: empirical default-alignment against the unguided baseline (bug 1332) ---

// firedShelf is a two-model shelf where the class's target behavior fired
// unguided on BOTH models — the "trained default across the shelf" shape.
var firedShelf = BaselineOutcome{Models: []BaselineModelResult{
	{ModelID: "Mistral-7B-Instruct-v0.3", Version: "v0.3-q4km", Fired: true, Runs: 8},
	{ModelID: "phi-4", Version: "phi4-q4km", Fired: true, Runs: 8},
}}

// mixedShelf is a two-model shelf where the behavior fired on one model but was
// MISSED on the other — so the glyph does real work on at least part of the
// population.
var mixedShelf = BaselineOutcome{Models: []BaselineModelResult{
	{ModelID: "Mistral-7B-Instruct-v0.3", Version: "v0.3-q4km", Fired: true, Runs: 8},
	{ModelID: "phi-4", Version: "phi4-q4km", Fired: false, Runs: 8},
}}

func TestItem12DefersWhenNoBaseline(t *testing.T) {
	for _, b := range []BaselineOutcome{{}, {Models: []BaselineModelResult{}}} {
		out := Item12DefaultAlignment(context.Background(), &State{Content: "entry", Baseline: b})
		if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindDeferred {
			t.Fatalf("no baseline must defer (degrade-never-fail), got %+v", out)
		}
		if !strings.Contains(out.Verdict.Pending, "no unguided baseline supplied") {
			t.Fatalf("defer pending should name the missing baseline, got %q", out.Verdict.Pending)
		}
	}
}

func TestItem12FailsWhenFiresAcrossShelf(t *testing.T) {
	out := Item12DefaultAlignment(context.Background(), &State{Content: "entry", Baseline: firedShelf})
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("a behavior that fires unguided across the shelf is a trained default → FAIL, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 12 {
		t.Fatalf("fail must tag item 12, got %v", out.Verdict.Item)
	}
	for _, want := range []string{"fired UNGUIDED", "trained default", "no friction = fail", "Mistral-7B-Instruct-v0.3", "phi-4"} {
		if !strings.Contains(out.Verdict.Reason, want) {
			t.Fatalf("fail reason must carry %q (population-relative + honest), got %q", want, out.Verdict.Reason)
		}
	}
}

func TestItem12PassesWhenMissedOnAnyShelfModel(t *testing.T) {
	out := Item12DefaultAlignment(context.Background(), &State{Content: "entry", Baseline: mixedShelf})
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPassWithCondition {
		t.Fatalf("a behavior missed unguided on any model → the glyph does work → PASS, got %+v", out)
	}
	// The pass is honest about the population it is relative to.
	for _, want := range []string{"MISSED unguided", "measured population", "phi-4 missed unguided"} {
		if !strings.Contains(out.Verdict.Condition, want) {
			t.Fatalf("pass condition must carry %q, got %q", want, out.Verdict.Condition)
		}
	}
}

func TestBaselineFiresAcrossShelfUnanimity(t *testing.T) {
	cases := []struct {
		name string
		b    BaselineOutcome
		want bool
	}{
		{"all-fired", firedShelf, true},
		{"one-missed", mixedShelf, false},
		{"empty", BaselineOutcome{}, false},
		{"single-fired", BaselineOutcome{Models: []BaselineModelResult{{ModelID: "m", Fired: true, Runs: 8}}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.b.firesAcrossShelf(); got != c.want {
				t.Fatalf("firesAcrossShelf() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestBaselineDescribeNamesEveryModel(t *testing.T) {
	// An unnamed model still describes without crashing (record what ran, even
	// when a model id is missing).
	b := BaselineOutcome{Models: []BaselineModelResult{
		{ModelID: "", Fired: true, Runs: 4},
		{ModelID: "phi-4", Fired: false, Runs: 8},
	}}
	got := b.describe()
	for _, want := range []string{"unnamed-model fired unguided (4 runs)", "phi-4 missed unguided (8 runs)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("describe() = %q, want substring %q", got, want)
		}
	}
}

func TestBaselineSuppliedReflectsModelCount(t *testing.T) {
	if (BaselineOutcome{}).supplied() {
		t.Fatal("zero-value outcome is not supplied")
	}
	if !firedShelf.supplied() {
		t.Fatal("a two-model outcome is supplied")
	}
}
