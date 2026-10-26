package loopstudy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/agentloop"
	"corpos-lab/internal/setupcompletion"
	"corpos-lab/internal/study"
)

// A stamped study loads and validates as written, pins the scaffold's current
// loop image and the loop's call_tokens default, and offers only the tools its
// sandbox can serve.
func TestWriteStampsALoadableStudy(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "S1")
	wrote, err := Write(dir, Options{Name: "specimen-x", ItemID: "L9"})
	if err != nil {
		t.Fatal(err)
	}
	if len(wrote) != 4 {
		t.Fatalf("wrote %v", wrote)
	}
	def, err := study.LoadDef(filepath.Join(dir, StudyFile))
	if err != nil {
		t.Fatalf("LoadDef: %v", err)
	}
	if def.Image != setupcompletion.LoopImg || def.Loop.CallTokens != agentloop.DefaultCallTokens || def.Loop.StepCap != StepCap {
		t.Errorf("def image %s call_tokens %d step_cap %d", def.Image, def.Loop.CallTokens, def.Loop.StepCap)
	}
	if def.Name != "specimen-x" || def.ItemID != "L9" || def.RunsPerCell != 16 || len(def.Sampling.Seeds) != 16 {
		t.Errorf("def = %+v", def)
	}
	pre, _ := os.ReadFile(filepath.Join(dir, PreambleFile))
	if strings.Contains(string(pre), "run_query") {
		t.Error("preamble offers run_query with no query data")
	}
	if err := def.Materialize(t.TempDir()); err != nil {
		t.Errorf("Materialize: %v", err)
	}
}

func TestWriteWithQuery(t *testing.T) {
	dir := t.TempDir()
	if _, err := Write(dir, Options{Name: "q", ItemID: "q", Query: true}); err != nil {
		t.Fatal(err)
	}
	pre, _ := os.ReadFile(filepath.Join(dir, PreambleFile))
	sb, _ := os.ReadFile(filepath.Join(dir, SandboxFile))
	if !strings.Contains(string(pre), "run_query <text>") || !strings.Contains(string(sb), "_db.txt") {
		t.Errorf("query study missing run_query or its data:\n%s\n%s", pre, sb)
	}
}

func TestWriteRefuses(t *testing.T) {
	dir := t.TempDir()
	if _, err := Write(dir, Options{Name: "a", ItemID: "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(dir, Options{Name: "a", ItemID: "b"}); err == nil {
		t.Error("second write: want a refusal")
	}
	for _, o := range []Options{{Name: "", ItemID: "b"}, {Name: "a b", ItemID: "b"}, {Name: "a", ItemID: "\""}} {
		if _, err := Write(t.TempDir(), o); err == nil {
			t.Errorf("%+v: want an error", o)
		}
	}
	file := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(filepath.Join(file, "sub"), Options{Name: "a", ItemID: "b"}); err == nil {
		t.Error("dir under a file: want an error")
	}
}
