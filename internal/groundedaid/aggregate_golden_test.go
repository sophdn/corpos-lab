package groundedaid

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAggregateGolden freezes Aggregate's whole output, for one and two raters,
// over a key that spans classes, models, scenarios and conditions, with ids only
// one rater coded and codes that differ exactly but agree on C-vs-not-C.
func TestAggregateGolden(t *testing.T) {
	key := map[string]KeyEntry{
		"p1": {Cls: Classes[0], Scenario: "1", Model: "mistral", Condition: "baseline", Run: 1},
		"p2": {Cls: Classes[0], Scenario: "1", Model: "mistral", Condition: "baseline", Run: 2},
		"p3": {Cls: Classes[0], Scenario: "2", Model: "phi4", Condition: "ground_only", Run: 1},
		"g1": {Cls: Classes[1], Scenario: "1", Model: "qwen38", Condition: "ground_nonprescriptive", Run: 1},
		"g2": {Cls: Classes[1], Scenario: "2", Model: "qwen38", Condition: "domain_imperative_only", Run: 1},
		"s1": {Cls: Classes[2], Scenario: "1", Model: "phi4", Condition: "baseline", Run: 1},
		"s2": {Cls: Classes[2], Scenario: "2", Model: "mistral", Condition: "ground_only", Run: 1},
	}
	a := map[string]string{"p1": "C", "p2": "Ic", "p3": "C", "g1": "I", "g2": "C", "s1": "N"}
	b := map[string]string{"p1": "C", "p2": "Ii", "p3": "Ic", "g1": "C", "g2": "C", "s2": "C"}
	out := "### one rater\n" + Aggregate(key, a, nil) + "### two raters\n" + Aggregate(key, a, b)

	golden := filepath.Join("testdata", "aggregate.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("Aggregate output drifted from %s:\n%s", golden, out)
	}
}
