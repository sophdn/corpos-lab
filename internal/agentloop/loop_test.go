package agentloop

import (
	"context"
	"errors"
	"strings"
	"testing"

	"corpos-lab/internal/model"
)

// scriptedClient returns one canned reply per turn, and records the prompt it
// was sent so a test can assert the transcript grew with real observations.
type scriptedClient struct {
	replies []string
	prompts []string
	err     error
	turn    int
}

func (c *scriptedClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	if c.err != nil {
		return model.Response{}, c.err
	}
	c.prompts = append(c.prompts, prompt)
	text := ""
	if c.turn < len(c.replies) {
		text = c.replies[c.turn]
	}
	c.turn++
	return model.Response{
		Text:              text,
		Model:             "fake-model",
		SystemFingerprint: "bTEST",
		Timings:           model.Timings{PredictedN: 7, PredictedPerSecond: 42.0},
	}, nil
}

func (c *scriptedClient) Name() string    { return "fake-model" }
func (c *scriptedClient) Version() string { return "v0" }
func (c *scriptedClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func baseParams() model.GenParams {
	return model.GenParams{Temperature: model.Float64(0.8), Seed: model.Int(1)}
}

func TestRunActsThenFinishes(t *testing.T) {
	client := &scriptedClient{replies: []string{
		"CALL read_file CHANGELOG.md",
		"CALL edit_file CHANGELOG.md ||| # Changelog\n## v1.5.0\n- ChainedFilter\n- NullFilter fix",
		"FINAL updated the changelog for v1.5.0",
	}}
	sb := NewSandbox(map[string]string{"CHANGELOG.md": "# Changelog\n## v1.4.0"})

	res, err := Run(context.Background(), client, "PREAMBLE + scenario", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalFinal {
		t.Fatalf("terminal = %q, want final", res.Terminal)
	}
	if res.Turns != 3 {
		t.Fatalf("turns = %d, want 3", res.Turns)
	}
	if res.Edits != 1 {
		t.Fatalf("edits = %d, want 1", res.Edits)
	}
	if !strings.Contains(sb.Snapshot()["CHANGELOG.md"], "v1.5.0") {
		t.Fatalf("sandbox CHANGELOG.md was not updated: %q", sb.Snapshot()["CHANGELOG.md"])
	}
	// The second prompt must carry the first turn's real observation.
	if len(client.prompts) < 2 || !strings.Contains(client.prompts[1], "OBSERVATION: # Changelog\n## v1.4.0") {
		t.Fatalf("turn 2 prompt did not carry the read observation: %q", client.prompts[1])
	}
	// Provenance aggregates across turns.
	if res.PredictedTokens != 21 || res.TokensPerSecond != 42.0 || res.Model != "fake-model" {
		t.Fatalf("provenance = %d tok / %.1f tps / %s", res.PredictedTokens, res.TokensPerSecond, res.Model)
	}
}

func TestRunStallEndsRun(t *testing.T) {
	client := &scriptedClient{replies: []string{
		"I recognise the changelog must be updated, but I will just describe it.",
	}}
	sb := NewSandbox(map[string]string{"CHANGELOG.md": "x"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalStall {
		t.Fatalf("terminal = %q, want stall", res.Terminal)
	}
	if res.Edits != 0 {
		t.Fatalf("edits = %d, want 0 on a stall", res.Edits)
	}
}

func TestRunHitsStepCap(t *testing.T) {
	// Always reads, never finishes — must stop at the cap.
	client := &scriptedClient{replies: []string{
		"CALL read_file a", "CALL read_file a", "CALL read_file a",
	}}
	sb := NewSandbox(map[string]string{"a": "1"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{StepCap: 2})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalCap {
		t.Fatalf("terminal = %q, want cap", res.Terminal)
	}
	if res.Turns != 2 {
		t.Fatalf("turns = %d, want 2 (the cap)", res.Turns)
	}
}

func TestRunSetsStopAndTokenCap(t *testing.T) {
	// A capturing client asserts the loop overrode MaxTokens and set the stop list.
	var gotParams model.GenParams
	client := &capturingClient{onCall: func(p model.GenParams) { gotParams = p }, reply: "FINAL done"}
	sb := NewSandbox(nil)
	if _, err := Run(context.Background(), client, "base", baseParams(), sb, Config{CallTokens: 128}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotParams.MaxTokens == nil || *gotParams.MaxTokens != 128 {
		t.Fatalf("call tokens = %v, want 128", gotParams.MaxTokens)
	}
	if len(gotParams.Stop) != 2 {
		t.Fatalf("stop = %v, want the default two-element boundary list", gotParams.Stop)
	}
}

func TestRunPropagatesModelError(t *testing.T) {
	client := &scriptedClient{err: errors.New("boom")}
	sb := NewSandbox(nil)
	if _, err := Run(context.Background(), client, "base", baseParams(), sb, Config{}); err == nil {
		t.Fatal("expected the model error to abort the run")
	}
}

func TestTranscriptRendersTurnsAndCapNote(t *testing.T) {
	client := &scriptedClient{replies: []string{"CALL read_file a", "CALL read_file a"}}
	sb := NewSandbox(map[string]string{"a": "1"})
	res, _ := Run(context.Background(), client, "base", baseParams(), sb, Config{StepCap: 2})
	tr := res.Transcript()
	if !strings.Contains(tr, "[turn 1]") || !strings.Contains(tr, "OBSERVATION: 1") {
		t.Fatalf("transcript missing turns/observations:\n%s", tr)
	}
	if !strings.Contains(tr, "step cap reached") {
		t.Fatalf("capped transcript should note the cap:\n%s", tr)
	}
}

func TestActionLabelCoversEveryKind(t *testing.T) {
	cases := map[ActionKind]string{
		ActionList:    "list_files .",
		ActionRead:    "read_file a",
		ActionQuery:   "run_query q",
		ActionEdit:    "edit_file f",
		ActionFinal:   "FINAL",
		ActionUnknown: "unknown:git",
		ActionNone:    "none",
	}
	arg := map[ActionKind]string{ActionList: ".", ActionRead: "a", ActionQuery: "q", ActionEdit: "f"}
	for kind, want := range cases {
		got := actionLabel(Action{Kind: kind, Arg: arg[kind], Tool: "git"})
		if got != want {
			t.Fatalf("actionLabel(%v) = %q, want %q", kind, got, want)
		}
	}
}

func TestRunHonoursCustomStop(t *testing.T) {
	var gotParams model.GenParams
	client := &capturingClient{onCall: func(p model.GenParams) { gotParams = p }, reply: "FINAL done"}
	custom := []string{"###STOP###"}
	if _, err := Run(context.Background(), client, "base", baseParams(), NewSandbox(nil), Config{Stop: custom}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(gotParams.Stop) != 1 || gotParams.Stop[0] != "###STOP###" {
		t.Fatalf("stop = %v, want the custom override", gotParams.Stop)
	}
}

func TestTranscriptFinalRun(t *testing.T) {
	client := &scriptedClient{replies: []string{"FINAL all set"}}
	res, _ := Run(context.Background(), client, "base", baseParams(), NewSandbox(nil), Config{})
	tr := res.Transcript()
	if !strings.Contains(tr, "[turn 1] FINAL all set") || strings.Contains(tr, "step cap") {
		t.Fatalf("final transcript = %q", tr)
	}
}

// A path that equals the queried directory contributes no child (the rest==""
// branch of list).
func TestSandboxListPathEqualsDir(t *testing.T) {
	sb := NewSandbox(map[string]string{"dir/": "placeholder", "dir/a.md": "x"})
	if got := sb.list("dir/"); got != "a.md" {
		t.Fatalf("list(dir/) = %q, want just a.md", got)
	}
}

// capturingClient reports the params of each call and returns one fixed reply.
type capturingClient struct {
	onCall func(model.GenParams)
	reply  string
}

func (c *capturingClient) Generate(_ context.Context, _ string, p model.GenParams) (model.Response, error) {
	c.onCall(p)
	return model.Response{Text: c.reply}, nil
}
func (c *capturingClient) Name() string    { return "cap" }
func (c *capturingClient) Version() string { return "v0" }
func (c *capturingClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}
