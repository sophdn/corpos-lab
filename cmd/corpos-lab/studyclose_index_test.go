package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStudyCloseFailsWithoutIndexRow(t *testing.T) {
	root := t.TempDir()
	study := filepath.Join(root, "studies", "alpha")
	if err := os.MkdirAll(study, 0o750); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"study-close", study}); code != 1 {
		t.Errorf("no INDEX.md: exit %d, want 1", code)
	}
	writeFileT(t, filepath.Join(root, "studies", "INDEX.md"), "| study | verdict |\n|---|---|\n| beta | x |\n")
	if code := run([]string{"study-close", study}); code != 1 {
		t.Errorf("no row: exit %d, want 1", code)
	}
	writeFileT(t, filepath.Join(root, "studies", "INDEX.md"), "| study | verdict |\n|---|---|\n| alpha | x |\n")
	if code := run([]string{"study-close", study}); code != 0 {
		t.Errorf("row present: exit %d, want 0", code)
	}
	elsewhere := filepath.Join(root, "battery-runs", "x")
	if err := os.MkdirAll(elsewhere, 0o750); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"study-close", elsewhere}); code != 0 {
		t.Errorf("outside studies/: exit %d, want 0", code)
	}
}
