package pystats

import (
	"math"
	"testing"
)

const z95 = 1.959963985

func floatEq(a, b float64) bool { return math.Abs(a-b) < 1e-12 }

func TestWilson(t *testing.T) {
	cases := []struct {
		k, n           int
		wantLo, wantHi float64
	}{
		{0, 0, 0.0, 0.0},
		{1, 2, 0.0945312057016527, 0.9054687942983473},
		{3, 4, 0.30064184250845644, 0.9544127392056013},
	}
	for _, c := range cases {
		lo, hi := Wilson(c.k, c.n, z95)
		if !floatEq(lo, c.wantLo) || !floatEq(hi, c.wantHi) {
			t.Errorf("Wilson(%d,%d) = (%v,%v), want (%v,%v)", c.k, c.n, lo, hi, c.wantLo, c.wantHi)
		}
	}
}

func TestFisherTwoSided(t *testing.T) {
	cases := []struct {
		a, b, c, d int
		want       float64
	}{
		{3, 1, 1, 3, 0.4857142857142857},
		{10, 0, 0, 10, 1.082508822446903e-05},
		{5, 5, 5, 5, 1.0},
	}
	for _, tc := range cases {
		got := FisherTwoSided(tc.a, tc.b, tc.c, tc.d)
		if !floatEq(got, tc.want) {
			t.Errorf("FisherTwoSided(%d,%d,%d,%d) = %v, want %v", tc.a, tc.b, tc.c, tc.d, got, tc.want)
		}
	}
}

func TestCombOutOfRange(t *testing.T) {
	// The Fisher sum never asks for an out-of-range coefficient, but the guard
	// must return 0 for k<0 or k>n rather than panic.
	if comb(3, -1).Sign() != 0 {
		t.Error("comb(3,-1) should be 0")
	}
	if comb(3, 5).Sign() != 0 {
		t.Error("comb(3,5) should be 0")
	}
	if comb(5, 2).Int64() != 10 {
		t.Errorf("comb(5,2) = %v, want 10", comb(5, 2))
	}
}

func TestCRate(t *testing.T) {
	c, n, r := CRate([]string{"C", "x", "C"})
	if c != 2 || n != 3 || math.Abs(r-2.0/3.0) > 1e-9 {
		t.Errorf("CRate = (%d,%d,%v)", c, n, r)
	}
	if _, _, r := CRate(nil); !math.IsNaN(r) {
		t.Errorf("CRate(empty) rate = %v, want NaN", r)
	}
}
