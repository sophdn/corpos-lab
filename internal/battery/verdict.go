// Package battery ports lab-app's ALPHABET Entry Battery: typed verdicts,
// the fail-fast sequence runner, the 15-item battery assembly (6 implemented
// items, 9 deferred stubs), and the inference steps that score entries via a
// local model. Behavioral contract inherited from the Rust source at
// lab-app@ae6611d; see docs/PORT_BATTERY_INVENTORY.md for what was ported,
// dropped, and why.
package battery

import (
	"fmt"
)

// VerdictKind discriminates Verdict variants. Values match the Rust serde
// tag strings so the JSON shape is stable across the port.
type VerdictKind string

// Verdict kinds, in the source's declaration order.
const (
	KindPass              VerdictKind = "pass"
	KindPassWithCondition VerdictKind = "pass_with_condition"
	KindFlag              VerdictKind = "flag"
	KindDeferred          VerdictKind = "deferred"
	KindFail              VerdictKind = "fail"
	KindNotApplicable     VerdictKind = "not_applicable"
)

// Verdict is the canonical typed pass/fail decision for a single evaluated
// item. Construct via the Pass/Fail/Flag/Deferred/NotApplicable helpers so
// invalid combinations (a pass with a failure reason, a deferral without a
// pending note) don't occur.
type Verdict struct {
	Kind VerdictKind `json:"kind"`
	// Condition accompanies pass_with_condition.
	Condition string `json:"condition,omitempty"`
	// Reason accompanies flag and fail.
	Reason string `json:"reason,omitempty"`
	// Pending accompanies deferred.
	Pending string `json:"pending,omitempty"`
	// Item is the 1-based battery item for fail verdicts, nil when no item
	// number applies (non-battery contexts).
	Item *int `json:"item,omitempty"`
	// Note accompanies not_applicable.
	Note string `json:"note,omitempty"`
}

// Pass returns a passing verdict.
func Pass() Verdict { return Verdict{Kind: KindPass} }

// PassWithCondition returns a pass carrying a noted condition.
func PassWithCondition(condition string) Verdict {
	return Verdict{Kind: KindPassWithCondition, Condition: condition}
}

// Flag returns an advisory flag verdict (soft fail; does not stop a sequence).
func Flag(reason string) Verdict { return Verdict{Kind: KindFlag, Reason: reason} }

// Deferred returns a cannot-judge-yet verdict with its pending reason.
func Deferred(pending string) Verdict { return Verdict{Kind: KindDeferred, Pending: pending} }

// Fail returns a failing verdict with no battery item number.
func Fail(reason string) Verdict { return Verdict{Kind: KindFail, Reason: reason} }

// FailItem returns a failing verdict at a specific 1-based battery item.
func FailItem(item int, reason string) Verdict {
	return Verdict{Kind: KindFail, Item: &item, Reason: reason}
}

// NotApplicable returns a structurally-inapplicable verdict. Distinct from
// PassWithCondition: there is no pass/fail judgement to make.
func NotApplicable(note string) Verdict { return Verdict{Kind: KindNotApplicable, Note: note} }

// Passed reports whether the verdict is a passing outcome (pass,
// pass_with_condition, or not_applicable).
func (v Verdict) Passed() bool {
	switch v.Kind {
	case KindPass, KindPassWithCondition, KindNotApplicable:
		return true
	case KindFlag, KindDeferred, KindFail:
		return false
	}
	return false
}

// String renders the ALPHABET_ENTRY_BATTERY.md notation, matching the Rust
// Display impl verbatim.
func (v Verdict) String() string {
	switch v.Kind {
	case KindPass:
		return "PASS"
	case KindPassWithCondition:
		return fmt.Sprintf("PASS* (%s)", v.Condition)
	case KindFlag:
		return "FLAG: " + v.Reason
	case KindDeferred:
		return "DEFERRED: " + v.Pending
	case KindFail:
		if v.Item != nil {
			return fmt.Sprintf("FAIL at item %d: %s", *v.Item, v.Reason)
		}
		return "FAIL: " + v.Reason
	case KindNotApplicable:
		return "N/A: " + v.Note
	}
	return string(v.Kind)
}

// FailureClassKind discriminates FailureClass values.
type FailureClassKind string

// Failure classifications for battery failures.
const (
	ClassExecutorCorrectable FailureClassKind = "executor_correctable"
	ClassDefinitionGap       FailureClassKind = "definition_gap"
	ClassUnknown             FailureClassKind = "unknown"
)

// FailureClass classifies a battery failure for routing: correctable by the
// executor vs a gap in the definition itself. Detail carries the free-form
// note the Rust Unknown(String) variant held.
type FailureClass struct {
	Kind   FailureClassKind `json:"kind"`
	Detail string           `json:"detail,omitempty"`
}

// Failure is one classified failing item inside a Block verdict.
type Failure struct {
	// Item is the failing battery item (1-based; 0 when no item applies).
	Item int `json:"item"`
	// Reason is the human-readable failure reason.
	Reason string `json:"reason"`
	// Class routes the failure (definition stop vs re-queue).
	Class FailureClass `json:"class"`
}

// RunVerdictKind discriminates RunVerdict variants.
type RunVerdictKind string

// Battery-level outcomes.
const (
	RunPromote RunVerdictKind = "promote"
	RunBlock   RunVerdictKind = "block"
	RunDefer   RunVerdictKind = "defer"
)

// RunVerdict is the overall outcome of a battery run, composed from
// per-item verdicts: promote if every item passed (or was N/A), block if any
// item failed, defer if any item deferred and none failed.
type RunVerdict struct {
	Kind RunVerdictKind `json:"kind"`
	// Failures carries one entry per failing item (block only).
	Failures []Failure `json:"failures,omitempty"`
	// Pending carries one entry per deferred item (defer only).
	Pending []string `json:"pending,omitempty"`
}

// IsPromote reports whether the verdict is Promote.
func (b RunVerdict) IsPromote() bool { return b.Kind == RunPromote }

// String matches the Rust Display impl: PROMOTE / BLOCK (N failure[s]) /
// DEFER (N pending).
func (b RunVerdict) String() string {
	switch b.Kind {
	case RunPromote:
		return "PROMOTE"
	case RunBlock:
		plural := "s"
		if len(b.Failures) == 1 {
			plural = ""
		}
		return fmt.Sprintf("BLOCK (%d failure%s)", len(b.Failures), plural)
	case RunDefer:
		return fmt.Sprintf("DEFER (%d pending)", len(b.Pending))
	}
	return string(b.Kind)
}

// ComposeRunVerdict folds per-item verdicts into a battery-level
// verdict. Precedence: Block wins over Defer wins over Promote. Flags are
// soft-fails included in Block's failure list (class executor_correctable)
// so reviewers see them alongside hard fails.
func ComposeRunVerdict(items []Verdict) RunVerdict {
	var failures []Failure
	var pending []string
	for _, v := range items {
		switch v.Kind {
		case KindFail:
			item := 0
			if v.Item != nil {
				item = *v.Item
			}
			failures = append(failures, Failure{
				Item:   item,
				Reason: v.Reason,
				Class:  FailureClass{Kind: ClassUnknown},
			})
		case KindFlag:
			failures = append(failures, Failure{
				Item:   0,
				Reason: "FLAG: " + v.Reason,
				Class:  FailureClass{Kind: ClassExecutorCorrectable},
			})
		case KindDeferred:
			pending = append(pending, v.Pending)
		case KindPass, KindPassWithCondition, KindNotApplicable:
		}
	}
	switch {
	case len(failures) > 0:
		return RunVerdict{Kind: RunBlock, Failures: failures}
	case len(pending) > 0:
		return RunVerdict{Kind: RunDefer, Pending: pending}
	default:
		return RunVerdict{Kind: RunPromote}
	}
}

// FailureRoute is the chain action a set of battery failures routes to.
type FailureRoute string

// Failure routes. PromoteAnyway exists as a route value but is never
// returned by RouteChainFailure — it requires an explicit operator override.
const (
	RouteStopAndStudyDefinition FailureRoute = "stop_and_study_definition"
	RouteRequeueToResearcher    FailureRoute = "requeue_to_researcher"
	RoutePromoteAnyway          FailureRoute = "promote_anyway"
)

// RouteChainFailure routes battery failures: any definition-gap or unknown
// classification stops the chain for definition study (conservative);
// all-executor-correctable re-queues; an empty failure set re-queues.
func RouteChainFailure(failures []Failure) FailureRoute {
	for _, f := range failures {
		if f.Class.Kind == ClassDefinitionGap || f.Class.Kind == ClassUnknown {
			return RouteStopAndStudyDefinition
		}
	}
	return RouteRequeueToResearcher
}
