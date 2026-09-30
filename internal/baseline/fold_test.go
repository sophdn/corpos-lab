package baseline

import "testing"

// capWith builds a capture for model id with runs numbered 1..n and slice ids
// item|id|runK, so a test can key rater codes by the same ids Fold reads.
func capWith(item, id, version string, n int) Capture {
	c := Capture{Item: item, ModelID: id, Version: version}
	for k := 1; k <= n; k++ {
		c.Runs = append(c.Runs, CaptureRun{Run: k, SliceID: SliceID(item, id, id, k)})
	}
	return c
}

func TestFoldMajorityCFires(t *testing.T) {
	c := capWith("pwv", "Mistral-7B", "v0.3", 4)
	// Three of four resolved to C -> a strict majority -> fired.
	scores := map[string]string{
		"pwv|Mistral-7B|run1": "C",
		"pwv|Mistral-7B|run2": "C",
		"pwv|Mistral-7B|run3": "C",
		"pwv|Mistral-7B|run4": "I",
	}
	out, gaps := Fold([]Capture{c}, []map[string]string{scores})
	if len(gaps) != 0 {
		t.Fatalf("gaps = %v, want none", gaps)
	}
	if len(out.Models) != 1 {
		t.Fatalf("models = %d, want 1", len(out.Models))
	}
	m := out.Models[0]
	if !m.Fired || m.Runs != 4 || m.ModelID != "Mistral-7B" || m.Version != "v0.3" {
		t.Errorf("model = %+v, want fired, 4 runs, Mistral-7B/v0.3", m)
	}
}

func TestFoldTieDoesNotFire(t *testing.T) {
	c := capWith("pwv", "m", "", 4)
	scores := map[string]string{
		"pwv|m|run1": "C", "pwv|m|run2": "C",
		"pwv|m|run3": "I", "pwv|m|run4": "I",
	}
	out, _ := Fold([]Capture{c}, []map[string]string{scores})
	if out.Models[0].Fired {
		t.Error("a 2-2 tie must not fire (strict majority)")
	}
}

func TestFoldTwoRatersMustAgree(t *testing.T) {
	c := capWith("pwv", "m", "", 2)
	r1 := map[string]string{"pwv|m|run1": "C", "pwv|m|run2": "C"}
	r2 := map[string]string{"pwv|m|run1": "C", "pwv|m|run2": "I"} // disagree on run2
	out, _ := Fold([]Capture{c}, []map[string]string{r1, r2})
	if len(out.Models) != 1 {
		t.Fatalf("models = %d, want 1", len(out.Models))
	}
	// run1 resolves to C (agreed); run2 is unresolved. One resolved run, it is C.
	if out.Models[0].Runs != 1 || !out.Models[0].Fired {
		t.Errorf("model = %+v, want 1 resolved run that fired", out.Models[0])
	}
}

func TestFoldAllUnresolvedIsAGap(t *testing.T) {
	c := capWith("pwv", "Qwen2.5-32B", "", 2)
	r1 := map[string]string{"pwv|Qwen2.5-32B|run1": "C", "pwv|Qwen2.5-32B|run2": "C"}
	r2 := map[string]string{"pwv|Qwen2.5-32B|run1": "I", "pwv|Qwen2.5-32B|run2": "I"}
	out, gaps := Fold([]Capture{c}, []map[string]string{r1, r2})
	if len(out.Models) != 0 {
		t.Errorf("models = %v, want none (all runs unresolved)", out.Models)
	}
	if len(gaps) != 1 || gaps[0].ModelID != "Qwen2.5-32B" {
		t.Fatalf("gaps = %+v, want one for Qwen2.5-32B", gaps)
	}
}

func TestFoldMissingCodeIsUnresolved(t *testing.T) {
	c := capWith("pwv", "m", "", 2)
	// run2 has no code at all -> unresolved; run1 is C -> majority of the 1 resolved.
	scores := map[string]string{"pwv|m|run1": "C"}
	out, gaps := Fold([]Capture{c}, []map[string]string{scores})
	if len(gaps) != 0 || len(out.Models) != 1 || out.Models[0].Runs != 1 || !out.Models[0].Fired {
		t.Errorf("out=%+v gaps=%+v, want one model, 1 resolved run, fired", out.Models, gaps)
	}
}

func TestFoldBlankCodeIsUnresolved(t *testing.T) {
	c := capWith("pwv", "m", "", 1)
	scores := map[string]string{"pwv|m|run1": ""}
	out, gaps := Fold([]Capture{c}, []map[string]string{scores})
	if len(out.Models) != 0 || len(gaps) != 1 {
		t.Errorf("a blank code must be unresolved: out=%+v gaps=%+v", out.Models, gaps)
	}
}

func TestFoldNoRatersIsAllGaps(t *testing.T) {
	c := capWith("pwv", "m", "", 2)
	out, gaps := Fold([]Capture{c}, nil)
	if len(out.Models) != 0 || len(gaps) != 1 {
		t.Errorf("no raters -> no resolved runs: out=%+v gaps=%+v", out.Models, gaps)
	}
}

func TestFoldEmptyModelIDUsesRequested(t *testing.T) {
	c := Capture{Item: "pwv", ModelID: "", Requested: "role-primary"}
	c.Runs = []CaptureRun{{Run: 1, SliceID: "pwv|role-primary|run1"}}
	scores := map[string]string{"pwv|role-primary|run1": "I"}
	out, _ := Fold([]Capture{c}, []map[string]string{scores})
	if len(out.Models) != 1 || out.Models[0].ModelID != "role-primary" {
		t.Fatalf("model id fallback failed: %+v", out.Models)
	}
	if out.Models[0].Fired {
		t.Error("single I must not fire")
	}
}

func TestFoldShelfWideFireVersusMiss(t *testing.T) {
	// Two models both fire -> BaselineOutcome an item-12 read would FAIL on.
	a := capWith("pwv", "Mistral-7B", "", 1)
	b := capWith("pwv", "Qwen2.5-32B", "", 1)
	scores := map[string]string{
		"pwv|Mistral-7B|run1":  "C",
		"pwv|Qwen2.5-32B|run1": "C",
	}
	out, _ := Fold([]Capture{a, b}, []map[string]string{scores})
	if len(out.Models) != 2 || !out.Models[0].Fired || !out.Models[1].Fired {
		t.Fatalf("want both models fired, got %+v", out.Models)
	}
}
