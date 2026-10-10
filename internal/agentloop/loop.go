package agentloop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"corpos-lab/internal/model"
)

// Default loop bounds. A turn is one action, so the token cap per call is small;
// the step cap bounds the whole run. Both are overridable per study.
const (
	// DefaultStepCap bounds the number of turns before the loop gives up.
	DefaultStepCap = 10
	// DefaultCallTokens caps one turn's generation. It must fit the single
	// largest action a turn can legitimately produce — which is an edit_file
	// writing a full multi-field prose record, not a one-line tool call.
	//
	// It was 256 (inherited from the single-turn felt-completion-tail-drop
	// study). That chopped a multi-field Item-12 record mid-write and left the
	// sandbox holding a partial record, which a deterministic end-state scorer
	// read as a false early-exit fire (bug 1425). 768 fits a normal multi-field
	// record with headroom and was the value that gave complete records and
	// 0/26 fires when minimum-viable-step-exit was re-smoked. It stays BOUNDED
	// on purpose: call_tokens also caps runaway per-turn generation, and the
	// step cap bounds the whole run — a longer-record scenario raises its own
	// study's call_tokens rather than this default growing without limit. When a
	// turn does bind this cap, the loop now records it per turn (TruncatedEdit),
	// so a chopped edit is a visible, scoreable-apart fact rather than a silent
	// corruption.
	DefaultCallTokens = 768
)

// defaultStop halts a turn the moment the subject starts writing its own
// OBSERVATION line, so the harness — not the model — supplies the observation
// (the feasibility smoke showed the model fabricating observations otherwise).
//
// It deliberately does NOT stop on a CALL. An earlier version stopped on "\nCALL"
// to block a second action, but that also truncated the FIRST call whenever the
// subject wrote a sentence of prose before it — the calibration run then scored a
// real tool call as a stall. A CALL-form turn runs its first directive and extra
// CALL lines are echoes; a native tool-call turn runs every call it carries
// (ParseActions). Either way the harness injects the real observations, so text
// around the calls is harmless and needs no stop. Fabrication is the only thing
// worth halting.
var defaultStop = []string{"\nOBSERVATION"}

// Config tunes one loop run.
type Config struct {
	// StepCap bounds the turns; <= 0 uses DefaultStepCap.
	StepCap int
	// CallTokens caps each turn's generation; <= 0 uses DefaultCallTokens.
	CallTokens int
	// Stop overrides the tool-call-boundary stop list; nil uses defaultStop.
	Stop []string
}

func (c Config) stepCap() int {
	if c.StepCap > 0 {
		return c.StepCap
	}
	return DefaultStepCap
}

func (c Config) callTokens() int {
	if c.CallTokens > 0 {
		return c.CallTokens
	}
	return DefaultCallTokens
}

func (c Config) stop() []string {
	if len(c.Stop) > 0 {
		return c.Stop
	}
	return defaultStop
}

// EffectiveStop returns the stop list this config resolves to — the study's
// override when set, else the loop's structural default. Exposed so the runner
// records the stop the loop ACTUALLY sent rather than an invisible constant:
// the OBSERVATION boundary is a real generation parameter (record-what-ran), and
// the loop and the run record must agree on it from one source.
func (c Config) EffectiveStop() []string { return c.stop() }

// Terminal is why a loop run ended.
type Terminal string

const (
	// TerminalFinal: the subject wrote FINAL.
	TerminalFinal Terminal = "final"
	// TerminalStall: a turn named no tool and no FINAL — recognition may be
	// present, but the subject did not act. The loop's analysis-mode signal.
	TerminalStall Terminal = "stall"
	// TerminalCap: the step cap was hit before a FINAL. Recorded so a capped run
	// is a visible fact for scoring, not mistaken for a stall.
	TerminalCap Terminal = "cap"
	// TerminalLoop: the subject repeated the same tool call, with the same
	// observation, repeatedCallLimit times inside a short window without progress.
	// The loop breaks on this unproductive repeat rather than burning to the step
	// cap. It is a distinct, scoreable signal: a capped run tried many different
	// things, a looped run was stuck on one. (bug 1428: Qwen re-read one file 10
	// times. bug 1430: Qwen alternated near-identical native-XML edits.)
	TerminalLoop Terminal = "loop"
	// TerminalContext: the server refused the next request because the prompt
	// exceeded the model context (n_ctx). The cell ends here with its transcript
	// and sandbox kept. It used to be a CELL ERROR with neither (bug 1442).
	TerminalContext Terminal = "context"
)

// repeatedCallLimit is how many times one executed-call signature may recur
// inside the recent window before the loop breaks with TerminalLoop. Three
// leaves room for a model that repeats once and then recovers, while still
// cutting a genuine repeat far short of a 16-turn cap.
const repeatedCallLimit = 3

// repeatWindow is how many recent turns the loop watches for a recurring
// signature. The check counts a signature's occurrences inside this window, not
// only consecutive repeats: Qwen's native-XML edit loop ALTERNATES its arguments
// (bug 1430), so the same unproductive call recurs every other turn rather than
// back-to-back. A consecutive-only check missed it and the transcript grew until
// it overflowed n_ctx (16385 > 16384) and the server returned 400. Eight turns
// holds three occurrences of a two- or three-turn cycle and still breaks well
// before the step cap.
const repeatWindow = 8

// ExecutedCall is one tool call the harness ran this turn and the observation it
// fed back. A CALL-form turn runs one; a native tool-call turn runs every call it
// carried (bug 1428).
type ExecutedCall struct {
	Action      string `json:"action"`
	Observation string `json:"observation"`
}

// Step is one turn: what the subject wrote and the tool calls the harness ran
// from it. Calls is empty for a FINAL or a stall.
type Step struct {
	ModelText string `json:"model_text"`
	// Calls are the tool calls executed this turn, in order. A native tool-call
	// turn carries several; a CALL-form turn carries one.
	Calls []ExecutedCall `json:"calls,omitempty"`
	// Truncated is true when this turn's generation hit the per-call token cap
	// (call_tokens) rather than stopping naturally. For a turn that carried an
	// edit_file action this is load-bearing: the recorded action — the file the
	// sandbox now holds — may be cut off mid-write.
	Truncated bool `json:"truncated,omitempty"`
	// Parse is how the harness read this turn (see ParseOutcome).
	Parse ParseOutcome `json:"parse"`
	// LostCalls and IgnoredCalls are this turn's TurnParse counts.
	LostCalls    int `json:"lost_calls,omitempty"`
	IgnoredCalls int `json:"ignored_calls,omitempty"`
	// PromptTokens is the prompt size the server reported for this turn, so a
	// reader can see which turn grew the prompt toward n_ctx.
	PromptTokens int `json:"prompt_tokens,omitempty"`
	// CollapsedCalls counts exact repeats of a call already run this turn. They
	// were answered with a one-line note instead of being run again (see Run).
	CollapsedCalls int `json:"collapsed_calls,omitempty"`
}

// signature is the turn's executed-call fingerprint: each call's label and
// observation, joined. The loop compares it turn to turn to catch an unproductive
// repeated-identical-call loop (bug 1428). An empty step (FINAL/stall) has no
// signature and never counts as a repeat.
func (s Step) signature() string {
	if len(s.Calls) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range s.Calls {
		b.WriteString(c.Action)
		b.WriteByte('\n')
		b.WriteString(c.Observation)
		b.WriteByte('\x1e')
	}
	return b.String()
}

// Result is a whole loop run — the transcript, why it ended, and the aggregated
// server-reported provenance across the turns.
type Result struct {
	Steps    []Step   `json:"steps"`
	Final    string   `json:"final,omitempty"`
	Terminal Terminal `json:"terminal"`
	Turns    int      `json:"turns"`
	Edits    int      `json:"edits"`
	// Model and BuildInfo are the server's account of itself on the last turn.
	Model     string `json:"model"`
	BuildInfo string `json:"build_info"`
	// PredictedTokens sums the tokens generated across all turns.
	PredictedTokens int `json:"predicted_tokens"`
	// TokensPerSecond is the last turn's throughput — the substrate tripwire.
	TokensPerSecond float64 `json:"tokens_per_second"`
	// Truncated is true if ANY turn hit the per-call token cap. Kept as the
	// coarse aggregate for backward compatibility; prefer the per-turn fields
	// below, which are what a scorer should actually key on. A bare bool here
	// cannot tell a harmless truncated read apart from a chopped edit.
	Truncated bool `json:"truncated"`
	// TruncatedTurns counts how many turns hit the per-call token cap. Zero
	// turns means no truncation; a count localises how much of the run was cut.
	TruncatedTurns int `json:"truncated_turns"`
	// TruncatedEdit is true if any TRUNCATED turn carried an edit_file action.
	// This is the dangerous case the aggregate bool hides: a chopped edit leaves
	// the sandbox holding a partial record, so a deterministic end-state scorer
	// reading that sandbox sees a truncation artifact, not the agent's
	// trajectory — the cell is UNSCOREABLE on end state. (bug 1425: a 256-token
	// cap chopped a multi-field Item-12 record mid-write and manufactured a false
	// early-exit fire; truncated=1 while the trajectory was carry-through.)
	TruncatedEdit bool `json:"truncated_edit"`
	// LostCalls sums the turns' lost calls and IgnoredCalls their ignored
	// directives (see TurnParse). Instrument notes, not scores.
	LostCalls    int `json:"lost_calls"`
	IgnoredCalls int `json:"ignored_calls"`
	// PromptTokens is the largest prompt the server reported across the turns.
	PromptTokens int `json:"prompt_tokens"`
	// CollapsedCalls sums the turns' collapsed repeat calls.
	CollapsedCalls int `json:"collapsed_calls"`
	// ContextError is the server's message when the run ended on TerminalContext.
	ContextError string `json:"context_error,omitempty"`
}

// ParseOutcomes lists each turn's parse outcome, in turn order.
func (r Result) ParseOutcomes() []string {
	out := make([]string, len(r.Steps))
	for i, s := range r.Steps {
		out[i] = string(s.Parse)
	}
	return out
}

// Run drives the loop: it sends base (the preamble plus the aided scenario) to
// the subject, parses every tool call the turn carries, executes each against sb,
// appends the real observation of each, and repeats until FINAL, a stall, the
// step cap, or a repeated-identical-call loop. base is re-sent each turn with the
// running transcript appended, matching the raw /completion path (cache_prompt
// off). A model error aborts the run.
func Run(ctx context.Context, client model.Client, base string, params model.GenParams, sb *Sandbox, cfg Config) (Result, error) {
	params.Stop = cfg.stop()
	params.MaxTokens = model.Int(cfg.callTokens())

	res := Result{}
	transcript := base
	steps := cfg.stepCap()
	var repeats repeatTracker
	for turn := 1; turn <= steps; turn++ {
		resp, err := client.Generate(ctx, transcript, params)
		if model.IsContextOverflow(err) {
			// The server generated nothing for this request, so the run so far is
			// complete as it stands: end it with what the subject did.
			var apiErr *model.APIError
			errors.As(err, &apiErr)
			res.Terminal = TerminalContext
			res.ContextError = serverMessage(apiErr.Body)
			res.Edits = sb.Edits()
			return res, nil
		}
		if err != nil {
			return Result{}, fmt.Errorf("agentloop: turn %d: %w", turn, err)
		}
		modelText := strings.TrimSpace(resp.Text)
		actions := ParseActions(modelText)
		step := res.recordTurn(turn, resp, modelText, actions)

		if terminal, ok := singleActionEnd(actions); ok {
			if terminal == TerminalFinal {
				res.Final = actions[0].Summary
			}
			res.Steps = append(res.Steps, step)
			res.Terminal = terminal
			res.Edits = sb.Edits()
			return res, nil
		}

		transcript += "\n" + modelText + res.runActions(actions, resp.Truncated, &step, sb)
		res.CollapsedCalls += step.CollapsedCalls
		res.Steps = append(res.Steps, step)

		if repeats.looping(step.signature()) {
			res.Terminal = TerminalLoop
			res.Edits = sb.Edits()
			return res, nil
		}
	}
	res.Terminal = TerminalCap
	res.Edits = sb.Edits()
	return res, nil
}

// recordTurn folds one reply's model and token figures into the result and
// returns the turn's step, before any tool runs.
func (res *Result) recordTurn(turn int, resp model.Response, modelText string, actions []Action) Step {
	res.Turns = turn
	res.Model = resp.Model
	res.BuildInfo = resp.SystemFingerprint
	res.PredictedTokens += resp.Timings.PredictedN
	res.TokensPerSecond = resp.Timings.PredictedPerSecond

	tp := ClassifyTurn(modelText, actions)
	step := Step{ModelText: modelText, Truncated: resp.Truncated, Parse: tp.Outcome,
		LostCalls: tp.Lost, IgnoredCalls: tp.Ignored, PromptTokens: resp.Timings.PromptN}
	res.LostCalls += tp.Lost
	res.IgnoredCalls += tp.Ignored
	res.PromptTokens = max(res.PromptTokens, resp.Timings.PromptN)
	if resp.Truncated {
		res.Truncated = true
		res.TruncatedTurns++
	}
	return step
}

// singleActionEnd reports the terminal a turn ends on before any tool runs.
//
// A FINAL or a stall ends the turn before any tool runs. Both only ever
// arrive as a single-action turn (the native path returns tool calls only).
func singleActionEnd(actions []Action) (Terminal, bool) {
	if len(actions) != 1 {
		return "", false
	}
	switch actions[0].Kind {
	case ActionFinal:
		return TerminalFinal, true
	case ActionNone:
		return TerminalStall, true
	}
	return "", false
}

// runActions runs the turn's calls against sb, records each on step, and returns
// the observation text to append to the transcript.
//
// Execute every tool call this turn carried, in order, and feed back one
// OBSERVATION per call. A single-call turn keeps the bare "OBSERVATION: …"
// shape; a multi-call turn labels each observation with the call it answers
// so the model can tell which read returned what.
func (res *Result) runActions(actions []Action, truncated bool, step *Step, sb *Sandbox) string {
	var out strings.Builder
	multi := len(actions) > 1
	ran := map[string]bool{}
	for _, action := range actions {
		label := actionLabel(action)
		// Run each distinct call once per turn. Qwen can repeat one native
		// read batch until it fills call_tokens (45 calls over four files),
		// and feeding every repeat's full file back overflowed n_ctx within
		// a few turns (bug 1442). An exact repeat — same tool, argument and
		// edit body — changes nothing, so it gets a one-line note instead.
		key := label + "\x00" + action.Content
		if ran[key] {
			step.CollapsedCalls++
			out.WriteString("\nOBSERVATION (" + label + "): (same call as above in this turn; not run again)")
			continue
		}
		ran[key] = true
		// An edit turn cut off mid-write corrupts the sandbox an end-state
		// scorer reads; record that distinctly from a harmless truncated read.
		if truncated && action.Kind == ActionEdit {
			res.TruncatedEdit = true
		}
		obs := sb.Do(action)
		step.Calls = append(step.Calls, ExecutedCall{Action: label, Observation: obs})
		if multi {
			out.WriteString("\nOBSERVATION (" + label + "): " + obs)
		} else {
			out.WriteString("\nOBSERVATION: " + obs)
		}
	}
	return out.String()
}

// repeatTracker holds the recent step signatures for the repeated-call breaker.
type repeatTracker struct{ recent []string }

// looping records sig and reports whether it now recurs repeatedCallLimit times
// within the last repeatWindow signatures. An empty sig is not recorded.
//
// Break an unproductive repeated-call loop before the step cap: the model
// emitting the same call with the same observation again and again is stuck,
// not working. The count is over a recent window, not consecutive turns, so
// an alternating near-identical loop is caught too (bug 1428: Qwen re-read
// one file; bug 1430: Qwen alternated near-identical native-XML edits).
func (rt *repeatTracker) looping(sig string) bool {
	if sig == "" {
		return false
	}
	rt.recent = append(rt.recent, sig)
	if len(rt.recent) > repeatWindow {
		rt.recent = rt.recent[1:]
	}
	occurrences := 0
	for _, s := range rt.recent {
		if s == sig {
			occurrences++
		}
	}
	return occurrences >= repeatedCallLimit
}

// serverMessage returns the "error.message" field of a llama-server error body,
// or the body itself when it carries none.
func serverMessage(body string) string {
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal([]byte(body), &e) == nil && e.Error.Message != "" {
		return e.Error.Message
	}
	return body
}

// actionLabel is a short, stable label for a step's action, for the transcript
// and for at-a-glance reading of a run.
func actionLabel(a Action) string {
	switch a.Kind {
	case ActionList:
		return "list_files " + a.Arg
	case ActionRead:
		return "read_file " + a.Arg
	case ActionQuery:
		return "run_query " + a.Arg
	case ActionEdit:
		return "edit_file " + a.Arg
	case ActionFinal:
		return "FINAL"
	case ActionUnknown:
		return "unknown:" + a.Tool
	default:
		return "none"
	}
}

// Transcript renders a Result as the flat text a blind rater scores: the
// subject's turns, the tool calls the harness actually executed each turn, the
// observation each returned, and the final answer. The setup (raw vs loop) is not
// named here — the rater sees only conduct.
//
// A turn's text can carry echoed "CALL …" lines and stray "FINAL …" text the
// harness never ran (a CALL-form turn runs its first directive; the rest is
// noise). Each EXECUTED line names one call that ran, in order, so a reader — and
// the order-based scorer downstream — keys on executed conduct, not on the
// model's own echoes. The raw ModelText is kept verbatim above it for provenance.
func (r Result) Transcript() string {
	var b strings.Builder
	for i, s := range r.Steps {
		fmt.Fprintf(&b, "[turn %d] %s\n", i+1, s.ModelText)
		if s.Truncated {
			fmt.Fprintf(&b, "[turn %d truncated at the call_tokens cap — this turn's action may be cut off mid-write]\n", i+1)
		}
		for _, c := range s.Calls {
			fmt.Fprintf(&b, "EXECUTED: %s\n", c.Action)
			fmt.Fprintf(&b, "OBSERVATION: %s\n", c.Observation)
		}
		if s.CollapsedCalls > 0 {
			fmt.Fprintf(&b, "[collapsed: %d repeated identical calls in turn %d were not run again]\n", s.CollapsedCalls, i+1)
		}
	}
	switch r.Terminal {
	case TerminalCap:
		b.WriteString("[loop ended: step cap reached]\n")
	case TerminalLoop:
		b.WriteString("[loop ended: repeated identical calls — the subject stopped making progress]\n")
	case TerminalContext:
		fmt.Fprintf(&b, "[loop ended: the next request exceeds the model context — %s]\n", r.ContextError)
	}
	if r.TruncatedEdit {
		b.WriteString("[WARNING: a truncated turn carried an edit — the final sandbox may hold a partial record; this cell is UNSCOREABLE on end state]\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
