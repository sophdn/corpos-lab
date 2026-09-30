package pubshared

import (
	"math/rand"
	"strings"
	"testing"
)

// isLorem reports whether w is one of the vocab-swap filler tokens.
func isLorem(w string) bool {
	for _, l := range loremWords {
		if w == l {
			return true
		}
	}
	return false
}

// identityShuffle leaves order untouched, so shuffle-mode output is deterministic
// in tests.
func identityShuffle(n int, swap func(i, j int)) {}

// reverseShuffle reverses the slice, a deterministic non-identity reorder.
func reverseShuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func TestTransformShuffleStructure(t *testing.T) {
	src := "# my-glyph\n\n## Y-Decision\n> alpha beta gamma\n\n---\n\n- one two three\n**Label:** four five\n"
	out := Transform(src, "shuffle", false, reverseShuffle)
	lines := strings.Split(out, "\n")
	// Title kept (no neutralize), blank/rule kept.
	if lines[0] != "# my-glyph" {
		t.Errorf("title changed: %q", lines[0])
	}
	// Blockquote prefix preserved, words reversed.
	if !strings.HasPrefix(lines[3], "> ") || !strings.Contains(lines[3], "gamma beta alpha") {
		t.Errorf("blockquote line = %q", lines[3])
	}
	// Bullet prefix preserved.
	if !strings.HasPrefix(lines[7], "- ") {
		t.Errorf("bullet line = %q", lines[7])
	}
	// Bold label prefix preserved.
	if !strings.HasPrefix(lines[8], "**Label:** ") {
		t.Errorf("label line = %q", lines[8])
	}
}

func TestTransformNeutralizeTitle(t *testing.T) {
	out := Transform("# parent-state-check-bypass\n## keep me\nbody words here\n", "shuffle", true, identityShuffle)
	lines := strings.Split(out, "\n")
	if lines[0] != "# glyph" {
		t.Errorf("title not neutralized: %q", lines[0])
	}
	if lines[1] != "## keep me" {
		t.Errorf("## heading should be kept verbatim: %q", lines[1])
	}
}

func TestTransformVocabSwap(t *testing.T) {
	out := Transform("> alpha beta\n- gamma\n", "vocab-swap", false, identityShuffle)
	// The lorem cursor advances across the document: 3 words -> first 3 lorem tokens.
	want := "> " + loremWords[0] + " " + loremWords[1] + "\n- " + loremWords[2] + "\n"
	if out != want {
		t.Errorf("vocab-swap:\n got %q\nwant %q", out, want)
	}
}

func TestTransformSingleWordNotShuffled(t *testing.T) {
	// A single-word body must pass through unchanged (len(words) <= 1).
	out := Transform("solo\n", "shuffle", false, reverseShuffle)
	if out != "solo\n" {
		t.Errorf("single word body = %q, want solo\\n", out)
	}
}

func TestTransformGradedZeroKeepsAllVocab(t *testing.T) {
	// strength 0 is the weak end: every word is kept (reordered), none vocab-swapped.
	rng := rand.New(rand.NewSource(1)) //nolint:gosec // reproducible test scramble
	out := TransformGraded("> alpha beta gamma delta\n", 0, false, rng)
	words := strings.Fields(strings.TrimPrefix(out, "> "))
	if len(words) != 4 {
		t.Fatalf("word count changed: %q", out)
	}
	got := map[string]bool{}
	for _, w := range words {
		if isLorem(w) {
			t.Errorf("strength 0 produced a lorem token: %q", out)
		}
		got[w] = true
	}
	for _, w := range []string{"alpha", "beta", "gamma", "delta"} {
		if !got[w] {
			t.Errorf("strength 0 dropped %q: %q", w, out)
		}
	}
	if !strings.HasPrefix(out, "> ") {
		t.Errorf("prefix lost: %q", out)
	}
}

func TestTransformGradedOneIsVocabSwap(t *testing.T) {
	// strength 1 is the strong end: identical to vocab-swap (lorem in order, cursor
	// advances across the document).
	rng := rand.New(rand.NewSource(99)) //nolint:gosec // reproducible test scramble
	out := TransformGraded("> alpha beta\n- gamma\n", 1, false, rng)
	want := "> " + loremWords[0] + " " + loremWords[1] + "\n- " + loremWords[2] + "\n"
	if out != want {
		t.Errorf("strength 1:\n got %q\nwant %q", out, want)
	}
}

func TestTransformGradedMixedAndDeterministic(t *testing.T) {
	src := "> alpha beta gamma delta epsilon zeta eta theta\n"
	out1 := TransformGraded(src, 0.5, false, rand.New(rand.NewSource(7))) //nolint:gosec // reproducible
	out2 := TransformGraded(src, 0.5, false, rand.New(rand.NewSource(7))) //nolint:gosec // reproducible
	if out1 != out2 {
		t.Fatalf("not deterministic for a fixed seed:\n%q\n%q", out1, out2)
	}
	words := strings.Fields(strings.TrimPrefix(out1, "> "))
	if len(words) != 8 {
		t.Fatalf("word count changed: %q", out1)
	}
	var lorem, orig int
	for _, w := range words {
		if isLorem(w) {
			lorem++
		} else {
			orig++
		}
	}
	if lorem == 0 || orig == 0 {
		t.Errorf("strength 0.5 should mix lorem and original, got lorem=%d orig=%d: %q", lorem, orig, out1)
	}
}

func TestTransformGradedClampAndPassthrough(t *testing.T) {
	// Negative clamps to the weak end (no lorem); >1 clamps to the strong end (all lorem).
	weak := TransformGraded("> alpha beta\n", -1, false, rand.New(rand.NewSource(3))) //nolint:gosec // reproducible
	for _, w := range strings.Fields(strings.TrimPrefix(weak, "> ")) {
		if isLorem(w) {
			t.Errorf("negative strength produced lorem: %q", weak)
		}
	}
	strong := TransformGraded("> alpha beta\n", 2, false, rand.New(rand.NewSource(3))) //nolint:gosec // reproducible
	for _, w := range strings.Fields(strings.TrimPrefix(strong, "> ")) {
		if !isLorem(w) {
			t.Errorf("strength>1 kept an original word: %q", strong)
		}
	}
	// Title neutralized, blank + rule preserved; at strength 0 a lone body word is kept.
	out := TransformGraded("# parent-state-check-bypass\n\n---\nsolo\n", 0, true, rand.New(rand.NewSource(5))) //nolint:gosec // reproducible
	lines := strings.Split(out, "\n")
	if lines[0] != "# glyph" || lines[1] != "" || lines[2] != "---" || lines[3] != "solo" {
		t.Errorf("passthrough/neutralize broken: %q", out)
	}
}

func TestPyStripLStrip(t *testing.T) {
	if pyStrip(" \tx\n ") != "x" {
		t.Error("pyStrip")
	}
	if pyLStrip("  x ") != "x " {
		t.Error("pyLStrip")
	}
}

func TestParseOrderedAndDumps(t *testing.T) {
	v, err := ParseOrdered([]byte(`{"b":1,"a":"two","c":[true,null,3.5],"d":{}}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// Key order preserved (b, a, c, d), separators ", " and ": ".
	got := DumpsPy(v, false)
	want := `{"b": 1, "a": "two", "c": [true, null, 3.5], "d": {}}`
	if got != want {
		t.Errorf("DumpsPy:\n got %q\nwant %q", got, want)
	}
}

func TestParseOrderedErrors(t *testing.T) {
	if _, err := ParseOrdered([]byte(`{`)); err == nil {
		t.Error("expected error on truncated object")
	}
	if _, err := ParseOrdered([]byte(`{} garbage`)); err == nil {
		t.Error("expected error on trailing data")
	}
	if _, err := ParseOrdered([]byte(`[1,`)); err == nil {
		t.Error("expected error on truncated array")
	}
}

func TestPyJSONString(t *testing.T) {
	// ensure_ascii=False passes non-ASCII through; C0 controls escaped.
	if got := pyJSONString("é\t\x01", false); got != "\"é\\t\\u0001\"" {
		t.Errorf("ensure_ascii=false: %q", got)
	}
	// ensure_ascii=True escapes non-ASCII; astral rune -> surrogate pair.
	if got := pyJSONString("é", true); got != `"\u00e9"` {
		t.Errorf("ensure_ascii=true bmp: %q", got)
	}
	if got := pyJSONString("\U0001F600", true); got != `"\ud83d\ude00"` {
		t.Errorf("ensure_ascii=true astral: %q", got)
	}
	// All short escapes.
	if got := pyJSONString("\\\"\b\f\n\r", false); got != `"\\\"\b\f\n\r"` {
		t.Errorf("short escapes: %q", got)
	}
}

func TestEscapeScriptClose(t *testing.T) {
	if escapeScriptClose("a</script>b") != `a<\/script>b` {
		t.Error("escapeScriptClose")
	}
}

func TestDumpsPyBoolNumberFallback(t *testing.T) {
	// Go-native values (not from ParseOrdered) fall through to json.Marshal.
	if DumpsPy(map[string]int{"x": 1}, false) != `{"x":1}` {
		t.Errorf("fallback marshal wrong: %q", DumpsPy(map[string]int{"x": 1}, false))
	}
	if DumpsPy(true, false) != "true" || DumpsPy(false, false) != "false" {
		t.Error("bool dump")
	}
	if DumpsPy(nil, false) != "null" {
		t.Error("nil dump")
	}
}

func mustParse(t *testing.T, s string) any {
	t.Helper()
	v, err := ParseOrdered([]byte(s))
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}

func TestValidateItems(t *testing.T) {
	good := mustParse(t, `[{"id":"C00","blocks":[{"heading":"h","text":"t","kind":"context"}]}]`)
	if err := ValidateItems(good); err != nil {
		t.Errorf("good items rejected: %v", err)
	}
	cases := []struct{ json, substr string }{
		{`[]`, "non-empty JSON array"},
		{`{}`, "non-empty JSON array"},
		{`["x"]`, "each item must be an object"},
		{`[{"id":"a","arm":"t","blocks":[{"heading":"h"}]}]`, "forbidden field(s) ['arm']"},
		{`[{"blocks":[{"heading":"h"}]}]`, "missing a non-empty string id"},
		{`[{"id":"a","blocks":[{"heading":"h"}]},{"id":"a","blocks":[{"heading":"h"}]}]`, "duplicate item id 'a'"},
		{`[{"id":"a","blocks":[]}]`, "non-empty blocks array"},
		{`[{"id":"a","blocks":["x"]}]`, "each block must be an object"},
		{`[{"id":"a","blocks":[{"heading":"h","judge":"x"}]}]`, "block carries forbidden field(s) ['judge']"},
		{`[{"id":"a","blocks":[{"heading":"h","kind":"secret"}]}]`, "block kind 'secret' not one of"},
	}
	for _, c := range cases {
		err := ValidateItems(mustParse(t, c.json))
		if err == nil || !strings.Contains(err.Error(), c.substr) {
			t.Errorf("items %s: got %v, want substr %q", c.json, err, c.substr)
		}
	}
}

func TestValidateLabels(t *testing.T) {
	good := mustParse(t, `[{"code":"C","desc":"d","key":"1","criteria":["one"]}]`)
	if err := ValidateLabels(good); err != nil {
		t.Errorf("good labels rejected: %v", err)
	}
	cases := []struct{ json, substr string }{
		{`[]`, "non-empty JSON array"},
		{`["x"]`, "at least code and desc"},
		{`[{"code":"C"}]`, "at least code and desc"},
		{`[{"code":"C","desc":"d","machine":"x"}]`, "forbidden field(s) ['machine']"},
		{`[{"code":"C","desc":"d","criteria":[]}]`, "criteria must be a non-empty array"},
		{`[{"code":"C","desc":"d","criteria":["ok",""]}]`, "every criterion must be a non-empty string"},
	}
	for _, c := range cases {
		err := ValidateLabels(mustParse(t, c.json))
		if err == nil || !strings.Contains(err.Error(), c.substr) {
			t.Errorf("labels %s: got %v, want substr %q", c.json, err, c.substr)
		}
	}
}

func TestBuildApp(t *testing.T) {
	tpl := "T=__TITLE__ I=__ITEMS_JSON__ L=__LABELS_JSON__ R=__RUBRIC_JSON__ S=__STORE_KEY__"
	items := mustParse(t, `[{"id":"C00","blocks":[{"heading":"h","text":"</x>","kind":"context"}]}]`)
	labels := mustParse(t, `[{"code":"C","desc":"d","key":"1"},{"code":"N","desc":"n","key":"2"}]`)
	rub := "rubric </r>"
	html, err := BuildApp(tpl, items, labels, "My Title", "store_v1", &rub)
	if err != nil {
		t.Fatalf("BuildApp: %v", err)
	}
	for _, want := range []string{
		"T=My Title",
		`R="rubric <\/r>"`,
		`S="store_v1"`,
		`<\/x>`, // script-close rewrite applied inside items
	} {
		if !strings.Contains(html, want) {
			t.Errorf("html missing %q:\n%s", want, html)
		}
	}
	// nil rubric -> "null".
	html2, err := BuildApp(tpl, items, labels, "T", "k", nil)
	if err != nil {
		t.Fatalf("BuildApp nil rubric: %v", err)
	}
	if !strings.Contains(html2, "R=null") {
		t.Errorf("nil rubric not null: %s", html2)
	}
	// Validation failure propagates.
	bad := mustParse(t, `[{"id":"a","arm":"t","blocks":[{"heading":"h"}]}]`)
	if _, err := BuildApp(tpl, bad, labels, "T", "k", nil); err == nil {
		t.Error("expected BuildApp to reject a leaky item")
	}
	if _, err := BuildApp(tpl, items, mustParse(t, `[]`), "T", "k", nil); err == nil {
		t.Error("expected BuildApp to reject empty labels")
	}
}

func TestMissingCriteria(t *testing.T) {
	rub := "r"
	three := mustParse(t, `[{"code":"C","desc":"d"},{"code":"Ic","desc":"d"},{"code":"N","desc":"d"}]`)
	if MissingCriteria(three, nil) != true {
		t.Error("3 codes, no criteria, no rubric -> should be missing")
	}
	if MissingCriteria(three, &rub) != false {
		t.Error("rubric present -> not missing")
	}
	two := mustParse(t, `[{"code":"C","desc":"d"},{"code":"N","desc":"d"}]`)
	if MissingCriteria(two, nil) != false {
		t.Error("<=2 codes -> not flagged")
	}
	withCrit := mustParse(t, `[{"code":"C","desc":"d","criteria":["x"]},{"code":"Ic","desc":"d"},{"code":"N","desc":"d"}]`)
	if MissingCriteria(withCrit, nil) != false {
		t.Error("some criteria present -> not missing")
	}
	if MissingCriteria("notarray", nil) != false {
		t.Error("non-array -> not missing")
	}
}

func TestAgreement(t *testing.T) {
	ids := []string{"a", "b", "c", "d"}
	pred := map[string]string{"a": "C", "b": "N", "c": "Ic", "d": ""}
	gold := map[string]string{"a": "C", "b": "I", "c": "Ic", "d": "C"}
	e, c, n := Agreement(pred, gold, ids)
	// common = a,b,c (d has empty pred). exact: a,c = 2. cnc: a(C==C),b(not C both)->true,c->true =3.
	if e != 2 || c != 3 || n != 3 {
		t.Errorf("Agreement = %d,%d,%d want 2,3,3", e, c, n)
	}
}

func TestBuildHumanClaude(t *testing.T) {
	rids := []string{"r0", "r1", "r2"}
	human := BuildHuman(rids, map[string]string{"H01": "C", "H02": "N"}) // H03 missing -> ""
	if human["r0"] != "C" || human["r1"] != "N" || human["r2"] != "" {
		t.Errorf("BuildHuman = %v", human)
	}
	claude := BuildClaude(rids, map[string]string{"r0": "C", "r1": "Ic", "r2": "N"},
		map[string]string{"r0": "C", "r1": "C", "r2": "N"})
	if claude["r0"] != "C" || claude["r1"] != "SPLIT" || claude["r2"] != "N" {
		t.Errorf("BuildClaude = %v", claude)
	}
}

func TestAnchorEvalReport(t *testing.T) {
	rids := []string{"r0", "r1", "r2"}
	human := map[string]string{"r0": "C", "r1": "N", "r2": "Ic"}
	claude := map[string]string{"r0": "C", "r1": "N", "r2": "SPLIT"}
	rep := AnchorEvalReport("anc", rids, human, claude, nil, "")
	for _, want := range []string{
		"anchor anc: 3 items  |  human-N=1  claude-splits=1",
		"SELF-CHECK: claude-consensus vs human",
		"off-task N recall",
		"C-boundary:",
	} {
		if !strings.Contains(rep, want) {
			t.Errorf("report missing %q:\n%s", want, rep)
		}
	}
	// With a rater block.
	rater := map[string]string{"r0": "C", "r1": "I", "r2": "Ic"}
	rep2 := AnchorEvalReport("anc", rids, human, claude, rater, "granite")
	if !strings.Contains(rep2, "=== granite ===") {
		t.Errorf("rater block missing:\n%s", rep2)
	}
}

func TestRatioAndPf3(t *testing.T) {
	if ratio(0, 0) != 0 {
		t.Error("ratio n==0 should be 0")
	}
	if ratio(1, 2) != 0.5 {
		t.Error("ratio")
	}
	if pf3(0.6667) != "0.667" {
		t.Errorf("pf3 = %q", pf3(0.6667))
	}
}
