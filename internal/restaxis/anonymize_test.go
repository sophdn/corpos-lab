package restaxis

import (
	"math/rand"
	"strings"
	"testing"
)

// identityShuffle is a no-op shuffle: it leaves order untouched so a test can assert
// against the canonical input order.
func identityShuffle(n int, swap func(i, j int)) {}

// reverseShuffle reverses the run, a deterministic non-identity reordering.
func reverseShuffle(n int, swap func(i, j int)) {
	for i := 0; i < n/2; i++ {
		swap(i, n-1-i)
	}
}

func sampleScenarios() []AnonScenario {
	return []AnonScenario{
		{Scenario: 1, Task: "task one", GroundTruth: "gt one", Responses: []AnonResponse{
			{Arm: "baseline", Seed: 1, Text: "resp a1"},
			{Arm: "glyph_only", Seed: 2, Text: "resp a2"},
			{Arm: "glyph_minus_rest", Seed: 3, Text: "resp a3"},
		}},
		{Scenario: 2, Task: "task two", GroundTruth: "gt two", Responses: []AnonResponse{
			{Arm: "baseline", Seed: 1, Text: "resp b1"},
			{Arm: "glyph_only", Seed: 1, Text: "resp b2"},
		}},
	}
}

func TestBuildAnonBundleMappingCount(t *testing.T) {
	bundle := BuildAnonBundle(sampleScenarios(), identityShuffle)
	if len(bundle.Map) != 5 {
		t.Fatalf("map has %d entries, want 5", len(bundle.Map))
	}
}

func TestBuildAnonBundleIDsSequential(t *testing.T) {
	bundle := BuildAnonBundle(sampleScenarios(), identityShuffle)
	for _, want := range []string{"R0000", "R0001", "R0002", "R0003", "R0004"} {
		if _, ok := bundle.Map[want]; !ok {
			t.Errorf("missing id %q; map=%v", want, bundle.Map)
		}
	}
	// Identity shuffle: first scenario's baseline/seed1 is R0000.
	if got := bundle.Map["R0000"]; got != (AnonEntry{Scenario: 1, Arm: "baseline", Seed: 1}) {
		t.Errorf("R0000 = %+v, want scenario 1 baseline seed 1", got)
	}
	if got := bundle.Map["R0003"]; got != (AnonEntry{Scenario: 2, Arm: "baseline", Seed: 1}) {
		t.Errorf("R0003 = %+v, want scenario 2 baseline seed 1", got)
	}
}

func TestBuildAnonBundleBlind(t *testing.T) {
	bundle := BuildAnonBundle(sampleScenarios(), identityShuffle)
	// The rater-facing markdown must not name any arm: the arm lives only in Map.
	for _, arm := range Arms {
		if strings.Contains(bundle.Markdown, arm) {
			t.Errorf("markdown leaks arm %q", arm)
		}
	}
	// Every response text appears exactly once, under an opaque id.
	for _, sc := range sampleScenarios() {
		for _, r := range sc.Responses {
			if n := strings.Count(bundle.Markdown, r.Text); n != 1 {
				t.Errorf("response %q appears %d times, want 1", r.Text, n)
			}
		}
	}
	if n := strings.Count(bundle.Markdown, "### R"); n != 5 {
		t.Errorf("found %d response headers, want 5", n)
	}
}

func TestBuildAnonBundleRoundTrip(t *testing.T) {
	// The Map must de-anonymize each id back to a real (scenario, arm, seed) present
	// in the input, with no duplicates and none invented.
	want := map[AnonEntry]bool{}
	for _, sc := range sampleScenarios() {
		for _, r := range sc.Responses {
			want[AnonEntry{Scenario: sc.Scenario, Arm: r.Arm, Seed: r.Seed}] = true
		}
	}
	bundle := BuildAnonBundle(sampleScenarios(), reverseShuffle)
	got := map[AnonEntry]bool{}
	for _, e := range bundle.Map {
		if got[e] {
			t.Errorf("duplicate entry %+v", e)
		}
		got[e] = true
		if !want[e] {
			t.Errorf("invented entry %+v", e)
		}
	}
	if len(got) != len(want) {
		t.Errorf("round-trip set size %d, want %d", len(got), len(want))
	}
}

func TestBuildAnonBundleShuffleReorders(t *testing.T) {
	id := BuildAnonBundle(sampleScenarios(), identityShuffle)
	rev := BuildAnonBundle(sampleScenarios(), reverseShuffle)
	// Reversing scenario 1's three responses puts glyph_minus_rest/seed3 at R0000.
	if rev.Map["R0000"] == id.Map["R0000"] {
		t.Errorf("reverse shuffle did not change the id->entry assignment")
	}
	if got := rev.Map["R0000"]; got != (AnonEntry{Scenario: 1, Arm: "glyph_minus_rest", Seed: 3}) {
		t.Errorf("reversed R0000 = %+v, want scenario 1 glyph_minus_rest seed 3", got)
	}
}

func TestBuildAnonBundleEmptyScenarioEmitsHeader(t *testing.T) {
	scs := []AnonScenario{{Scenario: 7, Task: "t", GroundTruth: "g"}} // no responses
	bundle := BuildAnonBundle(scs, identityShuffle)
	if len(bundle.Map) != 0 {
		t.Errorf("empty scenario produced %d ids, want 0", len(bundle.Map))
	}
	if !strings.Contains(bundle.Markdown, "## Task group 7") {
		t.Errorf("header for an empty-but-present scenario was dropped")
	}
}

func TestBuildAnonBundleDeterministicGivenSeed(t *testing.T) {
	scs := sampleScenarios()
	one := BuildAnonBundle(scs, rand.New(rand.NewSource(42)).Shuffle)
	two := BuildAnonBundle(scs, rand.New(rand.NewSource(42)).Shuffle)
	if one.Markdown != two.Markdown {
		t.Errorf("same seed produced different markdown")
	}
	for rid, e := range one.Map {
		if two.Map[rid] != e {
			t.Errorf("same seed diverged at %q: %+v vs %+v", rid, e, two.Map[rid])
		}
	}
}
