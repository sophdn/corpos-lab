package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// pubScoreModes maps a pub-score mode ("<study>-<mode>") to its handler. Each
// ported provenance/published-scoring tool registers its mode from an init() in
// its own file, so a new tool is a purely additive file — no shared-file edit.
var pubScoreModes = map[string]func([]string) error{}

func registerPubScoreMode(name string, fn func([]string) error) {
	if _, dup := pubScoreModes[name]; dup {
		panic("pub-score: duplicate mode " + name)
	}
	pubScoreModes[name] = fn
}

func pubScoreModeList() string {
	names := make([]string, 0, len(pubScoreModes))
	for n := range pubScoreModes {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// runPubScore dispatches the Go rewrites of the per-paper provenance scoring and
// analysis python (chain eradicate-python-rewrite-in-go). Modes self-register.
func runPubScore(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: corpos-lab pub-score <mode> … (modes: %s)\n", pubScoreModeList())
		return 2
	}
	mode, rest := args[0], args[1:]
	fn, ok := pubScoreModes[mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "pub-score: unknown mode %q (modes: %s)\n", mode, pubScoreModeList())
		return 2
	}
	if err := fn(rest); err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: pub-score %s: %v\n", mode, err)
		return 1
	}
	return 0
}
