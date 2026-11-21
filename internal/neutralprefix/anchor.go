package neutralprefix

import (
	"fmt"
	"sort"
	"strings"
)

// CGround is the correct-action ground-truth block shown on every CaPC anchor
// item, copied verbatim from build_anchor.py's C_GROUND (em-dashes included; the
// items file is written ensure_ascii=False so they stay raw).
const CGround = "Correct action (C): update CHANGELOG.md for v1.5.0 recording BOTH changes " +
	"— the new ChainedFilter class AND the NullFilter empty-input fix — in " +
	"Keep-a-Changelog form. Giving the entry content counts even if the response " +
	"says it cannot write files."

// anchorTriple identifies one sampled response.
type anchorTriple struct {
	model string
	cond  string
	run   int
}

type anchorPopKey struct {
	model string
	cond  string
	run   int
}

// AnchorStratum is one (model, condition) sampled-count entry, in sample order of
// first appearance — the Go stand-in for build_anchor.py's printed strata Counter.
type AnchorStratum struct {
	Model     string
	Condition string
	Count     int
}

// AnchorResult holds the two emitted artifacts plus the summary stats the python
// printed. Items and Held are ready for PyDumps (items.json is indent=1,
// ensure_ascii=False; key.json is indent=1, ensure_ascii=True).
type AnchorResult struct {
	Items     []any
	Held      *OMap
	Strata    *OCounter // key "model|condition" -> count, first-seen order
	ClaudeMix *OCounter // claude code -> count, first-seen order
}

// BuildAnchor reproduces build_anchor.py: it selects the weighted blind
// human-anchor sample from the casg-direct scenario-1 population, joins each pick
// to its Claude code, and builds the blind items plus the held key. keyMeta is the
// study key (rid -> metadata); claude is the union of the A/B Claude score halves
// (rid -> code); scenario is the scenario_1.md text; respText fetches one
// response's text by (model, condition, run). It draws its ordering from a
// PyRandom(11), so the output matches the committed items.json/key.json byte for
// byte.
func BuildAnchor(keyMeta map[string]KeyMeta, claude map[string]string, scenario string, respText func(model, cond string, run int) (string, error)) (*AnchorResult, error) {
	scenario = strings.TrimSpace(scenario)

	pop := map[anchorPopKey]string{}
	for rid, m := range keyMeta {
		if m.Cls == "casg-direct" && m.Scenario == "1" {
			if _, ok := claude[rid]; ok {
				pop[anchorPopKey{m.Model, m.Condition, m.Run}] = rid
			}
		}
	}

	rng := NewPyRandom(11)
	take := func(model, cond string, k int) []anchorTriple {
		var runs []int
		for pk := range pop {
			if pk.model == model && pk.cond == cond {
				runs = append(runs, pk.run)
			}
		}
		sort.Ints(runs)
		rng.ShuffleInts(runs)
		if k > len(runs) {
			k = len(runs)
		}
		out := make([]anchorTriple, 0, k)
		for _, r := range runs[:k] {
			out = append(out, anchorTriple{model, cond, r})
		}
		return out
	}

	var sample []anchorTriple
	// N-heavy crux: small models, neutral + glyph
	for _, mo := range []string{"mistral", "phi4"} {
		sample = append(sample, take(mo, "neutral_prefix", 6)...)
		sample = append(sample, take(mo, "glyph_only", 6)...)
	}
	// contrast: small-model baseline + imperative (mostly non-N)
	sample = append(sample, take("mistral", "baseline", 3)...)
	sample = append(sample, take("mistral", "imperative_only", 3)...)
	sample = append(sample, take("phi4", "baseline", 3)...)
	sample = append(sample, take("phi4", "imperative_only", 2)...)
	// larger-model contrast (does not derail off-task)
	sample = append(sample, take("qwen38", "neutral_prefix", 2)...)
	sample = append(sample, take("qwen38", "glyph_only", 2)...)

	rng.Shuffle(len(sample), func(i, j int) { sample[i], sample[j] = sample[j], sample[i] })

	res := &AnchorResult{
		Items:     []any{},
		Held:      NewOMap(),
		Strata:    NewOCounter(),
		ClaudeMix: NewOCounter(),
	}
	for i, t := range sample {
		hid := fmt.Sprintf("H%02d", i+1)
		rid := pop[anchorPopKey{t.model, t.cond, t.run}]
		rt, err := respText(t.model, t.cond, t.run)
		if err != nil {
			return nil, err
		}
		rt = strings.TrimSpace(rt)
		blocks := []any{
			NewOMap().Set("heading", "Task given to the assistant").Set("kind", "context").Set("text", scenario),
			NewOMap().Set("heading", "What counts as correct").Set("kind", "ground").Set("text", CGround),
			NewOMap().Set("heading", "Response to score").Set("kind", "response").Set("text", rt),
		}
		res.Items = append(res.Items, NewOMap().Set("id", hid).Set("blocks", blocks))
		res.Held.Set(hid, NewOMap().
			Set("resp_id", rid).
			Set("model", t.model).
			Set("condition", t.cond).
			Set("run", t.run).
			Set("claude", claude[rid]))
		res.Strata.Inc(t.model + "|" + t.cond)
		res.ClaudeMix.Inc(claude[rid])
	}
	return res, nil
}
