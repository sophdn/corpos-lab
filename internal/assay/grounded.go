// Package assay runs behavioral probes against a local model and scores the
// responses into typed rows. It is the real assay logic the registry-lab
// containers shipped only as placeholder shims. The first assay is the
// grounded-glyph probe, ported from lab-app's studies/grounded_glyph_probe.rs.
package assay

import (
	"context"
	"fmt"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/model"
)

// Condition is the aid a grounded-glyph probe run is given.
type Condition string

// Grounded-glyph probe conditions. Values are the wire strings used in
// study.json and response filenames (parity with the legacy run.py flags).
const (
	// Baseline: bare scenario, no aid.
	Baseline Condition = "baseline"
	// GlyphOnly: universal glyph prepended, no ground.
	GlyphOnly Condition = "glyph_only"
	// GroundedGlyph: universal glyph + domain ground prepended.
	GroundedGlyph Condition = "grounded_glyph"
)

// Materials are the text inputs a probe assembles a prompt from. Glyph and
// Ground may be empty for conditions that don't use them.
type Materials struct {
	Scenario string
	Glyph    string
	Ground   string
}

// AssemblePrompt builds the probe prompt for a condition. The "\n---\n"
// delimiter matches the legacy blueprint separator verbatim so Mistral/Claude
// prompt formats stay compatible:
//
//	baseline       → scenario
//	glyph_only     → glyph "---" scenario
//	grounded_glyph → glyph "---" ground "---" scenario
//
// It returns an error when a condition's required material is missing, rather
// than silently emitting a malformed prompt.
func AssemblePrompt(cond Condition, m Materials) (string, error) {
	switch cond {
	case Baseline:
		return m.Scenario, nil
	case GlyphOnly:
		if m.Glyph == "" {
			return "", fmt.Errorf("assay: %s condition requires a glyph", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Glyph, m.Scenario), nil
	case GroundedGlyph:
		if m.Glyph == "" {
			return "", fmt.Errorf("assay: %s condition requires a glyph", cond)
		}
		if m.Ground == "" {
			return "", fmt.Errorf("assay: %s condition requires a ground", cond)
		}
		return fmt.Sprintf("%s\n---\n%s\n---\n%s", m.Glyph, m.Ground, m.Scenario), nil
	default:
		return "", fmt.Errorf("assay: unknown condition %q", cond)
	}
}

// ScoreRow pairs an evaluated item with its typed verdict and a rationale
// pointer — the Go analogue of lab-app-types ScoreGridRow.
type ScoreRow struct {
	Item      string          `json:"item"`
	Condition Condition       `json:"condition"`
	Run       int             `json:"run"`
	Verdict   battery.Verdict `json:"verdict"`
	Rationale string          `json:"rationale"`
}

// ProbeResponse is the raw model reply for one condition/run, retained for
// audit alongside the scored row.
type ProbeResponse struct {
	Condition Condition
	Run       int
	Prompt    string
	Text      string
}

// probeGenParams pins the grounded-glyph probe sampling: deterministic,
// 512-token cap (verbatim from the Rust runner).
func probeGenParams() model.GenParams {
	return model.GenParams{
		Temperature: model.Float64(0.0),
		MaxTokens:   model.Int(512),
	}
}

// RunProbe assembles the prompt for one condition, calls the model once, and
// scores the reply. It returns both the scored row and the raw response so
// the caller can persist the response for audit. run is the 1-based replicate
// index, recorded on the row and used for the response filename.
func RunProbe(ctx context.Context, m model.Client, itemID string, cond Condition, run int, mats Materials) (ScoreRow, ProbeResponse, error) {
	prompt, err := AssemblePrompt(cond, mats)
	if err != nil {
		return ScoreRow{}, ProbeResponse{}, err
	}

	resp, err := m.Generate(ctx, prompt, probeGenParams())
	if err != nil {
		return ScoreRow{}, ProbeResponse{}, fmt.Errorf("assay: %s run %d: %w", cond, run, err)
	}

	verdict := battery.ParseModelVerdict(resp.Text, 1)
	rationale := fmt.Sprintf("grounded-glyph-probe:%s:response=%dchars", cond, len(resp.Text))

	row := ScoreRow{
		Item:      itemID,
		Condition: cond,
		Run:       run,
		Verdict:   verdict,
		Rationale: rationale,
	}
	response := ProbeResponse{
		Condition: cond,
		Run:       run,
		Prompt:    prompt,
		Text:      resp.Text,
	}
	return row, response, nil
}
