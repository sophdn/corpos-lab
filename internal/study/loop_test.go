package study

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/runner"
)

// loopDef is a valid agentic-loop definition body.
const loopDef = `
name = "setup-loop-smoke"
assay = "agentic-loop-probe"
item_id = "casg-direct"
image = "localhost/lab-agentic-loop-probe:dev"
network = "corpos-net"
conditions = ["baseline", "glyph_only"]
runs_per_cell = 1

[model]
base_url = "http://llama-server:8081/v1"
model_id = "Qwen3.8-27B-Q4_K_M.gguf"
version = "q4km"
endpoint = "completion"
prompt_template = "<|im_start|>user\n{prompt}<|im_end|>\n<|im_start|>assistant\n"

[materials]
scenario = "scenario.md"
glyph = "glyph.md"

[loop]
preamble = "preamble.md"
sandbox = "sandbox.json"
step_cap = 8
call_tokens = 200

[sampling]
temperature = 0.8
seeds = [1]
max_tokens = 1024
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

func loopMaterials() map[string]string {
	return map[string]string{
		"scenario.md":  "SCENARIO",
		"glyph.md":     "GLYPH",
		"preamble.md":  "PREAMBLE",
		"sandbox.json": `{"CHANGELOG.md":"# Changelog"}`,
	}
}

func TestLoadDefAcceptsLoopAssayAndMaterializes(t *testing.T) {
	d, err := LoadDef(writeDef(t, loopDef, loopMaterials()))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	if d.Assay != runner.LoopAssay {
		t.Fatalf("assay = %q", d.Assay)
	}

	in := t.TempDir()
	if err := d.Materialize(in); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	// preamble + sandbox copied under canonical names.
	for _, name := range []string{"preamble.md", "sandbox.json", "scenario.md", "glyph.md", "study.json"} {
		if _, err := os.Stat(filepath.Join(in, name)); err != nil {
			t.Fatalf("missing materialized %s: %v", name, err)
		}
	}
	// study.json carries the loop spec pointing at the canonical names.
	raw, err := os.ReadFile(filepath.Join(in, "study.json"))
	if err != nil {
		t.Fatal(err)
	}
	var spec runner.StudySpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("study.json: %v", err)
	}
	if spec.Loop.Preamble != "preamble.md" || spec.Loop.Sandbox != "sandbox.json" {
		t.Fatalf("loop spec = %+v", spec.Loop)
	}
	if spec.Loop.StepCap != 8 || spec.Loop.CallTokens != 200 {
		t.Fatalf("loop bounds not carried: %+v", spec.Loop)
	}
	// The materialized spec must still load (round-trip through the runner).
	if _, err := runner.LoadSpec(in); err != nil {
		t.Fatalf("materialized loop spec does not load: %v", err)
	}
}

func TestValidateRejectsLoopAssayWithoutPreamble(t *testing.T) {
	body := strings.Replace(loopDef, "preamble = \"preamble.md\"\n", "", 1)
	if _, err := LoadDef(writeDef(t, body, loopMaterials())); err == nil {
		t.Fatal("expected rejection when loop.preamble is missing")
	}
}

func TestValidateRejectsLoopAssayWithoutSandbox(t *testing.T) {
	body := strings.Replace(loopDef, "sandbox = \"sandbox.json\"\n", "", 1)
	if _, err := LoadDef(writeDef(t, body, loopMaterials())); err == nil {
		t.Fatal("expected rejection when loop.sandbox is missing")
	}
}

func TestMaterializeLoopReportsMissingPreambleFile(t *testing.T) {
	mats := loopMaterials()
	delete(mats, "preamble.md")
	d, err := LoadDef(writeDef(t, loopDef, mats))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	if err := d.Materialize(t.TempDir()); err == nil {
		t.Fatal("expected Materialize to fail when the preamble file is absent")
	}
}
