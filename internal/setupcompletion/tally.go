package setupcompletion

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Setups is the setup-arm order used in the tally table, from tally.py SETUPS.
var Setups = []string{"raw", "loop"}

// Kappa computes Cohen's kappa over (raterA, raterB) code pairs, ported from
// tally.kappa. It returns NaN for an empty input and 1.0 when chance agreement
// is total.
func Kappa(pairs [][2]string) float64 {
	n := len(pairs)
	if n == 0 {
		return math.NaN()
	}
	same := 0
	aMarg := map[string]int{}
	bMarg := map[string]int{}
	for _, p := range pairs {
		if p[0] == p[1] {
			same++
		}
		aMarg[p[0]]++
		bMarg[p[1]]++
	}
	po := float64(same) / float64(n)
	codes := map[string]struct{}{}
	for c := range aMarg {
		codes[c] = struct{}{}
	}
	for c := range bMarg {
		codes[c] = struct{}{}
	}
	// Sum in a fixed code order so the result is deterministic across runs.
	ordered := make([]string, 0, len(codes))
	for c := range codes {
		ordered = append(ordered, c)
	}
	sort.Strings(ordered)
	pe := 0.0
	for _, c := range ordered {
		pe += (float64(aMarg[c]) / float64(n)) * (float64(bMarg[c]) / float64(n))
	}
	if pe == 1 {
		return 1.0
	}
	return (po - pe) / (1 - pe)
}

type tallyCell struct {
	glyph, setup, cond string
}

// TallyReport joins the key and two rater outputs on id and renders the exact
// per-glyph (setup x condition) table tally.py prints: strict-consensus C and Ii
// counts, per-rater C/Ii counts, and each glyph's raw agreement and kappa. The
// returned string reproduces tally.py's stdout byte-for-byte (it begins with a
// blank line, as the python's first print does).
func TallyReport(key map[string]KeyEntry, a, b map[string]string) string {
	cells, perGlyph, glyphs := joinRaters(key, a, b)
	var sb strings.Builder
	for _, glyph := range glyphs {
		fmt.Fprintf(&sb, "\n=== %s ===\n", glyph)
		fmt.Fprintf(&sb, "%-5s %-16s %3s %6s %7s  per-rater-C(A/B)  per-rater-Ii(A/B)\n",
			"setup", "condition", "n", "consC", "consIi")
		for _, setup := range Setups {
			for _, cond := range Conditions {
				writeTallyRow(&sb, setup, cond, cells[tallyCell{glyph, setup, cond}])
			}
		}
		gp := perGlyph[glyph]
		fmt.Fprintf(&sb, "  raw agreement: %s   kappa: %s   (n=%d)\n",
			fmt3(rawAgreement(gp)), fmt3(Kappa(gp)), len(gp))
	}
	return sb.String()
}

// joinRaters joins the key with both raters on id: the code pairs per cell and
// per glyph (ids either rater left uncoded are dropped), and every key glyph,
// sorted (a glyph with no pairs still gets its section).
func joinRaters(key map[string]KeyEntry, a, b map[string]string) (map[tallyCell][][2]string, map[string][][2]string, []string) {
	cells := map[tallyCell][][2]string{}
	perGlyph := map[string][][2]string{}
	glyphSet := map[string]struct{}{}
	for rid, meta := range key {
		glyphSet[meta.Glyph] = struct{}{}
		ca, oka := a[rid]
		cb, okb := b[rid]
		if !oka || !okb {
			continue
		}
		pair := [2]string{ca, cb}
		cell := tallyCell{meta.Glyph, meta.Setup, meta.Condition}
		cells[cell] = append(cells[cell], pair)
		perGlyph[meta.Glyph] = append(perGlyph[meta.Glyph], pair)
	}
	glyphs := make([]string, 0, len(glyphSet))
	for g := range glyphSet {
		glyphs = append(glyphs, g)
	}
	sort.Strings(glyphs)
	return cells, perGlyph, glyphs
}

// writeTallyRow writes one (setup, condition) row; an empty cell writes nothing.
func writeTallyRow(sb *strings.Builder, setup, cond string, pairs [][2]string) {
	n := len(pairs)
	if n == 0 {
		return
	}
	var consC, consIi, aC, bC, aIi, bIi int
	for _, p := range pairs {
		if p[0] == "C" && p[1] == "C" {
			consC++
		}
		if p[0] == "Ii" && p[1] == "Ii" {
			consIi++
		}
		aC += b2i(p[0] == "C")
		bC += b2i(p[1] == "C")
		aIi += b2i(p[0] == "Ii")
		bIi += b2i(p[1] == "Ii")
	}
	fmt.Fprintf(sb, "%-5s %-16s %3d %6d %7d  %6d/%-6d  %6d/%d\n",
		setup, cond, n, consC, consIi, aC, bC, aIi, bIi)
}

// rawAgreement is the share of pairs whose two codes match, NaN when empty.
func rawAgreement(pairs [][2]string) float64 {
	if len(pairs) == 0 {
		return math.NaN()
	}
	same := 0
	for _, p := range pairs {
		if p[0] == p[1] {
			same++
		}
	}
	return float64(same) / float64(len(pairs))
}

// b2i is 1 for true and 0 for false.
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// fmt3 formats a float to 3 decimals, matching python's "{:.3f}" — including its
// lowercase "nan" for a NaN, which Go's %f would otherwise render as "NaN".
func fmt3(x float64) string {
	if math.IsNaN(x) {
		return "nan"
	}
	return fmt.Sprintf("%.3f", x)
}
