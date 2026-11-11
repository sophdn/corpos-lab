// Package restaxis ports the rest-axis-overhead-benchmark scoring harness from
// python to Go. The study is the Rest-axis ablation grid (chain 424): three
// matched arms — baseline, the full glyph, and the glyph with its Rest axis
// dropped — scored for over-firing on neutral scenarios.
//
// The logic is sans-IO: callers read the study's run and score files and pass
// the data in. cmd/corpos-lab/restaxis.go does that wiring.
package restaxis

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Arms are the three matched arms of the ablation grid, in report order.
var Arms = []string{"baseline", "glyph_only", "glyph_minus_rest"}

// lexiconSrc is the mechanical over-fire marker set: a regex list per category.
// A hit is an objective over-fire signal — a cross-check on the LLM judges, not
// the judge. Ported verbatim from mech_score.py; the text is lower-cased before
// matching, so the patterns are lower-case and carry no ignore-case flag.
var lexiconSrc = map[string][]string{
	"offer": {`\bwould you like\b`, `\bdo you want\b`, `\bshould i\b`, `\blet me know\b`,
		`\bwant me to\b`, `\bwould you prefer\b`, `\bif you.d like\b`},
	"hedge": {`\byou may want\b`, `\byou might want\b`, `\bi.d recommend\b`, `\bi recommend\b`,
		`\bworth noting\b`, `\bnote that\b`, `\bkeep in mind\b`, `\bbe aware\b`,
		`\bcaveat\b`, `\bone thing to\b`, `\bjust a heads up\b`},
	"sync": {`\bin sync\b`, `\bkeep .* (?:in sync|consistent|up to date)\b`, `\balso update\b`,
		`\bcorresponding\b`, `\bcompanion\b`, `\bto match\b`, `\bstay consistent\b`},
	"gate": {`\bbefore (?:i|we|you)\b`, `\bfirst,? (?:i|we|you|make|check|verify|confirm)\b`,
		`\bprerequisite\b`, `\bi can.?t\b`, `\bi cannot\b`, `\bunable to\b`,
		`\bneed to (?:verify|check|confirm|ensure)\b`, `\bmake sure to\b`,
		`\byou.ll (?:also )?need to\b`, `\bdon.t forget\b`, `\byou should also\b`},
}

// lexiconCats is the category order used when building a Hits map, kept stable so
// output is deterministic.
var lexiconCats = []string{"offer", "hedge", "sync", "gate"}

// lexicon holds the compiled patterns, built once.
var lexicon = func() map[string][]*regexp.Regexp {
	m := make(map[string][]*regexp.Regexp, len(lexiconSrc))
	for cat, pats := range lexiconSrc {
		rs := make([]*regexp.Regexp, len(pats))
		for i, p := range pats {
			rs[i] = regexp.MustCompile(p)
		}
		m[cat] = rs
	}
	return m
}()

// Hits is a per-category lexicon hit count.
type Hits map[string]int

// Markers counts lexicon hits over the lower-cased text and reports whether the
// text ends with a question mark. Ported from mech_score.markers.
func Markers(text string) (Hits, int) {
	t := strings.ToLower(text)
	h := make(Hits, len(lexiconCats))
	for _, cat := range lexiconCats {
		n := 0
		for _, re := range lexicon[cat] {
			n += len(re.FindAllString(t, -1))
		}
		h[cat] = n
	}
	trailingQ := 0
	if strings.HasSuffix(strings.TrimRight(text, " \t\r\n\f\v"), "?") {
		trailingQ = 1
	}
	return h, trailingQ
}

// Response is one grid cell's text, keyed by its coordinates.
type Response struct {
	Glyph    string
	Scenario int
	Arm      string
	Seed     int
	Text     string
}

// Row is the per-response mechanical score, shaped to match the python
// scores/<model>.mech.json output.
type Row struct {
	Glyph     string `json:"glyph"`
	Scenario  int    `json:"scenario"`
	Arm       string `json:"arm"`
	Seed      int    `json:"seed"`
	Hits      Hits   `json:"hits"`
	TrailingQ int    `json:"trailing_q"`
	LenDelta  int    `json:"len_delta"`
	MechFlag  int    `json:"mech_flag"`
}

type cell struct {
	glyph    string
	scenario int
}

// MechScore computes the per-response mechanical over-fire rows. It first takes
// the median baseline-arm response length per (glyph, scenario), then flags a
// response when it has any lexicon hit, ends with a question, or exceeds 1.8x the
// baseline median in length. Character length is counted in runes, matching
// python's len() over str. Ported from mech_score.main.
//
// The returned rows are sorted by (glyph, scenario, arm, seed) for a stable
// output; callers that compare against the python output should compare by key,
// since the python row order followed filesystem iteration.
func MechScore(responses []Response) []Row {
	baseLen := map[cell][]int{}
	for _, r := range responses {
		if r.Arm == "baseline" {
			c := cell{r.Glyph, r.Scenario}
			baseLen[c] = append(baseLen[c], utf8.RuneCountInString(r.Text))
		}
	}
	rows := make([]Row, 0, len(responses))
	for _, r := range responses {
		hits, tq := Markers(r.Text)
		total := 0
		for _, cat := range lexiconCats {
			total += hits[cat]
		}
		ln := utf8.RuneCountInString(r.Text)
		bl := baseLen[cell{r.Glyph, r.Scenario}]
		var med float64
		if len(bl) == 0 {
			med = float64(ln) // python: (base_len.get(...) or [len(txt)])
		} else {
			med = medianInts(bl)
		}
		lenDelta := float64(ln) - med
		// flag: any hit, a trailing question, or length > 1.8x the median. The
		// "(median or 1)" guard uses 1 when the median is zero.
		medOrOne := med
		if medOrOne == 0 {
			medOrOne = 1
		}
		flag := 0
		if total >= 1 || tq == 1 || lenDelta > 1.8*medOrOne {
			flag = 1
		}
		rows = append(rows, Row{
			Glyph: r.Glyph, Scenario: r.Scenario, Arm: r.Arm, Seed: r.Seed,
			Hits: hits, TrailingQ: tq, LenDelta: int(pyRound(lenDelta)), MechFlag: flag,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Glyph != b.Glyph {
			return a.Glyph < b.Glyph
		}
		if a.Scenario != b.Scenario {
			return a.Scenario < b.Scenario
		}
		if a.Arm != b.Arm {
			return armIndex(a.Arm) < armIndex(b.Arm)
		}
		return a.Seed < b.Seed
	})
	return rows
}

// ArmRate is a flagged-count-over-total for one arm.
type ArmRate struct {
	Arm     string
	Flagged int
	Total   int
}

// ArmRates aggregates the mechanical flag rate per arm, in Arms order.
func ArmRates(rows []Row) []ArmRate {
	agg := map[string]*ArmRate{}
	for _, r := range rows {
		a := agg[r.Arm]
		if a == nil {
			a = &ArmRate{Arm: r.Arm}
			agg[r.Arm] = a
		}
		a.Flagged += r.MechFlag
		a.Total++
	}
	out := make([]ArmRate, 0, len(Arms))
	for _, arm := range Arms {
		if a := agg[arm]; a != nil {
			out = append(out, *a)
		} else {
			out = append(out, ArmRate{Arm: arm})
		}
	}
	return out
}

func armIndex(arm string) int {
	for i, a := range Arms {
		if a == arm {
			return i
		}
	}
	return len(Arms)
}

// medianInts returns python statistics.median: the middle of a sorted odd-length
// list, or the mean of the two middle values for an even-length list.
func medianInts(xs []int) float64 {
	s := append([]int(nil), xs...)
	sort.Ints(s)
	n := len(s)
	if n%2 == 1 {
		return float64(s[n/2])
	}
	return float64(s[n/2-1]+s[n/2]) / 2
}

// pyRound rounds half to even, matching python 3's round().
func pyRound(x float64) float64 {
	r := math.Round(x)
	if math.Abs(x-math.Trunc(x)) == 0.5 {
		// halfway: round to even
		lo := math.Floor(x)
		if math.Mod(lo, 2) == 0 {
			return lo
		}
		return lo + 1
	}
	return r
}
