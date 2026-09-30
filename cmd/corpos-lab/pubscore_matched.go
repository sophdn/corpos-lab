package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"corpos-lab/internal/matchedcontent"
)

func init() {
	registerPubScoreMode("matched-build-human-anchor", raMatchedBuildHumanAnchor)
}

// respDirForModel maps the sample's model name to its runs response directory
// under the casg-direct study.
var respDirForModel = map[string]string{
	"mistral": "ground-ext-casg-direct-mistral",
	"qwen":    "ground-ext-casg-direct-qwen38",
}

// scoreCellForModel maps the sample's model name to its blind-score bundle cell.
var scoreCellForModel = map[string]string{
	"mistral": "casg-mistral",
	"qwen":    "casg-qwen",
}

// raMatchedBuildHumanAnchor ports build_human_anchor.py: it reads the casg-direct
// runs and the machine raters' blind-score bundles, draws the stratified sample,
// and writes the blind calibration app plus the operator-held calib_map.json.
//
// The sample is drawn with a seeded shuffle (default seed 7, matching
// random.Random(7)); the draw is structurally identical to python's but not
// byte-identical (different RNG).
//
// Usage: corpos-lab pub-score matched-build-human-anchor --study <matched-content-experiment-dir> \
//
//	--score-dir <dir> --template <file> --out <dir> [--seed <n>]
func raMatchedBuildHumanAnchor(args []string) error {
	study, args := flagValue(args, "--study")
	scoreDir, args := flagValue(args, "--score-dir")
	templatePath, args := flagValue(args, "--template")
	outDir, args := flagValue(args, "--out")
	seedStr, args := flagValue(args, "--seed")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if study == "" || scoreDir == "" || templatePath == "" || outDir == "" {
		return fmt.Errorf("need --study, --score-dir, --template and --out")
	}
	seed := 7
	if seedStr != "" {
		s, err := fmtAtoi(seedStr)
		if err != nil {
			return err
		}
		seed = s
	}

	// Build the per-model code maps from the blind-score bundles.
	codes := map[string]map[matchedcontent.CondRun]string{}
	for model, cell := range scoreCellForModel {
		cm, err := loadCodeMap(filepath.Join(scoreDir, cell))
		if err != nil {
			return fmt.Errorf("code map for %s: %w", cell, err)
		}
		codes[model] = cm
	}

	// Draw the stratified sample.
	rng := rand.New(rand.NewSource(int64(seed))) //nolint:gosec // not cryptographic: a reproducible sampling shuffle
	sample := matchedcontent.BuildSample(rng.Shuffle)

	// Read the scenario task and each selected response.
	scenarioB, err := os.ReadFile(filepath.Join(study, "casg-direct", "materials", "scenario.md")) //nolint:gosec // a study materials path
	if err != nil {
		return err
	}
	task := matchedcontent.Task(string(scenarioB))
	resps := make([]string, len(sample))
	for i, sel := range sample {
		dir := respDirForModel[sel.Model]
		if dir == "" {
			return fmt.Errorf("unknown model %q", sel.Model)
		}
		p := filepath.Join(study, "casg-direct", "runs", dir, "out", "responses", fmt.Sprintf("%s_%d.txt", sel.Condition, sel.Run))
		b, rerr := os.ReadFile(p) //nolint:gosec // a study response path
		if rerr != nil {
			return rerr
		}
		resps[i] = strings.TrimSpace(string(b))
	}

	items, order, cmap, err := matchedcontent.BuildItemsAndMap(sample, task, matchedcontent.GT, resps, codes)
	if err != nil {
		return err
	}

	tplB, err := os.ReadFile(templatePath) //nolint:gosec // an operator-supplied template path
	if err != nil {
		return err
	}
	html := matchedcontent.FillTemplate(string(tplB), matchedcontent.RenderItemsJSON(items))
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return err
	}
	calibPath := filepath.Join(outDir, "calibration.html")
	if err := os.WriteFile(calibPath, []byte(html), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "calib_map.json"), []byte(matchedcontent.RenderCmapJSON(order, cmap)), 0o600); err != nil {
		return err
	}

	fmt.Printf("built %s with %d blind items (%d KB)\n", calibPath, len(items), len(html)/1024)
	fmt.Printf("hidden strata: %s\n", matchedcontent.StrataCounterRepr(order, cmap))
	fmt.Printf("hidden claude-code mix: %s\n", matchedcontent.ClaudeCounterRepr(order, cmap))
	return nil
}

// loadCodeMap reads a blind-score bundle cell (KEY.json, raterA.json, raterB.json,
// optional adjudication.json) and builds its (condition, run) -> code map.
func loadCodeMap(cellDir string) (map[matchedcontent.CondRun]string, error) {
	var rawKey map[string]struct {
		Condition string `json:"condition"`
		Run       int    `json:"run"`
	}
	if err := readJSON(filepath.Join(cellDir, "KEY.json"), &rawKey); err != nil {
		return nil, err
	}
	key := make(map[string]matchedcontent.CondRun, len(rawKey))
	for rid, v := range rawKey {
		key[rid] = matchedcontent.CondRun{Condition: v.Condition, Run: v.Run}
	}
	raterA, err := loadRaterEntries(filepath.Join(cellDir, "raterA.json"))
	if err != nil {
		return nil, err
	}
	raterB, err := loadRaterEntries(filepath.Join(cellDir, "raterB.json"))
	if err != nil {
		return nil, err
	}
	adj := map[string]string{}
	adjPath := filepath.Join(cellDir, "adjudication.json")
	if _, statErr := os.Stat(adjPath); statErr == nil {
		if err := readJSON(adjPath, &adj); err != nil {
			return nil, err
		}
	}
	return matchedcontent.BuildCodeMap(key, raterA, raterB, adj)
}

// loadRaterEntries reads a rater bundle (a list of {id, code}) into RaterEntry slice.
func loadRaterEntries(path string) ([]matchedcontent.RaterEntry, error) {
	var raw []struct {
		ID   string `json:"id"`
		Code string `json:"code"`
	}
	if err := readJSON(path, &raw); err != nil {
		return nil, err
	}
	out := make([]matchedcontent.RaterEntry, len(raw))
	for i, r := range raw {
		out[i] = matchedcontent.RaterEntry{ID: r.ID, Code: r.Code}
	}
	return out, nil
}

// fmtAtoi parses an int, wrapping strconv for a clearer error.
func fmtAtoi(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, fmt.Errorf("seed must be an integer: %w", err)
	}
	return n, nil
}
