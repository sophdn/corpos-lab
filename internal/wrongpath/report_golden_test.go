package wrongpath

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReportGolden freezes the full Report over a deterministic grid: every arm,
// model, glyph and polarity with eight runs mixing verdicts, field sources and
// reasoning traces. A few cells are left unscored (no results.json) so H2 skips
// them. A second report over the same runs with no cells covers the all-NaN case.
func TestReportGolden(t *testing.T) {
	cells, all := reportGrid()
	out := Report(cells, all) + "### no cells\n" + Report(map[CellKey]CellScore{}, all)

	golden := filepath.Join("testdata", "report.golden")
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
		t.Fatalf("Report drifted from %s:\n%s", golden, out)
	}
}

// reportGrid builds TestReportGolden's runs and scored cells.
func reportGrid() (map[CellKey]CellScore, []RunFile) {
	verdicts := []string{"VERDICT: yes", "VERDICT: no", "verdict: YES", ""}
	fields := []string{"FIELD SOURCE: Scope block", "FIELD SOURCE: the marker and scope", "FIELD SOURCE: aim", "FIELD SOURCE: pull", "FIELD SOURCE: rest", "FIELD SOURCE: something else", ""}
	traces := []string{"", "checked the operative clause", "thought about scope", "no mention"}
	var all []RunFile
	cells := map[CellKey]CellScore{}
	n := 0
	for ai, arm := range Arms {
		for _, model := range Models {
			for gi, glyph := range Glyphs {
				for _, ab := range AB {
					var texts []string
					for run := 0; run < 8; run++ {
						v := verdicts[(n*7+run*(ai+1)+gi*run)%4]
						if glyph == "cas" && ab == "b" {
							v = "" // no verdict at all: a low-accuracy, uninformative cell
						}
						text := v + "\n" + fields[(n*3+run)%7]
						tr := traces[(n+run)%4]
						all = append(all, RunFile{Arm: arm, Model: model, Glyph: glyph, AB: ab, Text: text, Reasoning: tr, HasReasoning: tr != ""})
						texts = append(texts, text)
					}
					if (gi+n)%5 != 0 {
						cells[CellKey{arm, model, glyph, ab}] = ScoreCell(texts, GT[ab])
					}
					n++
				}
			}
		}
	}
	return cells, all
}
