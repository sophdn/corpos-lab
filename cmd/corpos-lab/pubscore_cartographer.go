package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"corpos-lab/internal/cartographer"
)

// This file wires the cartographer-duty-format provenance-scoring python
// (provenance/published-scoring/cartographer-duty-format/scores/*.py) to the pure
// logic in internal/cartographer, under `corpos-lab pub-score cartographer-*`.
// Each mode self-registers from init(); the IO — reading the study's JSON and
// response trees, writing the blind set / grids, printing the reports — lives
// here, the scoring lives in the package. Chain eradicate-python-rewrite-in-go.

// cartoFName matches a produced-response filename: "<condition>_<run>.txt".
// Ported from build_blind_set.FNAME / slug_c2.FNAME.
var cartoFName = regexp.MustCompile(`^([a-z_]+)_(\d+)\.txt$`)

const cartoDefaultSeed = 20260912

// raCartographerAnalyze ports analyze.py: it reads <study>/taboo_set.json and
// <study>/scores/coverage_grid.json and prints the coverage-rate, interaction,
// scan-recovery, and per-taboo report.
func raCartographerAnalyze(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score cartographer-analyze <study-dir>")
	}
	study := args[0]
	taboos, err := cartoReadTaboos(study)
	if err != nil {
		return err
	}
	grid, err := cartoReadGrid(filepath.Join(study, "scores", "coverage_grid.json"))
	if err != nil {
		return err
	}
	fmt.Print(cartographer.Analyze(taboos, grid))
	return nil
}

// raCartographerAnalyzeDuty ports analyze_duty.py: the duty-level, clustering-
// robust reanalysis plus the rater-free keyword corroboration. It reads the same
// taboo set and grid plus the per-run response texts under
// <study>/runs/qwen38/out/responses. The two permutation p-values are seeded Go
// Monte Carlo estimates (structural parity), every other line is exact.
func raCartographerAnalyzeDuty(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score cartographer-analyze-duty <study-dir>")
	}
	study := args[0]
	taboos, err := cartoReadTaboos(study)
	if err != nil {
		return err
	}
	grid, err := cartoReadGrid(filepath.Join(study, "scores", "coverage_grid.json"))
	if err != nil {
		return err
	}
	respDir := filepath.Join(study, "runs", "qwen38", "out", "responses")
	responses, err := cartoReadDutyResponses(respDir)
	if err != nil {
		return err
	}
	fmt.Print(cartographer.AnalyzeDuty(taboos, grid, responses,
		cartographer.DefaultPermReps, cartographer.DefaultPermSeed))
	return nil
}

// raCartographerBuildBlindSet ports build_blind_set.py: it reads a run's
// out/responses, shuffles the duties under a seed, writes each to
// <study>/scores/blind/<id>.md, and records the id->(condition, run) mapping in
// <study>/scores/blind_map.json.
func raCartographerBuildBlindSet(args []string) error {
	seedStr, rest := flagValue(args, "--seed")
	if len(rest) != 2 {
		return fmt.Errorf("usage: corpos-lab pub-score cartographer-build-blind-set <run-dir> <study-dir> [--seed N]")
	}
	runDir, study := rest[0], rest[1]
	seed := int64(cartoDefaultSeed)
	if seedStr != "" {
		n, err := strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return fmt.Errorf("--seed: %w", err)
		}
		seed = n
	}
	duties, err := cartoReadRunDuties(runDir)
	if err != nil {
		return err
	}
	results := cartographer.BuildBlindSet(duties, seed)

	blindDir := filepath.Join(study, "scores", "blind")
	if err := os.MkdirAll(blindDir, 0o750); err != nil {
		return err
	}
	mapping := make(map[string]map[string]any, len(results))
	for _, r := range results {
		if err := os.WriteFile(filepath.Join(blindDir, r.ID+".md"), []byte(r.Text), 0o600); err != nil {
			return err
		}
		mapping[r.ID] = map[string]any{"condition": r.Condition, "run": r.Run}
	}
	if err := writeJSON(filepath.Join(study, "scores", "blind_map.json"), mapping); err != nil {
		return err
	}
	fmt.Printf("wrote %d blind duties to %s; map -> scores/blind_map.json\n", len(results), blindDir)
	return nil
}

// raCartographerSlugC2 ports slug_c2.py: the deterministic taboo-slug citation
// scorer. It reads every response under each run dir's out/responses, counts
// exact slug citations, prints a per-condition summary, and writes slug_c2.json
// next to the first run dir.
func raCartographerSlugC2(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: corpos-lab pub-score cartographer-slug-c2 <run-dir> [<run-dir> ...]")
	}
	grid := map[string]cartographer.SlugC2Row{}
	var rows []cartographer.SlugC2Row
	for _, rd := range args {
		duties, err := cartoReadRunDuties(rd)
		if err != nil {
			return err
		}
		for _, d := range duties {
			hits := cartographer.CitedSlugs(d.Text)
			row := cartographer.SlugC2Row{Condition: d.Condition, Run: d.Run, NSlugs: len(hits), Slugs: hits}
			grid[fmt.Sprintf("%s_%d", d.Condition, d.Run)] = row
			rows = append(rows, row)
		}
	}
	summary, table := cartographer.SlugC2Summary(rows)
	fmt.Print(table)
	out := filepath.Join(args[0], "slug_c2.json")
	if err := writeJSON(out, map[string]any{"summary": summary, "grid": grid}); err != nil {
		return err
	}
	fmt.Printf("\nwrote %s\n", out)
	return nil
}

// raCartographerUnblind ports unblind.py: it joins the judge's blind coverage
// calls (<study>/scores/<judge>, default judge_raw.json) to the blind map and
// writes <study>/scores/coverage_grid.json keyed by condition_run.
func raCartographerUnblind(args []string) error {
	judgeName, rest := flagValue(args, "--judge")
	if len(rest) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score cartographer-unblind <study-dir> [--judge judge_raw.json]")
	}
	study := rest[0]
	if judgeName == "" {
		judgeName = "judge_raw.json"
	}
	scores := filepath.Join(study, "scores")
	mapping, err := cartoReadBlindMap(filepath.Join(scores, "blind_map.json"))
	if err != nil {
		return err
	}
	var judge map[string]map[string]int
	if err := readJSON(filepath.Join(scores, judgeName), &judge); err != nil {
		return err
	}
	grid, missing := cartographer.Unblind(mapping, judge)
	out := filepath.Join(scores, "coverage_grid.json")
	payload := map[string]any{"judge": "claude-opus-4-8, blind to condition", "grid": grid}
	if err := writeJSON(out, payload); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%d rows)\n", out, len(grid))
	if len(missing) > 0 {
		head := missing
		if len(head) > 10 {
			head = head[:10]
		}
		fmt.Printf("WARNING: %d blind ids had no judge call: %v\n", len(missing), head)
	}
	return nil
}

// cartoReadTaboos reads <study>/taboo_set.json into the taboo slice, in file order.
func cartoReadTaboos(study string) ([]cartographer.Taboo, error) {
	var ts struct {
		Taboos []cartographer.Taboo `json:"taboos"`
	}
	if err := readJSON(filepath.Join(study, "taboo_set.json"), &ts); err != nil {
		return nil, err
	}
	return ts.Taboos, nil
}

// cartoReadGrid reads a coverage_grid.json, preserving the grid's key order so
// analyze_duty reports per-condition means in first-appearance order.
func cartoReadGrid(path string) ([]cartographer.GridEntry, error) {
	var top struct {
		Grid json.RawMessage `json:"grid"`
	}
	if err := readJSON(path, &top); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(top.Grid))
	if _, err := dec.Token(); err != nil { // opening '{'
		return nil, err
	}
	var out []cartographer.GridEntry
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := keyTok.(string)
		var row struct {
			Condition string         `json:"condition"`
			Run       int            `json:"run"`
			Coverage  map[string]int `json:"coverage"`
		}
		if err := dec.Decode(&row); err != nil {
			return nil, err
		}
		out = append(out, cartographer.GridEntry{
			Key: key, Condition: row.Condition, Run: row.Run, Coverage: row.Coverage,
		})
	}
	return out, nil
}

// cartoReadBlindMap reads blind_map.json, preserving id order so a missing-id
// warning lists ids in map order.
func cartoReadBlindMap(path string) ([]cartographer.BlindMapEntry, error) {
	b, err := os.ReadFile(path) //nolint:gosec // a study score path
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	if _, err := dec.Token(); err != nil { // opening '{'
		return nil, err
	}
	var out []cartographer.BlindMapEntry
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		id, _ := keyTok.(string)
		var meta struct {
			Condition string `json:"condition"`
			Run       int    `json:"run"`
		}
		if err := dec.Decode(&meta); err != nil {
			return nil, err
		}
		out = append(out, cartographer.BlindMapEntry{ID: id, Condition: meta.Condition, Run: meta.Run})
	}
	return out, nil
}

// cartoReadRunDuties reads every "<cond>_<run>.txt" under <run-dir>/out/responses
// in sorted-filename order, matching python's sorted(os.listdir). It is the input
// to the blind-set and slug-c2 tools.
func cartoReadRunDuties(runDir string) ([]cartographer.Duty, error) {
	respDir := filepath.Join(runDir, "out", "responses")
	ents, err := os.ReadDir(respDir)
	if err != nil {
		return nil, err
	}
	var out []cartographer.Duty
	for _, e := range ents { // ReadDir returns entries sorted by name
		m := cartoFName.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		run, _ := strconv.Atoi(m[2])
		b, err := os.ReadFile(filepath.Join(respDir, e.Name())) //nolint:gosec // a study response path
		if err != nil {
			return nil, err
		}
		out = append(out, cartographer.Duty{Condition: m[1], Run: run, Text: string(b)})
	}
	return out, nil
}

// cartoReadDutyResponses reads the per-condition response texts for the keyword
// pass: for each condition, files <cond>_<i>.txt for i in 1..30, skipping any that
// are absent, matching analyze_duty's loop.
func cartoReadDutyResponses(respDir string) (map[string][]string, error) {
	out := make(map[string][]string, len(cartographer.Conds))
	for _, cond := range cartographer.Conds {
		for i := 1; i <= 30; i++ {
			p := filepath.Join(respDir, fmt.Sprintf("%s_%d.txt", cond, i))
			b, err := os.ReadFile(p) //nolint:gosec // a study response path
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			out[cond] = append(out[cond], string(b))
		}
	}
	return out, nil
}

func init() {
	registerPubScoreMode("cartographer-analyze", raCartographerAnalyze)
	registerPubScoreMode("cartographer-analyze-duty", raCartographerAnalyzeDuty)
	registerPubScoreMode("cartographer-build-blind-set", raCartographerBuildBlindSet)
	registerPubScoreMode("cartographer-slug-c2", raCartographerSlugC2)
	registerPubScoreMode("cartographer-unblind", raCartographerUnblind)
}
