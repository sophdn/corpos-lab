package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"corpos-lab/internal/wrongpath"
)

func init() {
	registerPubScoreMode("wrongpath-gen-study-defs", raWrongpathGenStudyDefs)
	registerPubScoreMode("wrongpath-make-ablated", raWrongpathMakeAblated)
	registerPubScoreMode("wrongpath-score", raWrongpathScore)
}

// raWrongpathGenStudyDefs ports gen_study_defs.py: it reads
// <materials>/base/INSTRUCTION.md and writes the 72-file study-def fan-out under
// <out>/<arm>/<model>/<glyph>-<ab>.toml.
//
// Usage: corpos-lab pub-score wrongpath-gen-study-defs --materials <dir> --out <dir>
func raWrongpathGenStudyDefs(args []string) error {
	materials, args := flagValue(args, "--materials")
	out, args := flagValue(args, "--out")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if materials == "" || out == "" {
		return fmt.Errorf("need --materials and --out")
	}
	b, err := os.ReadFile(filepath.Join(materials, "base", "INSTRUCTION.md")) //nolint:gosec // a study materials path
	if err != nil {
		return err
	}
	instruction := strings.TrimRight(string(b), "\n")
	defs, err := wrongpath.GenStudyDefs(instruction)
	if err != nil {
		return err
	}
	for _, d := range defs {
		p := filepath.Join(out, d.RelPath)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(d.Content), 0o600); err != nil {
			return err
		}
	}
	fmt.Printf("wrote %d study definitions under %s/\n", len(defs), out)
	return nil
}

// raWrongpathMakeAblated ports materials/make_ablated.py: it reads each
// <base>/GLYPH_<g>-terrain.md, removes the Marker and Aim axes, and writes the
// result to <out>/GLYPH_<g>-terrain.md.
//
// Usage: corpos-lab pub-score wrongpath-make-ablated --base <dir> --out <dir>
func raWrongpathMakeAblated(args []string) error {
	base, args := flagValue(args, "--base")
	out, args := flagValue(args, "--out")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if base == "" || out == "" {
		return fmt.Errorf("need --base and --out")
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	for _, g := range wrongpath.Glyphs {
		src := filepath.Join(base, fmt.Sprintf("GLYPH_%s-terrain.md", g))
		srcB, err := os.ReadFile(src) //nolint:gosec // a study materials path
		if err != nil {
			return err
		}
		ablated, err := wrongpath.Ablate(string(srcB), g)
		if err != nil {
			return err
		}
		dst := filepath.Join(out, fmt.Sprintf("GLYPH_%s-terrain.md", g))
		if err := os.WriteFile(dst, []byte(ablated), 0o600); err != nil {
			return err
		}
		fmt.Printf("%s: %d -> %d lines\n", g, wrongpath.LineCount(string(srcB)), wrongpath.LineCount(ablated))
	}
	return nil
}

// raWrongpathScore ports score.py: it reads the runs tree and prints the
// deterministic scoring report to stdout.
//
// Usage: corpos-lab pub-score wrongpath-score --runs <dir>
func raWrongpathScore(args []string) error {
	runs, args := flagValue(args, "--runs")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if runs == "" {
		return fmt.Errorf("need --runs")
	}
	cells := map[wrongpath.CellKey]wrongpath.CellScore{}
	var allRuns []wrongpath.RunFile
	for _, arm := range wrongpath.Arms {
		for _, model := range wrongpath.Models {
			for _, glyph := range wrongpath.Glyphs {
				for _, ab := range wrongpath.AB {
					key := wrongpath.CellKey{Arm: arm, Model: model, Glyph: glyph, AB: ab}
					cellRuns, score, scored := wpReadCell(runs, key)
					allRuns = append(allRuns, cellRuns...)
					if scored {
						cells[key] = score
					}
				}
			}
		}
	}
	fmt.Print(wrongpath.Report(cells, allRuns))
	return nil
}

// wpReadCell reads one cell's response files (with their reasoning traces for the
// qwen38 base arm) and scores the cell. scored is true only when the cell has a
// results.json; a cell with no responses dir yields nothing.
func wpReadCell(runs string, k wrongpath.CellKey) (cellRuns []wrongpath.RunFile, score wrongpath.CellScore, scored bool) {
	cellDir := filepath.Join(runs, k.Arm, k.Model, fmt.Sprintf("%s-%s", k.Glyph, k.AB), "out")
	respFiles, err := sortedResponses(filepath.Join(cellDir, "responses"))
	if err != nil {
		return nil, score, false // no responses for this cell
	}
	texts := make([]string, 0, len(respFiles))
	for _, rf := range respFiles {
		b, rerr := os.ReadFile(rf) //nolint:gosec // a study response path
		if rerr != nil {
			continue
		}
		text := string(b)
		texts = append(texts, text)
		run := wrongpath.RunFile{Arm: k.Arm, Model: k.Model, Glyph: k.Glyph, AB: k.AB, Text: text}
		if k.Arm == "base" && k.Model == "qwen38" {
			run.Reasoning, run.HasReasoning = wpReasoning(rf)
		}
		cellRuns = append(cellRuns, run)
	}
	// A cell counts toward the scored set only when it has a results.json.
	if _, err := os.Stat(filepath.Join(cellDir, "results.json")); err != nil {
		return cellRuns, score, false
	}
	return cellRuns, wrongpath.ScoreCell(texts, wrongpath.GT[k.AB]), true
}

// wpReasoning reads the reasoning trace paired with a response file (the same
// name under reasoning/ instead of responses/), reporting whether one exists.
func wpReasoning(responsePath string) (string, bool) {
	sep := string(os.PathSeparator)
	reasPath := strings.Replace(responsePath, sep+"responses"+sep, sep+"reasoning"+sep, 1)
	rb, err := os.ReadFile(reasPath) //nolint:gosec // a study reasoning path
	if err != nil {
		return "", false
	}
	return string(rb), true
}

// sortedResponses returns the grounded_glyph_*.txt files in dir, lexicographically
// sorted (matching python sorted(glob(...))).
func sortedResponses(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "grounded_glyph_") && strings.HasSuffix(name, ".txt") {
			out = append(out, filepath.Join(dir, name))
		}
	}
	sort.Strings(out)
	return out, nil
}
