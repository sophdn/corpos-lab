package control

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/runner"
	"corpos-lab/internal/study"
	"corpos-lab/internal/substrate"
)

const validDef = `
name = "smoke"
assay = "grounded-glyph-probe"
item_id = "casg-direct"
image = "localhost/lab-grounded-glyph-probe:dev"
network = "corpos-net"
conditions = ["baseline", "glyph_only"]
runs_per_cell = 1

[model]
base_url = "http://llama-server:8081/v1"
model_id = "qwen"
version = "q4"

[materials]
scenario = "scenario.md"
glyph = "glyph.md"
ground = "ground.md"

[sampling]
temperature = 0.8
seeds = [1]
max_tokens = 512
top_n_sigma = -1.0
top_k = 0
typical_p = 1.0
top_p = 1.0
min_p = 0.05
repeat_penalty = 1.0
repeat_last_n = 0
presence_penalty = 0.0
frequency_penalty = 0.0
xtc_probability = 0.0
dry_multiplier = 0.0
`

func loadTestDef(t *testing.T) study.Def {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "study.toml"), []byte(validDef), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"scenario.md", "glyph.md", "ground.md"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(n+"-content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d, err := study.LoadDef(filepath.Join(dir, "study.toml"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func fixedDigest(_ context.Context, _ string) (string, error) {
	return "sha256:" + "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"[:64], nil
}

// testDeps supplies stub host probes for tests whose subject is the control
// flow. Deps requires them explicitly rather than defaulting to nil, so a test
// cannot accidentally assert against a controller that observes nothing — the
// shape that let internal/provenance sit dead behind a doc claiming otherwise.
func testDeps(l Launcher, d DigestReader) Deps {
	return Deps{
		Launcher: l,
		DigestOf: d,
		Substrate: func(context.Context) substrate.Info {
			return substrate.Info{Kind: substrate.KindGPU, Device: "test-gpu", DetectedBy: "stub"}
		},
		Provenance: func(context.Context) (provenance.Stamp, error) {
			return provenance.Stamp{CommitSHA: "0123456789abcdef0123456789abcdef01234567"}, nil
		},
	}
}

// fakeLauncher writes a canned /out (results.json + manifest.json) and returns
// a chosen exit code, standing in for a real container run.
type fakeLauncher struct {
	exitCode      int
	stderr        string
	launchErr     error
	writeResults  bool
	rows          int
	badJSON       bool
	wrongIdentity bool
	noExtraction  bool
	badExtraction bool
}

func (f *fakeLauncher) Launch(_ context.Context, spec LaunchSpec) (LaunchResult, error) {
	if f.launchErr != nil {
		return LaunchResult{}, f.launchErr
	}
	if err := os.MkdirAll(spec.OutDir, 0o755); err != nil {
		return LaunchResult{}, err
	}
	// Drop an extraction manifest, like the real container (unless the test
	// simulates a missing or corrupt one).
	switch {
	case f.noExtraction:
		// leave it absent
	case f.badExtraction:
		_ = os.WriteFile(filepath.Join(spec.OutDir, "manifest.json"), []byte("{corrupt"), 0o644)
	default:
		ext := map[string]any{"image": "lab-grounded-glyph-probe", "mode": "run", "exit_code": f.exitCode, "results_path": "/out/results.json"}
		extRaw, _ := json.Marshal(ext)
		_ = os.WriteFile(filepath.Join(spec.OutDir, "manifest.json"), extRaw, 0o644)
	}

	if f.writeResults {
		var raw []byte
		if f.badJSON {
			raw = []byte("{not json")
		} else {
			results := runner.Results{Assay: "grounded-glyph-probe", ItemID: "casg-direct", ModelID: "qwen"}
			if f.wrongIdentity {
				results.Assay = "decomposition"
			}
			for i := 0; i < f.rows; i++ {
				results.Rows = append(results.Rows, assay.ScoreRow{
					Item: "casg-direct", Condition: assay.Baseline, Run: i + 1, Score: assay.Unscored, Rationale: "r",
				})
			}
			raw, _ = json.MarshalIndent(results, "", "  ")
		}
		_ = os.WriteFile(filepath.Join(spec.OutDir, "results.json"), raw, 0o644)
	}
	return LaunchResult{ExitCode: f.exitCode, Stderr: f.stderr}, nil
}

func TestRunStudyCompletesAndRecords(t *testing.T) {
	def := loadTestDef(t)
	work := t.TempDir()
	// 2 conditions × 1 run = 2 rows expected.
	run, err := RunStudy(context.Background(), def, work, testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 2}, fixedDigest))
	if err != nil {
		t.Fatalf("RunStudy: %v", err)
	}
	if run.Status != StatusCompleted {
		t.Fatalf("status = %s, want completed", run.Status)
	}
	if run.Results == nil || len(run.Results.Rows) != 2 {
		t.Fatalf("results not collected: %+v", run.Results)
	}
	if run.ImageDigest == "" || run.Manifest.Image.Digest == "" {
		t.Fatal("run not digest-pinned")
	}
	if run.Extraction == nil {
		t.Fatal("extraction manifest not collected")
	}
	// Run record written and re-parseable.
	raw, err := os.ReadFile(filepath.Join(work, "run-record.json"))
	if err != nil {
		t.Fatalf("run record not written: %v", err)
	}
	var back StudyRun
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("run record invalid: %v", err)
	}
	if back.Status != StatusCompleted {
		t.Fatalf("recorded status = %s", back.Status)
	}
}

func TestRunStudyContainerNonZeroExitFailsWithRecord(t *testing.T) {
	def := loadTestDef(t)
	work := t.TempDir()
	run, err := RunStudy(context.Background(), def, work, testDeps(&fakeLauncher{exitCode: 1, stderr: "model unreachable"}, fixedDigest))
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
	var ce *ContainerExitError
	if !errors.As(err, &ce) {
		t.Fatalf("expected ContainerExitError, got %T", err)
	}
	if ce.ExitCode != 1 {
		t.Fatalf("exit code = %d", ce.ExitCode)
	}
	if run.Status != StatusFailed || run.Error == "" {
		t.Fatalf("failed run not recorded: %+v", run)
	}
	// The failed-run row is persisted — never a silent skip.
	if _, err := os.Stat(filepath.Join(work, "run-record.json")); err != nil {
		t.Fatalf("failed run record not written: %v", err)
	}
}

func TestRunStudyMalformedResultsFails(t *testing.T) {
	work := t.TempDir()
	// Exit 0 but no results.json at all.
	run, err := RunStudy(context.Background(), loadTestDef(t), work, testDeps(&fakeLauncher{exitCode: 0, writeResults: false}, fixedDigest))
	var me *MalformedResultsError
	if !errors.As(err, &me) {
		t.Fatalf("expected MalformedResultsError, got %T (%v)", err, err)
	}
	if run.Status != StatusFailed {
		t.Fatal("status should be failed")
	}
}

func TestRunStudyBadJSONResultsFails(t *testing.T) {
	_, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(&fakeLauncher{exitCode: 0, writeResults: true, badJSON: true}, fixedDigest))
	var me *MalformedResultsError
	if !errors.As(err, &me) {
		t.Fatalf("expected MalformedResultsError for bad JSON, got %T", err)
	}
}

func TestRunStudyShortGridFails(t *testing.T) {
	// Exit 0 but only 1 row when 2 are expected — a truncated grid.
	_, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 1}, fixedDigest))
	var me *MalformedResultsError
	if !errors.As(err, &me) {
		t.Fatalf("expected MalformedResultsError for short grid, got %T", err)
	}
}

func TestRunStudyWrongIdentityFails(t *testing.T) {
	_, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 2, wrongIdentity: true}, fixedDigest))
	var me *MalformedResultsError
	if !errors.As(err, &me) {
		t.Fatalf("expected MalformedResultsError for identity mismatch, got %T", err)
	}
}

func TestRunStudyLauncherErrorFails(t *testing.T) {
	run, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(&fakeLauncher{launchErr: errors.New("podman missing")}, fixedDigest))
	if err == nil {
		t.Fatal("expected launcher error")
	}
	if run.Status != StatusFailed {
		t.Fatal("status should be failed")
	}
}

func TestRunStudyDigestErrorFails(t *testing.T) {
	badDigest := func(_ context.Context, _ string) (string, error) {
		return "", errors.New("no such image")
	}
	run, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(&fakeLauncher{}, badDigest))
	if err == nil {
		t.Fatal("expected digest error")
	}
	if run.Status != StatusFailed {
		t.Fatal("status should be failed")
	}
}

func TestRunStudyMaterializeErrorFails(t *testing.T) {
	def := loadTestDef(t)
	// A file where the work dir should be makes Materialize's MkdirAll fail.
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := RunStudy(context.Background(), def, filepath.Join(blocker, "work"), testDeps(&fakeLauncher{}, fixedDigest))
	if err == nil {
		t.Fatal("expected materialize error")
	}
}

func TestRunStudyToleratesMissingOrBadExtraction(t *testing.T) {
	// A missing or corrupt extraction manifest is diagnostic-only: the run
	// still completes and Extraction is simply nil.
	for _, tc := range []struct {
		name string
		l    *fakeLauncher
	}{
		{"missing", &fakeLauncher{exitCode: 0, writeResults: true, rows: 2, noExtraction: true}},
		{"corrupt", &fakeLauncher{exitCode: 0, writeResults: true, rows: 2, badExtraction: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), testDeps(tc.l, fixedDigest))
			if err != nil {
				t.Fatalf("run should complete despite %s extraction: %v", tc.name, err)
			}
			if run.Status != StatusCompleted {
				t.Fatalf("status = %s", run.Status)
			}
			if run.Extraction != nil {
				t.Fatalf("expected nil extraction for %s manifest", tc.name)
			}
		})
	}
}

func TestErrorTypeMessages(t *testing.T) {
	ce := &ContainerExitError{ExitCode: 2, Stderr: "boom"}
	if ce.Error() == "" || !contains(ce.Error(), "exited 2") {
		t.Fatalf("ContainerExitError message: %q", ce.Error())
	}
	me := &MalformedResultsError{Path: "/out/results.json", Reason: "short"}
	if !contains(me.Error(), "malformed results") {
		t.Fatalf("MalformedResultsError message: %q", me.Error())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// The two host-side facts a run could never report on its own. The container
// cannot stamp a repo it has no copy of, and no llama.cpp endpoint names the
// processor — so if the controller doesn't record these, nothing does. That was
// literally the state: internal/provenance was dead code while CLAUDE.md
// claimed every result row carried the repo commit.
func TestRunStudyRecordsSubstrateAndProvenance(t *testing.T) {
	work := t.TempDir()
	run, err := RunStudy(context.Background(), loadTestDef(t), work,
		testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 2}, fixedDigest))
	if err != nil {
		t.Fatalf("RunStudy: %v", err)
	}
	if run.Substrate.Kind != substrate.KindGPU || run.Substrate.Device != "test-gpu" {
		t.Errorf("substrate not recorded: %+v", run.Substrate)
	}
	if run.Provenance.CommitSHA != "0123456789abcdef0123456789abcdef01234567" {
		t.Errorf("provenance not recorded: %+v", run.Provenance)
	}

	// And they survive to the artifact that outlives the process.
	raw, err := os.ReadFile(filepath.Join(work, "run-record.json"))
	if err != nil {
		t.Fatal(err)
	}
	var back StudyRun
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.Substrate.Kind != substrate.KindGPU || back.Provenance.CommitSHA == "" {
		t.Errorf("run-record.json lost the host observations: %+v", back)
	}
}

// A run that cannot describe itself is still a run. Refusing one here would be
// the freeze reflex — the rule that told an agent to file a broken instrument
// and keep going, inverted but the same mistake: letting bookkeeping decide
// whether real work counts.
func TestRunStudyRecordsProvenanceFailureWithoutFailingTheRun(t *testing.T) {
	deps := testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 2}, fixedDigest)
	deps.Provenance = func(context.Context) (provenance.Stamp, error) {
		return provenance.Stamp{}, errors.New("not a git repository")
	}
	run, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), deps)
	if err != nil {
		t.Fatalf("a failed provenance stamp must not fail the run: %v", err)
	}
	if run.Status != StatusCompleted {
		t.Errorf("status = %s, want completed", run.Status)
	}
	if !contains(run.ProvenanceError, "not a git repository") {
		t.Errorf("provenance error not recorded: %q", run.ProvenanceError)
	}
	if run.Provenance.CommitSHA != "" {
		t.Error("a failed stamp must leave the commit empty, not invented")
	}
}

// An unknown substrate is recorded as unknown, not silently omitted or
// upgraded to a guess.
func TestRunStudyRecordsUnknownSubstrateHonestly(t *testing.T) {
	deps := testDeps(&fakeLauncher{exitCode: 0, writeResults: true, rows: 2}, fixedDigest)
	deps.Substrate = func(context.Context) substrate.Info {
		return substrate.Info{Kind: substrate.KindUnknown, DetectedBy: "none", Note: "nvidia-smi unavailable"}
	}
	run, err := RunStudy(context.Background(), loadTestDef(t), t.TempDir(), deps)
	if err != nil {
		t.Fatalf("RunStudy: %v", err)
	}
	if run.Substrate.Kind != substrate.KindUnknown || run.Substrate.DetectedBy != "none" {
		t.Errorf("unknown substrate must be recorded as unknown: %+v", run.Substrate)
	}
}

// A failed run is exactly when knowing the host matters most — and the record
// is written for failures too, so the observations must be on it.
func TestRunStudyRecordsHostObservationsOnFailedRuns(t *testing.T) {
	work := t.TempDir()
	run, _ := RunStudy(context.Background(), loadTestDef(t), work,
		testDeps(&fakeLauncher{exitCode: 1, stderr: "model unreachable"}, fixedDigest))
	if run.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", run.Status)
	}
	if run.Substrate.Kind != substrate.KindGPU || run.Provenance.CommitSHA == "" {
		t.Errorf("a failed run must still say what it ran on: substrate=%+v provenance=%+v",
			run.Substrate, run.Provenance)
	}
}
