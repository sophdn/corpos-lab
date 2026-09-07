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

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

// SupportedAssay is the only assay variant this runner executes so far.
// Other variants (behavioral-equivalence, structural-glyph-probe,
// decomposition) register as new cases when their logic lands.
const SupportedAssay = "grounded-glyph-probe"

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
	Scenario   string `json:"scenario"`
	Glyph      string `json:"glyph,omitempty"`
	Ground     string `json:"ground,omitempty"`
	Imperative string `json:"imperative,omitempty"`
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
	if s.Assay != SupportedAssay {
		return fmt.Errorf("runner: unsupported assay %q (only %q implemented)", s.Assay, SupportedAssay)
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
	return assay.Materials{Scenario: scenario, Glyph: glyph, Ground: ground, Imperative: imperative}, nil
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
// writes outDir/results.json plus outDir/responses/<condition>_<run>.txt.
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

	rows := []assay.ScoreRow{}
	for _, cond := range spec.Conditions {
		for run := 1; run <= spec.RunsPerCell; run++ {
			row, resp, err := assay.RunProbe(ctx, client, spec.ItemID, cond, run, mats, spec.Sampling)
			if err != nil {
				return Results{}, err
			}
			rows = append(rows, row)

			respPath := filepath.Join(responsesDir, fmt.Sprintf("%s_%d.txt", cond, run))
			if err := os.WriteFile(respPath, []byte(resp.Text), 0o644); err != nil {
				return Results{}, fmt.Errorf("runner: write response %s: %w", respPath, err)
			}
			if resp.RenderedPrompt != "" {
				promptPath := filepath.Join(promptsDir, fmt.Sprintf("%s_%d.txt", cond, run))
				if err := os.WriteFile(promptPath, []byte(resp.RenderedPrompt), 0o644); err != nil {
					return Results{}, fmt.Errorf("runner: write prompt %s: %w", promptPath, err)
				}
			}
		}
	}

	results := Results{
		Assay:   spec.Assay,
		ItemID:  spec.ItemID,
		ModelID: client.Name(),
		Rows:    rows,
		Sampler: spec.Sampling,
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
