package neutralprefix

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestAggregateGolden freezes Aggregate's whole output for one rater, two raters,
// and two raters with no id in common (the agreement ratios divide by zero).
func TestAggregateGolden(t *testing.T) {
	kf := &KeyFile{Meta: map[string]KeyMeta{}}
	add := func(rid, cls, model, sc, cond string) {
		kf.Order = append(kf.Order, rid)
		kf.Meta[rid] = KeyMeta{Cls: cls, Model: model, Scenario: sc, Condition: cond}
	}
	add("c1", "casg-direct", "mistral", "1", "baseline")
	add("c2", "casg-direct", "mistral", "1", "baseline")
	add("c3", "casg-direct", "phi4", "2", "glyph_only")
	add("f1", "formal-step-context-bypass", "qwen38", "1", "neutral_prefix")
	add("p1", "parent-state-check-bypass", "phi4", "2", "imperative_only")
	add("u1", "conditional-gate-uniform-default", "mistral", "2", "baseline")
	a := map[string]string{"c1": "C", "c2": "Ic", "c3": "C", "f1": "I", "p1": "C"}
	b := map[string]string{"c1": "C", "c2": "Ii", "c3": "Ic", "f1": "C", "u1": "C"}
	out := "### one rater\n" + Aggregate(kf, a, nil, false) +
		"### two raters\n" + Aggregate(kf, a, b, true) +
		"### no common ids\n" + Aggregate(kf, map[string]string{"c1": "C"}, map[string]string{"c2": "C"}, true)

	golden := filepath.Join("testdata", "aggregate.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("Aggregate output drifted from %s:\n%s", golden, out)
	}
}

// TestReconcileGolden freezes Reconcile's output over two classes in a given
// order, with uneven per-condition counts (one condition empty, so its rates are
// NaN) and a mix of agreed, split and old codes.
func TestReconcileGolden(t *testing.T) {
	kf := &KeyFile{Meta: map[string]KeyMeta{}}
	codes := []string{"C", "I", "Ic", "N"}
	classes := []ReconcileClass{}
	for ci, cls := range []string{"zeta", "alpha"} {
		rc := ReconcileClass{Name: cls, RaterA: map[string]string{}, RaterB: map[string]string{}, Old: map[string]string{}}
		// zeta: 200 baseline, 184 glyph_only, none in neutral_prefix or imperative_only.
		// alpha: 96 in each of the four conditions.
		sizes := map[string]int{"baseline": 200, "glyph_only": 184}
		if ci == 1 {
			sizes = map[string]int{"baseline": 96, "neutral_prefix": 96, "glyph_only": 96, "imperative_only": 96}
		}
		for _, cond := range []string{"imperative_only", "glyph_only", "neutral_prefix", "baseline"} {
			for i := 0; i < sizes[cond]; i++ {
				rid := fmt.Sprintf("%s_%s_%d", cls, cond, i)
				kf.Order = append(kf.Order, rid)
				kf.Meta[rid] = KeyMeta{Cls: cls, Condition: cond}
				rc.RaterA[rid] = codes[i%4]
				rc.RaterB[rid] = codes[(i/3)%4]
				rc.Old[rid] = codes[(i+ci)%3]
			}
		}
		classes = append(classes, rc)
	}
	out, err := Reconcile(kf, classes)
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "reconcile.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("Reconcile output drifted from %s:\n%s", golden, out)
	}
}
