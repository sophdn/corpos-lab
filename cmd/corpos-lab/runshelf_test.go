package main

import (
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/study"
)

func TestShelfPlan(t *testing.T) {
	roles := []study.ShelfRole{{Role: "anchor", ModelID: "Mistral.gguf"}, {Role: "primary", ModelID: "Qwen.gguf"}}
	steps := shelfPlan(roles, "/x/with-model.sh", "scripts/run-grid.sh", "defs", 5)
	if len(steps) != 2 || steps[1].Role != "primary" {
		t.Fatalf("steps = %+v", steps)
	}
	a := strings.Join(steps[0].Args, " ")
	for _, want := range []string{"/x/with-model.sh Mistral -- bash -c", `--model "role:anchor"`, `--model "Mistral"`, "--chunk 5"} {
		if !strings.Contains(a, want) {
			t.Errorf("anchor step %q missing %q", a, want)
		}
	}
}

func TestRunRunShelf(t *testing.T) {
	dir := t.TempDir()
	shelf := filepath.Join(dir, "shelf.toml")
	writeFileT(t, shelf, "[roles.primary]\nmodel_id = \"Q.gguf\"\n")
	if code := run([]string{"run-shelf", "defs", "-shelf", shelf, "-dry-run"}); code != 0 {
		t.Errorf("dry run exit %d", code)
	}
	for _, args := range [][]string{
		{"run-shelf"},
		{"run-shelf", "defs", "-nope"},
		{"run-shelf", "defs", "-shelf", filepath.Join(dir, "absent.toml")},
		{"run-shelf", "defs", "-shelf", shelf, "-llama-server-repo", dir},
	} {
		if code := run(args); code == 0 {
			t.Errorf("run(%v) = 0, want non-zero", args)
		}
	}
	bad := filepath.Join(dir, "bad.toml")
	writeFileT(t, bad, "[roles.x]\n")
	if code := run([]string{"run-shelf", "defs", "-shelf", bad, "-dry-run"}); code != 1 {
		t.Errorf("bad shelf exit %d", code)
	}
}
