package wrongpath

import (
	"fmt"
	"strings"
	"unicode"
)

// The axis headers make_ablated.py keys on, matched against each line's stripped
// form.
const (
	markerHeader = "### Marker axis"
	restHeader   = "### Rest axis"
	aimHeader    = "### Aim axis"
)

// pyStrip reproduces python str.strip() with no argument (trim leading/trailing
// whitespace), used for header matching.
func pyStrip(s string) string {
	return strings.TrimFunc(s, unicode.IsSpace)
}

// splitLinesKeepends reproduces python str.splitlines(keepends=True): it splits
// on the full CPython universal-newline set and keeps each line's terminator, so
// "".join(result) reconstructs the input byte-for-byte.
func splitLinesKeepends(s string) []string {
	var lines []string
	runes := []rune(s)
	start := 0
	i := 0
	for i < len(runes) {
		r := runes[i]
		if isLineBoundary(r) {
			// \r\n counts as one boundary.
			if r == '\r' && i+1 < len(runes) && runes[i+1] == '\n' {
				i += 2
			} else {
				i++
			}
			lines = append(lines, string(runes[start:i]))
			start = i
			continue
		}
		i++
	}
	if start < len(runes) {
		lines = append(lines, string(runes[start:]))
	}
	return lines
}

func isLineBoundary(r rune) bool {
	switch r {
	case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

// findUnique returns the single index of a line whose stripped form equals
// header, erroring if the count is not exactly one. Ports make_ablated._find.
func findUnique(lines []string, header, name string) (int, error) {
	var hits []int
	for i, ln := range lines {
		if pyStrip(ln) == header {
			hits = append(hits, i)
		}
	}
	if len(hits) != 1 {
		return 0, fmt.Errorf("%s: expected exactly one %q, found %d", name, header, len(hits))
	}
	return hits[0], nil
}

// Ablate ports make_ablated.ablate + its post-checks: it removes the Marker axis
// and the Aim axis blocks (everything from the "### Marker axis" line up to but
// not including the "### Rest axis" line), keeping the header, the Y-Decision
// block, and the Rest axis. name is the glyph slug, used only in error messages.
func Ablate(text, name string) (string, error) {
	lines := splitLinesKeepends(text)
	markerIdx, err := findUnique(lines, markerHeader, name)
	if err != nil {
		return "", err
	}
	restIdx, err := findUnique(lines, restHeader, name)
	if err != nil {
		return "", err
	}
	if !(markerIdx < restIdx) {
		return "", fmt.Errorf("%s: Marker axis does not precede Rest axis", name)
	}
	out := strings.Join(lines[:markerIdx], "") + strings.Join(lines[restIdx:], "")
	// Post-checks mirroring make_ablated.main.
	if strings.Contains(out, markerHeader) || strings.Contains(out, aimHeader) {
		return "", fmt.Errorf("%s: ablation left a Marker or Aim axis behind", name)
	}
	if !strings.Contains(out, restHeader) {
		return "", fmt.Errorf("%s: ablation dropped the Rest axis", name)
	}
	return out, nil
}

// LineCount reproduces len(text.splitlines()) — the count without keepends,
// used for make_ablated's per-glyph stdout summary.
func LineCount(text string) int {
	return len(splitLinesKeepends(text))
}
