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

	client := &http.Client{Timeout: 130 * time.Second}
	for _, g := range glyphs {
		gtMD, err := os.ReadFile(filepath.Join(study, g, "CORRECT_COMPLETIONS.md")) //nolint:gosec // study path
		if err != nil {
			return err
		}
		gt := restaxis.ParseCorrectCompletions(string(gtMD))

		scenDirs, err := listSubdirs(filepath.Join(study, "runs", model, g))
		if err != nil {
			return err
		}
		if onlyScen != 0 {
			scenDirs = []string{fmt.Sprintf("s%d", onlyScen)}
		} else {
			sort.Slice(scenDirs, func(i, j int) bool { return scenNum(scenDirs[i]) < scenNum(scenDirs[j]) })
		}

		var results []map[string]any
		for _, sd := range scenDirs {
			snum := scenNum(sd)
			scenTxt, err := os.ReadFile(filepath.Join(study, g, "materials", fmt.Sprintf("scenario_%d.md", snum))) //nolint:gosec // study path
			if err != nil {
				continue
			}
			rdir := filepath.Join(study, "runs", model, g, sd, "out", "responses")
			for _, arm := range restaxis.Arms {
				for seed := 1; seed <= 3; seed++ {
					resp, err := os.ReadFile(filepath.Join(rdir, fmt.Sprintf("%s_%d.txt", arm, seed))) //nolint:gosec // study path
					if err != nil {
						continue
					}
					prompt := restaxis.BuildPhi4Prompt(string(scenTxt), gt[snum], string(resp))
					content, err := phi4Complete(client, portal, prompt)
					if err != nil {
						return fmt.Errorf("%s/%s/%s %s_%d: %w", model, g, sd, arm, seed, err)
					}
					label, reason, coerced := restaxis.ParsePhi4Label(content)
					results = append(results, map[string]any{
						"glyph": g, "scenario": snum, "arm": arm, "seed": seed,
						"label": label, "reason": reason, "coerced": coerced,
					})
				}
			}
		}

		outp := filepath.Join(study, "scores", fmt.Sprintf("%s.%s.B.json", model, g))
		if onlyScen != 0 {
			if prev, err := os.ReadFile(outp); err == nil { //nolint:gosec // study path
				var old []map[string]any
				if json.Unmarshal(prev, &old) == nil {
					kept := old[:0]
					for _, r := range old {
						if int(toInt(r["scenario"])) != onlyScen {
							kept = append(kept, r)
						}
					}
					results = append(kept, results...)
				}
			}
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
