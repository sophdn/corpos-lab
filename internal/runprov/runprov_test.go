package runprov

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/control"
	"corpos-lab/internal/model"
	"corpos-lab/internal/runner"
)

func writeRecord(t *testing.T, path string, run control.StudyRun) {
	t.Helper()
	b, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func run(modelID, build string, seeds []int, truncated bool, mismatch string) control.StudyRun {
	return control.StudyRun{
		ImageDigest: "sha256:aaa",
		Results: &runner.Results{
			ModelID:       modelID,
			Sampler:       assay.Sampling{Temperature: 0.8, MaxTokens: 1024, MinP: 0.05, Seeds: seeds},
			Server:        model.ServerProps{BuildInfo: build, NCtx: 16384},
			ModelMismatch: mismatch,
			Rows: []assay.ScoreRow{
				{Observed: assay.Observed{BuildInfo: build, Truncated: truncated}},
				{Observed: assay.Observed{BuildInfo: build}},
			},
		},
	}
}

func TestLoadGroupsByModel(t *testing.T) {
	dir := t.TempDir()
	writeRecord(t, filepath.Join(dir, "a", "runs", "s1", "run-record.json"), run("qwen", "b1", []int{1, 2}, true, ""))
	writeRecord(t, filepath.Join(dir, "a", "runs", "s2", "run-record.json"), run("qwen", "b2", []int{2, 3}, false, "served other"))
	writeRecord(t, filepath.Join(dir, "b", "run-record.json"), run("mistral", "b1", []int{1}, false, ""))
	unread := run("mistral", "", []int{1}, false, "")
	unread.Results.ServerReadbackError = "connection refused"
	unread.Results.Server = model.ServerProps{}
	writeRecord(t, filepath.Join(dir, "c", "run-record.json"), unread)
	writeRecord(t, filepath.Join(dir, "d", "run-record.json"), control.StudyRun{Status: "failed"})

	models, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].Model != "mistral" || models[1].Model != "qwen" {
		t.Fatalf("models = %+v", models)
	}
	q := models[1]
	if q.Runs != 2 || q.Rows != 4 || q.Truncated != 1 || q.Mismatched != 1 ||
		strings.Join(q.Builds, ",") != "b1,b2" || len(q.Seeds) != 3 || len(q.Samplers) != 1 || q.NCtx[0] != 16384 {
		t.Errorf("qwen = %+v", q)
	}
	if m := models[0]; m.Unread != 1 || m.Runs != 2 {
		t.Errorf("mistral = %+v", m)
	}

	var tbl, js bytes.Buffer
	if err := WriteTable(&tbl, models); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tbl.String(), "b1, b2") || !strings.Contains(tbl.String(), "min_p=0.05") {
		t.Errorf("table:\n%s", tbl.String())
	}
	if err := WriteJSON(&js, models); err != nil {
		t.Fatal(err)
	}
	if orNone(nil) != "(none recorded)" {
		t.Error("orNone(nil)")
	}
}

func TestLoadErrors(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Error("empty dir: want an error")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "run-record.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Error("bad json: want an error")
	}
	if _, err := Load(filepath.Join(dir, "absent")); err == nil {
		t.Error("absent dir: want an error")
	}
}
