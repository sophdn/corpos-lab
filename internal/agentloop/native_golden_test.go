package agentloop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseNativeGolden freezes parseNative's result for every native shape it
// reads, the malformed hybrids it recovers, and the inputs it must decline.
func TestParseNativeGolden(t *testing.T) {
	inputs := []string{
		`<tool_call>{"name":"read_file","arguments":{"path":"a.go"}}</tool_call>`,
		`<tool_call>{"name":"list_files","arguments":{"dir":"src"}}`,
		`<tool_call>{not json}</tool_call>`,
		"<function=read_file>a.go</function>",
		"<tool_call><function=list_files>tests/</function></tool_call>",
		"<function=edit_file\n<parameter=path>a.go</parameter>\n<parameter=content>x\ny</parameter>\n</function>",
		"<function=FINAL>all done</function>",
		"<function=final></tool_call>",
		"<function=CALL list_files tests/>",
		"<function=CALL read_file>b.go</function>",
		"<function=mystery>arg</function>",
		"<function=read_file no close",
		"<function=read_file>unterminated arg",
		"<tool_call>\n\n  read_file c.go\n</tool_call>",
		"<tool_call>\nedit_file d.go ||| line1\nline2\n</tool_call>",
		"<tool_call>\nwhatever text\n</tool_call>",
		"<tool_call>\n\n</tool_call>",
		"plain prose, no call",
	}
	var b strings.Builder
	for _, in := range inputs {
		a, ok := parseNative(in)
		fmt.Fprintf(&b, "%q\n  -> ok=%v %+v\n", in, ok, a)
	}
	out := b.String()
	golden := filepath.Join("testdata", "parse_native.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("parseNative drifted from %s:\n%s", golden, out)
	}
}
