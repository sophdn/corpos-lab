package restaxis

import "sort"

// Key identifies one grid response across models, glyphs, scenarios, arms, seeds.
type Key struct {
	Model    string
	Glyph    string
	Scenario int
	Arm      string
	Seed     int
}

// Labels maps a response to a rater's label (e.g. "OK", "OF", "OFc").
type Labels map[Key]string

// Flags maps a response to its mechanical over-fire flag (0 or 1).
type Flags map[Key]int

// overFire is the label set counted as over-firing. Ported from analyze.py OF.
var overFire = map[string]bool{"OF": true, "OFc": true}

// IsOverFire reports whether a label counts as over-firing.
func IsOverFire(label string) bool { return overFire[label] }

// Agreement is rater-A-vs-rater-B agreement over the responses both rated.
type Agreement struct {
	Shared        int     // responses both raters labeled
	OverFireAgree float64 // percent where the two agree on the over-fire axis
	ExactLabel    float64 // percent where the two gave the identical label
}

// ComputeAgreement measures A/B agreement over the shared responses. Ported from
// analyze.py's agreement block. Percentages are 0 when nothing is shared.
func ComputeAgreement(a, b Labels) Agreement {
	var shared, agree, exact int
	for k, la := range a {
		lb, ok := b[k]
		if !ok {
			continue
		}
		shared++
		if IsOverFire(la) == IsOverFire(lb) {
			agree++
		}
		if la == lb {
			exact++
		}
	}
	ag := Agreement{Shared: shared}
	if shared > 0 {
		ag.OverFireAgree = 100 * float64(agree) / float64(shared)
		ag.ExactLabel = 100 * float64(exact) / float64(shared)
	}
	return ag
}

// RaterRate is one rater's over-fire count over total for a (model, arm) cell.
// Rater is one of "A", "B", "cons" (A and B both over-fire), or "mech".
type RaterRate struct {
	Model string
	Arm   string
	Rater string
	OF    int
	N     int
}

// Rates computes per-(model, arm) over-fire counts for A, B, consensus, and
// mechanical, over the union of A and B keys. Ported from analyze.py's per-cell
// aggregation. The result is sorted by model, then Arms order, then rater order
// (A, B, cons, mech) for a stable output.
func Rates(a, b Labels, m Flags) []RaterRate {
	acc := rateAcc{}
	for k := range unionKeys(a, b) {
		la, okA := a[k]
		lb, okB := b[k]
		if okA {
			acc.add(k, "A", IsOverFire(la))
		}
		if okB {
			acc.add(k, "B", IsOverFire(lb))
		}
		if okA && okB {
			acc.add(k, "cons", IsOverFire(la) && IsOverFire(lb))
		}
		if fl, ok := m[k]; ok {
			r := acc.get(k, "mech")
			r.N++
			r.OF += fl
		}
	}
	return acc.sorted()
}

// unionKeys returns the set of keys labeled by a, b, or both.
func unionKeys(a, b Labels) map[Key]bool {
	keys := map[Key]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	return keys
}

// rateCell is one (model, arm) cell of the Rates accumulation.
type rateCell struct {
	model, arm string
}

// rateAcc accumulates one RaterRate per (model, arm, rater).
type rateAcc map[rateCell]map[string]*RaterRate

// get returns the rate for k's cell and rater, creating it on first use.
func (acc rateAcc) get(k Key, rater string) *RaterRate {
	ck := rateCell{k.Model, k.Arm}
	byR := acc[ck]
	if byR == nil {
		byR = map[string]*RaterRate{}
		acc[ck] = byR
	}
	r := byR[rater]
	if r == nil {
		r = &RaterRate{Model: k.Model, Arm: k.Arm, Rater: rater}
		byR[rater] = r
	}
	return r
}

// add counts one rated response for rater, as over-fire when of is true.
func (acc rateAcc) add(k Key, rater string, of bool) {
	r := acc.get(k, rater)
	r.N++
	if of {
		r.OF++
	}
}

// sorted flattens the rates, sorted by model, then Arms order, then rater order
// (A, B, cons, mech).
func (acc rateAcc) sorted() []RaterRate {
	raterOrder := map[string]int{"A": 0, "B": 1, "cons": 2, "mech": 3}
	var out []RaterRate
	for _, byR := range acc {
		for _, r := range byR {
			out = append(out, *r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Model != out[j].Model {
			return out[i].Model < out[j].Model
		}
		if out[i].Arm != out[j].Arm {
			return armIndex(out[i].Arm) < armIndex(out[j].Arm)
		}
		return raterOrder[out[i].Rater] < raterOrder[out[j].Rater]
	})
	return out
}

// Disagreement is one response where raters A and B disagree on the over-fire
// axis. The json shape matches the study's scores/disagreements.json.
type Disagreement struct {
	Model    string `json:"model"`
	Glyph    string `json:"glyph"`
	Scenario int    `json:"scenario"`
	Arm      string `json:"arm"`
	Seed     int    `json:"seed"`
	A        string `json:"A"`
	B        string `json:"B"`
}

// Disagreements lists responses both raters labeled where they disagree on the
// over-fire axis, sorted by (model, glyph, scenario, arm, seed). Ported from
// analyze.py's disagreement block.
func Disagreements(a, b Labels) []Disagreement {
	var out []Disagreement
	for k, la := range a {
		lb, ok := b[k]
		if !ok {
			continue
		}
		if IsOverFire(la) != IsOverFire(lb) {
			out = append(out, Disagreement{
				Model: k.Model, Glyph: k.Glyph, Scenario: k.Scenario,
				Arm: k.Arm, Seed: k.Seed, A: la, B: lb,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		x, y := out[i], out[j]
		switch {
		case x.Model != y.Model:
			return x.Model < y.Model
		case x.Glyph != y.Glyph:
			return x.Glyph < y.Glyph
		case x.Scenario != y.Scenario:
			return x.Scenario < y.Scenario
		case x.Arm != y.Arm:
			return armIndex(x.Arm) < armIndex(y.Arm)
		default:
			return x.Seed < y.Seed
		}
	})
	return out
}
