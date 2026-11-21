package restaxis

// This file ports compile_review.py: it assembles the neutral-scenario review
// document (SCENARIO_REVIEW.md) for human review. The transform is pure and
// sans-IO — the caller reads the scenario materials and passes them in, along
// with the candidate and study directory paths that the document cites. The
// output is byte-identical to the python "\n".join(out) + "\n" assembly.

import (
	"sort"
	"strconv"
	"strings"
)

// ReviewGlyph is one glyph's block in the review document: its slug, the
// candidate (firing-condition) filename cited under candDir, the blind auditor's
// borderline scenario numbers, and the 20 scenario texts (index i is scenario
// i+1). A scenario the caller could not read is passed as "(MISSING)".
type ReviewGlyph struct {
	Slug       string
	Candidate  string
	Borderline []int
	Scenarios  []string
}

// CompileReview builds the review document. candDir is the candidates directory
// (cited as candDir/<candidate>), studyDir is the study root (cited as
// studyDir/<slug>/CORRECT_COMPLETIONS.md). It mirrors compile_review.py exactly,
// including the fixed header, the per-glyph header lines, and the fenced
// scenario blocks numbered 1..20.
func CompileReview(candDir, studyDir string, glyphs []ReviewGlyph) string {
	var out []string
	out = append(out, "# Rest-axis benchmark — neutral scenarios for review\n")
	out = append(out, "20 neutral scenarios per glyph, 4 glyphs = 80 total. Scenarios 1–4 are the "+
		"original audited set; 5–20 are new this session. Every scenario was classified "+
		"**not-fire** by an independent blind audit (a fresh subagent that judged each "+
		"against the glyph's firing condition without being told the intent).\n")
	out = append(out, "A scenario is **neutral** when the correct behaviour is to just do the task — no "+
		"gate, hedge, suppression, refusal, or unrequested extra action. **⚠ borderline** "+
		"marks the scenarios the blind auditor was least sure about; those most deserve your eye.\n")
	out = append(out, "To cross-check a scenario against the real firing rule and the expected over-fire, "+
		"open the two files listed under each glyph — I keep out of those so my judgement of "+
		"neutrality stays uncontaminated; yours does not have that constraint.\n")
	out = append(out, "---\n")

	for _, g := range glyphs {
		blSet := map[int]bool{}
		for _, n := range g.Borderline {
			blSet[n] = true
		}
		out = append(out, "\n# Glyph: `"+g.Slug+"`\n")
		out = append(out, "- Firing condition (open to check): `"+candDir+"/"+g.Candidate+"`")
		out = append(out, "- Over-fire signatures (open to check): `"+studyDir+"/"+g.Slug+"/CORRECT_COMPLETIONS.md` "+
			"(scenarios 1–4) and `"+studyDir+"/"+g.Slug+"/CORRECT_COMPLETIONS_NEW.md` (5–20)")
		out = append(out, "- Blind-audit verdict: all 20 **not-fire**. Borderline (auditor least sure): "+borderlineStr(g.Borderline)+"\n")
		for n := 1; n <= 20; n++ {
			text := "(MISSING)"
			if n-1 < len(g.Scenarios) {
				text = g.Scenarios[n-1]
			}
			flag := ""
			if blSet[n] {
				flag = "  ⚠ **borderline**"
			}
			newtag := " · *original*"
			if n >= 5 {
				newtag = " · *new*"
			}
			out = append(out, "\n### "+g.Slug+" — scenario "+strconv.Itoa(n)+"  [not-fire"+newtag+"]"+flag+"\n")
			out = append(out, "```")
			out = append(out, text)
			out = append(out, "```")
		}
	}
	return strings.Join(out, "\n") + "\n"
}

// borderlineStr renders the borderline set as python does:
// ", ".join(str(n) for n in sorted(borderline)) or "none".
func borderlineStr(borderline []int) string {
	if len(borderline) == 0 {
		return "none"
	}
	nums := append([]int(nil), borderline...)
	sort.Ints(nums)
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ", ")
}
