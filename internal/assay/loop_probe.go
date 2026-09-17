package assay

import (
	"context"
	"fmt"

	"corpos-lab/internal/agentloop"
	"corpos-lab/internal/model"
)

// LoopPreambleDelimiter separates the loop preamble from the aided scenario. It
// matches AssemblePrompt's own "\n---\n" separator so the aid arrives to the
// subject exactly as it does in the single-turn probe — the two setups differ
// only in the preamble and the ability to act, never in how the aid is worded.
const LoopPreambleDelimiter = "\n---\n"

// LoopArtifacts are the extra outputs of a loop run the runner persists: the
// full loop result (transcript, terminal, provenance) and the sandbox state the
// subject left behind.
type LoopArtifacts struct {
	Result  agentloop.Result
	Sandbox map[string]string
}

// RunLoopProbe runs one agentic-loop cell: it assembles the same aided scenario
// the single-turn probe uses, prepends the loop preamble, and drives the minimal
// tool loop against a fresh sandbox. It returns the captured (unscored) row, the
// response whose text is the flat transcript a blind rater scores, and the loop
// artifacts for provenance.
//
// Scoring is deferred to a judge, exactly as RunProbe defers it: the loop's job
// is capture. The row records the terminal state (final / stall / cap) and the
// edit count in its rationale so an at-a-glance read is possible before scoring.
func RunLoopProbe(ctx context.Context, m model.Client, itemID string, cond Condition, run int, mats Materials, preamble string, sandbox map[string]string, samp Sampling, cfg agentloop.Config) (ScoreRow, ProbeResponse, LoopArtifacts, error) {
	aided, err := AssemblePrompt(cond, mats)
	if err != nil {
		return ScoreRow{}, ProbeResponse{}, LoopArtifacts{}, err
	}
	base := preamble + LoopPreambleDelimiter + aided

	sb := agentloop.NewSandbox(sandbox)
	res, err := agentloop.Run(ctx, m, base, samp.GenParamsForRun(run), sb, cfg)
	if err != nil {
		return ScoreRow{}, ProbeResponse{}, LoopArtifacts{}, fmt.Errorf("assay: %s run %d: %w", cond, run, err)
	}

	truncNote := ""
	if res.Truncated {
		truncNote = ":truncated"
	}
	rationale := fmt.Sprintf("agentic-loop-probe:%s:terminal=%s:turns=%d:edits=%d%s:unscored",
		cond, res.Terminal, res.Turns, res.Edits, truncNote)

	row := ScoreRow{
		Item:      itemID,
		Condition: cond,
		Run:       run,
		Score:     Unscored,
		Rationale: rationale,
		Observed: Observed{
			Model:           res.Model,
			BuildInfo:       res.BuildInfo,
			TokensPerSecond: res.TokensPerSecond,
			PredictedTokens: res.PredictedTokens,
			Truncated:       res.Truncated,
		},
	}
	response := ProbeResponse{
		Condition: cond,
		Run:       run,
		Prompt:    base,
		Text:      res.Transcript(),
	}
	return row, response, LoopArtifacts{Result: res, Sandbox: sb.Snapshot()}, nil
}
