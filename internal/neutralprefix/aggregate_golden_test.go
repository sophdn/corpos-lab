package neutralprefix

import (
	"fmt"
	"testing"

	"corpos-lab/internal/testutil"
)

// TestAggregateGolden freezes Aggregate's whole output for one rater, two raters,
// and two raters with no id in common (the agreement ratios divide by zero).
func TestAggregateGolden(t *testing.T) {
	kf := keyFile(
		[5]string{"c1", "casg-direct", "mistral", "1", "baseline"},
		[5]string{"c2", "casg-direct", "mistral", "1", "baseline"},
		[5]string{"c3", "casg-direct", "phi4", "2", "glyph_only"},
		[5]string{"f1", "formal-step-context-bypass", "qwen38", "1", "neutral_prefix"},
		[5]string{"p1", "parent-state-check-bypass", "phi4", "2", "imperative_only"},
		[5]string{"u1", "conditional-gate-uniform-default", "mistral", "2", "baseline"},
	)
	a := map[string]string{"c1": "C", "c2": "Ic", "c3": "C", "f1": "I", "p1": "C"}
	b := map[string]string{"c1": "C", "c2": "Ii", "c3": "Ic", "f1": "C", "u1": "C"}
	out := "### one rater\n" + Aggregate(kf, a, nil, false) +
		"### two raters\n" + Aggregate(kf, a, b, true) +
		"### no common ids\n" + Aggregate(kf, map[string]string{"c1": "C"}, map[string]string{"c2": "C"}, true)

	testutil.Golden(t, "aggregate.golden", "Aggregate output", out)
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
	testutil.Golden(t, "reconcile.golden", "Reconcile output", out)
}
