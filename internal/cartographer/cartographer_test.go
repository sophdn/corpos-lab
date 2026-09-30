package cartographer

import (
	"math"
	"reflect"
	"regexp"
	"testing"
)

// --- fixture -----------------------------------------------------------------
//
// The fixture mirrors the JSON study dir the python oracles were run on to
// capture the golden constants in golden_test.go. Grid entries are in the same
// first-appearance condition order the python dict preserved (cartographer,
// cartographer_scan, annotated, baseline), which analyze_duty reports by.

func fixtureTaboos() []Taboo {
	return []Taboo{
		{Slug: "fp-alpha", Class: "first_principles", ScanTarget: false},
		{Slug: "fp-beta", Class: "first_principles", ScanTarget: false},
		{Slug: "nf-gamma", Class: "meta_routing", ScanTarget: true},
		{Slug: "nf-delta", Class: "corpus_empirical", ScanTarget: false},
	}
}

func cov(a, b, g, d int) map[string]int {
	return map[string]int{"fp-alpha": a, "fp-beta": b, "nf-gamma": g, "nf-delta": d}
}

func fixtureGrid() []GridEntry {
	data := []struct {
		cond string
		rows []map[string]int
	}{
		{"cartographer_instrument", []map[string]int{cov(1, 1, 0, 0), cov(1, 0, 0, 0), cov(0, 1, 1, 0)}},
		{"cartographer_scan_instrument", []map[string]int{cov(1, 1, 1, 0), cov(0, 1, 1, 1), cov(1, 0, 1, 0)}},
		{"annotated_instrument", []map[string]int{cov(1, 1, 1, 1), cov(1, 1, 1, 1), cov(1, 1, 0, 1)}},
		{"baseline", []map[string]int{cov(0, 0, 0, 0), cov(1, 0, 0, 1), cov(0, 1, 0, 0)}},
	}
	var out []GridEntry
	for _, d := range data {
		for i, c := range d.rows {
			run := i + 1
			out = append(out, GridEntry{
				Key:       gridKey(d.cond, run),
				Condition: d.cond,
				Run:       run,
				Coverage:  c,
			})
		}
	}
	return out
}

func fixtureResponses() map[string][]string {
	return map[string][]string{
		"baseline":                     {"nothing here", "route the call", "discover a constraint"},
		"annotated_instrument":         {"routing matters", "advisory bypass durability", "constraint discovered"},
		"cartographer_instrument":      {"plain text", "ROUTE it", "no keywords at all"},
		"cartographer_scan_instrument": {"routing", "durability advisory", "plain"},
	}
}

// --- report golden tests -----------------------------------------------------

func TestAnalyzeGolden(t *testing.T) {
	got := Analyze(fixtureTaboos(), fixtureGrid())
	if got != wantAnalyzeFixture {
		t.Errorf("Analyze mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, wantAnalyzeFixture)
	}
}

func TestAnalyzeDutyGolden(t *testing.T) {
	got := AnalyzeDuty(fixtureTaboos(), fixtureGrid(), fixtureResponses(), DefaultPermReps, DefaultPermSeed)
	if got != wantAnalyzeDutyFixture {
		t.Errorf("AnalyzeDuty mismatch.\n--- got ---\n%s\n--- want ---\n%s", got, wantAnalyzeDutyFixture)
	}
}

// TestAnalyzeNaNAndEmptyCondition covers the two branches the golden fixture does
// not: a condition wholly absent from the grid (Wilson n==0) and a scan-recovery
// cell with no judgments (the float("nan") fallback formatted as "nan").
func TestAnalyzeNaNAndEmptyCondition(t *testing.T) {
	taboos := []Taboo{
		{Slug: "fp-a", Class: "first_principles"},
		{Slug: "nf-g", Class: "meta_routing", ScanTarget: true},
	}
	// No cartographer_scan_instrument rows at all: its rates are 0/0 and the scan
	// recovery p-value falls back to nan.
	grid := []GridEntry{
		{Key: "baseline_1", Condition: "baseline", Run: 1, Coverage: map[string]int{"fp-a": 0, "nf-g": 0}},
		{Key: "annotated_instrument_1", Condition: "annotated_instrument", Run: 1, Coverage: map[string]int{"fp-a": 1, "nf-g": 1}},
		{Key: "cartographer_instrument_1", Condition: "cartographer_instrument", Run: 1, Coverage: map[string]int{"fp-a": 1, "nf-g": 0}},
	}
	got := Analyze(taboos, grid)
	if !regexp.MustCompile(`nf-g .* -> scan 0/0  p=nan`).MatchString(got) {
		t.Errorf("expected a nan scan-recovery line, got:\n%s", got)
	}
	// cartographer_scan_instrument row: 0/0 [0.00,0.00] on both classes (Wilson n==0).
	if !regexp.MustCompile(`cartographer_scan_instrument\s+0/0 \[0.00,0.00\]\s+0/0 \[0.00,0.00\]`).MatchString(got) {
		t.Errorf("expected an empty-condition 0/0 row, got:\n%s", got)
	}
}

// --- numeric helpers ---------------------------------------------------------

func floatEq(a, b float64) bool { return math.Abs(a-b) < 1e-12 }

func TestWilson(t *testing.T) {
	cases := []struct {
		k, n           int
		wantLo, wantHi float64
	}{
		{0, 0, 0.0, 0.0},
		{1, 2, 0.0945312057016527, 0.9054687942983473},
		{3, 4, 0.30064184250845644, 0.9544127392056013},
	}
	for _, c := range cases {
		lo, hi := Wilson(c.k, c.n)
		if !floatEq(lo, c.wantLo) || !floatEq(hi, c.wantHi) {
			t.Errorf("Wilson(%d,%d) = (%v,%v), want (%v,%v)", c.k, c.n, lo, hi, c.wantLo, c.wantHi)
		}
	}
}

func TestFisherTwoSided(t *testing.T) {
	cases := []struct {
		a, b, c, d int
		want       float64
	}{
		{3, 1, 1, 3, 0.4857142857142857},
		{10, 0, 0, 10, 1.082508822446903e-05},
		{5, 5, 5, 5, 1.0},
	}
	for _, tc := range cases {
		got := FisherTwoSided(tc.a, tc.b, tc.c, tc.d)
		if !floatEq(got, tc.want) {
			t.Errorf("FisherTwoSided(%d,%d,%d,%d) = %v, want %v", tc.a, tc.b, tc.c, tc.d, got, tc.want)
		}
	}
}

func TestPermTestDeterministicAndBounds(t *testing.T) {
	xs := []float64{1, 1, 1, 1}
	ys := []float64{0, 0, 0, 0}
	// Identical seeds give identical results.
	p1 := PermTest(xs, ys, 500, 42)
	p2 := PermTest(xs, ys, 500, 42)
	if p1 != p2 {
		t.Errorf("PermTest not deterministic for a fixed seed: %v vs %v", p1, p2)
	}
	// Fully separated groups: the observed difference is maximal, so almost no
	// permutation meets it; p is near the (hits+1)/(reps+1) floor.
	if p1 <= 0 || p1 > 0.1 {
		t.Errorf("PermTest on separated groups = %v, want a small positive p", p1)
	}
	// Identical groups: every permutation ties the (zero) observed difference, so p==1.
	same := PermTest([]float64{2, 2}, []float64{2, 2}, 100, 7)
	if !floatEq(same, 1.0) {
		t.Errorf("PermTest on identical groups = %v, want 1.0", same)
	}
}

func TestMeanAndRatio(t *testing.T) {
	if m := mean([]float64{1, 2, 3}); !floatEq(m, 2) {
		t.Errorf("mean = %v, want 2", m)
	}
	if r := ratio(3, 4); !floatEq(r, 0.75) {
		t.Errorf("ratio(3,4) = %v, want 0.75", r)
	}
	if r := ratio(0, 0); r != 0 {
		t.Errorf("ratio(0,0) = %v, want 0", r)
	}
}

func TestFormatG(t *testing.T) {
	cases := []struct {
		x    float64
		prec int
		want string
	}{
		{2e-51, 3, "2e-51"},
		{1.0, 3, "1"},
		{0.0801, 3, "0.0801"},
		{0.44, 2, "0.44"},
		{1e-05, 2, "1e-05"},
		{math.NaN(), 3, "nan"},
		{math.Inf(1), 3, "inf"},
		{math.Inf(-1), 3, "-inf"},
	}
	for _, c := range cases {
		if got := formatG(c.x, c.prec); got != c.want {
			t.Errorf("formatG(%v,%d) = %q, want %q", c.x, c.prec, got, c.want)
		}
	}
}

func TestCondShort(t *testing.T) {
	cases := map[string]string{
		"baseline":                     "baseline",
		"annotated_instrument":         "annotated",
		"cartographer_instrument":      "cartograp",
		"cartographer_scan_instrument": "cartograp",
		"":                             "",
	}
	for in, want := range cases {
		if got := condShort(in); got != want {
			t.Errorf("condShort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLastSeg(t *testing.T) {
	if got := lastSeg("investigation-fix-scope-boundary"); got != "boundary" {
		t.Errorf("lastSeg = %q, want boundary", got)
	}
	if got := lastSeg("noseparator"); got != "noseparator" {
		t.Errorf("lastSeg no-hyphen = %q, want noseparator", got)
	}
}

func TestSlugAndClassHelpers(t *testing.T) {
	taboos := fixtureTaboos()
	if got := fpSlugs(taboos); !reflect.DeepEqual(got, []string{"fp-alpha", "fp-beta"}) {
		t.Errorf("fpSlugs = %v", got)
	}
	if got := nonFPSlugs(taboos); !reflect.DeepEqual(got, []string{"nf-gamma", "nf-delta"}) {
		t.Errorf("nonFPSlugs = %v", got)
	}
}

func TestRateSkipsMissingSlug(t *testing.T) {
	grid := []GridEntry{
		{Condition: "baseline", Coverage: map[string]int{"a": 1}},             // "b" missing -> not judged
		{Condition: "baseline", Coverage: map[string]int{"a": 0, "b": 1}},     // both judged
		{Condition: "annotated_instrument", Coverage: map[string]int{"a": 1}}, // wrong condition
	}
	k, n := rate(grid, "baseline", []string{"a", "b"})
	if k != 2 || n != 3 { // a:1,a:0,b:1 -> k=2 over n=3 present judgments
		t.Errorf("rate = %d/%d, want 2/3", k, n)
	}
}

func TestCovRate(t *testing.T) {
	e := GridEntry{Coverage: map[string]int{"a": 1, "b": 0, "c": 1}}
	if got := covRate(e, []string{"a", "b", "c"}); !floatEq(got, 2.0/3.0) {
		t.Errorf("covRate = %v, want 2/3", got)
	}
}

func TestKeywordCounting(t *testing.T) {
	re := regexp.MustCompile(`(?i)\brout(e|ing)\b`)
	texts := []string{"route it", "ROUTING now", "reroute", "nothing"}
	// "route it" matches, "ROUTING now" matches (case-insensitive), "reroute" does
	// not (no leading word boundary before rout), "nothing" no.
	if got := countMatches(re, texts); got != 2 {
		t.Errorf("countMatches = %d, want 2", got)
	}
}

// --- slug-c2 -----------------------------------------------------------------

func TestCitedSlugsEmptyIsNonNil(t *testing.T) {
	got := CitedSlugs("no slugs cited here")
	if got == nil {
		t.Fatal("CitedSlugs returned nil; must be non-nil so JSON serialises as []")
	}
	if len(got) != 0 {
		t.Errorf("CitedSlugs = %v, want empty", got)
	}
}

func TestCitedSlugsHits(t *testing.T) {
	text := "we honor known-constraint-documentation and triggers-routing per canon"
	got := CitedSlugs(text)
	// Result is in C2Slugs order: triggers-routing precedes known-constraint-documentation.
	want := []string{"triggers-routing", "known-constraint-documentation"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CitedSlugs = %v, want %v", got, want)
	}
}

func TestSlugC2Summary(t *testing.T) {
	rows := []SlugC2Row{
		{Condition: "annotated_instrument", Run: 1, NSlugs: 10},
		{Condition: "annotated_instrument", Run: 2, NSlugs: 8},
		{Condition: "baseline", Run: 1, NSlugs: 0},
		{Condition: "baseline", Run: 2, NSlugs: 0},
	}
	summary, table := SlugC2Summary(rows)
	if summary["annotated_instrument"] != (SlugC2CondSummary{Runs: 2, RunsCitingGe1: 2, MeanSlugs: 9.0}) {
		t.Errorf("annotated summary = %+v", summary["annotated_instrument"])
	}
	if summary["baseline"] != (SlugC2CondSummary{Runs: 2, RunsCitingGe1: 0, MeanSlugs: 0.0}) {
		t.Errorf("baseline summary = %+v", summary["baseline"])
	}
	// Table is sorted by condition (annotated before baseline) and has header + 2 rows.
	wantTable := "condition                        runs  runs citing >=1  mean slugs\n" +
		"annotated_instrument                2                2        9.00\n" +
		"baseline                            2                0        0.00\n"
	if table != wantTable {
		t.Errorf("table mismatch.\n--- got ---\n%s\n--- want ---\n%s", table, wantTable)
	}
}

func TestSlugC2SummaryEmpty(t *testing.T) {
	summary, table := SlugC2Summary(nil)
	if len(summary) != 0 {
		t.Errorf("empty summary = %v", summary)
	}
	// Only the header line.
	if table != "condition                        runs  runs citing >=1  mean slugs\n" {
		t.Errorf("empty table = %q", table)
	}
}

// --- blind set ---------------------------------------------------------------

func TestBuildBlindSetRoundTrip(t *testing.T) {
	duties := []Duty{
		{Condition: "baseline", Run: 1, Text: "t-b1"},
		{Condition: "annotated_instrument", Run: 2, Text: "t-a2"},
		{Condition: "cartographer_instrument", Run: 3, Text: "t-c3"},
		{Condition: "cartographer_scan_instrument", Run: 4, Text: "t-s4"},
	}
	res := BuildBlindSet(duties, DefaultPermSeed)
	if len(res) != len(duties) {
		t.Fatalf("got %d results, want %d", len(res), len(duties))
	}
	// IDs are contiguous d0001.. and every (cond,run,text) is preserved exactly once.
	seen := map[string]string{} // "cond_run" -> text
	for i, r := range res {
		wantID := "d000" + string(rune('1'+i))
		if r.ID != wantID {
			t.Errorf("result %d id = %q, want %q", i, r.ID, wantID)
		}
		seen[gridKey(r.Condition, r.Run)] = r.Text
	}
	for _, d := range duties {
		if got := seen[gridKey(d.Condition, d.Run)]; got != d.Text {
			t.Errorf("duty %s_%d text = %q, want %q (blinding must preserve text)", d.Condition, d.Run, got, d.Text)
		}
	}
	// Determinism for a fixed seed.
	res2 := BuildBlindSet(duties, DefaultPermSeed)
	if !reflect.DeepEqual(res, res2) {
		t.Error("BuildBlindSet not deterministic for a fixed seed")
	}
}

func TestBuildBlindSetDoesNotMutateInput(t *testing.T) {
	duties := []Duty{
		{Condition: "baseline", Run: 1, Text: "a"},
		{Condition: "baseline", Run: 2, Text: "b"},
		{Condition: "baseline", Run: 3, Text: "c"},
	}
	before := append([]Duty(nil), duties...)
	_ = BuildBlindSet(duties, 99)
	if !reflect.DeepEqual(duties, before) {
		t.Errorf("input mutated: %v", duties)
	}
}

// --- unblind -----------------------------------------------------------------

func TestUnblind(t *testing.T) {
	mapping := []BlindMapEntry{
		{ID: "d0001", Condition: "baseline", Run: 1},
		{ID: "d0002", Condition: "annotated_instrument", Run: 2},
		{ID: "d0003", Condition: "cartographer_instrument", Run: 3}, // no judge call -> missing
	}
	judge := map[string]map[string]int{
		"d0001": {"triggers-routing": 1},                                // one set, rest default 0
		"d0002": {"triggers-routing": 2, "fix-locus-identification": 1}, // non-zero normalises to 1
	}
	grid, missing := Unblind(mapping, judge)
	if len(grid) != 2 {
		t.Fatalf("grid has %d rows, want 2", len(grid))
	}
	row := grid["baseline_1"]
	if row.Condition != "baseline" || row.Run != 1 {
		t.Errorf("baseline_1 meta wrong: %+v", row)
	}
	if len(row.Coverage) != len(C2Slugs) {
		t.Errorf("coverage has %d slugs, want %d", len(row.Coverage), len(C2Slugs))
	}
	if row.Coverage["triggers-routing"] != 1 {
		t.Errorf("triggers-routing = %d, want 1", row.Coverage["triggers-routing"])
	}
	if row.Coverage["known-constraint-documentation"] != 0 {
		t.Errorf("unset slug = %d, want 0", row.Coverage["known-constraint-documentation"])
	}
	if grid["annotated_instrument_2"].Coverage["triggers-routing"] != 1 {
		t.Error("non-zero judge call should normalise to 1")
	}
	if !reflect.DeepEqual(missing, []string{"d0003"}) {
		t.Errorf("missing = %v, want [d0003]", missing)
	}
}

func TestCombOutOfRange(t *testing.T) {
	// The Fisher sum never asks for an out-of-range coefficient, but the guard
	// must return 0 for k<0 or k>n rather than panic.
	if comb(3, -1).Sign() != 0 {
		t.Error("comb(3,-1) should be 0")
	}
	if comb(3, 5).Sign() != 0 {
		t.Error("comb(3,5) should be 0")
	}
	if comb(5, 2).Int64() != 10 {
		t.Errorf("comb(5,2) = %v, want 10", comb(5, 2))
	}
}

// TestAnalyzeDutySpreadRange exercises the per-taboo spread min/max updates, which
// the equal-count golden fixture leaves untouched: here the two first-principles
// slugs have different cartographer coverage, so lo and hi diverge.
func TestAnalyzeDutySpreadRange(t *testing.T) {
	taboos := []Taboo{
		{Slug: "fp-mid", Class: "first_principles"},
		{Slug: "fp-lo", Class: "first_principles"},
		{Slug: "fp-hi", Class: "first_principles"},
		{Slug: "nf-x", Class: "meta_routing"},
	}
	// counts across the three fp slugs, in slug order, are mid=1, lo=0, hi=2, so
	// the first sets lo=hi=1, the second drops lo, the third raises hi.
	c := func(mid, lo, hi, x int) map[string]int {
		return map[string]int{"fp-mid": mid, "fp-lo": lo, "fp-hi": hi, "nf-x": x}
	}
	grid := []GridEntry{
		{Key: "cartographer_instrument_1", Condition: "cartographer_instrument", Run: 1, Coverage: c(1, 0, 1, 0)},
		{Key: "cartographer_instrument_2", Condition: "cartographer_instrument", Run: 2, Coverage: c(0, 0, 1, 0)},
		{Key: "annotated_instrument_1", Condition: "annotated_instrument", Run: 1, Coverage: c(1, 1, 1, 1)},
		{Key: "baseline_1", Condition: "baseline", Run: 1, Coverage: c(0, 0, 0, 0)},
	}
	out := AnalyzeDuty(taboos, grid, map[string][]string{}, 200, DefaultPermSeed)
	// cartographer covers fp-mid 1/30, fp-lo 0/30, fp-hi 2/30 -> range 0/30 to 2/30.
	if !regexp.MustCompile(`mid 1/30, lo 0/30, hi 2/30`).MatchString(out) {
		t.Errorf("expected per-taboo spread counts, got:\n%s", out)
	}
	if !regexp.MustCompile(`range 0/30 to 2/30`).MatchString(out) {
		t.Errorf("expected range 0/30 to 2/30, got:\n%s", out)
	}
}

func TestGridKey(t *testing.T) {
	if got := gridKey("baseline", 12); got != "baseline_12" {
		t.Errorf("gridKey = %q, want baseline_12", got)
	}
}
