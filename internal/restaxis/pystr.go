package restaxis

// Small helpers that reproduce python str whitespace-trimming semantics, shared
// by the ported doc-generation tools so their string transforms stay
// byte-identical to the python originals.

import (
	"strings"
	"unicode"
)

// pyStrip reproduces python str.strip() with no argument: it removes leading and
// trailing Unicode whitespace.
func pyStrip(s string) string {
	return strings.TrimFunc(s, unicode.IsSpace)
}

// StripText is the exported form of pyStrip, for the cmd wiring to strip file
// contents the same way the python tools do (text.read_text().strip()).
func StripText(s string) string {
	return pyStrip(s)
}

// pyRStrip reproduces python str.rstrip() with no argument: it removes trailing
// Unicode whitespace only.
func pyRStrip(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
