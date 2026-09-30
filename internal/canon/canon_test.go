package canon

import (
	"strings"
	"testing"
)

const cleanGlyphFile = `# Glyph Candidate: example-class

**Battery status:** promoted 2026-09-06; see ../battery-runs/BATTERY_RUN_example.md

**Fallout profile:** ../fallout-profiles/TASK_fallout-example.md

**Glyph:** ` + "`example-class`" + `

**Y-fire:** the agent is at the decision point.

### Marker axis
> Invariant: taking X from Y produces Z-marker.
`

func hasProblem(problems []string, substr string) bool {
	for _, p := range problems {
		if strings.Contains(p, substr) {
			return true
		}
	}
	return false
}

func TestGlyphFileSchemaClean(t *testing.T) {
	if got := GlyphFileSchema(cleanGlyphFile); len(got) != 0 {
		t.Fatalf("clean file reported problems: %v", got)
	}
}

func TestGlyphFileSchemaMissingParts(t *testing.T) {
	cases := []struct {
		name, content, want string
	}{
		{"no glyph block", "# Glyph Candidate: x\n\n**Fallout profile:** f\n", "no `**Glyph:**` line"},
		{"no title", "**Fallout profile:** f\n\n**Glyph:** `x`\n", "title"},
		{"no fallout", "# Glyph Candidate: x\n\n**Glyph:** `x`\n", "Fallout profile"},
		{"two glyph blocks", "# Glyph Candidate: x\n\n**Fallout profile:** f\n\n**Glyph:** `a`\n**Glyph:** `b`\n", "more than one `**Glyph:**`"},
		{"two fallout", "# Glyph Candidate: x\n\n**Fallout profile:** a\n**Fallout profile:** b\n\n**Glyph:** `x`\n", "more than one `**Fallout profile:**`"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GlyphFileSchema(c.content); !hasProblem(got, c.want) {
				t.Errorf("want a problem containing %q, got %v", c.want, got)
			}
		})
	}
}

func TestGlyphFileSchemaBannedAboveBlock(t *testing.T) {
	head := "# Glyph Candidate: x\n\n**Fallout profile:** f\n\n"
	tail := "\n**Glyph:** `x`\n"
	cases := []struct {
		name, banned, want string
	}{
		{"returned heading", "## RETURNED — 2026-04-03", "history banner"},
		{"blockquote promoted", "> **PROMOTED** — lifted", "history banner"},
		{"derivation", "**Derivation:** fresh context", "Derivation"},
		{"decision class", "**Decision class:** an agent…", "Decision class"},
		{"ac heading", "## AC-2 — Rest axis", "AC-N"},
		{"battery findings", "**Battery findings (run 1):** …", "Battery findings"},
		{"table header", "| Entry | 1 | 2 | Overall |", "table header"},
		{"table row", "| x | PASS | PASS | **PASS** |", "table row"},
		{"two battery status", "**Battery status:** a\n**Battery status:** b", "more than one `**Battery status:**`"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GlyphFileSchema(head + c.banned + tail)
			if !hasProblem(got, c.want) {
				t.Errorf("want a problem containing %q, got %v", c.want, got)
			}
		})
	}
}

func TestScenarioSchemaClean(t *testing.T) {
	fire := "---\nglyph: example-class\nrole: fire\n---\nYou maintain a registry capped at 15 entries; it holds 15. Add one more and give the steps.\n"
	if got := ScenarioSchema(fire); len(got) != 0 {
		t.Fatalf("clean fire scenario reported problems: %v", got)
	}
	notFire := "---\nglyph: example-class\nrole: not-fire\n---\nThe registry holds 3 of 15 entries. Add one and give the steps.\n"
	if got := ScenarioSchema(notFire); len(got) != 0 {
		t.Fatalf("clean not-fire scenario reported problems: %v", got)
	}
}

func TestScenarioSchemaViolations(t *testing.T) {
	cases := []struct {
		name, content, want string
	}{
		{"no frontmatter", "just a prompt with no frontmatter\n", "frontmatter"},
		{"missing glyph", "---\nrole: fire\n---\nbody\n", "missing `glyph:"},
		{"bad role", "---\nglyph: x\nrole: maybe\n---\nbody\n", "must be fire or not-fire"},
		{"empty body", "---\nglyph: x\nrole: fire\n---\n\n", "empty scenario prompt body"},
		{"unterminated frontmatter", "---\nglyph: x\nrole: fire\nbody with no closing fence\n", "frontmatter"},
		{"meta leak glyph", "---\nglyph: x\nrole: fire\n---\nThis glyph fires when…\n", "mentions the word 'glyph'"},
		{"meta leak yfire", "---\nglyph: x\nrole: fire\n---\nAt Y-fire the agent…\n", "Y-fire"},
		{"meta leak axis", "---\nglyph: x\nrole: fire\n---\nThe Rest axis says…\n", "axis name"},
		{"meta leak decision class", "---\nglyph: x\nrole: fire\n---\nThis decision class covers…\n", "decision class"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ScenarioSchema(c.content); !hasProblem(got, c.want) {
				t.Errorf("want a problem containing %q, got %v", c.want, got)
			}
		})
	}
}

func TestRegistryShape(t *testing.T) {
	digest := strings.Repeat("a", 64)
	valid := "# Registry\n\n" +
		"| slug | canonical definition file | definition digest (sha256) |\n" +
		"|------|---------------------------|----------------------------|\n" +
		"| example-class | `candidates/CANDIDATE_example-class_2026-09-27.md` | `" + digest + "` |\n"
	if err := RegistryShape(valid); err != nil {
		t.Fatalf("valid registry: %v", err)
	}
	if err := RegistryShape("no table here\n"); err == nil {
		t.Error("malformed registry: want error, got nil")
	}
}
