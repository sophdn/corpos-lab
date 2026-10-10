package runner

import (
	"context"
	"errors"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

func cellSpec() StudySpec {
	return StudySpec{ItemID: "it", Conditions: []assay.Condition{assay.Baseline, assay.GlyphOnly}, RunsPerCell: 3}
}

// Rows come back in cell order: conditions in study order, runs ascending.
func TestForEachCellKeepsCellOrder(t *testing.T) {
	var seen []string
	rows, err := forEachCell(context.Background(), cellSpec(), func(_ context.Context, cond assay.Condition, run int) (assay.ScoreRow, error) {
		seen = append(seen, string(cond))
		return assay.ScoreRow{Item: "it", Condition: cond, Run: run}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 6 || rows[0].Condition != assay.Baseline || rows[0].Run != 1 || rows[5].Condition != assay.GlyphOnly || rows[5].Run != 3 {
		t.Errorf("rows = %+v", rows)
	}
	if len(seen) != 6 {
		t.Errorf("ran %d cells, want 6", len(seen))
	}
}

// A failing cell stops the study at that cell.
func TestForEachCellStopsAtFirstError(t *testing.T) {
	boom := errors.New("boom")
	ran := 0
	_, err := forEachCell(context.Background(), cellSpec(), func(_ context.Context, cond assay.Condition, run int) (assay.ScoreRow, error) {
		ran++
		if run == 2 {
			return assay.ScoreRow{}, boom
		}
		return assay.ScoreRow{Item: "it"}, nil
	})
	if !errors.Is(err, boom) || ran != 2 {
		t.Errorf("err = %v after %d cells, want boom after 2", err, ran)
	}
}

// The raw /completion path reports no build id per response, so every row's
// build_info was empty while the paper tables asserted a build. Rows now take
// the build from the /props readback and say so; a row that already carries
// one keeps it.
func TestExecuteFillsRowBuildInfoFromProps(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND",
	})
	client := &fakeClient{text: "PASS", props: model.ServerProps{BuildInfo: "b9445-af6528e6d"}}
	results, err := Execute(context.Background(), in, t.TempDir(), client)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results.Rows {
		if r.Observed.BuildInfo != "b9445-af6528e6d" || r.Observed.BuildInfoSource != "props" {
			t.Fatalf("row build = %q (%s), want the props build", r.Observed.BuildInfo, r.Observed.BuildInfoSource)
		}
	}

	served := &fakeClient{resp: &model.Response{Text: "PASS", SystemFingerprint: "b1-own"}, props: model.ServerProps{BuildInfo: "b2-props"}}
	results, err = Execute(context.Background(), in, t.TempDir(), served)
	if err != nil {
		t.Fatal(err)
	}
	if r := results.Rows[0].Observed; r.BuildInfo != "b1-own" || r.BuildInfoSource != "" {
		t.Errorf("row with its own build = %q (%s), want it kept", r.BuildInfo, r.BuildInfoSource)
	}
}
