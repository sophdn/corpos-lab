package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStderr runs fn with os.Stderr redirected and returns what it printed.
func captureStderr(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	code := fn()
	os.Stderr = orig
	_ = w.Close()
	return <-done, code
}

// TestDisclosureGolden freezes runDisclosure's exit code and stderr for its
// argument errors, verify pass and fail, every promote failure, and a promote
// that removes a stale published glyph.
func TestDisclosureGolden(t *testing.T) {
	root := t.TempDir()
	w := func(rel, body string) string {
		p := filepath.Join(root, rel)
		writeFileT(t, p, body)
		return p
	}
	emptyPub := filepath.Join(root, "pub-empty")
	if err := os.MkdirAll(emptyPub, 0o750); err != nil {
		t.Fatal(err)
	}
	leakPub := filepath.Dir(w("pub-leak/leaked.md", "x"))
	stalePub := filepath.Dir(w("pub-stale/old.md", "x"))
	okMan := w("approved.txt", "# approved\n")
	badMan := w("bad-approved.txt", "not a slug!\n")
	unknownMan := w("unknown-approved.txt", "no-such-glyph\n")
	reg := w("ALPHABET.md", "| slug | canonical definition file | certified | prior promotion | judge | battery | verdict | definition digest (sha256) |\n"+
		"|---|---|---|---|---|---|---|---|\n"+
		"| casg-direct | `candidates/C.md` | 2026-09-13 | - | j | b | PROMOTE | `f9196e663c86546f05fb149b751ac0e2e8df8fe3130cb46123ed4523c187c6f7` |\n")
	badReg := w("BAD.md", "no table here\n")
	canon := filepath.Join(root, "canon")
	cases := [][]string{
		{},
		{"verify", "-public-dir"},
		{"verify", "-manifest"},
		{"promote", "-canon"},
		{"promote", "-registry"},
		{"verify", "extra"},
		{"bogus"},
		{"verify", "-public-dir", emptyPub, "-manifest", okMan},
		{"verify", "-public-dir", leakPub, "-manifest", okMan},
		{"promote", "-registry", filepath.Join(root, "absent.md"), "-manifest", okMan, "-public-dir", emptyPub, "-canon", canon},
		{"promote", "-registry", badReg, "-manifest", okMan, "-public-dir", emptyPub, "-canon", canon},
		{"promote", "-registry", reg, "-manifest", filepath.Join(root, "absent.txt"), "-public-dir", emptyPub, "-canon", canon},
		{"promote", "-registry", reg, "-manifest", badMan, "-public-dir", emptyPub, "-canon", canon},
		{"promote", "-registry", reg, "-manifest", unknownMan, "-public-dir", emptyPub, "-canon", canon},
		{"promote", "-registry", reg, "-manifest", okMan, "-public-dir", stalePub, "-canon", canon},
	}
	var b strings.Builder
	for _, c := range cases {
		stderr, code := captureStderr(t, func() int { return runDisclosure(c) })
		fmt.Fprintf(&b, "$ %s\n%s=> %d\n", strings.Join(c, " "), stderr, code)
	}
	got := strings.ReplaceAll(b.String(), root, "<root>")
	golden := filepath.Join("testdata", "disclosure.golden")
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
		t.Fatalf("disclosure drifted from %s:\n%s", golden, got)
	}
}

// TestBatteryArgErrors pins runBattery's argument errors, all raised before any
// model call: each valued flag missing its value, and the wrong number of
// positional args.
func TestBatteryArgErrors(t *testing.T) {
	cases := [][]string{{}, {"a.md", "b.md"}, {"-all-items"}}
	for _, f := range []string{"-out", "-base", "-model", "-version", "-repo", "-registry", "-definition", "-provenance-types", "-baseline"} {
		cases = append(cases, []string{"c.md", f})
	}
	var b strings.Builder
	for _, c := range cases {
		stderr, code := captureStderr(t, func() int { return runBattery(c) })
		fmt.Fprintf(&b, "$ %s\n%s=> %d\n", strings.Join(c, " "), stderr, code)
	}
	got := b.String()
	golden := filepath.Join("testdata", "battery_args.golden")
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
		t.Fatalf("battery args drifted from %s:\n%s", golden, got)
	}
}

// TestParseBatteryArgs pins the defaults and every flag, in any position.
func TestParseBatteryArgs(t *testing.T) {
	o, code := parseBatteryArgs([]string{"c.md"})
	want := batteryOpts{base: "http://localhost:8081/v1", modelName: "qwen3.6-27b", repoDir: ".",
		registryPath: "corpus/private/glyph-model/ALPHABET.md", definitionPath: "corpus/glyph-model/GLYPH_DEFINITION.md",
		provenanceTypesPath: "corpus/glyph-model/GLYPH_PROVENANCE_TYPES.md", candidatePath: "c.md"}
	if code != 0 || o != want {
		t.Fatalf("defaults = %+v (%d)", o, code)
	}
	o, code = parseBatteryArgs([]string{"-all-items", "-out", "o", "-base", "b", "c.md", "-model", "m", "-version", "v",
		"-repo", "r", "-registry", "g", "-definition", "d", "-provenance-types", "p", "-baseline", "bl"})
	want = batteryOpts{outPath: "o", base: "b", modelName: "m", version: "v", repoDir: "r", registryPath: "g",
		definitionPath: "d", provenanceTypesPath: "p", baselinePath: "bl", allItems: true, candidatePath: "c.md"}
	if code != 0 || o != want {
		t.Fatalf("all flags = %+v (%d)", o, code)
	}
}

// TestBaselineCaptureArgErrors pins baseline capture's argument errors and the
// scenario read failure, all raised before any model call.
func TestBaselineCaptureArgErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.md")
	cases := [][]string{{}, {"item"}, {"-scenario", "s.md"}, {"a", "b", "-scenario", "s.md"},
		{"item", "-scenario", "s.md", "-runs", "0"}, {"item", "-scenario", "s.md", "-runs", "x"},
		{"item", "-scenario", "s.md", "-temperature", "hot"}, {"item", "-scenario", "s.md", "-max-tokens", "-3"},
		{"item", "-scenario", missing, "-out", "o", "-base", "b", "-model", "m", "-version", "v",
			"-runs", "2", "-temperature", "0.5", "-max-tokens", "64"}}
	for _, f := range []string{"-scenario", "-out", "-base", "-model", "-version", "-runs", "-temperature", "-max-tokens"} {
		cases = append(cases, []string{"item", f})
	}
	var b strings.Builder
	for _, c := range cases {
		stderr, code := captureStderr(t, func() int { return runBaselineCapture(c) })
		line := strings.Join(c, " ")
		fmt.Fprintf(&b, "$ %s\n%s=> %d\n", strings.ReplaceAll(line, missing, "<absent>"), strings.ReplaceAll(stderr, missing, "<absent>"), code)
	}
	got := b.String()
	golden := filepath.Join("testdata", "baseline_capture_args.golden")
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
		t.Fatalf("baseline capture args drifted from %s:\n%s", golden, got)
	}
}

// TestParseCaptureArgs pins baseline capture's defaults and every flag.
func TestParseCaptureArgs(t *testing.T) {
	o, code := parseCaptureArgs([]string{"item", "-scenario", "s.md"})
	want := captureOpts{scenarioPath: "s.md", outDir: ".", base: "http://localhost:8081/v1", runs: 8, temperature: 0.8, maxTokens: 512, item: "item"}
	if code != 0 || o != want {
		t.Fatalf("defaults = %+v (%d)", o, code)
	}
	o, code = parseCaptureArgs([]string{"-version", "v", "item", "-scenario", "s", "-out", "o", "-base", "b", "-model", "m",
		"-runs", "3", "-temperature", "0", "-max-tokens", "9"})
	want = captureOpts{scenarioPath: "s", outDir: "o", base: "b", modelName: "m", version: "v", versionSet: true,
		runs: 3, temperature: 0, maxTokens: 9, item: "item"}
	if code != 0 || o != want {
		t.Fatalf("all flags = %+v (%d)", o, code)
	}
}
