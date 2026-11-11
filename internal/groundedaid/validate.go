package groundedaid

import (
	"fmt"
	"sort"
)

// RunDirNameShape is the run-dir name shape ParseRunDir accepts, named in the
// mismatch errors so a caller sees why nothing parsed.
const RunDirNameShape = "gnp-<class>-s<scenario>-<model>"

// CheckBuildInputs guards grounded-build-slices against emitting a success-shaped
// empty (or silently partial) result. dirsFound is the number of run-dir result
// files the glob matched; dirsParsed is how many of those run-dir names
// ParseRunDir accepted; rows are the completed-run rows built from the parsed
// dirs. It returns a named, non-nil error when:
//
//   - the glob matched run dirs but ParseRunDir matched none of them (naming the
//     dirs-found-vs-parsed counts);
//   - zero rows were built (nothing to score);
//   - a parsed row carries a class outside the reported Classes set, which the
//     per-class slice loop in BuildSlices would silently drop (naming the
//     unknown classes vs the reported set).
//
// A correctly-matching study — at least one row, every row's class in Classes —
// returns nil, so the scored output is unchanged.
func CheckBuildInputs(dirsFound, dirsParsed int, rows []RawRow) error {
	if dirsFound > 0 && dirsParsed == 0 {
		return fmt.Errorf("grounded-build-slices: glob matched %d run dir(s) but ParseRunDir matched none of them (expected names like %q); nothing was scored", dirsFound, RunDirNameShape)
	}
	if len(rows) == 0 {
		return fmt.Errorf("grounded-build-slices: built 0 items (%d run dir(s) found, %d parsed); nothing was scored", dirsFound, dirsParsed)
	}
	if unknown := unknownClasses(rows); len(unknown) > 0 {
		return fmt.Errorf("grounded-build-slices: parsed run(s) carry class(es) %v absent from the reported class set %v; those rows would be silently dropped", unknown, Classes)
	}
	return nil
}

// unknownClasses returns the distinct row classes not in Classes, sorted.
func unknownClasses(rows []RawRow) []string {
	known := map[string]bool{}
	for _, c := range Classes {
		known[c] = true
	}
	seen := map[string]bool{}
	var out []string
	for _, r := range rows {
		if !known[r.Cls] && !seen[r.Cls] {
			seen[r.Cls] = true
			out = append(out, r.Cls)
		}
	}
	sort.Strings(out)
	return out
}

// CheckAggregateKey guards grounded-aggregate against silently dropping data from
// its pooled and per-cell tables, which iterate the hardcoded Classes and Conds.
// It returns a named error when the key is empty, or when a key entry carries a
// class or condition outside those reported sets (naming the offenders vs the
// reported lists). A key whose every entry is in Classes and Conds returns nil,
// so the byte-identical aggregate output is unchanged.
func CheckAggregateKey(key map[string]KeyEntry) error {
	if len(key) == 0 {
		return fmt.Errorf("grounded-aggregate: key has 0 entries; nothing to aggregate")
	}
	knownCls := map[string]bool{}
	for _, c := range Classes {
		knownCls[c] = true
	}
	knownCond := map[string]bool{}
	for _, c := range Conds {
		knownCond[c] = true
	}
	clsSeen, condSeen := map[string]bool{}, map[string]bool{}
	var badCls, badCond []string
	for _, m := range key {
		if !knownCls[m.Cls] && !clsSeen[m.Cls] {
			clsSeen[m.Cls] = true
			badCls = append(badCls, m.Cls)
		}
		if !knownCond[m.Condition] && !condSeen[m.Condition] {
			condSeen[m.Condition] = true
			badCond = append(badCond, m.Condition)
		}
	}
	if len(badCls) == 0 && len(badCond) == 0 {
		return nil
	}
	sort.Strings(badCls)
	sort.Strings(badCond)
	return fmt.Errorf("grounded-aggregate: key contains class(es) %v / condition(s) %v absent from the reported classes %v / conditions %v; those rows would be silently dropped from the tables", badCls, badCond, Classes, Conds)
}
