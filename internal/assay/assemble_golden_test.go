package assay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAssemblePromptGolden freezes, for every condition and an unknown one, the
// prompt built from full materials, the error for scenario-only materials, and
// (for grounded_glyph) the error when only the ground is missing.
func TestAssemblePromptGolden(t *testing.T) {
	full := Materials{
		Scenario: "SCEN", Glyph: "GLYPH", Ground: "GROUND", Imperative: "IMP", Scrambled: "SCR",
		OffTarget: "OFF", Neutral: "NEU", GlyphMinusRest: "GMR", Duty: "DUTY", Corpus: "CORP",
		Annotated: "ANN", Cartographer: "CART", CartographerScan: "CSCAN", DomainImperative: "DIMP",
		NonPrescriptiveGround: "NPG", CanonAligned: "CAL", CanonConflict: "CCON",
		ScrambledCanon: "SCAN", OffTargetCanon: "OTC",
	}
	conds := []Condition{Baseline, GlyphOnly, GroundedGlyph, ImperativeOnly, ScrambledGlyph, OffTargetGlyph,
		NeutralPrefix, GlyphMinusRest, DutyOnly, CorpusOnly, AnnotatedInstrument, CartographerInstrument,
		CartographerScanInstrument, GroundOnly, DomainImperativeOnly, GroundNonPrescriptive, CanonAligned,
		CanonConflict, ScrambledCanon, OffTargetCanon, Condition("nope")}
	show := func(s string, err error) string {
		if err != nil {
			return "ERR " + err.Error()
		}
		return strings.ReplaceAll(s, "\n", `\n`)
	}
	var b strings.Builder
	for _, c := range conds {
		fmt.Fprintf(&b, "%s full: %s\n", c, show(AssemblePrompt(c, full)))
		fmt.Fprintf(&b, "%s bare: %s\n", c, show(AssemblePrompt(c, Materials{Scenario: "SCEN"})))
	}
	fmt.Fprintf(&b, "grounded_glyph no-ground: %s\n", show(AssemblePrompt(GroundedGlyph, Materials{Scenario: "SCEN", Glyph: "GLYPH"})))
	out := b.String()

	golden := filepath.Join("testdata", "assemble.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Fatalf("AssemblePrompt drifted from %s:\n%s", golden, out)
	}
}
