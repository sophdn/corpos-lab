package setupcompletion

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestTallyReportWideGolden freezes the report over two glyphs, both setups and
// every condition, with ids only one rater coded and a mix of C/Ii/other codes.
func TestTallyReportWideGolden(t *testing.T) {
	key := map[string]KeyEntry{}
	a := map[string]string{}
	b := map[string]string{}
	codes := []string{"C", "Ii", "I", "N"}
	n := 0
	for _, glyph := range []string{"beta", "alpha"} {
		for _, setup := range Setups {
			for _, cond := range Conditions {
				for run := 1; run <= 3; run++ {
					rid := fmt.Sprintf("%s-%s-%s-%d", glyph, setup, cond, run)
					key[rid] = KeyEntry{Glyph: glyph, Setup: setup, Condition: cond, Run: run}
					if n%7 != 3 {
						a[rid] = codes[n%4]
					}
					if n%5 != 1 {
						b[rid] = codes[(n/2)%4]
					}
					n++
				}
			}
		}
	}
	out := TallyReport(key, a, b)
	golden := filepath.Join("testdata", "tally_wide.golden")
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
		t.Fatalf("TallyReport drifted from %s:\n%s", golden, out)
	}
}
