// Package digest computes the hex-encoded SHA-256 content digests used
// across corpos-lab for instrument freezing and container pinning: every
// executor-visible study artifact is pinned by digest at study
// registration and re-checked at scoring time (CHARTER.md
// freeze-by-digest rule), and assay container images are referenced by
// digest, never by tag.
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
