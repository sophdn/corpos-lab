package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("merge-completions", raMergeCompletions) }

// mergeGlyphs is the fixed glyph set merged by default, matching GLYPHS in
// merge_completions.py.
var mergeGlyphs = []string{
	"casg-direct", "conditional-gate-uniform-default",
	"formal-step-context-bypass", "parent-state-check-bypass",
}

// raMergeCompletions merges each glyph's CORRECT_COMPLETIONS_NEW.md into
// CORRECT_COMPLETIONS.md, bumps the scenario count to 20, and deletes the _NEW
// file. Ported from merge_completions.py.
//
//	corpos-lab rest-axis merge-completions [--glyphs a,b,c] <study-dir>
//
// A glyph with no _NEW file is skipped with a note (the merge is idempotent — a
// study whose files are already merged has no _NEW files left).
func raMergeCompletions(args []string) error {
	glyphsFlag, args := flagValue(args, "--glyphs")
	if len(args) != 1 {
		return fmt.Errorf("need a study-dir")
	}
	study := args[0]
	glyphs := mergeGlyphs
	if glyphsFlag != "" {
		glyphs = strings.Split(glyphsFlag, ",")
	}

	for _, g := range glyphs {
		mainPath := filepath.Join(study, g, "CORRECT_COMPLETIONS.md")
		newPath := filepath.Join(study, g, "CORRECT_COMPLETIONS_NEW.md")
		nt, err := os.ReadFile(newPath) //nolint:gosec // a study ground-truth path
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "%s: no _NEW file, skipped\n", g)
				continue
			}
			return err
		}
		mt, err := os.ReadFile(mainPath) //nolint:gosec // a study ground-truth path
		if err != nil {
			return err
		}
		merged := restaxis.MergeCompletions(string(mt), string(nt))
		if err := os.WriteFile(mainPath, []byte(merged), 0o600); err != nil {
			return err
		}
		if err := os.Remove(newPath); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s: merged -> %d lines; _NEW deleted; scenario-count bumped=%t\n",
			g, len(strings.Split(strings.TrimRight(merged, "\n"), "\n")), strings.Contains(merged, "scenarios: 20"))
	}
	return nil
}
