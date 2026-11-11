// Package pystats holds the small-sample statistics the scoring ports share,
// each matching its python original (analyze.py's wilson, comb, and
// fisher_two_sided) to the printed precision.
package pystats

import (
	"math"
	"math/big"
)

// Wilson returns the (lo, hi) Wilson score interval for k successes in n trials
// at normal quantile z. n==0 yields (0, 0).
func Wilson(k, n int, z float64) (lo, hi float64) {
	if n == 0 {
		return 0.0, 0.0
	}
	p := float64(k) / float64(n)
	nn := float64(n)
	d := 1 + z*z/nn
	c := p + z*z/(2*nn)
	m := z * math.Sqrt((p*(1-p)+z*z/(4*nn))/nn)
	return (c - m) / d, (c + m) / d
}

// comb returns the binomial coefficient C(n, k) as an exact big integer, or 0 for
// out-of-range k, matching python math.comb.
func comb(n, k int) *big.Int {
	if k < 0 || k > n {
		return big.NewInt(0)
	}
	return new(big.Int).Binomial(int64(n), int64(k))
}

// HyperPMF is the hypergeometric probability of x for row totals r1, r2, first
// column total c1, and grand total n. It is computed as the exact rational
// comb(r1,x)*comb(r2,c1-x)/comb(n,c1) and rounded to the nearest float64,
// matching python's (int*int)/int true division.
func HyperPMF(r1, r2, c1, n, x int) float64 {
	num := new(big.Int).Mul(comb(r1, x), comb(r2, c1-x))
	f, _ := new(big.Rat).SetFrac(num, comb(n, c1)).Float64()
	return f
}

// FisherTwoSided returns the two-sided Fisher exact p-value for the 2x2 table
// [[a,b],[c,d]]: the sum of the point masses no greater than the observed mass
// times (1 + 1e-7). Exact big-int combinatorics keep the result identical to
// python's arbitrary-precision path.
func FisherTwoSided(a, b, c, d int) float64 {
	r1, r2, c1, n := a+b, c+d, a+c, a+b+c+d
	thr := HyperPMF(r1, r2, c1, n, a) * (1 + 1e-7)
	lo := max(0, c1-r2)
	hi := min(r1, c1)
	sum := 0.0
	for x := lo; x <= hi; x++ {
		if px := HyperPMF(r1, r2, c1, n, x); px <= thr {
			sum += px
		}
	}
	return sum
}

// CRate is aggregate.py's crate(): the count of "C" codes in items, their total,
// and the rate (NaN when items is empty).
func CRate(items []string) (c, n int, r float64) {
	n = len(items)
	for _, x := range items {
		if x == "C" {
			c++
		}
	}
	if n == 0 {
		return c, n, math.NaN()
	}
	return c, n, float64(c) / float64(n)
}
