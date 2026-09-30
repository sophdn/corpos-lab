package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/baseline"
	"corpos-lab/internal/model"
	"corpos-lab/internal/rater"
	"corpos-lab/internal/study"
)

const baselineUsage = "usage:\n" +
	"  corpos-lab baseline capture <candidate> -scenario S.md [-runs N] [-out DIR] [-base URL] [-model NAME] [-version V] [-temperature T] [-max-tokens N]\n" +
	"  corpos-lab baseline fold -captures DIR -scores DIR [-out BASELINE.json]"

// runBaseline drives the two halves of the unguided-baseline measurement item 12
// consumes. They bracket the rater step: capture records the class scenario run
// UNGUIDED against the one model loaded on the portal, the operator rates the
// captured replies with the `rate` subcommand, and fold turns the rater consensus
// into the battery.BaselineOutcome JSON `battery -baseline` reads.
//
// The subcommand runs ONE model per invocation on purpose. The operator swaps the
// portal to the next shelf model between invocations, so a model that fails is
// caught as it runs, not at the end of a whole-shelf sweep.
func runBaseline(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, baselineUsage)
		return 2
	}
	switch args[0] {
	case "capture":
		return runBaselineCapture(args[1:])
	case "fold":
		return runBaselineFold(args[1:])
	default:
		fmt.Fprintln(os.Stderr, baselineUsage)
		return 2
	}
}

// runBaselineCapture runs the scenario UNGUIDED against the loaded model and
// writes two artifacts keyed by a filesystem-safe stem: a rater-ready slice
// JSONL of the replies (under <out>/slices) and a capture record with the
// provenance of what answered (under <out>/captures). It appends across
// invocations — a second model writes a second pair of files, never overwriting
// the first.
func runBaselineCapture(args []string) int {
	scenarioPath := ""
	outDir := "."
	base := "http://localhost:8081/v1"
	modelName := ""
	version := ""
	versionSet := false
	runs := 8
	temperature := 0.8
	maxTokens := 512
	rest := []string{}
	for i := 0; i < len(args); i++ {
		needsValue := func() (string, bool) {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "corpos-lab: %s needs a value\n", args[i])
				return "", false
			}
			i++
			return args[i], true
		}
		switch args[i] {
		case "-scenario":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			scenarioPath = v
		case "-out":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			outDir = v
		case "-base":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			base = v
		case "-model":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			modelName = v
		case "-version":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			version = v
			versionSet = true
		case "-runs":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			n, err := parsePositiveInt(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: -runs %v\n", err)
				return 2
			}
			runs = n
		case "-temperature":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			f, err := parseFloat(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: -temperature %v\n", err)
				return 2
			}
			temperature = f
		case "-max-tokens":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			n, err := parsePositiveInt(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: -max-tokens %v\n", err)
				return 2
			}
			maxTokens = n
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 || scenarioPath == "" {
		fmt.Fprintln(os.Stderr, baselineUsage)
		return 2
	}
	item := candidateSlug(rest[0])

	scenario, err := os.ReadFile(scenarioPath) //nolint:gosec // a scenario path from argv, not a secret
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: read scenario: %v\n", err)
		return 1
	}

	// Sampling must vary run to run so replicates are not identical — greedy
	// decoding would make every cell 0/8 or 8/8. Seed one RNG per replicate.
	seeds := make([]int, runs)
	for i := range seeds {
		seeds[i] = i + 1
	}
	samp := assay.Sampling{Temperature: temperature, Seeds: seeds, MaxTokens: maxTokens}

	client := model.NewOpenAI(base, modelName, version,
		model.WithHTTPClient(&http.Client{Timeout: 5 * time.Minute}))

	rec, err := baseline.RunCapture(context.Background(), client, item, string(scenario), runs, samp, time.Now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	// Fill the version from the shelf when the operator did not pass one by hand:
	// match the model the server reported against deploy/shelf.toml. A missing
	// shelf, an unlisted model, or a lookup error never blocks the capture — an
	// unresolved version is a recorded gap, not a failure (record-what-ran).
	if !versionSet {
		if versions, verr := study.ShelfVersions("."); verr != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: shelf version lookup: %v (recording an empty version)\n", verr)
		} else if v := versions[rec.ModelID]; v != "" {
			rec.Version = v
		}
	}

	stem := captureStem(item, rec.ModelID, rec.Requested)
	slicePath := filepath.Join(outDir, "slices", stem+".jsonl")
	capturePath := filepath.Join(outDir, "captures", stem+".json")
	if err := writeSliceJSONL(slicePath, rec.SliceLines()); err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(capturePath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: make capture dir: %v\n", err)
		return 1
	}
	if err := writeJSON(capturePath, rec); err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	reported := rec.ModelID
	if reported == "" {
		reported = "(server named none)"
	}
	fmt.Fprintf(os.Stderr, "corpos-lab: baseline capture %q — model %s, %d unguided runs\n", item, reported, len(rec.Runs))
	fmt.Fprintf(os.Stderr, "  slice   -> %s\n", slicePath)
	fmt.Fprintf(os.Stderr, "  capture -> %s\n", capturePath)
	fmt.Fprintf(os.Stderr, "  next: rate this slice per rater family, then `corpos-lab baseline fold`\n")
	return 0
}

// runBaselineFold reads the captures and the rater score dirs and writes the
// battery.BaselineOutcome JSON. It reports any model it could not resolve as a
// gap on stderr — a model with no consensus is omitted from the outcome, never
// given a guessed verdict.
func runBaselineFold(args []string) int {
	capturesDir := ""
	scoresDir := ""
	outPath := ""
	for i := 0; i < len(args); i++ {
		needsValue := func() (string, bool) {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "corpos-lab: %s needs a value\n", args[i])
				return "", false
			}
			i++
			return args[i], true
		}
		switch args[i] {
		case "-captures":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			capturesDir = v
		case "-scores":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			scoresDir = v
		case "-out":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			outPath = v
		default:
			fmt.Fprintln(os.Stderr, baselineUsage)
			return 2
		}
	}
	if capturesDir == "" || scoresDir == "" {
		fmt.Fprintln(os.Stderr, baselineUsage)
		return 2
	}

	captures, err := baseline.LoadCaptures(capturesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}
	scores, raterIDs, err := baseline.LoadRaterScores(scoresDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	outcome, gaps := baseline.Fold(captures, scores)

	encoded, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: marshal outcome: %v\n", err)
		return 1
	}
	if outPath == "" {
		fmt.Println(string(encoded))
	} else {
		if err := os.WriteFile(outPath, append(encoded, '\n'), 0o644); err != nil { //nolint:gosec // a result record, not a secret
			fmt.Fprintf(os.Stderr, "corpos-lab: write %s: %v\n", outPath, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "corpos-lab: baseline outcome -> %s\n", outPath)
	}

	fmt.Fprintf(os.Stderr, "corpos-lab: folded %d model(s) from %d rater famil(ies): %s\n",
		len(outcome.Models), len(scores), strings.Join(raterIDs, ", "))
	for _, m := range outcome.Models {
		state := "missed"
		if m.Fired {
			state = "fired"
		}
		fmt.Fprintf(os.Stderr, "  %-40s %s unguided (%d resolved runs)\n", m.ModelID, state, m.Runs)
	}
	for _, g := range gaps {
		fmt.Fprintf(os.Stderr, "  GAP %-36s %s\n", g.ModelID, g.Reason)
	}
	return 0
}

// captureStem builds a filesystem-safe stem for one model's capture artifacts:
// <item>__<model>, where model is the reported model, falling back to the
// requested one. Path separators become underscores so the stem is one path
// component.
func captureStem(item, reported, requested string) string {
	who := reported
	if who == "" {
		who = requested
	}
	if who == "" {
		who = "unknown-model"
	}
	safe := func(s string) string {
		s = strings.ReplaceAll(s, string(os.PathSeparator), "_")
		return strings.ReplaceAll(s, "/", "_")
	}
	return safe(item) + "__" + safe(who)
}

// writeSliceJSONL writes one rater slice line per file line — the JSONL the
// `rate` subcommand reads.
func writeSliceJSONL(path string, lines []rater.SliceLine) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("make slice dir: %w", err)
	}
	var b strings.Builder
	for _, l := range lines {
		encoded, err := json.Marshal(l)
		if err != nil {
			return fmt.Errorf("encode slice line: %w", err)
		}
		b.Write(encoded)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil { //nolint:gosec // a slice input, not a secret
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// parsePositiveInt parses a flag value that must be a positive integer.
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("not an integer: %q", s)
	}
	if n < 1 {
		return 0, fmt.Errorf("must be >= 1, got %d", n)
	}
	return n, nil
}

// parseFloat parses a flag value that must be a number.
func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("not a number: %q", s)
	}
	return f, nil
}
