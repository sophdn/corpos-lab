package rater

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"corpos-lab/internal/model"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func codeCReplies() *fakeSeam {
	return &fakeSeam{handler: func(_ int, _ string) (model.RawChatResult, error) {
		return model.RawChatResult{Content: "C"}, nil
	}}
}

func TestSliceLinesSkipsBlankAndParses(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "s.jsonl")
	writeFile(t, p, "{\"id\":\"a\",\"text\":\"x\"}\n\n  \n{\"id\":\"b\",\"text\":\"y\",\"scenario\":\"api-version\"}\n")
	lines, err := SliceLines(p)
	if err != nil {
		t.Fatalf("SliceLines: %v", err)
	}
	if len(lines) != 2 || lines[0].ID != "a" || lines[1].Scenario != "api-version" {
		t.Fatalf("lines = %+v", lines)
	}
}

func TestSliceLinesErrors(t *testing.T) {
	if _, err := SliceLines(filepath.Join(t.TempDir(), "missing.jsonl")); err == nil {
		t.Fatal("want error on a missing file")
	}
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.jsonl")
	writeFile(t, bad, "{not json}\n")
	if _, err := SliceLines(bad); err == nil {
		t.Fatal("want error on malformed json")
	}
}

func TestIsComplete(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.json")
	writeFile(t, good, `{"a":"C","b":"I"}`)
	if !IsComplete(good, []string{"a", "b"}) {
		t.Fatal("matching keys should be complete")
	}
	if IsComplete(good, []string{"a", "b", "c"}) {
		t.Fatal("a missing id should be incomplete")
	}
	if IsComplete(good, []string{"a", "z"}) {
		t.Fatal("a wrong id should be incomplete")
	}
	if IsComplete(filepath.Join(dir, "nope.json"), []string{"a"}) {
		t.Fatal("a missing file is not complete")
	}
	badJSON := filepath.Join(dir, "bad.json")
	writeFile(t, badJSON, "{oops")
	if IsComplete(badJSON, []string{"a"}) {
		t.Fatal("unparseable output is not complete")
	}
}

func TestRunSlicesScoresAndWritesProvenance(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "s0.jsonl")
	writeFile(t, slice, "{\"id\":\"a\",\"text\":\"x\"}\n{\"id\":\"b\",\"text\":\"y\"}\n")
	outDir := filepath.Join(dir, "scores")

	r := NewGrounded(Config{Seam: codeCReplies(), RaterID: "mech", Endpoint: "e"}, "RUBRIC")
	sum, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: outDir, Jobs: 1, WriteProvenance: true},
		[]string{slice})
	if err != nil {
		t.Fatalf("RunSlices: %v", err)
	}
	if sum.Scored != 1 {
		t.Fatalf("Scored = %d, want 1", sum.Scored)
	}
	canonical := filepath.Join(outDir, "mech", "s0.json")
	got := map[string]string{}
	data, err := os.ReadFile(canonical)
	if err != nil {
		t.Fatalf("read canonical: %v", err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("canonical not valid json: %v", err)
	}
	if got["a"] != "C" || got["b"] != "C" {
		t.Fatalf("scores = %v", got)
	}
	if _, err := os.Stat(canonical + ".prov.json"); err != nil {
		t.Fatalf("provenance sidecar missing: %v", err)
	}
}

func TestRunSlicesResumesSkippingComplete(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "s0.jsonl")
	writeFile(t, slice, "{\"id\":\"a\",\"text\":\"x\"}\n")
	outDir := filepath.Join(dir, "scores")
	// Pre-seed a valid canonical result; the run must skip it and not call the seam.
	writeFile(t, filepath.Join(outDir, "mech", "s0.json"), `{"a":"I"}`)

	seam := &fakeSeam{handler: func(_ int, _ string) (model.RawChatResult, error) {
		t.Error("seam must not be called for a complete slice")
		return model.RawChatResult{}, nil
	}}
	r := NewGrounded(Config{Seam: seam, RaterID: "mech"}, "RUBRIC")
	sum, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: outDir, Jobs: 1}, []string{slice})
	if err != nil {
		t.Fatalf("RunSlices: %v", err)
	}
	if sum.AlreadyComplete != 1 || sum.Scored != 0 {
		t.Fatalf("summary = %+v, want one already-complete", sum)
	}
}

func TestRunFileKeepsExistingOnCreateRace(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "s0.jsonl")
	writeFile(t, slice, "{\"id\":\"a\",\"text\":\"x\"}\n")
	outDir := filepath.Join(dir, "scores")
	// A canonical file whose keys do NOT match the slice: IsComplete is false, so
	// the slice is rated, but the create-only promotion then finds the file and
	// keeps it.
	writeFile(t, filepath.Join(outDir, "mech", "s0.json"), `{"stale":"X"}`)

	r := NewGrounded(Config{Seam: codeCReplies(), RaterID: "mech"}, "RUBRIC")
	sum, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: outDir, Jobs: 1}, []string{slice})
	if err != nil {
		t.Fatalf("RunSlices: %v", err)
	}
	if sum.KeptExisting != 1 {
		t.Fatalf("KeptExisting = %d, want 1", sum.KeptExisting)
	}
	// The existing (stale) file was not overwritten.
	data, _ := os.ReadFile(filepath.Join(outDir, "mech", "s0.json"))
	if string(data) != `{"stale":"X"}` {
		t.Fatalf("existing result was overwritten: %s", data)
	}
}

func TestRunSlicesConcurrentAllScored(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "scores")
	var paths []string
	for _, name := range []string{"s0", "s1", "s2", "s3"} {
		p := filepath.Join(dir, "slices", name+".jsonl")
		writeFile(t, p, "{\"id\":\"a\",\"text\":\"x\"}\n")
		paths = append(paths, p)
	}
	r := NewGrounded(Config{Seam: codeCReplies(), RaterID: "mech"}, "RUBRIC")
	sum, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: outDir, Jobs: 3}, paths)
	if err != nil {
		t.Fatalf("RunSlices: %v", err)
	}
	if sum.Scored != 4 {
		t.Fatalf("Scored = %d, want 4", sum.Scored)
	}
	if len(sum.Outcomes) != 4 {
		t.Fatalf("outcomes = %d, want 4", len(sum.Outcomes))
	}
}

func TestRunSlicesReturnsFirstError(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "s0.jsonl")
	writeFile(t, slice, "{\"id\":\"a\",\"text\":\"x\"}\n")
	boom := errors.New("down")
	seam := &fakeSeam{handler: func(_ int, _ string) (model.RawChatResult, error) {
		return model.RawChatResult{}, boom
	}}
	r := NewGrounded(Config{Seam: seam, RaterID: "mech"}, "RUBRIC")
	_, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: filepath.Join(dir, "out"), Jobs: 1}, []string{slice})
	if err == nil {
		t.Fatal("want an error when a slice fails")
	}
}

func TestRunSlicesJobsDefaultsToOne(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "s0.jsonl")
	writeFile(t, slice, "{\"id\":\"a\",\"text\":\"x\"}\n")
	r := NewGrounded(Config{Seam: codeCReplies(), RaterID: "mech"}, "RUBRIC")
	sum, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: filepath.Join(dir, "out"), Jobs: 0}, []string{slice})
	if err != nil {
		t.Fatalf("RunSlices: %v", err)
	}
	if sum.Scored != 1 {
		t.Fatalf("Scored = %d, want 1", sum.Scored)
	}
}

func TestDiscoverSlices(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"b.jsonl", "a.jsonl", "notes.txt"} {
		writeFile(t, filepath.Join(dir, n), "x")
	}
	got, err := DiscoverSlices(dir, "")
	if err != nil {
		t.Fatalf("DiscoverSlices dir: %v", err)
	}
	if len(got) != 2 || filepath.Base(got[0]) != "a.jsonl" || filepath.Base(got[1]) != "b.jsonl" {
		t.Fatalf("glob = %v, want sorted a,b jsonl", got)
	}
	single, err := DiscoverSlices("", "/tmp/one.jsonl")
	if err != nil {
		t.Fatalf("DiscoverSlices single: %v", err)
	}
	if len(single) != 1 || single[0] != "/tmp/one.jsonl" {
		t.Fatalf("single = %v", single)
	}
}

func TestDiscoverSlicesBadPatternErrors(t *testing.T) {
	if _, err := DiscoverSlices("[", ""); err == nil {
		t.Fatal("want an error on a malformed glob pattern")
	}
}

func TestRunSlicesSurfacesSliceReadError(t *testing.T) {
	dir := t.TempDir()
	slice := filepath.Join(dir, "slices", "bad.jsonl")
	writeFile(t, slice, "{not json}\n")
	r := NewGrounded(Config{Seam: codeCReplies(), RaterID: "mech"}, "RUBRIC")
	_, err := RunSlices(context.Background(), RunConfig{Rater: r, OutDir: filepath.Join(dir, "out"), Jobs: 1}, []string{slice})
	if err == nil {
		t.Fatal("want an error when a slice cannot be read")
	}
}

func TestWriteCreateOnlyErrorsOnMissingDir(t *testing.T) {
	p := filepath.Join(t.TempDir(), "no-such-dir", "x.json")
	if _, err := writeCreateOnly(p, []byte("x")); err == nil {
		t.Fatal("want an error writing into a missing directory")
	}
}

func TestWriteProvenanceErrorsOnBadPath(t *testing.T) {
	dir := t.TempDir()
	// The target path is an existing directory, so WriteFile fails.
	if err := writeProvenance(dir, &Provenance{ModelReported: "m"}); err == nil {
		t.Fatal("want an error writing provenance to a directory path")
	}
}

func TestWriteCreateOnly(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.json")
	created, err := writeCreateOnly(p, []byte("first"))
	if err != nil || !created {
		t.Fatalf("first write: created=%v err=%v", created, err)
	}
	created, err = writeCreateOnly(p, []byte("second"))
	if err != nil {
		t.Fatalf("second write err: %v", err)
	}
	if created {
		t.Fatal("second write must not create over an existing file")
	}
	data, _ := os.ReadFile(p)
	if string(data) != "first" {
		t.Fatalf("file overwritten: %s", data)
	}
}
