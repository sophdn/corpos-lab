package disclosure

import (
	"fmt"
	"strings"
)

// ParseApproved parses an approved-subset manifest: one glyph slug per line.
// Blank lines and lines whose first non-space character is "#" are ignored, and
// surrounding whitespace is trimmed. Slugs are returned in file order.
//
// A non-comment line that still contains inner whitespace after trimming is an
// error, so a stray sentence or a two-token line cannot be read as a slug and
// silently widen what is published.
func ParseApproved(text string) ([]string, error) {
	var slugs []string
	for i, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.ContainsAny(line, " \t") {
			return nil, fmt.Errorf("disclosure: approved manifest line %d is not a bare slug: %q", i+1, line)
		}
		slugs = append(slugs, line)
	}
	return slugs, nil
}
