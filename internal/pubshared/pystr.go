// Package pubshared ports the standalone reference helpers under
// provenance/published-scoring/_shared/ to Go (chain
// eradicate-python-rewrite-in-go): scramble.py, build_app.py, and anchor_eval.py.
// The logic is sans-IO; cmd/corpos-lab/pubscore_pubshared.go does the file wiring.
package pubshared

import (
	"strings"
	"unicode"
)

// pyStrip reproduces python str.strip() with no argument.
func pyStrip(s string) string {
	return strings.TrimFunc(s, unicode.IsSpace)
}

// pyLStrip reproduces python str.lstrip() with no argument.
func pyLStrip(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}
