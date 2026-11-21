// Package baseline runs the unguided-baseline measurement item 12 consumes: the
// class scenario run with NO glyph across the weak local treatment shelf, one
// model at a time. It has two halves that bracket the rater step:
//
//   - RunCapture runs the scenario UNGUIDED against the one model currently
//     loaded on the inference portal and records every reply, keyed by the model
//     the server reports (record what ran). The operating agent swaps the portal
//     to the next shelf model and captures again, so a model that fails is caught
//     as it runs, not at the end of a whole-shelf sweep.
//   - Fold reads the captures and the consensus rater codes and renders the
//     battery.BaselineOutcome item 12 reads through `battery -baseline`.
//
// Scoring is deferred to the rater between the two halves, matching the assay:
// the probe captures conduct, a judge scores it against the rubric. This package
// never asks a model whether the behavior fired; it folds the codes the raters
// returned.
package baseline

import (
	"context"
	"fmt"
	"time"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
	"corpos-lab/internal/rater"
)

// TargetCode is the grounded-probe code that means the class's target
// (avoidance) behavior FIRED unguided: the model produced the correct conduct
// with no glyph in the prompt. Item 12 reads a shelf-wide fire as a trained
// default the glyph merely restates (FAIL); a miss on any model means the glyph
// does real work there (PASS). See internal/battery/baseline.go.
const TargetCode = string(assay.ScoreC)

// CaptureRun is one unguided replicate: its slice id (the key the rater scores
// it under), the reply text under judgment, and the server's account of what
// produced it.
type CaptureRun struct {
	Run       int            `json:"run"`
	SliceID   string         `json:"slice_id"`
	ReplyText string         `json:"reply_text"`
	Observed  assay.Observed `json:"observed"`
}

// Capture is one shelf model's unguided-baseline capture. ModelID is the model
// the server REPORTED (record what ran), not the one requested — a swap behind
// the portal shows up here. An empty reported model is recorded as-is and folds
// to a gap, never a guessed verdict.
type Capture struct {
	Item      string       `json:"item"`
	ModelID   string       `json:"model_id"`
	Version   string       `json:"version"`
	Requested string       `json:"requested_model"`
	RunDate   string       `json:"run_date"`
	Runs      []CaptureRun `json:"runs"`
}

// SliceID is the stable key a run is rated under: item, the model that answered
// (the reported model, falling back to the requested one), and the replicate
// index. Fold reads the code back by this id, so a run scored by a rater maps to
// exactly one (model, run) cell.
func SliceID(item, reported, requested string, run int) string {
	who := reported
	if who == "" {
		who = requested
	}
	if who == "" {
		who = "unknown-model"
	}
	return fmt.Sprintf("%s|%s|run%d", item, who, run)
}

// RunCapture runs the scenario UNGUIDED against the loaded model runs times and
// returns the capture. ModelID takes the first non-empty reported model, so the
// record names what actually answered. It captures conduct and never scores it —
// a judge does that later, between capture and Fold. now supplies the run date;
// a nil now uses the wall clock.
func RunCapture(ctx context.Context, m model.Client, item, scenario string, runs int, samp assay.Sampling, now func() time.Time) (Capture, error) {
	if scenario == "" {
		return Capture{}, fmt.Errorf("baseline: capture needs a scenario")
	}
	if runs < 1 {
		return Capture{}, fmt.Errorf("baseline: capture needs runs >= 1, got %d", runs)
	}
	if now == nil {
		now = time.Now
	}
	rec := Capture{
		Item:      item,
		Requested: m.Name(),
		Version:   m.Version(),
		RunDate:   now().UTC().Format(time.RFC3339),
	}
	for run := 1; run <= runs; run++ {
		row, resp, err := assay.RunProbe(ctx, m, item, assay.Baseline, run, assay.Materials{Scenario: scenario}, samp)
		if err != nil {
			return Capture{}, fmt.Errorf("baseline: capture run %d: %w", run, err)
		}
		if rec.ModelID == "" {
			rec.ModelID = row.Observed.Model
		}
		rec.Runs = append(rec.Runs, CaptureRun{
			Run:       run,
			SliceID:   SliceID(item, row.Observed.Model, rec.Requested, run),
			ReplyText: resp.Text,
			Observed:  row.Observed,
		})
	}
	return rec, nil
}

// SliceLines renders the capture's runs as rater slice lines — one line per
// replicate, id = its SliceID, text = the reply under judgment. The grounded
// rater scores text against the class rubric, so Scenario is left empty.
func (c Capture) SliceLines() []rater.SliceLine {
	lines := make([]rater.SliceLine, 0, len(c.Runs))
	for _, r := range c.Runs {
		lines = append(lines, rater.SliceLine{ID: r.SliceID, Text: r.ReplyText})
	}
	return lines
}
