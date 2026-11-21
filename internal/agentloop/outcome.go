package agentloop

import "strings"

// ParseOutcome is how the harness read one model turn. It is an instrument note
// for the reader of a run, not a score: a NULL or step-capped cell whose turns
// were mostly lost calls says something about the parser, not about the model's
// disposition. Agents used to count lost calls by grepping transcripts for
// empty-argument EXECUTED lines (18 in one grid).
type ParseOutcome string

const (
	// OutcomeParsed: every call the turn carried was read with its arguments.
	OutcomeParsed ParseOutcome = "parsed"
	// OutcomeFinal: the turn was a FINAL.
	OutcomeFinal ParseOutcome = "final"
	// OutcomeNone: the turn named no tool and no FINAL, and had no call markup —
	// a stall the subject chose, not a parse loss.
	OutcomeNone ParseOutcome = "none"
	// OutcomeEmptyArg: a call was read but lost its argument (a read or query
	// with no target, an edit with no path or no content).
	OutcomeEmptyArg ParseOutcome = "empty-arg"
	// OutcomeUnknownTool: a call named a tool the loop does not offer.
	OutcomeUnknownTool ParseOutcome = "unknown-tool"
	// OutcomeMalformed: the turn carried call markup the parser could not read,
	// or a call whose argument still holds tag text.
	OutcomeMalformed ParseOutcome = "malformed"
)

// severity orders outcomes so a turn reports its worst call.
var severity = map[ParseOutcome]int{
	OutcomeParsed: 0, OutcomeEmptyArg: 1, OutcomeUnknownTool: 2, OutcomeMalformed: 3,
}

// TurnParse is the parse record of one turn.
type TurnParse struct {
	Outcome ParseOutcome
	// Lost counts calls the turn attempted that did not run as a usable call.
	Lost int
	// Ignored counts directives the harness did not run by design: CALL lines
	// after the first (one call per turn is the protocol), native calls after a
	// CALL that won the turn, and a FINAL after native calls. A planner that
	// writes its whole plan in one turn shows here, so a reader can tell an
	// apparatus fit problem from a disposition (suggestion 239).
	Ignored int
}

// ClassifyTurn reads the parse outcome of one turn from its text and the actions
// ParseActions returned for it.
func ClassifyTurn(modelText string, acts []Action) TurnParse {
	directives := countDirectives(modelText)
	if len(acts) == 1 {
		switch acts[0].Kind {
		case ActionFinal:
			return TurnParse{Outcome: OutcomeFinal}
		case ActionNone:
			if firstNativeMarker(modelText) >= 0 {
				return TurnParse{Outcome: OutcomeMalformed, Lost: 1}
			}
			return TurnParse{Outcome: OutcomeNone}
		}
	}
	tp := TurnParse{Outcome: OutcomeParsed}
	for _, a := range acts {
		o := callOutcome(a)
		if o != OutcomeParsed {
			tp.Lost++
		}
		if severity[o] > severity[tp.Outcome] {
			tp.Outcome = o
		}
	}
	if extra := directives - len(acts); extra > 0 {
		tp.Ignored = extra
	}
	return tp
}

// callOutcome classifies one executed call.
func callOutcome(a Action) ParseOutcome {
	if a.Kind == ActionUnknown {
		return OutcomeUnknownTool
	}
	if hasTagText(a.Arg) {
		return OutcomeMalformed
	}
	switch a.Kind {
	case ActionRead, ActionQuery:
		if strings.TrimSpace(a.Arg) == "" {
			return OutcomeEmptyArg
		}
	case ActionEdit:
		if strings.TrimSpace(a.Arg) == "" || a.Content == "" {
			return OutcomeEmptyArg
		}
	}
	// list_files with no argument lists the root, which is a valid call.
	return OutcomeParsed
}

// hasTagText reports whether an argument still carries tool-call markup, the
// sign of a parse that ran the tag soup as the argument.
func hasTagText(s string) bool {
	for _, tok := range []string{"<parameter", "</parameter", "<function", "</function", "<tool_call", "</tool_call"} {
		if strings.Contains(s, tok) {
			return true
		}
	}
	return false
}

// countDirectives counts the calls a turn wrote: each CALL line (a CALL FINAL
// counts too, being written in the call form) plus each native block.
func countDirectives(text string) int {
	n := len(nativeBlocks(text))
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "CALL ") {
			n++
		}
	}
	return n
}

// ClaimedEdits lists the paths of every edit_file the turn wrote, in order:
// each "CALL edit_file" line and each native edit block, whether or not the
// harness ran it. A reader compares it with the edits that executed to see which
// claimed writes never happened.
func ClaimedEdits(modelText string) []string {
	var paths []string
	for _, line := range strings.Split(modelText, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "CALL edit_file"); ok {
			path, _, _ := strings.Cut(rest, editSeparator)
			paths = append(paths, strings.TrimSpace(path))
		}
	}
	for _, b := range nativeBlocks(modelText) {
		if a, ok := parseNative(b); ok && a.Kind == ActionEdit {
			paths = append(paths, a.Arg)
		}
	}
	return paths
}
