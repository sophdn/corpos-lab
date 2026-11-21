package main

import (
	"path/filepath"
	"testing"
)

func TestRunNewLoopStudy(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "S1")
	if code := run([]string{"new-loop-study", dir, "-query"}); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if code := run([]string{"new-loop-study", dir}); code != 1 {
		t.Errorf("rewrite exit = %d, want 1", code)
	}
	if code := run([]string{"new-loop-study", "-name", "x", filepath.Join(t.TempDir(), "S2")}); code != 0 {
		t.Errorf("flags-first exit = %d", code)
	}
	for _, args := range [][]string{{"new-loop-study"}, {"new-loop-study", dir, "-nope"}} {
		if code := run(args); code != 2 {
			t.Errorf("run(%v) = %d, want 2", args, code)
		}
	}
}
