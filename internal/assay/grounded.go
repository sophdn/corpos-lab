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
	// Observed is what the server reported while producing THIS row.
	Observed Observed `json:"observed"`
}

// Observed is the server's own account of a single generation, recorded per
// row rather than per run.
//
// Per-row is the point. A study's substrate was assumed constant for its whole
// life and went unrecorded; when one leg silently fell back to CPU, nothing in
// the data could show it. TokensPerSecond makes that visible in every cell —
// ~40-60 on GPU for a 7B, 5.1 on CPU — so a mid-study substrate change shows up
// as a step in the rows instead of never showing up at all.
type Observed struct {
	// Model is what the server said it used, which may differ from what the
	// study asked for.
	Model string `json:"model"`
	// BuildInfo is llama.cpp's build id, so a server upgrade mid-study is
	// visible in the rows either side of it.
	BuildInfo string `json:"build_info"`
	// TokensPerSecond is generation throughput — the substrate tripwire.
	TokensPerSecond float64 `json:"tokens_per_second"`
	PromptTokens    int     `json:"prompt_tokens"`
	PredictedTokens int     `json:"predicted_tokens"`
}

// ProbeResponse is the raw model reply for one condition/run, retained for
// audit alongside the scored row.
type ProbeResponse struct {
	Condition Condition
	Run       int
	Prompt    string
	Text      string
}

// Sampling is a study's declared sampling regime for the probe. It lives in
// the study definition rather than in this package so that a run records the
// sampler it actually used: declared here, it travels inside study.json and is
// captured with the run, instead of being a Go constant invisible to the
// record. (The original rationale cited CHARTER freeze-by-digest, retired
// 2026-07-14 — see INQUIRY.md. The placement outlives the rationale: what made
// it right was never the freeze, it was that a hardcoded sampler is a sampler
// nobody can see afterwards.)
//
// This is also what makes a graded grid possible. Greedy decoding
// (temperature 0) returns an identical reply for every replicate, so a cell
// can only ever score 0/8 or 8/8 — never a graded value like the casg-direct
// v3 target's 7/8. Run-to-run variation requires sampling; Seeds keep that
// variation reproducible.
// The regime is COMPLETE: every stage of llama.cpp's sampler chain is named
// here, not just temperature. Temperature alone does not define a sampling
// distribution — top_k, top_p, min_p, the penalties and the rest all shape it,
// and any one left unsaid is inherited from whatever binary is running and
// changes silently when that binary is upgraded.
type Sampling struct {
	Temperature float64 `json:"temperature"`
	// Seeds pins one RNG seed per replicate, indexed by 1-based run number.
	// Empty means deterministic single-shot decoding, which is only coherent
	// at temperature 0.
	Seeds     []int `json:"seeds,omitempty"`
	MaxTokens int   `json:"max_tokens"`

	// Truncation stages, in chain order.
	TopNSigma float64 `json:"top_n_sigma"`
	TopK      int     `json:"top_k"`
	TypicalP  float64 `json:"typical_p"`
	TopP      float64 `json:"top_p"`
	MinP      float64 `json:"min_p"`

	// Penalty stages.
	RepeatPenalty    float64 `json:"repeat_penalty"`
	RepeatLastN      int     `json:"repeat_last_n"`
	PresencePenalty  float64 `json:"presence_penalty"`
	FrequencyPenalty float64 `json:"frequency_penalty"`

	// Stage switches. Sub-parameters are inert while these are zero.
	XTCProbability float64 `json:"xtc_probability"`
	DryMultiplier  float64 `json:"dry_multiplier"`
}

// Deterministic reports whether this regime decodes greedily, in which case
// every replicate of a cell is identical by construction.
func (s Sampling) Deterministic() bool { return s.Temperature == 0 }

// GenParamsForRun builds the generation params for a 1-based replicate index,
// selecting that run's seed from the sequence. A run outside the sequence gets
// no seed rather than a wrong one — silently reusing seed 1 would make two
// replicates identical and quietly deflate a cell's variance.
// Every field is sent, none left to inherit — that is what makes the params
// recorded with the run also the params the run executed under.
func (s Sampling) GenParamsForRun(run int) model.GenParams {
	p := model.GenParams{
		Temperature: model.Float64(s.Temperature),
		MaxTokens:   model.Int(s.MaxTokens),

		TopNSigma: model.Float64(s.TopNSigma),
		TopK:      model.Int(s.TopK),
		TypicalP:  model.Float64(s.TypicalP),
		TopP:      model.Float64(s.TopP),
		MinP:      model.Float64(s.MinP),

		RepeatPenalty:    model.Float64(s.RepeatPenalty),
		RepeatLastN:      model.Int(s.RepeatLastN),
		PresencePenalty:  model.Float64(s.PresencePenalty),
		FrequencyPenalty: model.Float64(s.FrequencyPenalty),

		XTCProbability: model.Float64(s.XTCProbability),
		DryMultiplier:  model.Float64(s.DryMultiplier),
	}
	if run >= 1 && run <= len(s.Seeds) {
		p.Seed = model.Int(s.Seeds[run-1])
	}
	return p
}

// RunProbe assembles the prompt for one condition, calls the model once, and
// captures the reply. It returns both the row and the raw response so the
// caller can persist the response for audit. run is the 1-based replicate
// index: it is recorded on the row, used for the response filename, and
// selects this run's seed from samp.
func RunProbe(ctx context.Context, m model.Client, itemID string, cond Condition, run int, mats Materials, samp Sampling) (ScoreRow, ProbeResponse, error) {
	prompt, err := AssemblePrompt(cond, mats)
	if err != nil {
		return ScoreRow{}, ProbeResponse{}, err
	}

	resp, err := m.Generate(ctx, prompt, samp.GenParamsForRun(run))
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
		Observed: Observed{
			Model:           resp.Model,
			BuildInfo:       resp.SystemFingerprint,
			TokensPerSecond: resp.Timings.PredictedPerSecond,
			PromptTokens:    resp.Timings.PromptN,
			PredictedTokens: resp.Timings.PredictedN,
		},
	}
	response := ProbeResponse{
		Condition: cond,
		Run:       run,
		Prompt:    prompt,
		Text:      resp.Text,
	}
	return row, response, nil
}
