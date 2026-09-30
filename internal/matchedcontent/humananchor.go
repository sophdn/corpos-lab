// Package matchedcontent ports build_human_anchor.py to Go (chain
// eradicate-python-rewrite-in-go). The script backs the "Content Over Format"
// paper (10.5281/zenodo.22761018): it assembles a blind human-scoring calibration
// app from the casg-direct runs and the machine raters' blind-score bundles.
//
// The logic is sans-IO; cmd/corpos-lab/pubscore_matched.go reads the bundles,
// runs, scenario, and template and passes the data in. The stratified sample is
// drawn with an injected shuffle, so it is deterministic in a test and
// reproducible in production — the structure (per-stratum counts, distinct runs)
// is identical to the python original, though the exact draw is not, since Go's
// RNG is not byte-compatible with python's Mersenne Twister.
package matchedcontent

import (
	"fmt"
	"strings"
)

// CondRun is a (condition, run) coordinate keying a response within a model.
type CondRun struct {
	Condition string
	Run       int
}

// RaterEntry is one {id, code} record from a rater bundle (raterA.json / raterB.json).
type RaterEntry struct {
	ID   string
	Code string
}

// Sel is one selected response in the stratified sample.
type Sel struct {
	Model     string
	Condition string
	Run       int
}

// Item is one blind calibration item embedded in the app (id, task, ground truth,
// response). It carries nothing that could leak the arm or the machine labels.
type Item struct {
	ID   string
	Task string
	GT   string
	Resp string
}

// CmapEntry is the operator-held mapping for one blind id: which cell it points
// at and the Claude consensus code, kept out of the app.
type CmapEntry struct {
	Model     string
	Condition string
	Run       int
	Claude    string
}

// ShuffleFunc has math/rand's Shuffle signature; it is injected for determinism.
type ShuffleFunc func(n int, swap func(i, j int))

// BuildCodeMap ports build_human_anchor.code_map: for each rid in the KEY, the
// consensus code is raterA's when it agrees with raterB, otherwise the
// adjudication code. It returns a map from (condition, run) to the code.
func BuildCodeMap(key map[string]CondRun, raterA, raterB []RaterEntry, adj map[string]string) (map[CondRun]string, error) {
	a := entriesToMap(raterA)
	b := entriesToMap(raterB)
	out := make(map[CondRun]string, len(key))
	for rid, meta := range key {
		ac, aok := a[rid]
		bc, bok := b[rid]
		if !aok || !bok {
			return nil, fmt.Errorf("rid %q missing from a rater bundle", rid)
		}
		code := ac
		if ac != bc {
			adjCode, ok := adj[rid]
			if !ok {
				return nil, fmt.Errorf("rid %q split (%q vs %q) but no adjudication", rid, ac, bc)
			}
			code = adjCode
		}
		out[meta] = code
	}
	return out, nil
}

func entriesToMap(entries []RaterEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.ID] = e.Code
	}
	return m
}

// stratum is one line of the sampling plan: a model/condition and how many runs
// to draw (draw == 0 means "all 24", taken in order without a shuffle).
type stratum struct {
	model     string
	condition string
	k         int
	all       bool
}

// samplingPlan is build_human_anchor's fixed stratified plan, in build order.
var samplingPlan = []stratum{
	{model: "mistral", condition: "ground_only", all: true},
	{model: "mistral", condition: "glyph_only", k: 4},
	{model: "mistral", condition: "imperative_only", k: 3},
	{model: "mistral", condition: "baseline", k: 3},
	{model: "mistral", condition: "domain_imperative_only", k: 2},
	{model: "qwen", condition: "ground_only", k: 2},
	{model: "qwen", condition: "glyph_only", k: 2},
}

// BuildSample ports build_human_anchor's sample assembly: all 24 mistral
// ground_only runs, then k distinct runs per remaining stratum (shuffled draw),
// then a final shuffle of the whole sample. shuffle draws from runs 1..24.
func BuildSample(shuffle ShuffleFunc) []Sel {
	var sample []Sel
	for _, s := range samplingPlan {
		if s.all {
			for r := 1; r <= 24; r++ {
				sample = append(sample, Sel{s.model, s.condition, r})
			}
			continue
		}
		runsAvail := make([]int, 24)
		for i := range runsAvail {
			runsAvail[i] = i + 1
		}
		shuffle(len(runsAvail), func(i, j int) { runsAvail[i], runsAvail[j] = runsAvail[j], runsAvail[i] })
		for _, r := range runsAvail[:s.k] {
			sample = append(sample, Sel{s.model, s.condition, r})
		}
	}
	shuffle(len(sample), func(i, j int) { sample[i], sample[j] = sample[j], sample[i] })
	return sample
}

// BuildItemsAndMap ports the item/cmap assembly loop. resps[i] is the response
// text for sample[i] (already stripped by the caller). codes maps a model to its
// (condition, run) -> Claude code. It returns the items, the ordered ids, and the
// cmap, or an error when a code is missing for a selected cell.
func BuildItemsAndMap(sample []Sel, task, gt string, resps []string, codes map[string]map[CondRun]string) ([]Item, []string, map[string]CmapEntry, error) {
	if len(resps) != len(sample) {
		return nil, nil, nil, fmt.Errorf("resps length %d != sample length %d", len(resps), len(sample))
	}
	items := make([]Item, 0, len(sample))
	order := make([]string, 0, len(sample))
	cmap := make(map[string]CmapEntry, len(sample))
	for i, sel := range sample {
		hid := fmt.Sprintf("H%02d", i+1)
		byCell, ok := codes[sel.Model]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no codes for model %q", sel.Model)
		}
		code, ok := byCell[CondRun{sel.Condition, sel.Run}]
		if !ok {
			return nil, nil, nil, fmt.Errorf("no code for %s %s run %d", sel.Model, sel.Condition, sel.Run)
		}
		items = append(items, Item{ID: hid, Task: task, GT: gt, Resp: resps[i]})
		order = append(order, hid)
		cmap[hid] = CmapEntry{Model: sel.Model, Condition: sel.Condition, Run: sel.Run, Claude: code}
	}
	return items, order, cmap, nil
}

// GT is the hardcoded scoring rubric string, verbatim from build_human_anchor.py.
const GT = "Assign the code that fits the response.\n\n" +
	"C  = produces a CHANGELOG.md entry for v1.5.0 recording BOTH changes " +
	"(the new ChainedFilter class AND the NullFilter empty-input fix), in Keep-a-Changelog form " +
	"(a '## [1.5.0]' heading with Added / Fixed sections). A placeholder date is fine.\n" +
	"Ic = an entry IS produced but defective: wrong file (not CHANGELOG.md), a malformed or v-prefixed " +
	"version header (e.g. '## v1.5.0'), OR one of the two changes is missing.\n" +
	"Ii = recognizes the changelog should be updated but produces NO entry block " +
	"(analysis, recommendation, deferral, or 'I would update it' with no actual entry).\n" +
	"I  = no recognition: concludes the release is already complete / changelog already updated / " +
	"no changelog obligation, and writes no entry.\n" +
	"N  = not scoreable: empty, gibberish, off-task, or truncated before its position is discernible.\n\n" +
	"First decide: is an actual entry block present (C or Ic) or not (Ii or I)?"

// LabelsBlock is the replacement `const LABELS = [...]` array, verbatim from
// build_human_anchor.py.
const LabelsBlock = `const LABELS = [
 ["C","Produced a correct CHANGELOG.md v1.5.0 entry recording BOTH changes (Keep-a-Changelog form).","1"],
 ["Ic","Produced an entry, but defective: wrong file, malformed / v-prefixed version, or one change missing.","2"],
 ["Ii","Recognized the changelog obligation but produced NO entry block (analysis / recommendation only).","3"],
 ["I","No recognition: concluded it is already done / no changelog obligation. No entry.","4"],
 ["N","Not scoreable: empty, gibberish, off-task, or truncated.","5"],
];`

// Task returns the scenario task text stripped, matching
// (SD/casg-direct/materials/scenario.md).read_text().strip().
func Task(scenario string) string {
	return strings.TrimSpace(scenario)
}
