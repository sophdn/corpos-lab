// Command corpos-lab is the headless lab controller CLI. It runs a study
// end-to-end from a single definition file — reproducible with no agent in the
// loop — and is the surface corpos drives as an external tool.
//
// Usage:
//
//	corpos-lab run-study <def.toml> [-work DIR]
//	corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]
//
// Exit 0 on a completed run, non-zero on any failure (the record is written
// regardless).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/batteryrun"
	"corpos-lab/internal/control"
	"corpos-lab/internal/image"
	"corpos-lab/internal/model"
	"corpos-lab/internal/persist"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/study"
	"corpos-lab/internal/substrate"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

const usage = "usage:\n" +
	"  corpos-lab run-study <def.toml> [-work DIR]\n" +
	"  corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]"

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "run-study":
		return runStudy(args[1:])
	case "battery":
		return runBattery(args[1:])
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
}

// runBattery runs the mechanized ALPHABET battery against one candidate and
// writes the JSON result (the mechanized item verdicts, the Deferred stubs for
// the judged items, and run-level provenance). The judged items are resolved
// separately by an assessor; this surface produces only the measured half.
func runBattery(args []string) int {
	outPath := ""
	base := "http://localhost:8081/v1"
	modelName := "qwen3.6-27b"
	version := ""
	repoDir := "."
	allItems := false
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
		case "-out":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			outPath = v
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
		case "-repo":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			repoDir = v
		case "-all-items":
			allItems = true
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]")
		return 2
	}
	candidatePath := rest[0]
	itemID := candidateSlug(candidatePath)

	// A thinking model producing reasoning for two verdict calls is slow; give
	// each generation room rather than clip it into a transport error.
	client := model.NewOpenAI(base, modelName, version,
		model.WithHTTPClient(&http.Client{Timeout: 5 * time.Minute}))

	res, err := batteryrun.Run(context.Background(), candidatePath, itemID, client, batteryrun.Deps{
		Substrate:  substrateProbe,
		Provenance: repoStamp(repoDir),
		Props:      client.Props,
	}, batteryrun.Options{AllItems: allItems})
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	encoded, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: marshal result: %v\n", err)
		return 1
	}
	if outPath == "" {
		fmt.Println(string(encoded))
	} else {
		if err := os.WriteFile(outPath, append(encoded, '\n'), 0o644); err != nil { //nolint:gosec // a result record, not a secret
			fmt.Fprintf(os.Stderr, "corpos-lab: write %s: %v\n", outPath, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "corpos-lab: battery result -> %s\n", outPath)
	}
	printBatterySummary(os.Stderr, res)
	return 0
}

// candidateSlug derives an item id from a candidate filename: it strips a
// leading "CANDIDATE_" and a trailing "_<date>.md", falling back to the base
// name without extension.
func candidateSlug(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name = strings.TrimPrefix(name, "CANDIDATE_")
	if i := strings.LastIndex(name, "_20"); i > 0 {
		name = name[:i]
	}
	return name
}

// printBatterySummary writes a one-line-per-item human summary of the
// mechanized outcome so a reader does not have to parse the JSON to see where a
// run landed.
func printBatterySummary(w *os.File, res batteryrun.Result) {
	fmt.Fprintf(w, "corpos-lab: battery %q — model %q, substrate %s, commit %s\n",
		res.ItemID, res.Provenance.ModelID, res.Provenance.Substrate.Kind, shortSHA(res.Provenance.Provenance.CommitSHA))
	for _, sr := range res.Sequence.StepResults {
		fmt.Fprintf(w, "  %-28s %s\n", sr.StepName, verdictLabel(sr.Outcome))
	}
	switch {
	case res.Sequence.Passed:
		fmt.Fprintln(w, "  => mechanized items PASS (judged items still Deferred to an assessor)")
	case res.AllItems:
		fmt.Fprintf(w, "  => FAIL, all items run: %s\n", res.Sequence.FailureReason)
	default:
		fmt.Fprintf(w, "  => stopped at step %d: %s\n", res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}
}

func verdictLabel(o battery.StepOutcome) string {
	switch o.Kind {
	case battery.OutcomeError:
		return "ERROR: " + o.Message
	case battery.OutcomeObservation:
		return "observation"
	default:
		if o.Verdict == nil {
			return "?"
		}
		return string(o.Verdict.Kind)
	}
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	if sha == "" {
		return "(unstamped)"
	}
	return sha
}

// repoStamp stamps the instrument repo (corpos-lab) that ran the battery.
// Defaults to the working directory, so run it from the checkout; -repo
// overrides. Records its own failure rather than inventing a commit.
func repoStamp(repoDir string) func(context.Context) (provenance.Stamp, error) {
	return func(ctx context.Context) (provenance.Stamp, error) {
		dir, err := filepath.Abs(repoDir)
		if err != nil {
			return provenance.Stamp{}, err
		}
		return provenance.Capture(ctx, dir)
	}
}

func runStudy(args []string) int {
	var defPath, workDir string
	toolkitURL := persist.DefaultToolkitURL
	project := "glyph-research"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-work":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -work needs a value")
				return 2
			}
			i++
			workDir = args[i]
		case "-toolkit-url":
			// Empty value disables persistence to the toolkit.
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -toolkit-url needs a value")
				return 2
			}
			i++
			toolkitURL = args[i]
		case "-project":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -project needs a value")
				return 2
			}
			i++
			project = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab run-study <def.toml> [-work DIR]")
		return 2
	}
	defPath = rest[0]

	def, err := study.LoadDef(defPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	if workDir == "" {
		workDir = filepath.Join(filepath.Dir(defPath), "runs", def.Name)
	}

	fmt.Printf("corpos-lab: running study %q (%s, %d conditions × %d) on %s\n",
		def.Name, def.Assay, len(def.Conditions), def.RunsPerCell, def.Network)

	runResult, err := control.RunStudy(context.Background(), def, workDir, control.Deps{
		Launcher:   podmanLauncher{},
		DigestOf:   digestReader,
		Substrate:  substrateProbe,
		Provenance: provenanceReader(defPath),
	})
	recordPath := filepath.Join(workDir, "run-record.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: study FAILED: %v\n", err)
		fmt.Fprintf(os.Stderr, "corpos-lab: failed-run record -> %s\n", recordPath)
		return 1
	}

	fmt.Printf("corpos-lab: study %s — %d rows, image %s\n",
		runResult.Status, len(runResult.Results.Rows), runResult.ImageDigest[:19])
	fmt.Printf("corpos-lab: run record -> %s\n", recordPath)

	// Persist to the toolkit for one queryable home (best-effort: the local
	// run record is the durable artifact, so a persist failure is a warning,
	// not a run failure). Raw responses stay on disk; only the pointer is sent.
	if toolkitURL != "" {
		responsesDir := filepath.Join(workDir, "out", "responses")
		client := persist.NewClient(toolkitURL, project)
		if err := client.Record(context.Background(), runResult, responsesDir); err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: WARN: persist to toolkit failed (run record on disk is unaffected): %v\n", err)
		} else {
			fmt.Printf("corpos-lab: persisted to toolkit (%s, project %s)\n", toolkitURL, project)
		}
	}
	return 0
}

// digestReader reads an image content digest via rootless podman.
func digestReader(ctx context.Context, ref string) (string, error) {
	return image.Digest(ctx, func(ctx context.Context, r string) ([]byte, error) {
		return exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", r).Output()
	}, ref)
}

// substrateProbe identifies the processor the inference server is running on.
func substrateProbe(ctx context.Context) substrate.Info {
	return substrate.Probe(ctx, substrate.ExecRunner)
}

// provenanceReader stamps the repo the study definition lives in — the tree
// that carries both the instrument and the materials.
//
// A caveat worth stating plainly rather than leaving for someone to discover:
// this stamps the repo at defPath, which is the instrument's repo when the
// documented invocation is used (run from the checkout) and is NOT when the
// binary is run from elsewhere against a def somewhere else. The robust form is
// a build-time stamp (-ldflags -X main.commit=$(git rev-parse HEAD)), which
// describes the binary regardless of where it runs. This is the honest
// available approximation, and it records its own failure rather than
// inventing a commit.
func provenanceReader(defPath string) control.ProvenanceReader {
	return func(ctx context.Context) (provenance.Stamp, error) {
		dir, err := filepath.Abs(filepath.Dir(defPath))
		if err != nil {
			return provenance.Stamp{}, err
		}
		return provenance.Capture(ctx, dir)
	}
}

// podmanLauncher runs one disposable assay container with rootless podman.
type podmanLauncher struct{}

func (podmanLauncher) Launch(ctx context.Context, spec control.LaunchSpec) (control.LaunchResult, error) {
	if err := os.MkdirAll(spec.OutDir, 0o755); err != nil {
		return control.LaunchResult{}, fmt.Errorf("corpos-lab: create out dir: %w", err)
	}
	// Argv assembly lives in control.PodmanArgs, not here: cmd/ has no tests and
	// the gate's coverage floor scopes to ./internal/..., so an inline argv was
	// the one translation in the tree with nowhere to assert on it. This
	// launcher keeps only the exec concern.
	argv, err := control.PodmanArgs(spec)
	if err != nil {
		return control.LaunchResult{}, err
	}
	cmd := exec.CommandContext(ctx, "podman", argv...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		return control.LaunchResult{ExitCode: 0}, nil
	}
	// A non-zero container exit is a classified run failure, not a launcher
	// error — return the exit code so the controller records it.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return control.LaunchResult{ExitCode: exitErr.ExitCode(), Stderr: strings.TrimSpace(stderr.String())}, nil
	}
	// podman itself could not run (missing binary, bad network name, etc.).
	return control.LaunchResult{}, fmt.Errorf("corpos-lab: podman run: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
}
