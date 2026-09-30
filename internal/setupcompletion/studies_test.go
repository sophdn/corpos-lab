package setupcompletion

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// These SHA-256s are of the python gen_studies.py output for casg-direct, so the
// test fails if the Go rendering drifts one byte from the oracle.
const (
	rawCasgDirectSHA  = "8703f1db78727c4582402ac3628ad96c7c0cc5111eeaf00f030c8a8e3e4f3dd3"
	loopCasgDirectSHA = "85656a0cc726a35b82ec07c4ce6b2cef5bccba14b26bcfe51f294bb78dd76434"
)

func sha(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestRawTOMLBytesMatchOracle(t *testing.T) {
	if got := sha(RawTOML("casg-direct")); got != rawCasgDirectSHA {
		t.Errorf("RawTOML(casg-direct) sha = %s, want %s\n---\n%s", got, rawCasgDirectSHA, RawTOML("casg-direct"))
	}
}

func TestLoopTOMLBytesMatchOracle(t *testing.T) {
	if got := sha(LoopTOML("casg-direct")); got != loopCasgDirectSHA {
		t.Errorf("LoopTOML(casg-direct) sha = %s, want %s\n---\n%s", got, loopCasgDirectSHA, LoopTOML("casg-direct"))
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
