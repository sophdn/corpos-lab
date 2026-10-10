package neutralprefix

import (
	"strings"
	"testing"
)

func aggKey() *KeyFile {
	kf := &KeyFile{Meta: map[string]KeyMeta{}}
	add := func(rid, cls, model, sc, cond string) {
		kf.Order = append(kf.Order, rid)
		kf.Meta[rid] = KeyMeta{Cls: cls, Model: model, Scenario: sc, Condition: cond}
	}
	add("a1", "casg-direct", "mistral", "1", "baseline")
	add("a2", "casg-direct", "mistral", "1", "baseline")
	add("a3", "casg-direct", "phi4", "2", "glyph_only")
	return kf
}

func TestAggregateTwoRater(t *testing.T) {
	kf := aggKey()
	a := map[string]string{"a1": "C", "a2": "C", "a3": "N"}
	b := map[string]string{"a1": "C", "a2": "x", "a3": "N"}
	got := Aggregate(kf, a, b, true)

	if !strings.HasPrefix(got, "rater-a: 3 ids | rater-b: 3 ids\n") {
		t.Fatalf("header line wrong:\n%s", got)
	}
	// exact-code agreement: a1 yes, a2 no, a3 yes -> 2/3
	if !strings.Contains(got, "agreement (exact code): 2/3 = 0.667") {
		t.Fatalf("exact agreement wrong:\n%s", got)
	}
	// C-vs-notC: a1 (C==C) yes, a2 (C vs not) no, a3 (not==not) yes -> 2/3
	if !strings.Contains(got, "agreement (C vs not-C): 2/3 = 0.667") {
		t.Fatalf("C-vs-notC agreement wrong:\n%s", got)
	}
	if !strings.Contains(got, "[strict-consensus]") {
		t.Fatalf("missing strict-consensus tag:\n%s", got)
	}
	// pooled casg-direct baseline: strict-consensus C = 1 (a1), n=2.
	if !strings.Contains(got, " 1/2  =0.50") {
		t.Fatalf("pooled baseline cell wrong:\n%s", got)
	}
	// an empty pooled cell renders " 0/0  = nan".
	if !strings.Contains(got, " 0/0  = nan") {
		t.Fatalf("nan cell missing:\n%s", got)
	}
	// header row and every class row appear.
	for _, cls := range aggregateClasses {
		if !strings.Contains(got, cls) {
			t.Errorf("missing class row %q", cls)
		}
	}
	// per-cell table: casg-direct/phi4/s2 glyph_only has 1 response (a3=N/N), which
	// is not a strict-consensus C, so 0/1.
	if !strings.Contains(got, "glyp= 0/1") {
		t.Fatalf("per-cell glyph_only count wrong:\n%s", got)
	}
}

func TestAggregateSingleRater(t *testing.T) {
	kf := aggKey()
	a := map[string]string{"a1": "C", "a2": "Ii", "a3": "N"}
	got := Aggregate(kf, a, nil, false)
	if strings.Contains(got, "rater-b") {
		t.Fatalf("single-rater should not mention rater-b:\n%s", got)
	}
	if strings.Contains(got, "agreement") {
		t.Fatalf("single-rater should print no agreement lines:\n%s", got)
	}
	if !strings.Contains(got, "[rater-a]") {
		t.Fatalf("missing rater-a tag:\n%s", got)
	}
	// rater-a code_for counts C directly: casg-direct baseline has 1 C of 2.
	if !strings.Contains(got, " 1/2  =0.50") {
		t.Fatalf("pooled baseline cell wrong:\n%s", got)
	}
}

func TestAggregateMissingRaterACode(t *testing.T) {
	// An id absent from rater A resolves to "?" in single-rater code_for; it is
	// simply not a C, so the pooled count treats it as non-C.
	kf := &KeyFile{
		Order: []string{"z1"},
		Meta:  map[string]KeyMeta{"z1": {Cls: "casg-direct", Model: "mistral", Scenario: "1", Condition: "baseline"}},
	}
	got := Aggregate(kf, map[string]string{}, nil, false)
	if !strings.Contains(got, " 0/1  =0.00") {
		t.Fatalf("missing-code cell wrong:\n%s", got)
	}
}
