package neutralprefix

import (
	"errors"
	"fmt"
	"testing"
)

// syntheticPop builds a casg-direct scenario-1 population with 8 runs for every
// (model, condition) the anchor samples from, all present in the Claude map.
func syntheticPop() (map[string]KeyMeta, map[string]string) {
	const runsPer = 8
	km := map[string]KeyMeta{}
	claude := map[string]string{}
	codes := []string{"C", "Ii", "Ic", "I", "N"}
	n := 0
	for _, model := range []string{"mistral", "phi4", "qwen38"} {
		for _, cond := range []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"} {
			for run := 1; run <= runsPer; run++ {
				rid := fmt.Sprintf("r_%s_%s_%d", model, cond, run)
				km[rid] = KeyMeta{Cls: "casg-direct", Scenario: "1", Model: model, Condition: cond, Run: run}
				claude[rid] = codes[n%len(codes)]
				n++
			}
		}
	}
	// Noise the sampler must ignore: other class, other scenario, not in Claude.
	km["x_other_class"] = KeyMeta{Cls: "parent-state-check-bypass", Scenario: "1", Model: "mistral", Condition: "baseline", Run: 1}
	claude["x_other_class"] = "C"
	km["x_other_scen"] = KeyMeta{Cls: "casg-direct", Scenario: "2", Model: "mistral", Condition: "baseline", Run: 1}
	claude["x_other_scen"] = "C"
	km["x_unscored"] = KeyMeta{Cls: "casg-direct", Scenario: "1", Model: "mistral", Condition: "baseline", Run: 99}
	return km, claude
}

func TestBuildAnchorFullSample(t *testing.T) {
	km, claude := syntheticPop()
	resp := func(model, cond string, run int) (string, error) {
		return fmt.Sprintf("resp-%s-%s-%d", model, cond, run), nil
	}
	res, err := BuildAnchor(km, claude, "  scenario text  ", resp)
	if err != nil {
		t.Fatal(err)
	}
	// 6+6+6+6 + 3+3+3+2 + 2+2 = 39
	if len(res.Items) != 39 || res.Held.Len() != 39 {
		t.Fatalf("items=%d held=%d, want 39", len(res.Items), res.Held.Len())
	}
	// hidden ids are H01..H39 in order.
	for i, k := range res.Held.keys {
		if want := fmt.Sprintf("H%02d", i+1); k != want {
			t.Fatalf("held key %d = %q, want %q", i, k, want)
		}
	}
	// strata counts sum to 39 and match the weighting.
	stratWant := map[string]int{
		"mistral|neutral_prefix": 6, "mistral|glyph_only": 6,
		"phi4|neutral_prefix": 6, "phi4|glyph_only": 6,
		"mistral|baseline": 3, "mistral|imperative_only": 3,
		"phi4|baseline": 3, "phi4|imperative_only": 2,
		"qwen38|neutral_prefix": 2, "qwen38|glyph_only": 2,
	}
	total := 0
	for _, k := range res.Strata.Keys {
		if res.Strata.Get(k) != stratWant[k] {
			t.Errorf("stratum %s = %d, want %d", k, res.Strata.Get(k), stratWant[k])
		}
		total += res.Strata.Get(k)
	}
	if total != 39 {
		t.Fatalf("strata total %d, want 39", total)
	}
	mix := 0
	for _, k := range res.ClaudeMix.Keys {
		mix += res.ClaudeMix.Get(k)
	}
	if mix != 39 {
		t.Fatalf("claude mix total %d, want 39", mix)
	}
	// scenario text is trimmed and shared across items; ground block is constant.
	first := res.Items[0].(*OMap)
	blocks := first.vals[1].([]any)
	b0 := blocks[0].(*OMap)
	if b0.vals[2].(string) != "scenario text" {
		t.Errorf("scenario not trimmed: %q", b0.vals[2])
	}
	if blocks[1].(*OMap).vals[2].(string) != CGround {
		t.Errorf("ground block mismatch")
	}
}

func TestBuildAnchorDeterministic(t *testing.T) {
	km, claude := syntheticPop()
	resp := func(model, cond string, run int) (string, error) { return "r", nil }
	a, err := BuildAnchor(km, claude, "s", resp)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildAnchor(km, claude, "s", resp)
	if err != nil {
		t.Fatal(err)
	}
	if PyDumps(a.Items, 1, false) != PyDumps(b.Items, 1, false) ||
		PyDumps(a.Held, 1, true) != PyDumps(b.Held, 1, true) {
		t.Fatal("BuildAnchor not deterministic")
	}
}

func TestBuildAnchorCapsShortStratum(t *testing.T) {
	km, claude := syntheticPop()
	// Remove all but one qwen38 neutral_prefix run so take(k=2) caps to 1.
	for run := 2; run <= 8; run++ {
		delete(km, fmt.Sprintf("r_qwen38_neutral_prefix_%d", run))
	}
	resp := func(model, cond string, run int) (string, error) { return "r", nil }
	res, err := BuildAnchor(km, claude, "s", resp)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 38 {
		t.Fatalf("items=%d, want 38 after capping", len(res.Items))
	}
}

func TestBuildAnchorRespError(t *testing.T) {
	km, claude := syntheticPop()
	sentinel := errors.New("boom")
	resp := func(model, cond string, run int) (string, error) { return "", sentinel }
	if _, err := BuildAnchor(km, claude, "s", resp); !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
}
