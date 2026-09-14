package battery

import (
	"fmt"
	"strings"
)

// headingPrefix ends the AC-4 block: extraction stops at the next Markdown
// section heading (e.g. a trailing `## Notes`) or at end of file. The block
// anchor glyphMarker and the fallout anchor falloutProfileMarker are defined
// alongside the items that also read them (static_steps.go), so extraction and
// assessment use the same markers.
const headingPrefix = "## "

// ExtractEntry pulls the assessable clean entry out of a candidate working-doc.
// A candidate file interleaves the metadata header, the AC-1/AC-2 derivation,
// the deliberately project-scoped AC-3 specimen, and the universalized AC-4
// assembled glyph. The battery must assess only the assembled glyph plus its
// fallout-profile referent; feeding it the whole doc makes the universality and
// intent items trip on non-structural text (metadata paths, the AC-3 specimen)
// and manufacture false failures.
//
// ExtractEntry returns exactly two things, joined:
//   - the `**Fallout profile:**` header line (Item 15's referent), and
//   - the AC-4 block from the first `**Glyph:**` line (skipping the
//     `[Executor: …]` template preamble above it) to the next `## ` heading or
//     end of file.
//
// It is the Go form of the retired extract_entries.py, so the battery reads the
// canonical candidate definition directly with no separate extract copy. The
// AC-4 block it returns is byte-identical to the block glyph_digest.py hashes,
// so the assessed text and the frozen digest describe the same definition.
//
// A candidate with no `**Glyph:**` line is malformed and returns an error. A
// missing fallout line is not an error here: the entry then carries no fallout
// field and Item 15 fails it closed, which is the correct verdict.
func ExtractEntry(content string) (string, error) {
	lines := strings.Split(content, "\n")

	glyphIdx := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, glyphMarker) {
			glyphIdx = i
			break
		}
	}
	if glyphIdx == -1 {
		return "", fmt.Errorf("battery: no %s line found; not a candidate working-doc", glyphMarker)
	}

	end := len(lines)
	for i := glyphIdx + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], headingPrefix) {
			end = i
			break
		}
	}
	block := strings.TrimRight(strings.Join(lines[glyphIdx:end], "\n"), " \t\r\n")

	fallout := ""
	for _, ln := range lines {
		if strings.HasPrefix(ln, falloutProfileMarker) {
			fallout = strings.TrimRight(ln, " \t\r\n")
			break
		}
	}
	if fallout == "" {
		return block, nil
	}
	return fallout + "\n\n" + block, nil
}
