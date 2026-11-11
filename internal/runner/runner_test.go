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
	"corpos-lab/internal/testutil"
)

// fakeClient is the shared scripted model client.
type fakeClient = testutil.FakeClient

// writeStudy lays out a valid /in dir (study.json + materials) and returns it.
func writeStudy(t *testing.T, spec StudySpec, materials map[string]string) string {
	t.Helper()
	return testutil.WriteStudyDir(t, spec, materials)
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
	return testutil.NeutralSampler[assay.Sampling](temp, seeds)
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

	results, err := Execute(context.Background(), in, out, &fakeClient{Text: "PASS", ModelName: "qwen"})
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
	results, f, out := runConditions(t, []assay.Condition{assay.ImperativeOnly},
		MaterialsSpec{Scenario: "scenario.md", Imperative: "imperative.md"},
		map[string]string{"scenario.md": "SCENARIO", "imperative.md": "IMPERATIVE"})
	if len(results.Rows) != 1 || results.Rows[0].Condition != assay.ImperativeOnly {
		t.Fatalf("rows: %+v", results.Rows)
	}
	// The imperative reached the prompt in the glyph's slot; no glyph leaked in.
	assertPrompts(t, f, "IMPERATIVE\n---\nSCENARIO")
	// Response captured under the condition's wire string.
	assertResponses(t, out, "imperative_only_1.txt")
}

// The mechanism-control conditions: the runner loads the scrambled and
// off-target materials and assembles them in the glyph's slot, carrying no glyph.
func TestExecuteRunsControlConditions(t *testing.T) {
	results, f, out := runConditions(t, []assay.Condition{assay.ScrambledGlyph, assay.OffTargetGlyph, assay.GlyphMinusRest},
		MaterialsSpec{Scenario: "scenario.md", Scrambled: "scrambled_glyph.md", OffTarget: "off_target_glyph.md", GlyphMinusRest: "glyph_minus_rest.md"},
		map[string]string{"scenario.md": "SCENARIO", "scrambled_glyph.md": "SCR", "off_target_glyph.md": "OTG", "glyph_minus_rest.md": "GMR"})
	if len(results.Rows) != 3 {
		t.Fatalf("rows: %+v", results.Rows)
	}
	assertPrompts(t, f, "SCR\n---\nSCENARIO", "OTG\n---\nSCENARIO", "GMR\n---\nSCENARIO")
	assertResponses(t, out, "scrambled_glyph_1.txt", "off_target_glyph_1.txt", "glyph_minus_rest_1.txt")
}

// The behavioral-equivalence conditions read the duty and corpus materials and
// assemble them into the guidance slot, and brief-only (baseline) carries no
// guidance. This exercises the duty/corpus reads in loadMaterials end-to-end.
func TestExecuteAssemblesBehavioralEquivalenceConditions(t *testing.T) {
	results, f, _ := runConditions(t, []assay.Condition{assay.Baseline, assay.DutyOnly, assay.CorpusOnly},
		MaterialsSpec{Scenario: "scenario.md", Duty: "duty.md", Corpus: "corpus.md"},
		map[string]string{"scenario.md": "SCENARIO", "duty.md": "DUTY", "corpus.md": "CORPUS"})
	if len(results.Rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(results.Rows))
	}
	// Prompts are recorded in condition order: baseline (scenario only), then
	// duty and corpus each in the guidance slot.
	assertPrompts(t, f, "SCENARIO", "DUTY\n---\nSCENARIO", "CORPUS\n---\nSCENARIO")
}

// The cartographer-duty-format conditions read the three design instruments and
// assemble each into the guidance slot. This exercises the annotated,
// cartographer, and cartographer_scan reads in loadMaterials end-to-end.
func TestExecuteAssemblesCartographerConditions(t *testing.T) {
	results, f, _ := runConditions(t,
		[]assay.Condition{assay.AnnotatedInstrument, assay.CartographerInstrument, assay.CartographerScanInstrument},
		MaterialsSpec{
			Scenario: "scenario.md", Annotated: "annotated.md",
			Cartographer: "cartographer.md", CartographerScan: "cartographer_scan.md",
		},
		map[string]string{
			"scenario.md": "TASK", "annotated.md": "ANN",
			"cartographer.md": "CART", "cartographer_scan.md": "SCAN",
		})
	if len(results.Rows) != 3 {
		t.Fatalf("want 3 rows, got %d", len(results.Rows))
	}
	assertPrompts(t, f, "ANN\n---\nTASK", "CART\n---\nTASK", "SCAN\n---\nTASK")
}

// The form × grounding 2×2 conditions: ground_only reuses the ground material
// and domain_imperative_only reads the new domain imperative, each assembled
// into the guidance slot with no glyph. This exercises the ground and
// domain_imperative reads in loadMaterials end-to-end.
func TestExecuteAssemblesFormGroundingConditions(t *testing.T) {
	results, f, out := runConditions(t, []assay.Condition{assay.GroundOnly, assay.DomainImperativeOnly},
		MaterialsSpec{Scenario: "scenario.md", Ground: "ground.md", DomainImperative: "domain_imperative.md"},
		map[string]string{"scenario.md": "SCENARIO", "ground.md": "GROUND", "domain_imperative.md": "DI"})
	if len(results.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(results.Rows))
	}
	assertPrompts(t, f, "GROUND\n---\nSCENARIO", "DI\n---\nSCENARIO")
	assertResponses(t, out, "ground_only_1.txt", "domain_imperative_only_1.txt")
}

// runConditions runs a one-run study of conds over mats, with files (each
// material's name -> content) in its /in dir. It returns the results, the
// client holding the assembled prompts, and the output dir.
func runConditions(t *testing.T, conds []assay.Condition, mats MaterialsSpec, files map[string]string) (Results, *fakeClient, string) {
	t.Helper()
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
		Conditions: conds, RunsPerCell: 1, Materials: mats, Sampling: validSampling(),
	}
	in := writeStudy(t, spec, files)
	out := t.TempDir()
	f := &fakeClient{Text: "reply"}
	results, err := Execute(context.Background(), in, out, f)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	return results, f, out
}

// assertPrompts fails unless f received exactly the prompts want, in order.
func assertPrompts(t *testing.T, f *fakeClient, want ...string) {
	t.Helper()
	if len(f.Prompts) != len(want) {
		t.Fatalf("assembled prompts = %q, want %q", f.Prompts, want)
	}
	for i, w := range want {
		if f.Prompts[i] != w {
			t.Fatalf("prompt %d = %q, want %q", i, f.Prompts[i], w)
		}
	}
}

// assertResponses fails unless each named response file exists under out.
func assertResponses(t *testing.T, out string, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := os.ReadFile(filepath.Join(out, "responses", name)); err != nil {
			t.Fatalf("missing response %s: %v", name, err)
		}
	}
}

// A study that names a condition but ships no material for it fails loudly,
// through that material's read branch, rather than assembling a bare scenario
// as if the arm were empty. The scenario is present each time; the condition's
// own material is deliberately absent.
func TestExecuteReportsMissingConditionMaterial(t *testing.T) {
	for _, c := range []struct {
		material string
		cond     assay.Condition
		mats     MaterialsSpec
	}{
		{"imperative", assay.ImperativeOnly, MaterialsSpec{Imperative: "imperative.md"}},
		{"scrambled", assay.ScrambledGlyph, MaterialsSpec{Scrambled: "scrambled_glyph.md"}},
		{"off-target", assay.OffTargetGlyph, MaterialsSpec{OffTarget: "off_target_glyph.md"}},
		{"cartographer", assay.CartographerInstrument, MaterialsSpec{Cartographer: "cartographer.md"}},
		{"annotated", assay.AnnotatedInstrument, MaterialsSpec{Annotated: "annotated.md"}},
		{"cartographer_scan", assay.CartographerScanInstrument, MaterialsSpec{CartographerScan: "cartographer_scan.md"}},
		{"duty", assay.DutyOnly, MaterialsSpec{Duty: "duty.md"}},
		{"corpus", assay.CorpusOnly, MaterialsSpec{Corpus: "corpus.md"}},
		{"domain-imperative", assay.DomainImperativeOnly, MaterialsSpec{DomainImperative: "domain_imperative.md"}},
	} {
		t.Run(c.material, func(t *testing.T) {
			c.mats.Scenario = "scenario.md"
			spec := StudySpec{
				Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m"},
				Conditions: []assay.Condition{c.cond}, RunsPerCell: 1,
				Materials: c.mats, Sampling: validSampling(),
			}
			in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
			if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{Text: "x"}); err == nil {
				t.Fatalf("expected missing-%s-material error", c.material)
			}
		})
	}
}

// When the client surfaces a rendered prompt (the raw /completion path), the
// runner records it per run so the run is self-describing down to its input.
func TestExecutePersistsRenderedPromptWhenSurfaced(t *testing.T) {
	got := executeOneCell(t, "completion", &model.Response{Text: "reply", RenderedPrompt: "WRAPPED-INPUT"}, "prompts")
	if got != "WRAPPED-INPUT" {
		t.Fatalf("recorded prompt = %q, want WRAPPED-INPUT", got)
	}
}

// executeOneCell runs a one-run baseline study on endpoint ("" for chat,
// "completion" for raw /completion) with resp as the model reply, and returns
// the run's recorded file under the subdir (prompts, reasoning, ...) of the
// output dir.
func executeOneCell(t *testing.T, endpoint string, resp *model.Response, subdir string) string {
	t.Helper()
	out := t.TempDir()
	if _, err := Execute(context.Background(), baselineIn(t, endpoint, 1), out, &fakeClient{Resp: resp}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, subdir, "baseline_1.txt"))
	if err != nil {
		t.Fatalf("read recorded %s: %v", subdir, err)
	}
	return string(got)
}

// baselineIn writes the smallest valid study, runs baseline runs over
// scenario.md on endpoint, and returns its /in dir.
func baselineIn(t *testing.T, endpoint string, runs int) string {
	t.Helper()
	spec := StudySpec{
		Assay: SupportedAssay, ItemID: "i", Model: ModelSpec{ModelID: "m", Endpoint: endpoint},
		Conditions: []assay.Condition{assay.Baseline}, RunsPerCell: runs,
		Materials: MaterialsSpec{Scenario: "scenario.md"}, Sampling: validSampling(),
	}
	return writeStudy(t, spec, map[string]string{"scenario.md": "S"})
}

// When the client surfaces a reasoning trace (a thinking model), the runner
// records it per run so a study can check the self-reported field source against
// the route the model actually reasoned through.
func TestExecutePersistsReasoningWhenSurfaced(t *testing.T) {
	got := executeOneCell(t, "completion", &model.Response{Text: "VERDICT: yes", Reasoning: "walked the scope check"}, "reasoning")
	if got != "walked the scope check" {
		t.Fatalf("recorded reasoning = %q, want 'walked the scope check'", got)
	}
}

// A non-thinking model surfaces no reasoning, so the runner writes no reasoning
// file rather than an empty one that would read as a captured blank trace.
func TestExecuteWritesNoReasoningFileForNonThinkingModel(t *testing.T) {
	in := baselineIn(t, "", 1)
	out := t.TempDir()
	if _, err := Execute(context.Background(), in, out, &fakeClient{Text: "VERDICT: no"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "reasoning", "baseline_1.txt")); !os.IsNotExist(err) {
		t.Fatal("a non-thinking model must not leave a reasoning file")
	}
}

// A reasoning trace that cannot be written aborts the run rather than dropping
// the trace silently — the same fail-loud contract as the response and prompt
// writes. Here the target path is pre-made a directory so WriteFile fails.
func TestExecuteReportsReasoningWriteError(t *testing.T) {
	in := baselineIn(t, "", 1)
	out := t.TempDir()
	if err := os.MkdirAll(filepath.Join(out, "reasoning", "baseline_1.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	f := &fakeClient{Resp: &model.Response{Text: "VERDICT: yes", Reasoning: "trace"}}
	if _, err := Execute(context.Background(), in, out, f); err == nil {
		t.Fatal("expected a reasoning write error")
	}
}

// The reasoning directory cannot be created when its path is already a file; the
// run fails at setup rather than proceeding with nowhere to record traces.
func TestExecuteReportsReasoningDirError(t *testing.T) {
	in := baselineIn(t, "", 1)
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(out, "reasoning"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Execute(context.Background(), in, out, &fakeClient{Text: "reply"}); err == nil {
		t.Fatal("expected a reasoning dir creation error")
	}
}

// A truncated reply (stopped at the token cap) is recorded on the row's Observed
// and flagged in the rationale, so a cut-off answer is a visible fact.
func TestExecuteRecordsTruncationOnTheRow(t *testing.T) {
	in := baselineIn(t, "", 1)
	f := &fakeClient{Resp: &model.Response{Text: "cut off here", Truncated: true}}
	results, err := Execute(context.Background(), in, t.TempDir(), f)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(results.Rows) != 1 || !results.Rows[0].Observed.Truncated {
		t.Fatalf("row Observed.Truncated not set: %+v", results.Rows)
	}
	if !strings.Contains(results.Rows[0].Rationale, ":truncated") {
		t.Fatalf("rationale missing :truncated marker: %q", results.Rows[0].Rationale)
	}
}

// The chat path surfaces no rendered prompt (the server templates server-side),
// so the runner writes no prompt file rather than recording a half-truth.
func TestExecuteWritesNoPromptFileInChatMode(t *testing.T) {
	in := baselineIn(t, "", 1)
	out := t.TempDir()
	if _, err := Execute(context.Background(), in, out, &fakeClient{Text: "reply"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "prompts", "baseline_1.txt")); !os.IsNotExist(err) {
		t.Fatal("chat mode must not write a prompt file")
	}
}

func TestExecuteIsNotPlaceholder(t *testing.T) {
	// Guards against the registry-lab regression: results.json must carry
	// real scored rows, never a {"status":"pending-adapter"} stub.
	in := baselineIn(t, "", 1)
	out := t.TempDir()

	if _, err := Execute(context.Background(), in, out, &fakeClient{Text: "PASS"}); err != nil {
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
	if _, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{Text: "PASS"}); err == nil {
		t.Fatal("expected missing-material error")
	}
}

func TestExecuteFailsFastOnProbeError(t *testing.T) {
	in := baselineIn(t, "", 3)
	_, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{Err: errContext})
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
		Text:      "reply",
		ModelName: "qwen",
		ServerProps: model.ServerProps{
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
		Text: "reply",
		ServerProps: model.ServerProps{
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
		Text:        "reply",
		ServerProps: model.ServerProps{ModelPath: "/models/Mistral-7B-Instruct-v0.3.Q4_K_M.gguf", ModelAlias: "mistral"},
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
	f := &fakeClient{Text: "reply", PropsErr: errors.New("props unreachable")}
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
	f := &fakeClient{Text: "reply"}
	if _, err := Execute(context.Background(), in, t.TempDir(), f); err != nil {
		t.Fatal(err)
	}
	if len(f.Params) == 0 {
		t.Fatal("no generations recorded")
	}
	for i, p := range f.Params {
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
	_, err := Execute(context.Background(), in, t.TempDir(), &fakeClient{Text: "x"})
	if err == nil || !strings.Contains(err.Error(), "repeat_penalty") {
		t.Fatalf("err = %v, want a refusal naming repeat_penalty", err)
	}
}

// errAtClient fails the Generate call at each index in fails with that error and
// returns reply for every other call. It lets a test drive exactly which cells
// hit a model error while the rest of the grid succeeds.
type errAtClient struct {
	fails map[int]error
	reply string
	n     int
}

func (c *errAtClient) Generate(_ context.Context, _ string, _ model.GenParams) (model.Response, error) {
	i := c.n
	c.n++
	if e, ok := c.fails[i]; ok {
		return model.Response{}, e
	}
	reply := c.reply
	if reply == "" {
		reply = "PASS"
	}
	return model.Response{Text: reply, Model: "qwen", SystemFingerprint: "bX", Timings: model.Timings{PredictedN: 5, PredictedPerSecond: 40}}, nil
}
func (c *errAtClient) Name() string    { return "qwen" }
func (c *errAtClient) Version() string { return "q4" }
func (c *errAtClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

// timeoutErr is a net.Error whose Timeout reports true — the shape a request
// deadline produces. A per-cell timeout is a cell error, not a study abort.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "request timed out" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

// Bug 1431: a single cell's model error (a 400 context-overflow, or a request
// timeout) must be recorded as a not-scoreable row and the batch must finish the
// rest of the grid and write results.json — not abort the whole study and lose
// every other cell.
func TestExecuteContinuesPastCellError(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH", "ground.md": "GROUND",
	})
	out := t.TempDir()
	client := &errAtClient{fails: map[int]error{
		1: &model.APIError{ModelID: "qwen", StatusCode: 400, Body: "the request exceeds the available context size"},
		3: timeoutErr{},
	}}

	results, err := Execute(context.Background(), in, out, client)
	if err != nil {
		t.Fatalf("a per-cell model error must not abort the study: %v", err)
	}
	// 3 conditions × 2 runs = 6 rows, every cell present including the failed ones.
	if len(results.Rows) != 6 {
		t.Fatalf("rows = %d, want 6 — a failed cell is still a row", len(results.Rows))
	}
	var nscored int
	for _, row := range results.Rows {
		if row.Score == assay.ScoreN {
			nscored++
			if !strings.Contains(row.Rationale, "cell-error") {
				t.Fatalf("error row rationale = %q, want it to mark the cell error", row.Rationale)
			}
		}
	}
	if nscored != 2 {
		t.Fatalf("not-scoreable rows = %d, want 2 (the 400 and the timeout)", nscored)
	}
	// results.json was written despite the mid-grid failures.
	if _, err := os.ReadFile(filepath.Join(out, "results.json")); err != nil {
		t.Fatalf("results.json must be written after a per-cell failure: %v", err)
	}
}

// The loop assay isolates per-cell failures the same way the single-turn assay
// does: one cell's model 400 records a null row and the other cells still run.
func TestExecuteLoopContinuesPastCellError(t *testing.T) {
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH",
		"preamble.md": "PREAMBLE", "sandbox.json": `{"REVIEW.md":""}`,
	})
	out := t.TempDir()
	// loopSpec has 2 conditions × 1 run; fail the first cell's only turn.
	client := &errAtClient{
		fails: map[int]error{0: &model.APIError{ModelID: "qwen", StatusCode: 400, Body: "context overflow"}},
		reply: "FINAL done",
	}

	results, err := Execute(context.Background(), in, out, client)
	if err != nil {
		t.Fatalf("a per-cell model error must not abort the loop study: %v", err)
	}
	if len(results.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(results.Rows))
	}
	var nscored int
	for _, row := range results.Rows {
		if row.Score == assay.ScoreN {
			nscored++
		}
	}
	if nscored != 1 {
		t.Fatalf("not-scoreable rows = %d, want 1", nscored)
	}
	if _, err := os.ReadFile(filepath.Join(out, "results.json")); err != nil {
		t.Fatalf("results.json must be written: %v", err)
	}
}

// A setup-level failure still aborts: an unreachable server (a transport error
// that is not a timeout) is not a per-cell error, so it fails the whole study
// rather than scoring every cell null against a server that never answered.
func TestExecuteAbortsOnUnreachableServer(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "R",
	})
	client := &errAtClient{fails: map[int]error{0: errors.New("dial tcp 127.0.0.1:8081: connect: connection refused")}}
	if _, err := Execute(context.Background(), in, t.TempDir(), client); err == nil {
		t.Fatal("an unreachable server must abort the study, not score cells null")
	}
}
