package calibrationstudies

import (
	"strings"
	"testing"
)

// cmrCasgMistral is the exact committed content of
// studies/calibrated-mechanism-retyping/defs/cmr-casg-direct-mistral.toml. It
// locks the byte-for-byte output the shell script produced. The \n sequences in
// prompt_template are two literal characters, as in the file.
const cmrCasgMistral = `# calibrated mechanism re-typing (task 4227) — casg-direct — mistral
name = "cmr-casg-direct-s1-mistral"
assay = "grounded-glyph-probe"
item_id = "casg-direct"

image = "localhost/lab-grounded-glyph-probe@sha256:43ac8df6d5ac63eff3c0bddecb881949074374ef9d57a4352c716aeedf9b22d3"
network = "corpos-net"

conditions = ["baseline", "glyph_only", "scrambled_glyph", "off_target_glyph"]
runs_per_cell = 8

[model]
base_url = "http://llama-server:8081/v1"
model_id = "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf"
version = "v0.3-q4km"
endpoint = "completion"
prompt_template = "[INST] {prompt}\n\n---\nNote: You are running as a single completion. There is no harness, no tools, no filesystem, no shell. You cannot call functions or read files. Give your answer directly as text. [/INST]"

[materials]
scenario   = "../casg-direct/materials/scenario.md"
glyph      = "../casg-direct/materials/glyph.md"
scrambled  = "../casg-direct/materials/scrambled_75.md"
off_target = "../casg-direct/materials/off_target_glyph.md"

[sampling]
temperature = 0.8
seeds = [1, 2, 3, 4, 5, 6, 7, 8]
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

func cellByName(cells []Cell, name string) (Cell, bool) {
	for _, c := range cells {
		if c.Name == name {
			return c, true
		}
	}
	return Cell{}, false
}

func TestCMRCells(t *testing.T) {
	cells := CMRCells()
	if len(cells) != 18 {
		t.Fatalf("CMRCells: got %d cells, want 18", len(cells))
	}
	got, ok := cellByName(cells, "cmr-casg-direct-mistral.toml")
	if !ok {
		t.Fatalf("CMRCells: missing cmr-casg-direct-mistral.toml")
	}
	if got.TOML != cmrCasgMistral {
		t.Errorf("cmr-casg-direct-mistral.toml content mismatch:\n%s", got.TOML)
	}
	// Emit order: class outer, model inner.
	if cells[0].Name != "cmr-casg-direct-mistral.toml" ||
		cells[1].Name != "cmr-casg-direct-phi4.toml" ||
		cells[2].Name != "cmr-casg-direct-qwen38.toml" {
		t.Errorf("CMRCells order: got %s, %s, %s", cells[0].Name, cells[1].Name, cells[2].Name)
	}
}

func TestSCBCells(t *testing.T) {
	cells := SCBCells()
	if len(cells) != 54 {
		t.Fatalf("SCBCells: got %d cells, want 54", len(cells))
	}

	ref, ok := cellByName(cells, "scb-casg-direct-ref-mistral.toml")
	if !ok {
		t.Fatalf("SCBCells: missing scb-casg-direct-ref-mistral.toml")
	}
	for _, want := range []string{
		`name = "scb-casg-direct-ref-s1-mistral"`,
		`item_id = "casg-direct"`,
		`conditions = ["baseline", "glyph_only"]`,
		`scenario = "../casg/materials/scenario.md"`,
		`glyph    = "../casg/materials/glyph.md"`,
	} {
		if !strings.Contains(ref.TOML, want) {
			t.Errorf("ref cell missing %q in:\n%s", want, ref.TOML)
		}
	}

	str, ok := cellByName(cells, "scb-casg-direct-str075-qwen38.toml")
	if !ok {
		t.Fatalf("SCBCells: missing scb-casg-direct-str075-qwen38.toml")
	}
	for _, want := range []string{
		`name = "scb-casg-direct-str075-s1-qwen38"`,
		`item_id = "casg-direct-str075"`,
		`conditions = ["scrambled_glyph"]`,
		`scenario  = "../casg/materials/scenario.md"`,
		`scrambled = "../casg/materials/scrambled_75.md"`,
		`model_id = "Qwen3.8-27B-Q4_K_M.gguf"`,
	} {
		if !strings.Contains(str.TOML, want) {
			t.Errorf("strength cell missing %q in:\n%s", want, str.TOML)
		}
	}

	// Emit order per arm+model: ref then the five strengths.
	if cells[0].Name != "scb-casg-direct-ref-mistral.toml" ||
		cells[1].Name != "scb-casg-direct-str000-mistral.toml" ||
		cells[5].Name != "scb-casg-direct-str100-mistral.toml" {
		t.Errorf("SCBCells order: got %s, %s, %s", cells[0].Name, cells[1].Name, cells[5].Name)
	}
}

func TestPromptTemplate(t *testing.T) {
	if promptTemplate("nope") != "" {
		t.Errorf("promptTemplate(unknown): want empty")
	}
	if !strings.HasPrefix(promptTemplate("qwen38"), "<|im_start|>user") {
		t.Errorf("promptTemplate(qwen38): unexpected start")
	}
}
