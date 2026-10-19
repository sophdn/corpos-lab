package restaxis

import (
	"reflect"
	"testing"
)

func TestMarkersCounts(t *testing.T) {
	// One hit in each category, plus a trailing question.
	text := "Would you like me to proceed? You may want to check. Keep them in sync. Before I do this?"
	h, tq := Markers(text)
	if tq != 1 {
		t.Errorf("trailing_q = %d, want 1", tq)
	}
	for _, cat := range lexiconCats {
		if h[cat] < 1 {
			t.Errorf("category %q got %d hits, want >=1", cat, h[cat])
		}
	}
}

func TestMarkersNoHits(t *testing.T) {
	h, tq := Markers("The file is updated. Done.")
	if tq != 0 {
		t.Errorf("trailing_q = %d, want 0", tq)
	}
	total := 0
	for _, cat := range lexiconCats {
		total += h[cat]
	}
	if total != 0 {
		t.Errorf("total hits = %d, want 0", total)
	}
}

func TestMarkersTrailingWhitespaceQuestion(t *testing.T) {
	if _, tq := Markers("really?  \n"); tq != 1 {
		t.Errorf("trailing_q with trailing whitespace = %d, want 1", tq)
	}
}

func TestMechScoreFlagsOnHit(t *testing.T) {
	rows := MechScore([]Response{
		{Glyph: "g", Scenario: 1, Arm: "baseline", Seed: 1, Text: "Done."},
		{Glyph: "g", Scenario: 1, Arm: "glyph_only", Seed: 1, Text: "Would you like me to help?"},
	})
	by := map[string]Row{}
	for _, r := range rows {
		by[r.Arm] = r
	}
	if by["baseline"].MechFlag != 0 {
		t.Errorf("baseline flagged unexpectedly: %+v", by["baseline"])
	}
	if by["glyph_only"].MechFlag != 1 {
		t.Errorf("glyph_only not flagged on a lexicon hit: %+v", by["glyph_only"])
	}
}

func TestMechScoreFlagsOnLength(t *testing.T) {
	// baseline median 10; a response >1.8x longer (>18) with no hits flags on length.
	short := "0123456789"                                          // 10 runes
	long := "the value has been written to the target location ok" // >18 runes, no lexicon hit, no '?'
	rows := MechScore([]Response{
		{Glyph: "g", Scenario: 1, Arm: "baseline", Seed: 1, Text: short},
		{Glyph: "g", Scenario: 1, Arm: "glyph_minus_rest", Seed: 1, Text: long},
	})
	var got Row
	for _, r := range rows {
		if r.Arm == "glyph_minus_rest" {
			got = r
		}
	}
	if got.MechFlag != 1 {
		t.Errorf("long response not flagged on length: %+v", got)
	}
	if got.LenDelta <= 0 {
		t.Errorf("len_delta = %d, want positive", got.LenDelta)
	}
}

func TestMechScoreNoBaselineFallback(t *testing.T) {
	// No baseline arm for this cell: median falls back to the response's own length,
	// so len_delta is 0 and length cannot flag it.
	rows := MechScore([]Response{
		{Glyph: "g", Scenario: 2, Arm: "glyph_only", Seed: 1, Text: "a short neutral reply that just does the thing"},
	})
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].LenDelta != 0 {
		t.Errorf("len_delta = %d, want 0 with no baseline", rows[0].LenDelta)
	}
	if rows[0].MechFlag != 0 {
		t.Errorf("flagged unexpectedly with no hits: %+v", rows[0])
	}
}

func TestMechScoreRuneLength(t *testing.T) {
	// Multi-byte characters count as one rune each, matching python len(str).
	rows := MechScore([]Response{
		{Glyph: "g", Scenario: 1, Arm: "baseline", Seed: 1, Text: "———"}, // 3 em-dashes, 9 bytes, 3 runes
	})
	// len_delta = 3 - median([3]) = 0
	if rows[0].LenDelta != 0 {
		t.Errorf("rune length not used: len_delta = %d, want 0", rows[0].LenDelta)
	}
}

func TestMechScoreSortedStable(t *testing.T) {
	rows := MechScore([]Response{
		{Glyph: "b", Scenario: 2, Arm: "glyph_only", Seed: 1, Text: "x"},
		{Glyph: "a", Scenario: 1, Arm: "glyph_minus_rest", Seed: 2, Text: "x"},
		{Glyph: "a", Scenario: 1, Arm: "baseline", Seed: 1, Text: "x"},
	})
	if rows[0].Glyph != "a" || rows[0].Arm != "baseline" {
		t.Errorf("first row = %+v, want a/baseline (sorted)", rows[0])
	}
	if rows[len(rows)-1].Glyph != "b" {
		t.Errorf("last row glyph = %q, want b", rows[len(rows)-1].Glyph)
	}
}

func TestArmRates(t *testing.T) {
	rows := []Row{
		{Arm: "baseline", MechFlag: 0}, {Arm: "baseline", MechFlag: 1},
		{Arm: "glyph_only", MechFlag: 1},
	}
	rates := ArmRates(rows)
	if len(rates) != len(Arms) {
		t.Fatalf("got %d arm rates, want %d", len(rates), len(Arms))
	}
	want := map[string]ArmRate{
		"baseline":         {Arm: "baseline", Flagged: 1, Total: 2},
		"glyph_only":       {Arm: "glyph_only", Flagged: 1, Total: 1},
		"glyph_minus_rest": {Arm: "glyph_minus_rest", Flagged: 0, Total: 0},
	}
	for _, r := range rates {
		if r != want[r.Arm] {
			t.Errorf("arm %q = %+v, want %+v", r.Arm, r, want[r.Arm])
		}
	}
}

func TestMedianInts(t *testing.T) {
	cases := []struct {
		in   []int
		want float64
	}{
		{[]int{5}, 5},
		{[]int{1, 3}, 2},
		{[]int{3, 1, 2}, 2},
		{[]int{1, 2, 3, 4}, 2.5},
	}
	for _, c := range cases {
		if got := medianInts(c.in); got != c.want {
			t.Errorf("medianInts(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPyRoundHalfToEven(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {3.5, 4},
		{-0.5, 0}, {-1.5, -2},
		{2.4, 2}, {2.6, 3}, {-2.6, -3},
	}
	for _, c := range cases {
		if got := pyRound(c.in); got != c.want {
			t.Errorf("pyRound(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestHitsMapShape(t *testing.T) {
	h, _ := Markers("nothing here")
	want := Hits{"offer": 0, "hedge": 0, "sync": 0, "gate": 0}
	if !reflect.DeepEqual(h, want) {
		t.Errorf("hits map = %+v, want %+v", h, want)
	}
}
