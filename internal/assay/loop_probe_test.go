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
	// truncated, if set, marks per turn (index-aligned with replies) whether the
	// server cut that turn off at the per-call token cap.
	truncated []bool
	turn      int
}

func (c *loopScript) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.prompts = append(c.prompts, prompt)
	text := "FINAL done"
	if c.turn < len(c.replies) {
		text = c.replies[c.turn]
	}
	trunc := false
	if c.turn < len(c.truncated) {
		trunc = c.truncated[c.turn]
	}
	c.turn++
	return model.Response{Text: text, Model: "qwen", SystemFingerprint: "bZ", Truncated: trunc, Timings: model.Timings{PredictedN: 9, PredictedPerSecond: 44}}, nil
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

func TestRunLoopProbeSurfacesTruncatedEdit(t *testing.T) {
	// The edit turn is chopped at the per-call token cap. The ScoreRow must mark
	// the cell as a truncated-EDIT (end-state unscoreable) case, both in the
	// human-readable rationale and in the structured Observed record — so a
	// downstream end-state scorer separates a format/truncation failure from a
	// reasoned terminal outcome without reading the transcript.
	client := &loopScript{
		replies: []string{
			"CALL edit_file rec.md ||| field one\nfield two\nfield thr", // chopped mid-record
			"FINAL done",
		},
		truncated: []bool{true, false},
	}
	row, _, arts, err := RunLoopProbe(context.Background(), client, "casg-direct", GlyphOnly, 1,
		Materials{Scenario: "SCENARIO", Glyph: "GLYPH"}, "PREAMBLE", map[string]string{"rec.md": ""}, testSampling(), agentloop.Config{})
	if err != nil {
		t.Fatalf("RunLoopProbe: %v", err)
	}
	if !row.Observed.Truncated {
		t.Fatal("Observed.Truncated must stay true for compatibility")
	}
	if row.Observed.TruncatedTurns != 1 {
		t.Fatalf("Observed.TruncatedTurns = %d, want 1", row.Observed.TruncatedTurns)
	}
	if !row.Observed.TruncatedEdit {
		t.Fatal("Observed.TruncatedEdit must be true — the truncated turn carried an edit")
	}
	if !strings.Contains(row.Rationale, "truncated-edit") || !strings.Contains(row.Rationale, "UNSCOREABLE") {
		t.Fatalf("rationale must flag the truncated-edit cell as unscoreable: %q", row.Rationale)
	}
	if arts.Result.TruncatedTurns != 1 || !arts.Result.TruncatedEdit {
		t.Fatalf("loop artifacts must carry the per-turn truncation signal: %+v", arts.Result)
	}
}

func TestRunLoopProbeTruncatedReadIsNotEditUnscoreable(t *testing.T) {
	// A truncated READ turn is counted but is not the sandbox-corrupting case:
	// the rationale marks :truncated(turns=N) without the UNSCOREABLE flag.
	client := &loopScript{
		replies:   []string{"CALL read_file rec.md", "FINAL done"},
		truncated: []bool{true, false},
	}
	row, _, _, err := RunLoopProbe(context.Background(), client, "casg-direct", GlyphOnly, 1,
		Materials{Scenario: "S", Glyph: "G"}, "P", map[string]string{"rec.md": "x"}, testSampling(), agentloop.Config{})
	if err != nil {
		t.Fatalf("RunLoopProbe: %v", err)
	}
	if row.Observed.TruncatedEdit {
		t.Fatal("a truncated read must not be flagged TruncatedEdit")
	}
	if !strings.Contains(row.Rationale, ":truncated(turns=1)") || strings.Contains(row.Rationale, "UNSCOREABLE") {
		t.Fatalf("rationale should note a plain truncation, not unscoreable: %q", row.Rationale)
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

// The row carries each turn's parse outcome and the cell's lost and ignored call
// counts, so results.json shows parser losses without reading transcripts.
func TestRunLoopProbeRecordsParseOutcomes(t *testing.T) {
	client := &loopScript{replies: []string{
		"<tool_call>\n<function=call>\n</function>\n</tool_call>",
		"CALL read_file a.md\nCALL read_file b.md",
		"FINAL done",
	}}
	row, _, _, err := RunLoopProbe(context.Background(), client, "item", Baseline, 1,
		Materials{Scenario: "S"}, "P", map[string]string{"a.md": "x"}, testSampling(), agentloop.Config{})
	if err != nil {
		t.Fatal(err)
	}
	o := row.Observed
	if strings.Join(o.ParseOutcomes, ",") != "unknown-tool,parsed,final" {
		t.Errorf("parse_outcomes = %v", o.ParseOutcomes)
	}
	if o.CollapsedCalls == nil || *o.CollapsedCalls != 0 {
		t.Errorf("collapsed_calls = %v, want 0 recorded", o.CollapsedCalls)
	}
	if o.LostCalls == nil || *o.LostCalls != 1 || o.IgnoredCalls == nil || *o.IgnoredCalls != 1 {
		t.Errorf("lost/ignored = %v/%v, want 1/1", o.LostCalls, o.IgnoredCalls)
	}
	if !strings.Contains(row.Rationale, ":lost=1:") {
		t.Errorf("rationale = %q, want lost=1", row.Rationale)
	}
}
