// Package canon validates the shape of the private glyph canon: the typed glyph
// file schema (suggestion 166), the typed scenario schema (typed-scenarios task),
// and — via disclosure.ParseRegistry — the certification-registry table.
//
// This is the Go home for the checks the private corpus's gate enforces. The
// logic lives here, under the corpos-lab gate and coverage floor; the private
// repo's pre-commit gate and CI invoke the built binary rather than a
// second-language script. It replaces the retired python linters
// (glyph_schema_lint.py, scenario_lint.py). The functions take file content as a
// string and embed no corpus data.
package canon

import (
	"fmt"
	"regexp"
	"strings"

	"corpos-lab/internal/disclosure"
)

const (
	glyphMarker   = "**Glyph:**"
	falloutMarker = "**Fallout profile:**"
)

var titleRE = regexp.MustCompile(`(?m)^#\s+Glyph Candidate:\s*\S`)

var batteryStatusRE = regexp.MustCompile(`^\s*\*\*Battery status:\*\*`)

// glyphBanned are the patterns rejected in the region ABOVE the AC-4 block — the
// bloat suggestion 166 removes. Matched line-by-line against the header region.
var glyphBanned = []struct {
	re    *regexp.Regexp
	label string
}{
	{regexp.MustCompile(`(?i)^\s*#{1,6}\s*(RETURNED|DEMOTED|PROMOTED)\b`), "history banner (RETURNED/DEMOTED/PROMOTED heading)"},
	{regexp.MustCompile(`(?i)^\s*>\s*\*{0,2}(RETURNED|DEMOTED|PROMOTED)\b`), "history banner (blockquote PROMOTED/RETURNED/DEMOTED)"},
	{regexp.MustCompile(`^\s*\*\*Derivation:\*\*`), "**Derivation:** provenance prose (recover from git history)"},
	{regexp.MustCompile(`^\s*\*\*Decision class:\*\*`), "**Decision class:** line (duplicates the AC-4 Y-fire block)"},
	{regexp.MustCompile(`^\s*#{1,6}\s*AC-[0-9]`), "AC-N derivation-scaffold heading (AC-1/2/3/4)"},
	{regexp.MustCompile(`^\s*\*\*Battery findings`), "**Battery findings** dump (belongs in battery-runs/, not the glyph file)"},
	{regexp.MustCompile(`(?i)^\s*\|\s*Entry\s*\|`), "battery PASS/FAIL table header"},
	{regexp.MustCompile(`(?i)^\s*\|.*\b(PASS|FAIL|DEFERRED)\b.*\|`), "battery PASS/FAIL table row"},
}

// GlyphFileSchema returns the typed-file-schema violations of a glyph candidate
// file (empty = clean). Required: a `# Glyph Candidate:` title, exactly one
// `**Fallout profile:**` line, and the AC-4 `**Glyph:**` block. Banned above the
// block: history banners, `**Derivation:**`, `**Decision class:**`, AC-N
// scaffold headings, and battery-findings dumps or PASS/FAIL tables. Only the
// region above the `**Glyph:**` line is inspected, so a diet that obeys the
// schema never changes the frozen digest.
func GlyphFileSchema(content string) []string {
	lines := strings.Split(content, "\n")

	glyphIdx, glyphCount := -1, 0
	for i, ln := range lines {
		if strings.HasPrefix(ln, glyphMarker) {
			if glyphIdx < 0 {
				glyphIdx = i
			}
			glyphCount++
		}
	}

	var problems []string
	header := lines
	if glyphIdx < 0 {
		problems = append(problems, "missing the AC-4 block: no `**Glyph:**` line")
	} else {
		if glyphCount > 1 {
			problems = append(problems, fmt.Sprintf("more than one `**Glyph:**` line (%d)", glyphCount))
		}
		header = lines[:glyphIdx]
	}

	if !titleRE.MatchString(content) {
		problems = append(problems, "missing the `# Glyph Candidate: <slug>` title")
	}

	fallout := 0
	for _, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), falloutMarker) {
			fallout++
		}
	}
	switch {
	case fallout == 0:
		problems = append(problems, "missing the `**Fallout profile:**` line (Item 15's referent)")
	case fallout > 1:
		problems = append(problems, fmt.Sprintf("more than one `**Fallout profile:**` line (%d)", fallout))
	}

	status := 0
	for _, ln := range header {
		if batteryStatusRE.MatchString(ln) {
			status++
		}
	}
	if status > 1 {
		problems = append(problems, fmt.Sprintf("more than one `**Battery status:**` line (%d); it must be a single pointer line", status))
	}

	for i, ln := range header {
		for _, b := range glyphBanned {
			if b.re.MatchString(ln) {
				problems = append(problems, fmt.Sprintf("line %d: banned — %s", i+1, b.label))
				break
			}
		}
	}
	return problems
}

var scenarioRoles = map[string]bool{"fire": true, "not-fire": true}

// scenarioMetaLeak are meta-layer terms a self-contained scenario prompt must not
// contain: a scenario is a task, not a description of the glyph (suggestion 170).
var scenarioMetaLeak = []struct {
	re    *regexp.Regexp
	label string
}{
	{regexp.MustCompile(`(?i)\bglyph\b`), "the word 'glyph'"},
	{regexp.MustCompile(`(?i)\bY-not-fire\b|\bY-fire\b`), "Y-fire / Y-not-fire"},
	{regexp.MustCompile(`(?i)\b(Marker|Aim|Rest)\s+axis\b`), "an axis name (Marker/Aim/Rest axis)"},
	{regexp.MustCompile(`(?i)\bdecision class\b`), "'decision class'"},
}

// ScenarioSchema returns the validity violations of a typed scenario file (empty
// = clean). A scenario is `--- glyph: <slug> / role: fire|not-fire ---` frontmatter
// over a non-empty, self-contained prompt body that names none of the glyph
// meta-layer (glyph, its axes, Y-fire/Y-not-fire, decision class).
func ScenarioSchema(content string) []string {
	fm, body, ok := parseFrontmatter(content)
	if !ok {
		return []string{"missing the `--- glyph: … / role: … ---` frontmatter"}
	}
	var problems []string
	if fm["glyph"] == "" {
		problems = append(problems, "frontmatter missing `glyph: <glyph-slug>`")
	}
	if role := fm["role"]; !scenarioRoles[role] {
		problems = append(problems, fmt.Sprintf("frontmatter `role:` must be fire or not-fire (got %q)", role))
	}
	if strings.TrimSpace(body) == "" {
		problems = append(problems, "empty scenario prompt body")
	}
	for _, m := range scenarioMetaLeak {
		if m.re.MatchString(body) {
			problems = append(problems, "not self-contained: prompt body mentions "+m.label+" — a scenario is a task, not a description of the glyph")
			break
		}
	}
	return problems
}

// parseFrontmatter splits a `---`-delimited YAML-ish frontmatter block from the
// body. It returns (fields, body, true) when a frontmatter block is present, or
// (nil, content, false) when it is absent.
func parseFrontmatter(content string) (map[string]string, string, bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, content, false
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, content, false
	}
	fm := make(map[string]string)
	for _, ln := range lines[1:end] {
		if k, v, found := strings.Cut(ln, ":"); found {
			fm[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return fm, strings.Join(lines[end+1:], "\n"), true
}

// RegistryShape validates the certification-registry table shape of an ALPHABET.md
// (the "expected shape of the alphabet" check). It reuses disclosure.ParseRegistry
// so the shape rule has one definition.
func RegistryShape(md string) error {
	_, err := disclosure.ParseRegistry(md)
	return err
}
