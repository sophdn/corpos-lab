package battery

// The ALPHABET Entry Battery (Phase-1 subset), ported from
// lab-app-server/src/sequences/battery.rs.
//
// In scope (implemented): items 1, 2, 4, 9, 10, 15.
// Deferred (register as Deferred stubs, exactly as in the source):
// items 3, 5, 6, 7, 8, 11, 12, 13, 14 — reasons in DeferredPending.

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
	case "item1-xyz-specificity",
		"item2-intent-language-scan",
		"item4-na",
		"item10-axis-presence":
		return "0.1.0"
	case "item3-duplicate-check",
		"item5-y-not-fire",
		"item6-entry-coherence",
		"item7-sister-mirror",
		"item8-phenomenological",
		"item11-safety-class",
		"item12-default-alignment",
		"item13-contamination-radius",
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
	seq.AddStep("item3-duplicate-check", DeferredStep(3))
	seq.AddStep("item4-na", Item4NA)
	seq.AddStep("item5-y-not-fire", DeferredStep(5))
	seq.AddStep("item6-entry-coherence", DeferredStep(6))

	// Stratum: Structural-behavioral (Items 7–10)
	seq.AddStep("item7-sister-mirror", DeferredStep(7))
	seq.AddStep("item8-phenomenological", DeferredStep(8))
	seq.AddStep("item9-universality", Item9Universality)
	seq.AddStep("item10-axis-presence", Item10AxisPresence)

	// Stratum: Interpretive (Items 11–12)
	seq.AddStep("item11-safety-class", DeferredStep(11))
	seq.AddStep("item12-default-alignment", DeferredStep(12))

	// Stratum: Systemic (Items 13–15)
	seq.AddStep("item13-contamination-radius", DeferredStep(13))
	seq.AddStep("item14-globality-demand", DeferredStep(14))
	seq.AddStep("item15-fallout-profile", Item15FalloutProfile)

	return seq
}
