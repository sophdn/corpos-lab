package study

import (
	"encoding/json"
	"os"
	"path/filepath"
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
`

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
		"no name":     `assay="grounded-glyph-probe"` + "\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n",
		"no image":    "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n",
		"no base_url": "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n",
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
	bad := "name=\"n\"\nassay=\"decomposition\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
	if _, err := LoadDef(writeDef(t, bad, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected unsupported-assay error")
	}
	zero := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=0\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
	if _, err := LoadDef(writeDef(t, zero, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected zero-runs error")
	}
}

func TestValidateRejectsConditionMaterialGaps(t *testing.T) {
	// glyph_only without a glyph material.
	noGlyph := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"glyph_only\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
	if _, err := LoadDef(writeDef(t, noGlyph, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected glyph-required error")
	}
	// grounded_glyph with glyph but no ground.
	noGround := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"grounded_glyph\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\nglyph=\"g.md\"\n"
	if _, err := LoadDef(writeDef(t, noGround, map[string]string{"s.md": "S", "g.md": "G"})); err == nil {
		t.Fatal("expected ground-required error")
	}
	// unknown condition.
	unknown := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"teleport\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
	if _, err := LoadDef(writeDef(t, unknown, map[string]string{"s.md": "S"})); err == nil {
		t.Fatal("expected unknown-condition error")
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
	body := "name=\"n\"\nassay=\"grounded-glyph-probe\"\nitem_id=\"i\"\nimage=\"x\"\nconditions=[\"baseline\"]\nruns_per_cell=1\n[model]\nbase_url=\"u\"\nmodel_id=\"m\"\n[materials]\nscenario=\"s.md\"\n"
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
