package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"corpos-lab/internal/alphabetassay"
)

// TestAlphabetCollectAssayGolden freezes every file alphabet-collect-assay writes
// for a small study: two classes with runs (one class has a run dir whose name
// does not parse and a response whose name has no seed, both skipped), and the
// other eight classes empty.
func TestAlphabetCollectAssayGolden(t *testing.T) {
	study := t.TempDir()
	c0, c1 := alphabetassay.Classes[0], alphabetassay.Classes[1]
	resp := func(cls, run, file, body string) {
		writeFileT(t, filepath.Join(study, cls, "runs", run, "out", "responses", file), body)
	}
	resp(c0, "asy-"+c0+"-s1-qwen38", "baseline_1.txt", "q base 1")
	resp(c0, "asy-"+c0+"-s1-qwen38", "alphabet_2.txt", "q alpha 2")
	resp(c0, "asy-"+c0+"-s1-qwen38", "noseed.txt", "skipped")
	resp(c0, "asy-"+c0+"-s2-mistral", "baseline_1.txt", "m base 1")
	resp(c0, "asy-"+c0+"-sX-bad", "baseline_1.txt", "unparsed dir")
	resp(c1, "asy-"+c1+"-s1-phi4", "alphabet_3.txt", "p alpha 3")
	out := filepath.Join(t.TempDir(), "collected")
	if err := paCollectAssay([]string{study, out}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		raw, _ := os.ReadFile(filepath.Join(out, n))
		fmt.Fprintf(&b, "== %s\n%s\n", n, raw)
	}
	got := b.String()
	golden := filepath.Join("testdata", "alphabet_collect.golden")
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
		t.Fatalf("collect output drifted from %s:\n%s", golden, got)
	}
}
