package disclosure

import (
	"strings"
	"testing"
)

const digestA = "f9196e663c86546f05fb149b751ac0e2e8df8fe3130cb46123ed4523c187c6f7"
const digestB = "1cae0a6e5284f73440005078270a16162906d758fc2d0bd3dea7a6315a59dcfc"

// validRegistry mirrors the real ALPHABET.md table shape: prose, then the table.
func validRegistry() string {
	return "# ALPHABET — glyph certification registry\n\nSome prose about the registry.\n\n" +
		"| slug | canonical definition file | certified | prior promotion | judge | battery | verdict | definition digest (sha256) |\n" +
		"|------|---------------------------|-----------|-----------------|-------|---------|---------|----------------------------|\n" +
		"| casg-direct | `candidates/CANDIDATE_casg-direct_2026-04-03.md` | 2026-09-13 | 2026-09-06 | Opus 4.8 | full 15-item | PROMOTE | `" + digestA + "` |\n" +
		"| casg-delegate | `candidates/CANDIDATE_casg-delegate_2026-04-03.md` | 2026-09-13 | 2026-09-06 | Opus 4.8 | full 15-item | PROMOTE | `" + digestB + "` |\n" +
		"\n## Certification provenance\n\nMore prose after the table.\n"
}

func TestParseRegistryValid(t *testing.T) {
	glyphs, err := ParseRegistry(validRegistry())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(glyphs) != 2 {
		t.Fatalf("got %d rows, want 2", len(glyphs))
	}
	if glyphs[0].Slug != "casg-direct" {
		t.Errorf("row 0 slug = %q, want casg-direct", glyphs[0].Slug)
	}
	if glyphs[0].CandidateFile != "candidates/CANDIDATE_casg-direct_2026-04-03.md" {
		t.Errorf("row 0 file = %q", glyphs[0].CandidateFile)
	}
	if glyphs[0].Digest != digestA {
		t.Errorf("row 0 digest = %q, want %q", glyphs[0].Digest, digestA)
	}
	if got, want := strings.Join(Slugs(glyphs), ","), "casg-direct,casg-delegate"; got != want {
		t.Errorf("Slugs() = %q, want %q", got, want)
	}
}

func TestParseRegistryNoTable(t *testing.T) {
	_, err := ParseRegistry("# just prose\n\nno table here.\n")
	if err == nil {
		t.Fatal("ParseRegistry() = nil error, want 'no table found'")
	}
	if !strings.Contains(err.Error(), "no certification-registry table") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestParseRegistryMalformedDigest(t *testing.T) {
	md := "| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n" +
		"| casg-direct | `candidates/x.md` | `not-a-digest` |\n"
	_, err := ParseRegistry(md)
	if err == nil {
		t.Fatal("ParseRegistry() = nil error, want malformed-digest error")
	}
	if !strings.Contains(err.Error(), "casg-direct") || !strings.Contains(err.Error(), "digest") {
		t.Errorf("error = %q, want it to name the slug and the digest problem", err.Error())
	}
}

func TestParseRegistryEmptySlug(t *testing.T) {
	md := "| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n" +
		"|  | `candidates/x.md` | `" + digestA + "` |\n"
	_, err := ParseRegistry(md)
	if err == nil || !strings.Contains(err.Error(), "empty slug") {
		t.Fatalf("error = %v, want empty-slug error", err)
	}
}

func TestParseRegistryMissingCandidateFile(t *testing.T) {
	md := "| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n" +
		"| casg-direct |  | `" + digestA + "` |\n"
	_, err := ParseRegistry(md)
	if err == nil || !strings.Contains(err.Error(), "candidate definition file") {
		t.Fatalf("error = %v, want missing-file error", err)
	}
}

func TestParseRegistryTooFewColumns(t *testing.T) {
	md := "| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n" +
		"| casg-direct | `candidates/x.md` |\n" // only 2 cells, digest column absent
	_, err := ParseRegistry(md)
	if err == nil || !strings.Contains(err.Error(), "columns") {
		t.Fatalf("error = %v, want too-few-columns error", err)
	}
}

func TestParseRegistrySkipsTablesMissingColumns(t *testing.T) {
	// A decoy table with slug but no canonical-file column, then a decoy with
	// slug and canonical file but no digest column, then the real registry.
	md := "| slug | note |\n|---|---|\n| a | b |\n\n" +
		"| slug | canonical definition file | certified |\n|---|---|---|\n| foo | `f.md` | 2026-01-01 |\n\n" +
		"prose between tables\n\n" +
		"| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n" +
		"| casg-direct | `candidates/x.md` | `" + digestA + "` |\n"

	glyphs, err := ParseRegistry(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(glyphs) != 1 || glyphs[0].Slug != "casg-direct" {
		t.Errorf("got %+v, want just casg-direct from the real table", glyphs)
	}
}

func TestParseRegistryNoDataRows(t *testing.T) {
	md := "| slug | canonical definition file | definition digest (sha256) |\n" +
		"|---|---|---|\n"
	_, err := ParseRegistry(md)
	if err == nil || !strings.Contains(err.Error(), "no data rows") {
		t.Fatalf("error = %v, want no-data-rows error", err)
	}
}
