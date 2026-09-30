package main

import (
	"fmt"
	"os"
	"strconv"

	"corpos-lab/internal/lengthdistraction"
	"corpos-lab/internal/neutralprefix"
)

// The length-distraction typing tool (task 4225, chain glyph-mechanism-taxonomy).
// It reads the same scoring artifacts the neutral-prefix tools write — key.json and
// the per-rater code files — but reports the OFF-TASK N rate per class x condition,
// which neutral-aggregate omits, and applies the length-distraction typing
// criterion. The classes come from the key, not a hardcoded list, so it types any
// grid the neutral-prefix builder produced.
func init() {
	registerPubScoreMode("length-aggregate", ldAggregate)
}

// ldAggregate reads key.json and one, two, or three rater code maps and prints the
// length-distraction typing report: N-rate and C-rate per class x condition, then a
// per-class verdict. --hi-n and --lo-n set the criterion thresholds (defaults 0.30
// and 0.10): a class is length-distraction when neutral N >= hi-n and imperative
// N <= lo-n. The thresholds are printed in the report header so a reader sees the
// rule that produced the verdicts.
//
// With three raters (--rater-a/-b/-c) the report reads per-id majority consensus
// across the families (>=2 of 3 agree; no majority folds to split). With two it is
// strict two-rater consensus; with one it is that rater's own codes.
//
//	pub-score length-aggregate --key <key.json> --rater-a <path> [--rater-b <path>] [--rater-c <path>] [--hi-n 0.30] [--lo-n 0.10]
func ldAggregate(args []string) error {
	keyPath, args := flagValue(args, "--key")
	raterA, args := flagValue(args, "--rater-a")
	raterB, args := flagValue(args, "--rater-b")
	raterC, args := flagValue(args, "--rater-c")
	hiNStr, args := flagValue(args, "--hi-n")
	loNStr, args := flagValue(args, "--lo-n")
	_ = args
	if keyPath == "" || raterA == "" {
		return fmt.Errorf("need --key <key.json> and --rater-a <dir|file>")
	}
	hiN, err := floatOr(hiNStr, 0.30)
	if err != nil {
		return fmt.Errorf("--hi-n: %w", err)
	}
	loN, err := floatOr(loNStr, 0.10)
	if err != nil {
		return fmt.Errorf("--lo-n: %w", err)
	}

	keyB, err := os.ReadFile(keyPath) //nolint:gosec // study path
	if err != nil {
		return err
	}
	key, err := neutralprefix.ParseKeyFile(keyB)
	if err != nil {
		return err
	}
	a, err := loadScores(raterA)
	if err != nil {
		return err
	}

	conds := []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}
	classes := lengthdistraction.SortedClasses(key)

	switch {
	case raterC != "":
		b, err := loadScores(raterB)
		if err != nil {
			return err
		}
		c, err := loadScores(raterC)
		if err != nil {
			return err
		}
		// Majority consensus across three families, passed as both a and b so the
		// strict-consensus path reports the majority code per id.
		maj := lengthdistraction.MajorityMap(key, a, b, c)
		fmt.Print(lengthdistraction.Report(key, maj, maj, true, classes, conds, hiN, loN))
	case raterB != "":
		b, err := loadScores(raterB)
		if err != nil {
			return err
		}
		fmt.Print(lengthdistraction.Report(key, a, b, true, classes, conds, hiN, loN))
	default:
		fmt.Print(lengthdistraction.Report(key, a, nil, false, classes, conds, hiN, loN))
	}
	return nil
}

// floatOr parses s as a float, returning def when s is empty.
func floatOr(s string, def float64) (float64, error) {
	if s == "" {
		return def, nil
	}
	return strconv.ParseFloat(s, 64)
}
