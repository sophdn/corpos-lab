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
	"strings"

	"github.com/BurntSushi/toml"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/runner"
)

// ModelDef is the model connection the assay runs against.
type ModelDef struct {
	BaseURL string `toml:"base_url"`
	ModelID string `toml:"model_id"`
	Version string `toml:"version"`
	// Endpoint selects the inference path: "" or "chat" (default) uses
	// /v1/chat/completions; "completion" uses the raw /completion endpoint with no
	// server-side chat template — the least-opinionated subject path
	// (studies/REPRODUCIBILITY.md).
	Endpoint string `toml:"endpoint"`
	// PromptTemplate is the study-declared instruct wrapper for "completion" mode,
	// carrying a "{prompt}" placeholder the assembled prompt is substituted into.
	// Published verbatim so a reproduction sees the exact input.
	PromptTemplate string `toml:"prompt_template"`
}

// MaterialsDef names the material files, as paths relative to the definition
// file (or absolute). Glyph, Ground, and Imperative are optional for conditions
// that don't use them.
type MaterialsDef struct {
	Scenario   string `toml:"scenario"`
	Glyph      string `toml:"glyph"`
	Ground     string `toml:"ground"`
	Imperative string `toml:"imperative"`
	Scrambled  string `toml:"scrambled"`
	OffTarget  string `toml:"off_target"`
	Duty       string `toml:"duty"`
	Corpus     string `toml:"corpus"`
	// Annotated, Cartographer, and CartographerScan are the design instruments for
	// the cartographer-duty-format assay's three format conditions.
	Annotated        string `toml:"annotated"`
	Cartographer     string `toml:"cartographer"`
	CartographerScan string `toml:"cartographer_scan"`
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
// Every sampler stage is a pointer for the same reason temperature is: each
// has a meaningful zero (top_k 0 disables it, repeat_penalty is neutral at 1.0
// but 0 is a real setting), so a plain value could not tell "I disabled this
// stage deliberately" from "I never mentioned it". Omission is rejected rather
// than defaulted — see validateSampling.
type SamplingDef struct {
	Temperature *float64 `toml:"temperature"`
	// Seeds pins one RNG seed per replicate; run N uses Seeds[N-1]. Required
	// whenever temperature > 0, so sampled runs stay reproducible.
	Seeds     []int `toml:"seeds"`
	MaxTokens *int  `toml:"max_tokens"`

	// Truncation stages, in sampler-chain order.
	TopNSigma *float64 `toml:"top_n_sigma"`
	TopK      *int     `toml:"top_k"`
	TypicalP  *float64 `toml:"typical_p"`
	TopP      *float64 `toml:"top_p"`
	MinP      *float64 `toml:"min_p"`

	// Penalty stages.
	RepeatPenalty    *float64 `toml:"repeat_penalty"`
	RepeatLastN      *int     `toml:"repeat_last_n"`
	PresencePenalty  *float64 `toml:"presence_penalty"`
	FrequencyPenalty *float64 `toml:"frequency_penalty"`

	// Stage switches; sub-parameters are inert while these are zero.
	XTCProbability *float64 `toml:"xtc_probability"`
	DryMultiplier  *float64 `toml:"dry_multiplier"`
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
	switch d.Model.Endpoint {
	case "", "chat":
	case "completion":
		if !strings.Contains(d.Model.PromptTemplate, "{prompt}") {
			return fmt.Errorf("study: model.endpoint = \"completion\" requires model.prompt_template " +
				"containing the \"{prompt}\" placeholder (the study-declared instruct wrapper)")
		}
	default:
		return fmt.Errorf("study: unknown model.endpoint %q (want \"chat\" or \"completion\")", d.Model.Endpoint)
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
		case assay.ImperativeOnly:
			if d.Materials.Imperative == "" {
				return fmt.Errorf("study: condition %q requires materials.imperative", c)
			}
		case assay.ScrambledGlyph:
			if d.Materials.Scrambled == "" {
				return fmt.Errorf("study: condition %q requires materials.scrambled", c)
			}
		case assay.OffTargetGlyph:
			if d.Materials.OffTarget == "" {
				return fmt.Errorf("study: condition %q requires materials.off_target", c)
			}
		case assay.DutyOnly:
			if d.Materials.Duty == "" {
				return fmt.Errorf("study: condition %q requires materials.duty", c)
			}
		case assay.CorpusOnly:
			if d.Materials.Corpus == "" {
				return fmt.Errorf("study: condition %q requires materials.corpus", c)
			}
		case assay.AnnotatedInstrument:
			if d.Materials.Annotated == "" {
				return fmt.Errorf("study: condition %q requires materials.annotated", c)
			}
		case assay.CartographerInstrument:
			if d.Materials.Cartographer == "" {
				return fmt.Errorf("study: condition %q requires materials.cartographer", c)
			}
		case assay.CartographerScanInstrument:
			if d.Materials.CartographerScan == "" {
				return fmt.Errorf("study: condition %q requires materials.cartographer_scan", c)
			}
		default:
			return fmt.Errorf("study: unknown condition %q", c)
		}
	}
	return nil
}

// validateSampling enforces that a study SAYS what it sampled. Every rule here
// refuses a silent default: an unrecorded sampling regime is unreproducible,
// and a sampler you cannot see afterwards is one you cannot claim to have
// chosen.
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
	if err := d.validateSamplerChain(); err != nil {
		return err
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

// validateSamplerChain refuses a study that names only some of the sampler.
//
// This is the same rule temperature already had, applied to the stages that
// were quietly exempt from it. Temperature does not define a sampling
// distribution on its own: llama.cpp runs a CHAIN — penalties, dry,
// top_n_sigma, top_k, typ_p, top_p, min_p, xtc, temperature — and every stage
// left undeclared is inherited from the server binary. A study that says
// "temperature 0.8, seed N" and nothing else has specified a fraction of its
// sampler and borrowed the rest from an accident of deployment; upgrade
// llama.cpp, have it move a default, and that study samples differently with
// nothing in the record to show it.
//
// Declaring a stage is not the same as enabling it. Most of these are meant to
// be pinned to their neutral value — the point is that the file SAYS so.
func (d Def) validateSamplerChain() error {
	s := d.Sampling
	stages := []struct {
		name string
		set  bool
	}{
		{"top_n_sigma", s.TopNSigma != nil},
		{"top_k", s.TopK != nil},
		{"typical_p", s.TypicalP != nil},
		{"top_p", s.TopP != nil},
		{"min_p", s.MinP != nil},
		{"repeat_penalty", s.RepeatPenalty != nil},
		{"repeat_last_n", s.RepeatLastN != nil},
		{"presence_penalty", s.PresencePenalty != nil},
		{"frequency_penalty", s.FrequencyPenalty != nil},
		{"xtc_probability", s.XTCProbability != nil},
		{"dry_multiplier", s.DryMultiplier != nil},
	}
	var missing []string
	for _, st := range stages {
		if !st.set {
			missing = append(missing, "sampling."+st.name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("study: sampler chain under-specified, missing %s — "+
			"temperature alone does not define a sampling distribution, and an "+
			"undeclared stage silently inherits the server's default "+
			"(pin it to its neutral value to disable it, but say so)",
			strings.Join(missing, ", "))
	}
	if *s.TopK < 0 {
		return fmt.Errorf("study: sampling.top_k must be >= 0 (0 disables it), got %d", *s.TopK)
	}
	if *s.RepeatPenalty <= 0 {
		return fmt.Errorf("study: sampling.repeat_penalty must be > 0 (1.0 is neutral), got %v", *s.RepeatPenalty)
	}
	if *s.RepeatLastN < 0 {
		return fmt.Errorf("study: sampling.repeat_last_n must be >= 0 (0 disables it), got %d", *s.RepeatLastN)
	}
	for _, p := range []struct {
		name string
		val  float64
	}{
		{"top_p", *s.TopP},
		{"min_p", *s.MinP},
		{"typical_p", *s.TypicalP},
		{"xtc_probability", *s.XTCProbability},
	} {
		if p.val < 0 || p.val > 1 {
			return fmt.Errorf("study: sampling.%s must be in [0,1], got %v", p.name, p.val)
		}
	}
	if *s.DryMultiplier < 0 {
		return fmt.Errorf("study: sampling.dry_multiplier must be >= 0 (0 disables it), got %v", *s.DryMultiplier)
	}
	return nil
}

// sampling returns the definition's sampling regime as the typed assay value.
// The bare dereferences are safe only because validateSampling ran in LoadDef —
// it is load-bearing against a nil panic here, not merely advisory.
func (d Def) sampling() assay.Sampling {
	return assay.Sampling{
		Temperature: *d.Sampling.Temperature,
		Seeds:       d.Sampling.Seeds,
		MaxTokens:   *d.Sampling.MaxTokens,

		TopNSigma: *d.Sampling.TopNSigma,
		TopK:      *d.Sampling.TopK,
		TypicalP:  *d.Sampling.TypicalP,
		TopP:      *d.Sampling.TopP,
		MinP:      *d.Sampling.MinP,

		RepeatPenalty:    *d.Sampling.RepeatPenalty,
		RepeatLastN:      *d.Sampling.RepeatLastN,
		PresencePenalty:  *d.Sampling.PresencePenalty,
		FrequencyPenalty: *d.Sampling.FrequencyPenalty,

		XTCProbability: *d.Sampling.XTCProbability,
		DryMultiplier:  *d.Sampling.DryMultiplier,
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
	if d.Materials.Imperative != "" {
		if err := d.copyMaterial(d.Materials.Imperative, filepath.Join(inDir, "imperative.md")); err != nil {
			return err
		}
		mats.Imperative = "imperative.md"
	}
	if d.Materials.Scrambled != "" {
		if err := d.copyMaterial(d.Materials.Scrambled, filepath.Join(inDir, "scrambled_glyph.md")); err != nil {
			return err
		}
		mats.Scrambled = "scrambled_glyph.md"
	}
	if d.Materials.OffTarget != "" {
		if err := d.copyMaterial(d.Materials.OffTarget, filepath.Join(inDir, "off_target_glyph.md")); err != nil {
			return err
		}
		mats.OffTarget = "off_target_glyph.md"
	}
	if d.Materials.Duty != "" {
		if err := d.copyMaterial(d.Materials.Duty, filepath.Join(inDir, "duty.md")); err != nil {
			return err
		}
		mats.Duty = "duty.md"
	}
	if d.Materials.Corpus != "" {
		if err := d.copyMaterial(d.Materials.Corpus, filepath.Join(inDir, "corpus.md")); err != nil {
			return err
		}
		mats.Corpus = "corpus.md"
	}
	if d.Materials.Annotated != "" {
		if err := d.copyMaterial(d.Materials.Annotated, filepath.Join(inDir, "annotated.md")); err != nil {
			return err
		}
		mats.Annotated = "annotated.md"
	}
	if d.Materials.Cartographer != "" {
		if err := d.copyMaterial(d.Materials.Cartographer, filepath.Join(inDir, "cartographer.md")); err != nil {
			return err
		}
		mats.Cartographer = "cartographer.md"
	}
	if d.Materials.CartographerScan != "" {
		if err := d.copyMaterial(d.Materials.CartographerScan, filepath.Join(inDir, "cartographer_scan.md")); err != nil {
			return err
		}
		mats.CartographerScan = "cartographer_scan.md"
	}

	spec := runner.StudySpec{
		Assay:  d.Assay,
		ItemID: d.ItemID,
		Model: runner.ModelSpec{
			BaseURL:        d.Model.BaseURL,
			ModelID:        d.Model.ModelID,
			Version:        d.Model.Version,
			Endpoint:       d.Model.Endpoint,
			PromptTemplate: d.Model.PromptTemplate,
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
