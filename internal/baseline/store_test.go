package baseline

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadCapturesReadsAndSorts(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "b.json"), `{"item":"pwv","model_id":"Qwen2.5-32B","runs":[{"run":1,"slice_id":"pwv|Qwen2.5-32B|run1"}]}`)
	writeFile(t, filepath.Join(dir, "a.json"), `{"item":"pwv","model_id":"Mistral-7B","runs":[{"run":1,"slice_id":"pwv|Mistral-7B|run1"}]}`)
	got, err := LoadCaptures(dir)
	if err != nil {
		t.Fatalf("LoadCaptures: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("captures = %d, want 2", len(got))
	}
	// a.json sorts before b.json.
	if got[0].ModelID != "Mistral-7B" || got[1].ModelID != "Qwen2.5-32B" {
		t.Errorf("captures not sorted by path: %q, %q", got[0].ModelID, got[1].ModelID)
	}
}

func TestLoadCapturesEmptyDir(t *testing.T) {
	got, err := LoadCaptures(t.TempDir())
	if err != nil {
		t.Fatalf("LoadCaptures: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("captures = %d, want 0 for an empty dir", len(got))
	}
}

func TestLoadCapturesRejectsMalformed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bad.json"), `{not json`)
	if _, err := LoadCaptures(dir); err == nil {
		t.Fatal("want an error on a malformed capture")
	}
}

func TestLoadRaterScoresMergesFamiliesAndSkipsProv(t *testing.T) {
	dir := t.TempDir()
	// Family devstral: two result files merge into one map.
	writeFile(t, filepath.Join(dir, "devstral", "pwv__mistral.json"), `{"pwv|Mistral-7B|run1":"C"}`)
	writeFile(t, filepath.Join(dir, "devstral", "pwv__qwen.json"), `{"pwv|Qwen2.5-32B|run1":"I"}`)
	// A provenance sidecar must be skipped, not parsed as codes.
	writeFile(t, filepath.Join(dir, "devstral", "pwv__mistral.json.prov.json"), `{"model_reported":"x"}`)
	// Family deepseek: one file.
	writeFile(t, filepath.Join(dir, "deepseek", "pwv__mistral.json"), `{"pwv|Mistral-7B|run1":"C"}`)
	// A stray file at the root is not a rater dir and is ignored.
	writeFile(t, filepath.Join(dir, "README.md"), "notes")

	maps, ids, err := LoadRaterScores(dir)
	if err != nil {
		t.Fatalf("LoadRaterScores: %v", err)
	}
	if len(maps) != 2 || len(ids) != 2 {
		t.Fatalf("families = %d/%d, want 2", len(maps), len(ids))
	}
	// ReadDir returns entries sorted: deepseek before devstral.
	if ids[0] != "deepseek" || ids[1] != "devstral" {
		t.Errorf("family order = %v, want [deepseek devstral]", ids)
	}
	devstral := maps[1]
	if devstral["pwv|Mistral-7B|run1"] != "C" || devstral["pwv|Qwen2.5-32B|run1"] != "I" {
		t.Errorf("devstral merge wrong: %v", devstral)
	}
	if len(devstral) != 2 {
		t.Errorf("devstral has %d entries, want 2 (prov sidecar skipped)", len(devstral))
	}
}

func TestLoadRaterScoresMissingDir(t *testing.T) {
	if _, _, err := LoadRaterScores(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("want an error for a missing scores dir")
	}
}

func TestLoadRaterScoresRejectsMalformedResult(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "devstral", "bad.json"), `{oops`)
	if _, _, err := LoadRaterScores(dir); err == nil {
		t.Fatal("want an error on a malformed rater result")
	}
}
