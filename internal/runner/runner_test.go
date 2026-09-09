package runner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

type fakeClient struct {
	text string
	err  error
	name string
	// resp, when set, is returned instead of a bare text response.
	resp *model.Response
	// props and propsErr drive the /props readback the runner records.
	props    model.ServerProps
	propsErr error
	// gotParams records every GenParams the runner sent, so a test can assert
	// the full sampler chain reached the wire rather than being dropped.
	gotParams []model.GenParams
	// gotPrompts records every assembled prompt, so a test can assert the
	// condition's materials reached the wire in the right shape.
	gotPrompts []string
}

func (f *fakeClient) Generate(_ context.Context, prompt string, p model.GenParams) (model.Response, error) {
	f.gotParams = append(f.gotParams, p)
	f.gotPrompts = append(f.gotPrompts, prompt)
	if f.err != nil {
		return model.Response{}, f.err
	}
	if f.resp != nil {
		return *f.resp, nil
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
func (f *fakeClient) Props(_ context.Context) (model.ServerProps, error) {
	if f.propsErr != nil {
		return model.ServerProps{}, f.propsErr
	}
	return f.props, nil
}

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
		Sampling:    neutralChain(0.8, []int{1, 2}),
	}
}

// neutralChain is a complete sampler regime: min_p the only live truncation
// stage, every other stage pinned to its disabled value. Complete because the
// runner now refuses an undeclared chain — a partial regime here would fail on
// sampling rather than on the rule its test names.
func neutralChain(temp float64, seeds []int) assay.Sampling {
	return assay.Sampling{
		Temperature: temp, Seeds: seeds, MaxTokens: 512,
		TopNSigma: -1.0, TopK: 0, TypicalP: 1.0, TopP: 1.0, MinP: 0.05,
		RepeatPenalty: 1.0, RepeatLastN: 0, PresencePenalty: 0.0, FrequencyPenalty: 0.0,
		XTCProbability: 0.0, DryMultiplier: 0.0,
	}
}

// validSampling is a deterministic regime for fixtures whose subject is some
// other rule, so they fail for the reason they name rather than on sampling.
func validSampling() assay.Sampling {
	return neutralChain(0, nil)
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

// The matched-content T2 condition: the runner loads the imperative material
// and assembles imperative + scenario, carrying no glyph.
func TestExecuteRunsImperativeOnlyCondition(t *testing.T) {
	spec := StudySpec{
		Assay:       SupportedAssay,
		ItemID:      "casg-direct",
		Model:       ModelSpec{ModelID: "qwen"},
		Conditions:  []assay.Condition{assay.ImperativeOnly},
		RunsPerCell: 1,
		Materials:   MaterialsSpec{Scenario: "scenario.md", Imperative: "imperative.md"},
		Sampling:    validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "SCENARIO", "imperative.md": "IMPERATIVE"})
	out := t.TempDir()

	f := &fakeClient{text: "reply", name: "qwen"}
	results, err := Execute(context.Background(), in, out, f)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(results.Rows) != 1 || results.Rows[0].Condition != assay.ImperativeOnly {
		t.Fatalf("rows: %+v", results.Rows)
	}
	// The imperative reached the prompt in the glyph's slot; no glyph leaked in.
	if len(f.gotPrompts) != 1 || f.gotPrompts[0] != "IMPERATIVE\n---\nSCENARIO" {
		t.Fatalf("assembled prompt = %q, want IMPERATIVE\\n---\\nSCENARIO", f.gotPrompts)
	}
	// Response captured under the condition's wire string.
	if _, err := os.ReadFile(filepath.Join(out, "responses", "imperative_only_1.txt")); err != nil {
		t.Fatalf("missing imperative_only response: %v", err)
	}
}

// A study that names imperative_only but ships no imperative material fails
// loudly, rather than assembling a bare scenario as if the T2 arm were empty.
func TestExecuteReportsMissingImperativeMaterial(t *testing.T) {
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.ImperativeOnly}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md", Imperative: "imperative.md"},
		Sampling:  validSampling(),
	}
	// scenario present, imperative.md deliberately absent.
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{text: "x"}); err == nil {
		t.Fatal("expected missing-imperative-material error")
	}
}

// The mechanism-control conditions: the runner loads the scrambled and
// off-target materials and assembles them in the glyph's slot, carrying no glyph.
func TestExecuteRunsControlConditions(t *testing.T) {
	spec := StudySpec{
		Assay:       SupportedAssay,
		ItemID:      "casg-direct",
		Model:       ModelSpec{ModelID: "qwen"},
		Conditions:  []assay.Condition{assay.ScrambledGlyph, assay.OffTargetGlyph},
		RunsPerCell: 1,
		Materials:   MaterialsSpec{Scenario: "scenario.md", Scrambled: "scrambled_glyph.md", OffTarget: "off_target_glyph.md"},
		Sampling:    validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "SCENARIO", "scrambled_glyph.md": "SCR", "off_target_glyph.md": "OTG"})
	out := t.TempDir()

	f := &fakeClient{text: "reply", name: "qwen"}
	results, err := Execute(context.Background(), in, out, f)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(results.Rows) != 2 {
		t.Fatalf("rows: %+v", results.Rows)
	}
	got := map[string]bool{}
	for _, p := range f.gotPrompts {
		got[p] = true
	}
	if !got["SCR\n---\nSCENARIO"] || !got["OTG\n---\nSCENARIO"] {
		t.Fatalf("assembled prompts = %q", f.gotPrompts)
	}
	for _, name := range []string{"scrambled_glyph_1.txt", "off_target_glyph_1.txt"} {
		if _, err := os.ReadFile(filepath.Join(out, "responses", name)); err != nil {
			t.Fatalf("missing response %s: %v", name, err)
		}
	}
}

// A study that names a control condition but ships no material fails loudly.
func TestExecuteReportsMissingControlMaterial(t *testing.T) {
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.ScrambledGlyph}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md", Scrambled: "scrambled_glyph.md"},
		Sampling:  validSampling(),
	}
	// scenario present, scrambled_glyph.md deliberately absent.
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{text: "x"}); err == nil {
		t.Fatal("expected missing-scrambled-material error")
	}
}

// The off-target read-error branch: named off_target material absent.
func TestExecuteReportsMissingOffTargetMaterial(t *testing.T) {
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.OffTargetGlyph}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md", OffTarget: "off_target_glyph.md"},
		Sampling:  validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{text: "x"}); err == nil {
		t.Fatal("expected missing-off-target-material error")
	}
}

// When the client surfaces a rendered prompt (the raw /completion path), the
// runner records it per run so the run is self-describing down to its input.
func TestExecutePersistsRenderedPromptWhenSurfaced(t *testing.T) {
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m", Endpoint: "completion"},
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md"}, Sampling: validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	out := t.TempDir()
	f := &fakeClient{resp: &model.Response{Text: "reply", RenderedPrompt: "WRAPPED-INPUT"}}
	if _, err := Execute(context.Background(), in, out, f); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, "prompts", "baseline_1.txt"))
	if err != nil {
		t.Fatalf("read recorded prompt: %v", err)
	}
	if string(got) != "WRAPPED-INPUT" {
		t.Fatalf("recorded prompt = %q, want WRAPPED-INPUT", got)
	}
}

// The chat path surfaces no rendered prompt (the server templates server-side),
// so the runner writes no prompt file rather than recording a half-truth.
func TestExecuteWritesNoPromptFileInChatMode(t *testing.T) {
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: 1,
		Materials: MaterialsSpec{Scenario: "scenario.md"}, Sampling: validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	out := t.TempDir()
	if _, err := Execute(context.Background(), in, out, &fakeClient{text: "reply"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "prompts", "baseline_1.txt")); !os.IsNotExist(err) {
		t.Fatal("chat mode must not write a prompt file")
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

// A run has to describe itself from the artifact alone. These assert the three
// things results.json gained: the sampler that was actually sent, the server's
// self-report, and the declared-vs-served model comparison nothing performed.
func TestExecuteRecordsSamplerAndServerReadback(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND",
	})
	out := t.TempDir()
	f := &fakeClient{
		text: "reply",
		name: "qwen",
		props: model.ServerProps{
			ModelPath:  "/models/qwen",
			ModelAlias: "qwen",
			BuildInfo:  "b9445-af6528e6d",
			NCtx:       8192,
			Defaults:   map[string]any{"top_k": float64(40)},
		},
	}
	res, err := Execute(context.Background(), in, out, f)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// The complete regime rides with the results, not just temperature.
	if res.Sampler.MinP != 0.05 || res.Sampler.TopK != 0 || res.Sampler.RepeatPenalty != 1.0 {
		t.Errorf("sampler not recorded with results: %+v", res.Sampler)
	}
	if res.Server.BuildInfo != "b9445-af6528e6d" || res.Server.NCtx != 8192 {
		t.Errorf("server readback = %+v", res.Server)
	}
	if res.ServerReadbackError != "" {
		t.Errorf("unexpected readback error: %q", res.ServerReadbackError)
	}
	// Declared "qwen", served "qwen" — no divergence to report.
	if res.ModelMismatch != "" {
		t.Errorf("model mismatch on matching models: %q", res.ModelMismatch)
	}

	// And it survives the trip to disk, which is the artifact that outlives us.
	raw, err := os.ReadFile(filepath.Join(out, "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var onDisk Results
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		t.Fatal(err)
	}
	if onDisk.Server.BuildInfo != "b9445-af6528e6d" || onDisk.Sampler.MinP != 0.05 {
		t.Errorf("results.json lost the self-description: %+v", onDisk)
	}
}

// The check nothing performed. A study named one model, the server served
// another, and the record read as if it got what it asked for — which is how a
// Mistral study ran without anyone being able to tell what answered.
func TestExecuteRecordsModelMismatchWithoutRefusingTheRun(t *testing.T) {
	spec := groundedSpec()
	spec.Model.ModelID = "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf"
	in := writeStudy(t, spec, map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND",
	})
	f := &fakeClient{
		text: "reply",
		props: model.ServerProps{
			ModelPath:  "/models/Qwen2.5-32B-Instruct-Q4_K_M.gguf",
			ModelAlias: "Qwen2.5-32B-Instruct-Q4_K_M.gguf",
		},
	}
	res, err := Execute(context.Background(), in, t.TempDir(), f)
	// RECORDED, NEVER ENFORCED — refusing here would be the freeze again.
	if err != nil {
		t.Fatalf("a model mismatch must not fail the run: %v", err)
	}
	if res.ModelMismatch == "" {
		t.Fatal("serving Qwen for a declared Mistral must be recorded as a divergence")
	}
	if !strings.Contains(res.ModelMismatch, "Mistral") || !strings.Contains(res.ModelMismatch, "Qwen") {
		t.Errorf("mismatch should name both models, got %q", res.ModelMismatch)
	}
	if len(res.Rows) == 0 {
		t.Error("the rows are real either way; they must still be returned")
	}
}

// A path matching the declared id is not a mismatch: llama-server reports an
// absolute /models path while a study names the bare artifact.
func TestExecuteTreatsPathBasenameMatchAsNoMismatch(t *testing.T) {
	spec := groundedSpec()
	spec.Model.ModelID = "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf"
	in := writeStudy(t, spec, map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "R",
	})
	f := &fakeClient{
		text:  "reply",
		props: model.ServerProps{ModelPath: "/models/Mistral-7B-Instruct-v0.3.Q4_K_M.gguf", ModelAlias: "mistral"},
	}
	res, err := Execute(context.Background(), in, t.TempDir(), f)
	if err != nil {
		t.Fatal(err)
	}
	if res.ModelMismatch != "" {
		t.Errorf("basename match must not read as divergence: %q", res.ModelMismatch)
	}
}

// A server that won't describe itself degrades the record; it does not void the
// run. The rows are real either way, and the silence is recorded rather than
// filled in with a guess.
func TestExecuteRecordsReadbackFailureWithoutFailingTheRun(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "R",
	})
	f := &fakeClient{text: "reply", propsErr: errors.New("props unreachable")}
	res, err := Execute(context.Background(), in, t.TempDir(), f)
	if err != nil {
		t.Fatalf("a failed readback must not fail the run: %v", err)
	}
	if !strings.Contains(res.ServerReadbackError, "props unreachable") {
		t.Errorf("readback error not recorded: %q", res.ServerReadbackError)
	}
	if res.Server.BuildInfo != "" {
		t.Error("a failed readback must leave the server report empty, not invented")
	}
	if len(res.Rows) == 0 {
		t.Error("rows must survive a failed readback")
	}
}

// The complete chain must reach every generation, not just be recorded beside
// it — recording a sampler the run didn't use is the exact false-confidence the
// record exists to prevent.
func TestExecuteSendsFullChainOnEveryGeneration(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "R",
	})
	f := &fakeClient{text: "reply"}
	if _, err := Execute(context.Background(), in, t.TempDir(), f); err != nil {
		t.Fatal(err)
	}
	if len(f.gotParams) == 0 {
		t.Fatal("no generations recorded")
	}
	for i, p := range f.gotParams {
		for _, c := range []struct {
			field string
			isNil bool
		}{
			{"temperature", p.Temperature == nil},
			{"max_tokens", p.MaxTokens == nil},
			{"top_n_sigma", p.TopNSigma == nil},
			{"top_k", p.TopK == nil},
			{"typical_p", p.TypicalP == nil},
			{"top_p", p.TopP == nil},
			{"min_p", p.MinP == nil},
			{"repeat_penalty", p.RepeatPenalty == nil},
			{"repeat_last_n", p.RepeatLastN == nil},
			{"presence_penalty", p.PresencePenalty == nil},
			{"frequency_penalty", p.FrequencyPenalty == nil},
			{"xtc_probability", p.XTCProbability == nil},
			{"dry_multiplier", p.DryMultiplier == nil},
		} {
			if c.isNil {
				t.Errorf("generation %d left %s unset — it would inherit the server default", i, c.field)
			}
		}
	}
}

// The container's own gate against an undeclared chain: repeat_penalty is
// neutral at 1.0 and has no valid zero, so a zero is the tell that study.json
// carried no sampler at all.
func TestValidateRejectsUndeclaredSamplerChain(t *testing.T) {
	spec := groundedSpec()
	spec.Sampling.RepeatPenalty = 0
	in := writeStudy(t, spec, map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "R",
	})
	_, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{text: "x"})
	if err == nil || !strings.Contains(err.Error(), "repeat_penalty") {
		t.Fatalf("err = %v, want a refusal naming repeat_penalty", err)
	}
}
