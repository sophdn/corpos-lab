package disclosure

import (
	"fmt"
	"regexp"
	"strings"
)

// CertifiedGlyph is one row of the private ALPHABET.md certification registry: a
// certified glyph's slug, the candidate working-doc that holds its canonical AC-4
// definition, and the sha256 digest of that definition block.
type CertifiedGlyph struct {
	Slug          string
	CandidateFile string
	Digest        string
}

var digestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

var separatorCellRE = regexp.MustCompile(`^:?-+:?$`)

// ParseRegistry parses the certification-registry markdown table in ALPHABET.md
// into its rows and validates the table shape. The table must carry the slug,
// canonical-definition-file, and definition-digest columns, and every data row
// must have a non-empty slug, a non-empty candidate-file path, and a
// 64-character lowercase-hex sha256 digest.
//
// ParseRegistry is the "expected shape of the alphabet" check: a malformed row
// is an error naming the offending slug, not a silently skipped line.
func ParseRegistry(md string) ([]CertifiedGlyph, error) {
	lines := strings.Split(md, "\n")

	headerIdx, cols := findHeader(lines)
	if headerIdx < 0 {
		return nil, fmt.Errorf("disclosure: no certification-registry table found " +
			"(need slug, canonical definition file, and definition digest columns)")
	}

	slugCol := cols["slug"]
	fileCol := cols["canonical definition file"]
	digestCol := digestColumn(cols)
	need := max(slugCol, max(fileCol, digestCol))

	var out []CertifiedGlyph
	for _, line := range lines[headerIdx+1:] {
		if strings.TrimSpace(line) == "" || !strings.Contains(line, "|") {
			break // the table ends at the first blank or non-table line
		}
		cells := splitTableRow(line)
		if isSeparatorRow(cells) {
			continue
		}
		if len(cells) <= need {
			return nil, fmt.Errorf("disclosure: registry row has %d columns, need at least %d: %q",
				len(cells), need+1, strings.TrimSpace(line))
		}
		slug := stripCell(cells[slugCol])
		file := stripCell(cells[fileCol])
		digest := stripCell(cells[digestCol])

		if slug == "" {
			return nil, fmt.Errorf("disclosure: registry row has an empty slug: %q", strings.TrimSpace(line))
		}
		if file == "" {
			return nil, fmt.Errorf("disclosure: registry row for %q has no candidate definition file", slug)
		}
		if !digestRE.MatchString(digest) {
			return nil, fmt.Errorf("disclosure: registry row for %q has a malformed digest %q (want 64 hex chars)", slug, digest)
		}
		out = append(out, CertifiedGlyph{Slug: slug, CandidateFile: file, Digest: digest})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("disclosure: certification-registry table has no data rows")
	}
	return out, nil
}

// Slugs returns just the slugs of a registry, in table order.
func Slugs(glyphs []CertifiedGlyph) []string {
	slugs := make([]string, len(glyphs))
	for i, g := range glyphs {
		slugs[i] = g.Slug
	}
	return slugs
}

// findHeader returns the index of the first markdown table header row that
// carries the slug, canonical-definition-file, and a definition-digest column,
// together with that row's column-name-to-index map. It returns (-1, nil) when
// no such header exists.
func findHeader(lines []string) (int, map[string]int) {
	for i, line := range lines {
		if !strings.Contains(line, "|") {
			continue
		}
		cols := columnIndex(splitTableRow(line))
		if _, ok := cols["slug"]; !ok {
			continue
		}
		if _, ok := cols["canonical definition file"]; !ok {
			continue
		}
		if digestColumn(cols) < 0 {
			continue
		}
		return i, cols
	}
	return -1, nil
}

// digestColumn returns the index of the column whose header names the definition
// digest, or -1 when absent.
func digestColumn(cols map[string]int) int {
	for name, idx := range cols {
		if strings.Contains(name, "definition digest") {
			return idx
		}
	}
	return -1
}

// splitTableRow splits a markdown table row into its trimmed cells, dropping the
// empty fields the outer pipes produce.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// columnIndex maps each lower-cased header cell to its column index.
func columnIndex(cells []string) map[string]int {
	m := make(map[string]int, len(cells))
	for i, c := range cells {
		m[strings.ToLower(c)] = i
	}
	return m
}

// isSeparatorRow reports whether every cell is a markdown table separator
// (dashes, optionally colon-anchored).
func isSeparatorRow(cells []string) bool {
	for _, c := range cells {
		if !separatorCellRE.MatchString(c) {
			return false
		}
	}
	return len(cells) > 0
}

// stripCell removes the backticks and surrounding whitespace a markdown cell
// wraps a path or digest in.
func stripCell(cell string) string {
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(cell), "`"))
}
