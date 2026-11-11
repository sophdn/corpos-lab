package actionconflict

import (
	"fmt"
	"strings"
)

// Row is one scored response from per_response.jsonl.
type Row struct {
	Cell       string `json:"cell"`
	Scenario   string `json:"scenario"`
	Condition  string `json:"condition"`
	Precision  string `json:"precision"`
	Run        int    `json:"run"`
	Confidence string `json:"confidence"`
	Verdict    string `json:"verdict"`
	Model      string `json:"model"`
}

// Rid is a response's stable id: cell::condition::run.
func Rid(r Row) string {
	return fmt.Sprintf("%s::%s::%d", r.Cell, r.Condition, r.Run)
}

// Lowconf returns every low-confidence row — the ones a cross-family rater must
// resolve.
func Lowconf(rows []Row) []Row {
	var out []Row
	for _, r := range rows {
		if r.Confidence == "low" {
			out = append(out, r)
		}
	}
	return out
}

// DetMap keys each row's deterministic score by its id, for later comparison
// against the rater verdicts.
func DetMap(rows []Row) map[string]DetRow {
	m := make(map[string]DetRow, len(rows))
	for _, r := range rows {
		m[Rid(r)] = DetRow{
			Verdict: r.Verdict, Confidence: r.Confidence, Scenario: r.Scenario,
			Condition: r.Condition, Precision: r.Precision, Model: r.Model,
		}
	}
	return m
}

func strataKey(r Row) string {
	return r.Scenario + "|" + condKey(r.Condition, r.Precision) + "|" + r.Confidence
}

// Pilot builds a stratified sample: 3 high-confidence and 4 low-confidence rows
// per (scenario, condition, confidence) stratum, deduped by id in stratum order.
//
// The reference shuffled each stratum with a seeded Python RNG to choose which
// rows fill it. Go's RNG is not byte-compatible with Python's, so this does NOT
// reproduce the committed pilot.jsonl: shuffle draws a FRESH sample of the same
// shape. Chain 543 is closed, so the committed slice stands as the historical
// record; a regeneration is a new draw. shuffle is injected so the selection is
// controllable in a test and random in production.
func Pilot(rows []Row, shuffle func([]Row)) []Row {
	strata := map[string][]Row{}
	var order []string
	for _, r := range rows {
		k := strataKey(r)
		if _, ok := strata[k]; !ok {
			order = append(order, k)
		}
		strata[k] = append(strata[k], r)
	}
	var pilot []Row
	seen := map[string]bool{}
	for _, k := range order {
		items := strata[k]
		take := 3
		if strings.HasSuffix(k, "|low") {
			take = 4 // oversample low-confidence
		}
		shuffle(items)
		for i := 0; i < take && i < len(items); i++ {
			id := Rid(items[i])
			if seen[id] {
				continue
			}
			seen[id] = true
			pilot = append(pilot, items[i])
		}
	}
	return pilot
}
