package main

import (
	"path/filepath"
	"testing"
)

func TestRunCells(t *testing.T) {
	dir := t.TempDir()
	writeFileT(t, filepath.Join(dir, "out", "results.json"), `{"rows": [{"condition": "baseline", "run": 1, "rationale": "agentic-loop-probe:baseline:terminal=final:turns=1:edits=0:unscored"}]}`)
	writeFileT(t, filepath.Join(dir, "out", "responses", "baseline_1.txt"), "[turn 1] FINAL done")
	for _, args := range [][]string{{"cells", dir}, {"cells", dir, "-json"}, {"cells", "-json", dir}} {
		if code := run(args); code != 0 {
			t.Errorf("run(%v) = %d, want 0", args, code)
		}
	}
	for _, args := range [][]string{{"cells"}, {"cells", t.TempDir()}, {"cells", dir, "-nope"}} {
		if code := run(args); code == 0 {
			t.Errorf("run(%v) = 0, want non-zero", args)
		}
	}
}
