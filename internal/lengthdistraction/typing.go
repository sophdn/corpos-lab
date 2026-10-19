// Package lengthdistraction types a glyph decision class by whether a long prefix
// suppresses correct action through OFF-TASK derailment (probe code N) rather than
// through comprehension of the decision.
//
// It names a fourth mechanism type beside the alphabet assay's comprehension,
// recognition, and mere structure (INQUIRY.md, How we measure — Mechanism
// controls). Those three are read on the correct-action rate (C) against the
// scrambled and off-target controls, both of which keep the three-axis glyph
// shape. None of them isolates plain length. The neutral-prefix control (chain
// 548) supplies the missing arm: a length-matched, content-free, form-free prose
// block. When that block reproduces the glyph's suppression the effect is a bare
// length effect, and the failure mode is code N — the model summarizes the prefix
// instead of doing the task.
//
// The existing neutralprefix.Aggregate reports the C rate only, so it cannot see
// the derailment. This package reports the N rate per class x condition and
// applies the typing criterion, over a class list supplied by the caller rather
// than the four classes chain 548 hardcoded.
package lengthdistraction

import (
	"fmt"
	"sort"
	"strings"

	"corpos-lab/internal/neutralprefix"
)

// Majority reduces several rater codes for one id to a single consensus code by
// vote. A code wins when strictly more than half of the raters supplied voted for
// it (empty from a rater is a non-vote, not a vote for a missing code). With no
// winner — a three-way tie, or no rater scoring the id — the result is CodeSplit,
// so a contested id lowers a rate rather than being credited to a plurality. This
// is the >2-rater generalization of strict two-rater consensus, matching the
// independent-family consensus the scoring standard calls the measure of record.
func Majority(id string, raters ...map[string]string) string {
	if len(raters) == 0 {
		return CodeSplit
	}
	votes := map[string]int{}
	for _, r := range raters {
		if c := r[id]; c != "" {
			votes[c]++
		}
	}
	// Deterministic scan: sort the voted codes so a tie resolves the same way
	// every run (a Go map's iteration order is randomized).
	codes := make([]string, 0, len(votes))
	for c := range votes {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	best, bestN := "", 0
	for _, c := range codes {
		if votes[c] > bestN {
			best, bestN = c, votes[c]
		}
	}
	if bestN*2 > len(raters) {
		return best
	}
	return CodeSplit
}

// MajorityMap builds the per-id majority-consensus code map over every id in the
// key, from the rater maps supplied. Passed to Report or Tally as both a and b
// (with hasB true), it makes the strict-consensus path report the majority code
// per id, with no-majority ids folded to CodeSplit.
func MajorityMap(key *neutralprefix.KeyFile, raters ...map[string]string) map[string]string {
	out := make(map[string]string, len(key.Order))
	for _, id := range key.Order {
		out[id] = Majority(id, raters...)
	}
	return out
}

// The four probe codes a rubric assigns, plus the two synthetic markers this
// package records when two raters do not agree or a rater left an id unscored.
// The definitions are the one scoring standard (CLAUDE.md): only a performed
// correct action is C; a recommendation or description without the action is Ii;
// a defective action is Ic; no recognition is I; off-task or malformed is N.
const (
	CodeC       = "C"
	CodeIi      = "Ii"
	CodeIc      = "Ic"
	CodeI       = "I"
	CodeN       = "N"
	CodeSplit   = "split" // the two raters disagreed
	CodeUnrated = "?"     // an id absent from a rater map
)

// Dist is the code distribution for one class x condition slice. Split counts the
// ids where two raters disagreed; Unscored counts ids missing from a rater map.
// Total is every id in the slice, so Frac denominators never silently drop a cell.
type Dist struct {
	C, Ii, Ic, I, N int
	Split           int
	Unscored        int
	Total           int
}

// NRate is the off-task share — the length-distraction signal. It divides by Total
// (every id in the slice), so a split or unscored id lowers the rate rather than
// vanishing from the denominator. Total 0 returns 0.
func (d Dist) NRate() float64 {
	if d.Total == 0 {
		return 0
	}
	return float64(d.N) / float64(d.Total)
}

// CRate is the correct-action share, on the same Total denominator as NRate.
func (d Dist) CRate() float64 {
	if d.Total == 0 {
		return 0
	}
	return float64(d.C) / float64(d.Total)
}

// add folds one consensus code into the distribution.
func (d *Dist) add(code string) {
	d.Total++
	switch code {
	case CodeC:
		d.C++
	case CodeIi:
		d.Ii++
	case CodeIc:
		d.Ic++
	case CodeI:
		d.I++
	case CodeN:
		d.N++
	case CodeSplit:
		d.Split++
	default:
		d.Unscored++
	}
}

// consensus reduces two rater codes for one id to a single recorded code. With
// hasB it is strict: the code when both raters agree, else CodeSplit. Without a
// second rater it is rater A's own code, or CodeUnraTed when A never scored the id.
// A blank code from either rater (an id the rater omitted) makes the pair a split
// under hasB and unscored under single-rater, never a silent agreement.
func consensus(aCode, bCode string, hasB bool) string {
	if !hasB {
		if aCode == "" {
			return CodeUnrated
		}
		return aCode
	}
	if aCode == "" || bCode == "" {
		return CodeSplit
	}
	if aCode == bCode {
		return aCode
	}
	return CodeSplit
}

// Tally builds the code distribution for one class x condition slice of the key.
// It walks the key in its recorded order and, for every id whose metadata matches
// class and cond, folds the consensus of the two rater maps. b is ignored when
// hasB is false.
func Tally(key *neutralprefix.KeyFile, a, b map[string]string, hasB bool, class, cond string) Dist {
	var d Dist
	for _, id := range key.Order {
		m := key.Meta[id]
		if m.Cls != class || m.Condition != cond {
			continue
		}
		d.add(consensus(a[id], b[id], hasB))
	}
	return d
}

// Classes returns the distinct class names in the key, in first-seen order, so a
// report lists exactly the classes the run produced without a hardcoded list.
func Classes(key *neutralprefix.KeyFile) []string {
	seen := map[string]bool{}
	var out []string
	for _, id := range key.Order {
		c := key.Meta[id].Cls
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}

// Verdict is a class's mechanism-type reading on the length-distraction axis.
type Verdict string

const (
	// IsLengthDistraction: the neutral prefix reproduces the off-task derailment
	// and the short imperative does not — the suppression is a bare length effect.
	IsLengthDistraction Verdict = "length-distraction"
	// NotLengthDistraction: the neutral prefix does not derail the model, so any
	// suppression the glyph produces is not a generic length effect.
	NotLengthDistraction Verdict = "not-length-distraction"
	// Inconclusive: the neutral prefix derails AND the imperative also derails, so
	// length alone is not isolated — the short content-matched control did not stay
	// clean, and the criterion cannot separate length from content here.
	Inconclusive Verdict = "inconclusive"
)

// Type applies the typing criterion for one class. A class is length-distraction
// when the length-matched neutral prefix reproduces the off-task derailment
// (neutral N rate at or above hiN) AND the short content-matched imperative does
// not (imperative N rate at or below loN). When the neutral prefix does not derail
// the class is not length-distraction; when both derail the read is inconclusive.
//
// hiN and loN are explicit inputs, not baked constants: the criterion is a reading
// rule the caller states with the study's chosen values, in the lab's observe-don't-
// assert spirit. The neutral and imperative distributions come from Tally.
func Type(neutral, imperative Dist, hiN, loN float64) Verdict {
	if neutral.NRate() < hiN {
		return NotLengthDistraction
	}
	if imperative.NRate() > loN {
		return Inconclusive
	}
	return IsLengthDistraction
}

// Report renders the per-class x condition N-rate and C-rate table and the
// per-class length-distraction verdict, over the classes and conditions given.
// conds is the report column order; the verdict reads the "neutral_prefix" and
// "imperative_only" columns, so both must be present in conds for a verdict line.
func Report(key *neutralprefix.KeyFile, a, b map[string]string, hasB bool, classes, conds []string, hiN, loN float64) string {
	var sb strings.Builder

	sb.WriteString("== length-distraction typing ==")
	if hasB {
		sb.WriteString("  [strict-consensus]")
	} else {
		sb.WriteString("  [rater-a]")
	}
	fmt.Fprintf(&sb, "  (criterion: neutral N>=%.2f AND imperative N<=%.2f)\n", hiN, loN)

	hdr := ljust("class", 36) + ljust("condition", 18) + ljust("N", 7) + ljust("C", 7) + "n"
	sb.WriteString(hdr + "\n")

	dists := map[string]map[string]Dist{}
	for _, cls := range classes {
		dists[cls] = map[string]Dist{}
		for _, cond := range conds {
			d := Tally(key, a, b, hasB, cls, cond)
			dists[cls][cond] = d
			row := ljust(cls, 36) + ljust(cond, 18) +
				ljust(f42(d.NRate()), 7) + ljust(f42(d.CRate()), 7) +
				fmt.Sprintf("%d", d.Total)
			sb.WriteString(row + "\n")
		}
	}

	sb.WriteString("\nverdict:\n")
	for _, cls := range classes {
		neu := dists[cls]["neutral_prefix"]
		imp := dists[cls]["imperative_only"]
		v := Type(neu, imp, hiN, loN)
		fmt.Fprintf(&sb, "  %s %s  (neutral N=%.2f, imperative N=%.2f)\n",
			ljust(cls, 36), v, neu.NRate(), imp.NRate())
	}
	return sb.String()
}

// SortedClasses returns the key's classes in stable alphabetical order, for a
// caller that wants a deterministic report independent of run order.
func SortedClasses(key *neutralprefix.KeyFile) []string {
	c := Classes(key)
	sort.Strings(c)
	return c
}

// f42 formats a rate as width-4 two-decimal, matching the neutralprefix report so
// the two tables line up when read side by side.
func f42(x float64) string { return fmt.Sprintf("%4.2f", x) }

// ljust left-justifies s to width w with spaces (ASCII, byte-for-byte).
func ljust(s string, w int) string {
	if len(s) >= w {
		return s + " "
	}
	return s + strings.Repeat(" ", w-len(s))
}
