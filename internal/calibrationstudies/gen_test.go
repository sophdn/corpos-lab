package calibrationstudies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cmrDefs is the committed studies/calibrated-mechanism-retyping/defs dir: the
// byte-for-byte output the shell script produced, which every CMR cell must
// reproduce.
var cmrDefs = filepath.Join("..", "..", "studies", "calibrated-mechanism-retyping", "defs")

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
	if _, ok := cellByName(cells, "cmr-casg-direct-mistral.toml"); !ok {
		t.Fatalf("CMRCells: missing cmr-casg-direct-mistral.toml")
	}
	for _, c := range cells {
		want, err := os.ReadFile(filepath.Join(cmrDefs, c.Name))
		if err != nil {
			t.Fatal(err)
		}
		if c.TOML != string(want) {
			t.Errorf("%s content mismatch with the committed def:\n%s", c.Name, c.TOML)
		}
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
