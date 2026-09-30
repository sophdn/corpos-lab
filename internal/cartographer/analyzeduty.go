package cartographer

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultPermReps and DefaultPermSeed mirror analyze_duty.perm_test's defaults.
const (
	DefaultPermReps = 100000
	DefaultPermSeed = 20260912
)

// keywordName is one rater-free keyword probe, kept in a slice to preserve the
// python dict's insertion order in the report.
type keywordName struct {
	name string
	re   *regexp.Regexp
}

// keywords is the deterministic keyword-presence set from analyze_duty: a
// case-insensitive regex per non-first-principles hazard. A response "mentions"
// the hazard when the regex matches anywhere in its text.
var keywords = []keywordName{
	{"routing", regexp.MustCompile(`(?i)\brout(e|ing)\b`)},
	{"durability(advisory)", regexp.MustCompile(`(?i)\b(advisory|durabilit|bypass)\w*`)},
	{"constraint(discovered)", regexp.MustCompile(`(?i)\b(constraint|discover)\w*`)},
}

// duty is one judged duty's per-class coverage rates: the fraction of
// first-principles slugs covered and the fraction of non-first-principles slugs
// covered. Ported from the (fpr, nfr) tuples analyze_duty builds.
type duty struct {
	fp    float64
	nonfp float64
}

// orderedDuties groups grid rows into per-condition duty lists, preserving the
// order in which conditions first appear in the grid (analyze_duty reports means
// in dict-insertion order). It returns the grouping and that condition order.
func orderedDuties(taboos []Taboo, grid []GridEntry) (map[string][]duty, []string) {
	fp := fpSlugs(taboos)
	nonfp := nonFPSlugs(taboos)
	duties := map[string][]duty{}
	var order []string
	for _, e := range grid {
		if _, seen := duties[e.Condition]; !seen {
			order = append(order, e.Condition)
		}
		fpr := covRate(e, fp)
		nfr := covRate(e, nonfp)
		duties[e.Condition] = append(duties[e.Condition], duty{fpr, nfr})
	}
	return duties, order
}

// covRate is the mean coverage over the given slugs for one row. Unlike
// analyze.rate it indexes coverage directly (python `cov[s]`), because
// analyze_duty assumes every slug is present in every row.
func covRate(e GridEntry, slugs []string) float64 {
	s := 0
	for _, slug := range slugs {
		s += e.Coverage[slug]
	}
	return float64(s) / float64(len(slugs))
}

// AnalyzeDuty produces the text of analyze_duty.py: per-duty mean coverage by
// condition; the registered duty-level condition-by-class interaction (a
// permutation test); the non-first-principles duty-level Fisher contrast; the
// first-principles cartographer-vs-baseline permutation contrast; the
// cartographer first-principles per-taboo spread; and the rater-free keyword pass.
//
// responses maps each condition to the texts of its response files (the caller
// reads runs/qwen38/out/responses/<cond>_<i>.txt for i in 1..30, skipping
// missing). The two permutation p-values are Monte Carlo estimates from a seeded
// Go RNG, so those two lines match python's within sampling error, not bit for
// bit; every other line is exact. reps and seed feed both permutation tests.
func AnalyzeDuty(taboos []Taboo, grid []GridEntry, responses map[string][]string, reps int, seed int64) string {
	fp := fpSlugs(taboos)
	duties, order := orderedDuties(taboos, grid)

	var b strings.Builder
	line := func(s string) { b.WriteString(s); b.WriteByte('\n') }

	line("== per-duty mean coverage rate (n=30 duties/condition) ==")
	for _, cond := range order {
		rows := duties[cond]
		fpm := meanField(rows, func(d duty) float64 { return d.fp })
		nfm := meanField(rows, func(d duty) float64 { return d.nonfp })
		line(fmt.Sprintf("  %-32s FP %.3f  nonFP %.3f", cond, fpm, nfm))
	}

	const annot, cart, base = "annotated_instrument", "cartographer_instrument", "baseline"
	da := diffField(duties[annot])
	dc := diffField(duties[cart])
	pInt := PermTest(da, dc, reps, seed)
	line("")
	line("== registered interaction (duty-level, permutation) ==")
	line(fmt.Sprintf("  per-duty (FP-nonFP): annotated mean %+.3f, cartographer mean %+.3f", mean(da), mean(dc)))
	line(fmt.Sprintf("  permutation p (annotated vs cartographer) = %s", formatG(pInt, 2)))

	ak := anyNonFP(duties[annot])
	ck := anyNonFP(duties[cart])
	pNF := FisherTwoSided(ak, 30-ak, ck, 30-ck)
	line("")
	line("== non-first-principles, duty-level (covers >=1 non-FP hazard) ==")
	line(fmt.Sprintf("  annotated %d/30, cartographer %d/30, Fisher p=%s", ak, ck, formatG(pNF, 2)))

	ca := fpField(duties[cart])
	ba := fpField(duties[base])
	pFP := PermTest(ca, ba, reps, seed)
	line("")
	line("== first-principles, duty-level (cartographer vs baseline) ==")
	line(fmt.Sprintf("  cartographer mean %.3f, baseline mean %.3f, permutation p=%s", mean(ca), mean(ba), formatG(pFP, 2)))

	line("")
	line("== cartographer first-principles per-taboo spread ==")
	counts := make([]int, len(fp))
	for i, s := range fp {
		counts[i] = tabooCount(grid, cart, s)
	}
	parts := make([]string, len(fp))
	lo, hi := counts[0], counts[0]
	for i, s := range fp {
		parts[i] = fmt.Sprintf("%s %d/30", lastSeg(s), counts[i])
		if counts[i] < lo {
			lo = counts[i]
		}
		if counts[i] > hi {
			hi = counts[i]
		}
	}
	line("  " + strings.Join(parts, ", "))
	line(fmt.Sprintf("  range %d/30 to %d/30", lo, hi))

	line("")
	line("== rater-free keyword presence (responses mentioning the hazard) ==")
	for _, cond := range Conds {
		texts := responses[cond]
		cells := make([]string, len(keywords))
		for i, kw := range keywords {
			cells[i] = fmt.Sprintf("%s %d/30", kw.name, countMatches(kw.re, texts))
		}
		line(fmt.Sprintf("  %-32s %s", cond, strings.Join(cells, "  ")))
	}

	return b.String()
}

// meanField averages one field over the duties, matching python's per-condition
// means. It assumes a non-empty slice.
func meanField(rows []duty, f func(duty) float64) float64 {
	s := 0.0
	for _, r := range rows {
		s += f(r)
	}
	return s / float64(len(rows))
}

// diffField returns the per-duty (FP - nonFP) values for a condition.
func diffField(rows []duty) []float64 {
	out := make([]float64, len(rows))
	for i, r := range rows {
		out[i] = r.fp - r.nonfp
	}
	return out
}

// fpField returns the per-duty first-principles rates for a condition.
func fpField(rows []duty) []float64 {
	out := make([]float64, len(rows))
	for i, r := range rows {
		out[i] = r.fp
	}
	return out
}

// anyNonFP counts duties covering at least one non-first-principles hazard.
func anyNonFP(rows []duty) int {
	n := 0
	for _, r := range rows {
		if r.nonfp > 0 {
			n++
		}
	}
	return n
}

// tabooCount counts rows of one condition whose coverage of the slug is set,
// matching analyze_duty.taboo_count.
func tabooCount(grid []GridEntry, cond, slug string) int {
	n := 0
	for _, e := range grid {
		if e.Condition == cond && e.Coverage[slug] != 0 {
			n++
		}
	}
	return n
}

// lastSeg returns the text after the final hyphen of a slug, python's
// s.split('-')[-1].
func lastSeg(s string) string {
	if i := strings.LastIndexByte(s, '-'); i >= 0 {
		return s[i+1:]
	}
	return s
}

// countMatches counts how many texts the regex matches at least once.
func countMatches(re *regexp.Regexp, texts []string) int {
	n := 0
	for _, t := range texts {
		if re.MatchString(t) {
			n++
		}
	}
	return n
}
