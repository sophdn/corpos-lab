package battery

import (
	"context"
	"fmt"
	"strings"

	"corpos-lab/internal/model"
)

// verdictGenParams pins the sampling for battery verdict steps: deterministic
// scoring (greedy) with room for a thinking model's reasoning before the
// verdict. The token budget covers a reasoning model's <think> block plus the
// short PASS/FAIL answer — a 256-token cap truncated a thinking assessor
// mid-reasoning, leaving no answer to parse and manufacturing a FAIL. A
// non-thinking model simply stops at its short answer well under the cap.
// Measured: an item-1 assessment of a real entry spent ~2400 tokens on
// reasoning before the verdict, so the cap sits well above that (a truncated
// completion has no answer to parse and reads as a FAIL). A non-thinking model
// still stops at its short answer, and greedy decoding has not run away.
func verdictGenParams() model.GenParams {
	return model.GenParams{
		Temperature: model.Float64(0.0),
		MaxTokens:   model.Int(8000),
	}
}

// VerdictGenParams exposes the sampler the model-assessed steps send, so a run
// can record what it actually ran under without duplicating the values.
func VerdictGenParams() model.GenParams { return verdictGenParams() }

// ParseModelVerdict converts a model's PASS / FAIL response into a typed
// Verdict. The single seam where LLM prose becomes a typed scoring outcome —
// every inference step goes through here so the string→Verdict mapping has
// exactly one spec-and-one-tested implementation.
//
// Rules (source parity):
//   - Leading PASS (any case) → Pass.
//   - Leading FAIL (any case) → Fail with the remaining text trimmed as the
//     reason ("Item N: <reason>", or "Item N: unspecified failure").
//   - Anything else → Fail with a malformed-response reason pointing at the
//     model's actual reply.
func ParseModelVerdict(responseText string, item int) Verdict {
	trimmed := strings.TrimSpace(responseText)
	upper := strings.ToUpper(trimmed)
	switch {
	case strings.HasPrefix(upper, "PASS"):
		return Pass()
	case strings.HasPrefix(upper, "FAIL"):
		reason := strings.TrimSpace(trimmed[4:])
		if reason == "" {
			return FailItem(item, fmt.Sprintf("Item %d: unspecified failure", item))
		}
		return FailItem(item, fmt.Sprintf("Item %d: %s", item, reason))
	default:
		return FailItem(item, fmt.Sprintf(
			"Item %d: model response did not start with PASS or FAIL: %s", item, trimmed))
	}
}

// runVerdictStep sends the prompt, maps transport errors to an error
// outcome ("Item N: model error: …"), and parses the reply into a verdict.
func runVerdictStep(ctx context.Context, st *State, item int, prompt string) StepOutcome {
	resp, err := st.Model.Generate(ctx, prompt, verdictGenParams())
	if err != nil {
		return ErrorOutcome(fmt.Sprintf("Item %d: model error: %v", item, err))
	}
	return VerdictOutcome(ParseModelVerdict(resp.Text, item))
}

// Item1XYZSpecificity — X/Y/Z specificity: the model evaluates whether the
// Marker invariant fills X (action), Y (context), Z (violation) with
// particulars rather than category labels or paraphrases.
//
// The prompt resolves X and Y before judging them. The corpus's canonical
// Marker form is the literal sentence "Taking X from Y → <Z prose>"
// (GLYPH_DEFINITION.md:72) — X and Y are notation, standing for the action and
// the context condition the entry defines elsewhere (its decision point,
// firing condition, and does-not-fire-on carve-out). Reading the two symbols
// as the components themselves makes every entry in the corpus an instant
// FAIL on a naming convention, which is not what this item tests: the item
// tests whether the entry's action and context are specifiable, and the
// 2026-04-03 human assessor resolved them from the surrounding prose before
// scoring. An entry that gives no resolvable definition anywhere still fails —
// the resolution step is where an unspecifiable X or Y shows itself, so
// scoping the read this way tightens the item rather than excusing it.
func Item1XYZSpecificity(ctx context.Context, st *State) StepOutcome {
	prompt := fmt.Sprintf(
		"You are evaluating a glyph entry for structural specificity.\n\n"+
			"The entry's Marker invariant must name:\n"+
			"- X: a specific action, structural position, design decision, or omission\n"+
			"- Y: a specific, specifiable context condition\n"+
			"- Z: a specific structural violation (a named invariant that fails)\n\n"+
			"RESOLVE BEFORE JUDGING. This corpus writes its Marker invariant in a "+
			"canonical notation: \"Taking X from Y → <violation prose>\". The letters "+
			"X and Y in that sentence are that notation — symbols standing for "+
			"components the entry defines elsewhere — not the components themselves, "+
			"and their presence is not by itself a failure. Before scoring, resolve "+
			"each component from the whole entry:\n"+
			"- X = the action or omission the entry's decision point and firing "+
			"condition describe the agent taking.\n"+
			"- Y = the context condition the entry's decision point, does-not-fire-on "+
			"carve-out, and discriminating conditions describe.\n"+
			"- Z = the violation prose written after the arrow in the invariant.\n\n"+
			"Then judge the RESOLVED X, Y, and Z for specificity. If the entry gives "+
			"no resolvable definition for a component anywhere — the symbol is never "+
			"cashed out into an action or a context condition — that component is an "+
			"unresolved placeholder and the entry FAILS.\n\n"+
			"None of the resolved components may be category labels, paraphrases, or "+
			"placeholders "+
			"(e.g. \"skips a step\", \"makes an error\", \"in a context where a gate is absent\").\n\n"+
			"Entry to evaluate:\n---\n%s\n---\n\n"+
			"Respond with exactly one of:\n"+
			"PASS — if all three resolved components name specifics\n"+
			"FAIL <reason> — if any component is unresolvable, a placeholder, or a "+
			"category label, naming which component and why",
		st.Content)
	return runVerdictStep(ctx, st, 1, prompt)
}

// Item9Universality — the model checks whether structural fields contain
// project-specific references (file paths, protocol slugs, artifact names,
// project vocabulary). The prompt carries the 2026-04-23 calibration-instance
// ruling (lab-app/fidelity/item9-criterion-ruling.md): a calibration instance
// labelled "recognition illustration" or "does not define scope" gets no
// exemption from the scan, so project-specific vocabulary inside such a
// labelled instance is still a FAIL. Before this the runner embodied the
// interpretation the ruling voided.
func Item9Universality(ctx context.Context, st *State) StepOutcome {
	prompt := fmt.Sprintf(
		"You are evaluating a glyph entry for universality.\n\n"+
			"Check whether any structural fields (Y marker, invariants, firing condition, "+
			"does-not-fire-on, violation signal) contain project-specific references:\n"+
			"- File paths (e.g. process-docs/ALPHABET.md)\n"+
			"- Protocol slugs by name\n"+
			"- Project-specific artifact names\n"+
			"- Project-specific vocabulary that would not exist in another agent system\n\n"+
			"Assess only the structural fields named above. The entry may also carry a "+
			"`**Fallout profile:**` reference to a separate fallout document — that is a "+
			"metadata pointer, not a structural field, so its path is not a Sub-check A "+
			"reference and must be ignored.\n\n"+
			"Calibration instances in the violation signal and illustrative carve-outs in "+
			"does-not-fire-on are structural fields and must meet the same universality "+
			"standard. No carve-outs. A label marking a calibration instance as "+
			"\"recognition illustration\" or \"does not define scope\" describes its "+
			"evidentiary function — it does not exempt the instance from this scan. "+
			"Project-specific vocabulary inside such a labelled calibration instance is "+
			"still a FAIL (ruled 2026-04-23).\n\n"+
			"The decision class must be recognizable by an agent in a different project "+
			"that has this decision class, without importing this project's infrastructure.\n\n"+
			"Entry to evaluate:\n---\n%s\n---\n\n"+
			"Respond with exactly one of:\n"+
			"PASS — if all structural fields use project-agnostic vocabulary\n"+
			"FAIL <reason> — if any project-specific reference is found, naming it",
		st.Content)
	return runVerdictStep(ctx, st, 9, prompt)
}
