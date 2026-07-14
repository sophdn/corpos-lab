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

// Fake client for scoring/run tests.
type fakeClient struct {
	text    string
	err     error
	prompts []string
}

func (f *fakeClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	f.prompts = append(f.prompts, prompt)
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{Text: f.text}, nil
}
func (f *fakeClient) Name() string    { return "fake" }
func (f *fakeClient) Version() string { return "0.0.0" }

func TestRunProbeCapturesRowIdentityAndResponse(t *testing.T) {
	f := &fakeClient{text: "I'd update the changelog first."}
	row, resp, err := RunProbe(context.Background(), f, "i-42", GlyphOnly, 3, Materials{Scenario: "S", Glyph: "G"})
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
		row, _, err := RunProbe(context.Background(), f, "i", Baseline, 1, Materials{Scenario: "S"})
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
	_, _, err := RunProbe(context.Background(), f, "i", Baseline, 2, Materials{Scenario: "S"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "baseline") || !strings.Contains(err.Error(), "run 2") {
		t.Fatalf("error should name condition and run: %v", err)
	}
}

func TestRunProbeSurfacesAssemblyError(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	_, _, err := RunProbe(context.Background(), f, "i", GlyphOnly, 1, Materials{Scenario: "S"})
	if err == nil {
		t.Fatal("expected assembly error for missing glyph")
	}
	if len(f.prompts) != 0 {
		t.Fatal("model should not be called when assembly fails")
	}
}

func TestProbeGenParamsPinDeterministicSampling(t *testing.T) {
	p := probeGenParams()
	if p.Temperature == nil || *p.Temperature != 0.0 {
		t.Fatalf("temperature: %v", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 512 {
		t.Fatalf("max tokens: %v", p.MaxTokens)
	}
}
