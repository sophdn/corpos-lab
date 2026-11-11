package battery

import (
	"context"
	"testing"
)

const validGlyphEntry = "**Fallout profile:** ../fallout/demo.md\n\n" +
	"**Glyph:** `demo-class`\n" +
	"**Y-fire:** the agent stands at a decision point.\n" +
	"\n" +
	"### Marker axis\n" +
	"> Invariant: taking X from Y.\n" +
	"\n" +
	"### Aim axis\n" +
	"> Invariant: taking M from Y.\n" +
	"\n" +
	"### Rest axis\n" +
	"> Characterization: neutral territory.\n"

func runLint(content string, reg RegistryReader) Result {
	return RunSequence(context.Background(), BuildStructuralLint(), Input{
		ItemID:         "demo-class",
		Content:        content,
		Registry:       reg,
		ContinueOnFail: true,
	})
}

func TestStructuralLint_StepCount(t *testing.T) {
	if got := BuildStructuralLint().StepCount(); got != 4 {
		t.Fatalf("structural lint has %d steps, want 4", got)
	}
}

func TestStructuralLint_ValidGlyphPasses(t *testing.T) {
	res := runLint(validGlyphEntry, &fakeRegistry{})
	if !res.Passed {
		t.Fatalf("valid glyph failed lint: %s", res.FailureReason)
	}
}

func TestStructuralLint_MissingRestAxisFails(t *testing.T) {
	content := "**Glyph:** `demo-class`\n**Y-fire:** here.\n### Marker axis\n> a\n### Aim axis\n> b\n"
	if runLint(content, &fakeRegistry{}).Passed {
		t.Fatal("expected lint to fail on missing Rest axis")
	}
}

func TestStructuralLint_IntentLanguageFails(t *testing.T) {
	content := "**Glyph:** `demo-class`\n**Y-fire:** fires when the agent believes it should.\n" +
		"### Marker axis\n> a\n### Aim axis\n> b\n### Rest axis\n> c\n"
	if runLint(content, &fakeRegistry{}).Passed {
		t.Fatal("expected lint to fail on intent-modeling language")
	}
}

func TestStructuralLint_DuplicateIdentityFails(t *testing.T) {
	// The identity appears twice in the registry: the first is self-excluded as
	// the candidate's own promoted entry, the second is a genuine duplicate.
	res := runLint(validGlyphEntry, &fakeRegistry{identities: []string{"demo-class", "demo-class"}})
	if res.Passed {
		t.Fatal("expected lint to fail on a duplicate identity")
	}
}

func TestStructuralLint_ContinueOnFailRunsEveryItem(t *testing.T) {
	// Intent language (item 2) and a missing Aim/Rest (item 10) both fail; under
	// ContinueOnFail all four items are measured.
	content := "**Glyph:** `demo-class`\n**Y-fire:** when the agent decides.\n### Marker axis\n> a\n"
	res := runLint(content, &fakeRegistry{})
	if res.Passed {
		t.Fatal("expected failure")
	}
	if len(res.StepResults) != 4 {
		t.Fatalf("ran %d steps, want all 4 under ContinueOnFail", len(res.StepResults))
	}
}
