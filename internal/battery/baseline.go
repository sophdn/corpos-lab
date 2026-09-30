package battery

import (
	"context"
	"fmt"
	"strings"
)

// BaselineOutcome is the empirical unguided-baseline measurement item 12
// consumes IN PLACE OF a model's introspection about its own training. For the
// candidate's decision class, it records whether the class's target (avoidance)
// behavior FIRES when the class scenario is run UNGUIDED — no glyph prefix — the
// baseline (assay.Baseline) arm run across the weak local treatment shelf (the
// models the studies actually run, e.g. Mistral-7B; see deploy/shelf.toml).
//
// It exists because the old item 12 asked a model judge to introspect on its
// own training to decide whether a glyph's behavior was a "trained default"
// (bug item12-default-alignment-strong-model-introspection, 1332). That is
// model-relative and gives false FAILs: a model cannot reliably report what it
// would do unprompted, and a stronger judge does not fix it — the blindness is
// structural. The empirical rebuild replaces the introspection with a
// measurement: run the scenario without the glyph and observe whether the
// behavior fires anyway.
//
// The verdict it drives is empirical and population-relative, and it says so:
//   - fires unguided ACROSS the shelf   -> the behavior IS a trained default the
//     glyph merely restates -> item 12 FAIL (no friction = fail).
//   - MISSED unguided on any shelf model -> the glyph does real work there, so
//     the entry is distinguishable from that model's trained default -> item 12
//     PASS (relative to the measured population).
//
// The zero value (no models) means no baseline was supplied. Item 12 then
// records a gap (a Deferred verdict) rather than crashing or manufacturing a
// FAIL — an empirical item with no measurement cannot render a verdict, and
// refusing the run for want of it is the freeze reflex the record-what-ran
// invariant rejects (CLAUDE.md: nothing may fail a run). This is the injected
// seam item 12 reads instead of driving inference inside the sans-IO battery
// package, mirroring item 6's inline ReferenceMaterial: the baseline inference
// runs at the IO edge and the result is delivered here.
type BaselineOutcome struct {
	// Models is the per-model unguided result, one entry per weak-shelf model the
	// baseline arm ran. Recorded so the verdict names the population it is
	// relative to (record what ran — the verdict is honest about being
	// population-relative).
	Models []BaselineModelResult `json:"models"`
}

// BaselineModelResult is one shelf model's unguided-baseline measurement: did
// the class's target behavior fire on this model with no glyph in the prompt.
type BaselineModelResult struct {
	// ModelID is the concrete gguf the baseline arm ran on this cell (the shelf
	// model, resolved from deploy/shelf.toml — never the role token).
	ModelID string `json:"model_id"`
	// Version is that model's recorded version string.
	Version string `json:"version"`
	// Fired reports whether the class's target (avoidance) behavior fired on this
	// model with NO glyph in the prompt — the behavior the agent produced anyway.
	Fired bool `json:"fired"`
	// Runs is how many unguided replicates this measurement folded, so a reader
	// can weigh the cell rather than trust a single integer (read cells, not
	// counts).
	Runs int `json:"runs"`
}

// supplied reports whether any shelf model's unguided baseline was measured. An
// empty outcome (zero value, or an empty non-nil slice) is "no baseline", the
// degrade path item 12 defers on.
func (b BaselineOutcome) supplied() bool { return len(b.Models) > 0 }

// firesAcrossShelf reports whether the class's target behavior fired UNGUIDED on
// EVERY shelf model measured. "Across the shelf" is unanimous on purpose: a
// single shelf model that MISSED the behavior unguided is a model for which the
// glyph does real work, so the entry is distinguishable from that model's
// trained default and item 12 must not fail on it. Requiring unanimity is the
// conservative choice against the false-FAIL failure that motivated the
// empirical rebuild (bug 1332) — the item fails only when the behavior is a
// trained default across the whole measured population.
func (b BaselineOutcome) firesAcrossShelf() bool {
	if !b.supplied() {
		return false
	}
	for _, m := range b.Models {
		if !m.Fired {
			return false
		}
	}
	return true
}

// describe renders the measured shelf population for the verdict reason, so the
// population-relative verdict names what it ran against (record what ran).
func (b BaselineOutcome) describe() string {
	parts := make([]string, 0, len(b.Models))
	for _, m := range b.Models {
		state := "missed"
		if m.Fired {
			state = "fired"
		}
		id := m.ModelID
		if id == "" {
			id = "unnamed-model"
		}
		parts = append(parts, fmt.Sprintf("%s %s unguided (%d runs)", id, state, m.Runs))
	}
	return strings.Join(parts, "; ")
}

// Item12DefaultAlignment — default-alignment, EMPIRICAL (not introspective).
//
// The verdict is measured against the weak local treatment shelf, not asked of
// a judge: item 12 reads the unguided-baseline outcome supplied through the
// injected seam (st.Baseline) and renders PASS/FAIL from it. The baseline arm
// runs the class scenario with NO glyph prefix across the shelf; whether the
// class's target (avoidance) behavior fires without the glyph decides the item.
//
// This replaces the introspective judge (bug 1332), which asked a model to
// report whether a glyph's behavior was one of its own trained defaults — a
// model-relative question that gives false FAILs and that a stronger judge
// cannot fix, because the blindness is structural. The empirical verdict is
// population-relative and honest about it: the reason names the measured shelf,
// and the run provenance captures which models and runs produced it.
//
// Polarity is unchanged from the old item 12 — no friction = fail — but it is
// now measured: fires unguided across the shelf means the glyph merely restates
// a trained default and does no work (FAIL); missed unguided on any shelf model
// means the glyph does real work there (PASS).
//
// It no longer calls the model, so the context is unused. With no baseline
// supplied it defers with a recorded gap rather than failing the run.
func Item12DefaultAlignment(_ context.Context, st *State) StepOutcome {
	b := st.Baseline
	if !b.supplied() {
		return VerdictOutcome(Deferred(
			"Item 12: no unguided baseline supplied — the empirical default-alignment " +
				"verdict needs the class scenario run UNGUIDED across the weak local " +
				"treatment shelf, and none was wired for this run"))
	}
	if b.firesAcrossShelf() {
		return FailItemOutcome(12, fmt.Sprintf(
			"Item 12: the class's target behavior fired UNGUIDED across the whole "+
				"measured shelf (%s) — it is a trained default the glyph merely restates, "+
				"so it does no work (no friction = fail)", b.describe()))
	}
	return VerdictOutcome(PassWithCondition(fmt.Sprintf(
		"item 12 empirical: the class's target behavior was MISSED unguided on at "+
			"least one shelf model, so the glyph does real work — verdict relative to "+
			"the measured population (%s)", b.describe())))
}
