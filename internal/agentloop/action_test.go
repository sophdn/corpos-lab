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
