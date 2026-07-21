package battery

import (
	"context"
	"fmt"
	"path/filepath"
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

// falloutProfileMarker is the entry field Item 15 reads: the text following it
// on the same line is the reference to the fallout-profile document.
const falloutProfileMarker = "**Fallout profile:**"

// falloutDimensions are the five dimensions ALPHABET_ENTRY_BATTERY.md Item 15
// requires the referenced profile to address. A null finding on a dimension is
// a valid result, but the dimension must be explicitly present.
var falloutDimensions = []string{
	"attentional shift",
	"over-application",
	"meta-awareness",
	"scope creep",
	"suppression effects",
}

// Item15FalloutProfile — fallout characterization. The entry must carry a
// `**Fallout profile:**` field, the referenced document must resolve and be
// readable, and it must address all five fallout dimensions. This is a
// characterization gate: only presence and dimensional coverage are checked,
// never the correctness of the analysis.
//
// The step performs pure path math (resolving the reference relative to the
// entry file) and delegates the read to the injected ProfileReader, so the
// package stays sans-IO. It fails closed on a missing field, a missing reader,
// an unreadable referent, or any unaddressed dimension — a battery must never
// manufacture a pass it did not verify, which is exactly what the old
// substring-only check did for the whole window the referents were deleted.
func Item15FalloutProfile(_ context.Context, st *State) StepOutcome {
	ref, ok := falloutProfileRef(st.Content)
	if !ok {
		return FailItemOutcome(15, "no **Fallout profile:** field found in entry")
	}
	if st.Profiles == nil {
		return FailItemOutcome(15,
			fmt.Sprintf("no profile reader configured to resolve %q", ref))
	}

	resolved := ref
	if st.EntryPath != "" {
		resolved = filepath.Join(filepath.Dir(st.EntryPath), ref)
	}

	doc, err := st.Profiles.ReadProfile(resolved)
	if err != nil {
		return FailItemOutcome(15,
			fmt.Sprintf("fallout profile %q is missing or unreadable: %v", resolved, err))
	}

	if missing := missingFalloutDimensions(doc); len(missing) > 0 {
		return FailItemOutcome(15,
			"fallout profile does not address: "+strings.Join(missing, ", "))
	}
	return PassOutcome()
}

// falloutProfileRef extracts the reference following the `**Fallout profile:**`
// marker on its line, returning false when the field is absent or empty.
func falloutProfileRef(content string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		idx := strings.Index(line, falloutProfileMarker)
		if idx < 0 {
			continue
		}
		ref := strings.TrimSpace(line[idx+len(falloutProfileMarker):])
		if ref == "" {
			continue
		}
		return ref, true
	}
	return "", false
}

// missingFalloutDimensions returns the dimensions the profile does not address,
// in canonical order; an empty result means all five are covered.
func missingFalloutDimensions(doc string) []string {
	var missing []string
	for _, dim := range falloutDimensions {
		if !dimensionAddressed(doc, dim) {
			missing = append(missing, dim)
		}
	}
	return missing
}

// dimensionAddressed reports whether the profile carries a section heading for
// the dimension in either form the restored profiles use: a bold run
// (`**Attentional shift.**`) or an ATX heading (`### Attentional shift`), at
// the start of a line. Matching the heading forms — not a bare substring — is
// deliberate: every profile opens with a five-item definition preamble that
// lists all five dimensions as `- **<dim>:**` list items, so a substring scan
// would report full coverage even for a profile whose analysis omits a
// dimension. List items begin with "-", so anchoring on a leading "###" or
// "**" excludes the preamble and matches only genuine section headings.
func dimensionAddressed(doc, dim string) bool {
	for _, line := range strings.Split(doc, "\n") {
		trimmed := strings.TrimSpace(line)
		var heading string
		switch {
		case strings.HasPrefix(trimmed, "###"):
			heading = strings.TrimLeft(trimmed, "#")
		case strings.HasPrefix(trimmed, "**"):
			heading = strings.TrimLeft(trimmed, "*")
		default:
			continue
		}
		heading = strings.ToLower(strings.TrimSpace(heading))
		if strings.HasPrefix(heading, dim) {
			return true
		}
	}
	return false
}
