package battery

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

const digestDoc = "**Metadata**\n" +
	"**Fallout profile:** ../fallout/x.md\n" +
	"\n" +
	"## AC-3 specimen\n" +
	"junk specimen text\n" +
	"\n" +
	"[Executor: fill below.]\n" +
	"\n" +
	"**Glyph:** `demo`\n" +
	"**Y-fire:** the terrain.\n" +
	"\n" +
	"### Marker axis\n" +
	"> Invariant: X.\n" +
	"\n" +
	"## Notes\n" +
	"trailing\n"

const wantBlock = "**Glyph:** `demo`\n" +
	"**Y-fire:** the terrain.\n" +
	"\n" +
	"### Marker axis\n" +
	"> Invariant: X."

func TestGlyphBlock_ExtractsAC4Block(t *testing.T) {
	got, err := GlyphBlock(digestDoc)
	if err != nil {
		t.Fatalf("GlyphBlock: %v", err)
	}
	if got != wantBlock {
		t.Fatalf("block mismatch:\n got %q\nwant %q", got, wantBlock)
	}
}

func TestGlyphBlock_EOFTerminator(t *testing.T) {
	// No trailing "## " heading: the block runs to EOF, right-trimmed.
	doc := "**Glyph:** `x`\nY line\n\n"
	got, err := GlyphBlock(doc)
	if err != nil {
		t.Fatalf("GlyphBlock: %v", err)
	}
	if want := "**Glyph:** `x`\nY line"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGlyphBlock_TripleHashDoesNotTerminate(t *testing.T) {
	// "### " is not "## " — an axis subheading stays inside the block.
	doc := "**Glyph:** `x`\n### Marker axis\nbody\n## Notes\nafter\n"
	got, err := GlyphBlock(doc)
	if err != nil {
		t.Fatalf("GlyphBlock: %v", err)
	}
	if want := "**Glyph:** `x`\n### Marker axis\nbody"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGlyphBlock_NoGlyphLine(t *testing.T) {
	if _, err := GlyphBlock("no marker here\n## heading\n"); err == nil {
		t.Fatal("expected error for content with no **Glyph:** line")
	}
}

func TestGlyphDigest_IsSHA256OfBlock(t *testing.T) {
	got, err := GlyphDigest(digestDoc)
	if err != nil {
		t.Fatalf("GlyphDigest: %v", err)
	}
	sum := sha256.Sum256([]byte(wantBlock))
	want := hex.EncodeToString(sum[:])
	if got != want {
		t.Fatalf("digest = %s, want %s (sha256 of the AC-4 block)", got, want)
	}
}

func TestGlyphDigest_NoGlyphLineErrors(t *testing.T) {
	if _, err := GlyphDigest("nothing here"); err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractEntry_PrependsFalloutToGlyphBlock(t *testing.T) {
	block, err := GlyphBlock(digestDoc)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := ExtractEntry(digestDoc)
	if err != nil {
		t.Fatal(err)
	}
	if want := "**Fallout profile:** ../fallout/x.md\n\n" + block; entry != want {
		t.Fatalf("entry mismatch:\n got %q\nwant %q", entry, want)
	}
}
