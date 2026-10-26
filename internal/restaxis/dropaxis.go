package restaxis

import (
	"regexp"
	"strings"
)

// axisWords are the three axis words a section-start line may name. Ported from
// drop_axis.py AXES.
var axisWords = map[string]bool{"marker": true, "aim": true, "rest": true}

// sectionRE matches a section-start line: a Markdown heading (#..######) or a
// line-leading bold label, capturing the first word so we can tell which axis it
// names. Ported verbatim from drop_axis.py SECTION_RE.
var sectionRE = regexp.MustCompile(`^(?:#{1,6}\s+|\*\*)\s*([A-Za-z][A-Za-z-]*)`)

// headingRE matches any Markdown heading line; an axis block ends at the next
// heading. Ported from the inline r"^#{1,6}\s+" in drop_axis.drop_axis.
var headingRE = regexp.MustCompile(`^#{1,6}\s+`)

// sectionAxis returns the lowercased axis word if the line starts an axis
// section, or "" if it does not. Ported from drop_axis.py _section_axis (which
// returns None; "" is the Go equivalent, since a valid axis word is never empty).
func sectionAxis(line string) string {
	m := sectionRE.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	word := strings.ToLower(m[1])
	if axisWords[word] {
		return word
	}
	return ""
}

// DropAxis drops exactly one axis block (typically "Rest") from a glyph markdown
// text, structure-preserving. An axis block starts at a section-start line naming
// the axis — a heading `## Rest …` or a line-leading bold label `**Rest …**` —
// and runs until the next axis section start (Marker / Aim / Rest), the next
// heading, or end of file. The seam is then tidied: a doubled blank line is
// collapsed and a now-dangling trailing `---` separator (or trailing blanks) is
// dropped. It returns the new text and whether the axis was found; when the axis
// is absent the text is returned unchanged and found is false. The returned text
// always ends with a trailing newline. Ported from drop_axis.py drop_axis.
func DropAxis(text, axis string) (string, bool) {
	axis = strings.ToLower(axis)
	lines := strings.Split(text, "\n")

	start := -1
	for i, line := range lines {
		if sectionAxis(line) == axis {
			start = i
			break
		}
	}
	if start == -1 {
		return text, false
	}

	// End at the next axis section, the next heading, or EOF.
	end := len(lines)
	for j := start + 1; j < len(lines); j++ {
		if sectionAxis(lines[j]) != "" || headingRE.MatchString(lines[j]) {
			end = j
			break
		}
	}

	kept := make([]string, 0, start+(len(lines)-end))
	kept = append(kept, lines[:start]...)
	kept = append(kept, lines[end:]...)

	kept = tidySeam(kept, start)

	return strings.Join(kept, "\n") + "\n", true
}

// tidySeam cleans the join left by removing a block whose first kept-tail line is
// at index seam: it collapses a doubled blank line at the seam into a single
// blank, then drops any trailing `---` separator or blank lines. Ported from the
// two tidy loops in drop_axis.py drop_axis. Split out so the collapse branch —
// which drop_axis never reaches, since a block always ends at a non-blank heading
// or at EOF — stays directly testable.
func tidySeam(kept []string, seam int) []string {
	// Collapse a doubled blank line at the seam; keep a single blank as
	// paragraph separation and drop extras.
	for seam < len(kept) && strings.TrimSpace(kept[seam]) == "" {
		if seam > 0 && strings.TrimSpace(kept[seam-1]) == "" {
			kept = append(kept[:seam], kept[seam+1:]...)
		} else {
			seam++
		}
	}
	// A `---` (or blank) that is now the last content is a dangling separator
	// from the removed block.
	for len(kept) > 0 {
		last := strings.TrimSpace(kept[len(kept)-1])
		if last == "" || last == "---" {
			kept = kept[:len(kept)-1]
		} else {
			break
		}
	}
	return kept
}
