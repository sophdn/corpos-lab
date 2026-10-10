package alphabetassay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestAnalyzeGolden freezes Analyze's whole result (as JSON) and FormatTable's
// text over three present classes and seven absent ones: one fully rated across
// several cells, one with ids only one rater coded, and one with no jointly-rated
// id (null agreement).
func TestAnalyzeGolden(t *testing.T) {
	inputs := map[string]ClassInput{}
	full := ClassInput{Key: map[string]KeyEntry{}, ACodes: map[string]string{}, BCodes: map[string]string{}}
	for i := 0; i < 24; i++ {
		id := fmt.Sprintf("f%02d", i)
		full.Key[id] = KeyEntry{Model: TableModels[i%3], Condition: []string{"baseline", "alphabet"}[i%2]}
		full.ACodes[id] = Codes[i%5]
		full.BCodes[id] = Codes[(i/2)%5]
	}
	inputs[Classes[0]] = full
	partial := ClassInput{
		Key:    map[string]KeyEntry{"p1": {Model: "phi4", Condition: "baseline"}, "p2": {Model: "phi4", Condition: "baseline"}, "p3": {Model: "mistral", Condition: "alphabet"}},
		ACodes: map[string]string{"p1": "C", "p2": "C", "p3": "I"},
		BCodes: map[string]string{"p1": "C", "p3": "I"},
	}
	inputs[Classes[3]] = partial
	inputs[Classes[5]] = ClassInput{Key: map[string]KeyEntry{"z": {Model: "qwen38", Condition: "baseline"}}, ACodes: map[string]string{"z": "C"}, BCodes: map[string]string{}}

	res := Analyze(inputs)
	js, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	out := string(js) + "\n" + FormatTable(res)

	golden := filepath.Join("testdata", "analyze.golden")
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
		t.Fatalf("Analyze output drifted from %s:\n%s", golden, out)
	}
}
