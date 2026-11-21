package neutralprefix

import (
	"fmt"
	"strings"
	"testing"
)

// build384 returns a KeyFile for one class with 96 ids per condition (384 total),
// in the given insertion order, plus rater maps produced by codeFn(cond, index).
func build384(class string) *KeyFile {
	kf := &KeyFile{Meta: map[string]KeyMeta{}}
	conds := []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}
	for _, cond := range conds {
		for i := 0; i < 96; i++ {
			rid := fmt.Sprintf("%s_%s_%d", class, cond, i)
			kf.Order = append(kf.Order, rid)
			kf.Meta[rid] = KeyMeta{Cls: class, Condition: cond}
		}
	}
	return kf
}

func TestReconcileAllC(t *testing.T) {
	kf := build384("cls")
	a := map[string]string{}
	b := map[string]string{}
	old := map[string]string{}
	for _, rid := range kf.Order {
		a[rid], b[rid], old[rid] = "C", "C", "C"
	}
	got, err := Reconcile(kf, []ReconcileClass{{Name: "cls", RaterA: a, RaterB: b, Old: old}})
	if err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	fmt.Fprintf(&want, "\n=== cls ===\n")
	fmt.Fprintf(&want, "inter-rater A/B agreement: 384/384 = %.3f\n", 1.0)
	for _, cond := range []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"} {
		fmt.Fprintf(&want, "  %-16s n=%2d  old C-rate=%.3f (%d)  new strict-C=%.3f (%d)  consensus=%s\n",
			cond, 96, 1.0, 96, 1.0, 96, "{'C': 96}")
	}
	if got != want.String() {
		t.Fatalf("reconcile all-C mismatch:\n got %q\nwant %q", got, want.String())
	}
}

func TestReconcileSplitAndOrder(t *testing.T) {
	kf := build384("cls")
	a := map[string]string{}
	b := map[string]string{}
	old := map[string]string{}
	// For baseline: first id N/N (agree N), second id C/Ii (split), rest I/I.
	base := []string{}
	for _, rid := range kf.Order {
		if kf.Meta[rid].Condition == "baseline" {
			base = append(base, rid)
		}
		a[rid], b[rid], old[rid] = "I", "I", "x"
	}
	a[base[0]], b[base[0]] = "N", "N"
	a[base[1]], b[base[1]] = "C", "Ii"
	old[base[0]] = "C"
	got, err := Reconcile(kf, []ReconcileClass{{Name: "cls", RaterA: a, RaterB: b, Old: old}})
	if err != nil {
		t.Fatal(err)
	}
	// First-seen order in baseline: N (id0), SPLIT (id1), then I (id2..).
	if !strings.Contains(got, "consensus={'N': 1, 'SPLIT': 1, 'I': 94}") {
		t.Fatalf("baseline consensus histogram wrong:\n%s", got)
	}
	// baseline agreement: only base[1] disagrees -> 383/384 exact across the class.
	if !strings.Contains(got, "inter-rater A/B agreement: 383/384") {
		t.Fatalf("agreement wrong:\n%s", got)
	}
	// baseline old C-rate: 1/96
	if !strings.Contains(got, fmt.Sprintf("old C-rate=%.3f (1)", 1.0/96.0)) {
		t.Fatalf("old C-rate wrong:\n%s", got)
	}
}

func TestReconcileWrongCount(t *testing.T) {
	kf := &KeyFile{Meta: map[string]KeyMeta{"a": {Cls: "cls", Condition: "baseline"}}, Order: []string{"a"}}
	_, err := Reconcile(kf, []ReconcileClass{{Name: "cls"}})
	if err == nil || !strings.Contains(err.Error(), "1 ids") {
		t.Fatalf("expected 384-count error, got %v", err)
	}
}
