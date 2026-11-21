package agentloop

import "testing"

func TestClassifyTurn(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		want    ParseOutcome
		lost    int
		ignored int
	}{
		{"call form", "CALL read_file a.md", OutcomeParsed, 0, 0},
		{"final", "FINAL done", OutcomeFinal, 0, 0},
		{"call final", "CALL FINAL done", OutcomeFinal, 0, 0},
		{"stall", "I think the change is fine as it is.", OutcomeNone, 0, 0},
		{"empty native wrapper", "<tool_call>\n<function=call>\n</function>\n</tool_call>", OutcomeUnknownTool, 1, 0},
		{"unparseable native", "<tool_call>{not json}</tool_call>", OutcomeMalformed, 1, 0},
		{"unparseable call line", "CALL", OutcomeNone, 0, 0},
		{"edit with no content", "CALL edit_file notes.md", OutcomeEmptyArg, 1, 0},
		{"read with no path", "<function=read_file></function>", OutcomeEmptyArg, 1, 0},
		{"list root is fine", "CALL list_files", OutcomeParsed, 0, 0},
		{"tag soup arg", "<function=read_file>\n<parameter\ndocs/a.md\n</function>", OutcomeMalformed, 1, 0},
		{"echoed call lines", "CALL read_file a.md\nCALL read_file b.md\nCALL list_files .", OutcomeParsed, 0, 2},
		{"call before native", "CALL read_file a.md\n<tool_call>\n<function=read_file>\n<parameter=path>\nb.md\n</parameter>\n</function>\n</tool_call>", OutcomeParsed, 0, 1},
		{"native read plus unknown", "<tool_call>\n<function=read_file>\n<parameter=path>\na.md\n</parameter>\n</function>\n</tool_call>\n<tool_call>\n<function=computer_use>\n</function>\n</tool_call>", OutcomeUnknownTool, 1, 0},
		{"native final after read", "<tool_call>\n<function=read_file>\n<parameter=path>\na.md\n</parameter>\n</function>\n</tool_call>\n<tool_call>\n<function=FINAL>\n</function>\n</tool_call>", OutcomeParsed, 0, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ClassifyTurn(c.text, ParseActions(c.text))
			if got.Outcome != c.want || got.Lost != c.lost || got.Ignored != c.ignored {
				t.Errorf("ClassifyTurn = %+v, want outcome %s lost %d ignored %d", got, c.want, c.lost, c.ignored)
			}
		})
	}
}

func TestClaimedEdits(t *testing.T) {
	text := "CALL edit_file a.md ||| one\nCALL edit_file b.md ||| two\n<tool_call>\n<function=edit_file>\n<parameter=path>\nc.md\n</parameter>\n<parameter=content>\nx\n</parameter>\n</function>\n</tool_call>\n<tool_call>\n<function=read_file>\n<parameter=path>\nd.md\n</parameter>\n</function>\n</tool_call>"
	got := ClaimedEdits(text)
	if len(got) != 3 || got[0] != "a.md" || got[1] != "b.md" || got[2] != "c.md" {
		t.Errorf("ClaimedEdits = %q, want [a.md b.md c.md]", got)
	}
	if got := ClaimedEdits("CALL read_file a.md"); len(got) != 0 {
		t.Errorf("no edits claimed, got %q", got)
	}
}
