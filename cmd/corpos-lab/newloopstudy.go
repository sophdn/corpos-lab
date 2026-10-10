package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"corpos-lab/internal/loopstudy"
)

// runNewLoopStudy stamps a new agentic-loop study dir from current defaults.
func runNewLoopStudy(args []string) int {
	fs := flag.NewFlagSet("new-loop-study", flag.ContinueOnError)
	name := fs.String("name", "", "study name (default: the dir name)")
	item := fs.String("item", "", "item id (default: the dir name)")
	query := fs.Bool("query", false, "offer run_query and stamp its data file")
	var dir string
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		dir, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if dir == "" && fs.NArg() == 1 {
		dir = fs.Arg(0)
	}
	if dir == "" {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab new-loop-study <dir> [-name N] [-item ID] [-query]")
		return 2
	}
	base := filepath.Base(filepath.Clean(dir))
	o := loopstudy.Options{Name: *name, ItemID: *item, Query: *query}
	if o.Name == "" {
		o.Name = base
	}
	if o.ItemID == "" {
		o.ItemID = base
	}
	wrote, err := loopstudy.Write(dir, o)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, p := range wrote {
		fmt.Println("wrote", p)
	}
	fmt.Printf("next: write %s and %s, then smoke one cell: corpos-lab run-study %s -runs 1\n",
		filepath.Join(dir, loopstudy.ScenarioFile), filepath.Join(dir, loopstudy.SandboxFile), filepath.Join(dir, loopstudy.StudyFile))
	return 0
}
