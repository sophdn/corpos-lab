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
}

// MaterialsSpec names the material files (relative to the input dir) the
// assay assembles prompts from. Glyph and Ground are optional for conditions
// that don't use them.
type MaterialsSpec struct {
	Scenario string `json:"scenario"`
	Glyph    string `json:"glyph,omitempty"`
	Ground   string `json:"ground,omitempty"`
}

// StudySpec is the container's /in/study.json — a self-contained description
// of one assay run: which assay, which item, which model, which conditions,
// how many replicates, and where the materials are.
type StudySpec struct {
	Assay       string            `json:"assay"`
	ItemID      string            `json:"item_id"`
	Model       ModelSpec         `json:"model"`
	Conditions  []assay.Condition `json:"conditions"`
	RunsPerCell int               `json:"runs_per_cell"`
	Materials   MaterialsSpec     `json:"materials"`
}

// Results is the container's /out/results.json — the scored rows plus the
// identifying context needed to attribute them.
type Results struct {
	Assay   string           `json:"assay"`
	ItemID  string           `json:"item_id"`
	ModelID string           `json:"model_id"`
	Rows    []assay.ScoreRow `json:"rows"`
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
	return assay.Materials{Scenario: scenario, Glyph: glyph, Ground: ground}, nil
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

	rows := []assay.ScoreRow{}
	for _, cond := range spec.Conditions {
		for run := 1; run <= spec.RunsPerCell; run++ {
			row, resp, err := assay.RunProbe(ctx, client, spec.ItemID, cond, run, mats)
			if err != nil {
				return Results{}, err
			}
			rows = append(rows, row)

			respPath := filepath.Join(responsesDir, fmt.Sprintf("%s_%d.txt", cond, run))
			if err := os.WriteFile(respPath, []byte(resp.Text), 0o644); err != nil {
				return Results{}, fmt.Errorf("runner: write response %s: %w", respPath, err)
			}
		}
	}

	results := Results{
		Assay:   spec.Assay,
		ItemID:  spec.ItemID,
		ModelID: client.Name(),
		Rows:    rows,
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
