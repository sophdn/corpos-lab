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

import (
	"encoding/json"
	"strings"
)

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

// ParseAction reads one model turn and returns the first directive it finds. It
// accepts two protocols: the loop's own "CALL <tool> <arg>" text form, and the
// model's native trained tool-call syntax. Qwen3 reverts to its native format
// mid-run — the setup-vs-agentic-loop smoke saw it switch after turn 1 — so a
// loop that only read the CALL form would score a real action as a stall.
// Native forms are tried first; then a line scan for CALL / FINAL; a turn with
// none is ActionNone (a stall).
//
// edit_file is the exception to the line-at-a-time rule: its contents are a
// whole file and span the rest of the message, so once an edit_file directive is
// found the parse takes everything after "|||" through the end of the turn.
func ParseAction(modelText string) Action {
	if a, ok := parseNative(modelText); ok {
		return a
	}
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

// parseNative recognises the model's own trained tool-call syntax, in the two
// shapes Qwen3 emits: a Hermes JSON call inside <tool_call> … </tool_call>, and
// the <function=NAME> … </function> tag form. It returns false when the turn
// carries neither, so the CALL text protocol still gets its chance.
func parseNative(text string) (Action, bool) {
	inner, hasCall := between(text, "<tool_call>", "</tool_call>")
	if hasCall {
		if a, ok := parseJSONCall(inner); ok {
			return a, true
		}
	}
	// The <function=NAME> … </function> form, whether bare or inside a <tool_call>.
	if idx := strings.Index(text, "<function="); idx >= 0 {
		rest := text[idx+len("<function="):]
		name, after, ok := strings.Cut(rest, ">")
		if ok {
			name = strings.TrimSpace(name)
			arg := after
			if end := strings.Index(after, "</function>"); end >= 0 {
				arg = after[:end]
			}
			arg = strings.TrimSpace(arg)
			// Malformed hybrid: Qwen sometimes stuffs the whole CALL text directive
			// into the function name, e.g. <function=CALL list_files tests/> or
			// <function=CALL read_file x</function>. The name is then not a bare
			// tool, so the tag reads as unknown-tool and the real read/edit is lost.
			// Recover the tool and any inline argument from the directive, falling
			// back to the tag's own argument. A well-formed named tool is handled
			// above and never reaches this recovery.
			if !isKnownTool(name) {
				directive := strings.TrimSpace(strings.TrimPrefix(stripTagNoise(name), "CALL"))
				if tool, inlineArg, _ := strings.Cut(directive, " "); isKnownTool(tool) {
					callArg := strings.TrimSpace(inlineArg)
					if callArg == "" {
						callArg = stripTagNoise(arg)
					}
					return namedCall(tool, callArg), true
				}
			}
			return namedCall(name, arg), true
		}
	}
	// The bare form: <tool_call> then "toolname args" on its own line, no JSON and
	// no <function=> wrapper. Qwen3 emits this too (calibration run). Read the
	// first non-empty line as "toolname args", but only when the first token names
	// a real tool — otherwise malformed JSON ("{not json}") would be misread as a
	// bare call rather than falling through.
	if hasCall {
		for _, line := range strings.Split(inner, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			name, arg, _ := strings.Cut(line, " ")
			if isKnownTool(name) {
				return namedCall(name, strings.TrimSpace(arg)), true
			}
			break
		}
	}
	return Action{}, false
}

// stripTagNoise removes the tool-call XML wrapper fragments a model leaves inside
// a malformed <function=…> value, so the CALL directive trapped there can be read.
// It runs only on the recovery path; a well-formed call never reaches it.
func stripTagNoise(s string) string {
	for _, tok := range []string{"</function>", "</function", "</tool_call>", "</tool_call", "<tool_call>", "<function="} {
		s = strings.ReplaceAll(s, tok, " ")
	}
	s = strings.ReplaceAll(s, "<", " ")
	s = strings.ReplaceAll(s, ">", " ")
	return strings.TrimSpace(s)
}

// isKnownTool reports whether name is one of the loop's four tools.
func isKnownTool(name string) bool {
	switch name {
	case "list_files", "read_file", "run_query", "edit_file":
		return true
	default:
		return false
	}
}

// between returns the text between the first open and the next closeTag marker.
func between(s, open, closeTag string) (string, bool) {
	i := strings.Index(s, open)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(open):]
	j := strings.Index(rest, closeTag)
	if j < 0 {
		return strings.TrimSpace(rest), true // unclosed: take the remainder
	}
	return strings.TrimSpace(rest[:j]), true
}

// jsonCall is the Hermes tool-call shape: a name and an arguments object.
type jsonCall struct {
	Name      string `json:"name"`
	Arguments struct {
		Dir     string `json:"dir"`
		Path    string `json:"path"`
		Query   string `json:"query"`
		Content string `json:"content"`
	} `json:"arguments"`
}

// parseJSONCall reads a Hermes JSON tool call and maps it to an Action. The
// argument keys are the ones the preamble names (dir/path/query/content).
func parseJSONCall(inner string) (Action, bool) {
	var c jsonCall
	if err := json.Unmarshal([]byte(inner), &c); err != nil || c.Name == "" {
		return Action{}, false
	}
	switch c.Name {
	case "list_files":
		return Action{Kind: ActionList, Tool: c.Name, Arg: c.Arguments.Dir}, true
	case "read_file":
		return Action{Kind: ActionRead, Tool: c.Name, Arg: c.Arguments.Path}, true
	case "run_query":
		return Action{Kind: ActionQuery, Tool: c.Name, Arg: c.Arguments.Query}, true
	case "edit_file":
		return Action{Kind: ActionEdit, Tool: c.Name, Arg: c.Arguments.Path, Content: c.Arguments.Content}, true
	default:
		return Action{Kind: ActionUnknown, Tool: c.Name}, true
	}
}

// namedCall maps a tool name plus a freeform argument (the <function=NAME> form)
// to an Action. For edit_file the freeform argument is split at "|||" as in the
// CALL form; failing that, the first line is the path and the rest the contents.
func namedCall(name, arg string) Action {
	switch name {
	case "list_files":
		return Action{Kind: ActionList, Tool: name, Arg: arg}
	case "read_file":
		return Action{Kind: ActionRead, Tool: name, Arg: arg}
	case "run_query":
		return Action{Kind: ActionQuery, Tool: name, Arg: arg}
	case "edit_file":
		if path, content, found := strings.Cut(arg, editSeparator); found {
			return Action{Kind: ActionEdit, Tool: name, Arg: strings.TrimSpace(path), Content: strings.TrimPrefix(content, " ")}
		}
		path, content, _ := strings.Cut(arg, "\n")
		return Action{Kind: ActionEdit, Tool: name, Arg: strings.TrimSpace(path), Content: content}
	default:
		return Action{Kind: ActionUnknown, Tool: name, Arg: arg}
	}
}
