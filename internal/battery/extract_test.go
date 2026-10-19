package battery

import (
	"strings"
	"testing"
)

// A candidate working-doc: metadata header (with the fallout line), a
// project-scoped AC-3 specimen carrying intent language, the AC-4 assembled
// glyph, and a trailing Notes section. Extraction must keep only the fallout
// line and the AC-4 block, dropping the header body, AC-3, and Notes.
const candidateDoc = "# Candidate: test\n" +
	"\n" +
	"**Decision class:** structural\n" +
	"**Fallout profile:** ../fallout-profiles/test.md\n" +
	"**Source:** /home/foo/project/specimen.md\n" +
	"\n" +
	"## AC-3 — Decomp source material\n" +
	"\n" +
	"Specimen: when the agent decides to model the user's intent.\n" +
	"Project path: /home/foo/project/x.md\n" +
	"\n" +
	"## AC-4 — Assembled glyph\n" +
	"\n" +
	"[Executor: fill the template below.]\n" +
	"\n" +
	"**Glyph:** `test`\n" +
	"**Y — Decision terrain**\n" +
	"Some terrain description.\n" +
	"**Marker axis:** the failure direction.\n" +
	"**Aim axis:** the correct path.\n" +
	"**Rest axis:** irrelevant territory.\n" +
	"\n" +
	"## Notes\n" +
	"\n" +
	"When the agent wants more, that is out of scope.\n"

func TestExtractEntryKeepsFalloutAndAC4Block(t *testing.T) {
	got, err := ExtractEntry(candidateDoc)
	if err != nil {
		t.Fatalf("ExtractEntry: %v", err)
	}

	// The fallout line leads, so Item 15 finds its referent.
	if !strings.HasPrefix(got, "**Fallout profile:** ../fallout-profiles/test.md\n\n") {
		t.Fatalf("expected the fallout line first, got:\n%s", got)
	}
	// The AC-4 block is present, starting at the glyph line.
	if !strings.Contains(got, "**Glyph:** `test`") {
		t.Fatalf("expected the glyph line, got:\n%s", got)
	}
	if !strings.Contains(got, "**Rest axis:** irrelevant territory.") {
		t.Fatalf("expected the last axis line, got:\n%s", got)
	}

	// The [Executor: …] preamble above the glyph line is skipped.
	if strings.Contains(got, "[Executor:") {
		t.Errorf("the executor preamble must be dropped, got:\n%s", got)
	}
	// The AC-3 specimen, its intent language, and its project paths are dropped.
	if strings.Contains(got, "when the agent decides") {
		t.Errorf("AC-3 intent language must be dropped, got:\n%s", got)
	}
	if strings.Contains(got, "/home/foo/project") {
		t.Errorf("AC-3 project paths must be dropped, got:\n%s", got)
	}
	// The trailing Notes section is dropped.
	if strings.Contains(got, "out of scope") {
		t.Errorf("the trailing Notes section must be dropped, got:\n%s", got)
	}
	// No blank line dangles at the end of the block.
	if strings.HasSuffix(got, "\n") {
		t.Errorf("the block must be right-trimmed, got a trailing newline")
	}
}

func TestExtractEntryErrorsWithoutGlyphLine(t *testing.T) {
	_, err := ExtractEntry("# Not a candidate\n\nJust prose, no glyph marker.\n")
	if err == nil {
		t.Fatal("expected an error when no **Glyph:** line is present")
	}
}

func TestExtractEntryWithoutFalloutReturnsBlockOnly(t *testing.T) {
	doc := "## AC-4 — Assembled glyph\n\n**Glyph:** `x`\n**Marker axis:** m.\n"
	got, err := ExtractEntry(doc)
	if err != nil {
		t.Fatalf("ExtractEntry: %v", err)
	}
	if strings.Contains(got, "Fallout profile") {
		t.Errorf("no fallout line should appear, got:\n%s", got)
	}
	if !strings.HasPrefix(got, "**Glyph:** `x`") {
		t.Errorf("expected the block to start at the glyph line, got:\n%s", got)
	}
}

func TestExtractEntryBlockRunsToEOFWithoutTrailingHeading(t *testing.T) {
	doc := "**Fallout profile:** p.md\n\n**Glyph:** `y`\n**Marker axis:** m.\n**Aim axis:** a.\n"
	got, err := ExtractEntry(doc)
	if err != nil {
		t.Fatalf("ExtractEntry: %v", err)
	}
	if !strings.HasSuffix(got, "**Aim axis:** a.") {
		t.Errorf("block should run to EOF and be trimmed, got:\n%s", got)
	}
}
