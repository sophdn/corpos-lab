package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"corpos-lab/internal/restaxis"
)

// restAxisModes maps a rest-axis subcommand to its handler. Each ported tool
// registers its mode from an init() in its own file, so a new tool adds a file
// without editing this dispatcher.
var restAxisModes = map[string]func([]string) error{}

// registerRestAxisMode adds a mode handler. It panics on a duplicate name so a
// collision fails loudly at startup rather than silently shadowing.
func registerRestAxisMode(name string, fn func([]string) error) {
	if _, dup := restAxisModes[name]; dup {
		panic("rest-axis: duplicate mode " + name)
	}
	restAxisModes[name] = fn
}

func restAxisModeList() string {
	names := make([]string, 0, len(restAxisModes))
	for n := range restAxisModes {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func init() {
	registerRestAxisMode("mech", raMech)
	registerRestAxisMode("analyze", raAnalyze)
}

// runRestAxis dispatches to the ported rest-axis-overhead-benchmark tools in
// internal/restaxis. Modes register themselves (see restAxisModes).
func runRestAxis(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: corpos-lab rest-axis <mode> … (modes: %s)\n", restAxisModeList())
		return 2
	}
	mode, rest := args[0], args[1:]
	fn, ok := restAxisModes[mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "rest-axis: unknown mode %q (modes: %s)\n", mode, restAxisModeList())
		return 2
	}
	if err := fn(rest); err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: rest-axis %s: %v\n", mode, err)
		return 1
	}
	return 0
}

// raMech walks <study-dir>/runs/<model>/<glyph>/<scenario>/out/responses/<arm>_<seed>.txt,
// scores each response mechanically, writes <study-dir>/scores/<model>.mech.json,
// and prints per-arm flag rates. Ported from mech_score.py.
func raMech(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("need a study-dir")
	}
	study := args[0]
	models := args[1:]
	if len(models) == 0 {
		models = []string{"qwen38", "mistral", "phi4"}
	}
	for _, model := range models {
		resp, err := readRestAxisResponses(study, model)
		if err != nil {
			return err
		}
		rows := restaxis.MechScore(resp)
		out := filepath.Join(study, "scores", model+".mech.json")
		if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
			return err
		}
		b, err := json.MarshalIndent(rows, "", " ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, b, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s: %d responses -> %s\n", model, len(rows), filepath.Base(out))
		for _, ar := range restaxis.ArmRates(rows) {
			pct := 0.0
			if ar.Total > 0 {
				pct = 100 * float64(ar.Flagged) / float64(ar.Total)
			}
			fmt.Fprintf(os.Stderr, "  %-18s mech-flag rate %d/%d = %.0f%%\n", ar.Arm, ar.Flagged, ar.Total, pct)
		}
	}
	return nil
}

// raAnalyze joins rater A (de-anonymized blind bundles), rater B (scores/*.B.json),
// and the mechanical flags (scores/*.mech.json); prints A/B agreement and
// per-(model,arm) over-fire rates; and writes scores/disagreements.json. Ported
// from analyze.py. Rater A lives outside the repo (a blind-scoring bundle dir), so
// its path is a required flag rather than a hardcoded path.
func raAnalyze(args []string) error {
	raterA, rest := flagValue(args, "--rater-a")
	if raterA == "" {
		return fmt.Errorf("need --rater-a <dir> (the rater A blind bundles: <model>.<glyph>/{map,labels}.json)")
	}
	if len(rest) != 1 {
		return fmt.Errorf("need a study-dir")
	}
	study := rest[0]
	scores := filepath.Join(study, "scores")

	a, err := loadRaterA(raterA)
	if err != nil {
		return err
	}
	b, err := loadRaterB(scores)
	if err != nil {
		return err
	}
	m, err := loadMechFlags(scores)
	if err != nil {
		return err
	}

	ag := restaxis.ComputeAgreement(a, b)
	fmt.Fprintf(os.Stderr, "rater A: %d | rater B: %d | mech: %d\n", len(a), len(b), len(m))
	fmt.Fprintf(os.Stderr, "A/B on %d shared: over-fire agreement %.1f%% | exact-label %.1f%%\n",
		ag.Shared, ag.OverFireAgree, ag.ExactLabel)
	for _, r := range restaxis.Rates(a, b, m) {
		pct := 0.0
		if r.N > 0 {
			pct = 100 * float64(r.OF) / float64(r.N)
		}
		fmt.Fprintf(os.Stderr, "  %-8s %-18s %-4s %5.1f%% (%d/%d)\n", r.Model, r.Arm, r.Rater, pct, r.OF, r.N)
	}

	dis := restaxis.Disagreements(a, b)
	rows := make([]map[string]any, 0, len(dis))
	for _, d := range dis {
		rows = append(rows, map[string]any{
			"model": d.Model, "glyph": d.Glyph, "scenario": d.Scenario,
			"arm": d.Arm, "seed": d.Seed, "A": d.A, "B": d.B,
		})
	}
	out := filepath.Join(scores, "disagreements.json")
	bs, err := json.MarshalIndent(rows, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, bs, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "A/B over-fire disagreements: %d -> %s\n", len(dis), out)
	return nil
}

// loadRaterA reads the blind rater-A bundles under dir: one <model>.<glyph>/
// subdir per bundle, each holding map.json (rid -> {scenario, arm, seed}) and
// labels.json (rid -> label). It de-anonymizes to keyed labels.
func loadRaterA(dir string) (restaxis.Labels, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := restaxis.Labels{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		model, glyph, ok := strings.Cut(e.Name(), ".")
		if !ok {
			continue
		}
		bundle := filepath.Join(dir, e.Name())
		var meta map[string]struct {
			Scenario int    `json:"scenario"`
			Arm      string `json:"arm"`
			Seed     int    `json:"seed"`
		}
		var labels map[string]string
		if readJSONFile(filepath.Join(bundle, "map.json"), &meta) != nil ||
			readJSONFile(filepath.Join(bundle, "labels.json"), &labels) != nil {
			continue
		}
		for rid, mt := range meta {
			if lab, ok := labels[rid]; ok {
				out[restaxis.Key{Model: model, Glyph: glyph, Scenario: mt.Scenario, Arm: mt.Arm, Seed: mt.Seed}] = lab
			}
		}
	}
	return out, nil
}

// scoreRow is the shared shape of a rater-B / mech score row.
type scoreRow struct {
	Glyph    string `json:"glyph"`
	Scenario int    `json:"scenario"`
	Arm      string `json:"arm"`
	Seed     int    `json:"seed"`
	Label    string `json:"label"`
	MechFlag int    `json:"mech_flag"`
}

// loadRaterB reads scores/<model>.<glyph>.B.json (the model is the first
// dot-segment of the filename, the glyph comes from each row).
func loadRaterB(scoresDir string) (restaxis.Labels, error) {
	files, err := filepath.Glob(filepath.Join(scoresDir, "*.*.B.json"))
	if err != nil {
		return nil, err
	}
	out := restaxis.Labels{}
	for _, f := range files {
		model := strings.SplitN(filepath.Base(f), ".", 2)[0]
		var rows []scoreRow
		if err := readJSONFile(f, &rows); err != nil {
			return nil, err
		}
		for _, r := range rows {
			out[restaxis.Key{Model: model, Glyph: r.Glyph, Scenario: r.Scenario, Arm: r.Arm, Seed: r.Seed}] = r.Label
		}
	}
	return out, nil
}

// loadMechFlags reads scores/<model>.mech.json (the model is the first
// dot-segment of the filename).
func loadMechFlags(scoresDir string) (restaxis.Flags, error) {
	files, err := filepath.Glob(filepath.Join(scoresDir, "*.mech.json"))
	if err != nil {
		return nil, err
	}
	out := restaxis.Flags{}
	for _, f := range files {
		model := strings.SplitN(filepath.Base(f), ".", 2)[0]
		var rows []scoreRow
		if err := readJSONFile(f, &rows); err != nil {
			return nil, err
		}
		for _, r := range rows {
			out[restaxis.Key{Model: model, Glyph: r.Glyph, Scenario: r.Scenario, Arm: r.Arm, Seed: r.Seed}] = r.MechFlag
		}
	}
	return out, nil
}

func readJSONFile(path string, v any) error {
	b, err := os.ReadFile(path) //nolint:gosec // a study score path
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// readRestAxisResponses reads the response texts for one model from the study's
// runs tree: runs/<model>/<glyph>/<scenarioDir>/out/responses/<arm>_<seed>.txt,
// where the scenario dir is named like "s14".
func readRestAxisResponses(study, model string) ([]restaxis.Response, error) {
	root := filepath.Join(study, "runs", model)
	glyphs, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []restaxis.Response
	for _, g := range glyphs {
		if !g.IsDir() {
			continue
		}
		glyph := g.Name()
		scenDirs, err := os.ReadDir(filepath.Join(root, glyph))
		if err != nil {
			continue
		}
		for _, sd := range scenDirs {
			if !sd.IsDir() {
				continue
			}
			scen, err := strconv.Atoi(strings.TrimPrefix(sd.Name(), "s"))
			if err != nil {
				continue
			}
			rdir := filepath.Join(root, glyph, sd.Name(), "out", "responses")
			for _, arm := range restaxis.Arms {
				for seed := 1; seed <= 3; seed++ {
					p := filepath.Join(rdir, fmt.Sprintf("%s_%d.txt", arm, seed))
					b, err := os.ReadFile(p) //nolint:gosec // a study response path
					if err != nil {
						continue
					}
					out = append(out, restaxis.Response{
						Glyph: glyph, Scenario: scen, Arm: arm, Seed: seed, Text: string(b),
					})
				}
			}
		}
	}
	return out, nil
}
