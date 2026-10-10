package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"corpos-lab/internal/study"
)

// shelfStep is one invocation run-shelf makes: with-model.sh swaps to a shelf
// model, runs run-grid.sh over the defs for that model, and swaps back.
type shelfStep struct {
	Role string
	Args []string
}

// shelfPlan builds the invocations for a shelf sweep. Each role runs twice
// through run-grid.sh's --model substring filter: once for defs that name the
// role ("role:<name>") and once for defs that name the gguf, so a grid written
// either way is covered. run-grid.sh skips completed cells, so the second pass
// costs nothing when the first ran them. Roles come primary-last
// (study.ShelfRoles), so the sweep ends on the default model.
func shelfPlan(roles []study.ShelfRole, withModel, runGrid, defs string, chunk int) []shelfStep {
	var steps []shelfStep
	for _, r := range roles {
		stem := strings.TrimSuffix(r.ModelID, ".gguf")
		grid := fmt.Sprintf("%q --defs %q --model %q", runGrid, defs, "role:"+r.Role)
		grid += fmt.Sprintf(" && %q --defs %q --model %q", runGrid, defs, stem)
		if chunk > 0 {
			grid = strings.ReplaceAll(grid, " --model ", " --chunk "+strconv.Itoa(chunk)+" --model ")
		}
		steps = append(steps, shelfStep{Role: r.Role, Args: []string{withModel, stem, "--", "bash", "-c", grid}})
	}
	return steps
}

// runRunShelf sweeps a grid of study defs across every model on the shelf, one
// model at a time through the llama-server repo's with-model.sh, which restores
// the previously served model even when a leg fails or is interrupted.
func runRunShelf(args []string) int {
	fs := flag.NewFlagSet("run-shelf", flag.ContinueOnError)
	shelf := fs.String("shelf", "deploy/shelf.toml", "model shelf")
	llamaRepo := fs.String("llama-server-repo", filepath.Join(os.Getenv("HOME"), "dev", "the model-server repo"), "checkout holding scripts/with-model.sh")
	chunk := fs.Int("chunk", 0, "pass --chunk N to run-grid.sh (0 = all pending cells)")
	dryRun := fs.Bool("dry-run", false, "print the plan without running it")
	var defs string
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		defs, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if defs == "" && fs.NArg() == 1 {
		defs = fs.Arg(0)
	}
	if defs == "" {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab run-shelf <defs-dir> [-shelf deploy/shelf.toml] [-llama-server-repo DIR] [-chunk N] [-dry-run]")
		return 2
	}
	raw, err := os.ReadFile(*shelf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab run-shelf: %v\n", err)
		return 1
	}
	roles, err := study.ShelfRoles(string(raw))
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab run-shelf: %v\n", err)
		return 1
	}
	withModel := filepath.Join(*llamaRepo, "scripts", "with-model.sh")
	if _, err := os.Stat(withModel); err != nil && !*dryRun {
		fmt.Fprintf(os.Stderr, "corpos-lab run-shelf: %v (set -llama-server-repo)\n", err)
		return 1
	}
	for _, s := range shelfPlan(roles, withModel, filepath.Join("scripts", "run-grid.sh"), defs, *chunk) {
		fmt.Printf("run-shelf: %s: %s\n", s.Role, strings.Join(s.Args, " "))
		if *dryRun {
			continue
		}
		cmd := exec.Command(s.Args[0], s.Args[1:]...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab run-shelf: %s leg failed: %v (the portal was restored by with-model.sh)\n", s.Role, err)
			return 1
		}
	}
	return 0
}
