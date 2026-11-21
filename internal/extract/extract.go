// Package extract writes the container extraction manifest — the
// deterministic result-location contract the host-side controller reads to
// find a run's output regardless of which image produced it. It is the Go
// re-implementation of registry-lab's lab-write-manifest (bash+jq), preserving
// the /out/manifest.json shape while dropping the shell/jq dependency.
package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest is the extraction manifest written to /out/manifest.json on every
// container exit (success or crash), matching registry-lab's schema so the
// controller's result-location contract is unchanged.
type Manifest struct {
	Image       string `json:"image"`
	Mode        string `json:"mode"`
	StartedAt   string `json:"started_at"`
	FinishedAt  string `json:"finished_at"`
	ExitCode    int    `json:"exit_code"`
	ResultsPath string `json:"results_path"`
}

// Write marshals m to outDir/manifest.json. It creates outDir if needed so an
// early crash (before the assay made the dir) still records a manifest.
func Write(outDir string, m Manifest) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("extract: create out dir: %w", err)
	}
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("extract: marshal manifest: %w", err)
	}
	path := filepath.Join(outDir, "manifest.json")
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("extract: write manifest: %w", err)
	}
	return nil
}
