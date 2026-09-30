package disclosure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/battery"
)

// candidateDoc builds a candidate working-doc whose AC-4 block carries the given
// glyph slug. The "## Notes" section after the block is derivation material that
// GlyphBlock must exclude from what gets published.
func candidateDoc(slug string) string {
	return "# Glyph Candidate: " + slug + "\n\n" +
		"## AC-3 — Decomp source\n\nprivate derivation material.\n\n" +
		"## AC-4 — Assembled glyph\n\n" +
		"**Glyph:** `" + slug + "`\n" +
		"**Y-fire:** the condition fires.\n" +
		"**Y-not-fire:** the discriminating condition is present.\n\n" +
		"### Marker axis\n\n**Firing condition:** trace shows the miss.\n\n" +
		"## Notes\n\nprivate study-trajectory notes that must not be published.\n"
}

// writeCanon creates a private-canon glyph-model dir with the given candidates
// and returns the dir plus a registry whose digests are the real digests of
// those candidates.
func writeCanon(t *testing.T, slugs ...string) (string, []CertifiedGlyph) {
	t.Helper()
	canon := t.TempDir()
	if err := os.MkdirAll(filepath.Join(canon, "candidates"), 0o755); err != nil {
		t.Fatal(err)
	}
	var reg []CertifiedGlyph
	for _, slug := range slugs {
		doc := candidateDoc(slug)
		rel := filepath.Join("candidates", "CANDIDATE_"+slug+".md")
		if err := os.WriteFile(filepath.Join(canon, rel), []byte(doc), 0o644); err != nil {
			t.Fatal(err)
		}
		dg, err := battery.GlyphDigest(doc)
		if err != nil {
			t.Fatal(err)
		}
		reg = append(reg, CertifiedGlyph{Slug: slug, CandidateFile: rel, Digest: dg})
	}
	return canon, reg
}

func TestPromoteWritesApprovedSubsetOnly(t *testing.T) {
	canon, reg := writeCanon(t, "a", "b", "c", "d", "e")
	public := t.TempDir()

	res, err := Promote(canon, public, reg, []string{"a", "c", "d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(res.Published, ","); got != "a,c,d" {
		t.Errorf("Published = %q, want a,c,d", got)
	}

	present, err := PublishedSlugs(public)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(present, ","); got != "a,c,d" {
		t.Errorf("published dir has %q, want a,c,d", got)
	}

	body, err := os.ReadFile(filepath.Join(public, "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "**Glyph:** `a`") {
		t.Errorf("a.md missing the glyph line: %q", body)
	}
	if strings.Contains(string(body), "study-trajectory notes") {
		t.Error("a.md leaked the private Notes section — GlyphBlock should stop before ## Notes")
	}
	if strings.Contains(string(body), "private derivation material") {
		t.Error("a.md leaked the AC-3 derivation")
	}
}

func TestPromoteRemovesStalePublished(t *testing.T) {
	canon, reg := writeCanon(t, "a", "b")
	public := t.TempDir()
	// A previously published glyph that is no longer approved.
	if err := os.WriteFile(filepath.Join(public, "old.md"), []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Promote(canon, public, reg, []string{"a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(res.Removed, ","); got != "old" {
		t.Errorf("Removed = %q, want old", got)
	}
	present, _ := PublishedSlugs(public)
	if got := strings.Join(present, ","); got != "a" {
		t.Errorf("published dir = %q, want a", got)
	}
}

func TestPromoteRejectsUncertifiedApproval(t *testing.T) {
	canon, reg := writeCanon(t, "a")
	public := t.TempDir()

	_, err := Promote(canon, public, reg, []string{"a", "ghost"})
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("error = %v, want rejection naming the uncertified slug", err)
	}
}

func TestPromoteRejectsDriftedDigest(t *testing.T) {
	canon, reg := writeCanon(t, "a")
	public := t.TempDir()
	reg[0].Digest = strings.Repeat("0", 64) // pretend the frozen digest differs

	_, err := Promote(canon, public, reg, []string{"a"})
	if err == nil || !strings.Contains(err.Error(), "drifted") {
		t.Fatalf("error = %v, want a drifted-definition error", err)
	}
}

func TestPromoteMissingCandidateFileErrors(t *testing.T) {
	canon, reg := writeCanon(t, "a")
	public := t.TempDir()
	if err := os.Remove(filepath.Join(canon, reg[0].CandidateFile)); err != nil {
		t.Fatal(err)
	}
	_, err := Promote(canon, public, reg, []string{"a"})
	if err == nil || !strings.Contains(err.Error(), "read candidate") {
		t.Fatalf("error = %v, want a read-candidate error", err)
	}
}

func TestPublishedSlugsMissingDirIsEmpty(t *testing.T) {
	slugs, err := PublishedSlugs(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slugs) != 0 {
		t.Errorf("slugs = %v, want empty", slugs)
	}
}

func TestVerifyPublishedMatchesManifest(t *testing.T) {
	canon, reg := writeCanon(t, "a", "b")
	public := t.TempDir()
	if _, err := Promote(canon, public, reg, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(t.TempDir(), "PUBLIC_ALPHABET.txt")
	if err := os.WriteFile(manifest, []byte("# approved\na\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPublished(public, manifest); err != nil {
		t.Errorf("VerifyPublished() = %v, want nil", err)
	}
}

func TestVerifyPublishedDetectsLeak(t *testing.T) {
	public := t.TempDir()
	if err := os.WriteFile(filepath.Join(public, "leaked.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(t.TempDir(), "PUBLIC_ALPHABET.txt")
	if err := os.WriteFile(manifest, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	err := VerifyPublished(public, manifest)
	if err == nil || !strings.Contains(err.Error(), "LEAK") {
		t.Fatalf("error = %v, want a LEAK error", err)
	}
}

func TestVerifyPublishedMissingManifestErrors(t *testing.T) {
	err := VerifyPublished(t.TempDir(), filepath.Join(t.TempDir(), "nope.txt"))
	if err == nil || !strings.Contains(err.Error(), "approved manifest") {
		t.Fatalf("error = %v, want a read-manifest error", err)
	}
}

func TestVerifyPublishedBadManifestErrors(t *testing.T) {
	manifest := filepath.Join(t.TempDir(), "m.txt")
	if err := os.WriteFile(manifest, []byte("two tokens here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := VerifyPublished(t.TempDir(), manifest)
	if err == nil || !strings.Contains(err.Error(), "not a bare slug") {
		t.Fatalf("error = %v, want a manifest parse error", err)
	}
}

func TestPromoteMalformedCandidateErrors(t *testing.T) {
	canon := t.TempDir()
	if err := os.MkdirAll(filepath.Join(canon, "candidates"), 0o755); err != nil {
		t.Fatal(err)
	}
	rel := filepath.Join("candidates", "CANDIDATE_x.md")
	if err := os.WriteFile(filepath.Join(canon, rel), []byte("# no glyph line here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg := []CertifiedGlyph{{Slug: "x", CandidateFile: rel, Digest: strings.Repeat("a", 64)}}
	_, err := Promote(canon, t.TempDir(), reg, []string{"x"})
	if err == nil || !strings.Contains(err.Error(), "extract glyph") {
		t.Fatalf("error = %v, want an extract-glyph error", err)
	}
}

func TestPromoteMkdirErrors(t *testing.T) {
	canon, reg := writeCanon(t, "a")
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// publicDir's parent is a file, so MkdirAll cannot create it.
	_, err := Promote(canon, filepath.Join(blocker, "sub"), reg, []string{"a"})
	if err == nil || !strings.Contains(err.Error(), "create public dir") {
		t.Fatalf("error = %v, want a create-public-dir error", err)
	}
}

func TestPublishedSlugsOnFileErrors(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := PublishedSlugs(f)
	if err == nil || !strings.Contains(err.Error(), "read public dir") {
		t.Fatalf("error = %v, want a read-public-dir error", err)
	}
}

func TestPublishedSlugsSkipsDirsAndNonMarkdown(t *testing.T) {
	d := t.TempDir()
	if err := os.MkdirAll(filepath.Join(d, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "a.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	slugs, err := PublishedSlugs(d)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(slugs, ","); got != "a" {
		t.Errorf("PublishedSlugs() = %q, want a (dirs and non-markdown skipped)", got)
	}
}

func TestVerifyPublishedUnreadableTreeErrors(t *testing.T) {
	// publicDir is a file, so listing it fails after the manifest parses.
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(t.TempDir(), "m.txt")
	if err := os.WriteFile(manifest, []byte("# empty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := VerifyPublished(f, manifest)
	if err == nil || !strings.Contains(err.Error(), "read public dir") {
		t.Fatalf("error = %v, want a read-public-dir error", err)
	}
}
