package disclosure

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/digest"
)

// PromoteResult reports what a promotion did.
type PromoteResult struct {
	Published []string // slugs written to the public tree
	Removed   []string // stale slugs deleted from the public tree
}

// Promote makes the public glyph directory hold exactly the approved subset,
// each as its canonical AC-4 block, and nothing else.
//
// For every approved-and-certified glyph it reads the candidate working-doc from
// the private canon, extracts the AC-4 block (battery.GlyphBlock — the same text
// the certification digest is over, which stops before the derivation notes),
// verifies that block's sha256 matches the registry digest, and writes it to
// publicDir/<slug>.md. It removes any <slug>.md already in publicDir whose slug
// is not approved, so the public directory always equals the approved set.
//
// It refuses to publish when the manifest approves an uncertified glyph, or when
// a candidate's current definition no longer matches its frozen registry digest
// (a drifted glyph must be re-certified, not published).
//
// canonDir is the private canon glyph-model directory (holds candidates/);
// publicDir is the public published-glyph directory.
func Promote(canonDir, publicDir string, registry []CertifiedGlyph, approved []string) (PromoteResult, error) {
	sel := Select(Slugs(registry), approved)
	if err := sel.Err(); err != nil {
		return PromoteResult{}, err
	}

	bySlug := make(map[string]CertifiedGlyph, len(registry))
	for _, g := range registry {
		bySlug[g.Slug] = g
	}

	if err := os.MkdirAll(publicDir, 0o755); err != nil {
		return PromoteResult{}, fmt.Errorf("disclosure: create public dir: %w", err)
	}

	var res PromoteResult
	for _, slug := range sel.Publish {
		g := bySlug[slug]
		content, err := os.ReadFile(filepath.Join(canonDir, g.CandidateFile))
		if err != nil {
			return res, fmt.Errorf("disclosure: read candidate for %q: %w", slug, err)
		}
		block, err := battery.GlyphBlock(string(content))
		if err != nil {
			return res, fmt.Errorf("disclosure: extract glyph %q: %w", slug, err)
		}
		// The certification digest is sha256 of exactly this block, so hashing the
		// block reproduces the frozen registry value.
		if got := digest.Bytes([]byte(block)); got != g.Digest {
			return res, fmt.Errorf("disclosure: glyph %q definition digest %s does not match registry digest %s — "+
				"the definition drifted; re-certify before publishing", slug, got, g.Digest)
		}
		if err := os.WriteFile(filepath.Join(publicDir, slug+".md"), []byte(block+"\n"), 0o644); err != nil {
			return res, fmt.Errorf("disclosure: write public glyph %q: %w", slug, err)
		}
		res.Published = append(res.Published, slug)
	}

	present, err := PublishedSlugs(publicDir)
	if err != nil {
		return res, err
	}
	keep := toSet(sel.Publish)
	for _, slug := range present {
		if _, ok := keep[slug]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(publicDir, slug+".md")); err != nil {
			return res, fmt.Errorf("disclosure: remove stale public glyph %q: %w", slug, err)
		}
		res.Removed = append(res.Removed, slug)
	}
	sort.Strings(res.Removed)
	return res, nil
}

// PublishedSlugs lists the glyph slugs present in a published-glyph directory:
// the basename without .md of every *.md file. A missing directory is not an
// error — it means nothing has been published yet.
func PublishedSlugs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("disclosure: read public dir: %w", err)
	}
	var slugs []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slugs = append(slugs, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(slugs)
	return slugs, nil
}

// VerifyPublished checks that a published-glyph directory contains exactly the
// glyphs named in the approved-manifest file. It is the public-repo CI check and
// needs no access to the private canon.
func VerifyPublished(publicDir, approvedManifestPath string) error {
	data, err := os.ReadFile(approvedManifestPath)
	if err != nil {
		return fmt.Errorf("disclosure: read approved manifest: %w", err)
	}
	approved, err := ParseApproved(string(data))
	if err != nil {
		return err
	}
	present, err := PublishedSlugs(publicDir)
	if err != nil {
		return err
	}
	return ValidatePublicTree(present, approved)
}
