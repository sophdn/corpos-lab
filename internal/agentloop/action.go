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
//
// Why the harness parses tool calls from text, not llama-server's structured
// tool calling (decided 2026-10-09, task agentloop-cells-and-parser-corpus):
// structured calls need /v1/chat/completions with the model's jinja template and
// a tools schema, so the server would inject template framing and a tool
// description into the subject's prompt, and its own parser would sit between
// the model and the harness. That breaks the apparatus rule that the subject runs
// on the least-opinionated endpoint — raw /completion with a study-declared,
// recorded prompt — and a server parser would silently drop or repair the
// malformed calls this package records as data (see ParseOutcome). The cost of
// text parsing is new dialects; the testdata/dialects corpus and the per-turn
// parse outcomes are how that cost is caught and paid.
package agentloop

import (
	"encoding/json"
	"regexp"
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
	a := parseAction(modelText)
	if a.Kind == ActionEdit {
		a.Content = stripTrailingNoise(a.Content)
	}
	return a
}

// ParseActions reads one model turn and returns every directive it carries, in
// order. It exists because the two protocols differ in how many calls one turn
// holds:
//
//   - The loop's "CALL <tool> <arg>" text form is one tool per turn by
//     construction (the preamble says so), so a CALL or FINAL turn yields a
//     single directive. A second CALL line is an echo the harness does not run.
//   - The model's native tool-call XML is its trained BATCH form: Qwen writes
//     several <tool_call> blocks in one turn ("let me check … in parallel"), so a
//     native turn yields EVERY call it carried.
//
// Executing all native calls is the fix for bug 1428. The old loop parsed one
// call per turn, so a multi-call native turn left the model's other reads
// unanswered; the model re-emitted the same batch and re-read one file until the
// step cap, writing no decision. ParseActions returns at least one Action, so a
// turn with neither a native call nor a CALL/FINAL line is a single ActionNone.
//
// A CALL or FINAL line written BEFORE the first native marker wins: the model
// took the loop's own protocol first and the native text after it is trailing
// echo. Without this a turn of "CALL edit_file … ||| … FINAL …" followed by a
// native read ran only the read, so a reached decision was lost (2026-10-09).
func ParseActions(modelText string) []Action {
	if m := firstNativeMarker(modelText); m > 0 {
		if a := ParseAction(modelText[:m]); a.Kind != ActionNone {
			return []Action{a}
		}
	}
	if acts := parseNativeAll(modelText); len(acts) > 0 {
		// A native FINAL first ends the turn. A FINAL after real calls is dropped:
		// the subject has not yet seen what those calls return.
		if acts[0].Kind == ActionFinal {
			return acts[:1]
		}
		calls := acts[:0]
		for _, a := range acts {
			if a.Kind == ActionFinal {
				continue
			}
			if a.Kind == ActionEdit {
				a.Content = stripTrailingNoise(a.Content)
			}
			calls = append(calls, a)
		}
		return calls
	}
	return []Action{ParseAction(modelText)}
}

// firstNativeMarker returns the index of the first native tool-call marker
// (<tool_call> or <function=) in text, or -1 when there is none.
func firstNativeMarker(text string) int {
	m := -1
	for _, marker := range []string{"<tool_call>", "<function="} {
		if i := strings.Index(text, marker); i >= 0 && (m < 0 || i < m) {
			m = i
		}
	}
	return m
}

// parseNativeAll extracts every native tool-call the turn carries, in order. It
// splits the turn into native blocks — each <tool_call>…</tool_call> region, or
// each bare <function=…> region when no <tool_call> wrapper is present — and runs
// the single-block parseNative on each, keeping the calls it recognises. A block
// parseNative cannot read (e.g. "{not json}") is skipped; when no block parses it
// returns nil, so ParseActions falls back to the CALL/FINAL line scan and a turn
// such as "<tool_call>{not json}</tool_call>\nFINAL done" still finds its FINAL.
func parseNativeAll(text string) []Action {
	var acts []Action
	for _, b := range nativeBlocks(text) {
		if a, ok := parseNative(b); ok {
			acts = append(acts, a)
		}
	}
	return acts
}

// nativeBlocks splits text into one substring per native tool-call, each keeping
// its opening marker so parseNative reads it exactly as it reads a whole turn. It
// splits on <tool_call> when that wrapper is present (Qwen's usual shape), else on
// a bare <function=. Text before the first marker is dropped. A turn with neither
// marker yields no blocks.
func nativeBlocks(text string) []string {
	var delim string
	switch {
	case strings.Contains(text, "<tool_call>"):
		delim = "<tool_call>"
	case strings.Contains(text, "<function="):
		delim = "<function="
	default:
		return nil
	}
	rest := text
	if i := strings.Index(rest, delim); i > 0 {
		rest = rest[i:]
	}
	var blocks []string
	for {
		next := strings.Index(rest[len(delim):], delim)
		if next < 0 {
			blocks = append(blocks, rest)
			return blocks
		}
		cut := len(delim) + next
		blocks = append(blocks, rest[:cut])
		rest = rest[cut:]
	}
}

// parseAction is the raw parse; ParseAction wraps it to clean an edit body.
func parseAction(modelText string) Action {
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

// stripTrailingNoise removes the protocol text the subject sometimes leaves at
// the end of an edit_file body: its own FINAL line, a stray CALL directive, a
// </think> tag, or the closing tags of a native call. Each such text bled into
// the file the subject wrote (bug 1441). Only whole trailing lines are dropped,
// repeatedly from the end, so the same text inside the file body is untouched,
// and a body with no trailing noise is returned as-is.
func stripTrailingNoise(content string) string {
	lines := strings.Split(content, "\n")
	last := len(lines) - 1
	dropped := false
	for last >= 0 {
		t := strings.TrimSpace(lines[last])
		if t != "" && !isTrailingNoise(t) {
			break
		}
		dropped = dropped || t != ""
		last--
	}
	if !dropped {
		return content
	}
	return strings.TrimRight(strings.Join(lines[:last+1], "\n"), "\n")
}

// isTrailingNoise reports whether one trimmed line is protocol text, not file
// content.
func isTrailingNoise(line string) bool {
	switch line {
	case "FINAL", "</think>", "<think>", "</tool_call>", "<tool_call>", "</function>", paramTagClose:
		return true
	}
	return strings.HasPrefix(line, "FINAL ") || strings.HasPrefix(line, "CALL ")
}

// parseCall turns the text after "CALL " into a typed Action. rest is the tool
// name plus its (single-line) argument; rawLine is the directive line; tail is
// the lines after it, which an edit_file's contents continue into.
func parseCall(rest, rawLine string, tail []string) Action {
	tool, arg, _ := strings.Cut(rest, " ")
	switch tool {
	case "FINAL":
		// "CALL FINAL <summary>": the subject finished but wrapped FINAL in the
		// call form. It is a FINAL, not a call to an unknown "FINAL" tool.
		return Action{Kind: ActionFinal, Summary: strings.TrimSpace(arg), RawLine: rawLine}
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
	if a, ok := parseFunctionTag(text); ok {
		return a, true
	}
	if hasCall {
		return parseBareCall(inner)
	}
	return Action{}, false
}

// parseFunctionTag reads the first <function=NAME> … </function> tag in text.
func parseFunctionTag(text string) (Action, bool) {
	idx := strings.Index(text, "<function=")
	if idx < 0 {
		return Action{}, false
	}
	rest := text[idx+len("<function="):]
	name, after, ok := strings.Cut(rest, ">")
	if !ok {
		return Action{}, false
	}
	name, after = repairOverCapturedName(invokeName(strings.TrimSpace(name)), after)
	arg := after
	if end := strings.Index(after, "</function>"); end >= 0 {
		arg = after[:end]
	}
	arg = strings.TrimSpace(arg)
	// <function=FINAL>: the subject declared it was done in the native
	// form. Without this it ran as an unknown "FINAL" tool and the loop
	// went on past the subject's own end.
	if strings.EqualFold(name, "FINAL") {
		return Action{Kind: ActionFinal, Summary: stripTagNoise(arg)}, true
	}
	if !isKnownTool(name) {
		if a, ok := recoverHybridCall(name, arg); ok {
			return a, true
		}
	}
	return namedCall(name, arg), true
}

// repairOverCapturedName handles a missing close-'>' on the function name: Qwen
// writes "<function=edit_file\n<parameter=path>…", so the Cut on '>' runs on the
// <parameter=…>'s '>' and over-captures into the name. The real name is the first
// token; re-insert the consumed '>' so the parameter tags stay intact for
// parseParamTags. Only triggers when the capture swallowed a tag fragment, so the
// "<function=CALL list_files tests/>" hybrid (no "<parameter=") still reaches the
// recovery in parseFunctionTag unchanged.
func repairOverCapturedName(name, after string) (string, string) {
	if !strings.Contains(name, paramTagPrefix) {
		return name, after
	}
	firstTok := name
	if cut := strings.IndexAny(name, " \t\r\n"); cut >= 0 {
		firstTok = name[:cut]
	}
	if !isKnownTool(firstTok) {
		return name, after
	}
	return firstTok, strings.TrimPrefix(name, firstTok) + ">" + after
}

// recoverHybridCall recovers a malformed hybrid: Qwen sometimes stuffs the whole
// CALL text directive into the function name, e.g. <function=CALL list_files
// tests/> or <function=CALL read_file x</function>. The name is then not a bare
// tool, so the tag reads as unknown-tool and the real read/edit is lost. Recover
// the tool and any inline argument from the directive, falling back to the tag's
// own argument. A well-formed named tool never reaches this recovery.
func recoverHybridCall(name, arg string) (Action, bool) {
	directive := strings.TrimSpace(strings.TrimPrefix(stripTagNoise(name), "CALL"))
	tool, inlineArg, _ := strings.Cut(directive, " ")
	if !isKnownTool(tool) {
		return Action{}, false
	}
	callArg := strings.TrimSpace(inlineArg)
	if callArg == "" {
		callArg = stripTagNoise(arg)
	}
	return namedCall(tool, callArg), true
}

// parseBareCall reads the bare form: <tool_call> then "toolname args" on its own
// line, no JSON and no <function=> wrapper. Qwen3 emits this too (calibration
// run). Read the first non-empty line as "toolname args", but only when the first
// token names a real tool — otherwise malformed JSON ("{not json}") would be
// misread as a bare call rather than falling through.
func parseBareCall(inner string) (Action, bool) {
	for _, line := range strings.Split(inner, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, arg, _ := strings.Cut(line, " ")
		if name == "edit_file" {
			// An edit's contents span the lines after the directive. Taking
			// only its first line cut a multi-line file to one line and wiped
			// the rest of it (bug 1441).
			_, body, _ := strings.Cut(inner, "edit_file")
			return namedCall(name, stripTrailingNoise(strings.TrimSpace(body))), true
		}
		if isKnownTool(name) {
			return namedCall(name, strings.TrimSpace(arg)), true
		}
		break
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

// jsonCall is the Hermes tool-call shape: a name and an arguments object. The
// argument object carries the preamble's key names plus the synonyms Qwen3 emits
// (filename/file for path, directory for dir), resolved the same way the
// <parameter=…> tag path resolves them.
type jsonCall struct {
	Name      string `json:"name"`
	Arguments struct {
		Dir         string `json:"dir"`
		Directory   string `json:"directory"`
		Path        string `json:"path"`
		Filename    string `json:"filename"`
		File        string `json:"file"`
		Query       string `json:"query"`
		Content     string `json:"content"`
		Contents    string `json:"contents"`
		NewContent  string `json:"new_content"`
		NewContents string `json:"new_contents"`
	} `json:"arguments"`
}

// parseJSONCall reads a Hermes JSON tool call and maps it to an Action. The
// argument keys are the ones the preamble names (dir/path/query/content), each
// resolved against the synonym set so a model that writes "filename" for a read
// path still lands.
func parseJSONCall(inner string) (Action, bool) {
	var c jsonCall
	if err := json.Unmarshal([]byte(inner), &c); err != nil || c.Name == "" {
		return Action{}, false
	}
	first := func(vals ...string) string {
		for _, v := range vals {
			if v != "" {
				return v
			}
		}
		return ""
	}
	a := c.Arguments
	switch c.Name {
	case "list_files":
		return Action{Kind: ActionList, Tool: c.Name, Arg: first(a.Dir, a.Directory, a.Path)}, true
	case "read_file":
		return Action{Kind: ActionRead, Tool: c.Name, Arg: first(a.Path, a.Filename, a.File)}, true
	case "run_query":
		return Action{Kind: ActionQuery, Tool: c.Name, Arg: a.Query}, true
	case "edit_file":
		return Action{Kind: ActionEdit, Tool: c.Name, Arg: first(a.Path, a.Filename, a.File), Content: first(a.Content, a.Contents, a.NewContent, a.NewContents)}, true
	default:
		return Action{Kind: ActionUnknown, Tool: c.Name}, true
	}
}

// namedCall maps a tool name plus a freeform argument (the <function=NAME> form)
// to an Action. Qwen3 often fills the body with <parameter=KEY>VALUE</parameter>
// sub-tags instead of a bare argument; when it does, the parameters decide the
// action. Otherwise, for edit_file the freeform argument is split at "|||" as in
// the CALL form; failing that, the first line is the path and the rest the
// contents.
func namedCall(name, arg string) Action {
	arg = normalizeParamAttr(arg)
	if params, ok := parseParamTags(arg); ok {
		return actionFromParams(name, arg, params)
	}
	if vals, ok := parseBareParams(arg); ok {
		return actionFromPositional(name, vals)
	}
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

// bareParamOpen is the unnamed parameter tag Qwen3 sometimes writes, with no key:
// <function=list_files><parameter>.</parameter>.
const bareParamOpen = "<parameter>"

// parseBareParams pulls the values of unnamed <parameter>VALUE</parameter> tags
// out of a <function=NAME> body, in order, trimming whitespace. It reports false
// when the body has none. Without it the whole tag soup ran as the argument, so
// list_files listed "<parameter> . </parameter>" as an empty directory and the
// subject re-listed until the cell went NULL (bug 1441). A value whose closing tag
// is missing takes the rest of the body.
func parseBareParams(body string) ([]string, bool) {
	var vals []string
	rest := body
	for {
		open := strings.Index(rest, bareParamOpen)
		if open < 0 {
			break
		}
		rest = rest[open+len(bareParamOpen):]
		value, tail, closed := strings.Cut(rest, paramTagClose)
		vals = append(vals, stripTrailingNoise(strings.TrimSpace(value)))
		if !closed {
			break
		}
		rest = tail
	}
	return vals, len(vals) > 0
}

// actionFromPositional builds an Action from unnamed parameter values. A
// single-argument tool takes the first value. edit_file takes the path from the
// first and the contents from the second, or splits a lone value at "|||".
func actionFromPositional(name string, vals []string) Action {
	if name == "edit_file" {
		if len(vals) > 1 {
			return Action{Kind: ActionEdit, Tool: name, Arg: vals[0], Content: vals[1]}
		}
		path, content, _ := strings.Cut(vals[0], editSeparator)
		return Action{Kind: ActionEdit, Tool: name, Arg: strings.TrimSpace(path), Content: strings.TrimPrefix(content, " ")}
	}
	return namedCall(name, vals[0])
}

// invokeAttr matches the <function=invoke name="TOOL"> form, where Qwen3 puts the
// real tool name in an attribute of a generic "invoke".
var invokeAttr = regexp.MustCompile(`^invoke\s+name\s*=\s*["']?([A-Za-z_]+)["']?$`)

// invokeName unwraps the invoke form to its tool name. Any other name is returned
// unchanged.
func invokeName(name string) string {
	if m := invokeAttr.FindStringSubmatch(name); m != nil {
		return m[1]
	}
	return name
}

// Qwen3's sub-tag brackets, e.g. <parameter=path>board/tasks/x.md</parameter>.
const (
	paramTagPrefix = "<parameter="
	paramTagClose  = "</parameter>"
)

// paramAttrTag matches the attribute form of a parameter open tag,
// <parameter name="KEY">, which Qwen3 emits instead of <parameter=KEY>.
var paramAttrTag = regexp.MustCompile(`<parameter\s+name\s*=\s*["']([^"']*)["']\s*>`)

// normalizeParamAttr rewrites every <parameter name="KEY"> open tag to the
// <parameter=KEY> form, so one parser handles both. Without it the attribute form
// carries no "<parameter=" at all and the call ran with the raw tag text as its
// argument, so the cell stalled on a repeated no-op call.
func normalizeParamAttr(body string) string {
	return paramAttrTag.ReplaceAllString(body, paramTagPrefix+"$1>")
}

// cutParamKey splits the text after a "<parameter=" prefix into the key and the
// text after the key. The key ends at the first ">" or newline. Qwen3 sometimes
// drops the ">" and writes the key alone on its line,
// <parameter=thumbsvc/cache.py\n</parameter>; ending the key at the newline keeps
// that key clean instead of running it into the closing tag. tagClosed reports
// whether the key ended at ">" (a normal open tag). ok is false when neither
// delimiter follows.
func cutParamKey(rest string) (key, after string, tagClosed, ok bool) {
	i := strings.IndexAny(rest, ">\n")
	if i < 0 {
		return "", "", false, false
	}
	key = strings.TrimSpace(rest[:i])
	// <parameter=config/services.yaml</parameter>: the open tag lost its ">" and
	// the close tag follows on the same line, so the first ">" is the close tag's.
	// The key is the text before "</parameter"; it carries no value (bug 1441).
	if k, _, found := strings.Cut(key, "</parameter"); found {
		return strings.TrimSpace(k), rest[i+1:], false, true
	}
	return key, rest[i+1:], rest[i] == '>', true
}

// parseParamTags pulls every <parameter=KEY>VALUE</parameter> block out of a
// <function=NAME> body into a KEY→VALUE map, trimming surrounding whitespace and
// newlines from each value. It reports false when the body carries no such block,
// so the freeform-arg path still handles the bare <function=NAME>arg</function>
// form. A parameter whose closing tag is missing takes the rest of the body as its
// value; a malformed open tag with no ">" after the key ends the scan.
func parseParamTags(body string) (map[string]string, bool) {
	if !strings.Contains(body, paramTagPrefix) {
		return nil, false
	}
	params := map[string]string{}
	rest := body
	for {
		open := strings.Index(rest, paramTagPrefix)
		if open < 0 {
			break
		}
		rest = rest[open+len(paramTagPrefix):]
		key, after, tagClosed, ok := cutParamKey(rest)
		if !ok {
			break
		}
		if !tagClosed {
			// Open tag lost its ">": the key stands alone with no value.
			params[key] = ""
			rest = after
			continue
		}
		value, tail, closed := strings.Cut(after, paramTagClose)
		params[strings.TrimSpace(key)] = strings.TrimSpace(value)
		if !closed {
			break
		}
		rest = tail
	}
	if len(params) == 0 {
		return nil, false
	}
	return params, true
}

// actionFromParams builds an Action from the parameters parsed out of a
// <function=NAME> body. Keys map the same way parseJSONCall maps the JSON argument
// keys: dir→list_files, path→read_file/edit_file, query→run_query,
// content→edit_file contents. For edit_file without a content parameter it falls
// back to the "|||" separator in the raw body, so the hybrid where the model gives
// a <parameter=path> sub-tag and then puts the file after "|||" still lands.
//
// Qwen3 does not hold the preamble's exact parameter names: it writes read_file's
// path as <parameter=filename>, list_files' dir as <parameter=path>, and wraps the
// real tool in a generic <function=call> whose <parameter=tool> names it. Each key
// is therefore resolved against a synonym set, and a generic wrapper is dispatched
// on its inner tool name. This recovers only calls that are otherwise lost to an
// empty argument; a correctly-named call resolves to the same value it always did.
func actionFromParams(name, body string, params map[string]string) Action {
	// A value that still holds a tag fragment is garbage from an unclosed
	// parameter tag swallowing the next one; blank it so the JSON overlay and the
	// synonym resolution can supply the real value.
	for k, v := range params {
		if strings.Contains(v, paramTagPrefix) || strings.Contains(v, "<function=") {
			params[k] = ""
		}
	}
	// Args carried as a JSON object — under <parameter=arguments> or, when the
	// param tags are mangled, embedded loose in the body (the hybrid param-tag +
	// Hermes-JSON shape Qwen emits). Lift its string fields into the param map so
	// the resolution below finds path/dir/content as usual.
	overlayArgsJSON(params, body)
	// Generic wrapper: <function=call> (or any non-tool name) with the real tool
	// in a "tool"/"name" parameter. Dispatch on that inner tool when it is ours.
	if !isKnownTool(name) {
		if inner := firstParam(params, "tool", "tool_name", "name", "action", "function"); isKnownTool(inner) {
			name = inner
		}
	}
	// "argument"/"arg" are the generic single-argument keys the <function=CALL>
	// wrapper uses; they sit last so a specific key always wins.
	switch name {
	case "list_files":
		return Action{Kind: ActionList, Tool: name, Arg: orPathLikeKey(firstParam(params, "dir", "directory", "path", "argument", "arg"), body)}
	case "read_file":
		return Action{Kind: ActionRead, Tool: name, Arg: orPathLikeKey(firstParam(params, "path", "filename", "file", "filepath", "argument", "arg"), body)}
	case "run_query":
		return Action{Kind: ActionQuery, Tool: name, Arg: firstParam(params, "query", "sql", "q", "argument", "arg")}
	case "edit_file":
		a := Action{Kind: ActionEdit, Tool: name, Arg: firstParam(params, "path", "filename", "file", "filepath")}
		if a.Arg == "" {
			// Bug 1430: Qwen writes the path as the FIRST parameter's NAME, not a
			// path value: <function=edit_file><parameter=REVIEW.md>…</parameter>
			// <parameter=new_content>…</parameter>. No path synonym matches, so the
			// write landed with an empty path and the decision scored NULL. Recover
			// the path from the first parameter key that names a file.
			a.Arg = pathLikeParamKey(body)
		}
		if content := firstParam(params, "content", "contents", "new_content", "new_contents", "text", "body"); content != "" {
			a.Content = content
		} else if _, after, found := strings.Cut(body, editSeparator); found {
			a.Content = strings.TrimSpace(after)
		}
		return a
	default:
		return Action{Kind: ActionUnknown, Tool: name, Arg: body}
	}
}

// firstParam returns the first non-empty value among keys, so a tool's argument
// is recovered whichever synonym the model used for the parameter name. Values in
// params are already whitespace-trimmed by parseParamTags.
func firstParam(params map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := params[k]; v != "" {
			return v
		}
	}
	return ""
}

// orPathLikeKey returns arg, or when it is empty the first path-like parameter
// key in body. read_file and list_files need this for the same reason edit_file
// does (bug 1430): Qwen3 writes the path as the parameter NAME, e.g.
// <parameter=thumbsvc/cache.py>, so no path synonym holds a value.
func orPathLikeKey(arg, body string) string {
	if arg != "" {
		return arg
	}
	return pathLikeParamKey(body)
}

// pathLikeParamKey returns the first <parameter=KEY> in body order whose KEY
// names a file — it holds a "/" or a ".". This is how Qwen encodes an edit_file
// path when it writes the path as the parameter NAME rather than a path value:
// <function=edit_file><parameter=REVIEW.md>…</parameter>… (bug 1430). It scans
// body rather than the params map so the FIRST key wins deterministically; a map
// has no order. None of the loop's parameter keywords (path, content, dir, …)
// hold a "/" or ".", so a key that does is unambiguously a path, not a keyword.
// Returns "" when no parameter key looks like a path.
func pathLikeParamKey(body string) string {
	rest := body
	for {
		open := strings.Index(rest, paramTagPrefix)
		if open < 0 {
			return ""
		}
		rest = rest[open+len(paramTagPrefix):]
		key, _, _, ok := cutParamKey(rest)
		if !ok {
			return ""
		}
		// A key with whitespace or "<" is a malformed tag run into its value or
		// the next tag, not a path.
		if strings.ContainsAny(key, "/.") && !strings.ContainsAny(key, " \t<") {
			return key
		}
	}
}

// overlayArgsJSON lifts the string fields of a JSON object carried under an
// "arguments" parameter into params, filling only keys that are empty or absent.
// Qwen3 sometimes writes <parameter=arguments>{"path":…,"contents":…}</parameter>
// instead of one <parameter=…> per field; without this the real path/content are
// trapped in the JSON blob and the call lands with an empty argument. A missing or
// non-object "arguments" value is left untouched.
func overlayArgsJSON(params map[string]string, body string) {
	// Prefer an explicit arguments parameter; fall back to the first JSON object
	// embedded loose in the body (what remains when the param tags are mangled).
	candidates := []string{strings.TrimSpace(params["arguments"])}
	if i := strings.IndexByte(body, '{'); i >= 0 {
		candidates = append(candidates, body[i:])
	}
	for _, c := range candidates {
		if !strings.HasPrefix(strings.TrimSpace(c), "{") {
			continue
		}
		// Decode one value and ignore trailing junk, so "{…}\n</function>" parses.
		var obj map[string]any
		if json.NewDecoder(strings.NewReader(c)).Decode(&obj) != nil {
			continue
		}
		for k, v := range obj {
			if s, ok := v.(string); ok && params[k] == "" {
				params[k] = s
			}
		}
		return
	}
}
