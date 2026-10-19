package pubshared

import (
	"fmt"
	"strings"
)

// Codes are represented as strings; the empty string stands for python's None
// (a missing or unscored id). Real codes are always non-empty.

// Agreement ports anchor_eval.agreement: over the ids where both pred and gold
// have a code, it returns the exact-match count, the C-vs-not-C match count, and
// the common count.
func Agreement(pred, gold map[string]string, ids []string) (exact, cnc, n int) {
	for _, i := range ids {
		if pred[i] == "" || gold[i] == "" {
			continue
		}
		n++
		if pred[i] == gold[i] {
			exact++
		}
		if (pred[i] == "C") == (gold[i] == "C") {
			cnc++
		}
	}
	return exact, cnc, n
}

// BuildHuman maps each rid to the human code at its positional slot: the kth rid
// (0-based) takes human code "H{k+1:02d}". A missing key yields "" (python None).
// Ports anchor_eval.main's human map construction.
func BuildHuman(rids []string, humanByH map[string]string) map[string]string {
	human := make(map[string]string, len(rids))
	for k, rid := range rids {
		human[rid] = humanByH[fmt.Sprintf("H%02d", k+1)]
	}
	return human
}

// BuildClaude ports anchor_eval.main's claude map: raterA's code when it agrees
// with raterB, else "SPLIT".
func BuildClaude(rids []string, ra, rb map[string]string) map[string]string {
	claude := make(map[string]string, len(rids))
	for _, rid := range rids {
		if ra[rid] == rb[rid] {
			claude[rid] = ra[rid]
		} else {
			claude[rid] = "SPLIT"
		}
	}
	return claude
}

// reportBlock ports anchor_eval.report: the agreement block for one rater under
// test.
func reportBlock(b *strings.Builder, label string, pred, human, claude map[string]string, ids []string) {
	fmt.Fprintf(b, "\n=== %s ===\n", label)
	for _, pair := range []struct {
		name string
		gold map[string]string
	}{{"human", human}, {"claude-consensus", claude}} {
		e, c, n := Agreement(pred, pair.gold, ids)
		fmt.Fprintf(b, "  vs %-16s exact %2d/%d = %s   C-vs-notC %2d/%d = %s\n",
			pair.name, e, n, pf3(ratio(e, n)), c, n, pf3(ratio(c, n)))
	}
	var humanN []string
	for _, i := range ids {
		if human[i] == "N" {
			humanN = append(humanN, i)
		}
	}
	gotN := 0
	for _, i := range humanN {
		if pred[i] == "N" {
			gotN++
		}
	}
	overN := 0
	for _, i := range ids {
		if pred[i] == "N" && human[i] != "N" && human[i] != "" {
			overN++
		}
	}
	fmt.Fprintf(b, "  off-task N recall (human-N -> rater-N): %d/%d   over-called N (rater-N, human-not-N): %d\n",
		gotN, len(humanN), overN)
	lenient, strict := 0, 0
	for _, i := range ids {
		if pred[i] == "C" && human[i] != "C" && human[i] != "" {
			lenient++
		}
		if human[i] == "C" && pred[i] != "C" && pred[i] != "" {
			strict++
		}
	}
	fmt.Fprintf(b, "  C-boundary: rater-C / human-not-C = %d (lenient)   human-C / rater-not-C = %d (strict)\n",
		lenient, strict)
}

// AnchorEvalReport ports anchor_eval.main's output: the header line, the
// claude-vs-human self-check, and an optional rater-under-test block. rater is
// nil for the self-check-only run.
func AnchorEvalReport(anchorName string, rids []string, human, claude map[string]string, rater map[string]string, raterLabel string) string {
	var b strings.Builder
	humanNCount := 0
	for _, v := range human {
		if v == "N" {
			humanNCount++
		}
	}
	splitCount := 0
	for _, v := range claude {
		if v == "SPLIT" {
			splitCount++
		}
	}
	fmt.Fprintf(&b, "anchor %s: %d items  |  human-N=%d  claude-splits=%d\n",
		anchorName, len(rids), humanNCount, splitCount)
	reportBlock(&b, "SELF-CHECK: claude-consensus vs human", claude, human, claude, rids)
	if rater != nil {
		reportBlock(&b, raterLabel, rater, human, claude, rids)
	}
	return b.String()
}

// ratio is e/n as a float; n==0 yields NaN (python would raise, but real anchors
// always have n>0).
func ratio(e, n int) float64 {
	if n == 0 {
		return 0 // avoid a NaN that would never occur on real data
	}
	return float64(e) / float64(n)
}

func pf3(x float64) string {
	return fmt.Sprintf("%.3f", x)
}
