package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

// scriptClient returns one canned reply per turn, cycling per run so each cell
// replays the same script. It lets a loop test drive read → edit → FINAL.
type scriptClient struct {
	replies []string
	turn    int
}

func (c *scriptClient) Generate(_ context.Context, _ string, _ model.GenParams) (model.Response, error) {
	text := "FINAL done"
	if c.turn < len(c.replies) {
		text = c.replies[c.turn]
	}
	c.turn++
	return model.Response{Text: text, Model: "qwen", SystemFingerprint: "bX", Timings: model.Timings{PredictedN: 5, PredictedPerSecond: 40}}, nil
}
func (c *scriptClient) Name() string    { return "qwen" }
func (c *scriptClient) Version() string { return "q4" }
func (c *scriptClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func loopSpec() StudySpec {
	return StudySpec{
		Assay:       LoopAssay,
		ItemID:      "casg-direct",
		Model:       ModelSpec{BaseURL: "http://llama-server:8081/v1", ModelID: "qwen", Version: "q4", Endpoint: "completion", PromptTemplate: "{prompt}"},
		Conditions:  []assay.Condition{assay.Baseline, assay.GlyphOnly},
		RunsPerCell: 1,
		Materials:   MaterialsSpec{Scenario: "scenario.md", Glyph: "glyph.md"},
		Sampling:    neutralChain(0.8, []int{1}),
		Loop:        LoopSpec{Preamble: "preamble.md", Sandbox: "sandbox.json", StepCap: 6, CallTokens: 128},
	}
}

func TestExecuteRunsLoopAssay(t *testing.T) {
	sandbox := `{"CHANGELOG.md":"# Changelog\n## v1.4.0"}`
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "SCENARIO", "glyph.md": "GLYPH",
		"preamble.md": "PREAMBLE tools here", "sandbox.json": sandbox,
	})
	out := t.TempDir()

	client := &scriptClient{replies: []string{
		"CALL read_file CHANGELOG.md",
		"CALL edit_file CHANGELOG.md ||| # Changelog\n## v1.5.0\n- ChainedFilter",
		"FINAL updated changelog",
	}}
	results, err := Execute(context.Background(), in, out, client)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	// 2 conditions × 1 run = 2 rows.
	if len(results.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(results.Rows))
	}
	if results.Assay != LoopAssay {
		t.Fatalf("assay = %q", results.Assay)
	}
	if !strings.Contains(results.Rows[0].Rationale, "agentic-loop-probe") {
		t.Fatalf("rationale = %q", results.Rows[0].Rationale)
	}

	// The response file is the flat transcript, carrying the fed-back observation.
	tr, err := os.ReadFile(filepath.Join(out, "responses", "baseline_1.txt"))
	if err != nil {
		t.Fatalf("read transcript: %v", err)
	}
	if !strings.Contains(string(tr), "OBSERVATION:") || !strings.Contains(string(tr), "[turn 1]") {
		t.Fatalf("transcript missing turns/observations:\n%s", tr)
	}

	// The base prompt was written (preamble + aided scenario).
	pr, err := os.ReadFile(filepath.Join(out, "prompts", "baseline_1.txt"))
	if err != nil || !strings.Contains(string(pr), "PREAMBLE") {
		t.Fatalf("prompt file missing preamble: %v / %q", err, pr)
	}

	// The final sandbox snapshot was written and reflects the edit.
	snapRaw, err := os.ReadFile(filepath.Join(out, "sandboxes", "baseline_1.txt"))
	if err != nil {
		t.Fatalf("read sandbox snapshot: %v", err)
	}
	var snap map[string]string
	if err := json.Unmarshal(snapRaw, &snap); err != nil {
		t.Fatalf("snapshot not JSON: %v", err)
	}
	if !strings.Contains(snap["CHANGELOG.md"], "v1.5.0") {
		t.Fatalf("snapshot CHANGELOG.md = %q, want the edit", snap["CHANGELOG.md"])
	}
}

// The loop's stop list is a real generation parameter, not an invisible Go
// constant: whatever halted each turn must appear in the run record, and it
// must be exactly what was sent (record-what-ran).
func TestExecuteLoopRecordsDefaultStop(t *testing.T) {
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "preamble.md": "P", "sandbox.json": `{"a":"b"}`,
	})
	client := &fakeClient{text: "FINAL done"} // terminate on turn 1
	results, err := Execute(context.Background(), in, t.TempDir(), client)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(results.Sampler.Stop) != 1 || results.Sampler.Stop[0] != "\nOBSERVATION" {
		t.Fatalf("recorded stop = %v, want the default OBSERVATION boundary", results.Sampler.Stop)
	}
	// The recorded stop is exactly what reached the wire.
	if len(client.gotParams) == 0 {
		t.Fatal("client received no generation call")
	}
	sent := client.gotParams[0].Stop
	if len(sent) != 1 || sent[0] != "\nOBSERVATION" {
		t.Fatalf("sent stop = %v, want the default; record and wire must agree", sent)
	}
}

func TestExecuteLoopRecordsDeclaredStopOverride(t *testing.T) {
	spec := loopSpec()
	spec.Loop.Stop = []string{"\nOBS", "\nRESULT"}
	in := writeStudy(t, spec, map[string]string{
		"scenario.md": "S", "glyph.md": "G", "preamble.md": "P", "sandbox.json": `{"a":"b"}`,
	})
	client := &fakeClient{text: "FINAL done"}
	results, err := Execute(context.Background(), in, t.TempDir(), client)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	want := []string{"\nOBS", "\nRESULT"}
	if len(results.Sampler.Stop) != 2 || results.Sampler.Stop[0] != want[0] || results.Sampler.Stop[1] != want[1] {
		t.Fatalf("recorded stop = %v, want the declared override %v", results.Sampler.Stop, want)
	}
	sent := client.gotParams[0].Stop
	if len(sent) != 2 || sent[0] != want[0] || sent[1] != want[1] {
		t.Fatalf("sent stop = %v, want the declared override %v", sent, want)
	}
}

// The single-turn grounded-glyph probe parses full completions and sends no
// stop, so its record shows none rather than a borrowed loop default.
func TestExecuteSingleTurnRecordsNoStop(t *testing.T) {
	in := writeStudy(t, groundedSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "ground.md": "GR",
	})
	client := &fakeClient{text: "PASS", name: "qwen"}
	results, err := Execute(context.Background(), in, t.TempDir(), client)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(results.Sampler.Stop) != 0 {
		t.Fatalf("single-turn recorded stop = %v, want none", results.Sampler.Stop)
	}
	if len(client.gotParams[0].Stop) != 0 {
		t.Fatalf("single-turn sent stop = %v, want none", client.gotParams[0].Stop)
	}
}

func TestExecuteLoopMissingPreambleFile(t *testing.T) {
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "sandbox.json": `{"a":"b"}`,
		// preamble.md omitted on purpose.
	})
	if _, err := Execute(context.Background(), in, t.TempDir(), &scriptClient{}); err == nil {
		t.Fatal("expected an error when the loop preamble file is missing")
	}
}

func TestExecuteLoopBadSandboxJSON(t *testing.T) {
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "preamble.md": "P",
		"sandbox.json": `["not","an","object"]`,
	})
	if _, err := Execute(context.Background(), in, t.TempDir(), &scriptClient{}); err == nil {
		t.Fatal("expected an error when the sandbox is not a JSON path→contents object")
	}
}

func TestLoadSpecRejectsLoopAssayWithoutLoopFields(t *testing.T) {
	spec := loopSpec()
	spec.Loop = LoopSpec{} // strip the required preamble/sandbox
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S", "glyph.md": "G"})
	if _, err := LoadSpec(in); err == nil {
		t.Fatal("expected LoadSpec to reject a loop assay with no loop.preamble/sandbox")
	}
}

func TestExecuteLoopMissingSandboxFile(t *testing.T) {
	in := writeStudy(t, loopSpec(), map[string]string{
		"scenario.md": "S", "glyph.md": "G", "preamble.md": "P",
		// sandbox.json omitted on purpose.
	})
	if _, err := Execute(context.Background(), in, t.TempDir(), &scriptClient{}); err == nil {
		t.Fatal("expected an error when the loop sandbox file is missing")
	}
}

func TestLoadSpecRejectsLoopAssayMissingSandbox(t *testing.T) {
	spec := loopSpec()
	spec.Loop = LoopSpec{Preamble: "preamble.md"} // sandbox left empty
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S", "glyph.md": "G", "preamble.md": "P"})
	if _, err := LoadSpec(in); err == nil {
		t.Fatal("expected rejection when loop.sandbox is empty")
	}
}

// The single-turn path writes a reasoning trace when the client surfaces one.
func TestExecuteWritesReasoningTrace(t *testing.T) {
	spec := StudySpec{
		Assay:       SupportedAssay,
		ItemID:      "casg-direct",
		Model:       ModelSpec{ModelID: "qwen"},
		Conditions:  []assay.Condition{assay.Baseline},
		RunsPerCell: 1,
		Materials:   MaterialsSpec{Scenario: "scenario.md"},
		Sampling:    validSampling(),
	}
	in := writeStudy(t, spec, map[string]string{"scenario.md": "S"})
	out := t.TempDir()
	f := &fakeClient{resp: &model.Response{Text: "answer", Reasoning: "the trace"}}
	if _, err := Execute(context.Background(), in, out, f); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(out, "reasoning", "baseline_1.txt"))
	if err != nil || string(b) != "the trace" {
		t.Fatalf("reasoning trace = %q, err %v", b, err)
	}
}

func TestSupportedAssays(t *testing.T) {
	if !SupportedAssays(SupportedAssay) || !SupportedAssays(LoopAssay) {
		t.Fatal("both implemented assays must be supported")
	}
	if SupportedAssays("nonexistent-probe") {
		t.Fatal("an unknown assay must not be supported")
	}
}
