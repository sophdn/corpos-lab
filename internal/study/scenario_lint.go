package study

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// A scenario must be SELF-CONTAINED: it has to include everything the subject
// needs to perform the task, because the subject sees only the assembled prompt.
// On the rest-axis benchmark a conditional-gate scenario referenced GUIDE.md
// content that was absent from the prompt, so its runs were unscoreable — caught
// only at scoring time, after the whole grid had run (suggestion
// neutrality-audit-should-check-scenario-self-containedness).
//
// ScenarioSelfContainmentWarnings scans a scenario's text for references to
// content that a reader would expect to be inlined but that a pointer leaves
// out. It is a heuristic and it only WARNS: a scenario that legitimately names a
// file the task acts on will trip it, and that is acceptable for an advisory
// check the author reads and dismisses. It never blocks a run.

// docRefPattern matches a reference to a document file by name — the shape of the
// GUIDE.md failure. A bare data file the task operates on (config.json) also
// matches; the warning names the reference so the author can judge.
var docRefPattern = regexp.MustCompile(`(?i)\b[\w-]+\.(md|markdown|txt|rst|pdf|docx?)\b`)

// pointerPattern matches phrases that promise content from elsewhere — "see the
// guide", "as described in the document" — the verbal form of the same gap.
var pointerPattern = regexp.MustCompile(`(?i)\b(?:see|refer to|as (?:described|shown|noted|defined|specified)(?: in)?|according to|per|following)\s+(?:the\s+|your\s+|our\s+)?(guide|document|doc|spec|specification|readme|instructions|instruction|attachment|attached|manual|handbook|appendix|above|below)\b`)

// ScenarioSelfContainmentWarnings returns one advisory warning per distinct
// dangling reference found in the scenario text. It returns nil for a
// self-contained scenario. The warnings are stable-ordered so a run record is
// reproducible.
func ScenarioSelfContainmentWarnings(text string) []string {
	found := map[string]bool{}

	for _, m := range docRefPattern.FindAllString(text, -1) {
		key := "references the document " + strings.ToLower(m)
		found[key] = true
	}
	for _, m := range pointerPattern.FindAllString(text, -1) {
		key := "points outside the prompt: " + strings.ToLower(strings.Join(strings.Fields(m), " "))
		found[key] = true
	}

	if len(found) == 0 {
		return nil
	}
	keys := make([]string, 0, len(found))
	for k := range found {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	warnings := make([]string, 0, len(keys))
	for _, k := range keys {
		warnings = append(warnings, fmt.Sprintf(
			"scenario self-containment: %s — confirm that content is inlined in the "+
				"prompt, or the subject cannot perform the task and its runs are unscoreable", k))
	}
	return warnings
}
