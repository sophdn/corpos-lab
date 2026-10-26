package main

import (
	"fmt"
	"os"
	"path/filepath"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("compile-review", raCompileReview) }

// reviewGlyphs is the fixed glyph set of the review document, matching the GLYPHS
// table in compile_review.py: slug, candidate (firing-condition) filename, and
// the blind auditor's borderline scenario numbers.
var reviewGlyphs = []restaxis.ReviewGlyph{
	{Slug: "casg-direct", Candidate: "CANDIDATE_casg-direct_2026-04-03.md",
		Borderline: []int{3, 4, 6, 13, 14, 15, 16, 17, 18, 19, 20}},
	{Slug: "conditional-gate-uniform-default", Candidate: "CANDIDATE_conditional-gate-uniform-default_2026-03-30.md",
		Borderline: []int{3, 6, 16}},
	{Slug: "formal-step-context-bypass", Candidate: "CANDIDATE_formal-step-context-bypass_2026-03-29.md",
		Borderline: []int{2, 4, 7, 11, 14, 19, 20}},
	{Slug: "parent-state-check-bypass", Candidate: "CANDIDATE_parent-state-check-bypass_2026-03-30.md",
		Borderline: []int{3, 4, 6, 9, 10, 11}},
}

// defaultCandDir is the candidates directory cited by the review document,
// matching CAND in compile_review.py.
const defaultCandDir = "corpus/private/glyph-model/candidates"

// raCompileReview compiles the neutral-scenario review document. Ported from
// compile_review.py.
//
//	corpos-lab rest-axis compile-review [--cand <candidates-dir>] [--out <file>] <study-dir>
//
// It reads each glyph's 20 scenario materials, tags them, and assembles the
// document. The candidate and _NEW paths are cited in the text for the human to
// open; the tool does not read them.
func raCompileReview(args []string) error {
	cand, args := flagValue(args, "--cand")
	out, args := flagValue(args, "--out")
	if len(args) != 1 {
		return fmt.Errorf("need a study-dir")
	}
	study := args[0]
	if cand == "" {
		cand = defaultCandDir
	}
	if out == "" {
		out = filepath.Join(study, "SCENARIO_REVIEW.md")
	}

	glyphs := make([]restaxis.ReviewGlyph, len(reviewGlyphs))
	for i, g := range reviewGlyphs {
		g.Scenarios = make([]string, 20)
		for n := 1; n <= 20; n++ {
			p := filepath.Join(study, g.Slug, "materials", fmt.Sprintf("scenario_%d.md", n))
			if b, err := os.ReadFile(p); err == nil { //nolint:gosec // a study materials path
				g.Scenarios[n-1] = restaxis.StripText(string(b))
			} else {
				g.Scenarios[n-1] = "(MISSING)"
			}
		}
		glyphs[i] = g
	}

	doc := restaxis.CompileReview(cand, study, glyphs)
	if err := os.WriteFile(out, []byte(doc), 0o600); err != nil {
		return err
	}
	scenarios := 0
	for range glyphs {
		scenarios += 20
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes, %d scenarios)\n", out, len(doc), scenarios)
	return nil
}
