package wrongpath

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Models and AB are the score.py iteration axes (MODELS, AB). Glyphs/Arms are
// shared with the generator.
var (
	Models = []string{"mistral", "qwen2532", "qwen38"}
	AB     = []string{"a", "b"}
)

// GT is score.py's ground truth: the a-scenario fires (yes), the b-scenario is a
// carve-out (no).
var GT = map[string]string{"a": "yes", "b": "no"}

// families is score.py FAMILIES, checked by earliest occurrence in the cited
// FIELD SOURCE string.
var families = []string{"scope", "marker", "aim", "pull", "rest"}

var (
	verdictRe = regexp.MustCompile(`(?i)VERDICT:\s*(yes|no)`)
	fieldRe   = regexp.MustCompile(`(?i)FIELD SOURCE:\s*(.+)`)
)

// ParseVerdict ports score.parse_verdict: the lower-cased yes/no after VERDICT:,
// or "" (python None) when absent.
func ParseVerdict(text string) string {
	m := verdictRe.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

// ClassifyField ports score.classify_field: "none" when no FIELD SOURCE line,
// else the earliest family keyword (by index, then family order) in the cited
// string lower-cased, else "other".
func ClassifyField(text string) string {
	m := fieldRe.FindStringSubmatch(text)
	if m == nil {
		return "none"
	}
	s := strings.ToLower(m[1])
	best := -1
	bestFam := ""
	for _, f := range families {
		idx := strings.Index(s, f)
		if idx < 0 {
			continue
		}
		if best == -1 || idx < best {
			best, bestFam = idx, f
		}
	}
	if best == -1 {
		return "other"
	}
	return bestFam
}

// CellKey identifies a grid cell.
type CellKey struct {
	Arm, Model, Glyph, AB string
}

// CellScore is score.score_cell's per-cell result.
type CellScore struct {
	N               int
	Correct         int
	CorrectScope    int
	CorrectShortcut int
}

// ScoreCell ports score.score_cell: it counts n, correct verdicts, and among
// those how many cited Scope, over the response texts (already in sorted order).
func ScoreCell(texts []string, gt string) CellScore {
	var c CellScore
	for _, text := range texts {
		c.N++
		if ParseVerdict(text) == gt {
			c.Correct++
			if ClassifyField(text) == "scope" {
				c.CorrectScope++
			}
		}
	}
	c.CorrectShortcut = c.Correct - c.CorrectScope
	return c
}

// RunFile is one response file, with its coordinates, the response text, and the
// paired reasoning trace (only populated for qwen38 base cells).
type RunFile struct {
	Arm, Model, Glyph, AB string
	Text                  string
	Reasoning             string
	HasReasoning          bool
}

// Report ports score.main: the full deterministic scoring report as a string.
// cells holds the scored cells (those with a results.json), keyed by coordinate;
// allRuns holds every response file found under the runs tree (for the
// distribution sections), in a caller-supplied deterministic order.
//
// The Counter-repr sections print counts in first-seen order over allRuns; the
// underlying counts are order-independent, so a comparison against the python
// oracle for those lines is semantic-by-key.
func Report(cells map[CellKey]CellScore, allRuns []RunFile) string {
	var b strings.Builder

	fmt.Fprintf(&b, "scored %d cells\n\n", len(cells))

	// H1 — dissociation, per model (base arm).
	b.WriteString("== H1: verdict accuracy vs scope engagement (BASE arm) ==\n")
	fmt.Fprintf(&b, "%-10s %6s %20s %20s\n", "model", "acc", "scope-cite/correct", "wrong-path/correct")
	for _, model := range Models {
		var n, corr, cs int
		for k, c := range cells {
			if k.Arm == "base" && k.Model == model {
				n += c.N
				corr += c.Correct
				cs += c.CorrectScope
			}
		}
		acc := 0.0
		if n != 0 {
			acc = float64(corr) / float64(n)
		}
		scopeRate := math.NaN()
		wpRate := math.NaN()
		if corr != 0 {
			scopeRate = float64(cs) / float64(corr)
			wpRate = 1 - scopeRate
		}
		fmt.Fprintf(&b, "%-10s %s %s %s   (correct %d/%d, scope %d)\n",
			model, pf2w(acc, 6), pf2w(scopeRate, 20), pf2w(wpRate, 20), corr, n, cs)
	}

	// H1 by scenario polarity, all models pooled (base arm).
	b.WriteString("\n== H1 by scenario polarity (BASE arm), all models pooled ==\n")
	for _, ab := range AB {
		var n, corr, cs int
		for k, c := range cells {
			if k.Arm == "base" && k.AB == ab {
				n += c.N
				corr += c.Correct
				cs += c.CorrectScope
			}
		}
		sr := math.NaN()
		if corr != 0 {
			sr = float64(cs) / float64(corr)
		}
		fmt.Fprintf(&b, "  %s (%s): acc %s  scope-cite/correct %s  (correct %d/%d)\n",
			ab, GT[ab], pf2(float64(corr)/float64(n)), pf2(sr), corr, n)
	}

	// H2 — brittleness (base vs ablated drop, by base-arm routing).
	b.WriteString("\n== H2: brittleness (BASE vs ABLATED accuracy drop, by base-arm routing) ==\n")
	for _, model := range Models {
		var shortcutDrops, scopeDrops []float64
		type detailRow struct {
			glyph, ab, routing string
			bc, ac             int
			drop               float64
		}
		var detail []detailRow
		for _, glyph := range Glyphs {
			for _, ab := range AB {
				bc, okb := cells[CellKey{"base", model, glyph, ab}]
				ac, oka := cells[CellKey{"ablated", model, glyph, ab}]
				if !okb || !oka {
					continue
				}
				routing := "uninform"
				if bc.Correct >= 2 {
					if bc.CorrectShortcut > bc.CorrectScope {
						routing = "shortcut"
					} else {
						routing = "scope"
					}
				}
				drop := float64(bc.Correct-ac.Correct) / float64(bc.N)
				detail = append(detail, detailRow{glyph, ab, routing, bc.Correct, ac.Correct, drop})
				switch routing {
				case "shortcut":
					shortcutDrops = append(shortcutDrops, drop)
				case "scope":
					scopeDrops = append(scopeDrops, drop)
				}
			}
		}
		sc := meanOrNaN(shortcutDrops)
		scp := meanOrNaN(scopeDrops)
		diff := math.NaN()
		if len(shortcutDrops) > 0 && len(scopeDrops) > 0 {
			diff = sc - scp
		}
		fmt.Fprintf(&b, "\n  %s: shortcut-cell mean drop %s (n=%d), scope-cell mean drop %s (n=%d), differential %s\n",
			model, pf2(sc), len(shortcutDrops), pf2(scp), len(scopeDrops), pf2p(diff))
		for _, d := range detail {
			fmt.Fprintf(&b, "    %s-%s %-8s base %d/8 -> ablated %d/8  drop %s\n",
				d.glyph, d.ab, d.routing, d.bc, d.ac, pf2p(d.drop))
		}
	}

	// Field-source class distribution per model/arm (all runs).
	b.WriteString("\n== Field-source class distribution per model/arm (all runs) ==\n")
	for _, model := range Models {
		for _, arm := range Arms {
			oc := newOrderedCounter()
			for _, r := range allRuns {
				if r.Arm == arm && r.Model == model {
					oc.add(ClassifyField(r.Text))
				}
			}
			fmt.Fprintf(&b, "  %-8s %-9s %s\n", arm, model, oc.repr())
		}
	}

	// Field-source among correct verdicts, by model and polarity (base arm).
	b.WriteString("\n== Field-source among correct verdicts, by model and polarity (BASE arm) ==\n")
	for _, model := range Models {
		for _, ab := range AB {
			oc := newOrderedCounter()
			for _, glyph := range Glyphs {
				for _, r := range allRuns {
					if r.Arm == "base" && r.Model == model && r.Glyph == glyph && r.AB == ab {
						if ParseVerdict(r.Text) == GT[ab] {
							oc.add(ClassifyField(r.Text))
						}
					}
				}
			}
			fmt.Fprintf(&b, "  %-9s %s (%-3s) %s\n", model, ab, GT[ab], oc.repr())
		}
	}

	// W1/W2: qwen3.8 base traces that worked scope despite a non-scope citation.
	b.WriteString("\n== Qwen3.8 base: scope worked in the trace despite a non-scope citation ==\n")
	nonscope, mention := 0, 0
	for _, r := range allRuns {
		if r.Arm != "base" || r.Model != "qwen38" {
			continue
		}
		if ClassifyField(r.Text) == "scope" {
			continue
		}
		nonscope++
		if r.HasReasoning {
			low := strings.ToLower(r.Reasoning)
			if strings.Contains(low, "scope") || strings.Contains(low, "operative") {
				mention++
			}
		}
	}
	fmt.Fprintf(&b, "  non-scope-cited base runs: %d; trace mentions scope/operative: %d\n", nonscope, mention)

	return b.String()
}

// meanOrNaN reproduces statistics.mean over a non-empty slice, NaN when empty.
func meanOrNaN(xs []float64) float64 {
	if len(xs) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// pf2 formats a float as python f"{x:.2f}" (nan -> "nan").
func pf2(x float64) string {
	if math.IsNaN(x) {
		return "nan"
	}
	return fmt.Sprintf("%.2f", x)
}

// pf2w formats a float right-justified in width w as python f"{x:{w}.2f}".
func pf2w(x float64, w int) string {
	if math.IsNaN(x) {
		return fmt.Sprintf("%*s", w, "nan")
	}
	return fmt.Sprintf("%*.2f", w, x)
}

// pf2p formats a float as python f"{x:+.2f}" (nan -> "+nan").
func pf2p(x float64) string {
	if math.IsNaN(x) {
		return "+nan"
	}
	return fmt.Sprintf("%+.2f", x)
}

// orderedCounter reproduces collections.Counter's dict repr: counts by key with
// keys in first-seen order.
type orderedCounter struct {
	order []string
	count map[string]int
}

func newOrderedCounter() *orderedCounter {
	return &orderedCounter{count: map[string]int{}}
}

func (o *orderedCounter) add(k string) {
	if _, ok := o.count[k]; !ok {
		o.order = append(o.order, k)
	}
	o.count[k]++
}

// repr renders like python repr(dict(counter)): {} or {'k': v, 'k2': v2}.
func (o *orderedCounter) repr() string {
	if len(o.order) == 0 {
		return "{}"
	}
	var parts []string
	for _, k := range o.order {
		parts = append(parts, fmt.Sprintf("'%s': %d", k, o.count[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// ByKey returns the counts as a plain map, for semantic (order-independent)
// comparison of the Counter-repr sections against the python oracle.
func (o *orderedCounter) ByKey() map[string]int {
	out := make(map[string]int, len(o.count))
	for k, v := range o.count {
		out[k] = v
	}
	return out
}
