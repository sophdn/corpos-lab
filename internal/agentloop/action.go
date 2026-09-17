// Package agentloop is the minimal tool-using loop for the agentic-loop probe.
// It wraps the same llama.cpp portal the single-turn probe uses and gives the
// subject the one thing raw completion withholds: the ability to act. The
// subject emits one tool call per turn as a line of text; the harness executes
// it against an in-memory sandbox and feeds back the real result; the loop
// repeats until the subject finishes or a step cap is hit.
//
// This is the harness arm of chain 550 (setup-completion-vs-agentic-loop). Its
// question is whether analysis-mode — recognition without execution — survives
// when the subject can act (see studies/setup-completion-vs-agentic-loop). The
// loop is deliberately minimal: one portal, a tiny tool vocabulary, a text
// protocol with no server-side tool parsing, and a sandbox with no real
// filesystem so it runs unchanged inside the distroless assay container.
package agentloop

import "strings"

// ActionKind classifies one parsed model turn.
type ActionKind int

const (
	// ActionNone: the turn named no tool and no FINAL — the subject stalled.
	// This is the loop's analogue of the raw-completion tool-stall.
	ActionNone ActionKind = iota
	// ActionList: list_files <dir>.
	ActionList
	// ActionRead: read_file <path>.
	ActionRead
	// ActionQuery: run_query <query>.
	ActionQuery
	// ActionEdit: edit_file <path> ||| <contents>.
	ActionEdit
	// ActionFinal: FINAL <summary> — the subject declares it is done.
	ActionFinal
	// ActionUnknown: a CALL to a tool the loop does not provide.
	ActionUnknown
)

// editSeparator splits an edit_file argument into its path and its new contents.
const editSeparator = "|||"

// Action is one parsed model turn: either a tool call, a FINAL, or nothing.
type Action struct {
	Kind ActionKind
	// Tool is the raw tool name as the subject wrote it (for ActionUnknown, the
	// unrecognised name).
	Tool string
	// Arg is the single argument for list_files / read_file / run_query, or the
	// path for edit_file.
	Arg string
	// Content is the new file contents for edit_file, empty otherwise.
	Content string
	// Summary is the text after FINAL, empty otherwise.
	Summary string
	// RawLine is the directive line the parse matched, kept for the transcript.
	RawLine string
}

// ParseAction reads one model turn and returns the first directive it finds. A
// turn may open with prose; the parse scans line by line and returns the first
// line that is a CALL or a FINAL, so leading reasoning does not hide the action.
// A turn with neither is ActionNone (a stall).
//
// edit_file is the exception to the line-at-a-time rule: its contents are a
// whole file and span the rest of the message, so once an edit_file directive is
// found the parse takes everything after "|||" through the end of the turn.
func ParseAction(modelText string) Action {
	lines := strings.Split(modelText, "\n")
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		switch {
		case line == "FINAL" || strings.HasPrefix(line, "FINAL "):
			return Action{Kind: ActionFinal, Summary: strings.TrimSpace(strings.TrimPrefix(line, "FINAL")), RawLine: line}
		case strings.HasPrefix(line, "CALL "):
			rest := strings.TrimSpace(strings.TrimPrefix(line, "CALL"))
			return parseCall(rest, line, lines[i+1:])
		}
	}
	return Action{Kind: ActionNone}
}

// parseCall turns the text after "CALL " into a typed Action. rest is the tool
// name plus its (single-line) argument; rawLine is the directive line; tail is
// the lines after it, which an edit_file's contents continue into.
func parseCall(rest, rawLine string, tail []string) Action {
	tool, arg, _ := strings.Cut(rest, " ")
	switch tool {
	case "list_files":
		return Action{Kind: ActionList, Tool: tool, Arg: strings.TrimSpace(arg), RawLine: rawLine}
	case "read_file":
		return Action{Kind: ActionRead, Tool: tool, Arg: strings.TrimSpace(arg), RawLine: rawLine}
	case "run_query":
		return Action{Kind: ActionQuery, Tool: tool, Arg: strings.TrimSpace(arg), RawLine: rawLine}
	case "edit_file":
		// The contents are a whole file: rejoin this line's argument with every
		// line after it, then split off the path at the "|||" separator.
		full := strings.Join(append([]string{arg}, tail...), "\n")
		path, content, found := strings.Cut(full, editSeparator)
		a := Action{Kind: ActionEdit, Tool: tool, Arg: strings.TrimSpace(path), RawLine: rawLine}
		if found {
			// Drop only the single leading space of the " ||| " spelling; the file
			// body itself is the artifact under the bar and is kept verbatim.
			a.Content = strings.TrimPrefix(content, " ")
		}
		return a
	default:
		return Action{Kind: ActionUnknown, Tool: tool, Arg: strings.TrimSpace(arg), RawLine: rawLine}
	}
}
