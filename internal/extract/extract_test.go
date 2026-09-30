package extract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteProducesRegistryLabShape(t *testing.T) {
	out := t.TempDir()
	m := Manifest{
		Image:       "lab-grounded-glyph-probe",
		Mode:        "run",
		StartedAt:   "2026-07-08T12:00:00Z",
		FinishedAt:  "2026-07-08T12:01:00Z",
		ExitCode:    0,
		ResultsPath: "/out/results.json",
	}
	if err := Write(out, m); err != nil {
		t.Fatalf("Write: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(out, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var back Manifest
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("manifest not valid JSON: %v", err)
	}
	if back != m {
		t.Fatalf("round trip changed manifest: %+v -> %+v", m, back)
	}
	// Field names must match registry-lab's contract exactly.
	var keyed map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keyed); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"image", "mode", "started_at", "finished_at", "exit_code", "results_path"} {
		if _, ok := keyed[k]; !ok {
			t.Errorf("manifest missing contract field %q", k)
		}
	}
}

func TestWriteCreatesMissingOutDir(t *testing.T) {
	// A crash before the assay made /out must still record a manifest.
	nested := filepath.Join(t.TempDir(), "does", "not", "exist")
	if err := Write(nested, Manifest{Mode: "run", ExitCode: 2}); err != nil {
		t.Fatalf("Write should create the dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, "manifest.json")); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
}

func TestWriteErrorsWhenOutDirPathIsAFile(t *testing.T) {
	// A file where the out dir should be makes MkdirAll fail — Write must
	// surface it, not swallow it.
	base := t.TempDir()
	blocker := filepath.Join(base, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Try to write into a dir under the file, which MkdirAll can't create.
	if err := Write(filepath.Join(blocker, "out"), Manifest{Mode: "run"}); err == nil {
		t.Fatal("expected error when out dir path is under a file")
	}
}

func TestWriteRecordsCrashExitCode(t *testing.T) {
	out := t.TempDir()
	if err := Write(out, Manifest{Mode: "run", ExitCode: 1}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(out, "manifest.json"))
	var back Manifest
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", back.ExitCode)
	}
}
