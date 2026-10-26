package main

import "corpos-lab/internal/loopcells"

// runCells prints the per-cell outcomes of an agentic-loop run dir: files
// opened, claimed edits and whether each ran, sandbox changes, lost calls, and
// how each cell ended. It replaces the throwaway per-battery scorers.
func runCells(args []string) int {
	return runDirReport("cells", "usage: corpos-lab cells <run-dir> [-json]", args,
		loopcells.Load, loopcells.WriteTable, loopcells.WriteJSON)
}
