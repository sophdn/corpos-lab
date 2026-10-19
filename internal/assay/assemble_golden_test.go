package assay

import (
	"fmt"
	"strings"
	"testing"

	"corpos-lab/internal/testutil"
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

	testutil.Golden(t, "assemble.golden", "AssemblePrompt", out)
}
