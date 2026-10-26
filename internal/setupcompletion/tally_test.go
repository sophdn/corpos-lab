package setupcompletion

import (
	"math"
	"testing"
)

func TestKappaEmpty(t *testing.T) {
	if !math.IsNaN(Kappa(nil)) {
		t.Errorf("Kappa(nil) should be NaN")
	}
}

func TestKappaPerfectSingleCode(t *testing.T) {
	// Both raters always "C": chance agreement is total, so kappa is defined as 1.
	pairs := [][2]string{{"C", "C"}, {"C", "C"}, {"C", "C"}}
	if got := Kappa(pairs); got != 1.0 {
		t.Errorf("Kappa(all C) = %v, want 1.0", got)
	}
}

func TestKappaChanceLevel(t *testing.T) {
	// Two equally-likely codes, agreement exactly at chance -> kappa 0.
	pairs := [][2]string{{"C", "C"}, {"C", "I"}, {"I", "C"}, {"I", "I"}}
	if got := Kappa(pairs); math.Abs(got-0.0) > 1e-9 {
		t.Errorf("Kappa(chance) = %v, want 0", got)
	}
}

func TestKappaPositive(t *testing.T) {
	pairs := [][2]string{{"C", "C"}, {"C", "C"}, {"C", "C"}, {"I", "I"}, {"I", "C"}}
	got := Kappa(pairs)
	if got <= 0 || got >= 1 {
		t.Errorf("Kappa = %v, want strictly between 0 and 1", got)
	}
}

func TestTallyReportGolden(t *testing.T) {
	// Anchored to the python tally.py stdout for this exact input. 'ghost' has no
	// matched rater ids (nan footer); 'real' has two pairs.
	key := map[string]KeyEntry{
		"id1": {Glyph: "ghost", Condition: "baseline", Setup: "raw", Run: 1},
		"id2": {Glyph: "ghost", Condition: "glyph_only", Setup: "loop", Run: 2},
		"r1":  {Glyph: "real", Condition: "baseline", Setup: "raw", Run: 1},
		"r2":  {Glyph: "real", Condition: "baseline", Setup: "raw", Run: 2},
	}
	a := map[string]string{"r1": "C", "r2": "Ii"}
	b := map[string]string{"r1": "C", "r2": "C"}
	want := "\n=== ghost ===\n" +
		"setup condition          n  consC  consIi  per-rater-C(A/B)  per-rater-Ii(A/B)\n" +
		"  raw agreement: nan   kappa: nan   (n=0)\n" +
		"\n=== real ===\n" +
		"setup condition          n  consC  consIi  per-rater-C(A/B)  per-rater-Ii(A/B)\n" +
		"raw   baseline           2      1       0       1/2            1/0\n" +
		"  raw agreement: 0.500   kappa: 0.000   (n=2)\n"
	if got := TallyReport(key, a, b); got != want {
		t.Errorf("TallyReport mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestTallyReportConsensusCounts(t *testing.T) {
	// Two matched pairs in one cell: one consensus C, one consensus Ii.
	key := map[string]KeyEntry{
		"x": {Glyph: "g", Condition: "glyph_only", Setup: "loop", Run: 1},
		"y": {Glyph: "g", Condition: "glyph_only", Setup: "loop", Run: 2},
	}
	a := map[string]string{"x": "C", "y": "Ii"}
	b := map[string]string{"x": "C", "y": "Ii"}
	want := "\n=== g ===\n" +
		"setup condition          n  consC  consIi  per-rater-C(A/B)  per-rater-Ii(A/B)\n" +
		"loop  glyph_only         2      1       1       1/1            1/1\n" +
		"  raw agreement: 1.000   kappa: 1.000   (n=2)\n"
	if got := TallyReport(key, a, b); got != want {
		t.Errorf("TallyReport mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestFmt3(t *testing.T) {
	if got := fmt3(math.NaN()); got != "nan" {
		t.Errorf("fmt3(NaN) = %q, want nan", got)
	}
	if got := fmt3(0.5); got != "0.500" {
		t.Errorf("fmt3(0.5) = %q, want 0.500", got)
	}
}
