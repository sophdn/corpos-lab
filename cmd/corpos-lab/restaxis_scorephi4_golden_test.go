package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// phi4Portal stands in for the llama.cpp portal: it labels a response OK when its
// prompt mentions "good", OF otherwise, and records every request body.
type phi4Portal struct {
	mu     sync.Mutex
	bodies []string
	fail   bool
}

func (p *phi4Portal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	_ = json.NewDecoder(r.Body).Decode(&req)
	raw, _ := json.Marshal(req)
	p.mu.Lock()
	p.bodies = append(p.bodies, string(raw))
	p.mu.Unlock()
	if p.fail {
		http.Error(w, "boom", http.StatusInternalServerError)
		return
	}
	content := "OF - unwarranted gate"
	if strings.Contains(fmt.Sprint(req["prompt"]), "good") {
		content = "OK\nit does the task"
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"content": "  " + content + "  "})
}

// TestScorePhi4Golden freezes score-phi4's written score files and errors over a
// study with two glyphs, scenarios sorted numerically (s2 before s10), a scenario
// with no material (skipped), missing response files (skipped), the --scenario
// re-score that merges into an existing file, and the error paths.
func TestScorePhi4Golden(t *testing.T) {
	study := t.TempDir()
	w := func(rel, body string) { writeFileT(t, filepath.Join(study, rel), body) }
	for _, g := range []string{"g1", "g2"} {
		w(g+"/CORRECT_COMPLETIONS.md", "## Scenario 2\nthe right move for 2\n## Scenario 10\nthe right move for 10\n")
		w(g+"/materials/scenario_2.md", "scenario two "+g)
		w(g+"/materials/scenario_10.md", "scenario ten "+g)
		for _, sd := range []string{"s2", "s10", "s3"} {
			w("runs/m/"+g+"/"+sd+"/out/responses/baseline_1.txt", "good "+g+sd)
			w("runs/m/"+g+"/"+sd+"/out/responses/glyph_only_2.txt", "bad "+g+sd)
		}
	}
	portal := &phi4Portal{}
	srv := httptest.NewServer(portal)
	defer srv.Close()

	var b strings.Builder
	note := func(label string, err error) {
		msg := "ok"
		if err != nil {
			msg = strings.ReplaceAll(err.Error(), study, "<study>")
			msg = strings.ReplaceAll(msg, srv.URL, "<portal>")
		}
		fmt.Fprintf(&b, "%s: %s\n", label, msg)
	}
	note("score all", raScorePhi4([]string{"--study", study, "--model", "m", "--portal", srv.URL}))
	// Rescore one scenario of g1: its old rows are replaced, other scenarios kept.
	w("runs/m/g1/s2/out/responses/glyph_only_2.txt", "good now")
	note("rescore g1 s2", raScorePhi4([]string{"--study", study, "--model", "m", "--glyph", "g1", "--scenario", "2", "--portal", srv.URL}))
	note("no study", raScorePhi4([]string{"--model", "m"}))
	note("bad scenario", raScorePhi4([]string{"--study", study, "--model", "m", "--scenario", "two"}))
	note("no glyph dir", raScorePhi4([]string{"--study", study, "--model", "absent", "--portal", srv.URL}))
	portal.fail = true
	note("portal error", raScorePhi4([]string{"--study", study, "--model", "m", "--glyph", "g2", "--portal", srv.URL}))

	scores, _ := filepath.Glob(filepath.Join(study, "scores", "*"))
	sort.Strings(scores)
	for _, f := range scores {
		raw, _ := os.ReadFile(f)
		fmt.Fprintf(&b, "== %s\n%s\n", filepath.Base(f), raw)
	}
	fmt.Fprintf(&b, "== first request\n%s\n", portal.bodies[0])
	got := b.String()

	golden := filepath.Join("testdata", "score_phi4.golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("score-phi4 drifted from %s:\n%s", golden, got)
	}
}
