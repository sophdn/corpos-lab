package restaxis

import "testing"

func k(model, glyph string, scen int, arm string, seed int) Key {
	return Key{Model: model, Glyph: glyph, Scenario: scen, Arm: arm, Seed: seed}
}

func TestIsOverFire(t *testing.T) {
	for _, l := range []string{"OF", "OFc"} {
		if !IsOverFire(l) {
			t.Errorf("%q should be over-fire", l)
		}
	}
	for _, l := range []string{"OK", "", "N"} {
		if IsOverFire(l) {
			t.Errorf("%q should not be over-fire", l)
		}
	}
}

func TestComputeAgreement(t *testing.T) {
	a := Labels{
		k("m", "g", 1, "baseline", 1): "OK",
		k("m", "g", 1, "baseline", 2): "OF",
		k("m", "g", 1, "baseline", 3): "OK",
		k("m", "g", 2, "baseline", 1): "OF", // not in b -> ignored
	}
	b := Labels{
		k("m", "g", 1, "baseline", 1): "OK",  // exact + axis agree
		k("m", "g", 1, "baseline", 2): "OFc", // axis agree, not exact
		k("m", "g", 1, "baseline", 3): "OF",  // axis disagree
	}
	ag := ComputeAgreement(a, b)
	if ag.Shared != 3 {
		t.Fatalf("shared = %d, want 3", ag.Shared)
	}
	// axis agreement 2/3, exact 1/3
	if ag.OverFireAgree < 66.6 || ag.OverFireAgree > 66.7 {
		t.Errorf("over-fire agreement = %.2f, want ~66.67", ag.OverFireAgree)
	}
	if ag.ExactLabel < 33.3 || ag.ExactLabel > 33.4 {
		t.Errorf("exact-label = %.2f, want ~33.33", ag.ExactLabel)
	}
}

func TestComputeAgreementEmpty(t *testing.T) {
	ag := ComputeAgreement(Labels{}, Labels{})
	if ag.Shared != 0 || ag.OverFireAgree != 0 || ag.ExactLabel != 0 {
		t.Errorf("empty agreement = %+v, want zeros", ag)
	}
}

func TestDisagreements(t *testing.T) {
	a := Labels{
		k("m", "g", 1, "baseline", 1):   "OK",
		k("m", "g", 2, "glyph_only", 1): "OF",
		k("m", "g", 3, "baseline", 1):   "OF", // b agrees (both OF) -> no disagreement
	}
	b := Labels{
		k("m", "g", 1, "baseline", 1):   "OF", // disagree
		k("m", "g", 2, "glyph_only", 1): "OK", // disagree
		k("m", "g", 3, "baseline", 1):   "OFc",
	}
	dis := Disagreements(a, b)
	if len(dis) != 2 {
		t.Fatalf("got %d disagreements, want 2: %+v", len(dis), dis)
	}
	// sorted: scenario 1 before 2
	if dis[0].Scenario != 1 || dis[1].Scenario != 2 {
		t.Errorf("disagreements not sorted by scenario: %+v", dis)
	}
	if dis[0].A != "OK" || dis[0].B != "OF" {
		t.Errorf("first disagreement labels = %s/%s, want OK/OF", dis[0].A, dis[0].B)
	}
}

func TestRates(t *testing.T) {
	a := Labels{
		k("m", "g", 1, "baseline", 1): "OF",
		k("m", "g", 1, "baseline", 2): "OK",
	}
	b := Labels{
		k("m", "g", 1, "baseline", 1): "OF",
		k("m", "g", 1, "baseline", 2): "OF",
	}
	m := Flags{
		k("m", "g", 1, "baseline", 1): 1,
		k("m", "g", 1, "baseline", 2): 0,
	}
	rates := Rates(a, b, m)
	got := map[string]RaterRate{}
	for _, r := range rates {
		got[r.Rater] = r
	}
	if got["A"].OF != 1 || got["A"].N != 2 {
		t.Errorf("A rate = %+v, want 1/2", got["A"])
	}
	if got["B"].OF != 2 || got["B"].N != 2 {
		t.Errorf("B rate = %+v, want 2/2", got["B"])
	}
	if got["cons"].OF != 1 || got["cons"].N != 2 {
		t.Errorf("cons rate = %+v, want 1/2", got["cons"])
	}
	if got["mech"].OF != 1 || got["mech"].N != 2 {
		t.Errorf("mech rate = %+v, want 1/2", got["mech"])
	}
}

func TestRatesSortedByModelArmRater(t *testing.T) {
	a := Labels{
		k("z", "g", 1, "glyph_only", 1):       "OK",
		k("a", "g", 1, "glyph_minus_rest", 1): "OK",
		k("a", "g", 1, "baseline", 1):         "OK",
	}
	rates := Rates(a, Labels{}, Flags{})
	if rates[0].Model != "a" || rates[0].Arm != "baseline" {
		t.Errorf("first rate = %+v, want model a / baseline", rates[0])
	}
	if rates[len(rates)-1].Model != "z" {
		t.Errorf("last rate model = %q, want z", rates[len(rates)-1].Model)
	}
}

func TestDisagreementsSortAcrossAxes(t *testing.T) {
	// Exercises every sort branch: differing glyph, differing arm (including an
	// arm outside Arms, so armIndex hits its fallback), and differing seed.
	mk := func(glyph string, arm string, seed int) Key { return k("m", glyph, 1, arm, seed) }
	a := Labels{
		mk("b", "baseline", 1): "OF",
		mk("a", "zzz", 1):      "OF",
		mk("a", "baseline", 2): "OF",
		mk("a", "baseline", 1): "OF",
	}
	b := Labels{
		mk("b", "baseline", 1): "OK",
		mk("a", "zzz", 1):      "OK",
		mk("a", "baseline", 2): "OK",
		mk("a", "baseline", 1): "OK",
	}
	dis := Disagreements(a, b)
	if len(dis) != 4 {
		t.Fatalf("got %d, want 4", len(dis))
	}
	if dis[0].Glyph != "a" || dis[len(dis)-1].Glyph != "b" {
		t.Errorf("glyph not sorted: %+v", dis)
	}
	// within glyph a: baseline seed1, baseline seed2, then the unknown arm zzz last
	if dis[0].Arm != "baseline" || dis[0].Seed != 1 || dis[1].Seed != 2 || dis[2].Arm != "zzz" {
		t.Errorf("arm/seed not sorted: %+v", dis[:3])
	}
}
