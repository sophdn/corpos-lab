// Command lab-assay is the container-side assay entrypoint and the container's
// PID 1. It preserves registry-lab's /in-/out contract — report/run modes and
// an /out/manifest.json extraction manifest on every exit — but implements it
// as a single CGo-free static Go binary instead of bash+jq+python, matching
// the corpos family's static-binary-on-distroless model.
//
//	report (default CMD) — print environment + exit 0. Smoke test.
//	run                  — read /in, run the assay, write /out.
//
// Directories default to /in and /out (the container's only state channels)
// and can be overridden with LAB_IN_DIR / LAB_OUT_DIR for host-side dry runs.
// The image identity is carried in LAB_IMAGE (set per variant image).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"corpos-lab/internal/extract"
	"corpos-lab/internal/model"
	"corpos-lab/internal/runner"
)

func main() {
	os.Exit(dispatch())
}

// dispatch runs the selected mode and, via a named return, always writes the
// extraction manifest with the final exit code before returning — the Go
// analogue of registry-lab's EXIT trap.
func dispatch() (code int) {
	mode := "report"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	outDir := envOr("LAB_OUT_DIR", "/out")
	started := time.Now().UTC().Format(time.RFC3339)

	defer func() {
		m := extract.Manifest{
			Image:       os.Getenv("LAB_IMAGE"),
			Mode:        mode,
			StartedAt:   started,
			FinishedAt:  time.Now().UTC().Format(time.RFC3339),
			ExitCode:    code,
			ResultsPath: outDir + "/results.json",
		}
		if err := extract.Write(outDir, m); err != nil {
			fmt.Fprintf(os.Stderr, "[lab-assay] WARN: extraction manifest not written: %v\n", err)
		}
	}()

	switch mode {
	case "report":
		return reportMode(outDir)
	case "run":
		return runMode(outDir)
	default:
		fmt.Fprintf(os.Stderr, "[lab-assay] unknown mode %q (want report|run)\n", mode)
		return 2
	}
}

func reportMode(outDir string) int {
	fmt.Printf("[lab-assay] mode=report image=%s\n", os.Getenv("LAB_IMAGE"))
	fmt.Printf("[lab-assay] in=%s out=%s uid=%d\n", envOr("LAB_IN_DIR", "/in"), outDir, os.Getuid())
	return 0
}

func runMode(outDir string) int {
	inDir := envOr("LAB_IN_DIR", "/in")

	spec, err := runner.LoadSpec(inDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[lab-assay] %v\n", err)
		return 1
	}
	if spec.Model.BaseURL == "" || spec.Model.ModelID == "" {
		fmt.Fprintln(os.Stderr, "[lab-assay] study.json model.base_url and model.model_id are required")
		return 1
	}

	client := model.NewOpenAI(spec.Model.BaseURL, spec.Model.ModelID, spec.Model.Version)
	fmt.Printf("[lab-assay] assay=%s item=%s model=%s conditions=%v runs=%d\n",
		spec.Assay, spec.ItemID, spec.Model.ModelID, spec.Conditions, spec.RunsPerCell)

	results, err := runner.Execute(context.Background(), inDir, outDir, client)
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
