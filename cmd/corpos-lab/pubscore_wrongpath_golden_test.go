package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected and returns what it printed.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	ferr := fn()
	os.Stdout = orig
	_ = w.Close()
	return <-done, ferr
}

// TestWrongpathScoreGolden freezes the wrongpath-score report over a small runs
// tree: scored cells (with results.json) and an unscored one, a qwen38 base cell
// whose responses have reasoning traces (one missing), and a file that is not a
// grounded_glyph response (ignored).
func TestWrongpathScoreGolden(t *testing.T) {
	runs := t.TempDir()
	cell := func(arm, model, gab string, scored bool, resps map[string]string) {
		out := filepath.Join(runs, arm, model, gab, "out")
		for name, body := range resps {
			writeFileT(t, filepath.Join(out, "responses", name), body)
		}
		if scored {
			writeFileT(t, filepath.Join(out, "results.json"), "{}")
		}
	}
	cell("base", "mistral", "cas-a", true, map[string]string{
		"grounded_glyph_1.txt": "VERDICT: yes\nFIELD SOURCE: scope", "grounded_glyph_2.txt": "VERDICT: no\nFIELD SOURCE: aim",
		"notes.txt": "VERDICT: yes",
	})
	cell("ablated", "mistral", "cas-a", true, map[string]string{"grounded_glyph_1.txt": "VERDICT: no"})
	cell("base", "qwen38", "cgu-b", true, map[string]string{
		"grounded_glyph_1.txt": "VERDICT: no\nFIELD SOURCE: marker", "grounded_glyph_2.txt": "VERDICT: no\nFIELD SOURCE: pull",
	})
	writeFileT(t, filepath.Join(runs, "base", "qwen38", "cgu-b", "out", "reasoning", "grounded_glyph_1.txt"), "the operative clause")
	cell("base", "qwen2532", "fsb-a", false, map[string]string{"grounded_glyph_1.txt": "VERDICT: yes\nFIELD SOURCE: rest"})

	got, err := captureStdout(t, func() error { return raWrongpathScore([]string{"--runs", runs}) })
	if err != nil {
		t.Fatalf("score: %v", err)
	}
	for _, bad := range [][]string{{}, {"--runs", runs, "extra"}} {
		if err := raWrongpathScore(bad); err != nil {
			got += "error: " + strings.ReplaceAll(err.Error(), runs, "<runs>") + "\n"
		}
	}
	golden := filepath.Join("testdata", "wrongpath_score.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("wrongpath-score drifted from %s:\n%s", golden, got)
	}
}
