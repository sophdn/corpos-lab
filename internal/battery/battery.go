package battery

// The ALPHABET Entry Battery (Phase-1 subset), ported from
// lab-app-server/src/sequences/battery.rs.
//
// In scope (implemented): items 1, 2, 3, 4, 6, 9, 10, 11, 12, 13, 15.
//   - Deterministic (no model call): items 3 and 13 — structural checks the
//     archived structural-prober covered, where a reproducible count/lookup beats
//     a non-reproducible LLM verdict (sub-decision recorded on each function).
//   - Model-assessed verdict steps: items 6, 11 — semantic checks, on the
//     item 1/9 pattern with the injectable model.Client.
//   - Empirical (no judge call): item 12 — its default-alignment verdict is
//     measured against the unguided-baseline seam (baseline.go), not introspected
//     by a judge (bug 1332).
// Deferred (register as Deferred stubs): items 5, 7, 8, 14 — reasons in
// DeferredPending; a separate task decides these.

// StepVersion returns the per-step implementation version persisted
// alongside each step's outcome. Deferred items carry "0.0.0-deferred" so a
// post-hoc query can distinguish "ran with current logic" from "registered
// but not yet evaluable"; unknown step names return the reserved "0.0.0".
func StepVersion(stepName string) string {
	switch stepName {
	case "item9-universality",
		"item15-fallout-profile":
		// Repaired under task 3586: item 15 now resolves and dimensionally
		// checks the referenced profile instead of a bare substring; item 9
		// carries the 2026-04-23 calibration-instance ruling. Bumped past the
		// 0.1.0 cohort so a post-hoc query separates repaired-logic runs from
		// the pre-repair runs whose item-9/15 verdicts cannot be trusted.
		return "0.2.0"
	case "item1-xyz-specificity":
		// Repaired under task 3588: the prompt now resolves X and Y from the
		// entry's decision point and firing condition before judging them,
		// instead of reading the canonical "Taking X from Y" notation as the
		// components themselves. Bumped past the 0.1.0 cohort so a post-hoc
		// query separates resolved-reading runs from the literal-reading runs,
		// which failed every entry in the corpus at step 0.
		return "0.2.0"
	case "item6-entry-coherence":
		// Repaired under the model-assessed-battery hardening (bug 1334 +
		// suggestion 175): item 6 now delivers the glyph definition INLINE to the
		// judge — and the provenance taxonomy inline when the entry cites a
		// provenance type — instead of asking the judge to trace fields against
		// its own memory of the definition, which produced hallucinated
		// "missing rule" failures on coherent glyphs. Bumped past the 0.1.0 cohort
		// so a post-hoc query separates reference-inlined runs from the
		// definition-from-memory runs, whose item-6 verdicts cannot be trusted.
		return "0.2.0"
	case "item12-default-alignment":
		// Rebuilt EMPIRICAL under item12-empirical-default-check: item 12's verdict
		// is now measured against the weak-local-shelf unguided baseline (does the
		// class's target behavior fire WITHOUT the glyph?) instead of asking a judge
		// to introspect on its own training, which gave model-relative false FAILs
		// that a stronger judge could not fix (bug 1332). Bumped past the 0.1.0
		// cohort so a post-hoc query separates empirical-item-12 runs from the old
		// introspective ones, whose item-12 verdicts cannot be trusted.
		return "0.2.0"
	case "item2-intent-language-scan",
		"item4-na",
		"item10-axis-presence",
		"item3-duplicate-check",
		"item11-safety-class",
		"item13-contamination-radius":
		// Items 3, 11, 13 mechanized under port-structural-prober-items:
		// 3 and 13 deterministic (structural-prober logic), 11 a model-assessed
		// verdict step. They join the 0.1.0 cohort — first real logic, no prior
		// repaired-vs-pre-repair split to encode. (Item 12 started here too but was
		// rebuilt empirical and bumped to 0.2.0 above.)
		return "0.1.0"
	case "item5-y-not-fire",
		"item7-sister-mirror",
		"item8-phenomenological",
		"item14-globality-demand":
		return "0.0.0-deferred"
	default:
		return "0.0.0"
	}
}

// BuildBattery assembles the 15-item battery in stratum order (fail-fast:
// lower-cost items first): mechanical (1–6), structural-behavioral (7–10),
// interpretive (11–12), systemic (13–15). Deferred items occupy their slot
// and emit Deferred so the step count matches ALPHABET_ENTRY_BATTERY.md
// exactly and ComposeRunVerdict classifies the run correctly.
func BuildBattery() *Sequence {
	seq := NewSequence("battery")

	// Stratum: Mechanical (Items 1–6)
	seq.AddStep("item1-xyz-specificity", Item1XYZSpecificity)
	seq.AddStep("item2-intent-language-scan", Item2IntentLanguageScan)
	seq.AddStep("item3-duplicate-check", Item3DuplicateCheck)
	seq.AddStep("item4-na", Item4NA)
	seq.AddStep("item5-y-not-fire", DeferredStep(5))
	seq.AddStep("item6-entry-coherence", Item6EntryCoherence)

	// Stratum: Structural-behavioral (Items 7–10)
	seq.AddStep("item7-sister-mirror", DeferredStep(7))
	seq.AddStep("item8-phenomenological", DeferredStep(8))
	seq.AddStep("item9-universality", Item9Universality)
	seq.AddStep("item10-axis-presence", Item10AxisPresence)

	// Stratum: Interpretive (Items 11–12)
	seq.AddStep("item11-safety-class", Item11SafetyClass)
	seq.AddStep("item12-default-alignment", Item12DefaultAlignment)

	// Stratum: Systemic (Items 13–15)
	seq.AddStep("item13-contamination-radius", Item13ContaminationRadius)
	seq.AddStep("item14-globality-demand", DeferredStep(14))
	seq.AddStep("item15-fallout-profile", Item15FalloutProfile)

	return seq
}
