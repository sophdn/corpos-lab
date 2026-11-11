package cartographer

import (
	"fmt"
	"math"
	"math/rand"

	"corpos-lab/internal/pystats"
)

// Z95 is the two-sided 95% normal quantile, matching the python constant exactly
// so Wilson bounds agree to the printed precision.
const Z95 = 1.959963985

// Wilson returns the Wilson score interval for k successes in n trials at the 95%
// level. Ported verbatim from analyze.wilson; n==0 yields (0,0).
func Wilson(k, n int) (lo, hi float64) {
	return pystats.Wilson(k, n, Z95)
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
