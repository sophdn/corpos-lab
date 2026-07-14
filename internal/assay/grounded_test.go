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

// testSampling is a sampled (non-greedy) regime with one seed per replicate —
// the shape a real graded study declares.
func testSampling() Sampling {
	return Sampling{Temperature: 0.8, Seeds: []int{11, 22, 33}, MaxTokens: 512}
}

// Fake client for scoring/run tests.
type fakeClient struct {
	text    string
	err     error
	prompts []string
	params  []model.GenParams
}

func (f *fakeClient) Generate(_ context.Context, prompt string, p model.GenParams) (model.Response, error) {
	f.prompts = append(f.prompts, prompt)
	f.params = append(f.params, p)
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{Text: f.text}, nil
}
func (f *fakeClient) Name() string    { return "fake" }
func (f *fakeClient) Version() string { return "0.0.0" }

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
