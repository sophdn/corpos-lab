package studyclose

import (
	"strings"
	"testing"
)

func find(findings []Finding, check string) (Finding, bool) {
	for _, f := range findings {
		if f.Check == check {
			return f, true
		}
	}
	return Finding{}, false
}

func TestReportFullStudy(t *testing.T) {
	in := Inputs{
		Conditions:           []string{"baseline", "grounded_glyph", "scrambled_glyph", "off_target_glyph"},
		RunStatuses:          []string{"completed", "completed"},
		SecondRaters:         []string{"phi-4-14B"},
		ReconciliationMarker: true,
	}
	got := Report(in)
	for _, check := range []string{"run records", "mechanism controls", "second rater", "library reconciled"} {
		f, ok := find(got, check)
		if !ok {
			t.Fatalf("missing finding %q", check)
		}
		if f.Status != StatusPresent {
			t.Errorf("%q status = %q, want present", check, f.Status)
		}
	}
	if _, ok := find(got, "predictions"); ok {
		t.Error("no predictions file was given, so no predictions finding should appear")
	}
}

func TestReportGapsAndFlags(t *testing.T) {
	in := Inputs{
		Conditions:      []string{"baseline", "scrambled_glyph"}, // off_target absent -> flag
		RunStatuses:     []string{"completed", "failed"},         // partial -> flag
		SecondRaters:    []string{"Mistral-7B-Instruct"},         // failed the floor -> flag
		PredictionsFile: true,                                    // -> flag
		// ReconciliationMarker false -> absent
	}
	got := Report(in)

	if f, _ := find(got, "run records"); f.Status != StatusFlag {
		t.Errorf("run records status = %q, want flag", f.Status)
	}
	if f, _ := find(got, "mechanism controls"); f.Status != StatusFlag || !strings.Contains(f.Detail, "off_target_glyph absent") {
		t.Errorf("mechanism controls finding wrong: %+v", f)
	}
	if f, _ := find(got, "second rater"); f.Status != StatusFlag || !strings.Contains(f.Detail, "floor") {
		t.Errorf("second rater finding wrong: %+v", f)
	}
	if f, _ := find(got, "library reconciled"); f.Status != StatusAbsent {
		t.Errorf("library reconciled status = %q, want absent", f.Status)
	}
	if f, ok := find(got, "predictions"); !ok || f.Status != StatusFlag {
		t.Errorf("predictions finding wrong: %+v (ok=%v)", f, ok)
	}
}

func TestReportEmptyStudy(t *testing.T) {
	got := Report(Inputs{})
	if f, _ := find(got, "run records"); f.Status != StatusAbsent {
		t.Errorf("empty run records status = %q, want absent", f.Status)
	}
	if f, _ := find(got, "mechanism controls"); f.Status != StatusAbsent {
		t.Errorf("empty mechanism controls status = %q, want absent", f.Status)
	}
	if f, _ := find(got, "second rater"); f.Status != StatusAbsent {
		t.Errorf("empty second rater status = %q, want absent", f.Status)
	}
}
