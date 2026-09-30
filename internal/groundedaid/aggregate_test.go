package groundedaid

import (
	"math"
	"strings"
	"testing"
)

func testKey() map[string]KeyEntry {
	return map[string]KeyEntry{
		"i1": {Cls: "post-write-verification-absent", Scenario: "1", Model: "mistral", Condition: "baseline"},
		"i2": {Cls: "post-write-verification-absent", Scenario: "1", Model: "mistral", Condition: "ground_only"},
	}
}

func TestAggregateDualRater(t *testing.T) {
	key := testKey()
	a := map[string]string{"i1": "C", "i2": "Ic"}
	b := map[string]string{"i1": "C", "i2": "C"}
	out := Aggregate(key, a, b)

	if !strings.HasPrefix(out, "rater-a: 2 ids | rater-b: 2 ids\n") {
		t.Errorf("header line wrong:\n%s", out)
	}
	if !strings.Contains(out, "agreement (exact code): 1/2 = 0.500\n") {
		t.Errorf("exact-code agreement wrong:\n%s", out)
	}
	if !strings.Contains(out, "agreement (C vs not-C): 1/2 = 0.500\n") {
		t.Errorf("C-vs-notC agreement wrong:\n%s", out)
	}
	if !strings.Contains(out, "[strict-consensus]") {
		t.Errorf("expected strict-consensus tag:\n%s", out)
	}
	// pooled: baseline is a consensus C (1/1=1.00), ground_only is not (0/1=0.00).
	if !strings.Contains(out, "  1/1  =1.00") {
		t.Errorf("pooled baseline cell wrong:\n%s", out)
	}
	if !strings.Contains(out, "  0/1  =0.00") {
		t.Errorf("pooled ground_only cell wrong:\n%s", out)
	}
	// Empty cells render as a NaN rate.
	if !strings.Contains(out, "  0/0  = nan") {
		t.Errorf("empty cell should show ' nan':\n%s", out)
	}
	// Per-cell line for the populated (class, model, s1).
	if !strings.Contains(out, "  post-write-verification- mistral  s1  basel= 1/1  groun= 0/0  groun= 0/1  domai= 0/0") {
		t.Errorf("per-cell line wrong:\n%s", out)
	}
}

func TestAggregateSingleRater(t *testing.T) {
	key := testKey()
	a := map[string]string{"i1": "C"} // i2 absent -> code_for "?"
	out := Aggregate(key, a, nil)

	if !strings.HasPrefix(out, "rater-a: 1 ids\n") {
		t.Errorf("single-rater header wrong:\n%s", out)
	}
	if strings.Contains(out, "rater-b") || strings.Contains(out, "agreement (") {
		t.Errorf("single rater should print no rater-b / agreement lines:\n%s", out)
	}
	if !strings.Contains(out, "[rater-a]") {
		t.Errorf("expected rater-a tag:\n%s", out)
	}
	// baseline has one "C" -> 1/1; ground_only has "?" -> 0/1.
	if !strings.Contains(out, "  1/1  =1.00") || !strings.Contains(out, "  0/1  =0.00") {
		t.Errorf("single-rater pooled cells wrong:\n%s", out)
	}
}

func TestCrateAndHelpers(t *testing.T) {
	c, n, r := crate([]string{"C", "x", "C"})
	if c != 2 || n != 3 || math.Abs(r-2.0/3.0) > 1e-9 {
		t.Errorf("crate = (%d,%d,%v)", c, n, r)
	}
	if _, _, r := crate(nil); !math.IsNaN(r) {
		t.Errorf("crate(empty) rate = %v, want NaN", r)
	}
	if !math.IsNaN(ratio(1, 0)) {
		t.Errorf("ratio(_,0) should be NaN")
	}
	if got := f4dot2(math.NaN()); got != " nan" {
		t.Errorf("f4dot2(NaN) = %q", got)
	}
	if got := ljust("ab", 4); got != "ab  " {
		t.Errorf("ljust = %q", got)
	}
	if got := trunc("abcdef", 3); got != "abc" {
		t.Errorf("trunc = %q", got)
	}
}
