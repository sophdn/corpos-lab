//go:build parity

package battery

import (
	"os"
	"path/filepath"
	"testing"
)

// goldenDigests are glyph_digest.py's output for the certified candidates,
// captured 2026-09-15. TestDigestParity proves the Go GlyphDigest reproduces
// them byte-for-byte, so the Go tool and the retired Python tool are
// interchangeable. Excluded from the gate (parity build tag) because it reads
// the corpus tree; run it with:
//
//	go test -tags parity -run TestDigestParity ./internal/battery/
var goldenDigests = map[string]string{
	"casg-delegate":                      "1cae0a6e5284f73440005078270a16162906d758fc2d0bd3dea7a6315a59dcfc",
	"casg-direct":                        "f9196e663c86546f05fb149b751ac0e2e8df8fe3130cb46123ed4523c187c6f7",
	"conditional-gate-uniform-default":   "f6bb2f86c3feb301797aa475d35dace4ae4b4d364e71235fd2d3dd9d747e97ed",
	"discovery-event-non-recording":      "0d1990c72993824f200a6bd3672ecc03fe4d9718bd1a456556cb8f0bcc66f2ec",
	"formal-step-context-bypass":         "3fed53261c687b2113e221d647972207946d2c0610e21f03ae8d0f08c5e61c1e",
	"governed-operation-protocol-bypass": "908550cb75727b8ff71ff7f93050f662eab3b0d080bc1441318368fca8d968ec",
	"parent-state-check-bypass":          "1e3a90d441922303b0dccafb1e1a528ebc7e4b649d3c1e1ba292b2a83f27af77",
	"structural-ceiling-bypass":          "7cfa96687c96681fabbda4fe9e80769873670e9563b08099f09015a952eac8a0",
}

func TestDigestParity(t *testing.T) {
	dir := filepath.FromSlash("../../corpus/private/glyph-model/candidates")
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("candidate corpus not checked out (%v); run this where corpus/private is present", err)
	}
	for slug, want := range goldenDigests {
		matches, err := filepath.Glob(filepath.Join(dir, "CANDIDATE_"+slug+"_*.md"))
		if err != nil || len(matches) != 1 {
			t.Fatalf("%s: expected one candidate file, got %v (err %v)", slug, matches, err)
		}
		content, err := os.ReadFile(matches[0])
		if err != nil {
			t.Fatalf("%s: %v", slug, err)
		}
		got, err := GlyphDigest(string(content))
		if err != nil {
			t.Fatalf("%s: %v", slug, err)
		}
		if got != want {
			t.Errorf("%s: Go digest %s != Python golden %s", slug, got, want)
		}
	}
}
