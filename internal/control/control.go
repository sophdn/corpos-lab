// Package control is the host-side lab controller — the piece the old system
// never wired (e9-lab-controller). It takes a study definition, materializes
// the container /in, pins the run by digest, launches the disposable assay
// container, and collects + validates /out. Every outcome, success or failure,
// produces a persisted run record: a failure is a typed error plus a written
// failed-run row, never a silent skip.
package control

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"corpos-lab/internal/extract"
	"corpos-lab/internal/manifest"
	"corpos-lab/internal/runner"
	"corpos-lab/internal/study"
)

// LaunchSpec is one container launch: the pinned image, the bind-mount dirs,
// and the podman network to join.
type LaunchSpec struct {
	Image   string
	InDir   string
	OutDir  string
	Network string
}

// LaunchResult is the outcome of a container launch — the process exit code
// and captured stderr. A non-zero ExitCode is a run failure the controller
// classifies; the launcher itself only errors when podman could not run at all.
type LaunchResult struct {
	ExitCode int
	Stderr   string
}

// Launcher runs one assay container to completion. Injectable so the control
// flow is sans-IO testable; the production launcher shells to rootless podman
// (see cmd/corpos-lab).
type Launcher interface {
	Launch(ctx context.Context, spec LaunchSpec) (LaunchResult, error)
}

// DigestReader returns the content digest of an image ref (podman-backed in
// production; injected in tests).
type DigestReader func(ctx context.Context, ref string) (string, error)

// Status is a study run's terminal state.
type Status string

// Run statuses.
const (
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// StudyRun is the durable record of one study run — the artifact the
// persistence layer (persist-and-dashboard) ingests. It is written for every
// run, including failures.
type StudyRun struct {
	Name        string               `json:"name"`
	Assay       string               `json:"assay"`
	ItemID      string               `json:"item_id"`
	Image       string               `json:"image"`
	ImageDigest string               `json:"image_digest"`
	Status      Status               `json:"status"`
	Error       string               `json:"error,omitempty"`
	Manifest    manifest.RunManifest `json:"manifest"`
	Results     *runner.Results      `json:"results,omitempty"`
	Extraction  *extract.Manifest    `json:"extraction,omitempty"`
}

// ContainerExitError is a non-zero container exit — including the
// model-unreachable case, which surfaces as the assay binary exiting 1.
type ContainerExitError struct {
	ExitCode int
	Stderr   string
}

func (e *ContainerExitError) Error() string {
	return fmt.Sprintf("control: container exited %d: %s", e.ExitCode, e.Stderr)
}

// MalformedResultsError is a container that exited 0 but produced an absent,
// unparseable, or incomplete /out/results.json.
type MalformedResultsError struct {
	Path   string
	Reason string
}

func (e *MalformedResultsError) Error() string {
	return fmt.Sprintf("control: malformed results at %s: %s", e.Path, e.Reason)
}

// RunStudy executes def end-to-end under workDir and writes a run record to
// workDir/run-record.json. It returns the StudyRun (always) and a typed error
// (nil on success). On any failure the returned StudyRun has Status=Failed
// with the error recorded, and the run record is still written.
func RunStudy(ctx context.Context, def study.Def, workDir string, launcher Launcher, digestOf DigestReader) (StudyRun, error) {
	inDir := filepath.Join(workDir, "in")
	outDir := filepath.Join(workDir, "out")

	run := StudyRun{
		Name:   def.Name,
		Assay:  def.Assay,
		ItemID: def.ItemID,
		Image:  def.Image,
	}

	// Any exit path writes the record; a failure records the error first.
	finish := func(status Status, err error) (StudyRun, error) {
		run.Status = status
		if err != nil {
			run.Error = err.Error()
		}
		if writeErr := writeRecord(workDir, run); writeErr != nil {
			// A record we can't persist is itself a hard failure — never
			// let a run vanish silently. Carry the original outcome (if any)
			// in the message; only one error can be wrapped.
			orig := "none"
			if err != nil {
				orig = err.Error()
			}
			return run, fmt.Errorf("control: run %s ended %s but its record could not be written (original outcome: %s): %w",
				def.Name, status, orig, writeErr)
		}
		return run, err
	}

	if err := def.Materialize(inDir); err != nil {
		return finish(StatusFailed, err)
	}

	digest, err := digestOf(ctx, def.Image)
	if err != nil {
		return finish(StatusFailed, fmt.Errorf("control: read image digest for %s: %w", def.Image, err))
	}
	run.ImageDigest = digest

	pinned, err := manifest.Compute(inDir, manifest.ImagePin{Ref: def.Image, Digest: digest})
	if err != nil {
		return finish(StatusFailed, err)
	}
	run.Manifest = pinned

	res, err := launcher.Launch(ctx, LaunchSpec{
		Image:   def.Image,
		InDir:   inDir,
		OutDir:  outDir,
		Network: def.Network,
	})
	if err != nil {
		return finish(StatusFailed, fmt.Errorf("control: launch %s: %w", def.Image, err))
	}

	// The container's extraction manifest is best-effort context, collected
	// whether the run passed or failed.
	run.Extraction = readExtraction(outDir)

	if res.ExitCode != 0 {
		return finish(StatusFailed, &ContainerExitError{ExitCode: res.ExitCode, Stderr: res.Stderr})
	}

	results, err := collectResults(outDir, def)
	if err != nil {
		return finish(StatusFailed, err)
	}
	run.Results = results

	return finish(StatusCompleted, nil)
}

// collectResults reads and validates /out/results.json against the definition:
// it must parse, match the study's assay and item, and carry exactly the
// expected number of rows (conditions × runs_per_cell). A zero-exit container
// with a short grid is a malformed result, not a partial success.
func collectResults(outDir string, def study.Def) (*runner.Results, error) {
	path := filepath.Join(outDir, "results.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, &MalformedResultsError{Path: path, Reason: fmt.Sprintf("unreadable: %v", err)}
	}
	var results runner.Results
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, &MalformedResultsError{Path: path, Reason: fmt.Sprintf("unparseable: %v", err)}
	}
	if results.Assay != def.Assay || results.ItemID != def.ItemID {
		return nil, &MalformedResultsError{Path: path, Reason: fmt.Sprintf(
			"identity mismatch: results %s/%s vs def %s/%s", results.Assay, results.ItemID, def.Assay, def.ItemID)}
	}
	want := len(def.Conditions) * def.RunsPerCell
	if len(results.Rows) != want {
		return nil, &MalformedResultsError{Path: path, Reason: fmt.Sprintf(
			"expected %d rows (%d conditions × %d runs), got %d", want, len(def.Conditions), def.RunsPerCell, len(results.Rows))}
	}
	return &results, nil
}

// readExtraction reads the container's /out/manifest.json best-effort; a
// missing or bad manifest yields nil rather than an error (it is diagnostic
// context, not the result of record).
func readExtraction(outDir string) *extract.Manifest {
	raw, err := os.ReadFile(filepath.Join(outDir, "manifest.json"))
	if err != nil {
		return nil
	}
	var m extract.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return &m
}

func writeRecord(workDir string, run StudyRun) error {
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return fmt.Errorf("control: create work dir: %w", err)
	}
	raw, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return fmt.Errorf("control: marshal run record: %w", err)
	}
	path := filepath.Join(workDir, "run-record.json")
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("control: write run record: %w", err)
	}
	return nil
}
