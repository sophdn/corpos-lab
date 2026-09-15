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
	"sort"
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
	"  corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]\n" +
	"  corpos-lab glyph-digest <slug> | --all | --file <path> [-candidates DIR]\n" +
	"  corpos-lab glyph-lint <candidate.md> [-registry ALPHABET.md]"

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
	case "glyph-digest":
		return runGlyphDigest(args[1:])
	case "glyph-lint":
		return runGlyphLint(args[1:])
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
}

// runBattery runs the mechanized ALPHABET battery against one candidate and
// writes the JSON result (the mechanized item verdicts, the Deferred stubs for
// the judged items, and run-level provenance). The judged items are resolved
// separately by an assessor; this surface produces only the measured half.
//
// It takes the candidate working-doc directly: batteryrun.Run extracts the
// assessable clean entry (the fallout line plus the AC-4 assembled-glyph block)
// from it, so there is no separate extract copy to keep in sync. Point
// <candidate.md> at corpus/private/glyph-model/candidates/CANDIDATE_<slug>_*.md.
func runBattery(args []string) int {
	outPath := ""
	base := "http://localhost:8081/v1"
	modelName := "qwen3.6-27b"
	version := ""
	repoDir := "."
	registryPath := "corpus/private/glyph-model/ALPHABET.md"
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
		case "-registry":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			registryPath = v
		case "-all-items":
			allItems = true
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-registry ALPHABET.md] [-all-items]")
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
		Registry:   batteryrun.FileRegistryReader{Path: registryPath},
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

// runGlyphDigest prints the certification digest of one or every candidate's
// AC-4 glyph block — the Go form of the retired glyph_digest.py. The block
// extraction and hashing live in battery.GlyphDigest (tested); this stays file
// glue.
func runGlyphDigest(args []string) int {
	dir := "corpus/private/glyph-model/candidates"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "-candidates" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -candidates needs a value")
				return 2
			}
			i++
			dir = args[i]
			continue
		}
		rest = append(rest, args[i])
	}

	switch {
	case len(rest) == 1 && rest[0] == "--all":
		matches, err := filepath.Glob(filepath.Join(dir, "CANDIDATE_*.md"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		sort.Strings(matches)
		for _, p := range matches {
			d, err := fileDigest(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
				return 1
			}
			fmt.Printf("%s  %s\n", candidateSlug(p), d)
		}
		return 0
	case len(rest) == 2 && rest[0] == "--file":
		d, err := fileDigest(rest[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		fmt.Println(d)
		return 0
	case len(rest) == 1:
		matches, err := filepath.Glob(filepath.Join(dir, "CANDIDATE_"+rest[0]+"_*.md"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		switch len(matches) {
		case 0:
			fmt.Fprintf(os.Stderr, "corpos-lab: no candidate file for slug %q under %s\n", rest[0], dir)
			return 1
		case 1:
			d, err := fileDigest(matches[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
				return 1
			}
			fmt.Println(d)
			return 0
		default:
			fmt.Fprintf(os.Stderr, "corpos-lab: ambiguous slug %q: %v\n", rest[0], matches)
			return 1
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: corpos-lab glyph-digest <slug> | --all | --file <path> [-candidates DIR]")
		return 2
	}
}

// fileDigest reads a candidate file and returns its AC-4 glyph-block digest.
func fileDigest(path string) (string, error) {
	b, err := os.ReadFile(path) //nolint:gosec // a candidate file path from argv, not a secret
	if err != nil {
		return "", err
	}
	return battery.GlyphDigest(string(b))
}

// runGlyphLint runs the deterministic structural subset of the battery against
// one candidate — a fast, model-free pre-battery shape check. It lints the
// extracted entry, not the raw doc, so the AC-3 specimen does not trip the
// intent scan. The item logic lives in internal/battery (tested); this stays
// file glue, and it is the surface suggestion 166's schema enforcement reaches
// through once that item joins battery.BuildStructuralLint.
func runGlyphLint(args []string) int {
	registryPath := "corpus/private/glyph-model/ALPHABET.md"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "-registry" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -registry needs a value")
				return 2
			}
			i++
			registryPath = args[i]
			continue
		}
		rest = append(rest, args[i])
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab glyph-lint <candidate.md> [-registry ALPHABET.md]")
		return 2
	}
	path := rest[0]
	content, err := os.ReadFile(path) //nolint:gosec // a candidate file path from argv, not a secret
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}
	entry, err := battery.ExtractEntry(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	res := battery.RunSequence(context.Background(), battery.BuildStructuralLint(), battery.Input{
		ItemID:         candidateSlug(path),
		Content:        entry,
		Registry:       batteryrun.FileRegistryReader{Path: registryPath},
		ContinueOnFail: true,
	})
	for _, sr := range res.StepResults {
		fmt.Fprintf(os.Stderr, "  %-28s %s\n", sr.StepName, verdictLabel(sr.Outcome))
	}
	if res.Passed {
		fmt.Fprintf(os.Stderr, "corpos-lab: glyph-lint %q PASS (structural items)\n", candidateSlug(path))
		return 0
	}
	fmt.Fprintf(os.Stderr, "corpos-lab: glyph-lint %q FAIL: %s\n", candidateSlug(path), res.FailureReason)
	return 1
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
		Preflight:  preflightProbe(defPath),
	})
	// Preflight warnings are recorded on the run; surface them to the operator
	// too, on success or failure.
	for _, w := range runResult.Warnings {
		fmt.Fprintf(os.Stderr, "corpos-lab: WARN: %s\n", w)
	}
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
//
// When the ref is digest-pinned and the read fails — the shape that means a
// rebuild orphaned the pinned digest — it returns an *image.AbsentPinError that
// names the missing digest and the digests currently present for the repo,
// instead of the opaque podman "exit status 125" the raw failure produces. The
// enrichment logic lives in internal/image (tested); this stays exec glue.
func digestReader(ctx context.Context, ref string) (string, error) {
	digest, err := image.Digest(ctx, podmanInspectDigest, ref)
	if err == nil {
		return digest, nil
	}
	repo, missing, ok := image.ParsePinnedDigest(ref)
	if !ok {
		return "", err
	}
	// Best-effort: list what IS present for the repo. A failure here (podman
	// missing, no such repo) just yields an empty Available, which the error
	// message renders as "rebuild it".
	var available []string
	if out, listErr := exec.CommandContext(ctx, "podman", "images", "--format", "{{.Digest}}", repo).Output(); listErr == nil {
		available = image.ParseDigestList(out)
	}
	return "", &image.AbsentPinError{Ref: ref, Repo: repo, Digest: missing, Available: available}
}

// podmanInspectDigest reads one image's content digest via rootless podman.
func podmanInspectDigest(ctx context.Context, ref string) ([]byte, error) {
	return exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", ref).Output()
}

// preflightProbe warns when the pinned assay image predates the assay/runner
// code, so a run on it silently executes stale code. It compares the image's
// build time against the last git change to internal/assay or internal/runner.
// Every arm that cannot read a signal returns no warning rather than guessing —
// observe, don't assert — and it never blocks the run. The comparison logic
// lives in internal/image and internal/provenance (tested); this is exec glue.
func preflightProbe(defPath string) control.PreflightProbe {
	return func(ctx context.Context, def study.Def) []string {
		built, err := image.Created(ctx, podmanInspectCreated, def.Image)
		if err != nil {
			return nil
		}
		repoDir, err := filepath.Abs(filepath.Dir(defPath))
		if err != nil {
			return nil
		}
		changed, err := provenance.LastChange(ctx, repoDir, "internal/assay", "internal/runner")
		if err != nil {
			return nil
		}
		if msg, stale := image.StaleImageWarning(def.Image, built, changed); stale {
			return []string{msg}
		}
		return nil
	}
}

// podmanInspectCreated reads one image's build time (epoch seconds) via rootless
// podman.
func podmanInspectCreated(ctx context.Context, ref string) ([]byte, error) {
	return exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Created.Unix}}", ref).Output()
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
