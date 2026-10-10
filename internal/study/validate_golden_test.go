package study

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/runner"
)

// TestValidateGolden freezes the exact message (or acceptance) of every rule in
// validate, in rule order, by mutating one parsed valid definition per case.
func TestValidateGolden(t *testing.T) {
	base, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatalf("valid fixture: %v", err)
	}
	allMats := MaterialsDef{
		Scenario: "s", Glyph: "g", Ground: "gr", Imperative: "i", Scrambled: "sc", OffTarget: "o",
		Neutral: "n", GlyphMinusRest: "gm", Duty: "d", Corpus: "c", Annotated: "a", Cartographer: "ca",
		CartographerScan: "cs", DomainImperative: "di", NonPrescriptiveGround: "np", CanonAligned: "cal",
		CanonConflict: "cc", ScrambledCanon: "scc", OffTargetCanon: "otc",
	}
	conds := []string{"baseline", "glyph_only", "grounded_glyph", "imperative_only", "scrambled_glyph",
		"off_target_glyph", "neutral_prefix", "glyph_minus_rest", "duty_only", "corpus_only",
		"annotated_instrument", "cartographer_instrument", "cartographer_scan_instrument", "ground_only",
		"domain_imperative_only", "ground_nonprescriptive", "canon_aligned", "canon_conflict",
		"scrambled_canon", "off_target_canon", "teleport"}

	type tc struct {
		name string
		mut  func(*Def)
	}
	cases := []tc{
		{"valid", func(*Def) {}},
		{"no name", func(d *Def) { d.Name = "" }},
		{"no assay", func(d *Def) { d.Assay = "" }},
		{"no item_id", func(d *Def) { d.ItemID = "" }},
		{"no image", func(d *Def) { d.Image = "" }},
		{"no base_url", func(d *Def) { d.Model.BaseURL = "" }},
		{"no model_id", func(d *Def) { d.Model.ModelID = "" }},
		{"no scenario", func(d *Def) { d.Materials.Scenario = "" }},
		{"bad assay", func(d *Def) { d.Assay = "astrology" }},
		{"loop no preamble", func(d *Def) { d.Assay = runner.LoopAssay }},
		{"loop no sandbox", func(d *Def) { d.Assay = runner.LoopAssay; d.Loop.Preamble = "p" }},
		{"completion no template", func(d *Def) { d.Model.Endpoint = "completion" }},
		{"completion with template", func(d *Def) { d.Model.Endpoint = "completion"; d.Model.PromptTemplate = "<s>{prompt}" }},
		{"chat endpoint", func(d *Def) { d.Model.Endpoint = "chat" }},
		{"bad endpoint", func(d *Def) { d.Model.Endpoint = "smoke-signals" }},
		{"no conditions", func(d *Def) { d.Conditions = nil }},
		{"zero runs", func(d *Def) { d.RunsPerCell = 0 }},
		{"direction only", func(d *Def) { d.Direction = "lift" }},
		{"no temperature", func(d *Def) { d.Sampling.Temperature = nil }},
		{"grounded_glyph glyph only", func(d *Def) {
			d.Conditions = []string{"grounded_glyph"}
			d.Materials = MaterialsDef{Scenario: "s", Glyph: "g"}
		}},
	}
	for _, c := range conds {
		c := c
		cases = append(cases,
			tc{c + " bare", func(d *Def) { d.Conditions = []string{c}; d.Materials = MaterialsDef{Scenario: "s"} }},
			tc{c + " full", func(d *Def) { d.Conditions = []string{c}; d.Materials = allMats }})
	}
	var b strings.Builder
	for _, c := range cases {
		d := base
		d.Conditions = append([]string(nil), base.Conditions...)
		c.mut(&d)
		msg := "ok"
		if err := d.validate(); err != nil {
			msg = err.Error()
		}
		fmt.Fprintf(&b, "%s: %s\n", c.name, msg)
	}
	out := b.String()
	golden := filepath.Join("testdata", "validate.golden")
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
		t.Fatalf("validate drifted from %s:\n%s", golden, out)
	}
}
