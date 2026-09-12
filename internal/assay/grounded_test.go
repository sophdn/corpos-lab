package assay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"corpos-lab/internal/model"
)

// Prompt-assembly tests ported from grounded_glyph_probe.rs (the setup_*
// tests): the \n---\n delimiter and per-condition material requirements.

func TestAssemblePromptBaselineUsesOnlyScenario(t *testing.T) {
	got, err := AssemblePrompt(Baseline, Materials{Scenario: "S", Glyph: "G", Ground: "GR"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "S" {
		t.Fatalf("got %q, want S", got)
	}
}

func TestAssemblePromptGlyphOnlyConcatsGlyphAndScenario(t *testing.T) {
	got, err := AssemblePrompt(GlyphOnly, Materials{Scenario: "S", Glyph: "G"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "G\n---\nS" {
		t.Fatalf("got %q, want G\\n---\\nS", got)
	}
}

func TestAssemblePromptGroundedGlyphConcatsAllThree(t *testing.T) {
	got, err := AssemblePrompt(GroundedGlyph, Materials{Scenario: "S", Glyph: "G", Ground: "GR"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "G\n---\nGR\n---\nS" {
		t.Fatalf("got %q, want G\\n---\\nGR\\n---\\nS", got)
	}
}

// stubModel is a minimal model.Client for RunProbe tests.
type stubModel struct {
	resp model.Response
	err  error
}

func (s stubModel) Generate(context.Context, string, model.GenParams) (model.Response, error) {
	return s.resp, s.err
}
func (s stubModel) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}
func (s stubModel) Name() string    { return "stub" }
func (s stubModel) Version() string { return "v0" }

func TestRunProbeRecordsTruncation(t *testing.T) {
	m := stubModel{resp: model.Response{Text: "cut", Truncated: true}}
	row, _, err := RunProbe(context.Background(), m, "item", Baseline, 1, Materials{Scenario: "S"}, Sampling{Seeds: []int{1}})
	if err != nil {
		t.Fatal(err)
	}
	if !row.Observed.Truncated {
		t.Fatal("row Observed.Truncated should be true")
	}
	if !strings.Contains(row.Rationale, ":truncated") {
		t.Fatalf("rationale missing :truncated: %q", row.Rationale)
	}
}

func TestRunProbeSurfacesGenerateError(t *testing.T) {
	m := stubModel{err: errors.New("boom")}
	if _, _, err := RunProbe(context.Background(), m, "item", Baseline, 1, Materials{Scenario: "S"}, Sampling{Seeds: []int{1}}); err == nil {
		t.Fatal("expected RunProbe to surface the Generate error")
	}
}

func TestAssemblePromptImperativeOnlyConcatsImperativeAndScenario(t *testing.T) {
	// T2: the imperative rule takes the glyph's slot — same delimiter, same
	// shape as GlyphOnly, no glyph. A glyph present in Materials must not leak
	// into the prompt.
	got, err := AssemblePrompt(ImperativeOnly, Materials{Scenario: "S", Glyph: "G", Imperative: "IMP"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "IMP\n---\nS" {
		t.Fatalf("got %q, want IMP\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingImperative(t *testing.T) {
	// A glyph present but no imperative must still fail: T2 carries the
	// imperative, never the glyph as a fallback.
	if _, err := AssemblePrompt(ImperativeOnly, Materials{Scenario: "S", Glyph: "G"}); err == nil {
		t.Fatal("expected error for missing imperative")
	}
}

func TestAssemblePromptScrambledGlyphConcatsScrambledAndScenario(t *testing.T) {
	// The scrambled glyph takes the glyph's slot — same delimiter, same shape as
	// GlyphOnly. A real glyph present in Materials must not leak into the prompt.
	got, err := AssemblePrompt(ScrambledGlyph, Materials{Scenario: "S", Glyph: "G", Scrambled: "SCR"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "SCR\n---\nS" {
		t.Fatalf("got %q, want SCR\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingScrambled(t *testing.T) {
	// A glyph present but no scrambled material must still fail: the control
	// carries the scrambled glyph, never the real glyph as a fallback.
	if _, err := AssemblePrompt(ScrambledGlyph, Materials{Scenario: "S", Glyph: "G"}); err == nil {
		t.Fatal("expected error for missing scrambled glyph")
	}
}

func TestAssemblePromptOffTargetGlyphConcatsOffTargetAndScenario(t *testing.T) {
	got, err := AssemblePrompt(OffTargetGlyph, Materials{Scenario: "S", Glyph: "G", OffTarget: "OTG"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "OTG\n---\nS" {
		t.Fatalf("got %q, want OTG\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingOffTarget(t *testing.T) {
	if _, err := AssemblePrompt(OffTargetGlyph, Materials{Scenario: "S", Glyph: "G"}); err == nil {
		t.Fatal("expected error for missing off-target glyph")
	}
}

func TestAssemblePromptRejectsMissingGlyph(t *testing.T) {
	if _, err := AssemblePrompt(GlyphOnly, Materials{Scenario: "S"}); err == nil {
		t.Fatal("expected error for missing glyph")
	}
	if _, err := AssemblePrompt(GroundedGlyph, Materials{Scenario: "S", Ground: "GR"}); err == nil {
		t.Fatal("expected error for missing glyph in grounded")
	}
}

func TestAssemblePromptRejectsMissingGround(t *testing.T) {
	if _, err := AssemblePrompt(GroundedGlyph, Materials{Scenario: "S", Glyph: "G"}); err == nil {
		t.Fatal("expected error for missing ground")
	}
}

func TestAssemblePromptRejectsUnknownCondition(t *testing.T) {
	if _, err := AssemblePrompt(Condition("nonsense"), Materials{Scenario: "S"}); err == nil {
		t.Fatal("expected error for unknown condition")
	}
}

// Behavioral-equivalence Condition A: the duty specification takes the guidance
// slot — same delimiter and shape as GlyphOnly, no glyph. A glyph present in
// Materials must not leak into the prompt.
func TestAssemblePromptDutyOnlyConcatsDutyAndScenario(t *testing.T) {
	got, err := AssemblePrompt(DutyOnly, Materials{Scenario: "S", Glyph: "G", Duty: "DUTY"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "DUTY\n---\nS" {
		t.Fatalf("got %q, want DUTY\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingDuty(t *testing.T) {
	// A glyph present but no duty must still fail: Condition A carries the duty,
	// never the glyph as a fallback.
	if _, err := AssemblePrompt(DutyOnly, Materials{Scenario: "S", Glyph: "G"}); err == nil {
		t.Fatal("expected error for missing duty")
	}
}

// Behavioral-equivalence Condition B: the corpus takes the guidance slot.
func TestAssemblePromptCorpusOnlyConcatsCorpusAndScenario(t *testing.T) {
	got, err := AssemblePrompt(CorpusOnly, Materials{Scenario: "S", Duty: "DUTY", Corpus: "CORPUS"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "CORPUS\n---\nS" {
		t.Fatalf("got %q, want CORPUS\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingCorpus(t *testing.T) {
	if _, err := AssemblePrompt(CorpusOnly, Materials{Scenario: "S", Duty: "DUTY"}); err == nil {
		t.Fatal("expected error for missing corpus")
	}
}

func TestAssemblePromptAnnotatedInstrumentConcatsAnnotatedAndScenario(t *testing.T) {
	got, err := AssemblePrompt(AnnotatedInstrument, Materials{Scenario: "S", Annotated: "ANN"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "ANN\n---\nS" {
		t.Fatalf("got %q, want ANN\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingAnnotated(t *testing.T) {
	if _, err := AssemblePrompt(AnnotatedInstrument, Materials{Scenario: "S"}); err == nil {
		t.Fatal("expected error for missing annotated instrument")
	}
}

func TestAssemblePromptCartographerInstrumentConcatsCartographerAndScenario(t *testing.T) {
	got, err := AssemblePrompt(CartographerInstrument, Materials{Scenario: "S", Cartographer: "CART"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "CART\n---\nS" {
		t.Fatalf("got %q, want CART\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingCartographer(t *testing.T) {
	if _, err := AssemblePrompt(CartographerInstrument, Materials{Scenario: "S"}); err == nil {
		t.Fatal("expected error for missing cartographer instrument")
	}
}

func TestAssemblePromptCartographerScanConcatsScanAndScenario(t *testing.T) {
	got, err := AssemblePrompt(CartographerScanInstrument, Materials{Scenario: "S", CartographerScan: "SCAN"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "SCAN\n---\nS" {
		t.Fatalf("got %q, want SCAN\\n---\\nS", got)
	}
}

func TestAssemblePromptRejectsMissingCartographerScan(t *testing.T) {
	if _, err := AssemblePrompt(CartographerScanInstrument, Materials{Scenario: "S"}); err == nil {
		t.Fatal("expected error for missing cartographer-scan instrument")
	}
}

// testSampling is a sampled (non-greedy) regime with one seed per replicate —
// the shape a real graded study declares. The chain is complete and neutral:
// min_p is the only live truncation stage, every other stage pinned to its
// disabled value.
func testSampling() Sampling {
	return Sampling{
		Temperature: 0.8, Seeds: []int{11, 22, 33}, MaxTokens: 512,
		TopNSigma: -1.0, TopK: 0, TypicalP: 1.0, TopP: 1.0, MinP: 0.05,
		RepeatPenalty: 1.0, RepeatLastN: 0, PresencePenalty: 0.0, FrequencyPenalty: 0.0,
		XTCProbability: 0.0, DryMultiplier: 0.0,
	}
}

// Fake client for scoring/run tests.
type fakeClient struct {
	text    string
	err     error
	prompts []string
	params  []model.GenParams
	// resp, when set, is returned instead of a bare text response — for
	// asserting that server-reported observations reach the row.
	resp *model.Response
}

func (f *fakeClient) Generate(_ context.Context, prompt string, p model.GenParams) (model.Response, error) {
	f.prompts = append(f.prompts, prompt)
	f.params = append(f.params, p)
	if f.err != nil {
		return model.Response{}, f.err
	}
	if f.resp != nil {
		return *f.resp, nil
	}
	return model.Response{Text: f.text}, nil
}
func (f *fakeClient) Name() string    { return "fake" }
func (f *fakeClient) Version() string { return "0.0.0" }
func (f *fakeClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func TestRunProbeCapturesRowIdentityAndResponse(t *testing.T) {
	f := &fakeClient{text: "I'd update the changelog first."}
	row, resp, err := RunProbe(context.Background(), f, "i-42", GlyphOnly, 3, Materials{Scenario: "S", Glyph: "G"}, testSampling())
	if err != nil {
		t.Fatal(err)
	}
	if row.Item != "i-42" || row.Condition != GlyphOnly || row.Run != 3 {
		t.Fatalf("row identity: %+v", row)
	}
	if !strings.Contains(row.Rationale, "grounded-glyph-probe") {
		t.Fatalf("rationale: %q", row.Rationale)
	}
	// The response is the load-bearing artifact — the rubric path reads it.
	if resp.Prompt != "G\n---\nS" || resp.Text != "I'd update the changelog first." {
		t.Fatalf("response audit: %+v", resp)
	}
}

// A thinking model's reasoning reaches the response so a study can check a
// self-reported field source against the route the model actually reasoned
// through. The client surfaces it; the probe carries it verbatim.
func TestRunProbeCapturesReasoningTrace(t *testing.T) {
	f := &fakeClient{resp: &model.Response{
		Text:      "VERDICT: yes\nFIELD SOURCE: Scope — operative when",
		Reasoning: "The scope check: no companion update in the trace, so it fires.",
	}}
	_, resp, err := RunProbe(context.Background(), f, "i", GroundedGlyph, 1,
		Materials{Scenario: "S", Glyph: "G", Ground: "R"}, testSampling())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Reasoning != "The scope check: no companion update in the trace, so it fires." {
		t.Fatalf("reasoning not carried onto the response: %q", resp.Reasoning)
	}
}

// A non-thinking model produces no reasoning, and the probe must not invent one.
func TestRunProbeLeavesReasoningEmptyForNonThinkingModel(t *testing.T) {
	f := &fakeClient{text: "VERDICT: no"}
	_, resp, err := RunProbe(context.Background(), f, "i", Baseline, 1, Materials{Scenario: "S"}, testSampling())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Reasoning != "" {
		t.Fatalf("expected empty reasoning, got %q", resp.Reasoning)
	}
}

// The probe must never invent a code. These are the exact reply shapes that
// previously came back a spurious fail: behavioral prose opens with neither
// PASS nor FAIL because the probe never asks for either.
func TestRunProbeLeavesEveryReplyUnscored(t *testing.T) {
	for _, text := range []string{
		"## [1.5.0] - 2026-07-14\n### Added\n- ChainedFilter",            // a correct entry (would be C)
		"The Marker's obligation to update the changelog remains unmet.", // analysis mode (Ii)
		"Both artifacts appear to have been updated already.",            // reasoned away (I)
		"PASS", // the battery's token, arriving by coincidence
		"FAIL", // ditto
		"",     // empty reply
	} {
		f := &fakeClient{text: text}
		row, _, err := RunProbe(context.Background(), f, "i", Baseline, 1, Materials{Scenario: "S"}, testSampling())
		if err != nil {
			t.Fatal(err)
		}
		if row.Score != Unscored {
			t.Fatalf("reply %q scored %q; the probe must not judge — a judge does", text, row.Score)
		}
	}
}

func TestRunProbeSurfacesModelError(t *testing.T) {
	f := &fakeClient{err: errors.New("timeout")}
	_, _, err := RunProbe(context.Background(), f, "i", Baseline, 2, Materials{Scenario: "S"}, testSampling())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "baseline") || !strings.Contains(err.Error(), "run 2") {
		t.Fatalf("error should name condition and run: %v", err)
	}
}

func TestRunProbeSurfacesAssemblyError(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	_, _, err := RunProbe(context.Background(), f, "i", GlyphOnly, 1, Materials{Scenario: "S"}, testSampling())
	if err == nil {
		t.Fatal("expected assembly error for missing glyph")
	}
	if len(f.prompts) != 0 {
		t.Fatal("model should not be called when assembly fails")
	}
}

func TestRunProbeAppliesTheStudysSamplingRegime(t *testing.T) {
	f := &fakeClient{text: "reply"}
	if _, _, err := RunProbe(context.Background(), f, "i", Baseline, 2,
		Materials{Scenario: "S"}, testSampling()); err != nil {
		t.Fatal(err)
	}
	got := f.params[0]
	if got.Temperature == nil || *got.Temperature != 0.8 {
		t.Fatalf("temperature not taken from the study: %v", got.Temperature)
	}
	if got.MaxTokens == nil || *got.MaxTokens != 512 {
		t.Fatalf("max tokens: %v", got.MaxTokens)
	}
	// Run 2 must draw the SECOND seed. Off-by-one here would silently pair two
	// replicates to one seed and deflate the cell's variance.
	if got.Seed == nil || *got.Seed != 22 {
		t.Fatalf("run 2 should use seed 22, got %v", got.Seed)
	}
}

// Each replicate draws its own seed — the property that makes a graded grid
// reproducible rather than merely random.
func TestGenParamsForRunSelectsSeedByRunIndex(t *testing.T) {
	s := testSampling()
	for run, want := range map[int]int{1: 11, 2: 22, 3: 33} {
		if got := s.GenParamsForRun(run); got.Seed == nil || *got.Seed != want {
			t.Fatalf("run %d: seed = %v, want %d", run, got.Seed, want)
		}
	}
}

func TestGenParamsForRunOutsideSeedSequenceGetsNoSeed(t *testing.T) {
	s := testSampling() // 3 seeds
	// Better an unseeded run than a wrongly-seeded one: reusing seed 1 would
	// duplicate an existing replicate while appearing to be a fresh sample.
	for _, run := range []int{0, 4, 99} {
		if got := s.GenParamsForRun(run); got.Seed != nil {
			t.Fatalf("run %d: expected no seed, got %v", run, *got.Seed)
		}
	}
}

func TestDeterministicSamplingIsGreedyAndUnseeded(t *testing.T) {
	s := Sampling{Temperature: 0, MaxTokens: 256}
	if !s.Deterministic() {
		t.Fatal("temperature 0 must report deterministic")
	}
	p := s.GenParamsForRun(1)
	if p.Temperature == nil || *p.Temperature != 0 {
		t.Fatalf("temperature: %v", p.Temperature)
	}
	if p.Seed != nil {
		t.Fatalf("greedy decoding needs no seed, got %v", *p.Seed)
	}
	if testSampling().Deterministic() {
		t.Fatal("temperature 0.8 must not report deterministic")
	}
}

// Every row carries what the server said while producing THAT row. Per-row is
// the point: a study's substrate was assumed constant for its whole life and
// went unrecorded, so when one leg silently ran on CPU nothing in the data
// could show it. Throughput on the row makes that a step in the numbers rather
// than an invisible fact.
func TestRunProbeRecordsServerReportedObservationsOnEveryRow(t *testing.T) {
	f := &fakeClient{resp: &model.Response{
		Text:              "the reply",
		Model:             "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf",
		SystemFingerprint: "b9445-af6528e6d",
		Timings: model.Timings{
			PromptN: 210, PredictedN: 64,
			PromptPerSecond: 330.29, PredictedPerSecond: 46.23,
		},
	}}
	row, _, err := RunProbe(context.Background(), f, "i-1", GroundedGlyph, 1,
		Materials{Scenario: "S", Glyph: "G", Ground: "R"}, testSampling())
	if err != nil {
		t.Fatal(err)
	}
	if row.Observed.Model != "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf" {
		t.Errorf("observed model = %q", row.Observed.Model)
	}
	if row.Observed.BuildInfo != "b9445-af6528e6d" {
		t.Errorf("observed build = %q", row.Observed.BuildInfo)
	}
	// The tripwire: ~46 is GPU, ~5 is the CPU fallback that went unnoticed.
	if row.Observed.TokensPerSecond != 46.23 {
		t.Errorf("observed tok/s = %v, want 46.23", row.Observed.TokensPerSecond)
	}
	if row.Observed.PromptTokens != 210 || row.Observed.PredictedTokens != 64 {
		t.Errorf("observed token counts = %+v", row.Observed)
	}
}

// GenParamsForRun must send the whole chain, not just the fields it used to.
// A stage left nil here is a stage inherited from the server binary.
func TestGenParamsForRunCarriesEveryStage(t *testing.T) {
	s := Sampling{
		Temperature: 0.7, Seeds: []int{5}, MaxTokens: 128,
		TopNSigma: -1.0, TopK: 11, TypicalP: 0.66, TopP: 0.88, MinP: 0.04,
		RepeatPenalty: 1.07, RepeatLastN: 33, PresencePenalty: 0.22, FrequencyPenalty: 0.11,
		XTCProbability: 0.15, DryMultiplier: 0.44,
	}
	p := s.GenParamsForRun(1)
	for _, c := range []struct {
		field string
		got   any
		want  any
	}{
		{"temperature", *p.Temperature, 0.7},
		{"max_tokens", *p.MaxTokens, 128},
		{"seed", *p.Seed, 5},
		{"top_n_sigma", *p.TopNSigma, -1.0},
		{"top_k", *p.TopK, 11},
		{"typical_p", *p.TypicalP, 0.66},
		{"top_p", *p.TopP, 0.88},
		{"min_p", *p.MinP, 0.04},
		{"repeat_penalty", *p.RepeatPenalty, 1.07},
		{"repeat_last_n", *p.RepeatLastN, 33},
		{"presence_penalty", *p.PresencePenalty, 0.22},
		{"frequency_penalty", *p.FrequencyPenalty, 0.11},
		{"xtc_probability", *p.XTCProbability, 0.15},
		{"dry_multiplier", *p.DryMultiplier, 0.44},
	} {
		if c.got != c.want {
			t.Errorf("GenParams.%s = %v, want %v", c.field, c.got, c.want)
		}
	}
}
