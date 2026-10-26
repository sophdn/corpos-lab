package setupcompletion

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The generated study TOML for casg-direct is pinned to golden files in
// testdata, so the test fails if the Go rendering drifts one byte. The goldens
// mask the two probe-image pins (oracleText): the pins move on every probe
// rebuild and re-pinning is automated (scripts/build-lab-images.sh --repin), so
// they must not move the oracle. TestTOMLsPerGlyph still checks that every TOML
// carries the pinned images, and --check-pins checks the pins against the build.
//
// When the generated shape changes on purpose, run
//
//	go test ./internal/setupcompletion/ -run Oracle -update
//
// which rewrites the goldens and prints the diff; review it and commit it.
var update = flag.Bool("update", false, "rewrite the scaffold oracle goldens")

// oracleText masks the probe-image pins so the oracle covers the generated
// shape and not the current build's digests.
func oracleText(s string) string {
	return strings.NewReplacer(LoopImg, "<LoopImg>", RawImg, "<RawImg>").Replace(s)
}

func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	want, err := os.ReadFile(path)
	if *update {
		if err == nil && string(want) == got {
			return
		}
		if werr := os.WriteFile(path, []byte(got), 0o644); werr != nil {
			t.Fatal(werr)
		}
		t.Logf("updated %s:\n%s", path, lineDiff(string(want), got))
		return
	}
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create it)", path, err)
	}
	if string(want) != got {
		t.Errorf("%s drifted from its golden (run with -update if intended):\n%s", name, lineDiff(string(want), got))
	}
}

// lineDiff lists the lines that differ, by line number.
func lineDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	var b strings.Builder
	for i := 0; i < max(len(w), len(g)); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			fmt.Fprintf(&b, "line %d:\n  - %s\n  + %s\n", i+1, wl, gl)
		}
	}
	return b.String()
}

func TestRawTOMLBytesMatchOracle(t *testing.T) {
	checkGolden(t, "raw_casg-direct.toml", oracleText(RawTOML("casg-direct")))
}

func TestLoopTOMLBytesMatchOracle(t *testing.T) {
	checkGolden(t, "loop_casg-direct.toml", oracleText(LoopTOML("casg-direct")))
}

func TestLineDiff(t *testing.T) {
	if d := lineDiff("a\nb", "a\nc\nd"); !strings.Contains(d, "line 2") || !strings.Contains(d, "line 3") {
		t.Errorf("lineDiff = %q", d)
	}
}

func TestMaterials(t *testing.T) {
	want := `scenario   = "../../alphabet-wide-mechanism-and-grounding-assay/g/materials/scenario_1.md"
glyph      = "../../alphabet-wide-mechanism-and-grounding-assay/g/materials/glyph.md"
imperative = "../../alphabet-wide-mechanism-and-grounding-assay/g/materials/imperative.md"
`
	if got := Materials("g"); got != want {
		t.Errorf("Materials(g) = %q, want %q", got, want)
	}
}

func TestTOMLsPerGlyph(t *testing.T) {
	if len(Glyphs) != 10 {
		t.Fatalf("Glyphs has %d entries, want 10", len(Glyphs))
	}
	seen := map[string]bool{}
	for _, g := range Glyphs {
		raw := RawTOML(g)
		loop := LoopTOML(g)
		if !strings.Contains(raw, "name = \"setup-raw-"+g+"-qwen38\"") {
			t.Errorf("raw TOML for %q missing its name line", g)
		}
		if !strings.Contains(loop, "name = \"setup-loop-"+g+"-qwen38\"") {
			t.Errorf("loop TOML for %q missing its name line", g)
		}
		if !strings.Contains(raw, RawImg) || !strings.Contains(loop, LoopImg) {
			t.Errorf("glyph %q TOMLs missing their pinned image digests", g)
		}
		if !strings.Contains(loop, "[loop]\n") {
			t.Errorf("loop TOML for %q missing the [loop] section", g)
		}
		if strings.Contains(raw, "[loop]") {
			t.Errorf("raw TOML for %q unexpectedly has a [loop] section", g)
		}
		// Both end with the shared sampling block.
		if !strings.HasSuffix(raw, Sampling) || !strings.HasSuffix(loop, Sampling) {
			t.Errorf("glyph %q TOMLs do not end with the sampling block", g)
		}
		if seen[g] {
			t.Errorf("duplicate glyph %q", g)
		}
		seen[g] = true
	}
}
