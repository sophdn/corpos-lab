package restaxis

// This file ports merge_completions.py: it mechanically merges a per-glyph
// CORRECT_COMPLETIONS_NEW.md (scenarios 5-20) into CORRECT_COMPLETIONS.md
// (scenarios 1-4). The transform is pure and sans-IO — the caller reads both
// files and writes the merged result (and deletes the _NEW file). The output is
// byte-identical to the python string assembly.

import (
	"regexp"
	"strings"
)

// scenariosCountRe matches a "scenarios: <n>" frontmatter line, mirroring the
// python r"(?m)^scenarios:\s*\d+\s*$".
var scenariosCountRe = regexp.MustCompile(`(?m)^scenarios:\s*\d+\s*$`)

// leadingH1Re matches a leading top-level "# ..." heading line, mirroring the
// python r"^#\s.*\n" applied without the multiline flag (so it anchors at the
// start of the string only).
var leadingH1Re = regexp.MustCompile(`^#\s.*\n`)

// MergeCompletions merges the _NEW body into the main body: it bumps the main
// frontmatter's scenario count to 20, strips the _NEW frontmatter and a leading
// redundant H1, and joins them. It reproduces merge_completions.py:
//
//	mt2 = re.sub(r"(?m)^scenarios:\s*\d+\s*$", "scenarios: 20", mt)
//	body_new = strip_frontmatter(nt).strip()
//	body_new = re.sub(r"^#\s.*\n", "", body_new, count=1).strip()
//	merged = mt2.rstrip() + "\n\n" + body_new + "\n"
func MergeCompletions(mainText, newText string) string {
	mt2 := scenariosCountRe.ReplaceAllString(mainText, "scenarios: 20")
	bodyNew := pyStrip(stripFrontmatter(newText))
	bodyNew = pyStrip(replaceFirst(leadingH1Re, bodyNew, ""))
	return pyRStrip(mt2) + "\n\n" + bodyNew + "\n"
}

// stripFrontmatter removes a leading YAML frontmatter block (--- ... ---) if
// present, mirroring the python strip_frontmatter: split on "---" into at most
// three parts and, when there are exactly three, return the third with leading
// newlines removed.
func stripFrontmatter(text string) string {
	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) == 3 {
			return strings.TrimLeft(parts[2], "\n")
		}
	}
	return text
}

// replaceFirst replaces the first match of re in s with repl, matching python
// re.sub(..., count=1). With a start-anchored pattern there is at most one match.
func replaceFirst(re *regexp.Regexp, s, repl string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + repl + s[loc[1]:]
}
