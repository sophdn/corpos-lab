package wrongpath

import (
	"strings"
	"testing"
)

func TestTOMLBasic(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", `"plain"`},
		{`a"b`, `"a\"b"`},
		{`a\b`, `"a\\b"`},
		{"a\nb", `"a\nb"`},
		{"a\tb", `"a\tb"`},
		{"a\\\"\n\t", `"a\\\"\n\t"`},
	}
	for _, c := range cases {
		if got := TOMLBasic(c.in); got != c.want {
			t.Errorf("TOMLBasic(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSamplingBlock(t *testing.T) {
	b := SamplingBlock(4096)
	if !strings.Contains(b, "max_tokens = 4096\n") {
		t.Errorf("max_tokens not rendered: %q", b)
	}
	if !strings.Contains(b, "—") {
		t.Error("em-dash missing from sampling block")
	}
	if !strings.HasPrefix(b, "[sampling]\n") {
		t.Error("sampling block should start with [sampling]")
	}
}

func TestRenderKnownAndUnknown(t *testing.T) {
	got, err := Render("base", "mistral", "cas", "a", "INSTRUCTION")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(got, `name = "wrongpath-base-mistral-cas-a"`) {
		t.Error("name line missing")
	}
	if !strings.Contains(got, `ground   = "../../../materials/base/GLYPH_cas-terrain.md"`) {
		t.Error("base ground path wrong")
	}
	if !strings.Contains(got, "[INST] {prompt}") {
		t.Error("prompt template not embedded")
	}
	// ablated arm points ground at the ablated terrain dir.
	abl, err := Render("ablated", "qwen38", "gop", "b", "X")
	if err != nil {
		t.Fatalf("Render ablated error: %v", err)
	}
	if !strings.Contains(abl, `ground   = "../../../materials/ablated/GLYPH_gop-terrain.md"`) {
		t.Error("ablated ground path wrong")
	}
	if _, err := Render("base", "nope", "cas", "a", "X"); err == nil {
		t.Error("expected error for unknown model")
	}
}

func TestGenStudyDefsMatrix(t *testing.T) {
	defs, err := GenStudyDefs("INSTR")
	if err != nil {
		t.Fatalf("GenStudyDefs error: %v", err)
	}
	want := len(Arms) * len(Subjects) * len(Glyphs) * len(Scenarios)
	if len(defs) != want {
		t.Fatalf("got %d defs, want %d", len(defs), want)
	}
	// first def follows generation order: base/mistral/cas-a.
	if defs[0].RelPath != "base/mistral/cas-a.toml" {
		t.Errorf("first RelPath = %q, want base/mistral/cas-a.toml", defs[0].RelPath)
	}
	seen := map[string]bool{}
	for _, d := range defs {
		if seen[d.RelPath] {
			t.Errorf("duplicate RelPath %q", d.RelPath)
		}
		seen[d.RelPath] = true
		if d.Content == "" {
			t.Errorf("empty content for %q", d.RelPath)
		}
	}
}

func TestSubjectByName(t *testing.T) {
	if _, ok := subjectByName("qwen2532"); !ok {
		t.Error("qwen2532 should resolve")
	}
	if _, ok := subjectByName("ghost"); ok {
		t.Error("ghost should not resolve")
	}
}

const sampleTerrain = `# GLYPH cas

## Y-Decision
> decide something

---

### Marker axis
> marker stuff
> more marker

### Aim axis
> aim stuff

---

### Rest axis
> rest stuff
`

func TestAblate(t *testing.T) {
	out, err := Ablate(sampleTerrain, "cas")
	if err != nil {
		t.Fatalf("Ablate error: %v", err)
	}
	if strings.Contains(out, markerHeader) || strings.Contains(out, aimHeader) {
		t.Error("Marker/Aim axis survived ablation")
	}
	if !strings.Contains(out, restHeader) {
		t.Error("Rest axis was dropped")
	}
	if !strings.Contains(out, "## Y-Decision") {
		t.Error("Y-Decision block should survive")
	}
	// The kept-header + kept-rest join is byte-exact: it is the prefix up to Marker
	// concatenated with the Rest block onward.
	lines := splitLinesKeepends(sampleTerrain)
	mi, _ := findUnique(lines, markerHeader, "cas")
	ri, _ := findUnique(lines, restHeader, "cas")
	want := strings.Join(lines[:mi], "") + strings.Join(lines[ri:], "")
	if out != want {
		t.Errorf("ablate output mismatch:\n got %q\nwant %q", out, want)
	}
}

func TestAblateErrors(t *testing.T) {
	// No marker.
	if _, err := Ablate("### Rest axis\nx\n", "g"); err == nil {
		t.Error("expected error when Marker axis missing")
	}
	// No rest.
	if _, err := Ablate("### Marker axis\nx\n### Aim axis\n", "g"); err == nil {
		t.Error("expected error when Rest axis missing")
	}
	// Marker after Rest -> not marker<rest.
	if _, err := Ablate("### Rest axis\n### Marker axis\n", "g"); err == nil {
		t.Error("expected error when Marker does not precede Rest")
	}
	// Duplicate marker.
	if _, err := Ablate("### Marker axis\n### Marker axis\n### Rest axis\n", "g"); err == nil {
		t.Error("expected error on duplicate Marker axis")
	}
}

func TestSplitLinesKeepends(t *testing.T) {
	got := splitLinesKeepends("a\nb\r\nc\rd")
	want := []string{"a\n", "b\r\n", "c\r", "d"}
	if len(got) != len(want) {
		t.Fatalf("got %d lines %q, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
	if strings.Join(got, "") != "a\nb\r\nc\rd" {
		t.Error("splitLinesKeepends does not round-trip")
	}
	if len(splitLinesKeepends("")) != 0 {
		t.Error("empty string should give no lines")
	}
}

func TestLineCount(t *testing.T) {
	if n := LineCount("a\nb\nc"); n != 3 {
		t.Errorf("LineCount = %d, want 3", n)
	}
}

func TestPyStrip(t *testing.T) {
	if pyStrip("  x \t\n") != "x" {
		t.Error("pyStrip should trim surrounding whitespace")
	}
}

func TestParseVerdict(t *testing.T) {
	cases := map[string]string{
		"VERDICT: yes":       "yes",
		"verdict:   NO here": "no",
		"VERDICT:Yes":        "yes",
		"no verdict here":    "",
	}
	for in, want := range cases {
		if got := ParseVerdict(in); got != want {
			t.Errorf("ParseVerdict(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassifyField(t *testing.T) {
	cases := map[string]string{
		"FIELD SOURCE: Scope of the operation": "scope",
		"FIELD SOURCE: Pull character":         "pull",
		"field source: the Marker axis":        "marker",
		"FIELD SOURCE: something unrelated":    "other",
		"no field here":                        "none",
		// earliest family wins: "aim" appears before "scope".
		"FIELD SOURCE: aim then scope": "aim",
	}
	for in, want := range cases {
		if got := ClassifyField(in); got != want {
			t.Errorf("ClassifyField(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestScoreCell(t *testing.T) {
	texts := []string{
		"VERDICT: yes\nFIELD SOURCE: Scope", // correct + scope
		"VERDICT: yes\nFIELD SOURCE: Pull",  // correct + shortcut
		"VERDICT: no\nFIELD SOURCE: Scope",  // wrong verdict
		"nothing",                           // no verdict
	}
	c := ScoreCell(texts, "yes")
	if c.N != 4 || c.Correct != 2 || c.CorrectScope != 1 || c.CorrectShortcut != 1 {
		t.Errorf("ScoreCell = %+v", c)
	}
}

func TestOrderedCounterAndFormatters(t *testing.T) {
	oc := newOrderedCounter()
	if oc.repr() != "{}" {
		t.Errorf("empty repr = %q", oc.repr())
	}
	oc.add("scope")
	oc.add("pull")
	oc.add("scope")
	if oc.repr() != "{'scope': 2, 'pull': 1}" {
		t.Errorf("repr = %q", oc.repr())
	}
	if bk := oc.ByKey(); bk["scope"] != 2 || bk["pull"] != 1 {
		t.Errorf("ByKey = %v", bk)
	}
	if pf2(0.5) != "0.50" {
		t.Errorf("pf2(0.5) = %q", pf2(0.5))
	}
	if pf2w(1.0, 6) != "  1.00" {
		t.Errorf("pf2w = %q", pf2w(1.0, 6))
	}
	if pf2p(0.25) != "+0.25" || pf2p(-0.25) != "-0.25" {
		t.Errorf("pf2p wrong")
	}
	// NaN branches.
	nan := meanOrNaN(nil)
	if pf2(nan) != "nan" {
		t.Errorf("pf2(nan) = %q", pf2(nan))
	}
	if pf2w(nan, 20) != strings.Repeat(" ", 17)+"nan" {
		t.Errorf("pf2w(nan,20) = %q", pf2w(nan, 20))
	}
	if pf2p(nan) != "+nan" {
		t.Errorf("pf2p(nan) = %q", pf2p(nan))
	}
}

func TestMeanOrNaN(t *testing.T) {
	if got := meanOrNaN([]float64{1, 2, 3}); got != 2 {
		t.Errorf("mean = %v, want 2", got)
	}
}

func TestReportSectionsAndNaN(t *testing.T) {
	cs := map[CellKey]CellScore{
		// mistral base: two cells; one shortcut-routed, one scope-routed.
		{"base", "mistral", "cas", "a"}:    {N: 8, Correct: 6, CorrectScope: 1, CorrectShortcut: 5}, // shortcut
		{"ablated", "mistral", "cas", "a"}: {N: 8, Correct: 3, CorrectScope: 1, CorrectShortcut: 2},
		{"base", "mistral", "cgu", "b"}:    {N: 8, Correct: 6, CorrectScope: 5, CorrectShortcut: 1}, // scope
		{"ablated", "mistral", "cgu", "b"}: {N: 8, Correct: 5, CorrectScope: 4, CorrectShortcut: 1},
		// qwen2532 base cell with <2 correct -> uninform routing, no drop bucket.
		{"base", "qwen2532", "fsb", "a"}:    {N: 8, Correct: 1, CorrectScope: 0, CorrectShortcut: 1},
		{"ablated", "qwen2532", "fsb", "a"}: {N: 8, Correct: 0, CorrectScope: 0, CorrectShortcut: 0},
		// qwen38 base cell with 0 correct -> H1 scope_rate stays finite (corr>0 check).
	}
	runs := []RunFile{
		{Arm: "base", Model: "mistral", Glyph: "cas", AB: "a", Text: "VERDICT: yes\nFIELD SOURCE: Scope"},
		{Arm: "base", Model: "mistral", Glyph: "cas", AB: "a", Text: "VERDICT: yes\nFIELD SOURCE: Pull"},
		{Arm: "base", Model: "qwen38", Glyph: "cas", AB: "a", Text: "VERDICT: yes\nFIELD SOURCE: Pull",
			Reasoning: "the scope check was applied", HasReasoning: true},
		{Arm: "base", Model: "qwen38", Glyph: "cas", AB: "a", Text: "VERDICT: yes\nFIELD SOURCE: Marker",
			Reasoning: "no mention", HasReasoning: true},
	}
	rep := Report(cs, runs)
	for _, want := range []string{
		"scored 6 cells",
		"== H1: verdict accuracy vs scope engagement (BASE arm) ==",
		"differential",
		"== Field-source class distribution per model/arm (all runs) ==",
		"non-scope-cited base runs: 2; trace mentions scope/operative: 1",
	} {
		if !strings.Contains(rep, want) {
			t.Errorf("report missing %q\n%s", want, rep)
		}
	}
	// qwen38 has no scored cells in H1 -> nan formatting present as right-justified nan.
	if !strings.Contains(rep, "nan") {
		t.Error("expected a nan cell in the report")
	}
}
