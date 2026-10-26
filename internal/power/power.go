// Package power sizes a study's per-cell replicate count from a precision target
// and a budget, and provides the sequential stopping rule the runner prefers
// over a fixed n. It removes the recurring need to hand-set n: give a target (a
// desired 95% CI half-width, or an effect size resolved to one) and a budget (a
// per-cell run cap, e.g. derived from a wall-clock budget), and it returns the
// recommended n, which bound set it, and the rationale — recorded with the study
// so the choice is auditable (the record-what-ran invariant).
//
// The lab measures proportions (an override rate, a strict-consensus C rate), so
// the whole package is proportion sizing. Planning uses the normal approximation
// n = z^2 p(1-p) / w^2; the sequential readout uses the Wilson score interval,
// which is accurate at the small n and extreme p a graded grid produces.
package power

import (
	"fmt"
	"math"
)

// Z95 is the standard-normal quantile for a two-sided 95% interval.
const Z95 = 1.96

// Plan is a sized recommendation, JSON-serializable to record with a study.
type Plan struct {
	TargetHalfWidth float64 `json:"target_half_width"`
	WorstCaseP      float64 `json:"worst_case_p"`
	Z               float64 `json:"z"`
	BudgetCapN      int     `json:"budget_cap_n"` // 0 = no cap
	PrecisionN      int     `json:"precision_n"`  // n to hit the target at worst-case p
	RecommendedN    int     `json:"recommended_n"`
	Bound           string  `json:"bound"` // "precision" | "budget"
	Rationale       string  `json:"rationale"`
}

// defaults resolves the zero-value conveniences: p=0.5 (the widest interval),
// z=Z95. It validates the resolved values.
func defaults(worstCaseP, z float64) (float64, float64, error) {
	if worstCaseP == 0 {
		worstCaseP = 0.5
	}
	if worstCaseP < 0 || worstCaseP > 1 {
		return 0, 0, fmt.Errorf("power: worst-case p must be in [0,1], got %v", worstCaseP)
	}
	if z == 0 {
		z = Z95
	}
	if z <= 0 {
		return 0, 0, fmt.Errorf("power: z must be > 0, got %v", z)
	}
	return worstCaseP, z, nil
}

// PlanN sizes the per-cell n. PrecisionN is the smallest n whose 95% CI
// half-width at worstCaseP is <= targetHalfWidth (normal approx). When a budget
// cap is set and the precision n exceeds it, the cap wins and the plan is
// budget-bound (the CI will be wider than the target); otherwise the precision n
// wins. worstCaseP defaults to 0.5 and z to Z95 when passed as zero.
func PlanN(targetHalfWidth, worstCaseP, z float64, budgetCapN int) (Plan, error) {
	if targetHalfWidth <= 0 || targetHalfWidth >= 1 {
		return Plan{}, fmt.Errorf("power: target half-width must be in (0,1), got %v", targetHalfWidth)
	}
	if budgetCapN < 0 {
		return Plan{}, fmt.Errorf("power: budget cap must be >= 0, got %d", budgetCapN)
	}
	worstCaseP, z, err := defaults(worstCaseP, z)
	if err != nil {
		return Plan{}, err
	}
	precisionN := int(math.Ceil(z * z * worstCaseP * (1 - worstCaseP) / (targetHalfWidth * targetHalfWidth)))
	if precisionN < 1 {
		precisionN = 1
	}
	p := Plan{
		TargetHalfWidth: targetHalfWidth, WorstCaseP: worstCaseP, Z: z,
		BudgetCapN: budgetCapN, PrecisionN: precisionN,
	}
	switch {
	case budgetCapN > 0 && precisionN > budgetCapN:
		p.RecommendedN = budgetCapN
		p.Bound = "budget"
		p.Rationale = fmt.Sprintf("budget-bound: a %.3f half-width at p=%.2f needs n=%d, over the "+
			"budget cap of %d; running n=%d, so the CI will be wider than target.",
			targetHalfWidth, worstCaseP, precisionN, budgetCapN, budgetCapN)
	default:
		p.RecommendedN = precisionN
		p.Bound = "precision"
		capNote := ""
		if budgetCapN > 0 {
			capNote = fmt.Sprintf(", within the budget cap of %d", budgetCapN)
		}
		p.Rationale = fmt.Sprintf("precision-bound: n=%d reaches a %.3f 95%% CI half-width at "+
			"worst-case p=%.2f%s.", precisionN, targetHalfWidth, worstCaseP, capNote)
	}
	return p, nil
}

// BudgetCapPerCell turns a wall-clock (or compute) budget into a per-cell run
// cap: floor(budgetSeconds / perRunSeconds / cells). It is how "run for at most
// an hour" becomes a number the planner and the stopping rule can honour.
func BudgetCapPerCell(budgetSeconds, perRunSeconds float64, cells int) (int, error) {
	if budgetSeconds <= 0 {
		return 0, fmt.Errorf("power: budget seconds must be > 0, got %v", budgetSeconds)
	}
	if perRunSeconds <= 0 {
		return 0, fmt.Errorf("power: per-run seconds must be > 0, got %v", perRunSeconds)
	}
	if cells < 1 {
		return 0, fmt.Errorf("power: cells must be >= 1, got %d", cells)
	}
	return int(math.Floor(budgetSeconds / perRunSeconds / float64(cells))), nil
}

// HalfWidthForEffect resolves an effect-size target (a proportion difference to
// call real) to a CI half-width target. Rule of thumb: size each arm's interval
// so a gap of delta does not vanish into noise — a half-width of delta/2 keeps
// two arms delta apart from overlapping at the point estimate. Documented as a
// heuristic, not a power calculation.
func HalfWidthForEffect(delta float64) (float64, error) {
	if delta <= 0 || delta >= 1 {
		return 0, fmt.Errorf("power: effect size (proportion difference) must be in (0,1), got %v", delta)
	}
	return delta / 2, nil
}

// WilsonHalfWidth is the ± term of the Wilson score interval for k successes in
// n trials at quantile z (z defaults to Z95 when zero). It is the sequential
// precision readout: after each batch, this is how tight the estimate is now.
func WilsonHalfWidth(k, n int, z float64) (float64, error) {
	if n < 1 {
		return 0, fmt.Errorf("power: n must be >= 1, got %d", n)
	}
	if k < 0 || k > n {
		return 0, fmt.Errorf("power: k must be in [0,n], got k=%d n=%d", k, n)
	}
	if z == 0 {
		z = Z95
	}
	if z <= 0 {
		return 0, fmt.Errorf("power: z must be > 0, got %v", z)
	}
	nf := float64(n)
	phat := float64(k) / nf
	denom := 1 + z*z/nf
	margin := (z / denom) * math.Sqrt(phat*(1-phat)/nf+z*z/(4*nf*nf))
	return margin, nil
}

// SequentialStop is the stopping rule: given k successes so far in n runs, stop
// when the Wilson half-width is at or under the target (reason "precision") or
// the budget cap is reached (reason "budget", when budgetCapN > 0), whichever
// comes first; otherwise continue. This is the rule the runner prefers over a
// fixed n — run in batches, consult this after each, stop at the first bound.
func SequentialStop(k, n int, targetHalfWidth float64, budgetCapN int, z float64) (stop bool, reason string, hw float64, err error) {
	if targetHalfWidth <= 0 || targetHalfWidth >= 1 {
		return false, "", 0, fmt.Errorf("power: target half-width must be in (0,1), got %v", targetHalfWidth)
	}
	hw, err = WilsonHalfWidth(k, n, z)
	if err != nil {
		return false, "", 0, err
	}
	if hw <= targetHalfWidth {
		return true, "precision", hw, nil
	}
	if budgetCapN > 0 && n >= budgetCapN {
		return true, "budget", hw, nil
	}
	return false, "continue", hw, nil
}
