package agentloop

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The dialect corpus in testdata/dialects is the parser's regression net. Each
// case is one model turn (<name>.turn) and the actions the harness must run from
// it (<name>.want): one actionLabel per line, then optional "--- content N"
// sections giving the exact body of the Nth action's edit. Every native tool-call
// form Qwen has been seen to emit has a case, so a parser change that drops a
// known form fails the gate instead of a costly grid.
//
// The turns are synthetic: each copies the markup of a form found in a real run
// transcript, with neutral paths and content. The transcripts themselves live in
// the private corpus and never reach this public repo. To add a form, copy its
// markup into a new case; `corpos-lab cells` lists the lost calls of a run, which
// is where new forms are found.

type dialectWant struct {
	labels   []string
	contents map[int]string // 1-based action index -> edit body
}

func parseDialectWant(t *testing.T, raw string) dialectWant {
	t.Helper()
	w := dialectWant{contents: map[int]string{}}
	sections := strings.Split(strings.TrimRight(raw, "\n"), "\n--- content ")
	for _, l := range strings.Split(sections[0], "\n") {
		if l = strings.TrimSpace(l); l != "" {
			w.labels = append(w.labels, l)
		}
	}
	for _, s := range sections[1:] {
		head, body, _ := strings.Cut(s, "\n")
		n, err := strconv.Atoi(strings.TrimSpace(head))
		if err != nil {
			t.Fatalf("bad content header %q", head)
		}
		w.contents[n] = body
	}
	return w
}

func TestDialectCorpus(t *testing.T) {
	turns, err := filepath.Glob(filepath.Join("testdata", "dialects", "*.turn"))
	if err != nil {
		t.Fatal(err)
	}
	if len(turns) < 20 {
		t.Fatalf("dialect corpus has %d cases, want the full set (>= 20)", len(turns))
	}
	for _, path := range turns {
		name := strings.TrimSuffix(filepath.Base(path), ".turn")
		t.Run(name, func(t *testing.T) {
			turn, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			rawWant, err := os.ReadFile(strings.TrimSuffix(path, ".turn") + ".want")
			if err != nil {
				t.Fatalf("case %s has no .want: %v", name, err)
			}
			want := parseDialectWant(t, string(rawWant))
			got := ParseActions(string(turn))
			if g := summarize(got); strings.Join(g, "\n") != strings.Join(want.labels, "\n") {
				t.Fatalf("actions = %q, want %q", g, want.labels)
			}
			for n, body := range want.contents {
				if got[n-1].Content != body {
					t.Errorf("action %d content = %q, want %q", n, got[n-1].Content, body)
				}
			}
		})
	}
}
