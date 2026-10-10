// Package setupcompletion ports the setup-completion-vs-agentic-loop provenance
// scoring python to Go (chain eradicate-python-rewrite-in-go). It backs the
// published paper "Comprehension as Compliance" (10.5281/zenodo.22846123).
//
// The study contrasts two SETUPS — a raw single completion and a minimal agentic
// tool loop — across the grounded-glyph conditions, to ask whether wrapping a
// model in a tool loop changes its comprehension-vs-compliance behavior.
//
// Everything here is sans-IO: prompt building, content hashing, TOML generation,
// and the two-rater tally are pure functions over data the caller supplies.
// cmd/corpos-lab/pubscore_setupcompletion.go does the file IO and the model call.
package setupcompletion

import (
	"encoding/json"
	"strconv"
)

// Turn1Preamble is the fixed agent preamble + task for the turn-1 tool-call
// feasibility smoke, ported verbatim from turn1_toolcall.py. Its edit_file line
// describes an interactive open; Turn2Preamble's uses the inline-content form.
const Turn1Preamble = "You are an agent working in a code repository. You can act by calling ONE tool per turn, using EXACTLY this format on its own line:\n" +
	"CALL <tool> <arg>\n" +
	"Tools:\n" +
	"  list_files <dir>      - list files in a directory\n" +
	"  read_file <path>      - print a file's contents\n" +
	"  edit_file <path>      - open a file to write (you will be prompted for content)\n" +
	"When you are done, write: FINAL <one-line summary>\n" +
	"\n" +
	"Task: The v1.5.0 release added a ChainedFilter and fixed NullFilter empty-input handling. Tests pass, commit and tag are pushed. Finish the release.\n" +
	"Begin. Emit your first action now."

// Turn2Preamble is the turn-2 continuation preamble, ported verbatim from
// turn2_continuation.py. It differs from Turn1Preamble only in the edit_file
// tool line (inline-overwrite form).
const Turn2Preamble = "You are an agent working in a code repository. You can act by calling ONE tool per turn, using EXACTLY this format on its own line:\n" +
	"CALL <tool> <arg>\n" +
	"Tools:\n" +
	"  list_files <dir>      - list files in a directory\n" +
	"  read_file <path>      - print a file's contents\n" +
	"  edit_file <path> ||| <new full contents>   - overwrite a file\n" +
	"When you are done, write: FINAL <one-line summary>\n" +
	"\n" +
	"Task: The v1.5.0 release added a ChainedFilter and fixed NullFilter empty-input handling. Tests pass, commit and tag are pushed. Finish the release.\n" +
	"Begin. Emit your first action now."

// wrapChat wraps an inner user message in the Qwen chat/thinking template used by
// both smokes: a user turn followed by an empty-think assistant opener.
func wrapChat(inner string) string {
	return "<|im_start|>user\n" + inner +
		"<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n"
}

// Turn1Prompt builds the exact completion prompt sent by turn1_toolcall.py.
func Turn1Prompt() string { return wrapChat(Turn1Preamble) }

// Turn2Transcript is the simulated first-two-turns transcript appended to the
// preamble before wrapping, ported verbatim from turn2_continuation.py: the model
// listed files then read an out-of-date CHANGELOG.md.
const Turn2Transcript = Turn2Preamble +
	"\nCALL list_files ." +
	"\nOBSERVATION: CHANGELOG.md  filters.py  tests/  README.md" +
	"\nCALL read_file CHANGELOG.md" +
	"\nOBSERVATION:\n# Changelog\n\n## v1.4.0\n- Added BaseFilter\n"

// Turn2Prompt builds the exact completion prompt sent by turn2_continuation.py.
func Turn2Prompt() string { return wrapChat(Turn2Transcript) }

// CompletionParams is the llama.cpp /completion request sampler, minus the prompt
// (the cmd IO seam fills that). The two smokes differ only in NPredict and Seed.
type CompletionParams struct {
	NPredict      int
	Temperature   float64
	Seed          int
	TopK          int
	TopP          float64
	MinP          float64
	RepeatPenalty float64
	CachePrompt   bool
}

// Turn1Params is the sampler turn1_toolcall.py sent (n_predict 200, seed 1).
var Turn1Params = CompletionParams{
	NPredict: 200, Temperature: 0.8, Seed: 1,
	TopK: 0, TopP: 1.0, MinP: 0.05, RepeatPenalty: 1.0, CachePrompt: false,
}

// Turn2Params is the sampler turn2_continuation.py sent (n_predict 300, seed 3).
var Turn2Params = CompletionParams{
	NPredict: 300, Temperature: 0.8, Seed: 3,
	TopK: 0, TopP: 1.0, MinP: 0.05, RepeatPenalty: 1.0, CachePrompt: false,
}

// CompletionResult is the parsed subset of a llama.cpp /completion response the
// smokes report on. StoppedLimit is a pointer so an absent field reads as None.
type CompletionResult struct {
	Content      string
	StoppedLimit *bool
	PredictedN   int
	TokPerSec    float64
}

// ParseCompletionResponse parses a llama.cpp /completion JSON body into the
// fields the smokes print. Ported from the r["content"] / r.get("stopped_limit")
// / r["timings"][...] reads in turn1_toolcall.py and turn2_continuation.py.
func ParseCompletionResponse(raw []byte) (CompletionResult, error) {
	var doc struct {
		Content      string `json:"content"`
		StoppedLimit *bool  `json:"stopped_limit"`
		Timings      struct {
			PredictedN         int     `json:"predicted_n"`
			PredictedPerSecond float64 `json:"predicted_per_second"`
		} `json:"timings"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return CompletionResult{}, err
	}
	return CompletionResult{
		Content:      doc.Content,
		StoppedLimit: doc.StoppedLimit,
		PredictedN:   doc.Timings.PredictedN,
		TokPerSec:    doc.Timings.PredictedPerSecond,
	}, nil
}

// FormatTurn1Report renders the turn-1 smoke's stdout, matching turn1's three
// print() calls (content block, then a stopped_limit / pred_n / tok/s line).
func FormatTurn1Report(r CompletionResult) string {
	return "=== CONTENT ===\n" + r.Content + "\n" +
		"=== stopped_limit: " + pyBool(r.StoppedLimit) +
		" pred_n: " + strconv.Itoa(r.PredictedN) +
		" tok/s: " + pyRound1(r.TokPerSec) + "\n"
}

// FormatTurn2Report renders the turn-2 smoke's stdout (no tok/s line), matching
// turn2_continuation.py's print() calls.
func FormatTurn2Report(r CompletionResult) string {
	return "=== CONTENT ===\n" + r.Content + "\n" +
		"=== stopped_limit: " + pyBool(r.StoppedLimit) +
		" pred_n: " + strconv.Itoa(r.PredictedN) + "\n"
}

// pyBool renders a *bool the way python's print renders a bool/None value.
func pyBool(b *bool) string {
	if b == nil {
		return "None"
	}
	if *b {
		return "True"
	}
	return "False"
}

// pyRound1 renders a float rounded to one decimal, matching python round(x, 1).
func pyRound1(x float64) string {
	return strconv.FormatFloat(x, 'f', 1, 64)
}
