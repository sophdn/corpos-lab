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
	"strings"
	"time"

	"corpos-lab/internal/setupcompletion"
)

// The setup-completion-vs-agentic-loop provenance tools, ported from
// provenance/published-scoring/setup-completion-vs-agentic-loop/*.py. Each mode
// registers itself here; the pure logic lives in internal/setupcompletion.
func init() {
	registerPubScoreMode("setup-turn1-toolcall", raSetupTurn1Toolcall)
	registerPubScoreMode("setup-turn2-continuation", raSetupTurn2Continuation)
	registerPubScoreMode("setup-build-slices", raSetupBuildSlices)
	registerPubScoreMode("setup-gen-studies", raSetupGenStudies)
	registerPubScoreMode("setup-tally", raSetupTally)
}

// raSetupTurn1Toolcall runs the turn-1 tool-call feasibility smoke: it builds the
// prompt (pure), posts one completion to the llama.cpp portal (the IO seam), and
// prints the content + timing report. --print-prompt emits the prompt and exits
// without calling the model.
//
//	pub-score setup-turn1-toolcall [--portal URL] [--print-prompt]
func raSetupTurn1Toolcall(args []string) error {
	return runSetupSmoke(args, setupcompletion.Turn1Prompt(), setupcompletion.Turn1Params,
		setupcompletion.FormatTurn1Report)
}

// raSetupTurn2Continuation runs the turn-2 continuation smoke, same shape as
// turn-1 with the continuation prompt and sampler.
//
//	pub-score setup-turn2-continuation [--portal URL] [--print-prompt]
func raSetupTurn2Continuation(args []string) error {
	return runSetupSmoke(args, setupcompletion.Turn2Prompt(), setupcompletion.Turn2Params,
		setupcompletion.FormatTurn2Report)
}

func runSetupSmoke(args []string, prompt string, params setupcompletion.CompletionParams,
	format func(setupcompletion.CompletionResult) string) error {
	portal, args := flagValue(args, "--portal")
	printPrompt := hasFlag(args, "--print-prompt")
	if printPrompt {
		fmt.Print(prompt)
		return nil
	}
	if portal == "" {
		portal = "http://localhost:8081/completion"
	}
	raw, err := setupComplete(portal, prompt, params)
	if err != nil {
		return err
	}
	res, err := setupcompletion.ParseCompletionResponse(raw)
	if err != nil {
		return err
	}
	fmt.Print(format(res))
	return nil
}

// setupComplete posts one prompt+sampler to the llama.cpp /completion portal and
// returns the raw response body. This is the model-call IO seam.
func setupComplete(portal, prompt string, p setupcompletion.CompletionParams) ([]byte, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": prompt, "n_predict": p.NPredict, "temperature": p.Temperature,
		"seed": p.Seed, "top_k": p.TopK, "top_p": p.TopP, "min_p": p.MinP,
		"repeat_penalty": p.RepeatPenalty, "cache_prompt": p.CachePrompt,
	})
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodPost, portal, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("portal %s: %s", portal, resp.Status)
	}
	return raw, nil
}

// raSetupBuildSlices builds per-glyph blind slices and a de-anonymising key from
// the study runs tree, ported from build_slices.py. The FS walk is the IO seam;
// hashing and slice/key assembly are pure (setupcompletion.BuildSlices).
//
//	pub-score setup-build-slices --study-root <dir> --out <dir> [--glyph <name> ...]
func raSetupBuildSlices(args []string) error {
	studyRoot, args := flagValue(args, "--study-root")
	out, args := flagValue(args, "--out")
	var only []string
	for {
		var g string
		g, args = flagValue(args, "--glyph")
		if g == "" {
			break
		}
		only = append(only, g)
	}
	if studyRoot == "" || out == "" {
		return fmt.Errorf("need --study-root <dir> and --out <dir>")
	}

	records, err := collectSetupRuns(studyRoot)
	if err != nil {
		return err
	}
	if len(only) > 0 {
		keep := records[:0]
		for _, r := range records {
			if containsStr(only, r.Glyph) {
				keep = append(keep, r)
			}
		}
		records = keep
	}

	glyphs, perGlyph, key := setupcompletion.BuildSlices(records)
	if err := os.MkdirAll(filepath.Join(out, "slices"), 0o750); err != nil {
		return err
	}
	for _, g := range glyphs {
		items := perGlyph[g]
		rows := make([]map[string]any, 0, len(items))
		for _, it := range items {
			rows = append(rows, map[string]any{"id": it.ID, "text": it.Text})
		}
		slicePath := filepath.Join(out, "slices", g+".jsonl")
		if err := writeJSONL(slicePath, rows); err != nil {
			return err
		}
		fmt.Printf("%s: %d responses -> %s\n", g, len(items), slicePath)
	}
	keyPath := filepath.Join(out, "key.json")
	if err := writeJSON(keyPath, key); err != nil {
		return err
	}
	fmt.Printf("key: %d ids -> %s\n", len(key), keyPath)
	return nil
}

// collectSetupRuns walks the study runs tree and yields one RunRecord per response
// text, matching build_slices.collect: it globs */runs/setup-*-qwen38/out/responses,
// skips smoke studies, derives the setup from the study name, and reads each
// <cond>_<run>.txt.
func collectSetupRuns(studyRoot string) ([]setupcompletion.RunRecord, error) {
	respDirs, err := filepath.Glob(filepath.Join(studyRoot, "*", "runs", "setup-*-qwen38", "out", "responses"))
	if err != nil {
		return nil, err
	}
	sort.Strings(respDirs)
	var out []setupcompletion.RunRecord
	for _, respDir := range respDirs {
		// respDir = <root>/<glyph>/runs/<study>/out/responses
		study := filepath.Base(filepath.Dir(filepath.Dir(respDir)))
		glyph := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(respDir)))))
		if setupcompletion.IsSmokeStudy(study) {
			continue
		}
		setup, ok := setupcompletion.SetupFromStudy(study)
		if !ok {
			continue
		}
		for _, cond := range setupcompletion.Conditions {
			matches, err := filepath.Glob(filepath.Join(respDir, cond+"_*.txt"))
			if err != nil {
				return nil, err
			}
			sort.Strings(matches)
			for _, txt := range matches {
				stem := strings.TrimSuffix(filepath.Base(txt), ".txt")
				run, err := setupcompletion.RunFromStem(stem)
				if err != nil {
					return nil, err
				}
				data, err := os.ReadFile(txt) //nolint:gosec // a study runs-tree path
				if err != nil {
					return nil, err
				}
				out = append(out, setupcompletion.RunRecord{
					Glyph: glyph, Setup: setup, Condition: cond, Run: run, Text: string(data),
				})
			}
		}
	}
	return out, nil
}

// raSetupGenStudies writes each glyph's study.raw and study.loop TOMLs under the
// study root, ported from gen_studies.py. The templates are pure; only the
// mkdir/write is IO.
//
//	pub-score setup-gen-studies --root <study-root>
func raSetupGenStudies(args []string) error {
	root, _ := flagValue(args, "--root")
	if root == "" {
		return fmt.Errorf("need --root <study-root>")
	}
	for _, g := range setupcompletion.Glyphs {
		d := filepath.Join(root, g)
		if err := os.MkdirAll(d, 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(d, "study.raw.qwen38.toml"), []byte(setupcompletion.RawTOML(g)), 0o600); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(d, "study.loop.qwen38.toml"), []byte(setupcompletion.LoopTOML(g)), 0o600); err != nil {
			return err
		}
		fmt.Printf("wrote %s/study.raw.qwen38.toml and study.loop.qwen38.toml\n", g)
	}
	return nil
}

// raSetupTally joins the key and two rater outputs and prints the per-glyph
// consensus table, ported from tally.py. Reading the three JSON files is the IO
// seam; the join and formatting are pure (setupcompletion.TallyReport).
//
//	pub-score setup-tally --key <key.json> --rater-a <a.json> --rater-b <b.json>
func raSetupTally(args []string) error {
	keyPath, args := flagValue(args, "--key")
	aPath, args := flagValue(args, "--rater-a")
	bPath, _ := flagValue(args, "--rater-b")
	if keyPath == "" || aPath == "" || bPath == "" {
		return fmt.Errorf("need --key, --rater-a, and --rater-b")
	}
	var key map[string]setupcompletion.KeyEntry
	if err := readJSON(keyPath, &key); err != nil {
		return err
	}
	var a, b map[string]string
	if err := readJSON(aPath, &a); err != nil {
		return err
	}
	if err := readJSON(bPath, &b); err != nil {
		return err
	}
	fmt.Print(setupcompletion.TallyReport(key, a, b))
	return nil
}

func hasFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

func containsStr(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
