package actionconflict

import "testing"

func TestRid(t *testing.T) {
	got := Rid(Row{Cell: "cpuc-api-version-mistral-base", Condition: "baseline", Run: 1})
	if got != "cpuc-api-version-mistral-base::baseline::1" {
		t.Fatalf("Rid = %q", got)
	}
}

func TestLowconfFiltersLowOnly(t *testing.T) {
	rows := []Row{
		{Cell: "a", Run: 1, Confidence: "high"},
		{Cell: "a", Run: 2, Confidence: "low"},
		{Cell: "a", Run: 3, Confidence: "low"},
	}
	got := Lowconf(rows)
	if len(got) != 2 || got[0].Run != 2 || got[1].Run != 3 {
		t.Fatalf("Lowconf = %+v", got)
	}
}

func TestDetMap(t *testing.T) {
	rows := []Row{{Cell: "c", Condition: "baseline", Run: 5, Verdict: "A_local",
		Confidence: "high", Scenario: "api-version", Precision: "", Model: "m"}}
	m := DetMap(rows)
	d, ok := m["c::baseline::5"]
	if !ok || d.Verdict != "A_local" || d.Scenario != "api-version" || d.Model != "m" {
		t.Fatalf("DetMap = %+v", m)
	}
}

func TestPilotStratifiesAndCaps(t *testing.T) {
	var rows []Row
	// 6 high rows in one stratum, 6 low rows in another.
	for i := 0; i < 6; i++ {
		rows = append(rows, Row{Cell: "c", Condition: "baseline", Run: i,
			Scenario: "api-version", Confidence: "high"})
	}
	for i := 0; i < 6; i++ {
		rows = append(rows, Row{Cell: "c", Condition: "canon_conflict", Precision: "weak", Run: 100 + i,
			Scenario: "api-version", Confidence: "low"})
	}
	noShuffle := func([]Row) {}
	pilot := Pilot(rows, noShuffle)
	// 3 from the high stratum + 4 from the low stratum = 7.
	if len(pilot) != 7 {
		t.Fatalf("pilot size = %d, want 7 (3 high + 4 low)", len(pilot))
	}
	highs, lows := 0, 0
	for _, r := range pilot {
		if r.Confidence == "high" {
			highs++
		} else {
			lows++
		}
	}
	if highs != 3 || lows != 4 {
		t.Fatalf("pilot mix = %d high, %d low; want 3/4", highs, lows)
	}
}

func TestPilotDedupsAndHandlesSmallStrata(t *testing.T) {
	rows := []Row{
		{Cell: "c", Condition: "baseline", Run: 1, Scenario: "s", Confidence: "high"},
		{Cell: "c", Condition: "baseline", Run: 1, Scenario: "s", Confidence: "high"}, // dup id
		{Cell: "c", Condition: "baseline", Run: 2, Scenario: "s", Confidence: "low"},  // only 1 in a low stratum
	}
	pilot := Pilot(rows, func([]Row) {})
	if len(pilot) != 2 {
		t.Fatalf("pilot = %d, want 2 (dedup + a 1-item low stratum)", len(pilot))
	}
}
