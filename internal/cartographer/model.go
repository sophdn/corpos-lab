// Package cartographer ports the cartographer-duty-format provenance-scoring
// python to Go. The study (glyph-research, chain eradicate-python-rewrite-in-go)
// backs the published paper "Derived or Observed"
// (10.5281/zenodo.22726758): a blind coverage assay comparing four duty-format
// conditions — baseline, an annotated instrument that cites canon slugs, a
// cartographer instrument that derives its duties, and a cartographer instrument
// with a meta-taboo scan — over ten pre-declared taboo classes.
//
// The logic here is sans-IO: callers read the study's taboo_set.json,
// coverage_grid.json, blind_map.json, judge output, and per-run response texts,
// and pass the parsed data in. The report-producing functions return the exact
// text the python printed, so a caller can diff against the oracle byte for byte.
// cmd/corpos-lab/pubscore_cartographer.go does the file wiring.
package cartographer

// Conds are the four experiment conditions, in the fixed report order used by
// analyze and the keyword pass. Ported from analyze.CONDS.
var Conds = []string{
	"baseline",
	"annotated_instrument",
	"cartographer_instrument",
	"cartographer_scan_instrument",
}

// FirstPrinciples is the taboo class whose hazards are derivable by analyzing the
// repair act itself. It is the pivot of the condition-by-class interaction.
const FirstPrinciples = "first_principles"

// Taboo is one pre-declared taboo class. The python reads more fields from
// taboo_set.json (divergence_point, markers); the scoring uses only these three.
type Taboo struct {
	Slug       string `json:"slug"`
	Class      string `json:"class"`
	ScanTarget bool   `json:"scan_target"`
}

// GridEntry is one coverage-grid row: a judged duty with its per-slug coverage
// calls. Entries are kept in the grid's file order because analyze_duty reports
// per-condition means in the order conditions first appear in the grid.
type GridEntry struct {
	Key       string         // "<condition>_<run>"
	Condition string         `json:"condition"`
	Run       int            `json:"run"`
	Coverage  map[string]int `json:"coverage"`
}

// fpSlugs returns the first-principles taboo slugs in taboo-file order.
func fpSlugs(taboos []Taboo) []string {
	var out []string
	for _, t := range taboos {
		if t.Class == FirstPrinciples {
			out = append(out, t.Slug)
		}
	}
	return out
}

// nonFPSlugs returns the non-first-principles taboo slugs in taboo-file order.
func nonFPSlugs(taboos []Taboo) []string {
	var out []string
	for _, t := range taboos {
		if t.Class != FirstPrinciples {
			out = append(out, t.Slug)
		}
	}
	return out
}

// rate counts covered (k) over judged (n) for one condition across the given
// slugs. A slug absent from a row's coverage map is not judged and is skipped,
// matching python's `row["coverage"].get(s) is None` guard. Ported from
// analyze.rate.
func rate(grid []GridEntry, cond string, slugs []string) (k, n int) {
	for _, e := range grid {
		if e.Condition != cond {
			continue
		}
		for _, s := range slugs {
			v, ok := e.Coverage[s]
			if !ok {
				continue
			}
			n++
			if v != 0 {
				k++
			}
		}
	}
	return k, n
}
