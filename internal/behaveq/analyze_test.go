package behaveq

import (
	"math"
	"strings"
	"testing"
)

// wantAnalyzeN24 is analyze.py's full stdout for suffix "-n24" over the committed
// study runs, captured from the python oracle. It locks the report format and the
// statistics (Wilson, Fisher, Newcombe, n-selection) byte-for-byte.
const wantAnalyzeN24 = `# Analysis for run suffix '-n24'

## DP-1 clear-rate (primary), with 95% Wilson interval
  mistral:
    baseline     DP-1 cleared 0/24  [0.00,0.14]   DP-2 violated 0/24
    duty_only    DP-1 cleared 3/24  [0.04,0.31]   DP-2 violated 0/24
    corpus_only  DP-1 cleared 3/24  [0.04,0.31]   DP-2 violated 2/24
  qwen38:
    baseline     DP-1 cleared 0/24  [0.00,0.14]   DP-2 violated 0/24
    duty_only    DP-1 cleared 14/24  [0.39,0.76]   DP-2 violated 0/24
    corpus_only  DP-1 cleared 15/24  [0.43,0.79]   DP-2 violated 0/24
  qwen2532:
    baseline     DP-1 cleared 0/24  [0.00,0.14]   DP-2 violated 0/24
    duty_only    DP-1 cleared 23/24  [0.80,0.99]   DP-2 violated 0/24
    corpus_only  DP-1 cleared 4/24  [0.07,0.36]   DP-2 violated 0/24

## Contrasts (per model)
  mistral:
    corpus>brief (1-sided Fisher): p=0.117
    duty>brief   (1-sided Fisher): p=0.117
    corpus vs duty (2-sided Fisher): p=1
    corpus-duty diff 90% CI: [-0.17,+0.17]  (equivalent if within [-0.15,+0.15])
  qwen38:
    corpus>brief (1-sided Fisher): p=1.196e-06
    duty>brief   (1-sided Fisher): p=4.066e-06
    corpus vs duty (2-sided Fisher): p=1
    corpus-duty diff 90% CI: [-0.18,+0.26]  (equivalent if within [-0.15,+0.15])
  qwen2532:
    corpus>brief (1-sided Fisher): p=0.05461
    duty>brief   (1-sided Fisher): p=7.753e-13
    corpus vs duty (2-sided Fisher): p=2.304e-08
    corpus-duty diff 90% CI: [-0.89,-0.59]  (equivalent if within [-0.15,+0.15])

## n-selection (from Qwen3.8 corpus vs duty)
  observed rates: corpus 0.625, duty 0.583
  n per cell for 80% power, superiority: 2162
  n per cell for a 90% CI within +/-0.15 (equivalence, rough): 58
`

// studyN24Data reconstructs the committed study's per-cell counts (the numbers the
// python oracle scored), so the report test needs no filesystem.
func studyN24Data() map[string]ModelData {
	return map[string]ModelData{
		"mistral": {Present: true, Cells: map[string]CellScore{
			"baseline":    {Clear: 0, DP2Viol: 0, N: 24},
			"duty_only":   {Clear: 3, DP2Viol: 0, N: 24},
			"corpus_only": {Clear: 3, DP2Viol: 2, N: 24},
		}},
		"qwen38": {Present: true, Cells: map[string]CellScore{
			"baseline":    {Clear: 0, DP2Viol: 0, N: 24},
			"duty_only":   {Clear: 14, DP2Viol: 0, N: 24},
			"corpus_only": {Clear: 15, DP2Viol: 0, N: 24},
		}},
		"qwen2532": {Present: true, Cells: map[string]CellScore{
			"baseline":    {Clear: 0, DP2Viol: 0, N: 24},
			"duty_only":   {Clear: 23, DP2Viol: 0, N: 24},
			"corpus_only": {Clear: 4, DP2Viol: 0, N: 24},
		}},
	}
}

func TestReportByteIdentical(t *testing.T) {
	got := Report("-n24", studyN24Data())
	if got != wantAnalyzeN24 {
		t.Errorf("Report mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, wantAnalyzeN24)
	}
}

func TestReportAbsentModel(t *testing.T) {
	data := studyN24Data()
	data["mistral"] = ModelData{Present: false}
	got := Report("-n24", data)
	if !strings.Contains(got, "  mistral-n24: (no runs)\n") {
		t.Errorf("absent model line missing:\n%s", got)
	}
	// An absent model has no contrast block.
	if strings.Contains(got, "  mistral:\n    corpus>brief") {
		t.Error("absent model should have no contrast block")
	}
}

func TestReportNoQwen38SkipsNSelection(t *testing.T) {
	data := studyN24Data()
	data["qwen38"] = ModelData{Present: false}
	got := Report("-n24", data)
	if strings.Contains(got, "## n-selection") {
		t.Error("n-selection should be omitted when qwen38 is absent")
	}
}

func TestReportEqualRatesShowsNone(t *testing.T) {
	// Equal qwen38 corpus/duty rates make n_for_superiority return None.
	data := map[string]ModelData{
		"qwen38": {Present: true, Cells: map[string]CellScore{
			"baseline":    {Clear: 0, N: 10},
			"duty_only":   {Clear: 5, N: 10},
			"corpus_only": {Clear: 5, N: 10},
		}},
	}
	got := Report("-x", data)
	if !strings.Contains(got, "superiority: None\n") {
		t.Errorf("expected None for equal rates:\n%s", got)
	}
}

func TestScoreCell(t *testing.T) {
	texts := []string{
		"Hypothesis 1: a guess. The order-service batch job ran.", // cleared, mentions batch
		"caused by the cache error.",                              // violated, no batch -> dp2viol
		"Hypothesis 1: another guess. revalidation of 1,247.",     // cleared, mentions batch
	}
	cs := ScoreCell(texts)
	if cs.N != 3 {
		t.Errorf("N = %d, want 3", cs.N)
	}
	if cs.Clear != 2 {
		t.Errorf("Clear = %d, want 2", cs.Clear)
	}
	if cs.DP2Viol != 1 {
		t.Errorf("DP2Viol = %d, want 1", cs.DP2Viol)
	}
}

func TestMentionsBatch(t *testing.T) {
	for _, hit := range []string{"the order-service failed", "a batch job", "reconciliation ran",
		"1247 requests", "1,247 requests", "revalidation queue"} {
		if !MentionsBatch(hit) {
			t.Errorf("MentionsBatch(%q) = false, want true", hit)
		}
	}
	if MentionsBatch("the cache error only") {
		t.Error("MentionsBatch on unrelated text should be false")
	}
}

func TestWilsonZeroN(t *testing.T) {
	lo, hi := wilson(0, 0, z95)
	if lo != 0 || hi != 0 {
		t.Errorf("wilson(0,0) = (%v,%v), want (0,0)", lo, hi)
	}
}

func TestWilsonKnown(t *testing.T) {
	// 0/24 at 95% -> [0.00, 0.14] when rounded to 2dp (matches the oracle).
	lo, hi := wilson(0, 24, z95)
	if math.Round(lo*100) != 0 || math.Round(hi*100) != 14 {
		t.Errorf("wilson(0,24) = [%.4f,%.4f], want ~[0.00,0.14]", lo, hi)
	}
}

func TestCombOutOfRange(t *testing.T) {
	if comb(5, -1).Sign() != 0 {
		t.Error("comb(5,-1) should be 0")
	}
	if comb(5, 6).Sign() != 0 {
		t.Error("comb(5,6) should be 0")
	}
	if comb(5, 2).Int64() != 10 {
		t.Error("comb(5,2) should be 10")
	}
}

func TestFisherTwoSidedIdentical(t *testing.T) {
	// Identical rows give p = 1.
	p := fisherTwoSided(3, 21, 3, 21)
	if math.Abs(p-1.0) > 1e-9 {
		t.Errorf("fisherTwoSided identical rows = %v, want 1", p)
	}
}

func TestFisherOneSidedGreaterKnown(t *testing.T) {
	// mistral corpus>brief: a=3,b=21,c=0,d=24 -> p≈0.117 (from the oracle).
	p := fisherOneSidedGreater(3, 21, 0, 24)
	if math.Abs(p-0.117) > 5e-4 {
		t.Errorf("fisherOneSidedGreater = %v, want ~0.117", p)
	}
}

func TestNewcombeDiffCISymmetricZero(t *testing.T) {
	lo, hi := newcombeDiffCI(3, 24, 3, 24, z90)
	if math.Abs((lo + hi)) > 1e-9 {
		t.Errorf("newcombe for equal cells should be symmetric about 0: [%v,%v]", lo, hi)
	}
}

func TestNForSuperiority(t *testing.T) {
	if got := nForSuperiority(0.5, 0.5, z95, zb80); got != nil {
		t.Errorf("equal rates should give nil, got %v", *got)
	}
	// qwen38 rates 0.625 vs 0.583 -> 2162 (from the oracle).
	got := nForSuperiority(15.0/24.0, 14.0/24.0, z95, zb80)
	if got == nil || *got != 2162 {
		t.Errorf("nForSuperiority = %v, want 2162", got)
	}
}

func TestNForEquivalence(t *testing.T) {
	// qwen38 rates -> 58 (from the oracle).
	if got := nForEquivalence(15.0/24.0, 14.0/24.0, 0.15, z90); got != 58 {
		t.Errorf("nForEquivalence = %d, want 58", got)
	}
	// Both at the boundary -> hw_var clamps to 0.25.
	got := nForEquivalence(0, 0, 0.15, z90)
	want := int(math.Ceil(z90 * z90 * 0.25 / (0.15 * 0.15)))
	if got != want {
		t.Errorf("nForEquivalence boundary = %d, want %d", got, want)
	}
}

func TestNoneOrInt(t *testing.T) {
	if noneOrInt(nil) != "None" {
		t.Error("nil should be None")
	}
	v := 42
	if noneOrInt(&v) != "42" {
		t.Error("&42 should be 42")
	}
}
