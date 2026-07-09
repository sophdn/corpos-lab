package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/runner"
)

func spec() runner.StudySpec {
	return runner.StudySpec{
		Assay:       runner.SupportedAssay,
		ItemID:      "casg-direct",
		Model:       runner.ModelSpec{BaseURL: "http://llama-server:8081/v1", ModelID: "qwen", Version: "q4"},
		Conditions:  []assay.Condition{assay.Baseline, assay.GlyphOnly, assay.GroundedGlyph},
		RunsPerCell: 2,
		Materials:   runner.MaterialsSpec{Scenario: "scenario.md", Glyph: "glyph.md", Ground: "ground.md"},
	}
}

func writeIn(t *testing.T, s runner.StudySpec, materials map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "study.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	for name, content := range materials {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

var testImage = ImagePin{Ref: "localhost/lab-grounded-glyph-probe:dev", Digest: "sha256:aaaa1111bbbb2222"}

func TestComputePinsImageStudyAndMaterials(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	m, err := Compute(in, testImage)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if m.Image.Digest != testImage.Digest || m.Assay != runner.SupportedAssay {
		t.Fatalf("manifest identity: %+v", m)
	}
	if m.StudyDigest == "" {
		t.Fatal("study digest not set")
	}
	if len(m.Materials) != 3 {
		t.Fatalf("expected 3 material digests, got %d", len(m.Materials))
	}
	if m.Materials["scenario.md"] == m.Materials["glyph.md"] {
		t.Fatal("distinct material contents should have distinct digests")
	}
	if m.ModelID != "qwen" || m.ModelVersion != "q4" {
		t.Fatalf("model identity: %+v", m)
	}
}

func TestVerifyPassesOnIdenticalInputs(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(in, testImage, pinned); err != nil {
		t.Fatalf("verify should pass on identical inputs: %v", err)
	}
}

func TestVerifyRefusesChangedMaterial(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate a material by one byte — the freeze rule must catch it.
	if err := os.WriteFile(filepath.Join(in, "glyph.md"), []byte("G-EDITED"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = Verify(in, testImage, pinned)
	if err == nil {
		t.Fatal("expected hard refusal on changed material")
	}
	var mm *MismatchError
	if !errors.As(err, &mm) {
		t.Fatalf("expected MismatchError, got %T", err)
	}
	if !containsSubstr(mm.Divergences, `material "glyph.md" content`) {
		t.Fatalf("divergence should name glyph.md: %v", mm.Divergences)
	}
}

func TestVerifyRefusesChangedImageDigest(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt := ImagePin{Ref: testImage.Ref, Digest: "sha256:ffff9999eeee8888"}
	err = Verify(in, rebuilt, pinned)
	if err == nil {
		t.Fatal("expected hard refusal on changed image digest")
	}
	var mm *MismatchError
	if !errors.As(err, &mm) {
		t.Fatalf("expected MismatchError, got %T", err)
	}
	if !containsSubstr(mm.Divergences, "image digest") {
		t.Fatalf("divergence should name image digest: %v", mm.Divergences)
	}
}

func TestVerifyRefusesChangedStudyJSON(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	// Bump runs_per_cell — a study.json edit must invalidate the pin.
	changed := spec()
	changed.RunsPerCell = 8
	raw, _ := json.Marshal(changed)
	if err := os.WriteFile(filepath.Join(in, "study.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	err = Verify(in, testImage, pinned)
	if err == nil {
		t.Fatal("expected hard refusal on changed study.json")
	}
	var mm *MismatchError
	if !errors.As(err, &mm) {
		t.Fatalf("expected MismatchError, got %T", err)
	}
}

func TestVerifyReportsAddedAndRemovedMaterials(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	// Pin an extra material that the run won't have, and drop one the run has.
	pinned.Materials["phantom.md"] = "sha-of-nothing"
	delete(pinned.Materials, "ground.md")

	err = Verify(in, testImage, pinned)
	if err == nil {
		t.Fatal("expected refusal")
	}
	var mm *MismatchError
	if !errors.As(err, &mm) {
		t.Fatalf("expected MismatchError, got %T", err)
	}
	if !containsSubstr(mm.Divergences, `material "phantom.md": pinned but missing`) {
		t.Fatalf("should report missing pinned material: %v", mm.Divergences)
	}
	if !containsSubstr(mm.Divergences, `material "ground.md": present in run but not pinned`) {
		t.Fatalf("should report unpinned run material: %v", mm.Divergences)
	}
}

func TestVerifyRefusesChangedModelIdentity(t *testing.T) {
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	pinned, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}
	// Swap the model in study.json — the pin must reject a different model.
	changed := spec()
	changed.Model.ModelID = "different-model"
	changed.Model.Version = "v99"
	raw, _ := json.Marshal(changed)
	if err := os.WriteFile(filepath.Join(in, "study.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	err = Verify(in, testImage, pinned)
	if err == nil {
		t.Fatal("expected refusal on changed model")
	}
	var mm *MismatchError
	if !errors.As(err, &mm) {
		t.Fatalf("expected MismatchError, got %T", err)
	}
	if !containsSubstr(mm.Divergences, "model_id") || !containsSubstr(mm.Divergences, "model_version") {
		t.Fatalf("should report model divergences: %v", mm.Divergences)
	}
}

func TestShortTruncatesLongAndPassesTiny(t *testing.T) {
	if got := short("abc"); got != "abc" {
		t.Fatalf("short tiny = %q, want abc", got)
	}
	long := "sha256:" + "abcdef0123456789"
	if got := short(long); got != long[:12] {
		t.Fatalf("short long = %q, want %q", got, long[:12])
	}
}

func TestMismatchErrorMessageListsDivergences(t *testing.T) {
	e := &MismatchError{Divergences: []string{"a", "b"}}
	msg := e.Error()
	if !containsStr(msg, "2 divergence") || !containsStr(msg, "- a") || !containsStr(msg, "- b") {
		t.Fatalf("message: %q", msg)
	}
}

func TestComputePropagatesSpecErrors(t *testing.T) {
	if _, err := Compute(t.TempDir(), testImage); err == nil {
		t.Fatal("expected error for missing study.json")
	}
	// study.json names a material that isn't present.
	in := writeIn(t, spec(), map[string]string{"scenario.md": "S", "glyph.md": "G"}) // ground.md missing
	if _, err := Compute(in, testImage); err == nil {
		t.Fatal("expected error for missing material file")
	}
}

func containsSubstr(list []string, sub string) bool {
	for _, s := range list {
		if containsStr(s, sub) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
