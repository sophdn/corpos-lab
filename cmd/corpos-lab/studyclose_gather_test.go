package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestGatherStudyCloseCollectsEveryFact pins what the walk gathers from each kind
// of entry: conditions from every def (deduplicated and sorted, unparseable defs
// skipped), every parseable run-record status, rater ids from scores/ subdirs and
// loose scores/<rater>.json files, the reconciliation and predictions markers, and
// the INDEX.md row. An unreadable directory is skipped, not fatal.
func TestGatherStudyCloseCollectsEveryFact(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "studies", "gamma")
	files := map[string]string{
		"a.toml":                       `conditions = ["glyph_only", "baseline"]`,
		"sub/b.toml":                   `conditions = ["baseline", "neutral_prefix"]`,
		"broken.toml":                  `conditions = [`,
		"runs/r1/run-record.json":      `{"status": "completed"}`,
		"runs/r2/run-record.json":      `{"status": "failed"}`,
		"runs/r3/run-record.json":      `not json`,
		"scores/raterA/codes.json":     `{}`,
		"scores/raterB.v2.json":        `{}`,
		"scores/notes.txt":             "x",
		"nested/scores/raterC/x.json":  `{}`,
		"RECONCILED.md":                "done",
		"Predictions-2026.md":          "p",
		"locked/inner/run-record.json": `{"status": "hidden"}`,
		"../INDEX.md":                  "| study | verdict |\n|---|---|\n| `gamma` | x |\n",
	}
	for rel, body := range files {
		writeFileT(t, filepath.Join(dir, rel), body)
	}
	if os.Geteuid() != 0 {
		if err := os.Chmod(filepath.Join(dir, "locked"), 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "locked"), 0o700) })
	}
	in, err := gatherStudyClose(dir)
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	got := fmt.Sprintf("%+v", in)
	want := fmt.Sprintf("{Conditions:[baseline glyph_only neutral_prefix] RunStatuses:[completed failed] "+
		"SecondRaters:[raterA raterB raterC] ReconciliationMarker:true PredictionsFile:true "+
		"Study:gamma IndexPath:%s IndexRow:true}", filepath.Join(root, "studies", "INDEX.md"))
	if os.Geteuid() == 0 {
		t.Skip("root reads the locked dir; the skip-unreadable case needs a non-root run")
	}
	if got != want {
		t.Errorf("gatherStudyClose =\n%s\nwant\n%s", got, want)
	}
}

// A study outside a studies/ folder has no index to check.
func TestGatherStudyCloseOutsideStudies(t *testing.T) {
	dir := t.TempDir()
	writeFileT(t, filepath.Join(dir, "run-record.json"), `{"status": "completed"}`)
	in, err := gatherStudyClose(dir)
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if in.IndexPath != "" || in.IndexRow || len(in.RunStatuses) != 1 {
		t.Errorf("got %+v", in)
	}
}
