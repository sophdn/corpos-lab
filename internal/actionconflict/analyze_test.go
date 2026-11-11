package actionconflict

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type analyzeCase struct {
	ID      string `json:"id"`
	Verdict string `json:"det_verdict"`
	Conf    string `json:"det_conf"`
	DS      string `json:"ds"`
	DV      string `json:"dv"`
	CL      string `json:"cl"`
	Final   string `json:"final"`
	Source  string `json:"source"`
}

// The parity net for the consensus resolver: real det_map + rater votes paired
// with the reference final verdict, covering deterministic, consensus and split.
func TestResolveFinalMatchesReference(t *testing.T) {
	data, err := os.ReadFile("testdata/analyze_cases.json")
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	var cases []analyzeCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("bad testdata: %v", err)
	}
	sources := map[string]int{}
	for _, c := range cases {
		final, source := resolveFinal(c.Conf, c.Verdict, c.DS, c.DV, c.CL)
		if final != c.Final || source != c.Source {
			t.Errorf("%s: got (%s,%s), want (%s,%s)", c.ID, final, source, c.Final, c.Source)
		}
		sources[source]++
	}
	for _, want := range []string{"deterministic", "consensus", "split"} {
		if sources[want] == 0 {
			t.Errorf("parity sample never exercised source %q", want)
		}
	}
}

func TestResolveFinalTieBreak(t *testing.T) {
	// two agree -> consensus
	if f, s := resolveFinal("low", "", "A_canon", "A_canon", "A_local"); f != "A_canon" || s != "consensus" {
		t.Fatalf("majority: got (%s,%s)", f, s)
	}
	// three-way disagreement -> split
	if f, s := resolveFinal("low", "", "A_local", "A_canon", "neither"); f != "SPLIT" || s != "split" {
		t.Fatalf("3-way: got (%s,%s)", f, s)
	}
	// a majority of MISSING votes is a split, not a consensus on nothing
	if f, s := resolveFinal("low", "", "", "", "A_local"); f != "SPLIT" || s != "split" {
		t.Fatalf("missing-majority: got (%s,%s)", f, s)
	}
	// high confidence ignores the votes entirely
	if f, s := resolveFinal("high", "A_local", "A_canon", "A_canon", "A_canon"); f != "A_local" || s != "deterministic" {
		t.Fatalf("high: got (%s,%s)", f, s)
	}
}

func TestRate(t *testing.T) {
	if r := Rate(map[string]int{"A_canon": 3, "A_local": 1}); r != 75 {
		t.Fatalf("rate = %v, want 75", r)
	}
	if r := Rate(map[string]int{}); r != 0 {
		t.Fatalf("empty rate = %v, want 0", r)
	}
}

func TestOverrideReportSubQuestions(t *testing.T) {
	// Build final rows: strong override, controls that survive.
	final := map[string]FinalRow{}
	add := func(id, cond, prec, verdict string, n int) {
		for i := 0; i < n; i++ {
			final[id+string(rune('a'+i))] = FinalRow{
				DetRow: DetRow{Condition: cond, Precision: prec, Scenario: "api-version", Model: "m"},
				Final:  verdict,
			}
		}
	}
	add("b", "baseline", "", "A_local", 10)
	add("cs", "canon_conflict", "strong", "A_canon", 8) // strong override 80%
	add("cs2", "canon_conflict", "strong", "A_local", 2)
	add("sc", "scrambled_canon", "", "A_canon", 6) // 60% > 40% (strong/2) -> survives
	add("sc2", "scrambled_canon", "", "A_local", 4)
	rep := OverrideReport(final, 0)
	if !strings.Contains(rep, "-> YES") {
		t.Fatalf("q1 should be YES with an 80%% strong override:\n%s", rep)
	}
	if !strings.Contains(rep, "survives (structure)") {
		t.Fatalf("scrambled at 60%% vs strong 80%% should survive:\n%s", rep)
	}
	if !strings.Contains(rep, "3-way splits: 0") {
		t.Fatalf("split count missing:\n%s", rep)
	}

	// No override anywhere -> q1 is NO and the controls collapse.
	flat := map[string]FinalRow{}
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		flat["cc"+id] = FinalRow{DetRow: DetRow{Condition: "canon_conflict", Precision: "strong"}, Final: "A_local"}
		flat["sc"+id] = FinalRow{DetRow: DetRow{Condition: "scrambled_canon"}, Final: "A_local"}
	}
	repNo := OverrideReport(flat, 0)
	if !strings.Contains(repNo, "-> NO") {
		t.Fatalf("q1 should be NO with zero override:\n%s", repNo)
	}
	if !strings.Contains(repNo, "collapses (needs comprehension)") {
		t.Fatalf("scrambled at 0%% should collapse:\n%s", repNo)
	}
}

func TestFinalVerdictBuildsSplits(t *testing.T) {
	det := map[string]DetRow{
		"a": {Verdict: "A_local", Confidence: "high"},
		"b": {Confidence: "low"},
	}
	final, splits := FinalVerdict(det, map[string]string{"b": "A_local"},
		map[string]string{"b": "A_canon"}, map[string]string{"b": "neither"})
	if final["a"].Source != "deterministic" || final["a"].Final != "A_local" {
		t.Fatalf("a = %+v", final["a"])
	}
	if final["b"].Source != "split" || len(splits) != 1 || splits[0].ID != "b" {
		t.Fatalf("b = %+v splits=%+v", final["b"], splits)
	}
}
