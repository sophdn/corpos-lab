package agentloop

import (
	"context"
	"fmt"
	"strings"

	"corpos-lab/internal/model"
)

// Default loop bounds. A turn is one action, so the token cap per call is small;
// the step cap bounds the whole run. Both are overridable per study.
const (
	// DefaultStepCap bounds the number of turns before the loop gives up.
	DefaultStepCap = 10
	// DefaultCallTokens caps one turn's generation. One action plus brief
	// reasoning fits well inside this.
	DefaultCallTokens = 256
)

// defaultStop halts a turn the moment the subject starts writing its own
// OBSERVATION line, so the harness — not the model — supplies the observation
// (the feasibility smoke showed the model fabricating observations otherwise).
//
// It deliberately does NOT stop on a CALL. An earlier version stopped on "\nCALL"
// to block a second action, but that also truncated the FIRST call whenever the
// subject wrote a sentence of prose before it — the calibration run then scored a
// real tool call as a stall. The parser takes the first action of a turn and the
// harness injects the real observation, so extra text after the first call is
// harmless and needs no stop. Fabrication is the only thing worth halting.
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
)

// Step is one turn: what the subject wrote, the action parsed from it, and the
// observation the harness fed back (empty for a FINAL or a stall).
type Step struct {
	ModelText   string `json:"model_text"`
	Action      string `json:"action"`
	Observation string `json:"observation,omitempty"`
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
	// Truncated is true if any turn hit the per-call token cap.
	Truncated bool `json:"truncated"`
}

// Run drives the loop: it sends base (the preamble plus the aided scenario) to
// the subject, parses one action per turn, executes it against sb, appends the
// real observation, and repeats until FINAL, a stall, or the step cap. base is
// re-sent each turn with the running transcript appended, matching the raw
// /completion path (cache_prompt off). A model error aborts the run.
func Run(ctx context.Context, client model.Client, base string, params model.GenParams, sb *Sandbox, cfg Config) (Result, error) {
	params.Stop = cfg.stop()
	params.MaxTokens = model.Int(cfg.callTokens())

	res := Result{}
	transcript := base
	steps := cfg.stepCap()
	for turn := 1; turn <= steps; turn++ {
		resp, err := client.Generate(ctx, transcript, params)
		if err != nil {
			return Result{}, fmt.Errorf("agentloop: turn %d: %w", turn, err)
		}
		res.Turns = turn
		res.Model = resp.Model
		res.BuildInfo = resp.SystemFingerprint
		res.PredictedTokens += resp.Timings.PredictedN
		res.TokensPerSecond = resp.Timings.PredictedPerSecond
		if resp.Truncated {
			res.Truncated = true
		}

		modelText := strings.TrimSpace(resp.Text)
		action := ParseAction(modelText)
		step := Step{ModelText: modelText, Action: actionLabel(action)}

		switch action.Kind {
		case ActionFinal:
			res.Final = action.Summary
			res.Steps = append(res.Steps, step)
			res.Terminal = TerminalFinal
			res.Edits = sb.Edits()
			return res, nil
		case ActionNone:
			res.Steps = append(res.Steps, step)
			res.Terminal = TerminalStall
			res.Edits = sb.Edits()
			return res, nil
		default:
			obs := sb.Do(action)
			step.Observation = obs
			res.Steps = append(res.Steps, step)
			transcript += "\n" + modelText + "\nOBSERVATION: " + obs
		}
	}
	res.Terminal = TerminalCap
	res.Edits = sb.Edits()
	return res, nil
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
// subject's turns, the observations the harness returned, and the final answer.
// The setup (raw vs loop) is not named here — the rater sees only conduct.
func (r Result) Transcript() string {
	var b strings.Builder
	for i, s := range r.Steps {
		fmt.Fprintf(&b, "[turn %d] %s\n", i+1, s.ModelText)
		if s.Observation != "" {
			fmt.Fprintf(&b, "OBSERVATION: %s\n", s.Observation)
		}
	}
	if r.Terminal == TerminalCap {
		b.WriteString("[loop ended: step cap reached]\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
