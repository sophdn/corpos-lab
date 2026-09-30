package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"corpos-lab/internal/calibrationstudies"
)

// The calibration-study pub-score modes replace the four study shell scripts the
// whole-tree language policy forbids: gen-defs.sh and scoring/consensus_table.sh
// under studies/calibrated-mechanism-retyping and studies/scramble-strength-calibration.
//
//	pub-score cmr-gen-defs  [--out DIR]          regenerate the CMR cell TOMLs
//	pub-score scb-gen-defs  [--out DIR]          regenerate the SCB cell TOMLs
//	pub-score cmr-consensus [--scoring-dir DIR]  print the CMR consensus table
//	pub-score scb-consensus [--scoring-dir DIR]  print the SCB consensus table
func init() {
	registerPubScoreMode("cmr-gen-defs", cmrGenDefs)
	registerPubScoreMode("scb-gen-defs", scbGenDefs)
	registerPubScoreMode("cmr-consensus", cmrConsensus)
	registerPubScoreMode("scb-consensus", scbConsensus)
}

func cmrGenDefs(args []string) error {
	out, _ := flagValue(args, "--out")
	if out == "" {
		out = filepath.FromSlash("studies/calibrated-mechanism-retyping/defs")
	}
	return writeCells(out, calibrationstudies.CMRCells())
}

func scbGenDefs(args []string) error {
	out, _ := flagValue(args, "--out")
	if out == "" {
		out = filepath.FromSlash("studies/scramble-strength-calibration/defs")
	}
	return writeCells(out, calibrationstudies.SCBCells())
}

// writeCells writes each cell TOML into out, first clearing any existing top-level
// *.toml so a regeneration leaves no stale cell. It clears only the cell files,
// not the whole directory, so a populated defs/runs/ subtree survives — unlike the
// shell script's rm -rf, which destroyed it.
func writeCells(out string, cells []calibrationstudies.Cell) error {
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	stale, err := filepath.Glob(filepath.Join(out, "*.toml"))
	if err != nil {
		return err
	}
	for _, p := range stale {
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	for _, c := range cells {
		if err := os.WriteFile(filepath.Join(out, c.Name), []byte(c.TOML), 0o600); err != nil {
			return err
		}
	}
	fmt.Printf("wrote %d cell defs to %s/\n", len(cells), out)
	return nil
}

func cmrConsensus(args []string) error {
	dir, _ := flagValue(args, "--scoring-dir")
	if dir == "" {
		dir = filepath.FromSlash("studies/calibrated-mechanism-retyping/scoring")
	}
	return printConsensus(dir)
}

func scbConsensus(args []string) error {
	dir, _ := flagValue(args, "--scoring-dir")
	if dir == "" {
		dir = filepath.FromSlash("studies/scramble-strength-calibration/scoring")
	}
	return printConsensus(dir)
}

// printConsensus reads the study key and the three rater score directories, then
// prints the consensus table. This is the IO seam; the join and formatting are
// pure in calibrationstudies.ConsensusTable.
func printConsensus(dir string) error {
	var key map[string]calibrationstudies.KeyEntry
	if err := readJSON(filepath.Join(dir, "key.json"), &key); err != nil {
		return err
	}
	claude, err := mergeRaterDir(filepath.Join(dir, "scores", "claude"))
	if err != nil {
		return err
	}
	deepseek, err := mergeRaterDir(filepath.Join(dir, "scores", "deepseek"))
	if err != nil {
		return err
	}
	devstral, err := mergeRaterDir(filepath.Join(dir, "scores", "devstral"))
	if err != nil {
		return err
	}
	fmt.Print(calibrationstudies.ConsensusTable(key, claude, deepseek, devstral))
	return nil
}

// mergeRaterDir merges every *.json in a rater directory into one id->code map,
// reproducing `jq -s 'add'`: on a duplicate id, the later file (sorted by name)
// wins.
func mergeRaterDir(dir string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	merged := map[string]string{}
	for _, f := range files {
		var m map[string]string
		if err := readJSON(f, &m); err != nil {
			return nil, err
		}
		for id, code := range m {
			merged[id] = code
		}
	}
	return merged, nil
}
