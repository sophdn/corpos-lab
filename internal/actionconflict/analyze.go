package actionconflict

import (
	"fmt"
	"sort"
	"strings"
)

// DetRow is a response's deterministic score plus the cell coordinates, as read
// from det_map.json.
type DetRow struct {
	Verdict    string `json:"verdict"`
	Confidence string `json:"confidence"`
	Scenario   string `json:"scenario"`
	Condition  string `json:"condition"`
	Precision  string `json:"precision"`
	Model      string `json:"model"`
}

// FinalRow is a response's final verdict: its deterministic score when the score
// was high-confidence, or the cross-family consensus otherwise.
type FinalRow struct {
	DetRow
	Final  string
	Source string // "deterministic" | "consensus" | "split"
}

// Split records the three raters' votes for a low-confidence response with no
// majority.
type Split struct {
	ID       string `json:"id"`
	Deepseek string `json:"deepseek"`
	Devstral string `json:"devstral"`
	Claude   string `json:"claude"`
}

// topVote returns the value with the highest count, breaking ties by first
// appearance — the same rule Python's Counter.most_common(1) applies to the vote
// list in [deepseek, devstral, claude] order. An empty string is a missing vote.
func topVote(votes []string) (string, int) {
	order := make([]string, 0, len(votes))
	count := make(map[string]int, len(votes))
	for _, v := range votes {
		if _, seen := count[v]; !seen {
			order = append(order, v)
		}
		count[v]++
	}
	best, bestN := "", 0
	for _, v := range order {
		if count[v] > bestN {
			best, bestN = v, count[v]
		}
	}
	return best, bestN
}

// resolveFinal decides one response's final verdict. A high-confidence score
// stands; a low-confidence one is resolved by a majority of the three rater
// families, and a vote with no majority (or a majority of missing votes) is a
// split.
func resolveFinal(detConfidence, detVerdict, deepseek, devstral, claude string) (final, source string) {
	if detConfidence == "high" {
		return detVerdict, "deterministic"
	}
	best, n := topVote([]string{deepseek, devstral, claude})
	if n >= 2 && best != "" {
		return best, "consensus"
	}
	return "SPLIT", "split"
}

// FinalVerdict resolves every response and returns the final rows keyed by id
// plus the list of splits. ds, dv and cl map an id to a rater family's
// low-confidence vote; a missing entry is an empty string.
func FinalVerdict(det map[string]DetRow, ds, dv, cl map[string]string) (map[string]FinalRow, []Split) {
	final := make(map[string]FinalRow, len(det))
	var splits []Split
	// Sort ids so the split list is deterministic; the map result is order-free.
	ids := make([]string, 0, len(det))
	for id := range det {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		d := det[id]
		f, source := resolveFinal(d.Confidence, d.Verdict, ds[id], dv[id], cl[id])
		final[id] = FinalRow{DetRow: d, Final: f, Source: source}
		if source == "split" {
			splits = append(splits, Split{ID: id, Deepseek: ds[id], Devstral: dv[id], Claude: cl[id]})
		}
	}
	return final, splits
}

// Rate is the A_canon share of a verdict tally, as a percentage.
func Rate(counts map[string]int) float64 {
	n := 0
	for _, v := range counts {
		n += v
	}
	if n == 0 {
		return 0
	}
	return 100 * float64(counts["A_canon"]) / float64(n)
}

// condKey names a condition, appending the precision for the canon_conflict legs.
func condKey(condition, precision string) string {
	if precision != "" {
		return condition + "/" + precision
	}
	return condition
}

var conditionOrder = []string{
	"baseline", "canon_aligned", "canon_conflict/weak",
	"canon_conflict/medium", "canon_conflict/strong",
	"scrambled_canon", "off_target_canon",
}

// OverrideReport renders the override-rate table and the four sub-question
// verdicts from the final rows, matching analyze.py's OVERRIDE_TABLE.md.
func OverrideReport(final map[string]FinalRow, splitCount int) string {
	pooled := map[string]map[string]int{}
	for _, d := range final {
		k := condKey(d.Condition, d.Precision)
		if pooled[k] == nil {
			pooled[k] = map[string]int{}
		}
		pooled[k][d.Final]++
	}
	orr := func(k string) float64 { return Rate(pooled[k]) }

	var b strings.Builder
	fmt.Fprint(&b, "# Override-rate table (A_canon share) — final verdicts\n\n")
	fmt.Fprintf(&b, "Source mix: deterministic high-confidence + cross-family consensus on "+
		"low-confidence. 3-way splits: %d.\n\n", splitCount)
	fmt.Fprint(&b, "## Pooled over 4 models x 3 scenarios\n\n")
	fmt.Fprint(&b, "| condition | n | A_canon | A_local | neither | unscore | override% |\n")
	fmt.Fprint(&b, "|---|---|---|---|---|---|---|\n")
	for _, k := range conditionOrder {
		c := pooled[k]
		n := 0
		for _, v := range c {
			n += v
		}
		if n == 0 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d | %.1f%% |\n",
			k, n, c["A_canon"], c["A_local"], c["neither"], c["unscoreable"], Rate(c))
	}

	weak, med, strong := orr("canon_conflict/weak"), orr("canon_conflict/medium"), orr("canon_conflict/strong")
	scrambled, offTarget := orr("scrambled_canon"), orr("off_target_canon")
	q1 := weak > 5 || med > 5 || strong > 5
	fmt.Fprint(&b, "\n## The four sub-questions (pooled)\n\n")
	fmt.Fprintf(&b, "1. Override EXISTS: canon_conflict weak/med/strong = %.1f/%.1f/%.1f%% vs "+
		"baseline %.1f%% and canon_aligned %.1f%%. -> %s\n",
		weak, med, strong, orr("baseline"), orr("canon_aligned"), yesNo(q1))
	fmt.Fprintf(&b, "2. Scales with precision: weak %.1f -> medium %.1f -> strong %.1f%%.\n", weak, med, strong)
	fmt.Fprintf(&b, "3. Survives scrambled_canon: scrambled %.1f%% vs strong %.1f%% -> %s\n",
		scrambled, strong, survives(scrambled, strong, "structure", "needs comprehension"))
	fmt.Fprintf(&b, "4. Survives off_target_canon: off-target %.1f%% vs strong %.1f%% -> %s\n",
		offTarget, strong, survives(offTarget, strong, "structure", "needs relevance"))
	return b.String()
}

func yesNo(b bool) string {
	if b {
		return "YES"
	}
	return "NO"
}

// survives reports whether a control's override rate stays above half the strong
// rate — the analyze.py threshold separating a structural effect from one that
// needs comprehension or relevance.
func survives(control, strong float64, keeps, loses string) string {
	if control > strong/2 {
		return "survives (" + keeps + ")"
	}
	return "collapses (" + loses + ")"
}
