// Package rater scores a model's replies against a rubric. It ports the rater
// tools that lived as Python satellites (the grounded-glyph C/Ii/Ic/I/N codes
// and the action A_local/A_canon verdict) into the Go module, behind the same
// inference seam the battery uses.
//
// The scoring functions in this file are pure: they take the raw content and
// reasoning of a reply and return a code. They are the parity floor the port
// preserves, so they are characterized directly (code_test.go) against the
// reference behavior before anything calls a model.
package rater

import (
	"regexp"
	"strings"
)

// groundedCodeRE matches a grounded-glyph probe code as a whole word. The
// alternation order is load-bearing: Go's regexp is leftmost-first (the same
// semantics as Perl and Python re), so at the earliest matching position the
// first listed alternative wins — "Ii" is tried before "I". The word
// boundaries mean a code embedded in a longer word ("Icon", "In") does not
// match.
var groundedCodeRE = regexp.MustCompile(`\b(Ii|Ic|C|I|N)\b`)

// actionVerdictRE matches an action verdict as a whole word, case-insensitively.
var actionVerdictRE = regexp.MustCompile(`(?i)\b(A_local|A_canon|unscoreable|neither)\b`)

// GroundedCode extracts the rubric code from a reply. The prompt asks for the
// code as the first token, so the first match in content wins. When content
// carries no code — including empty content, from a reasoning model that spent
// its budget on reasoning_content — it falls back to the LAST match in
// reasoning: the verdict is the conclusion of the trace, not a code mentioned
// mid-deliberation. It returns "" when neither carries a code; the caller
// decides what an unparseable reply becomes.
func GroundedCode(content, reasoning string) string {
	if m := groundedCodeRE.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	all := groundedCodeRE.FindAllStringSubmatch(reasoning, -1)
	if len(all) == 0 {
		return ""
	}
	return all[len(all)-1][1]
}

// ActionVerdict extracts the committed-action verdict from a reply. It returns
// the first verdict token normalized to canonical case, or "unscoreable" when
// the reply is empty or carries no verdict token.
func ActionVerdict(reply string) string {
	if v := actionVerdict(reply); v != "" {
		return v
	}
	return "unscoreable"
}

// actionVerdict returns the canonical verdict token, or "" when the reply
// carries none. The empty return lets the rater loop count an unparseable reply
// and supply the "unscoreable" default the same way it supplies "N" for the
// grounded codes.
func actionVerdict(reply string) string {
	m := actionVerdictRE.FindStringSubmatch(reply)
	if m == nil {
		return ""
	}
	switch strings.ToLower(m[1]) {
	case "a_local":
		return "A_local"
	case "a_canon":
		return "A_canon"
	case "neither":
		return "neither"
	default:
		return ""
	}
}

// EffectiveMaxTokens is the per-rating token budget. An explicit budget always
// wins. Otherwise a hosted endpoint gets a generous default — it may be a
// reasoning model that spends the budget on reasoning_content and returns empty
// content at a small cap, coding every reply unparseable — while a local
// single-token rater keeps a tight budget.
func EffectiveMaxTokens(explicit *int, hosted bool) int {
	if explicit != nil {
		return *explicit
	}
	if hosted {
		return 2048
	}
	return 16
}
