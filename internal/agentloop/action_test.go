package agentloop

import "testing"

// summarize renders a slice of actions as kind:arg pairs for a compact assertion.
func summarize(acts []Action) []string {
	out := make([]string, len(acts))
	for i, a := range acts {
		out[i] = actionLabel(a)
	}
	return out
}

func TestParseActions(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{
			// The bug-1428 shape: two native tool_call reads in one turn.
			name: "two-native-reads",
			in:   "<tool_call>\n<function=read_file>\n<parameter=path>\nREADME.md\n</parameter>\n</function>\n</tool_call>\n<tool_call>\n<function=read_file>\n<parameter=path>\nsrc/duration.py\n</parameter>\n</function>\n</tool_call>",
			want: []string{"read_file README.md", "read_file src/duration.py"},
		},
		{
			name: "native-reads-with-leading-prose",
			in:   "Let me read both in parallel.\n<tool_call>\n<function=read_file> a\n</function>\n</tool_call>\n<tool_call>\n<function=read_file> b\n</function>\n</tool_call>",
			want: []string{"read_file a", "read_file b"},
		},
		{
			name: "bare-function-blocks-no-tool-call-wrapper",
			in:   "<function=list_files> . </function>\n<function=read_file> README.md </function>",
			want: []string{"list_files .", "read_file README.md"},
		},
		{
			// A single malformed native block parses to nothing, so the CALL/FINAL
			// line scan still finds the FINAL.
			name: "malformed-native-falls-through-to-final",
			in:   "<tool_call>{not json}</tool_call>\nFINAL done",
			want: []string{"FINAL"},
		},
		{
			name: "call-form-single-directive",
			in:   "Let me look first.\nCALL read_file README.md",
			want: []string{"read_file README.md"},
		},
		{
			name: "stall-is-one-none",
			in:   "I would update it but I will only describe it.",
			want: []string{"none"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := summarize(ParseActions(c.in))
			if len(got) != len(c.want) {
				t.Fatalf("ParseActions(%q) = %v, want %v", c.in, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("ParseActions(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestParseActionsEditStripsTrailingFinal(t *testing.T) {
	// A native edit whose body ends with a stray FINAL has that line stripped, the
	// same as the single-action ParseAction path.
	acts := ParseActions("<function=edit_file>\n<parameter=path>\nREVIEW.md\n</parameter>\n<parameter=content>\nAPPROVED\nFINAL done\n</parameter>\n</function>")
	if len(acts) != 1 || acts[0].Kind != ActionEdit {
		t.Fatalf("ParseActions edit = %+v, want one ActionEdit", acts)
	}
	if acts[0].Content != "APPROVED" {
		t.Fatalf("edit content = %q, want the trailing FINAL stripped", acts[0].Content)
	}
}

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

// The bare form: <tool_call> then "toolname args" on its own line — no JSON, no
// <function=>. This is the turn-1 shape the calibration run captured for
// parent-state, which the first parser scored as a stall.
func TestParseActionNativeBareToolCall(t *testing.T) {
	got := ParseAction("<tool_call>\nlist_files .")
	if got.Kind != ActionList || got.Arg != "." {
		t.Fatalf("bare tool_call list = %+v, want list_files .", got)
	}
	rd := ParseAction("<tool_call>\nread_file CHANGELOG.md\n</tool_call>")
	if rd.Kind != ActionRead || rd.Arg != "CHANGELOG.md" {
		t.Fatalf("bare tool_call read = %+v", rd)
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

// The malformed hybrid: Qwen3.8 stuffs the whole "CALL <tool> <arg>" directive
// inside the <function=…> value. These four strings are verbatim from the
// safety-check-bypass loop smoke, where each scored as unknown-tool and lost a
// real read or list. The parser must recover the tool and its argument.
func TestParseActionNativeHybridCallInFunctionTag(t *testing.T) {
	cases := []struct {
		name string
		in   string
		kind ActionKind
		tool string
		arg  string
	}{
		{"list-trailing-slash", "<tool_call>\n<function=CALL list_files tests/></function>\n</tool_call>", ActionList, "list_files", "tests/"},
		{"list-no-space-close", "<function=CALL list_files config</function>", ActionList, "list_files", "config"},
		{"read-path-close", "<function=CALL read_file config/validator.py</function>", ActionRead, "read_file", "config/validator.py"},
		{"read-clean-close", "<function=CALL read_file milestones/release-2026-Q2.md>", ActionRead, "read_file", "milestones/release-2026-Q2.md"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseAction(c.in)
			if got.Kind != c.kind || got.Tool != c.tool || got.Arg != c.arg {
				t.Fatalf("ParseAction(%q) = %+v, want kind %v tool %s arg %q", c.in, got, c.kind, c.tool, c.arg)
			}
		})
	}
}

// The recovery must not fire for a genuinely unknown native call: a clean
// <function=NAME> with an unrecognised tool stays ActionUnknown.
func TestParseActionNativeHybridPreservesUnknown(t *testing.T) {
	got := ParseAction("<function=computer_use> . </function>")
	if got.Kind != ActionUnknown || got.Tool != "computer_use" {
		t.Fatalf("clean unknown native = %+v, want ActionUnknown computer_use", got)
	}
}

// A FINAL declaration the subject leaks onto the last line of an edit_file body
// is stripped from the contents; a FINAL inside the body, or none, is untouched.
func TestParseActionStripsTrailingFinalFromEdit(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"trailing final on its own line",
			"CALL edit_file notes.md ||| # Notes\n- done the work\nFINAL wrote the notes",
			"# Notes\n- done the work"},
		{"trailing bare final",
			"CALL edit_file notes.md ||| body line\nFINAL",
			"body line"},
		{"final inside the body is kept",
			"CALL edit_file notes.md ||| FINAL is a section header\n- item",
			"FINAL is a section header\n- item"},
		{"no trailing final is unchanged",
			"CALL edit_file notes.md ||| just a body\n- item",
			"just a body\n- item"},
		{"body that is only a final collapses to empty",
			"CALL edit_file notes.md ||| FINAL all set",
			""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseAction(c.in)
			if got.Kind != ActionEdit {
				t.Fatalf("ParseAction(%q) kind = %v, want ActionEdit", c.in, got.Kind)
			}
			if got.Content != c.want {
				t.Fatalf("content = %q, want %q", got.Content, c.want)
			}
		})
	}
}

// Qwen3 emits <parameter=KEY>VALUE</parameter> sub-tags inside a <function=NAME>
// call. These are verbatim-shaped from the Item 12 tool-loop batteries, where the
// old parser took the whole body (parameter tags and all) as the freeform arg and
// so lost the real path/dir/query. read_file, list_files and run_query must map
// their single parameter the way parseJSONCall maps the JSON keys.
func TestParseActionNativeParamTagsSingleArg(t *testing.T) {
	cases := []struct {
		name string
		in   string
		kind ActionKind
		tool string
		arg  string
	}{
		{"read",
			"<function=read_file>\n<parameter=path>\nstatus/web-01.txt\n</parameter>\n</function>",
			ActionRead, "read_file", "status/web-01.txt"},
		{"read-in-tool-call",
			"<tool_call>\n<function=read_file>\n<parameter=path>\nstatus/web-01.txt\n</parameter>\n</function>\n</tool_call>",
			ActionRead, "read_file", "status/web-01.txt"},
		{"list",
			"<function=list_files>\n<parameter=dir>\nboard/tasks\n</parameter>\n</function>",
			ActionList, "list_files", "board/tasks"},
		{"query",
			"<function=run_query>\n<parameter=query>\naccounts region=EU\n</parameter>\n</function>",
			ActionQuery, "run_query", "accounts region=EU"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseAction(c.in)
			if got.Kind != c.kind || got.Tool != c.tool || got.Arg != c.arg {
				t.Fatalf("ParseAction(%q) = %+v, want kind %v tool %s arg %q", c.in, got, c.kind, c.tool, c.arg)
			}
		})
	}
}

// edit_file in the parameter-tag form, both with a <parameter=content> block and
// with the hybrid where a <parameter=path> sub-tag is followed by the ||| body.
// The old parser wrote the garbage string "<parameter=path> … </parameter>" as the
// path and the real write never landed (OBSERVATION: wrote <parameter=path> …).
func TestParseActionNativeParamTagsEdit(t *testing.T) {
	t.Run("content-param", func(t *testing.T) {
		got := ParseAction("<tool_call>\n<function=edit_file>\n<parameter=path>\nboard/tasks/search-reindex.md\n</parameter>\n<parameter=content>\n# Reindex\n- step one\n</parameter>\n</function>\n</tool_call>")
		if got.Kind != ActionEdit || got.Arg != "board/tasks/search-reindex.md" {
			t.Fatalf("got %+v, want edit of board/tasks/search-reindex.md", got)
		}
		if got.Content != "# Reindex\n- step one" {
			t.Fatalf("content = %q, want %q", got.Content, "# Reindex\n- step one")
		}
	})
	t.Run("path-param-then-separator", func(t *testing.T) {
		got := ParseAction("<function=edit_file>\n<parameter=path>\nboard/tasks/search-reindex.md\n</parameter>\n||| # Reindex\n- step one</function>")
		if got.Kind != ActionEdit || got.Arg != "board/tasks/search-reindex.md" {
			t.Fatalf("got %+v, want edit of board/tasks/search-reindex.md", got)
		}
		if got.Content != "# Reindex\n- step one" {
			t.Fatalf("content = %q, want %q", got.Content, "# Reindex\n- step one")
		}
	})
}

// A parameter-tag body on an unrecognised tool stays ActionUnknown — the param
// parse must not promote a tool the loop does not provide.
func TestParseActionNativeParamTagsUnknown(t *testing.T) {
	got := ParseAction("<function=computer_use>\n<parameter=path>\nx\n</parameter>\n</function>")
	if got.Kind != ActionUnknown || got.Tool != "computer_use" {
		t.Fatalf("got %+v, want ActionUnknown computer_use", got)
	}
}

// Qwen3 does not hold the preamble's exact parameter names. It writes read_file's
// path as <parameter=filename>, list_files' dir as <parameter=path>, and wraps the
// real tool in a generic <function=call> whose <parameter=tool> names it. Each of
// these was lost to an empty argument before the synonym resolution; they must now
// recover to the same action a correctly-named call produces.
func TestParseActionNativeParamTagsSynonyms(t *testing.T) {
	cases := []struct {
		name string
		in   string
		kind ActionKind
		tool string
		arg  string
	}{
		{"read-filename",
			"<function=read_file>\n<parameter=filename>\nmigrations/0001_init.sql\n</parameter>\n</function>",
			ActionRead, "read_file", "migrations/0001_init.sql"},
		{"read-file-key",
			"<function=read_file>\n<parameter=file>\nsrc/auth.py\n</parameter>\n</function>",
			ActionRead, "read_file", "src/auth.py"},
		{"list-path-key",
			"<function=list_files>\n<parameter=path>\nmigrations/\n</parameter>\n</function>",
			ActionList, "list_files", "migrations/"},
		{"generic-call-wrapper-read",
			"<function=call>\n<parameter=tool>\nread_file\n</parameter>\n<parameter=path>\nmigration/plan.md\n</parameter>\n</function>",
			ActionRead, "read_file", "migration/plan.md"},
		{"generic-call-wrapper-list",
			"<function=call>\n<parameter=tool>\nlist_files\n</parameter>\n<parameter=path>\n.\n</parameter>\n</function>",
			ActionList, "list_files", "."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ParseAction(c.in)
			if got.Kind != c.kind || got.Tool != c.tool || got.Arg != c.arg {
				t.Fatalf("ParseAction(%q) = %+v, want kind %v tool %s arg %q", c.in, got, c.kind, c.tool, c.arg)
			}
		})
	}
}

// edit_file recovers both its path synonym (filename) and its content synonym
// (contents), so a write the model forms with its own parameter names still lands.
func TestParseActionNativeParamTagsEditSynonyms(t *testing.T) {
	got := ParseAction("<function=edit_file>\n<parameter=filename>\nmigration/plan.md\n</parameter>\n<parameter=contents>\n# Plan\n- forward SQL\n</parameter>\n</function>")
	if got.Kind != ActionEdit || got.Arg != "migration/plan.md" {
		t.Fatalf("got %+v, want edit of migration/plan.md", got)
	}
	if got.Content != "# Plan\n- forward SQL" {
		t.Fatalf("content = %q, want %q", got.Content, "# Plan\n- forward SQL")
	}
}

// The Hermes JSON form resolves the same synonyms (filename for read path).
func TestParseActionNativeHermesJSONSynonyms(t *testing.T) {
	got := ParseAction(`<tool_call>{"name":"read_file","arguments":{"filename":"src/session.py"}}</tool_call>`)
	if got.Kind != ActionRead || got.Arg != "src/session.py" {
		t.Fatalf("got %+v, want read_file src/session.py", got)
	}
}

// Qwen3 omits the '>' that closes the function name and writes the <parameter=…>
// tags on the following lines; the real tool must still be recognised and its
// parameter value recovered.
func TestParseActionNativeFunctionMissingCloseBracket(t *testing.T) {
	t.Run("read-path-tag", func(t *testing.T) {
		got := ParseAction("<tool_call>\n<function=read_file\n<parameter=path>\nsrc/auth.py\n</parameter>\n</function>\n</tool_call>")
		if got.Kind != ActionRead || got.Arg != "src/auth.py" {
			t.Fatalf("got %+v, want read_file src/auth.py", got)
		}
	})
	t.Run("edit-path-tag", func(t *testing.T) {
		got := ParseAction("<tool_call>\n<function=edit_file\n<parameter=path>\nreview/round-1.md\n</parameter>\n<parameter=content>\n# done\n</parameter>\n</function>\n</tool_call>")
		if got.Kind != ActionEdit || got.Arg != "review/round-1.md" || got.Content != "# done" {
			t.Fatalf("got %+v, want edit review/round-1.md content '# done'", got)
		}
	})
}

// The hybrid where the function name lacks its '>' AND the arguments arrive as a
// JSON object under <parameter=arguments> — the exact shape that lost the round-1
// write in the scenario-2 grid.
func TestParseActionNativeArgumentsJSON(t *testing.T) {
	got := ParseAction("<tool_call>\n<function=edit_file\n<parameter=path>\n<parameter=arguments>\n{\"path\": \"review/round-1.md\", \"contents\": \"# findings\\n- F1\"}\n</parameter>\n</function>\n</tool_call>")
	if got.Kind != ActionEdit || got.Arg != "review/round-1.md" {
		t.Fatalf("got %+v, want edit of review/round-1.md", got)
	}
	if got.Content != "# findings\n- F1" {
		t.Fatalf("content = %q, want %q", got.Content, "# findings\n- F1")
	}
}

// The generic wrapper naming the tool in a "tool_name" parameter dispatches too.
func TestParseActionNativeToolNameSynonym(t *testing.T) {
	got := ParseAction("<function=call>\n<parameter=tool_name>\nlist_files\n</parameter>\n<parameter=path>\nsrc/\n</parameter>\n</function>")
	if got.Kind != ActionList || got.Tool != "list_files" || got.Arg != "src/" {
		t.Fatalf("got %+v, want list_files src/", got)
	}
}

// A malformed parameter tag with no closing ">" after the key is tolerated: the
// parse stops without inventing a parameter, and the body with no usable tag is
// treated as a bare freeform arg.
func TestParseActionNativeParamTagsMalformed(t *testing.T) {
	got := ParseAction("<function=read_file><parameter=path status/web-01.txt</function>")
	if got.Kind != ActionRead {
		t.Fatalf("got %+v, want ActionRead (freeform fallback)", got)
	}
}

// An unclosed final parameter tag (no </parameter>) takes the remainder of the
// body as its value.
func TestParseActionNativeParamTagsUnclosedValue(t *testing.T) {
	got := ParseAction("<function=read_file><parameter=path>status/web-01.txt</function>")
	if got.Kind != ActionRead || got.Arg != "status/web-01.txt" {
		t.Fatalf("got %+v, want read_file status/web-01.txt", got)
	}
}

// Bug 1430: Qwen3 writes an edit_file where the FILE PATH is the first
// parameter's NAME, not a path value, and the body rides a <parameter=new_content>
// block: <function=edit_file><parameter=REVIEW.md>…</parameter>
// <parameter=new_content>…</parameter></function>. The path synonyms find nothing
// (the path is a key, not a value) and "new_content" (singular) was absent from
// the content synonyms, so edit_file ran with empty arguments and wrote a 0-byte
// file — nulling a reached decision. The path must be recovered from the first
// path-like parameter key, and the body from the new_content parameter.
func TestParseActionNativeEditPathAsParamName(t *testing.T) {
	got := ParseAction("<function=edit_file>\n<parameter=REVIEW.md>\nplaceholder\n</parameter>\n<parameter=new_content>\nCHANGES-NEEDED: the proxy rule is not an acceptable criterion\n</parameter>\n</function>")
	if got.Kind != ActionEdit || got.Arg != "REVIEW.md" {
		t.Fatalf("got %+v, want edit of REVIEW.md", got)
	}
	if got.Content != "CHANGES-NEEDED: the proxy rule is not an acceptable criterion" {
		t.Fatalf("content = %q, want the new_content body", got.Content)
	}
}

// The singular new_content key resolves as a content synonym when the path is
// named correctly too, in the parameter-tag form and the Hermes JSON form.
func TestParseActionNativeEditNewContentSynonym(t *testing.T) {
	t.Run("param-tag", func(t *testing.T) {
		got := ParseAction("<function=edit_file>\n<parameter=path>\nREVIEW.md\n</parameter>\n<parameter=new_content>\nAPPROVED\n</parameter>\n</function>")
		if got.Kind != ActionEdit || got.Arg != "REVIEW.md" || got.Content != "APPROVED" {
			t.Fatalf("got %+v, want edit REVIEW.md -> APPROVED", got)
		}
	})
	t.Run("hermes-json", func(t *testing.T) {
		got := ParseAction(`<tool_call>{"name":"edit_file","arguments":{"path":"REVIEW.md","new_content":"APPROVED"}}</tool_call>`)
		if got.Kind != ActionEdit || got.Arg != "REVIEW.md" || got.Content != "APPROVED" {
			t.Fatalf("got %+v, want edit REVIEW.md -> APPROVED", got)
		}
	})
}

// Qwen3 drops the ">" after a parameter key and writes the path alone on its
// line: <parameter=thumbsvc/cache.py\n</parameter>. The key used to run on into
// the closing tag, so read_file ran with an empty path, the model repeated the
// call, and the cell stalled. The key must end at the newline, and read_file and
// list_files must take a path-like key as the path, as edit_file already does.
func TestParseActionNativeParamKeyMissingCloseBracket(t *testing.T) {
	t.Run("read_file", func(t *testing.T) {
		got := ParseAction("<tool_call>\n<function=read_file>\n<parameter=thumbsvc/cache.py\n</parameter>\n</function>\n</tool_call>")
		if got.Kind != ActionRead || got.Arg != "thumbsvc/cache.py" {
			t.Fatalf("got %+v, want read_file thumbsvc/cache.py", got)
		}
	})
	t.Run("list_files", func(t *testing.T) {
		got := ParseAction("<function=list_files>\n<parameter=notes/\n</parameter>\n</function>")
		if got.Kind != ActionList || got.Arg != "notes/" {
			t.Fatalf("got %+v, want list_files notes/", got)
		}
	})
	t.Run("later param still parsed", func(t *testing.T) {
		got := ParseAction("<function=edit_file>\n<parameter=docs/x.md\n<parameter=content>\nhello\n</parameter>\n</function>")
		if got.Kind != ActionEdit || got.Arg != "docs/x.md" || got.Content != "hello" {
			t.Fatalf("got %+v, want edit docs/x.md -> hello", got)
		}
	})
	t.Run("named value wins over path-like key", func(t *testing.T) {
		got := ParseAction("<function=read_file>\n<parameter=a/b.md\n<parameter=path>\nc/d.md\n</parameter>\n</function>")
		if got.Kind != ActionRead || got.Arg != "c/d.md" {
			t.Fatalf("got %+v, want read_file c/d.md", got)
		}
	})
}

// Qwen3 writes the attribute form <parameter name="KEY"> instead of
// <parameter=KEY>. The parser saw no "<parameter=" and ran the call with the raw
// tag text as its argument, so the cell stalled on a repeated no-op.
func TestParseActionNativeParamAttrForm(t *testing.T) {
	t.Run("path as key", func(t *testing.T) {
		got := ParseAction("<tool_call>\n<function=read_file>\n<parameter name=\"entries/personal/klm-2026-amphibian-survey.yaml\">\n</parameter>\n</function>\n</tool_call>")
		if got.Kind != ActionRead || got.Arg != "entries/personal/klm-2026-amphibian-survey.yaml" {
			t.Fatalf("got %+v, want read_file of the yaml path", got)
		}
	})
	t.Run("named path value", func(t *testing.T) {
		got := ParseAction("<function=read_file><parameter name='path'>docs/README.md</parameter></function>")
		if got.Kind != ActionRead || got.Arg != "docs/README.md" {
			t.Fatalf("got %+v, want read_file docs/README.md", got)
		}
	})
	t.Run("edit with content", func(t *testing.T) {
		got := ParseAction("<function=edit_file>\n<parameter name=\"path\">\nNOTES.md\n</parameter>\n<parameter name=\"content\">\nok\n</parameter>\n</function>")
		if got.Kind != ActionEdit || got.Arg != "NOTES.md" || got.Content != "ok" {
			t.Fatalf("got %+v, want edit NOTES.md -> ok", got)
		}
	})
}

// A key holding whitespace or "<" is a malformed tag, not a path, so the
// path-like fallback must not adopt it.
func TestPathLikeParamKeyRejectsMalformedKey(t *testing.T) {
	if got := pathLikeParamKey("<parameter=path status/web-01.txt</function>"); got != "" {
		t.Fatalf("pathLikeParamKey = %q, want empty", got)
	}
	if got := pathLikeParamKey("<parameter=content>x</parameter>"); got != "" {
		t.Fatalf("pathLikeParamKey = %q, want empty", got)
	}
}

// A CALL directive written before the first native marker wins. Qwen3 wrote a
// complete "CALL edit_file … ||| …" and a FINAL, then a native read of the file
// it had just written. The harness ran only the native read, so the edit never
// landed and a reached decision scored NULL (2026-10-09 smoke).
func TestParseActionsCallBeforeNativeWins(t *testing.T) {
	turn := "CALL edit_file entries/personal/klm.yaml ||| id: klm\nclass: personal\n\nFINAL Created the entry.\n\n<tool_call>\n<function=read_file>\n<parameter name=\"entries/personal/klm.yaml\">\n</parameter>\n</function>\n</tool_call>"
	acts := ParseActions(turn)
	if len(acts) != 1 || acts[0].Kind != ActionEdit || acts[0].Arg != "entries/personal/klm.yaml" {
		t.Fatalf("got %+v, want one edit of entries/personal/klm.yaml", acts)
	}
	if acts[0].Content != "id: klm\nclass: personal" {
		t.Fatalf("content = %q, want the yaml body without FINAL or native text", acts[0].Content)
	}
}

// A FINAL before a native marker ends the turn as FINAL.
func TestParseActionsFinalBeforeNativeWins(t *testing.T) {
	acts := ParseActions("FINAL done\n<tool_call>\n<function=read_file>\n<parameter=path>\nx.md\n</parameter>\n</function>\n</tool_call>")
	if len(acts) != 1 || acts[0].Kind != ActionFinal || acts[0].Summary != "done" {
		t.Fatalf("got %+v, want a single FINAL", acts)
	}
}

// Prose before a native call is not a directive, so the native calls still run.
func TestParseActionsProseBeforeNativeKeepsNative(t *testing.T) {
	acts := ParseActions("Let me check the README.\n<tool_call>\n<function=read_file>\n<parameter=path>\nREADME.md\n</parameter>\n</function>\n</tool_call>")
	if len(acts) != 1 || acts[0].Kind != ActionRead || acts[0].Arg != "README.md" {
		t.Fatalf("got %+v, want the native read of README.md", acts)
	}
}

func TestFirstNativeMarker(t *testing.T) {
	cases := map[string]int{"none": -1, "ab<function=x>": 2, "<tool_call><function=x>": 0, "a<function=x><tool_call>": 1}
	for in, want := range cases {
		if got := firstNativeMarker(in); got != want {
			t.Fatalf("firstNativeMarker(%q) = %d, want %d", in, got, want)
		}
	}
}

// A native FINAL among real calls: the calls run and the FINAL is dropped, since
// the subject has not yet seen what they return. A FINAL first ends the turn.
func TestParseActionsNativeFinalAmongCalls(t *testing.T) {
	final := "<tool_call>\n<function=FINAL>\n</function>\n</tool_call>"
	read := "<tool_call>\n<function=read_file>\n<parameter=path>\na.md\n</parameter>\n</function>\n</tool_call>"
	if got := summarize(ParseActions(read + "\n" + final)); len(got) != 1 || got[0] != "read_file a.md" {
		t.Errorf("read then FINAL = %q, want only the read", got)
	}
	if got := summarize(ParseActions(final + "\n" + read)); len(got) != 1 || got[0] != "FINAL" {
		t.Errorf("FINAL then read = %q, want only FINAL", got)
	}
}
