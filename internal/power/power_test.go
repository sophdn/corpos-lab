package power

import (
	"math"
	"testing"
)

func TestPlanNPrecisionBound(t *testing.T) {
	// target 0.10 half-width at p=0.5, z=1.96: n = 1.96^2 * 0.25 / 0.01 = 96.04 -> 97.
	p, err := PlanN(0.10, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.RecommendedN != 97 || p.PrecisionN != 97 {
		t.Fatalf("got recommended=%d precision=%d, want 97", p.RecommendedN, p.PrecisionN)
	}
	if p.Bound != "precision" {
		t.Fatalf("got bound %q, want precision", p.Bound)
	}
	if p.WorstCaseP != 0.5 || p.Z != Z95 {
		t.Fatalf("defaults not applied: p=%v z=%v", p.WorstCaseP, p.Z)
	}
}

func TestPlanNBudgetBound(t *testing.T) {
	// target 0.05 needs ~385; a cap of 100 wins and the plan is budget-bound.
	p, err := PlanN(0.05, 0.5, Z95, 100)
	if err != nil {
		t.Fatal(err)
	}
	if p.RecommendedN != 100 || p.Bound != "budget" {
		t.Fatalf("got recommended=%d bound=%q, want 100 budget", p.RecommendedN, p.Bound)
	}
	if p.PrecisionN <= 100 {
		t.Fatalf("precisionN should exceed the cap, got %d", p.PrecisionN)
	}
}

func TestPlanNRejectsBadInputs(t *testing.T) {
	for _, w := range []float64{0, -0.1, 1, 1.5} {
		if _, err := PlanN(w, 0, 0, 0); err == nil {
			t.Fatalf("expected error for target half-width %v", w)
		}
	}
	if _, err := PlanN(0.1, 1.5, 0, 0); err == nil {
		t.Fatal("expected error for p > 1")
	}
	if _, err := PlanN(0.1, 0.5, -1, 0); err == nil {
		t.Fatal("expected error for z < 0")
	}
	if _, err := PlanN(0.1, 0.5, Z95, -5); err == nil {
		t.Fatal("expected error for negative budget cap")
	}
}

func TestBudgetCapPerCell(t *testing.T) {
	n, err := BudgetCapPerCell(3600, 10, 12)
	if err != nil {
		t.Fatal(err)
	}
	if n != 30 { // floor(3600/10/12)
		t.Fatalf("got %d, want 30", n)
	}
	for _, c := range []struct {
		b, r  float64
		cells int
	}{{0, 10, 1}, {3600, 0, 1}, {3600, 10, 0}} {
		if _, err := BudgetCapPerCell(c.b, c.r, c.cells); err == nil {
			t.Fatalf("expected error for %+v", c)
		}
	}
}

func TestHalfWidthForEffect(t *testing.T) {
	w, err := HalfWidthForEffect(0.2)
	if err != nil {
		t.Fatal(err)
	}
	if w != 0.1 {
		t.Fatalf("got %v, want 0.1", w)
	}
	for _, d := range []float64{0, -0.1, 1} {
		if _, err := HalfWidthForEffect(d); err == nil {
			t.Fatalf("expected error for delta %v", d)
		}
	}
}

func TestWilsonHalfWidth(t *testing.T) {
	// k=4, n=8, z=1.96: Wilson half-width ~ 0.285.
	hw, err := WilsonHalfWidth(4, 8, 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(hw-0.285) > 0.01 {
		t.Fatalf("got %v, want ~0.285", hw)
	}
	// A tighter estimate at larger n.
	hw2, err := WilsonHalfWidth(50, 100, Z95)
	if err != nil {
		t.Fatal(err)
	}
	if hw2 >= hw {
		t.Fatalf("larger n should be tighter: n=100 %v vs n=8 %v", hw2, hw)
	}
	for _, c := range []struct{ k, n int }{{0, 0}, {-1, 8}, {9, 8}} {
		if _, err := WilsonHalfWidth(c.k, c.n, 0); err == nil {
			t.Fatalf("expected error for k=%d n=%d", c.k, c.n)
		}
	}
	if _, err := WilsonHalfWidth(4, 8, -1); err == nil {
		t.Fatal("expected error for z < 0")
	}
}

func TestSequentialStop(t *testing.T) {
	// Precision reached: hw ~0.285 <= 0.3 -> stop precision.
	stop, reason, _, err := SequentialStop(4, 8, 0.3, 0, 0)
	if err != nil || !stop || reason != "precision" {
		t.Fatalf("got stop=%v reason=%q err=%v, want stop precision", stop, reason, err)
	}
	// Budget reached before precision: target too tight, n at the cap -> stop budget.
	stop, reason, _, err = SequentialStop(4, 8, 0.1, 8, 0)
	if err != nil || !stop || reason != "budget" {
		t.Fatalf("got stop=%v reason=%q err=%v, want stop budget", stop, reason, err)
	}
	// Neither: keep going.
	stop, reason, _, err = SequentialStop(4, 8, 0.1, 100, 0)
	if err != nil || stop || reason != "continue" {
		t.Fatalf("got stop=%v reason=%q err=%v, want continue", stop, reason, err)
	}
	if _, _, _, err := SequentialStop(4, 8, 0, 0, 0); err == nil {
		t.Fatal("expected error for bad target half-width")
	}
	if _, _, _, err := SequentialStop(9, 8, 0.3, 0, 0); err == nil {
		t.Fatal("expected error for k > n")
	}
}
