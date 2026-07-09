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
