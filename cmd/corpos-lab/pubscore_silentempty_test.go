package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the regression tests for bug 1369: the scoring slice-builders and
// aggregators returned a success-shaped empty/partial result (exit 0, no error)
// when a study's run-dir names or conditions did not match the tool's fixed
// regex / glob / hardcoded class-condition set. Each test points the real CLI
// entrypoint at a fixture that triggers one silent-empty path and asserts the
// entrypoint now returns a non-nil error (which runPubScore maps to exit 1).
// Before the fix every one of these returns nil.

func writeFileT(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeJSONT(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	writeFileT(t, path, string(b))
}

// Path 1a: grounded-build-slices pointed at a study whose run dirs do not match
// the gnp-* glob (the scramble study's scb-* dirs). The glob matches nothing, so
// zero items are built — previously "wrote 0 items across 0 classes", exit 0.
func TestGroundedBuildSlicesErrorsOnGlobMiss(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, "scramble", "runs", "scb-foo-s1-mistral", "out", "results.json"), `{"rows":[]}`)
	err := pgBuildSlices([]string{"--root", root})
	if err == nil {
		t.Fatal("expected an error on a name-mismatch study (glob matched no run dirs), got nil")
	}
	if !strings.Contains(err.Error(), "0 items") && !strings.Contains(err.Error(), "nothing was scored") {
		t.Fatalf("error should name the empty result; got %q", err)
	}
}

// Path 1b: grounded-build-slices where the glob matches run dirs but ParseRunDir
// matches none of them (a gnp-* dir whose name fails the regex). Previously the
// rows stayed empty and the summary printed a default-class table, exit 0.
func TestGroundedBuildSlicesErrorsOnParseMiss(t *testing.T) {
	root := t.TempDir()
	writeFileT(t, filepath.Join(root, "c", "runs", "gnp-badname", "out", "results.json"), `{"rows":[]}`)
	err := pgBuildSlices([]string{"--root", root})
	if err == nil {
		t.Fatal("expected an error when the glob matched run dirs but ParseRunDir matched none, got nil")
	}
	if !strings.Contains(err.Error(), "ParseRunDir") {
		t.Fatalf("error should name the dirs-found-vs-parsed mismatch; got %q", err)
	}
}

// Path 2: length-aggregate where the key carries a condition (scrambled_glyph)
// absent from the hardcoded reported condition list. Previously those rows
// silently dropped from the table, exit 0.
func TestLengthAggregateErrorsOnUnreportedCondition(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.json")
	writeJSONT(t, keyPath, map[string]any{
		"r0": map[string]any{"cls": "classA", "scenario": "1", "model": "mistral", "condition": "baseline", "run": 1},
		"r1": map[string]any{"cls": "classA", "scenario": "1", "model": "mistral", "condition": "scrambled_glyph", "run": 1},
	})
	raterPath := filepath.Join(dir, "rater_a.json")
	writeJSONT(t, raterPath, map[string]string{"r0": "C", "r1": "N"})
	err := ldAggregate([]string{"--key", keyPath, "--rater-a", raterPath})
	if err == nil {
		t.Fatal("expected an error when the key contains a condition absent from the reported list, got nil")
	}
	if !strings.Contains(err.Error(), "scrambled_glyph") {
		t.Fatalf("error should name the dropped condition; got %q", err)
	}
}

// Sibling (grounded-aggregate): the same condition-absent-from-hardcoded-list
// drop in the Aggregate tool (aggregate.go's Conds/Classes).
func TestGroundedAggregateErrorsOnUnreportedCondition(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.json")
	writeJSONT(t, keyPath, map[string]any{
		"r0": map[string]any{"cls": "post-write-verification-absent", "scenario": "1", "model": "mistral", "condition": "baseline", "run": 1},
		"r1": map[string]any{"cls": "post-write-verification-absent", "scenario": "1", "model": "mistral", "condition": "scrambled_glyph", "run": 1},
	})
	raterPath := filepath.Join(dir, "rater_a.json")
	writeJSONT(t, raterPath, map[string]string{"r0": "C", "r1": "N"})
	err := pgAggregate([]string{"--key", keyPath, "--rater-a", raterPath})
	if err == nil {
		t.Fatal("expected an error when the key contains a condition absent from Conds, got nil")
	}
	if !strings.Contains(err.Error(), "scrambled_glyph") {
		t.Fatalf("error should name the dropped condition; got %q", err)
	}
}

// Path (alphabetassay): alphabet-collect-assay pointed at a study with no run
// dirs matching the per-class asy-* globs. Previously it printed "<cls>: 0
// responses" for every class and "done", exit 0.
func TestAlphabetCollectAssayErrorsOnEmpty(t *testing.T) {
	study := t.TempDir()
	out := filepath.Join(t.TempDir(), "collected")
	err := paCollectAssay([]string{study, out})
	if err == nil {
		t.Fatal("expected an error when no run dirs matched (zero responses collected), got nil")
	}
	if !strings.Contains(err.Error(), "collected 0 responses") && !strings.Contains(err.Error(), "nothing was collected") {
		t.Fatalf("error should name the empty result; got %q", err)
	}
}
