// Package studyclose reports the mechanical study-record facts INQUIRY.md keeps
// as human discipline, at study close. It ADVISES and never blocks: it prints
// which mechanical facts are present or missing and says nothing about
// scientific quality — the praxist "Local Reviewer" shape.
//
// It exists because those steps are in the exact shape corpos-lab already burned
// itself on: a documented invariant the run path never checked (the retired
// charter's manifest.Verify). This turns "reconcile the library, double-score,
// run the mechanism controls" from prose no code reads into a checklist that
// reports — without becoming a gate (suggestion
// wire-an-advisory-study-close-check-for-the-methodology-steps).
package studyclose

import (
	"fmt"
	"sort"
	"strings"
)

// Status is a finding's disposition. Every status is informational — none blocks.
type Status string

const (
	// StatusPresent: the mechanical fact is present in the study record.
	StatusPresent Status = "present"
	// StatusAbsent: the fact is absent; the finding advises what to add.
	StatusAbsent Status = "absent"
	// StatusFlag: the fact is present but something about it needs a human look.
	StatusFlag Status = "flag"
)

// Finding is one checked mechanical fact.
type Finding struct {
	Check  string
	Status Status
	Detail string
}

// Inputs are the mechanical facts a caller gathers from a study directory. The
// gathering (walk the dir, read the tomls, stat the markers) is I/O the command
// does; the checking logic here is pure so it is tested without a filesystem.
type Inputs struct {
	// Conditions is the union of conditions declared across the study's defs.
	Conditions []string
	// RunStatuses is the status of each run-record found (e.g. "completed").
	RunStatuses []string
	// SecondRaters lists the distinct rater ids found beyond the primary scorer.
	SecondRaters []string
	// ReconciliationMarker is true when a library-reconciliation record is present.
	ReconciliationMarker bool
	// PredictionsFile is true when a predictions file was found in the study dir.
	PredictionsFile bool
}

// The mechanism-control conditions INQUIRY.md names, and the second rater that
// cleared the capability floor versus the one that failed it.
const (
	scrambledControl  = "scrambled_glyph"
	offTargetControl  = "off_target_glyph"
	floorClearedRater = "phi-4-14B"
	floorFailedRater  = "mistral"
)

func has(conditions []string, want string) bool {
	for _, c := range conditions {
		if c == want {
			return true
		}
	}
	return false
}

// Report returns one finding per mechanical study-close fact. The findings are in
// a fixed order so the output is stable. None of them blocks.
func Report(in Inputs) []Finding {
	var out []Finding

	// Run records completed.
	completed := 0
	for _, s := range in.RunStatuses {
		if s == "completed" {
			completed++
		}
	}
	switch {
	case len(in.RunStatuses) == 0:
		out = append(out, Finding{"run records", StatusAbsent, "no run-record found in the study dir"})
	case completed == len(in.RunStatuses):
		out = append(out, Finding{"run records", StatusPresent,
			fmt.Sprintf("%d run(s), all completed", len(in.RunStatuses))})
	default:
		out = append(out, Finding{"run records", StatusFlag,
			fmt.Sprintf("%d of %d run(s) completed; the rest failed", completed, len(in.RunStatuses))})
	}

	// Mechanism controls.
	scr, off := has(in.Conditions, scrambledControl), has(in.Conditions, offTargetControl)
	switch {
	case scr && off:
		out = append(out, Finding{"mechanism controls", StatusPresent,
			"both scrambled_glyph and off_target_glyph are among the conditions"})
	case scr || off:
		present, missing := scrambledControl, offTargetControl
		if off {
			present, missing = offTargetControl, scrambledControl
		}
		out = append(out, Finding{"mechanism controls", StatusFlag,
			fmt.Sprintf("%s present, %s absent — run both when a mechanism or format claim is on the line", present, missing)})
	default:
		out = append(out, Finding{"mechanism controls", StatusAbsent,
			"neither scrambled_glyph nor off_target_glyph is a condition — run them when a mechanism or format claim is on the line"})
	}

	// Second rater and the capability floor.
	var failed []string
	for _, r := range in.SecondRaters {
		if strings.Contains(strings.ToLower(r), floorFailedRater) {
			failed = append(failed, r)
		}
	}
	switch {
	case len(in.SecondRaters) == 0:
		out = append(out, Finding{"second rater", StatusAbsent,
			"no second rater found — double-score with one that cleared the capability floor (default " + floorClearedRater + ")"})
	case len(failed) > 0:
		sort.Strings(failed)
		out = append(out, Finding{"second rater", StatusFlag,
			fmt.Sprintf("rater(s) %s did not clear the floor (Mistral-7B failed it); confirm the second rater is floor-cleared (default %s)",
				strings.Join(failed, ", "), floorClearedRater)})
	default:
		out = append(out, Finding{"second rater", StatusPresent,
			fmt.Sprintf("%d second rater(s): %s — confirm each cleared the capability floor", len(in.SecondRaters), strings.Join(in.SecondRaters, ", "))})
	}

	// Library reconciliation.
	if in.ReconciliationMarker {
		out = append(out, Finding{"library reconciled", StatusPresent, "a reconciliation record is present"})
	} else {
		out = append(out, Finding{"library reconciled", StatusAbsent,
			"no reconciliation record — record confirmed/refuted/partially-held and update the entry; the library data lives in seed-packet library-data/entries/"})
	}

	// Predictions held out of subject/judge view.
	if in.PredictionsFile {
		out = append(out, Finding{"predictions", StatusFlag,
			"a predictions file is present — confirm it is in no file a subject or judge model can see"})
	}

	return out
}
