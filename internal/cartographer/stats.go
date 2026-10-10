package cartographer

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
)

// Z95 is the two-sided 95% normal quantile, matching the python constant exactly
// so Wilson bounds agree to the printed precision.
const Z95 = 1.959963985

// Wilson returns the Wilson score interval for k successes in n trials at the 95%
// level. Ported verbatim from analyze.wilson; n==0 yields (0,0).
func Wilson(k, n int) (lo, hi float64) {
	if n == 0 {
		return 0.0, 0.0
	}
	z := Z95
	p := float64(k) / float64(n)
	nn := float64(n)
	d := 1 + z*z/nn
	c := p + z*z/(2*nn)
	m := z * math.Sqrt((p*(1-p)+z*z/(4*nn))/nn)
	return (c - m) / d, (c + m) / d
}

// comb returns the binomial coefficient C(n, k) as an exact big integer, or 0 for
// out-of-range k (matching the ranges the Fisher sum uses).
func comb(n, k int) *big.Int {
	if k < 0 || k > n {
		return big.NewInt(0)
	}
	return new(big.Int).Binomial(int64(n), int64(k))
}

// FisherTwoSided returns the two-sided Fisher exact p-value for the 2x2 table
// [[a,b],[c,d]]. Ported from analyze.fisher_two_sided: each hypergeometric point
// mass is the exact rational comb(r1,x)*comb(r2,c1-x)/comb(n,c1) rounded to
// float64 (matching python int/int true division), and the p-value sums the point
// masses no greater than the observed mass times (1 + 1e-7). Using exact big-int
// combinatorics keeps the result identical to python's arbitrary-precision path.
func FisherTwoSided(a, b, c, d int) float64 {
	r1, r2, c1, n := a+b, c+d, a+c, a+b+c+d
	den := comb(n, c1)
	pmf := func(x int) float64 {
		num := new(big.Int).Mul(comb(r1, x), comb(r2, c1-x))
		f, _ := new(big.Rat).SetFrac(num, den).Float64()
		return f
	}
	pObs := pmf(a)
	thr := pObs * (1 + 1e-7)
	lo := max(0, c1-r2)
	hi := min(r1, c1)
	sum := 0.0
	for x := lo; x <= hi; x++ {
		if px := pmf(x); px <= thr {
			sum += px
		}
	}
	return sum
}

// mean returns the arithmetic mean of xs. It panics on an empty slice, matching
// python's ZeroDivisionError intent; callers pass non-empty condition groups.
func mean(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// PermTest returns a two-sided permutation p-value on the difference of means of
// xs versus ys over reps relabelings. It mirrors analyze_duty.perm_test —
// observed absolute mean difference, pooled resampling, and the (hits+1)/(reps+1)
// estimate with a -1e-12 tolerance on the >= comparison — but uses Go's seeded
// Fisher-Yates shuffle rather than python's Mersenne Twister. The result is
// therefore a Monte Carlo estimate that agrees with python within sampling error,
// not bit for bit; it is deterministic for a fixed seed. This is the structural
// parity the port targets for the randomized tools.
func PermTest(xs, ys []float64, reps int, seed int64) float64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic parity RNG, not security
	obs := math.Abs(mean(xs) - mean(ys))
	n := len(xs)
	pool := make([]float64, 0, len(xs)+len(ys))
	pool = append(pool, xs...)
	pool = append(pool, ys...)
	hits := 0
	for i := 0; i < reps; i++ {
		rng.Shuffle(len(pool), func(a, b int) { pool[a], pool[b] = pool[b], pool[a] })
		d := math.Abs(mean(pool[:n]) - mean(pool[n:]))
		if d >= obs-1e-12 {
			hits++
		}
	}
	return float64(hits+1) / float64(reps+1)
}

// nan returns a quiet NaN, used where python falls back to float("nan").
func nan() float64 { return math.NaN() }

// formatG renders x the way python's format(x, ".<prec>g") does. Go's %g agrees
// with python for finite values; python prints non-finite values in lower case
// ("nan", "inf", "-inf"), which Go's %g does not, so those are special-cased.
func formatG(x float64, prec int) string {
	switch {
	case math.IsNaN(x):
		return "nan"
	case math.IsInf(x, 1):
		return "inf"
	case math.IsInf(x, -1):
		return "-inf"
	default:
		return fmt.Sprintf("%.*g", prec, x)
	}
}
