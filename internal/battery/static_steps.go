package battery

import (
	"context"
	"fmt"
	"strings"
)

// intentRejectPhrases are the intent-modeling constructions item 2 rejects:
// each makes the firing condition uncheckable from observable artifacts.
// Order matters — the first match is the reported phrase (source parity).
var intentRejectPhrases = []string{
	"when the agent decides",
	"when the agent believes",
	"when the agent thinks",
	"when the agent is confused",
	"when the agent is under pressure",
	"when the agent intends",
	"when the agent wants",
	"when the agent feels",
}

// Item2IntentLanguageScan — firing condition observability. Scans for
// intent-modeling language that makes the firing condition uncheckable
// from observable artifacts.
func Item2IntentLanguageScan(_ context.Context, st *State) StepOutcome {
	contentLower := strings.ToLower(st.Content)
	for _, phrase := range intentRejectPhrases {
		if strings.Contains(contentLower, phrase) {
			return FailItemOutcome(2,
				fmt.Sprintf("firing condition contains intent-modeling language: %q", phrase))
		}
	}
	return PassOutcome()
}

// DeferredPending is the single-source deferred-item registry: why each
// deferred item cannot yet run and what would unblock it. Kept as a switch
// so the list is grep-able by item number and edits are visible at review
// time (source parity with static_steps.rs).
func DeferredPending(item int) string {
	switch item {
	case 3:
		return "item 3: duplicate check — needs ALPHABET.md corpus comparison (pending corpus access)"
	case 5:
		return "item 5: Y-not-fire positive terrain — complex multi-sub-check, decomposition cross-reference to GLYPH_DECOMPOSITION_PROCESS.md pending"
	case 6:
		return "item 6: entry coherence — retrosynthetic field-tracing to GLYPH_DEFINITION.md pending"
	case 7:
		return "item 7: sister/mirror check — corpus-wide scan pending (corpus access)"
	case 8:
		return "item 8: phenomenological grounding — multi-faceted judgment across Y-fire, Y-not-fire, and three axes; underspecified for single-agent step"
	case 11:
		return "item 11: safety-class boundary — requires proto-ethos context not yet available to the model"
	case 12:
		return "item 12: default-alignment adversarial risk — requires understanding of trained defaults"
	case 13:
		return "item 13: contamination radius — needs full corpus + cross-entry dependency graph"
	case 14:
		return "item 14: globality demand — judgment-heavy, underspecified for single-agent step"
	default:
		return "item deferred pending integration"
	}
}

// DeferredStep builds the stub StepFn for a deferred item: it emits a
// Deferred verdict with the item's pending reason. Deferred is not a
// stopping condition, so the 15-item battery registers every item in
// stratum order while integration work is outstanding.
func DeferredStep(item int) StepFn {
	return func(_ context.Context, _ *State) StepOutcome {
		return VerdictOutcome(Deferred(DeferredPending(item)))
	}
}

// Item4NA — the registry-summary requirement applies to TABOO_REGISTRY
// entries only; for glyph entries item 4 is structurally inapplicable
// (NotApplicable, not "pass with a note").
func Item4NA(_ context.Context, _ *State) StepOutcome {
	return VerdictOutcome(NotApplicable("item 4 does not apply to glyph entries"))
}

// Item10AxisPresence — three-axis coverage or honest gap notation. Requires
// a Y marker section, then Marker/Aim/Rest each present as a heading
// (`**<axis>`), a mention (`<axis> axis`), or gap notation (`<axis>:`),
// case-insensitively.
func Item10AxisPresence(_ context.Context, st *State) StepOutcome {
	content := st.Content

	hasYMarker := strings.Contains(content, "**Y") ||
		strings.Contains(content, "Y —") ||
		strings.Contains(content, "Y marker")
	if !hasYMarker {
		return FailItemOutcome(10, "no Y marker section found")
	}

	contentLower := strings.ToLower(content)
	var missing []string
	for _, axis := range []string{"Marker", "Aim", "Rest"} {
		axisLower := strings.ToLower(axis)
		found := strings.Contains(contentLower, "**"+axisLower) ||
			strings.Contains(contentLower, axisLower+" axis") ||
			strings.Contains(contentLower, axisLower+":")
		if !found {
			missing = append(missing, axis)
		}
	}

	if len(missing) == 0 {
		return PassOutcome()
	}
	return FailItemOutcome(10,
		"missing axis coverage for: "+strings.Join(missing, ", "))
}

// Item15FalloutProfile — fallout characterization: the entry must carry a
// `**Fallout profile:**` field.
func Item15FalloutProfile(_ context.Context, st *State) StepOutcome {
	if strings.Contains(st.Content, "**Fallout profile:**") {
		return PassOutcome()
	}
	return FailItemOutcome(15, "no **Fallout profile:** field found in entry")
}
