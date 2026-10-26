package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("calib", raCalib) }

// raCalib builds a blind human-calibration subset for the operator: stratified
// across (model, arm) and weighted toward the A/B over-fire disagreements, with
// both-over-fire and both-OK agreements as accuracy anchors. It writes the sheet she
// scores (CALIBRATION_SHEET.md), the blank answer template (CALIBRATION_ANSWERS.md),
// and the operator-held keymap (calib_map.json, id -> {coords, A, B}). The arm and
// both judges' labels never enter the sheet. Ported from calib.py.
//
// Usage: rest-axis calib <study-dir> --rater-a <dir> --out <dir> [--seed <n>] [--target <n>]
//
// Rater A lives outside the repo (a blind-bundle dir of <model>.<glyph>/{map,labels}.json),
// so its path is a required flag, mirroring raAnalyze.
func raCalib(args []string) error {
	raterA, args := flagValue(args, "--rater-a")
	out, args := flagValue(args, "--out")
	seedStr, args := flagValue(args, "--seed")
	targetStr, args := flagValue(args, "--target")
	if raterA == "" {
		return fmt.Errorf("need --rater-a <dir> (the rater A blind bundles: <model>.<glyph>/{map,labels}.json)")
	}
	if out == "" {
		return fmt.Errorf("need --out <dir> (where to write the sheet, answers, and calib_map.json)")
	}
	if len(args) != 1 {
		return fmt.Errorf("need a study-dir")
	}
	study := args[0]
	scores := filepath.Join(study, "scores")

	seed := int64(7) // calib.py's default seed
	if seedStr != "" {
		v, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return fmt.Errorf("bad --seed %q: %w", seedStr, err)
		}
		seed = v
	}
	params := restaxis.DefaultCalibParams()
	if targetStr != "" {
		v, err := strconv.Atoi(targetStr)
		if err != nil {
			return fmt.Errorf("bad --target %q: %w", targetStr, err)
		}
		params.Target = v
	}

	a, err := loadRaterA(raterA)
	if err != nil {
		return err
	}
	b, err := loadRaterB(scores)
	if err != nil {
		return err
	}

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // not cryptographic: a reproducible calibration draw
	items := restaxis.CalibSelect(a, b, params, rng.Shuffle)

	rendered := make([]restaxis.CalibRendered, len(items))
	gtCache := map[string]map[int]string{}
	for i, it := range items {
		k := it.Key
		gt := gtCache[k.Glyph]
		if gt == nil {
			gt = map[int]string{}
			if comp, err := os.ReadFile(filepath.Join(study, k.Glyph, "CORRECT_COMPLETIONS.md")); err == nil { //nolint:gosec // a study path
				gt = parseRestAxisCompletions(string(comp))
			}
			gtCache[k.Glyph] = gt
		}
		rendered[i] = restaxis.CalibRendered{
			Item:        it,
			Task:        restAxisScenarioText(study, k.Glyph, k.Scenario),
			GroundTruth: gt[k.Scenario],
			Response:    restAxisResponseText(study, k),
		}
	}
	doc := restaxis.RenderCalib(rendered)

	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "CALIBRATION_SHEET.md"), []byte(doc.Sheet), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "CALIBRATION_ANSWERS.md"), []byte(doc.Answers), 0o600); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(out, "calib_map.json"), doc.Map); err != nil {
		return err
	}

	c := restaxis.SummarizeCalib(doc.Map)
	fmt.Fprintf(os.Stderr, "rater A: %d | rater B: %d\n", len(a), len(b))
	fmt.Fprintf(os.Stderr, "calibration: %d items (%d disagreements, %d agreements)\n", c.Total, c.Disagree, c.Agree)
	fmt.Fprintf(os.Stderr, "by model: %s\n", sortedCounts(c.ByModel))
	fmt.Fprintf(os.Stderr, "by arm: %s\n", sortedCounts(c.ByArm))
	fmt.Fprintf(os.Stderr, "sheet -> %s\n", filepath.Join(out, "CALIBRATION_SHEET.md"))
	return nil
}

// restAxisScenarioText reads a scenario prompt from the study's materials, trimmed.
// A missing file yields "" (the calibration sheet still renders).
func restAxisScenarioText(study, glyph string, scenario int) string {
	b, err := os.ReadFile(filepath.Join(study, glyph, "materials", fmt.Sprintf("scenario_%d.md", scenario))) //nolint:gosec // a study path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// restAxisResponseText reads one response from the runs tree, trimmed. Mirrors
// calib.py's response_text: a missing file yields "(missing)".
func restAxisResponseText(study string, k restaxis.Key) string {
	p := filepath.Join(study, "runs", k.Model, k.Glyph, fmt.Sprintf("s%d", k.Scenario), "out", "responses", fmt.Sprintf("%s_%d.txt", k.Arm, k.Seed))
	b, err := os.ReadFile(p) //nolint:gosec // a study response path
	if err != nil {
		return "(missing)"
	}
	return strings.TrimSpace(string(b))
}

// sortedCounts renders a count map as "k1=v1 k2=v2 …" in key order, for stable output.
func sortedCounts(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(parts, " ")
}
