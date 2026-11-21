package restaxis

import (
	"regexp"
	"strings"
	"testing"
)

func TestStripFrontmatter(t *testing.T) {
	in := "---\ntype: x\nscenarios: 4\n---\n\n# Title\n\nbody"
	if got := stripFrontmatter(in); got != "# Title\n\nbody" {
		t.Errorf("stripFrontmatter = %q", got)
	}
	// no frontmatter: unchanged
	if got := stripFrontmatter("# Title\nbody"); got != "# Title\nbody" {
		t.Errorf("no-frontmatter changed: %q", got)
	}
	// only one delimiter (fewer than 3 parts): unchanged
	if got := stripFrontmatter("---\nnot closed"); got != "---\nnot closed" {
		t.Errorf("unclosed frontmatter changed: %q", got)
	}
}

func TestReplaceFirst(t *testing.T) {
	re := regexp.MustCompile(`^#\s.*\n`)
	if got := replaceFirst(re, "# Heading\nrest", ""); got != "rest" {
		t.Errorf("replaceFirst = %q, want 'rest'", got)
	}
	// no match: unchanged
	if got := replaceFirst(re, "no heading here", "X"); got != "no heading here" {
		t.Errorf("no-match replaceFirst changed: %q", got)
	}
}

func TestMergeCompletions(t *testing.T) {
	main := "---\ntype: gt\nscenarios: 4\n---\n\n# glyph — completions\n\n## Scenario 1 — a\n\nbody1\n"
	nw := "---\ntype: gt\nscenarios: 16\n---\n\n# glyph — NEW\n\n## Scenario 5 — e\n\nbody5\n"
	got := MergeCompletions(main, nw)

	if !strings.Contains(got, "scenarios: 20") {
		t.Errorf("scenario count not bumped: %q", got)
	}
	if strings.Contains(got, "scenarios: 4") || strings.Contains(got, "scenarios: 16") {
		t.Errorf("stale scenario counts remain: %q", got)
	}
	// new frontmatter and its redundant H1 must be gone
	if strings.Contains(got, "type: gt\nscenarios") && strings.Count(got, "type: gt") != 1 {
		t.Errorf("new frontmatter not stripped: %q", got)
	}
	if strings.Contains(got, "# glyph — NEW") {
		t.Errorf("redundant new H1 not dropped: %q", got)
	}
	// original H1 kept; joined with a blank line; single trailing newline
	if !strings.Contains(got, "# glyph — completions") {
		t.Errorf("original H1 lost")
	}
	if !strings.Contains(got, "body1\n\n## Scenario 5 — e") {
		t.Errorf("join shape wrong: %q", got)
	}
	if !strings.HasSuffix(got, "body5\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("trailing newline not normalized: %q", got[len(got)-6:])
	}
}

func TestMergeCompletionsNoScenariosLine(t *testing.T) {
	// main lacks a scenarios: line -> no bump, still merges
	main := "no frontmatter here\n"
	nw := "just new body\n"
	got := MergeCompletions(main, nw)
	if got != "no frontmatter here\n\njust new body\n" {
		t.Errorf("merge without frontmatter = %q", got)
	}
}

func TestMergeCompletionsNewWithoutH1(t *testing.T) {
	main := "scenarios: 4\nmain body"
	nw := "plain new body, no heading"
	got := MergeCompletions(main, nw)
	if got != "scenarios: 20\nmain body\n\nplain new body, no heading\n" {
		t.Errorf("merge new-without-H1 = %q", got)
	}
}
