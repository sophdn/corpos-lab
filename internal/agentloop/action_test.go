package agentloop

import "testing"

func TestParseActionToolCalls(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Action
	}{
		{"list", "CALL list_files .", Action{Kind: ActionList, Tool: "list_files", Arg: "."}},
		{"read", "CALL read_file CHANGELOG.md", Action{Kind: ActionRead, Tool: "read_file", Arg: "CHANGELOG.md"}},
		{"query", "CALL run_query accounts region=EU", Action{Kind: ActionQuery, Tool: "run_query", Arg: "accounts region=EU"}},
		{"unknown", "CALL git_status .", Action{Kind: ActionUnknown, Tool: "git_status", Arg: "."}},
		{"final", "FINAL release complete", Action{Kind: ActionFinal, Summary: "release complete"}},
		{"final-bare", "FINAL", Action{Kind: ActionFinal, Summary: ""}},
		{"stall", "I would update the changelog but cannot.", Action{Kind: ActionNone}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseAction(c.in)
			if got.Kind != c.want.Kind || got.Tool != c.want.Tool || got.Arg != c.want.Arg || got.Summary != c.want.Summary {
				t.Fatalf("ParseAction(%q) = %+v, want kind/tool/arg/summary %+v", c.in, got, c.want)
			}
		})
	}
}

func TestParseActionEditSplitsPathAndContents(t *testing.T) {
	got := ParseAction("CALL edit_file CHANGELOG.md ||| # Changelog\n## v1.5.0")
	if got.Kind != ActionEdit {
		t.Fatalf("kind = %v, want ActionEdit", got.Kind)
	}
	if got.Arg != "CHANGELOG.md" {
		t.Fatalf("path = %q, want CHANGELOG.md", got.Arg)
	}
	if got.Content != "# Changelog\n## v1.5.0" {
		t.Fatalf("content = %q", got.Content)
	}
}

func TestParseActionEditWithoutSeparatorHasEmptyContent(t *testing.T) {
	got := ParseAction("CALL edit_file CHANGELOG.md")
	if got.Kind != ActionEdit || got.Arg != "CHANGELOG.md" || got.Content != "" {
		t.Fatalf("got %+v, want edit of CHANGELOG.md with empty content", got)
	}
}

// A turn may open with reasoning; the first directive still wins.
func TestParseActionSkipsLeadingProse(t *testing.T) {
	got := ParseAction("Let me look at the repo first.\nCALL read_file README.md")
	if got.Kind != ActionRead || got.Arg != "README.md" {
		t.Fatalf("got %+v, want read_file README.md past the prose", got)
	}
}

// A FINAL before a CALL wins because it appears first.
func TestParseActionFirstDirectiveWins(t *testing.T) {
	got := ParseAction("FINAL done\nCALL read_file x")
	if got.Kind != ActionFinal {
		t.Fatalf("got %+v, want the earlier FINAL", got)
	}
}

// Qwen3 reverts to its native tool-call syntax mid-run; the parser must read it.
// This is the exact turn-2 shape the setup-vs-agentic-loop smoke captured.
func TestParseActionNativeFunctionTag(t *testing.T) {
	got := ParseAction("<tool_call>\n<function=read_file> CHANGELOG.md\n</function>\n</tool_call>")
	if got.Kind != ActionRead || got.Arg != "CHANGELOG.md" {
		t.Fatalf("native <function=> read = %+v, want read_file CHANGELOG.md", got)
	}
}

func TestParseActionNativeFunctionTagList(t *testing.T) {
	got := ParseAction("<function=list_files> . </function>")
	if got.Kind != ActionList || got.Arg != "." {
		t.Fatalf("native list = %+v", got)
	}
}

func TestParseActionNativeHermesJSON(t *testing.T) {
	got := ParseAction(`<tool_call>{"name": "edit_file", "arguments": {"path": "CHANGELOG.md", "content": "# Changelog\n## v1.5.0"}}</tool_call>`)
	if got.Kind != ActionEdit || got.Arg != "CHANGELOG.md" || got.Content != "# Changelog\n## v1.5.0" {
		t.Fatalf("native JSON edit = %+v", got)
	}
}

func TestParseActionNativeHermesJSONReadQuery(t *testing.T) {
	r := ParseAction(`<tool_call>{"name":"read_file","arguments":{"path":"README.md"}}</tool_call>`)
	if r.Kind != ActionRead || r.Arg != "README.md" {
		t.Fatalf("json read = %+v", r)
	}
	q := ParseAction(`<tool_call>{"name":"run_query","arguments":{"query":"region=EU"}}</tool_call>`)
	if q.Kind != ActionQuery || q.Arg != "region=EU" {
		t.Fatalf("json query = %+v", q)
	}
}

func TestParseActionNativeUnknownTool(t *testing.T) {
	j := ParseAction(`<tool_call>{"name":"git_status","arguments":{}}</tool_call>`)
	if j.Kind != ActionUnknown || j.Tool != "git_status" {
		t.Fatalf("json unknown = %+v", j)
	}
	f := ParseAction("<function=git_status> . </function>")
	if f.Kind != ActionUnknown || f.Tool != "git_status" {
		t.Fatalf("tag unknown = %+v", f)
	}
}

// The native <function=> form of edit, with contents after a newline rather than
// the ||| separator.
func TestParseActionNativeFunctionEditNewline(t *testing.T) {
	got := ParseAction("<function=edit_file>CHANGELOG.md\n# Changelog\n## v1.5.0</function>")
	if got.Kind != ActionEdit || got.Arg != "CHANGELOG.md" || got.Content != "# Changelog\n## v1.5.0" {
		t.Fatalf("native edit newline = %+v", got)
	}
}

// The native <function=> form of edit, using the ||| separator.
func TestParseActionNativeFunctionEditSeparator(t *testing.T) {
	got := ParseAction("<function=edit_file> CHANGELOG.md ||| # Changelog</function>")
	if got.Kind != ActionEdit || got.Arg != "CHANGELOG.md" || got.Content != "# Changelog" {
		t.Fatalf("native edit separator = %+v", got)
	}
}

// A <tool_call> wrapper with malformed JSON and no <function=> tag falls through
// to the CALL/FINAL line scan rather than being read as an action.
func TestParseActionNativeMalformedFallsThrough(t *testing.T) {
	got := ParseAction("<tool_call>{not json}</tool_call>\nFINAL done")
	if got.Kind != ActionFinal {
		t.Fatalf("malformed native should fall through to FINAL, got %+v", got)
	}
}
