package restaxis

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseCompletions(t *testing.T) {
	txt := `# casg-direct — completions

Some intro text.

## Scenario 1 — append one entry

- Correct: return the list.

## Scenario 2 — another

second body

### scenario_5 mixed case and underscore

fifth body`
	got := ParseCompletions(txt)
	want := map[int]string{
		1: "- Correct: return the list.",
		2: "second body",
		5: "fifth body",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseCompletions = %#v, want %#v", got, want)
	}
}

func TestParseCompletionsNoHeadings(t *testing.T) {
	got := ParseCompletions("just a paragraph, no scenario headings")
	if len(got) != 0 {
		t.Errorf("expected empty map, got %#v", got)
	}
}

func TestParseCompletionsDuplicateKeepsLast(t *testing.T) {
	txt := "## Scenario 1 — first\nold\n## Scenario 1 — again\nnew"
	got := ParseCompletions(txt)
	if got[1] != "new" {
		t.Errorf("duplicate scenario should keep last: got %q", got[1])
	}
}

func TestParseCompletionsCaseInsensitiveAndUnderscore(t *testing.T) {
	txt := "# SCENARIO_3 heading\nbody three"
	got := ParseCompletions(txt)
	if got[3] != "body three" {
		t.Errorf("case-insensitive/underscore parse failed: %#v", got)
	}
}

func TestRenderCalibItemsJSONEmpty(t *testing.T) {
	if got := RenderCalibItemsJSON(nil); got != "[]" {
		t.Errorf("empty = %q, want []", got)
	}
}

func TestRenderCalibItemsJSONExactFormat(t *testing.T) {
	items := []CalibItem{
		{ID: "C00", Task: "t", GT: "g", Resp: "r"},
		{ID: "C01", Task: "u", GT: "h", Resp: "s"},
	}
	got := RenderCalibItemsJSON(items)
	want := `[{"id": "C00", "task": "t", "gt": "g", "resp": "r"}, {"id": "C01", "task": "u", "gt": "h", "resp": "s"}]`
	if got != want {
		t.Errorf("RenderCalibItemsJSON:\n got  %q\n want %q", got, want)
	}
}

func TestRenderCalibItemsJSONScriptEscape(t *testing.T) {
	got := RenderCalibItemsJSON([]CalibItem{{ID: "x", Task: "</script><div>", GT: "", Resp: ""}})
	if strings.Contains(got, "</") {
		t.Errorf("unescaped </ present: %q", got)
	}
	if !strings.Contains(got, `<\/script>`) {
		t.Errorf("expected <\\/script>, got %q", got)
	}
}

func TestPyJSONStringEscapes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", `"plain"`},
		{"a\"b", `"a\"b"`},
		{"a\\b", `"a\\b"`},
		{"a\nb", `"a\nb"`},
		{"a\tb", `"a\tb"`},
		{"a\rb", `"a\rb"`},
		{"a\bb", `"a\bb"`},
		{"a\fb", `"a\fb"`},
		{"a\x01b", `"a\u0001b"`},
		{"café — €", `"café — €"`}, // non-ASCII passes through (ensure_ascii=False)
		{"a/b", `"a/b"`},           // forward slash not escaped
	}
	for _, c := range cases {
		if got := pyJSONString(c.in); got != c.want {
			t.Errorf("pyJSONString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildCalibSPA(t *testing.T) {
	tpl := "const ITEMS = __ITEMS_JSON__; const STORE = __ANSWERS_KEY__; // again __ANSWERS_KEY__"
	items := []CalibItem{{ID: "C00", Task: "do it", GT: "gt", Resp: "resp"}}
	got := BuildCalibSPA(tpl, items, "v1")
	if strings.Contains(got, "__ITEMS_JSON__") || strings.Contains(got, "__ANSWERS_KEY__") {
		t.Errorf("placeholder not fully replaced: %q", got)
	}
	if !strings.Contains(got, `[{"id": "C00", "task": "do it", "gt": "gt", "resp": "resp"}]`) {
		t.Errorf("items not embedded: %q", got)
	}
	if strings.Count(got, `"v1"`) != 2 {
		t.Errorf("answers key not replaced at both sites: %q", got)
	}
}
