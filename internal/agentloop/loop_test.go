package agentloop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"corpos-lab/internal/model"
)

// scriptedClient returns one canned reply per turn, and records the prompt it
// was sent so a test can assert the transcript grew with real observations.
type scriptedClient struct {
	replies []string
	prompts []string
	// truncated, if set, marks per turn (index-aligned with replies) whether the
	// server cut that turn off at the per-call token cap (resp.Truncated).
	truncated []bool
	err       error
	turn      int
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
	trunc := false
	if c.turn < len(c.truncated) {
		trunc = c.truncated[c.turn]
	}
	c.turn++
	return model.Response{
		Text:              text,
		Model:             "fake-model",
		SystemFingerprint: "bTEST",
		Truncated:         trunc,
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

// nativeRead renders one Qwen native tool-call block that reads path, in the
// <tool_call><function=read_file><parameter=path>…</parameter></function></tool_call>
// shape the model actually emits (CALIBRATION run and bug 1428 transcripts).
func nativeRead(path string) string {
	return "<tool_call>\n<function=read_file>\n<parameter=path>\n" + path + "\n</parameter>\n</function>\n</tool_call>"
}

func TestRunExecutesAllNativeCallsInaTurn(t *testing.T) {
	// Bug 1428: Qwen reverts to native tool-call XML and dumps MANY calls in one
	// turn ("let me check … in parallel"). The old loop executed only the first
	// call, so the model never saw its other reads answered and re-read the same
	// file until the step cap — a NULL cell. The loop must execute EVERY native
	// call in the turn and feed back one OBSERVATION per call.
	client := &scriptedClient{replies: []string{
		// One turn, two native reads — the exact multi-call shape from the trace.
		nativeRead("README.md") + "\n" + nativeRead("src/duration.py"),
		// A second multi-call turn reads the remaining two files.
		nativeRead("ci/last_run.log") + "\n" + nativeRead("tests/test_duration.py"),
		"FINAL CHANGES-NEEDED: parse_duration drops the minutes component",
	}}
	sb := NewSandbox(map[string]string{
		"README.md":              "Approve once the suite passes.",
		"src/duration.py":        "def parse_duration(s): return 0",
		"ci/last_run.log":        "1 failed, 2 passed",
		"tests/test_duration.py": "assert parse_duration('1h30m') == 5400",
	})

	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalFinal {
		t.Fatalf("terminal = %q, want final — the read-heavy turn must reach a decision", res.Terminal)
	}
	// The SECOND prompt the model sees must carry BOTH of turn 1's observations.
	// Before the fix it carried only the first read, which is the whole bug.
	if len(client.prompts) < 2 {
		t.Fatalf("expected at least 2 prompts, got %d", len(client.prompts))
	}
	p2 := client.prompts[1]
	if !strings.Contains(p2, "Approve once the suite passes.") {
		t.Fatalf("turn 2 prompt missing the README observation:\n%s", p2)
	}
	if !strings.Contains(p2, "def parse_duration(s): return 0") {
		t.Fatalf("turn 2 prompt missing the duration.py observation — only the first call ran:\n%s", p2)
	}
}

func TestRunBreaksRepeatedIdenticalCallLoop(t *testing.T) {
	// Bug 1428, the across-turns half: the model emits the SAME native read every
	// turn, ignoring the observation. The loop must detect the unproductive
	// repeated-identical-call loop and stop with a distinct signal before the step
	// cap, not burn all 16 turns re-reading one file.
	same := nativeRead("src/duration.py")
	client := &scriptedClient{replies: []string{same, same, same, same, same, same}}
	sb := NewSandbox(map[string]string{"src/duration.py": "def parse_duration(s): return 0"})

	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{StepCap: 16})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalLoop {
		t.Fatalf("terminal = %q, want loop — the repeated read must break the loop", res.Terminal)
	}
	if res.Turns >= 16 {
		t.Fatalf("turns = %d, want well under the cap of 16", res.Turns)
	}
	tr := res.Transcript()
	if !strings.Contains(tr, "repeated identical") {
		t.Fatalf("transcript did not surface the repeated-call loop break:\n%s", tr)
	}
}

// nativeEditPathAsName renders the native-XML edit_file shape from bug 1430: the
// path is the FIRST parameter's NAME and the body rides a <parameter=new_content>
// block. This is what Qwen3 emits for its write, and it used to produce a 0-byte
// write that nulled the decision.
func nativeEditPathAsName(path, body string) string {
	return "<function=edit_file>\n<parameter=" + path + ">\nx\n</parameter>\n<parameter=new_content>\n" + body + "\n</parameter>\n</function>"
}

func TestRunWritesNativeXMLEditDecision(t *testing.T) {
	// Bug 1430, the write-path half: a native-XML edit_file turn must WRITE the
	// file with its content, not a 0-byte file. Before the fix the write landed
	// empty, the OBSERVATION read "wrote (0 bytes)", and the reached decision
	// scored NULL. This is the write analogue of the read-loop regression above.
	client := &scriptedClient{replies: []string{
		nativeEditPathAsName("REVIEW.md", "CHANGES-NEEDED: proxy rule rejected"),
		"FINAL recorded the decision",
	}}
	sb := NewSandbox(map[string]string{"REVIEW.md": ""})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalFinal {
		t.Fatalf("terminal = %q, want final", res.Terminal)
	}
	if res.Edits != 1 {
		t.Fatalf("edits = %d, want 1", res.Edits)
	}
	if got := sb.Snapshot()["REVIEW.md"]; got != "CHANGES-NEEDED: proxy rule rejected" {
		t.Fatalf("REVIEW.md = %q, want the written decision", got)
	}
	// The second prompt must carry a real write observation, not a 0-byte write.
	if len(client.prompts) < 2 {
		t.Fatalf("expected at least 2 prompts, got %d", len(client.prompts))
	}
	p2 := client.prompts[1]
	if !strings.Contains(p2, "wrote REVIEW.md (") || strings.Contains(p2, "(0 bytes)") {
		t.Fatalf("turn 2 prompt did not carry a real write observation:\n%s", p2)
	}
}

func TestRunBreaksAlternatingNearIdenticalLoop(t *testing.T) {
	// Bug 1430, the balloon half: Qwen loops on near-identical calls whose
	// ARGUMENTS alternate, so no two consecutive turns are byte-identical and the
	// old consecutive-only detector never fired — the transcript grew until it
	// overflowed n_ctx and the server returned 400. The loop must break a
	// signature that RECURS inside a window, not only one that repeats back-to-back.
	a := nativeRead("a.md")
	b := nativeRead("b.md")
	// a, b, a, b, … — each read recurs every other turn, never consecutively.
	client := &scriptedClient{replies: []string{a, b, a, b, a, b, a, b, a, b}}
	sb := NewSandbox(map[string]string{"a.md": "A", "b.md": "B"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{StepCap: 16})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalLoop {
		t.Fatalf("terminal = %q, want loop — the alternating repeat must break", res.Terminal)
	}
	if res.Turns >= 16 {
		t.Fatalf("turns = %d, want well under the cap of 16", res.Turns)
	}
}

func TestRunRecordsTruncatedEditTurn(t *testing.T) {
	// The edit turn is cut off at the per-call token cap mid-record. The loop must
	// record it as a per-turn, edit-carrying truncation — the dangerous case that
	// corrupts the sandbox a deterministic end-state scorer reads — not just flip
	// the coarse aggregate bool.
	client := &scriptedClient{
		replies: []string{
			"CALL read_file CHANGELOG.md",
			"CALL edit_file CHANGELOG.md ||| # Changelog\n## v1.5.0\n- field one\n- field two", // chopped here
			"FINAL done",
		},
		truncated: []bool{false, true, false},
	}
	sb := NewSandbox(map[string]string{"CHANGELOG.md": "x"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Truncated {
		t.Fatal("aggregate Truncated must stay true for compatibility")
	}
	if res.TruncatedTurns != 1 {
		t.Fatalf("TruncatedTurns = %d, want 1", res.TruncatedTurns)
	}
	if !res.TruncatedEdit {
		t.Fatal("TruncatedEdit must be true — a truncated turn carried an edit_file action")
	}
	// The truncated step itself must carry the per-turn flag (it was the edit turn).
	if len(res.Steps) < 2 || !res.Steps[1].Truncated {
		t.Fatalf("edit step should be marked truncated: %+v", res.Steps)
	}
	if res.Steps[0].Truncated {
		t.Fatal("the read turn was not truncated and must not be marked so")
	}
	// The transcript must surface the truncated-edit so a reader sees it without
	// reading provenance JSON.
	tr := res.Transcript()
	if !strings.Contains(tr, "truncated") {
		t.Fatalf("transcript did not surface truncation:\n%s", tr)
	}
}

func TestRunTruncatedNonEditTurnIsNotEditTruncation(t *testing.T) {
	// A truncated turn that is NOT an edit (a read) is counted but is NOT the
	// dangerous end-state-corrupting case: TruncatedEdit stays false.
	client := &scriptedClient{
		replies:   []string{"CALL read_file a", "FINAL done"},
		truncated: []bool{true, false},
	}
	sb := NewSandbox(map[string]string{"a": "1"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.TruncatedTurns != 1 {
		t.Fatalf("TruncatedTurns = %d, want 1", res.TruncatedTurns)
	}
	if res.TruncatedEdit {
		t.Fatal("TruncatedEdit must be false — the truncated turn was a read, not an edit")
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
	if len(gotParams.Stop) != 1 || gotParams.Stop[0] != "\nOBSERVATION" {
		t.Fatalf("stop = %v, want just the fabricated-observation boundary", gotParams.Stop)
	}
}

func TestConfigEffectiveStop(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{"default when unset", Config{}, []string{"\nOBSERVATION"}},
		{"override when set", Config{Stop: []string{"###STOP###"}}, []string{"###STOP###"}},
		{"multiple overrides", Config{Stop: []string{"\nOBS", "\nRESULT"}}, []string{"\nOBS", "\nRESULT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EffectiveStop()
			if len(got) != len(tt.want) {
				t.Fatalf("EffectiveStop() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("EffectiveStop()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
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

func TestTranscriptMarksExecutedActionAmidEchoes(t *testing.T) {
	// The turn text carries prose and an echoed second CALL the harness never ran;
	// the parser takes the first directive. EXECUTED must name that one action so a
	// reader does not score the echo.
	client := &scriptedClient{replies: []string{
		"Let me inspect it first.\nCALL read_file a\nCALL edit_file a ||| noise",
		"FINAL done",
	}}
	sb := NewSandbox(map[string]string{"a": "1"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	tr := res.Transcript()
	if !strings.Contains(tr, "EXECUTED: read_file a\nOBSERVATION: 1") {
		t.Fatalf("transcript did not mark the executed read before its observation:\n%s", tr)
	}
	// The echoed edit is visible in the raw turn text but was never executed.
	if !strings.Contains(tr, "CALL edit_file a ||| noise") {
		t.Fatalf("raw turn text should be preserved for provenance:\n%s", tr)
	}
	if res.Edits != 0 {
		t.Fatalf("edits = %d, want 0 — the echoed edit must not have run", res.Edits)
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

// Each turn records its parse outcome and the run sums lost and ignored calls,
// so a reader sees parser losses without grepping the transcript.
func TestRunRecordsParseOutcomes(t *testing.T) {
	c := &scriptedClient{replies: []string{
		"<tool_call>\n<function=call>\n</function>\n</tool_call>",
		"CALL read_file a.md\nCALL read_file b.md",
		"FINAL done",
	}}
	res, err := Run(context.Background(), c, "BASE", baseParams(), NewSandbox(map[string]string{"a.md": "x"}), Config{})
	if err != nil {
		t.Fatal(err)
	}
	want := []ParseOutcome{OutcomeUnknownTool, OutcomeParsed, OutcomeFinal}
	if len(res.Steps) != len(want) {
		t.Fatalf("steps = %d, want %d", len(res.Steps), len(want))
	}
	for i, w := range want {
		if res.Steps[i].Parse != w {
			t.Errorf("turn %d parse = %s, want %s", i+1, res.Steps[i].Parse, w)
		}
	}
	if got := res.ParseOutcomes(); len(got) != 3 || got[0] != string(OutcomeUnknownTool) {
		t.Errorf("ParseOutcomes = %v", got)
	}
	if res.LostCalls != 1 || res.IgnoredCalls != 1 {
		t.Errorf("lost = %d ignored = %d, want 1 and 1", res.LostCalls, res.IgnoredCalls)
	}
}

// overflowClient answers turns normally until overflowAt, then refuses the
// request the way llama-server does when the prompt exceeds n_ctx.
type overflowClient struct {
	scriptedClient
	overflowAt int
}

func (c *overflowClient) Generate(ctx context.Context, prompt string, p model.GenParams) (model.Response, error) {
	if c.turn+1 == c.overflowAt {
		return model.Response{}, &model.APIError{ModelID: "m", Op: "completion", StatusCode: 400,
			Body: `{"error":{"code":400,"message":"request (24275 tokens) exceeds the available context size (16384 tokens), try increasing it","type":"exceed_context_size_error","n_prompt_tokens":24275,"n_ctx":16384}}`}
	}
	return c.scriptedClient.Generate(ctx, prompt, p)
}

// bug 1442: a prompt that outgrew n_ctx made the server return 400, and the cell
// was recorded as CELL ERROR with no transcript or sandbox. The loop now ends the
// cell with the "context" terminal and keeps what the subject did.
func TestRunEndsWithContextTerminalOnOverflow(t *testing.T) {
	c := &overflowClient{scriptedClient: scriptedClient{replies: []string{
		"CALL edit_file notes.md ||| new body",
		"CALL read_file notes.md",
	}}, overflowAt: 3}
	sb := NewSandbox(map[string]string{"notes.md": "old"})
	res, err := Run(context.Background(), c, "BASE", baseParams(), sb, Config{})
	if err != nil {
		t.Fatalf("Run returned %v, want a clean context terminal", err)
	}
	if res.Terminal != TerminalContext || res.Turns != 2 || len(res.Steps) != 2 || res.Edits != 1 {
		t.Fatalf("res = terminal %s turns %d steps %d edits %d, want context/2/2/1", res.Terminal, res.Turns, len(res.Steps), res.Edits)
	}
	if !strings.Contains(res.ContextError, "24275 tokens") {
		t.Errorf("ContextError = %q, want the server's message", res.ContextError)
	}
	if sb.Snapshot()["notes.md"] != "new body" {
		t.Errorf("sandbox not kept")
	}
	if tr := res.Transcript(); !strings.Contains(tr, "[loop ended: the next request exceeds the model context") {
		t.Errorf("transcript has no context note:\n%s", tr)
	}
}

// Each turn records the prompt size the server reported, so a reader can see
// which turn grew the prompt.
func TestRunRecordsPromptTokensPerTurn(t *testing.T) {
	c := &promptSizeClient{sizes: []int{900, 4200}}
	res, err := Run(context.Background(), c, "BASE", baseParams(), NewSandbox(nil), Config{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps[0].PromptTokens != 900 || res.Steps[1].PromptTokens != 4200 || res.PromptTokens != 4200 {
		t.Errorf("prompt tokens = %d, %d (max %d), want 900, 4200 (max 4200)", res.Steps[0].PromptTokens, res.Steps[1].PromptTokens, res.PromptTokens)
	}
}

type promptSizeClient struct {
	sizes []int
	turn  int
}

func (c *promptSizeClient) Generate(context.Context, string, model.GenParams) (model.Response, error) {
	text := "CALL list_files ."
	if c.turn == len(c.sizes)-1 {
		text = "FINAL done"
	}
	n := c.sizes[c.turn]
	c.turn++
	return model.Response{Text: text, Timings: model.Timings{PromptN: n}}, nil
}

func (c *promptSizeClient) Name() string    { return "fake-model" }
func (c *promptSizeClient) Version() string { return "v0" }
func (c *promptSizeClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func TestServerMessage(t *testing.T) {
	if got := serverMessage(`{"error":{"message":"too long"}}`); got != "too long" {
		t.Errorf("serverMessage = %q", got)
	}
	if got := serverMessage("plain text"); got != "plain text" {
		t.Errorf("serverMessage fallback = %q", got)
	}
}

// bug 1442: Qwen repeated the same native read batch inside one turn until it
// filled call_tokens (45 calls over four files), and every repeat fed the whole
// file back, overflowing n_ctx within a few turns. Each distinct call now runs
// once per turn; an exact repeat gets a one-line observation and is counted.
func TestRunCollapsesRepeatedCallsWithinATurn(t *testing.T) {
	read := func(p string) string {
		return "<tool_call>\n<function=read_file>\n<parameter=path>\n" + p + "\n</parameter>\n</function>\n</tool_call>\n"
	}
	turn := read("a.md") + read("b.md") + read("a.md") + read("b.md") + read("a.md")
	c := &scriptedClient{replies: []string{turn, "FINAL done"}}
	res, err := Run(context.Background(), c, "BASE", baseParams(), NewSandbox(map[string]string{"a.md": "AAAA", "b.md": "BBBB"}), Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := res.Steps[0]
	if len(s.Calls) != 2 || s.CollapsedCalls != 3 || res.CollapsedCalls != 3 {
		t.Fatalf("calls %d collapsed %d (run %d), want 2 run and 3 collapsed", len(s.Calls), s.CollapsedCalls, res.CollapsedCalls)
	}
	p2 := c.prompts[1]
	if strings.Count(p2, "AAAA") != 1 || strings.Count(p2, "BBBB") != 1 {
		t.Errorf("file bodies fed back more than once:\n%s", p2)
	}
	if strings.Count(p2, "not run again") != 3 {
		t.Errorf("want 3 short repeat observations:\n%s", p2)
	}
	if tr := res.Transcript(); !strings.Contains(tr, "[collapsed: 3 repeated identical calls in turn 1 were not run again]") {
		t.Errorf("transcript lacks the collapse note:\n%s", tr)
	}
}

// A repeated edit with the same content is a repeat too; a different body for
// the same path is a different call and runs.
func TestRunCollapsesOnlyIdenticalEdits(t *testing.T) {
	turn := "<tool_call>\n{\"name\":\"edit_file\",\"arguments\":{\"path\":\"a.md\",\"content\":\"one\"}}\n</tool_call>\n" +
		"<tool_call>\n{\"name\":\"edit_file\",\"arguments\":{\"path\":\"a.md\",\"content\":\"one\"}}\n</tool_call>\n" +
		"<tool_call>\n{\"name\":\"edit_file\",\"arguments\":{\"path\":\"a.md\",\"content\":\"two\"}}\n</tool_call>\n"
	c := &scriptedClient{replies: []string{turn, "FINAL done"}}
	sb := NewSandbox(nil)
	res, err := Run(context.Background(), c, "BASE", baseParams(), sb, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Steps[0].CollapsedCalls != 1 || len(res.Steps[0].Calls) != 2 || sb.Snapshot()["a.md"] != "two" || sb.Edits() != 2 {
		t.Errorf("collapsed %d calls %d content %q edits %d", res.Steps[0].CollapsedCalls, len(res.Steps[0].Calls), sb.Snapshot()["a.md"], sb.Edits())
	}
}

// TestRunRepeatOutsideTheWindowIsNotALoop pins the window's edge: one read recurs
// three times, but each recurrence is more than repeatWindow turns after the last,
// so the oldest signature has left the window and the loop must run to the cap.
func TestRunRepeatOutsideTheWindowIsNotALoop(t *testing.T) {
	x := nativeRead("x.md")
	replies := []string{x}
	for i := 0; i < 2; i++ {
		for j := 0; j < repeatWindow-1; j++ {
			replies = append(replies, nativeRead(fmt.Sprintf("f%d-%d.md", i, j)))
		}
		replies = append(replies, x)
	}
	client := &scriptedClient{replies: replies}
	sb := NewSandbox(map[string]string{"x.md": "X"})
	res, err := Run(context.Background(), client, "base", baseParams(), sb, Config{StepCap: len(replies)})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Terminal != TerminalCap || res.Turns != len(replies) {
		t.Fatalf("terminal = %q after %d turns, want cap after %d — a repeat outside the window is not a loop", res.Terminal, res.Turns, len(replies))
	}
}
