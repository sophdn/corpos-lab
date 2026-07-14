// Command corpos-lab is the headless lab controller CLI. It runs a study
// end-to-end from a single definition file — reproducible with no agent in the
// loop — and is the surface corpos drives as an external tool.
//
// Usage:
//
//	corpos-lab run-study <def.toml> [-work DIR]
//
// Exit 0 on a completed study, non-zero on any failure (the failed-run record
// is written regardless).
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"corpos-lab/internal/control"
	"corpos-lab/internal/image"
	"corpos-lab/internal/persist"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/study"
	"corpos-lab/internal/substrate"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || args[0] != "run-study" {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab run-study <def.toml> [-work DIR]")
		return 2
	}
	return runStudy(args[1:])
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
	// --userns=keep-id maps the host user into the container; --user then runs
	// the process as that mapped uid/gid so it can write the host-owned /out
	// bind mount. Without --user the image's nonroot uid (65532) maps to an
	// unprivileged subuid that cannot write the host dir.
	userArg := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	argv := []string{
		"run", "--rm", "--userns=keep-id", "--user", userArg,
		"--network", spec.Network,
		"-v", spec.InDir + ":/in:ro,Z",
		"-v", spec.OutDir + ":/out:Z",
		spec.Image, "run",
	}
	cmd := exec.CommandContext(ctx, "podman", argv...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
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
