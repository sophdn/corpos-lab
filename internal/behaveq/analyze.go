package behaveq

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
)

// Analysis report for a behavioral-equivalence expansion run, ported from
// analyze.py. It scores every response with the DP-1 rule (primary) and a DP-2
// batch-job check, then reports per-model, per-condition clear-rates with Wilson
// intervals, the pre-registered contrasts (Fisher exact + a Newcombe CI), and the
// n needed for 80% power from the Qwen3.8 rates.

// AnalyzeModels and AnalyzeConds are the report's row/column order, ported from
// analyze.MODELS and analyze.CONDS.
var (
	AnalyzeModels = []string{"mistral", "qwen38", "qwen2532"}
	AnalyzeConds  = []string{"baseline", "duty_only", "corpus_only"}
)

// batchRE is the DP-2 batch-job marker, ported verbatim from analyze.BATCH_RE.
var batchRE = regexp.MustCompile(`(?i)order-service|batch|reconciliation|1,?247|revalidation`)

// Z scores, ported verbatim from analyze.py.
const (
	z95  = 1.959963985 // two-sided 95%
	z90  = 1.644853627 // one-sided 95% / two-sided 90%
	zb80 = 0.841621234 // power 0.80
)

// MentionsBatch reports whether the text mentions the order-service batch
// reconciliation job (the DP-2 factor). analyze.py counts a DP-2 violation when
// this is false.
func MentionsBatch(text string) bool {
	return batchRE.MatchString(text)
}

// CellScore is one (model, condition) cell's tally: DP-1 clears, DP-2 violations
// (responses that do NOT mention the batch job), and the response count.
type CellScore struct {
	Clear   int
	DP2Viol int
	N       int
}

// ScoreCell scores one cell's responses, matching analyze.score_cell.
func ScoreCell(texts []string) CellScore {
	var cs CellScore
	for _, txt := range texts {
		cs.N++
		if Score(txt) == "cleared" {
			cs.Clear++
		}
		if !MentionsBatch(txt) {
			cs.DP2Viol++
		}
	}
	return cs
}

// ModelData holds one model's cells for the report. Present is false when the run
// directory is absent (analyze prints "(no runs)" and skips its contrasts).
type ModelData struct {
	Present bool
	Cells   map[string]CellScore // keyed by condition
}

// wilson returns the (lo, hi) Wilson score interval, ported from analyze.wilson.
func wilson(k, n int, z float64) (float64, float64) {
	if n == 0 {
		return 0.0, 0.0
	}
	p := float64(k) / float64(n)
	d := 1 + z*z/float64(n)
	c := p + z*z/float64(2*n)
	m := z * math.Sqrt((p*(1-p)+z*z/float64(4*n))/float64(n))
	return (c - m) / d, (c + m) / d
}

// comb returns the binomial coefficient C(n, k) as an exact big integer, matching
// python math.comb (0 when out of range).
func comb(n, k int) *big.Int {
	if k < 0 || k > n {
		return big.NewInt(0)
	}
	return new(big.Int).Binomial(int64(n), int64(k))
}

// pmf is the hypergeometric probability of x, computed as an exact rational then
// rounded to the nearest float64 — matching python's (int*int)/int division.
func pmf(r1, r2, c1, n, x int) float64 {
	num := new(big.Int).Mul(comb(r1, x), comb(r2, c1-x))
	den := comb(n, c1)
	fv, _ := new(big.Rat).SetFrac(num, den).Float64()
	return fv
}

// fisherTwoSided is the two-sided Fisher exact p for [[a,b],[c,d]], ported from
// analyze.fisher_two_sided.
func fisherTwoSided(a, b, c, d int) float64 {
	r1, r2, c1, n := a+b, c+d, a+c, a+b+c+d
	pObs := pmf(r1, r2, c1, n, a)
	lo := max(0, c1-r2)
	hi := min(r1, c1)
	sum := 0.0
	for x := lo; x <= hi; x++ {
		px := pmf(r1, r2, c1, n, x)
		if px <= pObs*(1+1e-7) {
			sum += px
		}
	}
	return sum
}

// fisherOneSidedGreater is the one-sided p that row-1 proportion exceeds row-2,
// ported from analyze.fisher_one_sided_greater.
func fisherOneSidedGreater(a, b, c, d int) float64 {
	r1, r2, c1, n := a+b, c+d, a+c, a+b+c+d
	hi := min(r1, c1)
	sum := 0.0
	for x := a; x <= hi; x++ {
		sum += pmf(r1, r2, c1, n, x)
	}
	return sum
}

// NewcombeDiffCI is the Newcombe (method 10) CI for p1-p2 at normal quantile z,
// the one implementation in the lab (corpos-lab stats newcombe calls it). Ported from
// analyze.newcombe_diff_ci.
func NewcombeDiffCI(k1, n1, k2, n2 int, z float64) (float64, float64) {
	l1, u1 := wilson(k1, n1, z)
	l2, u2 := wilson(k2, n2, z)
	p1 := float64(k1) / float64(n1)
	p2 := float64(k2) / float64(n2)
	lo := (p1 - p2) - math.Sqrt((p1-l1)*(p1-l1)+(u2-p2)*(u2-p2))
	hi := (p1 - p2) + math.Sqrt((u1-p1)*(u1-p1)+(p2-l2)*(p2-l2))
	return lo, hi
}

// nForSuperiority returns the per-cell n for 80% power (nil = None, when the rates
// are equal), ported from analyze.n_for_superiority.
func nForSuperiority(p1, p2, alphaZ, betaZ float64) *int {
	if math.Abs(p1-p2) < 1e-9 {
		return nil
	}
	pbar := (p1 + p2) / 2
	inner := alphaZ*math.Sqrt(2*pbar*(1-pbar)) + betaZ*math.Sqrt(p1*(1-p1)+p2*(1-p2))
	num := inner * inner
	v := int(math.Ceil(num / ((p1 - p2) * (p1 - p2))))
	return &v
}

// nForEquivalence returns the rough per-cell n for a 90% CI within the margin,
// ported from analyze.n_for_equivalence.
func nForEquivalence(p1, p2, margin, z float64) int {
	hwVar := p1*(1-p1) + p2*(1-p2)
	if hwVar < 1e-9 {
		hwVar = 0.25
	}
	return int(math.Ceil(z * z * hwVar / (margin * margin)))
}

// Report renders the full analyze.py stdout for the given suffix and per-model
// data, byte-identical to analyze.main().
func Report(suffix string, data map[string]ModelData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Analysis for run suffix '%s'\n\n", suffix)
	fmt.Fprint(&b, "## DP-1 clear-rate (primary), with 95% Wilson interval\n")
	for _, model := range AnalyzeModels {
		md := data[model]
		if !md.Present {
			fmt.Fprintf(&b, "  %s%s: (no runs)\n", model, suffix)
			continue
		}
		fmt.Fprintf(&b, "  %s:\n", model)
		for _, cond := range AnalyzeConds {
			cs := md.Cells[cond]
			lo, hi := wilson(cs.Clear, cs.N, z95)
			fmt.Fprintf(&b, "    %-12s DP-1 cleared %d/%d  [%.2f,%.2f]   DP-2 violated %d/%d\n",
				cond, cs.Clear, cs.N, lo, hi, cs.DP2Viol, cs.N)
		}
	}
	fmt.Fprint(&b, "\n## Contrasts (per model)\n")
	for _, model := range AnalyzeModels {
		md := data[model]
		if !md.Present {
			continue
		}
		bl := md.Cells["baseline"]
		du := md.Cells["duty_only"]
		co := md.Cells["corpus_only"]
		fmt.Fprintf(&b, "  %s:\n", model)
		p := fisherOneSidedGreater(co.Clear, co.N-co.Clear, bl.Clear, bl.N-bl.Clear)
		fmt.Fprintf(&b, "    corpus>brief (1-sided Fisher): p=%.4g\n", p)
		p = fisherOneSidedGreater(du.Clear, du.N-du.Clear, bl.Clear, bl.N-bl.Clear)
		fmt.Fprintf(&b, "    duty>brief   (1-sided Fisher): p=%.4g\n", p)
		p = fisherTwoSided(co.Clear, co.N-co.Clear, du.Clear, du.N-du.Clear)
		lo, hi := NewcombeDiffCI(co.Clear, co.N, du.Clear, du.N, z90)
		fmt.Fprintf(&b, "    corpus vs duty (2-sided Fisher): p=%.4g\n", p)
		fmt.Fprintf(&b, "    corpus-duty diff 90%% CI: [%+.2f,%+.2f]  (equivalent if within [-0.15,+0.15])\n", lo, hi)
	}
	if md, ok := data["qwen38"]; ok && md.Present {
		co := md.Cells["corpus_only"]
		du := md.Cells["duty_only"]
		p1 := float64(co.Clear) / float64(co.N)
		p2 := float64(du.Clear) / float64(du.N)
		fmt.Fprint(&b, "\n## n-selection (from Qwen3.8 corpus vs duty)\n")
		fmt.Fprintf(&b, "  observed rates: corpus %.3f, duty %.3f\n", p1, p2)
		ns := nForSuperiority(p1, p2, z95, zb80)
		ne := nForEquivalence(p1, p2, 0.15, z90)
		fmt.Fprintf(&b, "  n per cell for 80%% power, superiority: %s\n", noneOrInt(ns))
		fmt.Fprintf(&b, "  n per cell for a 90%% CI within +/-0.15 (equivalence, rough): %d\n", ne)
	}
	return b.String()
}

func noneOrInt(p *int) string {
	if p == nil {
		return "None"
	}
	return fmt.Sprintf("%d", *p)
}
