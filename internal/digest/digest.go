// Package digest computes the hex-encoded SHA-256 content digests used across
// corpos-lab to record what a run executed against: every executor-visible
// study artifact is digested into the run record, and assay container images
// are referenced by digest, never by tag.
//
// RECORDED, NOT ENFORCED. These digests once backed a freeze-by-digest rule
// that refused to run on a mismatch; that rule was retired 2026-07-14 (see
// INQUIRY.md). A digest that differs from a prior run's is information about
// the two runs, not grounds for refusing either. The rule also covered the
// wrong things: it pinned the stimulus — scenario, glyph, ground, rubric —
// which by design never changes, while the variables that actually moved
// between runs were the sampler and the processor, which it never touched.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// Bytes returns the hex-encoded SHA-256 digest of b.
func Bytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// File returns the hex-encoded SHA-256 digest of the file at path. Study
// artifacts are small text/JSON files, so the file is read whole.
func File(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("digest: read %s: %w", path, err)
	}
	return Bytes(b), nil
}
