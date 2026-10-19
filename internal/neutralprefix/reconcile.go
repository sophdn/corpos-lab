package neutralprefix

import (
	"fmt"
	"strings"
)

// reconcileConds is reconcile_rerate.py's CONDS, in report order.
var reconcileConds = []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}

// ReconcileClass carries one class's rater inputs for Reconcile: the two new
// tightened-rubric blind raters (A/B, rid -> code) and the old single-rater union
// codes (rid -> code).
type ReconcileClass struct {
	Name   string
	RaterA map[string]string
	RaterB map[string]string
	Old    map[string]string
}

// Reconcile reproduces reconcile_rerate.py's stdout. For each class in the given
// order it reports A/B inter-rater agreement over that class's 384 ids, then per
// condition the old single-rater C-rate, the new strict-consensus C-rate, and the
// new consensus-code histogram. key supplies each id's class and condition, in the
// key file's insertion order (which fixes the histogram's key order). It returns
// the full report text (leading newline before each class header, trailing newline
// on every line), matching print() output byte for byte.
func Reconcile(key *KeyFile, classes []ReconcileClass) (string, error) {
	var b strings.Builder
	for _, c := range classes {
		ids := idsWhere(key, key.Order, func(m KeyMeta) bool { return m.Cls == c.Name })
		if len(ids) != 384 {
			return "", fmt.Errorf("%s: %d ids", c.Name, len(ids))
		}

		agree := 0
		for _, i := range ids {
			if c.RaterA[i] == c.RaterB[i] {
				agree++
			}
		}
		fmt.Fprintf(&b, "\n=== %s ===\n", c.Name)
		fmt.Fprintf(&b, "inter-rater A/B agreement: %d/%d = %.3f\n", agree, len(ids), float64(agree)/float64(len(ids)))

		for _, cond := range reconcileConds {
			cids := idsWhere(key, ids, func(m KeyMeta) bool { return m.Condition == cond })
			writeReconcileCond(&b, c, cond, cids)
		}
	}
	return b.String(), nil
}

// idsWhere returns the ids, in order, whose key metadata satisfies keep.
func idsWhere(key *KeyFile, ids []string, keep func(KeyMeta) bool) []string {
	var out []string
	for _, i := range ids {
		if keep(key.Meta[i]) {
			out = append(out, i)
		}
	}
	return out
}

// writeReconcileCond writes one condition's line: the old single-rater C-rate,
// the new strict-consensus C-rate, and the consensus-code histogram in
// first-seen order (a disagreement counts as SPLIT).
func writeReconcileCond(b *strings.Builder, c ReconcileClass, cond string, cids []string) {
	n := len(cids)
	newC, oldC := 0, 0
	cons := NewOCounter()
	for _, i := range cids {
		a, bb := c.RaterA[i], c.RaterB[i]
		if a == "C" && bb == "C" {
			newC++
		}
		if c.Old[i] == "C" {
			oldC++
		}
		if a == bb {
			cons.Inc(a)
		} else {
			cons.Inc("SPLIT")
		}
	}
	fmt.Fprintf(b, "  %-16s n=%2d  old C-rate=%.3f (%d)  new strict-C=%.3f (%d)  consensus=%s\n",
		cond, n, float64(oldC)/float64(n), oldC, float64(newC)/float64(n), newC, cons.PyDictRepr())
}
