package study

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// materialFields names every MaterialsDef source field with a setter, so a test
// can set all of them or blank one at a time.
var materialFields = []struct {
	name string
	set  func(*MaterialsDef, string)
}{
	{"scenario", func(m *MaterialsDef, v string) { m.Scenario = v }},
	{"glyph", func(m *MaterialsDef, v string) { m.Glyph = v }},
	{"ground", func(m *MaterialsDef, v string) { m.Ground = v }},
	{"imperative", func(m *MaterialsDef, v string) { m.Imperative = v }},
	{"scrambled", func(m *MaterialsDef, v string) { m.Scrambled = v }},
	{"off_target", func(m *MaterialsDef, v string) { m.OffTarget = v }},
	{"neutral", func(m *MaterialsDef, v string) { m.Neutral = v }},
	{"glyph_minus_rest", func(m *MaterialsDef, v string) { m.GlyphMinusRest = v }},
	{"duty", func(m *MaterialsDef, v string) { m.Duty = v }},
	{"corpus", func(m *MaterialsDef, v string) { m.Corpus = v }},
	{"annotated", func(m *MaterialsDef, v string) { m.Annotated = v }},
	{"cartographer", func(m *MaterialsDef, v string) { m.Cartographer = v }},
	{"cartographer_scan", func(m *MaterialsDef, v string) { m.CartographerScan = v }},
	{"domain_imperative", func(m *MaterialsDef, v string) { m.DomainImperative = v }},
	{"ground_nonprescriptive", func(m *MaterialsDef, v string) { m.NonPrescriptiveGround = v }},
	{"canon_aligned", func(m *MaterialsDef, v string) { m.CanonAligned = v }},
	{"canon_conflict", func(m *MaterialsDef, v string) { m.CanonConflict = v }},
	{"scrambled_canon", func(m *MaterialsDef, v string) { m.ScrambledCanon = v }},
	{"off_target_canon", func(m *MaterialsDef, v string) { m.OffTargetCanon = v }},
}

// TestMaterializeGolden freezes study.json and every written file for a
// definition naming all nineteen materials and both loop files, then the exact
// error when each material file in turn is missing.
func TestMaterializeGolden(t *testing.T) {
	base, err := LoadDef(writeDef(t, validDef, allMaterials()))
	if err != nil {
		t.Fatalf("valid fixture: %v", err)
	}
	src := t.TempDir()
	d := base
	d.baseDir = src
	for _, f := range materialFields {
		f.set(&d.Materials, "src_"+f.name+".md")
		writeFileT(t, filepath.Join(src, "src_"+f.name+".md"), "BODY "+f.name)
	}
	d.Loop = LoopDef{Preamble: "pre.md", Sandbox: "sb.json", StepCap: 7, CallTokens: 99, Stop: []string{"STOP"}}
	writeFileT(t, filepath.Join(src, "pre.md"), "PREAMBLE")
	writeFileT(t, filepath.Join(src, "sb.json"), `{"f":"x"}`)

	var b strings.Builder
	in := t.TempDir()
	if err := d.Materialize(in); err != nil {
		t.Fatalf("Materialize: %v", err)
	}
	entries, _ := os.ReadDir(in)
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, n := range names {
		raw, _ := os.ReadFile(filepath.Join(in, n))
		fmt.Fprintf(&b, "== %s\n%s\n", n, raw)
	}
	for _, f := range materialFields {
		missing := d
		missing.Materials = d.Materials
		f.set(&missing.Materials, "absent_"+f.name+".md")
		err := missing.Materialize(t.TempDir())
		msg := "ok"
		if err != nil {
			msg = strings.ReplaceAll(err.Error(), src, "<src>")
		}
		fmt.Fprintf(&b, "missing %s: %s\n", f.name, msg)
	}
	out := b.String()
	golden := filepath.Join("testdata", "materialize.golden")
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
		t.Fatalf("Materialize drifted from %s:\n%s", golden, out)
	}
}

func writeFileT(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
