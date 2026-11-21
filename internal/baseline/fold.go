package baseline

import "corpos-lab/internal/battery"

// Gap records a model whose unguided baseline could not be resolved into a fire
// or a miss, and why. A gap OMITS the model from the outcome rather than
// guessing: item 12 defers on a model it has no measurement for, and a
// manufactured false verdict is the exact failure the empirical rebuild exists
// to prevent.
type Gap struct {
	ModelID string `json:"model_id"`
	Reason  string `json:"reason"`
}

// Fold turns per-model captures and the raters' code maps into the
// battery.BaselineOutcome item 12 reads through `battery -baseline`. Each entry
// of raterScores is one independent rater family's {slice_id: code} result.
//
// A run is RESOLVED when the raters agree on its code; a disagreement, a blank
// code, or a missing code leaves it unresolved. A model FIRED when TargetCode is
// the resolved code on a strict majority of its resolved runs. A model with no
// resolved run is omitted and returned as a Gap.
func Fold(captures []Capture, raterScores []map[string]string) (battery.BaselineOutcome, []Gap) {
	var out battery.BaselineOutcome
	var gaps []Gap
	for _, c := range captures {
		fired, resolved := 0, 0
		for _, run := range c.Runs {
			code, ok := consensus(run.SliceID, raterScores)
			if !ok {
				continue
			}
			resolved++
			if code == TargetCode {
				fired++
			}
		}
		id := c.ModelID
		if id == "" {
			id = c.Requested
		}
		if resolved == 0 {
			gaps = append(gaps, Gap{ModelID: id, Reason: "no run resolved to a consensus code"})
			continue
		}
		out.Models = append(out.Models, battery.BaselineModelResult{
			ModelID: id,
			Version: c.Version,
			Fired:   fired*2 > resolved,
			Runs:    resolved,
		})
	}
	return out, gaps
}

// consensus returns the code the raters agree on for a slice id, and whether
// they agreed. With one rater its code stands (the standard prefers more than
// one, but folding never refuses to render what it was given). A blank or
// missing code, or any two raters disagreeing, is no consensus. With no raters
// at all there is nothing to agree on.
func consensus(id string, raterScores []map[string]string) (string, bool) {
	agreed := ""
	seen := false
	for _, scores := range raterScores {
		code, ok := scores[id]
		if !ok || code == "" {
			return "", false
		}
		if !seen {
			agreed, seen = code, true
			continue
		}
		if code != agreed {
			return "", false
		}
	}
	return agreed, seen
}
