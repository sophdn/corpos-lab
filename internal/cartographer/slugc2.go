package cartographer

import (
	"fmt"
	"sort"
	"strings"
)

// C2Slugs are the ten registered taboo slugs in the fixed order slug_c2.py
// declares. A slug is "cited" only when its exact kebab string appears in a
// response, so canon is a runtime dependency of the produced duty.
var C2Slugs = []string{
	"triggers-routing",
	"explicit-prerequisite-gate",
	"investigation-fix-criterion-predeclaration",
	"investigation-fix-scope-boundary",
	"fix-locus-identification",
	"fix-attempt-root-cause-reassessment",
	"investigation-advisory-fix-durability",
	"completion-evidence-required",
	"document-claim-verification",
	"known-constraint-documentation",
}

// CitedSlugs returns the C2Slugs (in C2Slugs order) whose exact string appears in
// text. Ported from slug_c2.cited_slugs; the match is a literal substring, so no
// slug is a prose false positive because every slug is a hyphenated token.
func CitedSlugs(text string) []string {
	out := []string{} // non-nil so an empty result serialises as [] not null, matching python
	for _, s := range C2Slugs {
		if strings.Contains(text, s) {
			out = append(out, s)
		}
	}
	return out
}

// SlugC2Row is one response's slug-citation record, shaped to match the python
// slug_c2.json grid values.
type SlugC2Row struct {
	Condition string   `json:"condition"`
	Run       int      `json:"run"`
	NSlugs    int      `json:"n_slugs"`
	Slugs     []string `json:"slugs"`
}

// SlugC2CondSummary is the per-condition summary, matching the python summary
// values.
type SlugC2CondSummary struct {
	Runs          int     `json:"runs"`
	RunsCitingGe1 int     `json:"runs_citing_ge1"`
	MeanSlugs     float64 `json:"mean_slugs"`
}

// SlugC2Summary aggregates rows into the per-condition summary and the printed
// table (the leading table only; the caller appends the "wrote <path>" line). The
// table iterates conditions in sorted order, matching slug_c2's `sorted(by_cond)`.
func SlugC2Summary(rows []SlugC2Row) (map[string]SlugC2CondSummary, string) {
	byCond := map[string][]int{}
	for _, r := range rows {
		byCond[r.Condition] = append(byCond[r.Condition], r.NSlugs)
	}
	conds := make([]string, 0, len(byCond))
	for c := range byCond {
		conds = append(conds, c)
	}
	sort.Strings(conds)

	summary := make(map[string]SlugC2CondSummary, len(conds))
	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }
	line(fmt.Sprintf("%-32s %4s %16s %11s", "condition", "runs", "runs citing >=1", "mean slugs"))
	for _, cond := range conds {
		counts := byCond[cond]
		n := len(counts)
		cited := 0
		total := 0
		for _, c := range counts {
			if c > 0 {
				cited++
			}
			total += c
		}
		m := 0.0
		if n != 0 {
			m = float64(total) / float64(n)
		}
		summary[cond] = SlugC2CondSummary{Runs: n, RunsCitingGe1: cited, MeanSlugs: m}
		line(fmt.Sprintf("%-32s %4d %16d %11.2f", cond, n, cited, m))
	}
	return summary, b.String()
}
