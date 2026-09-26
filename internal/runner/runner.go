// Package runner drives an assay from the container's /in → /out contract:
// it reads a typed study spec plus material files from the input directory,
// runs every condition/replicate against an injected model client, and writes
// the score rows and raw responses to the output directory. It is the real
// adapter the registry-lab shims deferred (task e9-lab-controller).
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"corpos-lab/internal/agentloop"
	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

// SupportedAssay is the single-turn probe variant: one model call per cell.
// Other variants (behavioral-equivalence, structural-glyph-probe,
// decomposition) register as new cases when their logic lands.
const SupportedAssay = "grounded-glyph-probe"

// LoopAssay is the agentic-loop probe variant: a minimal tool-using loop per
// cell instead of a single completion. It answers whether analysis-mode
// survives when the subject can act (chain 550). The aid delivery and scoring
// are shared with the single-turn probe; only the setup differs.
const LoopAssay = "agentic-loop-probe"

// SupportedAssays reports whether the runner executes an assay variant.
func SupportedAssays(name string) bool {
	return name == SupportedAssay || name == LoopAssay
}

// ModelSpec is the model connection a study runs against. The container
// receives it in study.json rather than baking it in, so one image serves
// any local endpoint the controller points it at.
type ModelSpec struct {
	BaseURL string `json:"base_url"`
	ModelID string `json:"model_id"`
	Version string `json:"version"`
	// Endpoint selects the inference path: "chat" (default) hits
	// /v1/chat/completions; "completion" hits the raw /completion endpoint with
	// no server-side chat template — the least-opinionated subject path.
	Endpoint string `json:"endpoint,omitempty"`
	// PromptTemplate is the study-declared instruct wrapper for "completion"
	// mode: the assembled prompt is substituted at its "{prompt}" placeholder and
	// sent verbatim. Unused in "chat" mode.
	PromptTemplate string `json:"prompt_template,omitempty"`
}

// MaterialsSpec names the material files (relative to the input dir) the
// assay assembles prompts from. Glyph, Ground, and Imperative are optional for
// conditions that don't use them.
type MaterialsSpec struct {
	Scenario       string `json:"scenario"`
	Glyph          string `json:"glyph,omitempty"`
	Ground         string `json:"ground,omitempty"`
	Imperative     string `json:"imperative,omitempty"`
	Scrambled      string `json:"scrambled,omitempty"`
	OffTarget      string `json:"off_target,omitempty"`
	Neutral        string `json:"neutral,omitempty"`
	GlyphMinusRest string `json:"glyph_minus_rest,omitempty"`
	Duty           string `json:"duty,omitempty"`
	Corpus         string `json:"corpus,omitempty"`
	// Annotated, Cartographer, and CartographerScan are the design instruments for
	// the cartographer-duty-format assay's format conditions.
	Annotated        string `json:"annotated,omitempty"`
	Cartographer     string `json:"cartographer,omitempty"`
	CartographerScan string `json:"cartographer_scan,omitempty"`
	// DomainImperative is the domain-specific directive for the
	// domain_imperative_only condition of the form × grounding 2×2.
	DomainImperative string `json:"domain_imperative,omitempty"`
	// NonPrescriptiveGround is the domain ground with its outcome sentences
	// removed, for the ground_nonprescriptive condition.
	NonPrescriptiveGround string `json:"ground_nonprescriptive,omitempty"`
	// CanonAligned, CanonConflict, ScrambledCanon, and OffTargetCanon are the
	// prepended blocks for the content-priority-under-conflict conditions.
	CanonAligned   string `json:"canon_aligned,omitempty"`
	CanonConflict  string `json:"canon_conflict,omitempty"`
	ScrambledCanon string `json:"scrambled_canon,omitempty"`
	OffTargetCanon string `json:"off_target_canon,omitempty"`
}

// StudySpec is the container's /in/study.json — a self-contained description
// of one assay run: which assay, which item, which model, which conditions,
// how many replicates, where the materials are, and how to sample.
//
// Sampling rides in this file deliberately: it is what the container actually
// applies, so putting it here means the run's own inputs record the sampler
// rather than it living as an invisible Go constant. (Originally justified by
// CHARTER freeze-by-digest, retired 2026-07-14 — see INQUIRY.md. The placement
// stands on its own: a hardcoded sampler is one nobody can audit afterwards.)
type StudySpec struct {
	Assay       string            `json:"assay"`
	ItemID      string            `json:"item_id"`
	Model       ModelSpec         `json:"model"`
	Conditions  []assay.Condition `json:"conditions"`
	RunsPerCell int               `json:"runs_per_cell"`
	Materials   MaterialsSpec     `json:"materials"`
	Sampling    assay.Sampling    `json:"sampling"`
	// Loop is the extra config the agentic-loop assay needs; ignored by the
	// single-turn assay. A study with Assay == LoopAssay must set Preamble and
	// Sandbox.
	Loop LoopSpec `json:"loop,omitempty"`
}

// LoopSpec configures the agentic-loop probe. Preamble and Sandbox name files in
// the container /in (canonical local names, like the material files). StepCap
// and CallTokens are optional; zero means the agentloop defaults.
type LoopSpec struct {
	// Preamble is the file holding the loop's fixed preamble — the tool
	// vocabulary and the "you can act" framing prepended to the aided scenario.
	Preamble string `json:"preamble,omitempty"`
	// Sandbox is the file holding the per-scenario sandbox as a JSON
	// path→contents map the tools act on.
	Sandbox string `json:"sandbox,omitempty"`
	// StepCap bounds the turns per run; 0 uses the agentloop default.
	StepCap int `json:"step_cap,omitempty"`
	// CallTokens caps each turn's generation; 0 uses the agentloop default.
	CallTokens int `json:"call_tokens,omitempty"`
	// Stop overrides the loop's stop list — the strings that halt a turn so the
	// harness supplies the OBSERVATION. Empty uses the agentloop default
	// ("\nOBSERVATION"). Whatever is effective is recorded in the results sampler
	// (record-what-ran), so a study need not set this to have its stop captured.
	Stop []string `json:"stop,omitempty"`
}

// Results is the container's /out/results.json — the scored rows plus the
// identifying context needed to attribute them.
//
// Sampler and Server are what make a run self-describing. Sampler is the
// complete regime the rows were produced under; Server is the inference
// server's own account of itself, read back after the loop. Together with each
// row's Observed block they answer "what actually ran" from the artifact alone,
// with no appeal to what anyone intended.
type Results struct {
	Assay   string           `json:"assay"`
	ItemID  string           `json:"item_id"`
	ModelID string           `json:"model_id"`
	Rows    []assay.ScoreRow `json:"rows"`
	// Sampler is the full chain actually sent with every generation.
	Sampler assay.Sampling `json:"sampler"`
	// Server is GET /props, read back after the run. Empty if the readback
	// failed — a run is not voided for failing to describe itself, but the
	// silence is visible rather than filled in.
	Server model.ServerProps `json:"server"`
	// ServerReadbackError records why Server is empty, when it is.
	ServerReadbackError string `json:"server_readback_error,omitempty"`
	// ModelMismatch is set when the server reports serving a model other than
	// the one the study declared. RECORDED, NEVER ENFORCED: a divergence is
	// information about this run, not grounds for refusing it. This is the
	// check nothing performed when a study believed it was running Mistral.
	ModelMismatch string `json:"model_mismatch,omitempty"`
}

// LoadSpec reads and validates study.json from inDir.
func LoadSpec(inDir string) (StudySpec, error) {
	raw, err := os.ReadFile(filepath.Join(inDir, "study.json"))
	if err != nil {
		return StudySpec{}, fmt.Errorf("runner: read study.json: %w", err)
	}
	var spec StudySpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return StudySpec{}, fmt.Errorf("runner: parse study.json: %w", err)
	}
	if err := spec.validate(); err != nil {
		return StudySpec{}, err
	}
	return spec, nil
}

func (s StudySpec) validate() error {
	if !SupportedAssays(s.Assay) {
		return fmt.Errorf("runner: unsupported assay %q (implemented: %q, %q)", s.Assay, SupportedAssay, LoopAssay)
	}
	if s.Assay == LoopAssay {
		if s.Loop.Preamble == "" {
			return fmt.Errorf("runner: assay %q requires loop.preamble", LoopAssay)
		}
		if s.Loop.Sandbox == "" {
			return fmt.Errorf("runner: assay %q requires loop.sandbox", LoopAssay)
		}
	}
	if s.ItemID == "" {
		return fmt.Errorf("runner: study.json missing item_id")
	}
	if len(s.Conditions) == 0 {
		return fmt.Errorf("runner: study.json lists no conditions")
	}
	if s.RunsPerCell < 1 {
		return fmt.Errorf("runner: runs_per_cell must be >= 1, got %d", s.RunsPerCell)
	}
	if s.Materials.Scenario == "" {
		return fmt.Errorf("runner: study.json missing materials.scenario")
	}
	// Sampling is re-checked container-side, not just host-side: a container
	// can be launched against any /in, and a spec with no sampling block
	// unmarshals to a zero value that would silently decode greedily with no
	// token cap. The executor refuses rather than inventing a regime.
	if s.Sampling.MaxTokens < 1 {
		return fmt.Errorf("runner: study.json missing sampling.max_tokens (got %d)", s.Sampling.MaxTokens)
	}
	// repeat_penalty is neutral at 1.0 and has no valid zero, so a zero here is
	// the tell that the sampler chain was never declared — the same
	// zero-value-means-absent reasoning as max_tokens above. The host catches
	// omission field-by-field (it has pointers); the container only has the
	// decoded values, so it checks the one field whose zero cannot be a choice.
	if s.Sampling.RepeatPenalty <= 0 {
		return fmt.Errorf("runner: study.json sampling.repeat_penalty is %v — the sampler chain "+
			"was not declared, and an undeclared chain inherits the server's defaults",
			s.Sampling.RepeatPenalty)
	}
	if s.Sampling.Temperature > 0 && len(s.Sampling.Seeds) != s.RunsPerCell {
		return fmt.Errorf("runner: study.json sampling.seeds has %d entries for %d runs_per_cell — "+
			"sampled runs need one seed each to reproduce", len(s.Sampling.Seeds), s.RunsPerCell)
	}
	return nil
}

// loadMaterials reads the material files named in the spec from inDir.
// Optional materials (glyph, ground) are read only when named; a named-but-
// missing file is an error, not a silent empty string.
func (s StudySpec) loadMaterials(inDir string) (assay.Materials, error) {
	read := func(name string) (string, error) {
		if name == "" {
			return "", nil
		}
		b, err := os.ReadFile(filepath.Join(inDir, name))
		if err != nil {
			return "", fmt.Errorf("runner: read material %q: %w", name, err)
		}
		return string(b), nil
	}
	scenario, err := read(s.Materials.Scenario)
	if err != nil {
		return assay.Materials{}, err
	}
	glyph, err := read(s.Materials.Glyph)
	if err != nil {
		return assay.Materials{}, err
	}
	ground, err := read(s.Materials.Ground)
	if err != nil {
		return assay.Materials{}, err
	}
	imperative, err := read(s.Materials.Imperative)
	if err != nil {
		return assay.Materials{}, err
	}
	scrambled, err := read(s.Materials.Scrambled)
	if err != nil {
		return assay.Materials{}, err
	}
	offTarget, err := read(s.Materials.OffTarget)
	if err != nil {
		return assay.Materials{}, err
	}
	neutral, err := read(s.Materials.Neutral)
	if err != nil {
		return assay.Materials{}, err
	}
	glyphMinusRest, err := read(s.Materials.GlyphMinusRest)
	if err != nil {
		return assay.Materials{}, err
	}
	duty, err := read(s.Materials.Duty)
	if err != nil {
		return assay.Materials{}, err
	}
	corpus, err := read(s.Materials.Corpus)
	if err != nil {
		return assay.Materials{}, err
	}
	annotated, err := read(s.Materials.Annotated)
	if err != nil {
		return assay.Materials{}, err
	}
	cartographer, err := read(s.Materials.Cartographer)
	if err != nil {
		return assay.Materials{}, err
	}
	cartographerScan, err := read(s.Materials.CartographerScan)
	if err != nil {
		return assay.Materials{}, err
	}
	domainImperative, err := read(s.Materials.DomainImperative)
	if err != nil {
		return assay.Materials{}, err
	}
	nonPrescriptiveGround, err := read(s.Materials.NonPrescriptiveGround)
	if err != nil {
		return assay.Materials{}, err
	}
	canonAligned, err := read(s.Materials.CanonAligned)
	if err != nil {
		return assay.Materials{}, err
	}
	canonConflict, err := read(s.Materials.CanonConflict)
	if err != nil {
		return assay.Materials{}, err
	}
	scrambledCanon, err := read(s.Materials.ScrambledCanon)
	if err != nil {
		return assay.Materials{}, err
	}
	offTargetCanon, err := read(s.Materials.OffTargetCanon)
	if err != nil {
		return assay.Materials{}, err
	}
	return assay.Materials{
		Scenario: scenario, Glyph: glyph, Ground: ground, Imperative: imperative,
		Scrambled: scrambled, OffTarget: offTarget, Neutral: neutral, GlyphMinusRest: glyphMinusRest,
		Duty: duty, Corpus: corpus,
		Annotated: annotated, Cartographer: cartographer, CartographerScan: cartographerScan,
		DomainImperative: domainImperative, NonPrescriptiveGround: nonPrescriptiveGround,
		CanonAligned: canonAligned, CanonConflict: canonConflict,
		ScrambledCanon: scrambledCanon, OffTargetCanon: offTargetCanon,
	}, nil
}

// modelMismatch compares the model the study DECLARED against the artifact the
// server says it actually loaded, and returns a human-readable divergence or
// "" when they agree.
//
// Nothing ever performed this comparison. `model_id` was echoed from the TOML
// into the record and never checked against reality, so a study could name one
// model, be served another, and read back as if it had got what it asked for.
// The result is reported, never enforced — per the record-what-ran invariant, a
// mismatch is information about the run, not grounds for refusing it.
func modelMismatch(declared string, props model.ServerProps) string {
	if declared == "" || (props.ModelAlias == "" && props.ModelPath == "") {
		return ""
	}
	if props.ModelAlias == declared || filepath.Base(props.ModelPath) == declared {
		return ""
	}
	return fmt.Sprintf("study declared model_id %q; server reports alias %q (path %q)",
		declared, props.ModelAlias, props.ModelPath)
}

// Execute runs the study described by inDir/study.json against client and
// writes outDir/results.json plus outDir/responses/<condition>_<run>.txt. A
// thinking model's reasoning trace, when the client surfaces one, is written to
// outDir/reasoning/<condition>_<run>.txt so a study can check a self-reported
// field source against the route the model actually reasoned through.
// It fails fast: the first probe error aborts the run (a partial results.json
// would misrepresent a study cell).
func Execute(ctx context.Context, inDir, outDir string, client model.Client) (Results, error) {
	spec, err := LoadSpec(inDir)
	if err != nil {
		return Results{}, err
	}
	mats, err := spec.loadMaterials(inDir)
	if err != nil {
		return Results{}, err
	}

	responsesDir := filepath.Join(outDir, "responses")
	if err := os.MkdirAll(responsesDir, 0o755); err != nil {
		return Results{}, fmt.Errorf("runner: create responses dir: %w", err)
	}
	// The rendered prompt is the literal input the subject received; recording it
	// per run makes a run self-describing down to its input (reproducibility
	// contract). Written only when the client surfaces it — the chat endpoint
	// applies its own template, so the true input is not ours to record there.
	promptsDir := filepath.Join(outDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		return Results{}, fmt.Errorf("runner: create prompts dir: %w", err)
	}
	// A thinking model's reasoning trace is recorded per run when the client
	// surfaces one. Written only when non-empty — a non-thinking model produces
	// none, and an empty file would misrepresent that as a captured blank trace.
	reasoningDir := filepath.Join(outDir, "reasoning")
	if err := os.MkdirAll(reasoningDir, 0o755); err != nil {
		return Results{}, fmt.Errorf("runner: create reasoning dir: %w", err)
	}

	dirs := outDirs{responses: responsesDir, prompts: promptsDir, reasoning: reasoningDir}
	// sampler is the regime the record reports. For the loop assay it gains the
	// stop the loop actually sent, resolved once from the loop config so the
	// record and the wire cannot disagree (record-what-ran).
	sampler := spec.Sampling
	var rows []assay.ScoreRow
	if spec.Assay == LoopAssay {
		var stop []string
		rows, stop, err = runLoopCells(ctx, client, spec, mats, inDir, outDir, dirs)
		sampler.Stop = stop
	} else {
		rows, err = runSingleTurnCells(ctx, client, spec, mats, dirs)
	}
	if err != nil {
		return Results{}, err
	}

	results := Results{
		Assay:   spec.Assay,
		ItemID:  spec.ItemID,
		ModelID: client.Name(),
		Rows:    rows,
		Sampler: sampler,
	}

	// Read the server's self-report back AFTER the loop, so it describes the
	// server that answered rather than one that might have been swapped since.
	// A failed readback degrades the record; it does not void the run — the
	// rows are real either way, and refusing them would be the freeze again.
	props, err := client.Props(ctx)
	if err != nil {
		results.ServerReadbackError = err.Error()
	} else {
		results.Server = props
		results.ModelMismatch = modelMismatch(spec.Model.ModelID, props)
	}

	if err := writeResults(outDir, results); err != nil {
		return Results{}, err
	}
	return results, nil
}

// outDirs are the per-run output directories Execute prepared.
type outDirs struct {
	responses string
	prompts   string
	reasoning string
}

// runSingleTurnCells runs the grounded-glyph probe: one model call per cell,
// writing the response, the rendered prompt (when surfaced), and the reasoning
// trace (when non-empty) per run. It is the raw-completion arm of chain 550 and
// the whole of the earlier program.
func runSingleTurnCells(ctx context.Context, client model.Client, spec StudySpec, mats assay.Materials, dirs outDirs) ([]assay.ScoreRow, error) {
	var rows []assay.ScoreRow
	for _, cond := range spec.Conditions {
		for run := 1; run <= spec.RunsPerCell; run++ {
			row, resp, err := assay.RunProbe(ctx, client, spec.ItemID, cond, run, mats, spec.Sampling)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)

			if err := writeCell(dirs.responses, cond, run, resp.Text); err != nil {
				return nil, err
			}
			if resp.RenderedPrompt != "" {
				if err := writeCell(dirs.prompts, cond, run, resp.RenderedPrompt); err != nil {
					return nil, err
				}
			}
			if resp.Reasoning != "" {
				if err := writeCell(dirs.reasoning, cond, run, resp.Reasoning); err != nil {
					return nil, err
				}
			}
		}
	}
	return rows, nil
}

// runLoopCells runs the agentic-loop probe: a minimal tool loop per cell against
// a fresh sandbox. It writes the flat transcript (the artifact a blind rater
// scores) as the response, the assembled base prompt, and the final sandbox
// state per run so the run records what the subject did, not just what it said.
//
// It returns the scored rows and the stop list the loop actually sent (the
// study override or the agentloop default), so Execute records the effective
// stop in the results sampler.
func runLoopCells(ctx context.Context, client model.Client, spec StudySpec, mats assay.Materials, inDir, outDir string, dirs outDirs) ([]assay.ScoreRow, []string, error) {
	preamble, sandbox, err := loadLoopInputs(inDir, spec.Loop)
	if err != nil {
		return nil, nil, err
	}
	sandboxDir := filepath.Join(outDir, "sandboxes")
	if err := os.MkdirAll(sandboxDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("runner: create sandboxes dir: %w", err)
	}
	cfg := agentloop.Config{StepCap: spec.Loop.StepCap, CallTokens: spec.Loop.CallTokens, Stop: spec.Loop.Stop}

	var rows []assay.ScoreRow
	for _, cond := range spec.Conditions {
		for run := 1; run <= spec.RunsPerCell; run++ {
			row, resp, arts, err := assay.RunLoopProbe(ctx, client, spec.ItemID, cond, run, mats, preamble, sandbox, spec.Sampling, cfg)
			if err != nil {
				return nil, nil, err
			}
			rows = append(rows, row)

			if err := writeCell(dirs.responses, cond, run, resp.Text); err != nil {
				return nil, nil, err
			}
			if err := writeCell(dirs.prompts, cond, run, resp.Prompt); err != nil {
				return nil, nil, err
			}
			snap, err := json.MarshalIndent(arts.Sandbox, "", "  ")
			if err != nil {
				return nil, nil, fmt.Errorf("runner: marshal sandbox %s run %d: %w", cond, run, err)
			}
			if err := writeCell(sandboxDir, cond, run, string(snap)); err != nil {
				return nil, nil, err
			}
		}
	}
	return rows, cfg.EffectiveStop(), nil
}

// loadLoopInputs reads the loop preamble and the sandbox map from the container
// /in. A named-but-unreadable file, or a sandbox that is not a JSON
// path→contents map, is an error rather than a silent empty loop.
func loadLoopInputs(inDir string, spec LoopSpec) (string, map[string]string, error) {
	preamble, err := os.ReadFile(filepath.Join(inDir, spec.Preamble))
	if err != nil {
		return "", nil, fmt.Errorf("runner: read loop preamble %q: %w", spec.Preamble, err)
	}
	raw, err := os.ReadFile(filepath.Join(inDir, spec.Sandbox))
	if err != nil {
		return "", nil, fmt.Errorf("runner: read loop sandbox %q: %w", spec.Sandbox, err)
	}
	var sandbox map[string]string
	if err := json.Unmarshal(raw, &sandbox); err != nil {
		return "", nil, fmt.Errorf("runner: parse loop sandbox %q (want a JSON path→contents object): %w", spec.Sandbox, err)
	}
	return string(preamble), sandbox, nil
}

// writeCell writes one per-run artifact under dir as <condition>_<run>.txt.
func writeCell(dir string, cond assay.Condition, run int, body string) error {
	path := filepath.Join(dir, fmt.Sprintf("%s_%d.txt", cond, run))
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("runner: write %s: %w", path, err)
	}
	return nil
}

func writeResults(outDir string, results Results) error {
	raw, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("runner: marshal results: %w", err)
	}
	path := filepath.Join(outDir, "results.json")
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("runner: write results.json: %w", err)
	}
	return nil
}
