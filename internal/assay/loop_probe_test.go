package assay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"corpos-lab/internal/agentloop"
	"corpos-lab/internal/model"
)

// loopScript returns one canned reply per turn, for driving a loop cell.
type loopScript struct {
	replies []string
	prompts []string
	turn    int
}

func (c *loopScript) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.prompts = append(c.prompts, prompt)
	text := "FINAL done"
	if c.turn < len(c.replies) {
		text = c.replies[c.turn]
	}
	c.turn++
	return model.Response{Text: text, Model: "qwen", SystemFingerprint: "bZ", Timings: model.Timings{PredictedN: 9, PredictedPerSecond: 44}}, nil
}
func (c *loopScript) Name() string    { return "qwen" }
func (c *loopScript) Version() string { return "q4" }
func (c *loopScript) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func TestRunLoopProbeAssemblesBaseActsAndCaptures(t *testing.T) {
	client := &loopScript{replies: []string{
		"CALL read_file CHANGELOG.md",
		"CALL edit_file CHANGELOG.md ||| # Changelog\n## v1.5.0",
		"FINAL updated",
	}}
	sandbox := map[string]string{"CHANGELOG.md": "# Changelog\n## v1.4.0"}

	row, resp, arts, err := RunLoopProbe(context.Background(), client, "casg-direct", GlyphOnly, 1,
		Materials{Scenario: "SCENARIO", Glyph: "GLYPH"}, "PREAMBLE", sandbox, testSampling(), agentloop.Config{})
	if err != nil {
		t.Fatalf("RunLoopProbe: %v", err)
	}
	// The base prompt is preamble + delimiter + (glyph "---" scenario).
	if !strings.HasPrefix(resp.Prompt, "PREAMBLE"+LoopPreambleDelimiter) || !strings.Contains(resp.Prompt, "GLYPH") || !strings.Contains(resp.Prompt, "SCENARIO") {
		t.Fatalf("base prompt shape wrong:\n%s", resp.Prompt)
	}
	// The first turn's prompt is exactly the base; later turns carry observations.
	if client.prompts[0] != resp.Prompt {
		t.Fatal("first loop turn did not receive the base prompt verbatim")
	}
	if row.Score != Unscored || !strings.Contains(row.Rationale, "terminal=final") || !strings.Contains(row.Rationale, "edits=1") {
		t.Fatalf("row = %+v", row)
	}
	if row.Observed.Model != "qwen" || row.Observed.PredictedTokens != 27 {
		t.Fatalf("observed = %+v, want qwen and 27 tokens summed", row.Observed)
	}
	if !strings.Contains(arts.Sandbox["CHANGELOG.md"], "v1.5.0") {
		t.Fatalf("sandbox not edited: %q", arts.Sandbox["CHANGELOG.md"])
	}
	if !strings.Contains(resp.Text, "OBSERVATION:") {
		t.Fatalf("transcript missing observations:\n%s", resp.Text)
	}
}

func TestRunLoopProbeRejectsMissingAidMaterial(t *testing.T) {
	// glyph_only with no glyph material — AssemblePrompt fails, so does the probe.
	_, _, _, err := RunLoopProbe(context.Background(), &loopScript{}, "x", GlyphOnly, 1,
		Materials{Scenario: "S"}, "P", nil, testSampling(), agentloop.Config{})
	if err == nil {
		t.Fatal("expected an error when glyph_only has no glyph")
	}
}

func TestRunLoopProbeSurfacesModelError(t *testing.T) {
	_, _, _, err := RunLoopProbe(context.Background(), &erroringClient{}, "x", Baseline, 1,
		Materials{Scenario: "S"}, "P", nil, testSampling(), agentloop.Config{})
	if err == nil {
		t.Fatal("expected the model error to surface")
	}
}

type erroringClient struct{}

func (erroringClient) Generate(context.Context, string, model.GenParams) (model.Response, error) {
	return model.Response{}, errors.New("boom")
}
func (erroringClient) Name() string    { return "e" }
func (erroringClient) Version() string { return "v" }
func (erroringClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}
