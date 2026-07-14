package manifest

import (
	"encoding/json"
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
		Sampling:    assay.Sampling{Temperature: 0.8, Seeds: []int{1, 2}, MaxTokens: 512},
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

func TestComputeRecordsImageStudyAndMaterials(t *testing.T) {
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

// Sampling is declared in study.json, so the run record's StudyDigest moves when
// the sampler changes. This documents that the sampler is part of a run's
// recorded inputs rather than an invisible Go constant.
//
// It is NOT a freeze check — the freeze is retired (INQUIRY.md). A moved digest
// means two runs had different inputs; that is information, not grounds for
// refusal. Note also that it only covers the sampler we DECLARE: top_p, min_p and
// repeat_penalty still inherit server defaults and no digest can see them. The
// fix for that is reading the effective config back from the server, not a
// better hash.
func TestStudyDigestCoversSamplingParams(t *testing.T) {
	base := spec()
	in := writeIn(t, base, map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
	m, err := Compute(in, testImage)
	if err != nil {
		t.Fatal(err)
	}

	for name, mutate := range map[string]func(s *runner.StudySpec){
		"temperature": func(s *runner.StudySpec) { s.Sampling.Temperature = 0.7 },
		"seeds":       func(s *runner.StudySpec) { s.Sampling.Seeds = []int{9, 10} },
		"max tokens":  func(s *runner.StudySpec) { s.Sampling.MaxTokens = 256 },
	} {
		t.Run(name, func(t *testing.T) {
			changed := spec()
			mutate(&changed)
			in2 := writeIn(t, changed, map[string]string{"scenario.md": "S", "glyph.md": "G", "ground.md": "GR"})
			m2, err := Compute(in2, testImage)
			if err != nil {
				t.Fatal(err)
			}
			if m2.StudyDigest == m.StudyDigest {
				t.Fatalf("changing %s left StudyDigest unchanged — sampling is outside the freeze", name)
			}
		})
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
