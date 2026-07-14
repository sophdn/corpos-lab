// Package study defines the host-authored study definition — the single file
// a researcher writes to describe one reproducible assay run — and its
// materialization into the container's /in contract. The definition is TOML
// (human-authored, corpos-family idiom); the container sees derived JSON.
package study

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/runner"
)

// ModelDef is the model connection the assay runs against.
type ModelDef struct {
	BaseURL string `toml:"base_url"`
	ModelID string `toml:"model_id"`
	Version string `toml:"version"`
}

// MaterialsDef names the material files, as paths relative to the definition
// file (or absolute). Glyph and Ground are optional for conditions that don't
// use them.
type MaterialsDef struct {
	Scenario string `toml:"scenario"`
	Glyph    string `toml:"glyph"`
	Ground   string `toml:"ground"`
}

// SamplingDef is the study's declared sampling regime.
//
// Temperature and MaxTokens are pointers so that an OMITTED field is
// distinguishable from a deliberate zero. This matters concretely: temperature
// 0.0 is a legitimate choice (deterministic single-shot), so a plain float64
// would make "I chose greedy decoding" and "I forgot to say" identical — and
// an unrecorded temperature is precisely the gap being closed here. The
// casg-direct v3 study.json never recorded its temperature, so the original
// value is now permanently unknowable and its reproduction carries an
// inferred-0.8 delta forever.
type SamplingDef struct {
	Temperature *float64 `toml:"temperature"`
	// Seeds pins one RNG seed per replicate; run N uses Seeds[N-1]. Required
	// whenever temperature > 0, so sampled runs stay reproducible.
	Seeds     []int `toml:"seeds"`
	MaxTokens *int  `toml:"max_tokens"`
}

// Def is a complete, self-contained study definition. A study is reproducible
// from this file alone (plus the pinned image) — no agent in the loop.
type Def struct {
	Name        string       `toml:"name"`
	Assay       string       `toml:"assay"`
	ItemID      string       `toml:"item_id"`
	Image       string       `toml:"image"`
	Network     string       `toml:"network"`
	Model       ModelDef     `toml:"model"`
	Conditions  []string     `toml:"conditions"`
	RunsPerCell int          `toml:"runs_per_cell"`
	Materials   MaterialsDef `toml:"materials"`
	Sampling    SamplingDef  `toml:"sampling"`

	// baseDir is the directory of the definition file, used to resolve
	// relative material paths. Not part of the wire format.
	baseDir string
}

// LoadDef reads, parses, and validates a study definition from path.
func LoadDef(path string) (Def, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Def{}, fmt.Errorf("study: read def %s: %w", path, err)
	}
	var d Def
	if err := toml.Unmarshal(raw, &d); err != nil {
		return Def{}, fmt.Errorf("study: parse def %s: %w", path, err)
	}
	d.baseDir = filepath.Dir(path)
	if err := d.validate(); err != nil {
		return Def{}, err
	}
	return d, nil
}

func (d Def) validate() error {
	for _, f := range []struct {
		name, val string
	}{
		{"name", d.Name}, {"assay", d.Assay}, {"item_id", d.ItemID},
		{"image", d.Image}, {"model.base_url", d.Model.BaseURL},
		{"model.model_id", d.Model.ModelID}, {"materials.scenario", d.Materials.Scenario},
	} {
		if f.val == "" {
			return fmt.Errorf("study: definition missing required field %q", f.name)
		}
	}
	if d.Assay != runner.SupportedAssay {
		return fmt.Errorf("study: unsupported assay %q (only %q implemented)", d.Assay, runner.SupportedAssay)
	}
	if len(d.Conditions) == 0 {
		return fmt.Errorf("study: definition lists no conditions")
	}
	if d.RunsPerCell < 1 {
		return fmt.Errorf("study: runs_per_cell must be >= 1, got %d", d.RunsPerCell)
	}
	if err := d.validateSampling(); err != nil {
		return err
	}
	for _, c := range d.Conditions {
		cond := assay.Condition(c)
		switch cond {
		case assay.Baseline:
		case assay.GlyphOnly:
			if d.Materials.Glyph == "" {
				return fmt.Errorf("study: condition %q requires materials.glyph", c)
			}
		case assay.GroundedGlyph:
			if d.Materials.Glyph == "" {
				return fmt.Errorf("study: condition %q requires materials.glyph", c)
			}
			if d.Materials.Ground == "" {
				return fmt.Errorf("study: condition %q requires materials.ground", c)
			}
		default:
			return fmt.Errorf("study: unknown condition %q", c)
		}
	}
	return nil
}

// validateSampling enforces that a study SAYS what it sampled. Every rule here
// refuses a silent default: an unrecorded sampling regime is unreproducible,
// and a result that cannot reproduce its manifest is an anecdote, not data
// (CHARTER freeze-by-digest).
func (d Def) validateSampling() error {
	s := d.Sampling
	if s.Temperature == nil {
		return fmt.Errorf("study: sampling.temperature must be declared explicitly " +
			"(0.0 for deterministic decoding) — an unrecorded temperature is unreproducible")
	}
	if *s.Temperature < 0 {
		return fmt.Errorf("study: sampling.temperature must be >= 0, got %v", *s.Temperature)
	}
	if s.MaxTokens == nil {
		return fmt.Errorf("study: sampling.max_tokens must be declared explicitly")
	}
	if *s.MaxTokens < 1 {
		return fmt.Errorf("study: sampling.max_tokens must be >= 1, got %d", *s.MaxTokens)
	}

	// Above greedy, replicates only differ if the sampler is seeded — and they
	// only REPRODUCE if the seeds are written down. Requiring one seed per
	// replicate is what buys reproducible-but-varied runs, the property greedy
	// decoding cannot offer at any seed.
	if *s.Temperature > 0 {
		if len(s.Seeds) != d.RunsPerCell {
			return fmt.Errorf("study: sampling.seeds must list exactly one seed per replicate "+
				"(runs_per_cell = %d, got %d) — sampled runs are only reproducible when seeded",
				d.RunsPerCell, len(s.Seeds))
		}
		seen := map[int]bool{}
		for _, seed := range s.Seeds {
			if seen[seed] {
				return fmt.Errorf("study: sampling.seeds repeats seed %d — "+
					"duplicate seeds make replicates identical and deflate the cell's variance", seed)
			}
			seen[seed] = true
		}
		return nil
	}

	// Deterministic mode stays available, but it must be chosen, not inherited.
	// Seeds are meaningless under greedy decoding, so accepting them here would
	// imply a variation the run cannot deliver.
	if len(s.Seeds) > 0 {
		return fmt.Errorf("study: sampling.seeds is set but temperature is 0 — " +
			"greedy decoding ignores seeds and returns an identical reply per replicate; " +
			"either raise the temperature or drop the seeds")
	}
	return nil
}

// sampling returns the definition's sampling regime as the typed assay value.
func (d Def) sampling() assay.Sampling {
	return assay.Sampling{
		Temperature: *d.Sampling.Temperature,
		Seeds:       d.Sampling.Seeds,
		MaxTokens:   *d.Sampling.MaxTokens,
	}
}

// conditions returns the definition's conditions as typed assay.Condition.
func (d Def) conditions() []assay.Condition {
	out := make([]assay.Condition, len(d.Conditions))
	for i, c := range d.Conditions {
		out[i] = assay.Condition(c)
	}
	return out
}

// Materialize writes the container /in contract into inDir: study.json (the
// runner's JSON spec, with materials renamed to canonical local names) plus
// the copied material files. Material paths in the definition are resolved
// relative to the definition file's directory.
func (d Def) Materialize(inDir string) error {
	if err := os.MkdirAll(inDir, 0o755); err != nil {
		return fmt.Errorf("study: create in dir: %w", err)
	}

	mats := runner.MaterialsSpec{Scenario: "scenario.md"}
	if err := d.copyMaterial(d.Materials.Scenario, filepath.Join(inDir, "scenario.md")); err != nil {
		return err
	}
	if d.Materials.Glyph != "" {
		if err := d.copyMaterial(d.Materials.Glyph, filepath.Join(inDir, "glyph.md")); err != nil {
			return err
		}
		mats.Glyph = "glyph.md"
	}
	if d.Materials.Ground != "" {
		if err := d.copyMaterial(d.Materials.Ground, filepath.Join(inDir, "ground.md")); err != nil {
			return err
		}
		mats.Ground = "ground.md"
	}

	spec := runner.StudySpec{
		Assay:  d.Assay,
		ItemID: d.ItemID,
		Model: runner.ModelSpec{
			BaseURL: d.Model.BaseURL,
			ModelID: d.Model.ModelID,
			Version: d.Model.Version,
		},
		Conditions:  d.conditions(),
		RunsPerCell: d.RunsPerCell,
		Materials:   mats,
		Sampling:    d.sampling(),
	}
	raw, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("study: marshal study.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(inDir, "study.json"), append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("study: write study.json: %w", err)
	}
	return nil
}

// copyMaterial resolves src relative to the definition dir and copies it to
// dst, failing loudly if the material is missing.
func (d Def) copyMaterial(src, dst string) error {
	if !filepath.IsAbs(src) {
		src = filepath.Join(d.baseDir, src)
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("study: read material %s: %w", src, err)
	}
	if err := os.WriteFile(dst, raw, 0o644); err != nil {
		return fmt.Errorf("study: write material %s: %w", dst, err)
	}
	return nil
}
