package matchedcontent

import (
	"strings"
	"testing"
)

func identityShuffle(n int, swap func(i, j int)) {}

func reverseShuffle(n int, swap func(i, j int)) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func TestBuildCodeMap(t *testing.T) {
	key := map[string]CondRun{
		"r0": {"ground_only", 1},
		"r1": {"glyph_only", 2},
		"r2": {"baseline", 3},
	}
	a := []RaterEntry{{"r0", "C"}, {"r1", "Ic"}, {"r2", "N"}}
	b := []RaterEntry{{"r0", "C"}, {"r1", "C"}, {"r2", "N"}} // r1 splits
	adj := map[string]string{"r1": "Ii"}
	m, err := BuildCodeMap(key, a, b, adj)
	if err != nil {
		t.Fatalf("BuildCodeMap: %v", err)
	}
	if m[CondRun{"ground_only", 1}] != "C" {
		t.Error("agreement code wrong")
	}
	if m[CondRun{"glyph_only", 2}] != "Ii" {
		t.Error("split should resolve via adjudication")
	}
	if m[CondRun{"baseline", 3}] != "N" {
		t.Error("agreement N wrong")
	}
	// Split without adjudication -> error.
	if _, err := BuildCodeMap(key, a, b, map[string]string{}); err == nil {
		t.Error("expected error for unadjudicated split")
	}
	// Missing rid in a rater bundle -> error.
	if _, err := BuildCodeMap(key, a[:1], b, adj); err == nil {
		t.Error("expected error for rid missing from a rater bundle")
	}
}

func TestBuildSampleStructure(t *testing.T) {
	for _, sh := range []ShuffleFunc{identityShuffle, reverseShuffle} {
		sample := BuildSample(sh)
		if len(sample) != 40 {
			t.Fatalf("sample size = %d, want 40", len(sample))
		}
		counts := map[[2]string]int{}
		runsByStratum := map[[2]string]map[int]bool{}
		for _, s := range sample {
			k := [2]string{s.Model, s.Condition}
			counts[k]++
			if runsByStratum[k] == nil {
				runsByStratum[k] = map[int]bool{}
			}
			if runsByStratum[k][s.Run] {
				t.Errorf("duplicate run %d in stratum %v", s.Run, k)
			}
			runsByStratum[k][s.Run] = true
			if s.Run < 1 || s.Run > 24 {
				t.Errorf("run %d out of range", s.Run)
			}
		}
		want := map[[2]string]int{
			{"mistral", "ground_only"}:            24,
			{"mistral", "glyph_only"}:             4,
			{"mistral", "imperative_only"}:        3,
			{"mistral", "baseline"}:               3,
			{"mistral", "domain_imperative_only"}: 2,
			{"qwen", "ground_only"}:               2,
			{"qwen", "glyph_only"}:                2,
		}
		for k, v := range want {
			if counts[k] != v {
				t.Errorf("stratum %v = %d, want %d", k, counts[k], v)
			}
		}
	}
}

func TestBuildItemsAndMap(t *testing.T) {
	sample := []Sel{
		{"mistral", "ground_only", 1},
		{"qwen", "glyph_only", 5},
	}
	codes := map[string]map[CondRun]string{
		"mistral": {{"ground_only", 1}: "C"},
		"qwen":    {{"glyph_only", 5}: "Ii"},
	}
	items, order, cmap, err := BuildItemsAndMap(sample, "TASK", "GTX", []string{"resp0", "resp1"}, codes)
	if err != nil {
		t.Fatalf("BuildItemsAndMap: %v", err)
	}
	if len(items) != 2 || order[0] != "H01" || order[1] != "H02" {
		t.Errorf("items/order wrong: %v %v", items, order)
	}
	if items[0].Task != "TASK" || items[0].GT != "GTX" || items[0].Resp != "resp0" {
		t.Errorf("item0 = %+v", items[0])
	}
	if cmap["H02"].Claude != "Ii" || cmap["H02"].Run != 5 {
		t.Errorf("cmap H02 = %+v", cmap["H02"])
	}
	// Length mismatch.
	if _, _, _, err := BuildItemsAndMap(sample, "T", "G", []string{"only-one"}, codes); err == nil {
		t.Error("expected length mismatch error")
	}
	// Missing model.
	if _, _, _, err := BuildItemsAndMap([]Sel{{"ghost", "x", 1}}, "T", "G", []string{"r"}, codes); err == nil {
		t.Error("expected missing-model error")
	}
	// Missing code.
	if _, _, _, err := BuildItemsAndMap([]Sel{{"mistral", "x", 9}}, "T", "G", []string{"r"}, codes); err == nil {
		t.Error("expected missing-code error")
	}
}

func TestTask(t *testing.T) {
	if Task("  hello \n") != "hello" {
		t.Error("Task should strip")
	}
}

func TestConstants(t *testing.T) {
	if !strings.Contains(GT, "Assign the code that fits the response.") {
		t.Error("GT constant wrong")
	}
	if !strings.HasPrefix(LabelsBlock, "const LABELS = [") || !strings.HasSuffix(LabelsBlock, "];") {
		t.Error("LabelsBlock constant wrong")
	}
}

func TestRenderItemsJSON(t *testing.T) {
	items := []Item{{ID: "H01", Task: "t </x>", GT: "g é", Resp: "r\nline"}}
	got := RenderItemsJSON(items)
	want := `[{"id": "H01", "task": "t <\/x>", "gt": "g é", "resp": "r\nline"}]`
	if got != want {
		t.Errorf("RenderItemsJSON:\n got %q\nwant %q", got, want)
	}
	if RenderItemsJSON(nil) != "[]" {
		t.Errorf("empty items = %q", RenderItemsJSON(nil))
	}
}

func TestRenderCmapJSON(t *testing.T) {
	order := []string{"H01", "H02"}
	cmap := map[string]CmapEntry{
		"H01": {Model: "mistral", Condition: "ground_only", Run: 3, Claude: "C"},
		"H02": {Model: "qwen", Condition: "glyph_only", Run: 1, Claude: "Ii"},
	}
	got := RenderCmapJSON(order, cmap)
	want := "{\n" +
		" \"H01\": {\n  \"model\": \"mistral\",\n  \"condition\": \"ground_only\",\n  \"run\": 3,\n  \"claude\": \"C\"\n },\n" +
		" \"H02\": {\n  \"model\": \"qwen\",\n  \"condition\": \"glyph_only\",\n  \"run\": 1,\n  \"claude\": \"Ii\"\n }\n" +
		"}"
	if got != want {
		t.Errorf("RenderCmapJSON:\n got %q\nwant %q", got, want)
	}
	if RenderCmapJSON(nil, nil) != "{}" {
		t.Errorf("empty cmap = %q", RenderCmapJSON(nil, nil))
	}
}

func TestFillTemplate(t *testing.T) {
	tpl := "H1 Rest-axis calibration\nconst LABELS = [\n [\"OK\",\"x\",\"1\"],\n];\nstore restaxis_calib_x\n" +
		"Ground truth &mdash; correct completion &amp; the over-fire to watch for\n" +
		"Ground truth \u2014 correct completion &amp; the over-fire to watch for\n" +
		"I=__ITEMS_JSON__ K=__ANSWERS_KEY__"
	out := FillTemplate(tpl, `[{"id":"H01"}]`)
	for _, want := range []string{
		"casg-direct scoring calibration",
		`["C","Produced a correct CHANGELOG.md`,
		"casg_human_anchor_x",
		"Scoring rubric &mdash; what each code means",
		"Scoring rubric \u2014 what each code means",
		`I=[{"id":"H01"}]`,
		`K="v1"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("FillTemplate missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Rest-axis calibration") || strings.Contains(out, "restaxis_calib_") {
		t.Error("old strings survived")
	}
	if strings.Contains(out, `["OK","x","1"]`) {
		t.Error("old LABELS block survived")
	}
}

func TestCounterReprs(t *testing.T) {
	order := []string{"H01", "H02", "H03", "H04"}
	cmap := map[string]CmapEntry{
		"H01": {Model: "mistral", Condition: "ground_only", Claude: "C"},
		"H02": {Model: "mistral", Condition: "ground_only", Claude: "N"},
		"H03": {Model: "mistral", Condition: "glyph_only", Claude: "C"},
		"H04": {Model: "qwen", Condition: "ground_only", Claude: "C"},
	}
	// most_common: ground_only(mistral)=2 first, then the two 1s in first-seen order.
	strata := StrataCounterRepr(order, cmap)
	want := "Counter({('mistral', 'ground_only'): 2, ('mistral', 'glyph_only'): 1, ('qwen', 'ground_only'): 1})"
	if strata != want {
		t.Errorf("StrataCounterRepr = %q, want %q", strata, want)
	}
	// claude: C=3, N=1.
	claude := ClaudeCounterRepr(order, cmap)
	if claude != "Counter({'C': 3, 'N': 1})" {
		t.Errorf("ClaudeCounterRepr = %q", claude)
	}
	if counterRepr(nil) != "Counter({})" {
		t.Errorf("empty counterRepr = %q", counterRepr(nil))
	}
}

func TestPyJSONString(t *testing.T) {
	// pyJSONString does not rewrite "</" (that happens in RenderItemsJSON).
	if got := pyJSONString("é</x>\t", false); got != "\"é</x>\\t\"" {
		t.Errorf("ensure_ascii=false: %q", got)
	}
	if pyJSONString("é", true) != `"\u00e9"` {
		t.Errorf("ensure_ascii=true bmp: %q", pyJSONString("é", true))
	}
	if pyJSONString("\U0001F600", true) != `"\ud83d\ude00"` {
		t.Errorf("astral: %q", pyJSONString("\U0001F600", true))
	}
	if pyJSONString("\\\"\b\f\n\r\x01", false) != `"\\\"\b\f\n\r\u0001"` {
		t.Errorf("escapes: %q", pyJSONString("\\\"\b\f\n\r\x01", false))
	}
}
