package restaxis

import (
	"math/rand"
	"strings"
	"testing"
)

func TestCalibSelectClassification(t *testing.T) {
	// One of each category plus one ignored (A=OK, B=I).
	dis := k("m", "g", 1, "baseline", 1)
	aof := k("m", "g", 2, "baseline", 1)
	aok := k("m", "g", 3, "baseline", 1)
	ign := k("m", "g", 4, "baseline", 1)
	onlyA := k("m", "g", 5, "baseline", 1) // not in B -> not shared
	a := Labels{dis: "OF", aof: "OFc", aok: "OK", ign: "OK", onlyA: "OF"}
	b := Labels{dis: "OK", aof: "OF", aok: "OK", ign: "I"}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 3 {
		t.Fatalf("selected %d, want 3 (dis+aof+aok)", len(got))
	}
	seen := map[Key]bool{}
	for _, it := range got {
		seen[it.Key] = true
	}
	if !seen[dis] || !seen[aof] || !seen[aok] {
		t.Errorf("missing an expected pick: %v", seen)
	}
	if seen[ign] || seen[onlyA] {
		t.Errorf("picked an ignored/unshared key: %v", seen)
	}
}

func TestCalibSelectStratumCap(t *testing.T) {
	// Six disagreements in one (model, arm) stratum: only 4 may be taken.
	a := Labels{}
	b := Labels{}
	for s := 1; s <= 6; s++ {
		key := k("m", "g", s, "baseline", 1)
		a[key] = "OF"
		b[key] = "OK"
	}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 4 {
		t.Fatalf("stratum cap: selected %d, want 4", len(got))
	}
}

func TestCalibSelectStrataSpread(t *testing.T) {
	// Two strata, 6 disagreements each: 4 from each = 8 total.
	a := Labels{}
	b := Labels{}
	for _, arm := range []string{"baseline", "glyph_only"} {
		for s := 1; s <= 6; s++ {
			key := k("m", "g", s, arm, 1)
			a[key] = "OF"
			b[key] = "OK"
		}
	}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 8 {
		t.Fatalf("selected %d, want 8 (4 per stratum x 2)", len(got))
	}
	per := map[string]int{}
	for _, it := range got {
		per[it.Arm]++
	}
	if per["baseline"] != 4 || per["glyph_only"] != 4 {
		t.Errorf("per-arm counts = %v, want 4 each", per)
	}
}

func TestCalibSelectDisagreeMax(t *testing.T) {
	// 40 disagreements, each in its own (model) stratum so the per-stratum cap never
	// bites: the global DisagreeMax=34 must cap the total.
	a := Labels{}
	b := Labels{}
	for i := 0; i < 40; i++ {
		key := k("m"+string(rune('A'+i)), "g", 1, "baseline", 1)
		a[key] = "OF"
		b[key] = "OK"
	}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 34 {
		t.Fatalf("selected %d disagreements, want 34 (DisagreeMax)", len(got))
	}
}

func TestCalibSelectAnchorCaps(t *testing.T) {
	// 10 both-OF and 10 both-OK agreements, no disagreements: 8 of each anchor.
	a := Labels{}
	b := Labels{}
	for s := 1; s <= 10; s++ {
		of := k("m", "g", s, "baseline", 1)
		a[of] = "OF"
		b[of] = "OFc"
		ok := k("m", "g", s, "glyph_only", 1)
		a[ok] = "OK"
		b[ok] = "OK"
	}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 16 {
		t.Fatalf("selected %d, want 16 (8 OF + 8 OK anchors)", len(got))
	}
	var nof, nok int
	for _, it := range got {
		switch {
		case IsOverFire(it.A) && IsOverFire(it.B):
			nof++
		case it.A == "OK" && it.B == "OK":
			nok++
		}
	}
	if nof != 8 || nok != 8 {
		t.Errorf("anchor split = %d OF, %d OK, want 8 and 8", nof, nok)
	}
}

func TestCalibSelectTargetCap(t *testing.T) {
	// Plenty of everything; a small target caps the final count.
	a := Labels{}
	b := Labels{}
	for i := 0; i < 40; i++ {
		key := k("m"+string(rune('A'+i)), "g", 1, "baseline", 1)
		a[key] = "OF"
		b[key] = "OK"
	}
	p := DefaultCalibParams()
	p.Target = 10
	got := CalibSelect(a, b, p, identityShuffle)
	if len(got) != 10 {
		t.Fatalf("selected %d, want 10 (Target)", len(got))
	}
}

func TestCalibSelectCarriesLabels(t *testing.T) {
	key := k("m", "g", 1, "baseline", 1)
	a := Labels{key: "OF"}
	b := Labels{key: "OK"}
	got := CalibSelect(a, b, DefaultCalibParams(), identityShuffle)
	if len(got) != 1 || got[0].A != "OF" || got[0].B != "OK" {
		t.Fatalf("item = %+v, want A=OF B=OK", got)
	}
}

func TestCalibSelectDeterministicGivenSeed(t *testing.T) {
	a := Labels{}
	b := Labels{}
	for s := 1; s <= 6; s++ {
		key := k("m", "g", s, "baseline", 1)
		a[key] = "OF"
		b[key] = "OK"
	}
	one := CalibSelect(a, b, DefaultCalibParams(), rand.New(rand.NewSource(7)).Shuffle)
	two := CalibSelect(a, b, DefaultCalibParams(), rand.New(rand.NewSource(7)).Shuffle)
	if len(one) != len(two) {
		t.Fatalf("nondeterministic length: %d vs %d", len(one), len(two))
	}
	for i := range one {
		if one[i] != two[i] {
			t.Errorf("same seed diverged at %d: %+v vs %+v", i, one[i], two[i])
		}
	}
}

func TestHeadKeys(t *testing.T) {
	ks := []Key{k("m", "g", 1, "baseline", 1), k("m", "g", 2, "baseline", 1)}
	if got := headKeys(ks, 5); len(got) != 2 {
		t.Errorf("headKeys over-length = %d, want 2", len(got))
	}
	if got := headKeys(ks, 1); len(got) != 1 {
		t.Errorf("headKeys(1) = %d, want 1", len(got))
	}
}

func TestKeyLess(t *testing.T) {
	// Arm ordering follows Arms, not lexicographic.
	base := k("m", "g", 1, "baseline", 1)
	gm := k("m", "g", 1, "glyph_minus_rest", 1)
	if !keyLess(base, gm) {
		t.Errorf("expected baseline < glyph_minus_rest by Arms order")
	}
	if keyLess(gm, base) {
		t.Errorf("arm order not antisymmetric")
	}
	// Seed is the final tiebreak.
	if !keyLess(k("m", "g", 1, "baseline", 1), k("m", "g", 1, "baseline", 2)) {
		t.Errorf("seed tiebreak wrong")
	}
	// Model, glyph, scenario ordering.
	if !keyLess(k("a", "g", 1, "baseline", 1), k("b", "g", 1, "baseline", 1)) {
		t.Errorf("model order wrong")
	}
	if !keyLess(k("m", "a", 1, "baseline", 1), k("m", "b", 1, "baseline", 1)) {
		t.Errorf("glyph order wrong")
	}
	if !keyLess(k("m", "g", 1, "baseline", 1), k("m", "g", 2, "baseline", 1)) {
		t.Errorf("scenario order wrong")
	}
}

func TestRenderCalib(t *testing.T) {
	items := []CalibRendered{
		{Item: CalibPick{Key: k("mistral", "casg", 3, "glyph_only", 2), A: "OF", B: "OK"},
			Task: "do the thing", GroundTruth: "the correct thing", Response: "here you go"},
		{Item: CalibPick{Key: k("phi4", "casg", 4, "baseline", 1), A: "OK", B: "OK"},
			Task: "second task", GroundTruth: "second gt", Response: "second resp"},
	}
	doc := RenderCalib(items)
	if len(doc.Map) != 2 {
		t.Fatalf("map has %d entries, want 2", len(doc.Map))
	}
	if _, ok := doc.Map["C00"]; !ok {
		t.Errorf("missing id C00")
	}
	e := doc.Map["C01"]
	if e.Model != "phi4" || e.Scenario != 4 || e.Arm != "baseline" || e.A != "OK" || e.B != "OK" {
		t.Errorf("C01 keymap = %+v, wrong round-trip", e)
	}
	// The sheet shows terrain but hides arm and both judges' labels.
	if !strings.Contains(doc.Sheet, "do the thing") || !strings.Contains(doc.Sheet, "here you go") {
		t.Errorf("sheet dropped task/response terrain")
	}
	for _, arm := range Arms {
		if strings.Contains(doc.Sheet, arm) {
			t.Errorf("sheet leaks arm %q", arm)
		}
	}
	if strings.Contains(doc.Sheet, "YOUR LABEL for C00") == false {
		t.Errorf("sheet missing label prompt")
	}
	// Answers has one row per item plus the header lines, and ends with a newline.
	if strings.Count(doc.Answers, "| C0") != 2 {
		t.Errorf("answers rows = %d, want 2", strings.Count(doc.Answers, "| C0"))
	}
	if !strings.HasSuffix(doc.Answers, "\n") {
		t.Errorf("answers must end with a newline")
	}
}

func TestRenderCalibEmpty(t *testing.T) {
	doc := RenderCalib(nil)
	if len(doc.Map) != 0 {
		t.Errorf("empty render produced %d map entries", len(doc.Map))
	}
	if !strings.Contains(doc.Sheet, "Calibration sheet") {
		t.Errorf("empty sheet missing header")
	}
}

func TestSummarizeCalib(t *testing.T) {
	m := map[string]CalibMapEntry{
		"C00": {Model: "m", Arm: "baseline", A: "OF", B: "OK"},   // disagree
		"C01": {Model: "m", Arm: "glyph_only", A: "OK", B: "OK"}, // agree
		"C02": {Model: "p", Arm: "baseline", A: "OFc", B: "OF"},  // agree (both OF)
	}
	c := SummarizeCalib(m)
	if c.Total != 3 || c.Disagree != 1 || c.Agree != 2 {
		t.Errorf("counts = %+v, want total 3 disagree 1 agree 2", c)
	}
	if c.ByModel["m"] != 2 || c.ByModel["p"] != 1 {
		t.Errorf("by model = %v", c.ByModel)
	}
	if c.ByArm["baseline"] != 2 || c.ByArm["glyph_only"] != 1 {
		t.Errorf("by arm = %v", c.ByArm)
	}
}

func TestSharedSortedKeys(t *testing.T) {
	x := k("m", "g", 2, "baseline", 1)
	y := k("m", "g", 1, "baseline", 1)
	z := k("m", "g", 3, "baseline", 1) // only in a
	a := Labels{x: "OK", y: "OK", z: "OK"}
	b := Labels{x: "OK", y: "OK"}
	got := sharedSortedKeys(a, b)
	if len(got) != 2 {
		t.Fatalf("shared = %d, want 2", len(got))
	}
	if got[0] != y || got[1] != x {
		t.Errorf("not sorted: %+v", got)
	}
}
