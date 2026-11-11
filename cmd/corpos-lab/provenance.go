package main

import "corpos-lab/internal/runprov"

// runProvenance prints one provenance block per model for every run under a
// study dir (builds, n_ctx, sampler chain, seeds, image digests, truncation and
// model-mismatch counts).
func runProvenance(args []string) int {
	return runDirReport("provenance", "usage: corpos-lab provenance <study-dir> [-json]", args,
		runprov.Load, runprov.WriteTable, runprov.WriteJSON)
}
