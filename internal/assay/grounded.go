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
	// ImperativeOnly: an information-matched imperative rule prepended, no
	// glyph. This is the matched-content experiment's T2 condition — the content
	// control against GlyphOnly (T1): same propositions, directive format
	// instead of the three-axis glyph. It carries no glyph on purpose; T1 and T2
	// are the two arms of the contrast, never combined.
	ImperativeOnly Condition = "imperative_only"
	// ScrambledGlyph: a glyph with the three-axis shape and length preserved but
	// the content word-scrambled into incoherence, prepended in place of the real
	// glyph. Mechanism control: read against GlyphOnly it separates format-structure
	// (if the scramble reproduces the glyph's effect) from content-comprehension
	// (if it collapses toward baseline).
	ScrambledGlyph Condition = "scrambled_glyph"
	// OffTargetGlyph: a coherent glyph for a DIFFERENT decision class than the
	// scenario, prepended in place of the matching glyph. Mechanism control: read
	// against GlyphOnly it tests whether recognition of the scenario is in the loop
	// (equal effect on- and off-target implicates glyph structure over recognition).
	OffTargetGlyph Condition = "off_target_glyph"
	// NeutralPrefix: a neutral, non-glyph, length-matched prose block prepended in
	// place of the glyph. The block is abstract, task-irrelevant prose — no decision
	// content, no domain vocabulary, and no three-axis shape — matched to the glyph's
	// length. Mechanism control for the generic-prefix rival: read against GlyphOnly
	// it separates "a prefix of this length sits before the scenario" from "the prefix
	// describes the decision the scenario poses". If it reproduces the glyph's
	// suppression the effect is a bare prefix or length effect; if it collapses toward
	// baseline the suppression needs decision content. It is the control the scrambled
	// and off-target arms cannot be, because both keep the three-axis glyph shape.
	NeutralPrefix Condition = "neutral_prefix"
	// GlyphMinusRest: the glyph with its Rest axis block removed and nothing else
	// changed, prepended in place of the full glyph. The rest-axis overhead
	// benchmark's ablation arm (Q3): read against GlyphOnly on a NEUTRAL scenario it
	// isolates the Rest axis — whether naming where the decision class does not apply
	// reduces false-positive over-firing when the glyph is loaded but not live.
	GlyphMinusRest Condition = "glyph_minus_rest"
	// DutyOnly: a duty specification prepended, no glyph. The behavioral-equivalence
	// assay's Condition A — the investigation duty (role file) delivered as the sole
	// guidance. It takes the same guidance slot as GlyphOnly, with the same delimiter
	// and shape; the two conditions never combine (a run is duty-only or glyph-only,
	// never both).
	DutyOnly Condition = "duty_only"
	// CorpusOnly: a corpus prepended, no glyph. The behavioral-equivalence assay's
	// Condition B — the same behavioral guidance delivered as a corpus of named
	// patterns instead of a duty specification. Read against DutyOnly it is the
	// behavioral-equivalence contrast: do two delivery routes for the same guidance
	// produce equivalent conduct? Baseline is the assay's Condition C (brief only).
	CorpusOnly Condition = "corpus_only"
	// AnnotatedInstrument: the annotated duty-writing instrument prepended as the
	// guidance. The cartographer-duty-format assay's Condition A — the design
	// instrument that directs the writer to cite taboo slugs per gate and explain
	// failure modes inline, so the produced duty carries the canon as a runtime
	// dependency. It takes the same guidance slot and delimiter as DutyOnly.
	AnnotatedInstrument Condition = "annotated_instrument"
	// CartographerInstrument: the cartographer duty-writing instrument prepended
	// as the guidance. The assay's Condition B — the design instrument that directs
	// the writer to encode avoidance structurally from first-principles failure-mode
	// analysis and cite no slugs. Read against AnnotatedInstrument it is the
	// cartographer-format contrast: does the structural form still cover the
	// corpus-empirical taboos the slug citations transmit?
	CartographerInstrument Condition = "cartographer_instrument"
	// CartographerScanInstrument: the cartographer instrument augmented with an
	// explicit Phase 3 meta-taboo scan — a named check for routing-class and
	// session-close-obligation-class taboos that first-principles defect analysis
	// does not surface. The assay's Condition D, testing whether the proposed fix
	// recovers corpus-empirical coverage without reintroducing slug citations. The
	// original study specified this scan but never ran it.
	CartographerScanInstrument Condition = "cartographer_scan_instrument"
	// GroundOnly: the informative domain-specific ground presented alone, no
	// glyph — the descriptive/domain-specific cell of the form × grounding 2×2.
	// Read against GroundedGlyph it isolates the glyph's contribution over the
	// ground it sits on; it reuses the existing Ground material and takes the
	// guidance slot with the same delimiter and shape as the other -only arms.
	GroundOnly Condition = "ground_only"
	// DomainImperativeOnly: a domain-specific directive presented alone — the
	// directive/domain-specific cell of the form × grounding 2×2. It carries a
	// domain-specific imperative in the guidance slot, distinct from the
	// information-matched ImperativeOnly rule, with the same delimiter and shape.
	DomainImperativeOnly Condition = "domain_imperative_only"
	// GroundNonPrescriptive: the domain-specific ground with its outcome
	// sentences removed — it names the file and the situation but not the correct
	// action or end state. Read against GroundOnly it isolates target
	// specification (naming the outcome) from grounding (naming the domain),
	// answering CaPC review item 1.3. It carries its own material in the guidance
	// slot with the same delimiter and shape as the other -only arms.
	GroundNonPrescriptive Condition = "ground_nonprescriptive"
	// CanonAligned prepends a block that AGREES with the scenario's local
	// instruction — it names the same action, A_local. The content-priority-
	// under-conflict assay's alignment control (chain 543): a prepended block that
	// points the same way must not move the action, so read against CanonConflict
	// it separates "content overrides the instruction" from "any prepended block
	// moves the action".
	CanonAligned Condition = "canon_aligned"
	// CanonConflict prepends a block (the "canon") that CONFLICTS with the
	// scenario's local instruction — it states a rule whose concrete action,
	// A_canon, differs from the instruction's A_local. The assay's treatment
	// condition: the override rate is the share of responses taking A_canon. The
	// precision ladder (weak/medium/strong) is three canon_conflict material
	// files, not three conditions.
	CanonConflict Condition = "canon_conflict"
	// ScrambledCanon prepends the strong conflict block with its shape and length
	// kept but its content word-scrambled into incoherence. Mechanism control:
	// read against CanonConflict it separates prepended structure (if the scramble
	// still overrides) from comprehension (if it collapses toward baseline).
	ScrambledCanon Condition = "scrambled_canon"
	// OffTargetCanon prepends a coherent conflicting block about a DIFFERENT domain
	// than the scenario. Mechanism control: read against CanonConflict it tests
	// whether the model must recognise the block as relevant to this decision
	// (equal override on- and off-target implicates prepended structure over
	// recognition).
	OffTargetCanon Condition = "off_target_canon"
)

// Materials are the text inputs a probe assembles a prompt from. Glyph, Ground,
// and Imperative may be empty for conditions that don't use them.
type Materials struct {
	Scenario string
	Glyph    string
	Ground   string
	// Imperative is the matched imperative rule for the ImperativeOnly (T2)
	// condition.
	Imperative string
	// Scrambled is the shape-preserved, content-scrambled glyph for the
	// ScrambledGlyph control condition.
	Scrambled string
	// OffTarget is the coherent-but-off-class glyph for the OffTargetGlyph
	// control condition.
	OffTarget string
	// Neutral is the neutral, non-glyph, length-matched prose block for the
	// NeutralPrefix control condition.
	Neutral string
	// GlyphMinusRest is the glyph with its Rest axis block removed, for the
	// GlyphMinusRest ablation condition of the rest-axis overhead benchmark.
	GlyphMinusRest string
	// Duty is the duty specification for the DutyOnly (Condition A) condition of
	// the behavioral-equivalence assay.
	Duty string
	// Corpus is the corpus of named behavioral patterns for the CorpusOnly
	// (Condition B) condition of the behavioral-equivalence assay.
	Corpus string
	// Annotated is the annotated duty-writing instrument for the
	// AnnotatedInstrument condition of the cartographer-duty-format assay.
	Annotated string
	// Cartographer is the cartographer duty-writing instrument for the
	// CartographerInstrument condition of the cartographer-duty-format assay.
	Cartographer string
	// CartographerScan is the cartographer instrument plus the Phase 3 meta-taboo
	// scan for the CartographerScanInstrument condition.
	CartographerScan string
	// DomainImperative is the domain-specific directive for the
	// DomainImperativeOnly condition of the form × grounding 2×2.
	DomainImperative string
	// NonPrescriptiveGround is the domain ground with its outcome sentences
	// removed, for the GroundNonPrescriptive condition — it names the domain but
	// not the correct action or end state.
	NonPrescriptiveGround string
	// CanonAligned is the prepended block agreeing with the local instruction
	// (names A_local), for the CanonAligned condition.
	CanonAligned string
	// CanonConflict is the prepended canon block pointing at A_canon, for the
	// CanonConflict condition. One of the weak/medium/strong precision files.
	CanonConflict string
	// ScrambledCanon is the shape-kept, content-scrambled conflict block, for the
	// ScrambledCanon control condition.
	ScrambledCanon string
	// OffTargetCanon is the coherent-but-off-domain conflict block, for the
	// OffTargetCanon control condition.
	OffTargetCanon string
}

// AssemblePrompt builds the probe prompt for a condition. The "\n---\n"
// delimiter matches the legacy blueprint separator verbatim so Mistral/Claude
// prompt formats stay compatible:
//
//	baseline        → scenario
//	glyph_only      → glyph "---" scenario
//	grounded_glyph  → glyph "---" ground "---" scenario
//	imperative_only → imperative "---" scenario
//	scrambled_glyph → scrambled "---" scenario
//	off_target_glyph → off-target glyph "---" scenario
//	neutral_prefix → neutral prose "---" scenario
//	glyph_minus_rest → glyph-minus-rest "---" scenario
//	duty_only       → duty "---" scenario
//	corpus_only     → corpus "---" scenario
//	annotated_instrument         → annotated "---" scenario
//	cartographer_instrument      → cartographer "---" scenario
//	cartographer_scan_instrument → cartographer-scan "---" scenario
//	ground_only            → ground "---" scenario
//	domain_imperative_only → domain-imperative "---" scenario
//	ground_nonprescriptive → non-prescriptive ground "---" scenario
//	canon_aligned    → canon-aligned block "---" scenario
//	canon_conflict   → canon-conflict block "---" scenario
//	scrambled_canon  → scrambled-canon block "---" scenario
//	off_target_canon → off-target-canon block "---" scenario
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
	case ImperativeOnly:
		if m.Imperative == "" {
			return "", fmt.Errorf("assay: %s condition requires an imperative", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Imperative, m.Scenario), nil
	case ScrambledGlyph:
		if m.Scrambled == "" {
			return "", fmt.Errorf("assay: %s condition requires a scrambled glyph", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Scrambled, m.Scenario), nil
	case OffTargetGlyph:
		if m.OffTarget == "" {
			return "", fmt.Errorf("assay: %s condition requires an off-target glyph", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.OffTarget, m.Scenario), nil
	case NeutralPrefix:
		if m.Neutral == "" {
			return "", fmt.Errorf("assay: %s condition requires a neutral prefix", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Neutral, m.Scenario), nil
	case GlyphMinusRest:
		if m.GlyphMinusRest == "" {
			return "", fmt.Errorf("assay: %s condition requires a glyph-minus-rest material", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.GlyphMinusRest, m.Scenario), nil
	case DutyOnly:
		if m.Duty == "" {
			return "", fmt.Errorf("assay: %s condition requires a duty specification", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Duty, m.Scenario), nil
	case CorpusOnly:
		if m.Corpus == "" {
			return "", fmt.Errorf("assay: %s condition requires a corpus", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Corpus, m.Scenario), nil
	case AnnotatedInstrument:
		if m.Annotated == "" {
			return "", fmt.Errorf("assay: %s condition requires an annotated instrument", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Annotated, m.Scenario), nil
	case CartographerInstrument:
		if m.Cartographer == "" {
			return "", fmt.Errorf("assay: %s condition requires a cartographer instrument", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Cartographer, m.Scenario), nil
	case CartographerScanInstrument:
		if m.CartographerScan == "" {
			return "", fmt.Errorf("assay: %s condition requires a cartographer-scan instrument", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.CartographerScan, m.Scenario), nil
	case GroundedGlyph:
		if m.Glyph == "" {
			return "", fmt.Errorf("assay: %s condition requires a glyph", cond)
		}
		if m.Ground == "" {
			return "", fmt.Errorf("assay: %s condition requires a ground", cond)
		}
		return fmt.Sprintf("%s\n---\n%s\n---\n%s", m.Glyph, m.Ground, m.Scenario), nil
	case GroundOnly:
		if m.Ground == "" {
			return "", fmt.Errorf("assay: %s condition requires a ground", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.Ground, m.Scenario), nil
	case DomainImperativeOnly:
		if m.DomainImperative == "" {
			return "", fmt.Errorf("assay: %s condition requires a domain imperative", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.DomainImperative, m.Scenario), nil
	case GroundNonPrescriptive:
		if m.NonPrescriptiveGround == "" {
			return "", fmt.Errorf("assay: %s condition requires a non-prescriptive ground", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.NonPrescriptiveGround, m.Scenario), nil
	case CanonAligned:
		if m.CanonAligned == "" {
			return "", fmt.Errorf("assay: %s condition requires a canon-aligned block", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.CanonAligned, m.Scenario), nil
	case CanonConflict:
		if m.CanonConflict == "" {
			return "", fmt.Errorf("assay: %s condition requires a canon-conflict block", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.CanonConflict, m.Scenario), nil
	case ScrambledCanon:
		if m.ScrambledCanon == "" {
			return "", fmt.Errorf("assay: %s condition requires a scrambled-canon block", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.ScrambledCanon, m.Scenario), nil
	case OffTargetCanon:
		if m.OffTargetCanon == "" {
			return "", fmt.Errorf("assay: %s condition requires an off-target-canon block", cond)
		}
		return fmt.Sprintf("%s\n---\n%s", m.OffTargetCanon, m.Scenario), nil
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
	// Truncated is true when the reply was cut off at the token cap rather than
	// stopping naturally. Recorded so a verbose answer cut off mid-artifact is a
	// visible fact for scoring and analysis, not something a rater has to infer.
	Truncated bool `json:"truncated"`
}

// ProbeResponse is the raw model reply for one condition/run, retained for
// audit alongside the scored row.
type ProbeResponse struct {
	Condition Condition
	Run       int
	// Prompt is the assembled probe prompt ([guidance] "---" [scenario]) — the
	// material the assay handed the client, before any endpoint-specific wrapping.
	Prompt string
	// RenderedPrompt is the exact string the client put on the wire, when it
	// surfaces one — the study-declared wrapper with the material substituted, for
	// the raw /completion path. Empty for the chat endpoint. This is the literal
	// input recorded for a self-describing, reproducible run.
	RenderedPrompt string
	Text           string
	// Reasoning is the thinking a reasoning model produced, when the server
	// surfaced it (reasoning_content, or the inline <think>…</think> block the
	// client strips from Text). Empty for a non-thinking model. Retained so a
	// study can check a self-reported field source against the route the model
	// actually reasoned through; never parsed for the answer.
	Reasoning string
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

	// Stop lists the strings that halt generation the moment the server emits
	// one. It is NOT a distribution-shaping stage of llama.cpp's sampler chain
	// (penalties → … → temperature); it is a generation-control parameter that
	// belongs in the recorded regime for the same reason the rest do — so a run
	// says what it actually sent. The single-turn grounded-glyph probe leaves it
	// empty (it parses full completions); the agentic-loop probe sets it to the
	// tool-call boundary so the harness, not the model, supplies each
	// OBSERVATION. Recorded per run: an unrecorded stop is a silent constant, the
	// exact failure the record-what-ran invariant exists to prevent.
	Stop []string `json:"stop,omitempty"`
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

		Stop: s.Stop,
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
	truncNote := ""
	if resp.Truncated {
		truncNote = ":truncated"
	}
	rationale := fmt.Sprintf("grounded-glyph-probe:%s:response=%dchars%s:unscored", cond, len(resp.Text), truncNote)

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
			Truncated:       resp.Truncated,
		},
	}
	response := ProbeResponse{
		Condition:      cond,
		Run:            run,
		Prompt:         prompt,
		RenderedPrompt: resp.RenderedPrompt,
		Text:           resp.Text,
		Reasoning:      resp.Reasoning,
	}
	return row, response, nil
}
