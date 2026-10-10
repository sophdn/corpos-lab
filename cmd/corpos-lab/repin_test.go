package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	repinOld = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	repinNew = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
)

func TestRunRepinRewritesNamedFiles(t *testing.T) {
	dir := t.TempDir()
	digests := filepath.Join(dir, "IMAGE_DIGESTS.txt")
	writeFileT(t, digests, "# header\nlab-agentic-loop-probe:dev "+repinNew+"\n")
	a := filepath.Join(dir, "a.toml")
	b := filepath.Join(dir, "b.toml")
	writeFileT(t, a, `image = "localhost/lab-agentic-loop-probe@`+repinOld+`"`+"\n")
	writeFileT(t, b, `image = "localhost/lab-agentic-loop-probe@`+repinNew+`"`+"\n")

	if code := run([]string{"repin", "-digests", digests, a, b}); code != 0 {
		t.Fatalf("repin exit = %d, want 0", code)
	}
	for _, p := range []string{a, b} {
		got, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(got), repinNew) || strings.Contains(string(got), repinOld) {
			t.Errorf("%s not re-pinned: %s", p, got)
		}
	}
}

func TestRunRepinFailures(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "IMAGE_DIGESTS.txt")
	writeFileT(t, good, "lab-agentic-loop-probe:dev "+repinNew+"\n")
	bad := filepath.Join(dir, "bad.txt")
	writeFileT(t, bad, "one two three\n")
	toml := filepath.Join(dir, "a.toml")
	writeFileT(t, toml, "x = 1\n")

	cases := map[string][]string{
		"no files":          {"repin", "-digests", good},
		"missing digests":   {"repin", "-digests", filepath.Join(dir, "absent.txt"), toml},
		"malformed digests": {"repin", "-digests", bad, toml},
		"missing file":      {"repin", "-digests", good, filepath.Join(dir, "absent.toml")},
		"bad flag":          {"repin", "-nope"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if code := run(args); code == 0 {
				t.Errorf("run(%v) = 0, want non-zero", args)
			}
		})
	}
}
