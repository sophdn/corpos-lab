// Command lab-assay is the container-side assay entrypoint. It reads the
// study spec and materials from the input dir, runs the assay against the
// model named in study.json, and writes results + responses to the output
// dir. It is the real runner that replaces registry-lab's placeholder shim.
//
// Directories default to /in and /out (the container's only state channels)
// and can be overridden with LAB_IN_DIR / LAB_OUT_DIR for host-side dry runs.
package main

import (
	"context"
	"fmt"
	"os"

	"corpos-lab/internal/model"
	"corpos-lab/internal/runner"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()

	inDir := envOr("LAB_IN_DIR", "/in")
	outDir := envOr("LAB_OUT_DIR", "/out")

	spec, err := runner.LoadSpec(inDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[lab-assay] %v\n", err)
		return 1
	}

	if spec.Model.BaseURL == "" || spec.Model.ModelID == "" {
		fmt.Fprintf(os.Stderr, "[lab-assay] study.json model.base_url and model.model_id are required\n")
		return 1
	}

	client := model.NewOpenAI(spec.Model.BaseURL, spec.Model.ModelID, spec.Model.Version)

	fmt.Printf("[lab-assay] assay=%s item=%s model=%s conditions=%v runs=%d\n",
		spec.Assay, spec.ItemID, spec.Model.ModelID, spec.Conditions, spec.RunsPerCell)

	results, err := runner.Execute(ctx, inDir, outDir, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[lab-assay] run failed: %v\n", err)
		return 1
	}

	fmt.Printf("[lab-assay] wrote %d rows to %s/results.json\n", len(results.Rows), outDir)
	return 0
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
