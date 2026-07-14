// Package assay runs behavioral probes against a local model and scores the
// responses into typed rows. It is the real assay logic the registry-lab
// containers shipped only as placeholder shims. The first assay is the
// grounded-glyph probe, ported from lab-app's studies/grounded_glyph_probe.rs.
package assay

import (
	"context"
	"fmt"

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

// ProbeScore is a grounded-glyph probe rubric code (SCORING_RUBRIC.md). It is
// deliberately NOT battery.Verdict: the battery asks a model a yes/no question
// about a glyph entry and reads back PASS/FAIL, whereas a probe presents a
// behavioral scenario and reads back conduct. The two are different
// measurements and share no scale.
//
// What counts as C is condition-specific — at baseline any execution attempt
// scores C, while the grounded condition demands a correctly formatted entry —
// so a code cannot be derived from response text alone. Scoring is deferred to
// a judge against the rubric; the probe's own job is capture.
type ProbeScore string

// Probe rubric codes. Unscored is the only value RunProbe ever assigns: it
// marks a captured-but-unjudged run, and exists so that "nobody has scored
// this yet" is impossible to confuse with "the agent failed".
const (
	// Unscored: response captured, no rubric code assigned yet.
	Unscored ProbeScore = "unscored"
	// ScoreC: recognition + correct action.
	ScoreC ProbeScore = "C"
	// ScoreIi: recognition, no action.
	ScoreIi ProbeScore = "Ii"
	// ScoreIc: recognition, wrong action.
	ScoreIc ProbeScore = "Ic"
	// ScoreI: no recognition.
	ScoreI ProbeScore = "I"
	// ScoreN: not scoreable (malformed / off-task / harness fault).
	ScoreN ProbeScore = "N"
)

// ScoreRow pairs an evaluated item with its rubric code and a rationale
// pointer — the Go analogue of lab-app-types ScoreGridRow. A row leaves the
// probe Unscored; a judge fills Score in later.
type ScoreRow struct {
	Item      string     `json:"item"`
	Condition Condition  `json:"condition"`
	Run       int        `json:"run"`
	Score     ProbeScore `json:"score"`
	Rationale string     `json:"rationale"`
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

	// No verdict is derived from resp.Text. The probe captures; a judge scores
	// against the rubric. Prefix-matching PASS/FAIL here (as this once did via
	// battery.ParseModelVerdict) scored every behavioral reply a spurious fail,
	// because a probe reply is conduct and never opens with either token.
	rationale := fmt.Sprintf("grounded-glyph-probe:%s:response=%dchars:unscored", cond, len(resp.Text))

	row := ScoreRow{
		Item:      itemID,
		Condition: cond,
		Run:       run,
		Score:     Unscored,
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
