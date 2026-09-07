package study

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/runner"
)

// validDef is the canonical grounded-glyph-probe definition body, with
// material files written alongside so relative paths resolve.
const validDef = `
name = "casg-direct-v3-smoke"
assay = "grounded-glyph-probe"
item_id = "casg-direct"
image = "localhost/lab-grounded-glyph-probe:dev"
network = "corpos-net"
conditions = ["baseline", "glyph_only", "grounded_glyph"]
runs_per_cell = 2

[model]
base_url = "http://llama-server:8081/v1"
model_id = "Qwen2.5-32B-Instruct-Q4_K_M.gguf"
version = "q4km"

[materials]
scenario = "scenario.md"
glyph = "glyph.md"
ground = "ground.md"

[sampling]
temperature = 0.8
seeds = [1, 2]
max_tokens = 512
top_n_sigma = -1.0
top_k = 0
typical_p = 1.0
top_p = 1.0
min_p = 0.05
repeat_penalty = 1.0
repeat_last_n = 0
presence_penalty = 0.0
frequency_penalty = 0.0
xtc_probability = 0.0
dry_multiplier = 0.0
`

// samplerChain is the neutral sampler declaration every valid fixture needs:
// one truncation stage live (min_p), every other stage pinned to its disabled
// value. Appended to fixtures rather than inlined so the chain-completeness
// rule has exactly one definition in the tests.
const samplerChain = "top_n_sigma=-1.0\ntop_k=0\ntypical_p=1.0\ntop_p=1.0\nmin_p=0.05\n" +
	"repeat_penalty=1.0\nrepeat_last_n=0\npresence_penalty=0.0\nfrequency_penalty=0.0\n" +
	"xtc_probability=0.0\ndry_multiplier=0.0\n"

// samplingBlock is a minimal VALID sampling declaration, appended to fixtures
// whose subject is some other validation rule. Without it those fixtures would
// trip the sampling check first and pass for the wrong reason — asserting only
// that "an" error occurred, not the one the test names.
const samplingBlock = "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + samplerChain

// samplingChainMinus returns a complete [sampling] block with exactly one
// stage omitted, so each chain case names the single omission it expects to be
// rejected instead of relying on a wholesale-missing block.
func samplingChainMinus(t *testing.T, field string) string {
	t.Helper()
	return "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + rewriteChain(t, field, "")
}

// chainWith returns the full chain with one stage set to val.
func chainWith(t *testing.T, field, val string) string {
	t.Helper()
	return rewriteChain(t, field, val)
}

// rewriteChain drops field (val == "") or reassigns it. It fails the test if
// field is absent from samplerChain — otherwise a renamed stage would silently
// turn its case into a no-op that passes for the wrong reason.
func rewriteChain(t *testing.T, field, val string) string {
	t.Helper()
	var b strings.Builder
	found := false
	for _, line := range strings.Split(strings.TrimSpace(samplerChain), "\n") {
		if strings.HasPrefix(line, field+"=") {
			found = true
			if val != "" {
				b.WriteString(field + "=" + val + "\n")
			}
			continue
		}
		b.WriteString(line + "\n")
	}
	if !found {
		t.Fatalf("samplerChain declares no %q — fixture and rule have drifted apart", field)
	}
	return b.String()
}

func writeDef(t *testing.T, body string, materials map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "study.toml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, content := range materials {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func allMaterials() map[string]string {
	return map[string]string{"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND"}
}

func TestLoadDefParsesAndValidates(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	if d.Name != "casg-direct-v3-smoke" || d.Assay != "grounded-glyph-probe" {
		t.Fatalf("parsed def: %+v", d)
	}
	if d.RunsPerCell != 2 || len(d.Conditions) != 3 {
		t.Fatalf("parsed counts: %+v", d)
	}
	if d.Model.ModelID != "Qwen2.5-32B-Instruct-Q4_K_M.gguf" || d.Network != "corpos-net" {
		t.Fatalf("parsed model/network: %+v", d)
	}
}

func TestLoadDefRejectsMissingFile(t *testing.T) {
	if _, err := LoadDef(filepath.Join(t.TempDir(), "nope.toml")); err == nil {
		t.Fatal("expected error for missing def")
	}
}

func TestLoadDefRejectsBadTOML(t *testing.T) {
	if _, err := LoadDef(writeDef(t, "name = = bad", nil)); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	cases := map[string]string{
		"no name":     `assay="grounded-glyph-probe"` + "\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock,
		"no image":    "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock,
		"no base_url": "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "S"})); err == nil {
				t.Fatalf("expected validation error for %s", name)
			}
		})
	}
}

func TestValidateRejectsUnsupportedAssayAndZeroRuns(t *testing.T) {
	bad := "name=\"n\"\nassay=\"decomposition\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, bad, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected unsupported-assay error")
	}
	zero := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=0\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, zero, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected zero-runs error")
	}
}

func TestValidateRejectsConditionMaterialGaps(t *testing.T) {
	// glyph_only without a glyph material.
	noGlyph := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"glyph_only\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, noGlyph, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected glyph-required error")
	}
	// grounded_glyph with glyph but no ground.
	noGround := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"grounded_glyph\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\nglyph=\"g.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, noGround, map[string]string{"s.md": "S", "g.md": "G"})); err == nil {
		t.Fatal("expected ground-required error")
	}
	// imperative_only without an imperative material.
	noImperative := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"imperative_only\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, noImperative, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected imperative-required error")
	}
	// unknown condition.
	unknown := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"teleport\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	if _, err := LoadDef(writeDef(t, unknown, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected unknown-condition error")
	}
}

func TestValidateCompletionEndpoint(t *testing.T) {
	base := func(model string) string {
		return "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\n" +
			"conditions=[\"baseline\"]\nruns_per_cell=1\n[model]\n" + model + "\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	}
	mats := map[string]string{"s.md": "S"}

	// completion mode without a {prompt} placeholder is rejected.
	noPlaceholder := base("base_url=\"u\"\nmodel_id=\"m\"\nendpoint=\"completion\"\nprompt_template=\"no placeholder here\"")
	if _, err := LoadDef(writeDef(t, noPlaceholder, mats)); err == nil {
		t.Fatal("expected error for completion without a {prompt} placeholder")
	}
	// an unknown endpoint is rejected.
	unknown := base("base_url=\"u\"\nmodel_id=\"m\"\nendpoint=\"telepathy\"")
	if _, err := LoadDef(writeDef(t, unknown, mats)); err == nil {
		t.Fatal("expected error for unknown endpoint")
	}
	// completion mode with a valid wrapper loads.
	ok := base("base_url=\"u\"\nmodel_id=\"m\"\nendpoint=\"completion\"\nprompt_template=\"<a>{prompt}<b>\"")
	if _, err := LoadDef(writeDef(t, ok, mats)); err != nil {
		t.Fatalf("valid completion def rejected: %v", err)
	}
	// the default (no endpoint) is chat and needs no template.
	if _, err := LoadDef(writeDef(t, base("base_url=\"u\"\nmodel_id=\"m\""), mats)); err != nil {
		t.Fatalf("default chat def rejected: %v", err)
	}
}

func TestMaterializeCarriesEndpointAndPromptTemplate(t *testing.T) {
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\n" +
		"conditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n" +
		"endpoint=\"completion\"\nprompt_template=\"<|im_start|>user\\n{prompt}<|im_end|>\\n<|im_start|>assistant\\n\"\n" +
		"[materials]\nscenario=\"s.md\"\n" + samplingBlock
	d, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "S"}))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	inDir := filepath.Join(t.TempDir(), "in")
	if err := d.Materialize(inDir); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	spec, err := runner.LoadSpec(inDir)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	if spec.Model.Endpoint != "completion" {
		t.Fatalf("endpoint = %q, want completion", spec.Model.Endpoint)
	}
	if !strings.Contains(spec.Model.PromptTemplate, "{prompt}") {
		t.Fatalf("prompt_template not carried: %q", spec.Model.PromptTemplate)
	}
}

// The matched-content T2 condition materializes the imperative into the
// container contract as imperative.md, alongside the scenario, with no glyph.
func TestMaterializeCopiesImperativeMaterial(t *testing.T) {
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"casg-direct\"\n" +
		"image=\"x\"\nconditions=[\"baseline\",\"imperative_only\"]\nruns_per_cell=1\n" +
		"[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n" +
		"[materials]\nscenario=\"s.md\"\nimperative=\"imp.md\"\n" + samplingBlock
	d, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "SCENARIO", "imp.md": "IMPERATIVE"}))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	inDir := filepath.Join(t.TempDir(), "in")
	if err := d.Materialize(inDir); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(inDir, "imperative.md"))
	if err != nil {
		t.Fatalf("read imperative.md: %v", err)
	}
	if string(got) != "IMPERATIVE" {
		t.Fatalf("imperative.md content = %q, want IMPERATIVE", got)
	}
	spec, err := runner.LoadSpec(inDir)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	if spec.Materials.Imperative != "imperative.md" {
		t.Fatalf("spec imperative = %q, want imperative.md", spec.Materials.Imperative)
	}
	if spec.Materials.Glyph != "" {
		t.Fatalf("imperative_only study must carry no glyph, got %q", spec.Materials.Glyph)
	}
	// A glyph must not have been written to the contract dir.
	if _, err := os.Stat(filepath.Join(inDir, "glyph.md")); !os.IsNotExist(err) {
		t.Fatal("glyph.md should not exist for an imperative_only study")
	}
}

// An unrecorded sampling regime is unreproducible, which is the exact gap the
// v3 study left behind: its study.json never recorded a temperature, so the
// original value is now permanently unknowable. Every rule here refuses a
// silent default rather than guessing one.
func TestValidateRejectsUnderspecifiedSampling(t *testing.T) {
	base := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=2\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
	cases := map[string]string{
		// The headline rule: silence is not consent to a default.
		"no sampling block at all": "",
		"no temperature":           "\n[sampling]\nmax_tokens=512\n" + samplerChain,
		"no max_tokens":            "\n[sampling]\ntemperature=0.0\n" + samplerChain,
		"negative temperature":     "\n[sampling]\ntemperature=-1\nmax_tokens=512\n" + samplerChain,
		"zero max_tokens":          "\n[sampling]\ntemperature=0.0\nmax_tokens=0\n" + samplerChain,
		// Sampled runs are only reproducible when every replicate is seeded.
		"sampled but unseeded":   "\n[sampling]\ntemperature=0.8\nmax_tokens=512\n" + samplerChain,
		"too few seeds for runs": "\n[sampling]\ntemperature=0.8\nseeds=[1]\nmax_tokens=512\n" + samplerChain,
		"too many seeds":         "\n[sampling]\ntemperature=0.8\nseeds=[1,2,3]\nmax_tokens=512\n" + samplerChain,
		// Two replicates on one seed are the same replicate twice.
		"duplicate seeds": "\n[sampling]\ntemperature=0.8\nseeds=[7,7]\nmax_tokens=512\n" + samplerChain,
		// Seeds under greedy decoding promise a variation that cannot happen.
		"seeds with greedy decoding": "\n[sampling]\ntemperature=0.0\nseeds=[1,2]\nmax_tokens=512\n" + samplerChain,

		// The chain rule: temperature does not define a distribution on its
		// own, so every stage must be named — the same "silence is not consent"
		// rule temperature already had, applied to the stages that were exempt.
		// One case per stage: each omits exactly that stage and declares the
		// rest, so a rule that silently stopped covering one field would fail
		// here rather than hide behind its neighbours.
		"chain: only temperature declared": "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n",
		"chain: no top_n_sigma":            samplingChainMinus(t, "top_n_sigma"),
		"chain: no top_k":                  samplingChainMinus(t, "top_k"),
		"chain: no typical_p":              samplingChainMinus(t, "typical_p"),
		"chain: no top_p":                  samplingChainMinus(t, "top_p"),
		"chain: no min_p":                  samplingChainMinus(t, "min_p"),
		"chain: no repeat_penalty":         samplingChainMinus(t, "repeat_penalty"),
		"chain: no repeat_last_n":          samplingChainMinus(t, "repeat_last_n"),
		"chain: no presence_penalty":       samplingChainMinus(t, "presence_penalty"),
		"chain: no frequency_penalty":      samplingChainMinus(t, "frequency_penalty"),
		"chain: no xtc_probability":        samplingChainMinus(t, "xtc_probability"),
		"chain: no dry_multiplier":         samplingChainMinus(t, "dry_multiplier"),

		// Declared-but-nonsensical is still rejected: naming a stage is not the
		// same as naming it a coherent value.
		"chain: negative top_k":         "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "top_k", "-1"),
		"chain: zero repeat_penalty":    "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "repeat_penalty", "0.0"),
		"chain: negative repeat_last_n": "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "repeat_last_n", "-5"),
		"chain: top_p above 1":          "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "top_p", "1.5"),
		"chain: min_p below 0":          "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "min_p", "-0.1"),
		"chain: typical_p above 1":      "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "typical_p", "2.0"),
		"chain: xtc_probability above 1": "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" +
			chainWith(t, "xtc_probability", "1.1"),
		"chain: negative dry_multiplier": "\n[sampling]\ntemperature=0.0\nmax_tokens=512\n" + chainWith(t, "dry_multiplier", "-1.0"),
	}
	for name, samp := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadDef(writeDef(t, base+samp, map[string]string{"s.md": "S"})); err == nil {
				t.Fatalf("expected sampling validation error for %q", name)
			}
		})
	}
}

func TestValidateAcceptsExplicitDeterministicMode(t *testing.T) {
	// Greedy single-shot stays available — but chosen, never inherited.
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=2\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n[sampling]\ntemperature=0.0\nmax_tokens=256\n" + samplerChain
	d, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "S"}))
	if err != nil {
		t.Fatalf("deterministic mode must remain selectable: %v", err)
	}
	if !d.sampling().Deterministic() {
		t.Fatal("temperature 0 should yield a deterministic regime")
	}
}

// The declared chain has to survive the trip into the container contract, or
// the study.toml describes a sampler the run never used. This is the assertion
// that makes "the params in the record are the params that ran" true: the
// values here are deliberately distinct from both the server's defaults and
// each other, so a field silently dropped, defaulted, or crossed with its
// neighbour shows up as a wrong number rather than a plausible one.
func TestMaterializeCarriesFullSamplerChain(t *testing.T) {
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\n" +
		"runs_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" +
		"[sampling]\ntemperature=0.7\nseeds=[3]\nmax_tokens=128\n" +
		"top_n_sigma=-1.0\ntop_k=11\ntypical_p=0.66\ntop_p=0.88\nmin_p=0.04\n" +
		"repeat_penalty=1.07\nrepeat_last_n=33\npresence_penalty=0.22\nfrequency_penalty=0.11\n" +
		"xtc_probability=0.15\ndry_multiplier=0.44\n"
	d, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "S"}))
	if err != nil {
		t.Fatal(err)
	}
	inDir := t.TempDir()
	if err := d.Materialize(inDir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(inDir, "study.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec runner.StudySpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}
	got := spec.Sampling
	for _, c := range []struct {
		field string
		got   any
		want  any
	}{
		{"temperature", got.Temperature, 0.7},
		{"max_tokens", got.MaxTokens, 128},
		{"top_n_sigma", got.TopNSigma, -1.0},
		{"top_k", got.TopK, 11},
		{"typical_p", got.TypicalP, 0.66},
		{"top_p", got.TopP, 0.88},
		{"min_p", got.MinP, 0.04},
		{"repeat_penalty", got.RepeatPenalty, 1.07},
		{"repeat_last_n", got.RepeatLastN, 33},
		{"presence_penalty", got.PresencePenalty, 0.22},
		{"frequency_penalty", got.FrequencyPenalty, 0.11},
		{"xtc_probability", got.XTCProbability, 0.15},
		{"dry_multiplier", got.DryMultiplier, 0.44},
	} {
		if c.got != c.want {
			t.Errorf("study.json sampling.%s = %v, want %v", c.field, c.got, c.want)
		}
	}
	if len(got.Seeds) != 1 || got.Seeds[0] != 3 {
		t.Errorf("study.json sampling.seeds = %v, want [3]", got.Seeds)
	}
}

func TestMaterializeWritesContainerContract(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatal(err)
	}
	inDir := filepath.Join(t.TempDir(), "in")
	if err := d.Materialize(inDir); err != nil {
		t.Fatalf("Materialize: %v", err)
	}

	// study.json parses back into the runner spec with canonical local names.
	raw, err := os.ReadFile(filepath.Join(inDir, "study.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec runner.StudySpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("study.json invalid: %v", err)
	}
	if spec.Materials.Scenario != "scenario.md" || spec.Materials.Glyph != "glyph.md" || spec.Materials.Ground != "ground.md" {
		t.Fatalf("materials not canonicalized: %+v", spec.Materials)
	}
	if spec.Assay != "grounded-glyph-probe" || spec.ItemID != "casg-direct" || spec.RunsPerCell != 2 {
		t.Fatalf("spec fields: %+v", spec)
	}

	// Sampling must reach the container, not stay behind on the host: the
	// executor is what applies it.
	if spec.Sampling.Temperature != 0.8 || spec.Sampling.MaxTokens != 512 {
		t.Fatalf("sampling not materialized: %+v", spec.Sampling)
	}
	if len(spec.Sampling.Seeds) != 2 || spec.Sampling.Seeds[0] != 1 || spec.Sampling.Seeds[1] != 2 {
		t.Fatalf("seed sequence not materialized: %+v", spec.Sampling.Seeds)
	}

	// Material files copied with content.
	for name, want := range map[string]string{"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND"} {
		got, err := os.ReadFile(filepath.Join(inDir, name))
		if err != nil {
			t.Fatalf("missing material %s: %v", name, err)
		}
		if string(got) != want {
			t.Fatalf("material %s = %q, want %q", name, got, want)
		}
	}
}

func TestMaterializeBaselineOnlyOmitsOptionalMaterials(t *testing.T) {
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n" + samplingBlock
	d, err := LoadDef(writeDef(t, body, map[string]string{"s.md": "S"}))
	if err != nil {
		t.Fatal(err)
	}
	inDir := filepath.Join(t.TempDir(), "in")
	if err := d.Materialize(inDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(inDir, "glyph.md")); !os.IsNotExist(err) {
		t.Fatal("glyph.md should not exist for a baseline-only study")
	}
}

func TestMaterializeFailsWhenInDirUncreatable(t *testing.T) {
	d, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatal(err)
	}
	// A file where the in dir's parent should be makes MkdirAll fail.
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := d.Materialize(filepath.Join(blocker, "in")); err == nil {
		t.Fatal("expected error when in dir is uncreatable")
	}
}

func TestMaterializeReportsMissingScenario(t *testing.T) {
	// Only glyph + ground present; scenario (the first copy) is absent.
	d, err := LoadDef(writeDef(t, validDef, map[string]string{"glyph.md": "G", "ground.md": "GR"}))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Materialize(filepath.Join(t.TempDir(), "in")); err == nil {
		t.Fatal("expected missing-scenario error")
	}
}

func TestMaterializeReportsMissingGlyph(t *testing.T) {
	// scenario + ground present; glyph absent (needed by glyph_only/grounded).
	d, err := LoadDef(writeDef(t, validDef, map[string]string{"scenario.md": "S", "ground.md": "GR"}))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Materialize(filepath.Join(t.TempDir(), "in")); err == nil {
		t.Fatal("expected missing-glyph error")
	}
}

func TestMaterializeReportsMissingMaterialFile(t *testing.T) {
	// Def names ground.md but it isn't on disk.
	d, err := LoadDef(writeDef(t, validDef, map[string]string{"scenario.md": "S", "glyph.md": "G"}))
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Materialize(filepath.Join(t.TempDir(), "in")); err == nil {
		t.Fatal("expected missing-material error")
	}
}
