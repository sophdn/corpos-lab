package loopcells

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Cell 1 reads two files, claims two edits in one turn (the second is an echo
// the harness never ran), and finishes. Cell 2 is an older row with no parse
// fields: one bare-parameter list (tag text in the argument) and an unknown tool.
// Cell 3 is a cell error.
const transcript1 = `[turn 1] Let me look.
CALL read_file config/app.yaml
EXECUTED: read_file config/app.yaml
OBSERVATION: timeout: 30
retries: 2
[turn 2] CALL read_file docs/notes.md
EXECUTED: read_file docs/notes.md
OBSERVATION: file not found: docs/notes.md
[turn 3] CALL edit_file config/app.yaml ||| timeout: 60
retries: 2
CALL edit_file CHANGELOG.md ||| bumped
EXECUTED: edit_file config/app.yaml
OBSERVATION: wrote config/app.yaml (23 bytes)
[turn 4] FINAL raised the timeout to 60`

const transcript2 = `[turn 1] <tool_call>
<function=list_files>
<parameter>
.
</parameter>
</tool_call>
EXECUTED: list_files <parameter>
.
</parameter>
</tool_call>
OBSERVATION: (empty directory)
[turn 2] <tool_call>
<function=call>
</function>
</tool_call>
EXECUTED: unknown:call
OBSERVATION: unknown tool: call (tools: list_files, read_file, run_query, edit_file)
[turn 2 truncated at the call_tokens cap — this turn's action may be cut off mid-write]
[collapsed: 4 repeated identical calls in turn 2 were not run again]
[loop ended: step cap reached]`

func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "in", "study.json"), `{"loop": {"sandbox": "sandbox.json"}}`)
	write(t, filepath.Join(dir, "in", "sandbox.json"), `{"config/app.yaml": "timeout: 30\nretries: 2", "README.md": "hi"}`)
	write(t, filepath.Join(dir, "out", "results.json"), `{
  "item_id": "item",
  "sampler": {"seeds": [11, 22, 33]},
  "rows": [
    {"condition": "baseline", "run": 1, "rationale": "agentic-loop-probe:baseline:terminal=final:turns=4:edits=1:lost=0:unscored",
     "observed": {"parse_outcomes": ["parsed", "parsed", "parsed", "final"], "lost_calls": 0, "ignored_calls": 1}},
    {"condition": "baseline", "run": 2, "rationale": "agentic-loop-probe:baseline:terminal=cap:turns=2:edits=0:unscored", "observed": {}},
    {"condition": "baseline", "run": 3, "rationale": "agentic-loop-probe:baseline:cell-error:unscoreable:boom", "observed": {}}
  ]}`)
	write(t, filepath.Join(dir, "out", "responses", "baseline_1.txt"), transcript1)
	write(t, filepath.Join(dir, "out", "responses", "baseline_2.txt"), transcript2)
	write(t, filepath.Join(dir, "out", "responses", "baseline_3.txt"), "CELL ERROR (not scoreable): boom")
	write(t, filepath.Join(dir, "out", "sandboxes", "baseline_1.txt"), `{"config/app.yaml": "timeout: 60\nretries: 2", "README.md": "hi", "NEW.md": "x"}`)
	write(t, filepath.Join(dir, "out", "sandboxes", "baseline_2.txt"), `{"config/app.yaml": "timeout: 30\nretries: 2"}`)
	return dir
}

func TestLoadReportsEachCell(t *testing.T) {
	cells, err := Load(fixture(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cells) != 3 {
		t.Fatalf("cells = %d, want 3", len(cells))
	}
	c := cells[0]
	if c.Condition != "baseline" || c.Run != 1 || c.Seed == nil || *c.Seed != 11 {
		t.Errorf("cell 1 identity = %s/%d seed %v", c.Condition, c.Run, c.Seed)
	}
	if c.Terminal != "final" || c.Turns != 4 {
		t.Errorf("cell 1 terminal/turns = %s/%d", c.Terminal, c.Turns)
	}
	if strings.Join(c.Opened, ",") != "config/app.yaml,docs/notes.md" {
		t.Errorf("cell 1 opened = %v", c.Opened)
	}
	if len(c.Edits) != 2 || !c.Edits[0].Executed || c.Edits[0].Path != "config/app.yaml" ||
		c.Edits[1].Executed || c.Edits[1].Path != "CHANGELOG.md" || c.Edits[1].Turn != 3 {
		t.Errorf("cell 1 edits = %+v", c.Edits)
	}
	if c.LostCalls != 0 || c.LostSource != "results" || c.IgnoredCalls == nil || *c.IgnoredCalls != 1 {
		t.Errorf("cell 1 lost = %d (%s) ignored %v", c.LostCalls, c.LostSource, c.IgnoredCalls)
	}
	if strings.Join(c.ParseOutcomes, ",") != "parsed,parsed,parsed,final" {
		t.Errorf("cell 1 parse outcomes = %v", c.ParseOutcomes)
	}
	if got := changes(c.Changed); got != "NEW.md:added,config/app.yaml:modified" {
		t.Errorf("cell 1 changed = %s", got)
	}
	if c.Final != "FINAL raised the timeout to 60" {
		t.Errorf("cell 1 final = %q", c.Final)
	}

	c = cells[1]
	if c.Terminal != "cap" || c.LostCalls != 2 || c.LostSource != "transcript" || c.TruncatedTurns != 1 || c.CollapsedCalls != 4 {
		t.Errorf("cell 2 = terminal %s lost %d (%s) truncated %d", c.Terminal, c.LostCalls, c.LostSource, c.TruncatedTurns)
	}
	if got := changes(c.Changed); got != "README.md:removed" {
		t.Errorf("cell 2 changed = %s", got)
	}
	if c.Final != "[loop ended: step cap reached]" {
		t.Errorf("cell 2 final = %q", c.Final)
	}

	c = cells[2]
	if c.Terminal != "cell-error" || c.Final != "CELL ERROR (not scoreable): boom" || c.Seed == nil || *c.Seed != 33 {
		t.Errorf("cell 3 = %+v", c)
	}
}

func changes(cs []Change) string {
	var parts []string
	for _, c := range cs {
		parts = append(parts, c.Path+":"+c.Kind)
	}
	return strings.Join(parts, ",")
}

func TestLoadAcceptsTheOutDir(t *testing.T) {
	dir := fixture(t)
	cells, err := Load(filepath.Join(dir, "out"))
	if err != nil || len(cells) != 3 || len(cells[0].Changed) == 0 {
		t.Fatalf("Load(out) = %d cells, err %v", len(cells), err)
	}
}

func TestLoadMissingResultsNamesThePath(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), filepath.Join(dir, "out", "results.json")) {
		t.Fatalf("err = %v, want it to name %s", err, filepath.Join(dir, "out", "results.json"))
	}
}

func TestLoadErrors(t *testing.T) {
	t.Run("bad results json", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "out", "results.json"), "{")
		if _, err := Load(dir); err == nil {
			t.Fatal("want an error")
		}
	})
	t.Run("missing transcript", func(t *testing.T) {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "out", "results.json"), `{"rows": [{"condition": "baseline", "run": 1, "rationale": "x"}]}`)
		if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "baseline_1.txt") {
			t.Fatalf("err = %v, want it to name the transcript", err)
		}
	})
	t.Run("bad sandbox json", func(t *testing.T) {
		dir := fixture(t)
		write(t, filepath.Join(dir, "out", "sandboxes", "baseline_1.txt"), "{")
		if _, err := Load(dir); err == nil {
			t.Fatal("want an error")
		}
	})
	t.Run("no inputs is not an error", func(t *testing.T) {
		dir := fixture(t)
		if err := os.RemoveAll(filepath.Join(dir, "in")); err != nil {
			t.Fatal(err)
		}
		cells, err := Load(dir)
		if err != nil || cells[0].Changed != nil {
			t.Fatalf("Load without in/ = %v, changed %v; want no change list", err, cells[0].Changed)
		}
	})
}

func TestWriteTableAndJSON(t *testing.T) {
	cells, err := Load(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	var tbl bytes.Buffer
	if err := WriteTable(&tbl, cells); err != nil {
		t.Fatal(err)
	}
	out := tbl.String()
	for _, want := range []string{"baseline_1", "final", "1/2", "config/app.yaml,docs/notes.md", "raised the timeout", "cell-error"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	var js bytes.Buffer
	if err := WriteJSON(&js, cells); err != nil {
		t.Fatal(err)
	}
	var back []Cell
	if err := json.Unmarshal(js.Bytes(), &back); err != nil || len(back) != 3 {
		t.Fatalf("json round trip: %v (%d cells)", err, len(back))
	}
}

func TestShorten(t *testing.T) {
	if got := shorten("a\nb", 10); got != "a b" {
		t.Errorf("shorten newline = %q", got)
	}
	if got := shorten(strings.Repeat("x", 20), 5); got != "xxxx…" {
		t.Errorf("shorten long = %q", got)
	}
}
