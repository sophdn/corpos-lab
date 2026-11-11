package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"corpos-lab/internal/restaxis"
)

func init() {
	registerRestAxisMode("gen-grid", raGenGrid)
}

// pyImage is gen_grid.py's hardcoded IMAGE digest, used as the default so a bare
// invocation reproduces the committed study TOMLs. Override with --image.
const pyImage = "localhost/lab-grounded-glyph-probe@sha256:305821cadd983337be7efac69d25d2b2c8df871430b2b2c7f903dba6d5541f26"

// pyGlyphs and pyScenarios are gen_grid.py's hardcoded GLYPHS and SCENARIOS,
// used as defaults for the same reason.
var pyGlyphs = []string{
	"casg-direct",
	"conditional-gate-uniform-default",
	"formal-step-context-bypass",
	"parent-state-check-bypass",
}

// raGenGrid generates the rest-axis ablation grid study TOMLs, one per
// (glyph, scenario), for a single model. It ports gen_grid.py, whose hardcoded
// study/output paths and model source TOMLs become flags here:
//
//	--out-dir <dir>     (required) base dir; a TOML is written to <dir>/<glyph>/study.<tag>.s<n>.toml
//	--model <tag>       (required) model tag, e.g. qwen38
//	--model-src <path>  (required) source TOML holding the [model] section to embed verbatim
//	--image <ref>       image digest reference (default: gen_grid.py's IMAGE)
//	--glyphs <csv>      comma-separated glyph list (default: gen_grid.py's GLYPHS)
//	--scenarios <spec>  scenarios as "a-b" or a comma list (default: 1-20)
//	--conditions <csv>  arm/condition list (default: baseline,glyph_only,glyph_minus_rest)
//	--runs-per-cell <n> runs per cell (default: 3)
//	--seeds <csv>       sampler seeds (default: 1,2,3)
func raGenGrid(args []string) error {
	outDir, args := flagValue(args, "--out-dir")
	modelTag, args := flagValue(args, "--model")
	modelSrc, args := flagValue(args, "--model-src")
	image, args := flagValue(args, "--image")
	glyphsCSV, args := flagValue(args, "--glyphs")
	scenariosSpec, args := flagValue(args, "--scenarios")
	conditionsCSV, args := flagValue(args, "--conditions")
	runsPerCellStr, args := flagValue(args, "--runs-per-cell")
	seedsCSV, args := flagValue(args, "--seeds")
	if len(args) != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}
	if outDir == "" {
		return fmt.Errorf("need --out-dir <dir>")
	}
	if modelTag == "" {
		return fmt.Errorf("need --model <tag>")
	}
	if modelSrc == "" {
		return fmt.Errorf("need --model-src <path> (source TOML with the [model] section)")
	}

	if image == "" {
		image = pyImage
	}
	glyphs := pyGlyphs
	if glyphsCSV != "" {
		glyphs = splitCSV(glyphsCSV)
	}
	scenarios, err := parseScenarios(scenariosSpec)
	if err != nil {
		return err
	}
	conditions := []string{"baseline", "glyph_only", "glyph_minus_rest"}
	if conditionsCSV != "" {
		conditions = splitCSV(conditionsCSV)
	}
	runsPerCell := 3
	if runsPerCellStr != "" {
		runsPerCell, err = strconv.Atoi(runsPerCellStr)
		if err != nil {
			return fmt.Errorf("bad --runs-per-cell %q: %w", runsPerCellStr, err)
		}
	}
	seeds := []int{1, 2, 3}
	if seedsCSV != "" {
		seeds, err = parseInts(seedsCSV)
		if err != nil {
			return fmt.Errorf("bad --seeds %q: %w", seedsCSV, err)
		}
	}

	srcBytes, err := os.ReadFile(modelSrc) //nolint:gosec // a study source path supplied by the operator
	if err != nil {
		return fmt.Errorf("read --model-src: %w", err)
	}
	modelBlock := restaxis.ExtractModelBlock(string(srcBytes))
	if strings.TrimSpace(modelBlock) == "" {
		return fmt.Errorf("no [model] section found in %s", modelSrc)
	}

	cells := restaxis.GenGrid(restaxis.GridSpec{
		Glyphs:      glyphs,
		Scenarios:   scenarios,
		ModelTag:    modelTag,
		ModelBlock:  modelBlock,
		Image:       image,
		Conditions:  conditions,
		RunsPerCell: runsPerCell,
		Seeds:       seeds,
	})

	for _, c := range cells {
		dir := filepath.Join(outDir, c.Glyph)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		dest := filepath.Join(dir, restaxis.StudyFileName(modelTag, c.Scenario))
		if err := os.WriteFile(dest, []byte(c.TOML), 0o600); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "wrote %d study tomls for model %s under %s\n", len(cells), modelTag, outDir)
	return nil
}

// parseScenarios parses a scenario spec: "" defaults to 1..20 (gen_grid.py's
// SCENARIOS), "a-b" is the inclusive range a..b, and otherwise a comma list.
func parseScenarios(spec string) ([]int, error) {
	if spec == "" {
		out := make([]int, 0, 20)
		for i := 1; i <= 20; i++ {
			out = append(out, i)
		}
		return out, nil
	}
	if lo, hi, ok := strings.Cut(spec, "-"); ok {
		a, err := strconv.Atoi(strings.TrimSpace(lo))
		if err != nil {
			return nil, fmt.Errorf("bad --scenarios range start %q: %w", lo, err)
		}
		b, err := strconv.Atoi(strings.TrimSpace(hi))
		if err != nil {
			return nil, fmt.Errorf("bad --scenarios range end %q: %w", hi, err)
		}
		if b < a {
			return nil, fmt.Errorf("bad --scenarios range %q: end before start", spec)
		}
		out := make([]int, 0, b-a+1)
		for i := a; i <= b; i++ {
			out = append(out, i)
		}
		return out, nil
	}
	return parseInts(spec)
}

// parseInts parses a comma-separated list of ints.
func parseInts(csv string) ([]int, error) {
	parts := splitCSV(csv)
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("bad int %q: %w", p, err)
		}
		out = append(out, n)
	}
	return out, nil
}

// splitCSV splits a comma-separated list, trimming whitespace and dropping empty
// fields.
func splitCSV(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
