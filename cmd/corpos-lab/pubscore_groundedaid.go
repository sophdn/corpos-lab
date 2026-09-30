package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"corpos-lab/internal/groundedaid"
)

// pgAggregate ports scoring/aggregate.py's IO: it reads the private key and one or
// two raters' score files (a dir of *.json or a single file) and prints the
// agreement stats and the pooled / per-cell C-rate tables.
//
// Usage: corpos-lab pub-score grounded-aggregate --key <key.json> --rater-a <path> [--rater-b <path>]
func pgAggregate(args []string) error {
	keyPath, args := flagValue(args, "--key")
	raterA, args := flagValue(args, "--rater-a")
	raterB, args := flagValue(args, "--rater-b")
	_ = args
	if keyPath == "" || raterA == "" {
		return fmt.Errorf("usage: corpos-lab pub-score grounded-aggregate --key <key.json> --rater-a <path> [--rater-b <path>]")
	}
	var key map[string]groundedaid.KeyEntry
	if err := readJSON(keyPath, &key); err != nil {
		return err
	}
	a, err := pgLoadScores(raterA)
	if err != nil {
		return err
	}
	var b map[string]string
	if raterB != "" {
		if b, err = pgLoadScores(raterB); err != nil {
			return err
		}
	}
	fmt.Print(groundedaid.Aggregate(key, a, b))
	return nil
}

// pgLoadScores reads a rater's codes (id -> code), from a dir of *.json files or a
// single json file, matching aggregate.py's load_scores.
func pgLoadScores(path string) (map[string]string, error) {
	var files []string
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		if files, err = filepath.Glob(filepath.Join(path, "*.json")); err != nil {
			return nil, err
		}
	} else {
		files = []string{path}
	}
	out := map[string]string{}
	for _, f := range files {
		var codes map[string]string
		if err := readJSON(f, &codes); err != nil {
			return nil, err
		}
		for id, code := range codes {
			out[id] = code
		}
	}
	return out, nil
}

var pgCompletedRe = regexp.MustCompile(`"status"\s*:\s*"completed"`)

type pgResults struct {
	Rows []struct {
		Condition string `json:"condition"`
		Run       int    `json:"run"`
	} `json:"rows"`
}

// pgBuildSlices ports scoring/build_slices.py's IO: it scans the study's completed
// runs, pairs each result row with its response text, and writes the blind slices,
// the id->key map, and the coverage manifest under <root>/scoring/.
//
// Usage: corpos-lab pub-score grounded-build-slices --root <study-root> [--per-slice 96] [--seed 12345]
func pgBuildSlices(args []string) error {
	root, args := flagValue(args, "--root")
	perSliceStr, args := flagValue(args, "--per-slice")
	seedStr, args := flagValue(args, "--seed")
	_ = args
	if root == "" {
		return fmt.Errorf("usage: corpos-lab pub-score grounded-build-slices --root <study-root> [--per-slice 96] [--seed 12345]")
	}
	perSlice := groundedaid.DefaultPerSlice
	if perSliceStr != "" {
		v, err := strconv.Atoi(perSliceStr)
		if err != nil {
			return fmt.Errorf("--per-slice: %w", err)
		}
		perSlice = v
	}
	seed := int64(groundedaid.DefaultSeed)
	if seedStr != "" {
		v, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return fmt.Errorf("--seed: %w", err)
		}
		seed = v
	}

	resultFiles, err := filepath.Glob(filepath.Join(root, "*", "runs", "gnp-*", "out", "results.json"))
	if err != nil {
		return err
	}
	sort.Strings(resultFiles)

	var rows []groundedaid.RawRow
	for _, results := range resultFiles {
		runDir := filepath.Dir(filepath.Dir(results))
		cls, sc, model, ok := groundedaid.ParseRunDir(filepath.Base(runDir))
		if !ok {
			continue
		}
		rec, err := os.ReadFile(filepath.Join(runDir, "run-record.json")) //nolint:gosec // a study run path
		if err != nil || !pgCompletedRe.Match(rec) {
			continue
		}
		var rj pgResults
		if err := readJSON(results, &rj); err != nil {
			return err
		}
		for _, r := range rj.Rows {
			resp := filepath.Join(runDir, "out", "responses", fmt.Sprintf("%s_%d.txt", r.Condition, r.Run))
			b, err := os.ReadFile(resp) //nolint:gosec // a study response path
			if err != nil {
				continue // missing response: skip, matching build_slices.py
			}
			rows = append(rows, groundedaid.RawRow{
				Cls: cls, Sc: sc, Model: model, Cond: r.Condition, Run: r.Run, Text: string(b),
			})
		}
	}

	res := groundedaid.BuildSlices(rows, perSlice, seed)
	outDir := filepath.Join(root, "scoring")
	if err := os.MkdirAll(filepath.Join(outDir, "slices"), 0o750); err != nil {
		return err
	}
	for _, s := range res.Slices {
		lines := make([]map[string]any, 0, len(s.Items))
		for _, it := range s.Items {
			lines = append(lines, map[string]any{"id": it.ID, "text": it.Text})
		}
		path := filepath.Join(outDir, "slices", fmt.Sprintf("%s__%02d.jsonl", s.Class, s.Index))
		if err := writeJSONL(path, lines); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(outDir, "key.json"), res.Key); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outDir, "MANIFEST.json"), res.Manifest); err != nil {
		return err
	}
	fmt.Print(groundedaid.FormatSummary(res))
	return nil
}

func init() {
	registerPubScoreMode("grounded-aggregate", pgAggregate)
	registerPubScoreMode("grounded-build-slices", pgBuildSlices)
}
