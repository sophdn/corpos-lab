package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("score-phi4", raScorePhi4) }

// raScorePhi4 is rater B: it grades each response with the phi-4 model over the
// llama.cpp portal and writes scores/<model>.<glyph>.B.json. Ported from
// score_phi4.py. The study dir and portal are flags (the python hardcoded them);
// the sampler is pinned exactly as the python sent it (temperature 0, top_k 1).
//
//	rest-axis score-phi4 --study <dir> --model <m> [--glyph <g>] [--scenario <n>] [--portal URL]
func raScorePhi4(args []string) error {
	study, args := flagValue(args, "--study")
	model, args := flagValue(args, "--model")
	glyph, args := flagValue(args, "--glyph")
	scenStr, args := flagValue(args, "--scenario")
	portal, args := flagValue(args, "--portal")
	_ = args
	if study == "" || model == "" {
		return fmt.Errorf("need --study <dir> and --model <name>")
	}
	if portal == "" {
		portal = "http://localhost:8081/completion"
	}
	var onlyScen int
	if scenStr != "" {
		n, err := strconv.Atoi(scenStr)
		if err != nil {
			return fmt.Errorf("--scenario: %w", err)
		}
		onlyScen = n
	}

	glyphs := []string{glyph}
	if glyph == "" {
		gs, err := listSubdirs(filepath.Join(study, "runs", model))
		if err != nil {
			return err
		}
		glyphs = gs
	}

	sc := phi4Scorer{client: &http.Client{Timeout: 130 * time.Second}, portal: portal, study: study, model: model}
	for _, g := range glyphs {
		results, err := sc.scoreGlyph(g, onlyScen)
		if err != nil {
			return err
		}
		outp := filepath.Join(study, "scores", fmt.Sprintf("%s.%s.B.json", model, g))
		if onlyScen != 0 {
			results = keepOtherScenarios(outp, onlyScen, results)
		}
		if err := os.MkdirAll(filepath.Dir(outp), 0o750); err != nil {
			return err
		}
		b, err := json.MarshalIndent(results, "", " ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(outp, b, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "%s/%s: scored %d -> %s\n", model, g, len(results), filepath.Base(outp))
	}
	return nil
}

// phi4Scorer grades one model's responses in a study through the portal.
type phi4Scorer struct {
	client               *http.Client
	portal, study, model string
}

// scoreGlyph grades every scenario of glyph g (only onlyScen when non-zero, else
// all scenario dirs in numeric order).
func (sc phi4Scorer) scoreGlyph(g string, onlyScen int) ([]map[string]any, error) {
	gtMD, err := os.ReadFile(filepath.Join(sc.study, g, "CORRECT_COMPLETIONS.md")) //nolint:gosec // study path
	if err != nil {
		return nil, err
	}
	gt := restaxis.ParseCorrectCompletions(string(gtMD))

	scenDirs, err := listSubdirs(filepath.Join(sc.study, "runs", sc.model, g))
	if err != nil {
		return nil, err
	}
	if onlyScen != 0 {
		scenDirs = []string{fmt.Sprintf("s%d", onlyScen)}
	} else {
		sort.Slice(scenDirs, func(i, j int) bool { return scenNum(scenDirs[i]) < scenNum(scenDirs[j]) })
	}

	var results []map[string]any
	for _, sd := range scenDirs {
		snum := scenNum(sd)
		scenTxt, err := os.ReadFile(filepath.Join(sc.study, g, "materials", fmt.Sprintf("scenario_%d.md", snum))) //nolint:gosec // study path
		if err != nil {
			continue // no material for this scenario: skip it, as score_phi4.py did
		}
		rows, err := sc.scoreScenario(g, sd, snum, string(scenTxt), gt[snum])
		if err != nil {
			return nil, err
		}
		results = append(results, rows...)
	}
	return results, nil
}

// scoreScenario grades each arm x seed response of scenario dir sd (number snum,
// task text scenTxt, ground truth gt). A missing response file is skipped.
func (sc phi4Scorer) scoreScenario(g, sd string, snum int, scenTxt, gt string) ([]map[string]any, error) {
	rdir := filepath.Join(sc.study, "runs", sc.model, g, sd, "out", "responses")
	var rows []map[string]any
	for _, arm := range restaxis.Arms {
		for seed := 1; seed <= 3; seed++ {
			resp, err := os.ReadFile(filepath.Join(rdir, fmt.Sprintf("%s_%d.txt", arm, seed))) //nolint:gosec // study path
			if err != nil {
				continue
			}
			prompt := restaxis.BuildPhi4Prompt(scenTxt, gt, string(resp))
			content, err := phi4Complete(sc.client, sc.portal, prompt)
			if err != nil {
				return nil, fmt.Errorf("%s/%s/%s %s_%d: %w", sc.model, g, sd, arm, seed, err)
			}
			label, reason, coerced := restaxis.ParsePhi4Label(content)
			rows = append(rows, map[string]any{
				"glyph": g, "scenario": snum, "arm": arm, "seed": seed,
				"label": label, "reason": reason, "coerced": coerced,
			})
		}
	}
	return rows, nil
}

// keepOtherScenarios merges a one-scenario re-score into the existing score file:
// the file's rows for every other scenario come first, then the new rows. A
// missing or unparseable file leaves the new rows as they are.
func keepOtherScenarios(outp string, onlyScen int, results []map[string]any) []map[string]any {
	prev, err := os.ReadFile(outp) //nolint:gosec // study path
	if err != nil {
		return results
	}
	var old []map[string]any
	if json.Unmarshal(prev, &old) != nil {
		return results
	}
	kept := old[:0]
	for _, r := range old {
		if int(toInt(r["scenario"])) != onlyScen {
			kept = append(kept, r)
		}
	}
	return append(kept, results...)
}

// phi4Complete posts one grader prompt to the llama.cpp portal with the sampler
// pinned exactly as score_phi4.py sent it, and returns the trimmed content.
func phi4Complete(client *http.Client, portal, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": prompt, "temperature": 0.0, "n_predict": 120, "top_k": 1, "cache_prompt": false,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, portal, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("portal %s: %s", portal, resp.Status)
	}
	var out struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Content), nil
}

func listSubdirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func scenNum(sd string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(sd, "s"))
	return n
}

func toInt(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	default:
		return 0
	}
}
