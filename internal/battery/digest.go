package battery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GlyphBlock returns the AC-4 assembled-glyph block on its own: the lines from
// the first **Glyph:** line (any `[Executor: …]` preamble above it is naturally
// excluded) to the next "## " heading or end of file, right-trimmed of trailing
// whitespace. It is the exact text glyph_digest.py hashes, so GlyphDigest
// reproduces the frozen certification digest.
//
// ExtractEntry prepends the fallout line to this block for battery assessment;
// the digest is over the block alone, which is why the two are separate. A
// candidate with no **Glyph:** line is malformed and returns an error.
func GlyphBlock(content string) (string, error) {
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
	return strings.TrimRight(strings.Join(lines[glyphIdx:end], "\n"), " \t\r\n"), nil
}

// GlyphDigest returns the sha256 hex of a candidate's GlyphBlock — the canonical
// certification digest. It is the Go form of glyph_digest.py: the one place the
// digest is computed on the Go side, so the frozen registry value, the freeze
// gate, and the assessed entry all describe one definition. Deterministic, no
// network.
func GlyphDigest(content string) (string, error) {
	block, err := GlyphBlock(content)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(block))
	return hex.EncodeToString(sum[:]), nil
}
