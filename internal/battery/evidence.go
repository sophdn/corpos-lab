package battery

import "fmt"

// EvidenceState is the maturity of a study artifact, made a typed fact the code
// reads instead of prose plus a file location. It is the state machine behind the
// candidate → ALPHABET promotion the docs describe as human discipline
// (suggestion give-corpos-lab-study-artifacts-typed-evidence-states). The rule
// that earns it its place: the code reads the state and runs promotion through
// one gate, so maturity is inspectable rather than reconstructed from where a
// file sits and what git remembers.
type EvidenceState string

const (
	// StateCandidate is identified but not yet battery-certified — the corpus
	// candidates stage.
	StateCandidate EvidenceState = "candidate"
	// StateCertified passed the 15-item entry battery; an ALPHABET member.
	StateCertified EvidenceState = "certified"
	// StateDerivedView is a regenerable projection of certified evidence (a table,
	// a figure). It never certifies; regenerate it rather than trust it as source.
	StateDerivedView EvidenceState = "derived_view"
	// StateAuditSnapshot is a frozen record of a past state, kept for audit. It
	// never certifies and never changes.
	StateAuditSnapshot EvidenceState = "audit_snapshot"
)

// Valid reports whether s is a known evidence state.
func (s EvidenceState) Valid() bool {
	switch s {
	case StateCandidate, StateCertified, StateDerivedView, StateAuditSnapshot:
		return true
	default:
		return false
	}
}

// PromoteOn applies the one promotion rule to an artifact in state `from`, given
// the battery RunVerdict of a run against it. It returns the resulting state and
// a human-readable note describing the transition.
//
// The rule:
//   - A candidate that the battery PROMOTES (all 15 items pass or are N/A)
//     certifies. A DEFER (items pending an assessor) or a BLOCK leaves it a
//     candidate — the mechanized battery alone does not certify.
//   - A certified artifact that BLOCKS on a re-run is DEMOTED to candidate. A
//     tightened battery demotes rather than grandfathers, the 2026-04-03 model:
//     old passes that no longer meet the bar lose their standing and say so.
//   - A derived view or an audit snapshot is not battery-eligible; PromoteOn
//     errors rather than silently certify something that must never certify.
func PromoteOn(from EvidenceState, v RunVerdict) (EvidenceState, string, error) {
	switch from {
	case StateDerivedView, StateAuditSnapshot:
		return from, "", fmt.Errorf("battery: %s is not battery-certifiable", from)
	case StateCandidate:
		switch v.Kind {
		case RunPromote:
			return StateCertified, "promoted: 15-item battery PROMOTE", nil
		case RunDefer:
			return StateCandidate, fmt.Sprintf("stays candidate: %s pending an assessor", v.String()), nil
		default:
			return StateCandidate, "stays candidate: battery " + v.String(), nil
		}
	case StateCertified:
		switch v.Kind {
		case RunPromote:
			return StateCertified, "recertified", nil
		case RunBlock:
			return StateCandidate, "DEMOTED to candidate: battery " + v.String() +
				" (a tightened battery demotes, the 2026-04-03 model)", nil
		default:
			return StateCertified, "stays certified: re-run " + v.String(), nil
		}
	default:
		return from, "", fmt.Errorf("battery: unknown evidence state %q", from)
	}
}
