package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

type fakeClient struct {
	text string
	err  error
	name string
}

func (f *fakeClient) Generate(_ context.Context, _ string, _ model.GenParams) (model.Response, error) {
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{Text: f.text}, nil
}
func (f *fakeClient) Name() string {
	if f.name != "" {
		return f.name
	}
	return "fake-model"
}
func (f *fakeClient) Version() string { return "0.0.0" }

// writeStudy lays out a valid /in dir (study.json + materials) and returns it.
func writeStudy(t *testing.T, spec StudySpec, materials map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	raw, err := json.Marshal(spec)
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

func groundedSpec() StudySpec {
	return StudySpec{
		Assay:       SupportedAssay,
		ItemID:      "casg-direct",
		Model:       ModelSpec{BaseURL: "http://llama-server:8081/v1", ModelID: "qwen", Version: "q4"},
		Conditions:  []assay.Condition{assay.Baseline, assay.GlyphOnly, assay.GroundedGlyph},
		RunsPerCell: 2,
		Materials:   MaterialsSpec{Scenario: "scenario.md", Glyph: "glyph.md", Ground: "ground.md"},
		Sampling:    assay.Sampling{Temperature: 0.8, Seeds: []int{1, 2}, MaxTokens: 512},
	}
}

// validSampling is a deterministic regime for fixtures whose subject is some
// other rule, so they fail for the reason they name rather than on sampling.
func validSampling() assay.Sampling {
	return assay.Sampling{Temperature: 0, MaxTokens: 512}
}

func TestExecuteProducesResultsAndResponses(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND",
	})
	out := t.TempDir()

	results, err := Execute(context.Background(), in, out, &fakeClient{text: "PASS", name: "qwen"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// 3 conditions × 2 runs = 6 rows.
	if len(results.Rows) != 6 {
		t.Fatalf("expected 6 rows, got %d", len(results.Rows))
	}
	if results.ModelID != "qwen" || results.Assay != SupportedAssay {
		t.Fatalf("results identity: %+v", results)
	}

	// results.json written and re-parseable.
	raw, err := os.ReadFile(filepath.Join(out, "results.json"))
	if err != nil {
		t.Fatalf("read results.json: %v", err)
	}
	var back Results
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("results.json not valid JSON: %v", err)
	}
	if len(back.Rows) != 6 {
		t.Fatalf("round-trip rows = %d", len(back.Rows))
	}

	// One response file per condition/run.
	for _, cond := range []assay.Condition{assay.Baseline, assay.GlyphOnly, assay.GroundedGlyph} {
		for run := 1; run <= 2; run++ {
			p := filepath.Join(out, "responses", string(cond)+"_"+itoa(run)+".txt")
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("missing response %s: %v", p, err)
			}
			if string(b) != "PASS" {
				t.Fatalf("response %s = %q", p, b)
			}
		}
	}
}

func TestExecuteIsNotPlaceholder(t *testing.T) {
	// Guards against the registry-lab regression: results.json must carry
	// real scored rows, never a {"status":"pending-adapter"} stub.
	in := writeStudy(t, StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md"}, Sampling: validSampling(),
	}, map[string]string{"scenario.md": "S"})
	out := t.TempDir()

	if _, err := Execute(context.Background(), in, out, &fakeClient{text: "PASS"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(out, "results.json"))
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		t.Fatal(err)
	}
	if _, isStub := probe["status"]; isStub {
		t.Fatal("results.json carries a placeholder status field")
	}
	if _, hasRows := probe["rows"]; !hasRows {
		t.Fatal("results.json has no rows")
	}
}

func TestLoadSpecRejectsUnsupportedAssay(t *testing.T) {
	in := writeStudy(t, StudySpec{
		Assay: "decomposition", ItemID: "i",
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "s.md"},
	}, map[string]string{"s.md": "S"})
	if _, err := LoadSpec(in); err == nil {
		t.Fatal("expected unsupported-assay error")
	}
}

func TestLoadSpecValidations(t *testing.T) {
	cases := []struct {
		name string
		spec StudySpec
	}{
		{"missing item_id", StudySpec{Assay: SupportedAssay, Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1, Materials: MaterialsSpec{Scenario: "s.md"}, Sampling: validSampling()}},
		{"no conditions", StudySpec{Assay: SupportedAssay, ItemID: "i", RunsPerCell: 1, Materials: MaterialsSpec{Scenario: "s.md"}, Sampling: validSampling()}},
		{"zero runs", StudySpec{Assay: SupportedAssay, ItemID: "i", Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 0, Materials: MaterialsSpec{Scenario: "s.md"}, Sampling: validSampling()}},
		{"missing scenario", StudySpec{Assay: SupportedAssay, ItemID: "i", Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1, Sampling: validSampling()}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := writeStudy(t, tc.spec, map[string]string{"s.md": "S"})
			if _, err := LoadSpec(in); err == nil {
				t.Fatalf("expected validation error for %s", tc.name)
			}
		})
	}
}

func TestLoadSpecMissingFileAndBadJSON(t *testing.T) {
	if _, err := LoadSpec(t.TempDir()); err == nil {
		t.Fatal("expected error for missing study.json")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "study.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestExecuteReportsMissingMaterial(t *testing.T) {
	// study.json names a glyph file that isn't present.
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "S", "ground.md": "G",
		// glyph.md deliberately absent
	})
	if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{text: "PASS"}); err == nil {
		t.Fatal("expected missing-material error")
	}
}

func TestExecuteFailsFastOnProbeError(t *testing.T) {
	in := writeStudy(t, StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 3,
		Materials: MaterialsSpec{Scenario: "scenario.md"}, Sampling: validSampling(),
	}, map[string]string{"scenario.md": "S"})
	_, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{err: errContext})
	if err == nil {
		t.Fatal("expected probe error to abort")
	}
}

var errContext = &probeErr{}

type probeErr struct{}

func (*probeErr) Error() string { return "model down" }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
