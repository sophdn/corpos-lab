package restaxis

import (
	"strings"
	"testing"
)

func TestBorderlineStr(t *testing.T) {
	if got := borderlineStr(nil); got != "none" {
		t.Errorf("empty borderline = %q, want none", got)
	}
	if got := borderlineStr([]int{6, 3, 16}); got != "3, 6, 16" {
		t.Errorf("borderlineStr sort/join = %q, want '3, 6, 16'", got)
	}
}

func TestCompileReviewHeader(t *testing.T) {
	doc := CompileReview("/cand", "/study", nil)
	if !strings.HasPrefix(doc, "# Rest-axis benchmark — neutral scenarios for review\n") {
		t.Errorf("header missing/wrong; starts: %q", doc[:min(80, len(doc))])
	}
	if !strings.Contains(doc, "**⚠ borderline**") {
		t.Errorf("borderline legend missing")
	}
	if !strings.HasSuffix(doc, "---\n\n") {
		t.Errorf("with no glyphs the doc should end after the --- separator; got tail %q", doc[len(doc)-8:])
	}
}

func TestCompileReviewGlyphBlock(t *testing.T) {
	scen := make([]string, 20)
	for i := range scen {
		scen[i] = "scenario body " + string(rune('a'+i))
	}
	g := ReviewGlyph{
		Slug:       "casg-direct",
		Candidate:  "CANDIDATE_casg-direct.md",
		Borderline: []int{3, 6},
		Scenarios:  scen,
	}
	doc := CompileReview("/cand", "/study", []ReviewGlyph{g})

	// glyph header + cited paths
	for _, want := range []string{
		"\n# Glyph: `casg-direct`\n",
		"- Firing condition (open to check): `/cand/CANDIDATE_casg-direct.md`",
		"`/study/casg-direct/CORRECT_COMPLETIONS.md` (scenarios 1–4) and `/study/casg-direct/CORRECT_COMPLETIONS_NEW.md` (5–20)",
		"Borderline (auditor least sure): 3, 6\n",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("missing segment: %q", want)
		}
	}
	// original vs new tag
	if !strings.Contains(doc, "### casg-direct — scenario 1  [not-fire · *original*]") {
		t.Errorf("scenario 1 original tag missing")
	}
	if !strings.Contains(doc, "### casg-direct — scenario 5  [not-fire · *new*]") {
		t.Errorf("scenario 5 new tag missing")
	}
	// borderline flag appears only on 3 and 6
	if !strings.Contains(doc, "### casg-direct — scenario 3  [not-fire · *original*]  ⚠ **borderline**") {
		t.Errorf("scenario 3 borderline flag missing")
	}
	if strings.Contains(doc, "scenario 1  [not-fire · *original*]  ⚠") {
		t.Errorf("scenario 1 should not be borderline")
	}
	// fenced body
	if !strings.Contains(doc, "```\nscenario body a\n```") {
		t.Errorf("scenario 1 body fence missing")
	}
}

func TestCompileReviewMissingScenario(t *testing.T) {
	// short Scenarios slice -> missing entries render as (MISSING)
	g := ReviewGlyph{Slug: "g", Candidate: "c.md", Scenarios: []string{"only one"}}
	doc := CompileReview("/cand", "/study", []ReviewGlyph{g})
	if !strings.Contains(doc, "```\n(MISSING)\n```") {
		t.Errorf("(MISSING) fallback not rendered")
	}
	if !strings.Contains(doc, "```\nonly one\n```") {
		t.Errorf("present scenario 1 not rendered")
	}
}
