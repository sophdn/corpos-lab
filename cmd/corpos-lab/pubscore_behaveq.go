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

	"corpos-lab/internal/behaveq"
)

// pubscore_behaveq.go wires the behavioral-equivalence-assay python ports
// (chain eradicate-python-rewrite-in-go, paper "Duty or Corpus",
// 10.5281/zenodo.22716214) to the pub-score registry. Pure logic lives in
// internal/behaveq; this file is the thin IO seam.

func init() {
	registerPubScoreMode("behaveq-gen-expansion-defs", raBehaveqGenExpansionDefs)
	registerPubScoreMode("behaveq-analyze", raBehaveqAnalyze)
	registerPubScoreMode("behaveq-double-score", raBehaveqDoubleScore)
	registerPubScoreMode("behaveq-dp1-rule-scorer", raBehaveqDP1RuleScorer)
}

// raBehaveqGenExpansionDefs writes the three arm study-def TOMLs for a given n,
// mirroring gen_expansion_defs.py. Files land at
// <outdir>/study-defs/n<N>/<arm>.toml (--outdir defaults to the cwd).
//
//	pub-score behaveq-gen-expansion-defs <N> [--outdir DIR]
func raBehaveqGenExpansionDefs(args []string) error {
	outdir, rest := flagValue(args, "--outdir")
	if outdir == "" {
		outdir = "."
	}
	if len(rest) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score behaveq-gen-expansion-defs <N> [--outdir DIR]")
	}
	n, err := strconv.Atoi(rest[0])
	if err != nil {
		return fmt.Errorf("N: %w", err)
	}
	dir := filepath.Join(outdir, "study-defs", fmt.Sprintf("n%d", n))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	for _, arm := range behaveq.ExpansionArms {
		path := filepath.Join(dir, arm.Key+".toml")
		if err := os.WriteFile(path, []byte(behaveq.TOMLFor(arm, n)), 0o600); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", path)
	}
	return nil
}

// raBehaveqAnalyze reproduces analyze.py: it scores every response under the study
// and prints the clear-rate + contrast report to stdout.
//
//	pub-score behaveq-analyze --study <dir> [<suffix>]   (suffix defaults to -n24)
func raBehaveqAnalyze(args []string) error {
	study, rest := flagValue(args, "--study")
	if study == "" {
		return fmt.Errorf("need --study <dir>")
	}
	suffix := "-n24"
	if len(rest) > 0 {
		suffix = rest[0]
	}
	data := map[string]behaveq.ModelData{}
	for _, model := range behaveq.AnalyzeModels {
		runDir := filepath.Join(study, "runs", model+suffix)
		info, err := os.Stat(runDir)
		if err != nil || !info.IsDir() {
			data[model] = behaveq.ModelData{Present: false}
			continue
		}
		md := behaveq.ModelData{Present: true, Cells: map[string]behaveq.CellScore{}}
		respDir := filepath.Join(runDir, "out", "responses")
		for _, cond := range behaveq.AnalyzeConds {
			texts, err := readCondResponses(respDir, cond)
			if err != nil {
				return err
			}
			md.Cells[cond] = behaveq.ScoreCell(texts)
		}
		data[model] = md
	}
	fmt.Print(behaveq.Report(suffix, data))
	return nil
}

// readCondResponses reads all "<cond>_*.txt" response files in a directory, in the
// same lexical order analyze.py's sorted() used (order does not affect the tally).
func readCondResponses(dir, cond string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, cond+"_") && strings.HasSuffix(n, ".txt") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	texts := make([]string, 0, len(names))
	for _, n := range names {
		b, err := os.ReadFile(filepath.Join(dir, n)) //nolint:gosec // a study response path
		if err != nil {
			return nil, err
		}
		texts = append(texts, string(b))
	}
	return texts, nil
}

// raBehaveqDP1RuleScorer reproduces dp1_rule_scorer.py's two CLI modes: --validate
// scores the hand-scored sets and prints agreement; a single path prints the
// response's score.
//
//	pub-score behaveq-dp1-rule-scorer --validate --study <dir>
//	pub-score behaveq-dp1-rule-scorer <response.txt>
func raBehaveqDP1RuleScorer(args []string) error {
	study, rest := flagValue(args, "--study")
	validate := false
	var files []string
	for _, a := range rest {
		if a == "--validate" {
			validate = true
			continue
		}
		files = append(files, a)
	}
	if validate {
		if study == "" {
			return fmt.Errorf("--validate needs --study <dir>")
		}
		res, err := behaveq.Validate(func(model, dirSuffix, cond string, run int) (string, error) {
			p := filepath.Join(study, "runs", model+dirSuffix, "out", "responses", fmt.Sprintf("%s_%d.txt", cond, run))
			b, err := os.ReadFile(p) //nolint:gosec // a study response path
			return string(b), err
		})
		if err != nil {
			return err
		}
		fmt.Printf("agreement: %d/%d\n", res.Agree, res.Total)
		if len(res.Mismatches) > 0 {
			fmt.Println("mismatches (model condition run: hand -> rule):")
			for _, m := range res.Mismatches {
				fmt.Printf("  %s %s %d: %s -> %s\n", m.Model, m.Condition, m.Run, m.Hand, m.Got)
			}
		} else {
			fmt.Println("no mismatches: the rule reproduces every hand-score.")
		}
		return nil
	}
	if len(files) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score behaveq-dp1-rule-scorer <response.txt> | --validate --study <dir>")
	}
	b, err := os.ReadFile(files[0]) //nolint:gosec // a response path
	if err != nil {
		return err
	}
	fmt.Println(behaveq.Score(string(b)))
	return nil
}

// raBehaveqDoubleScore reproduces double_score.py: it re-scores the fixed sample of
// responses with the phi-4 rater over the llama.cpp portal and writes the results
// as JSON to stdout. The model call is a thin IO seam; the prompt-building and
// parsing are the pure, unit-tested behaveq functions. NOTE: this issues a live
// model call — it is not exercised by the byte-identical verification, which covers
// only the deterministic tools.
//
//	pub-score behaveq-double-score --study <dir> [--portal URL]
func raBehaveqDoubleScore(args []string) error {
	study, rest := flagValue(args, "--study")
	portal, rest := flagValue(rest, "--portal")
	_ = rest
	if study == "" {
		return fmt.Errorf("need --study <dir>")
	}
	if portal == "" {
		portal = "http://localhost:8081/completion"
	}
	client := &http.Client{Timeout: 130 * time.Second}
	type result struct {
		Model     string `json:"model"`
		Condition string `json:"condition"`
		Run       int    `json:"run"`
		DP1       string `json:"dp1"`
		DP2       string `json:"dp2"`
		Raw       string `json:"raw"`
	}
	var results []result
	for _, s := range behaveq.DoubleScoreSample {
		path := filepath.Join(study, "runs", s.ModelDir, "out", "responses", fmt.Sprintf("%s_%d.txt", s.Condition, s.Run))
		b, err := os.ReadFile(path) //nolint:gosec // a study response path
		if err != nil {
			return err
		}
		prompt := behaveq.BuildDoubleScorePrompt(string(b))
		raw, err := doubleScoreRate(client, portal, prompt)
		if err != nil {
			return fmt.Errorf("%s %s %d: %w", s.ModelDir, s.Condition, s.Run, err)
		}
		dp1, dp2 := behaveq.ParseDoubleScore(raw)
		results = append(results, result{
			Model: s.ModelDir, Condition: s.Condition, Run: s.Run,
			DP1: dp1, DP2: dp2, Raw: strings.TrimSpace(raw),
		})
		fmt.Fprintf(os.Stderr, "%s %s %d: DP1=%s DP2=%s\n", s.ModelDir, s.Condition, s.Run, dp1, dp2)
	}
	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// doubleScoreRate posts one rater prompt to the llama.cpp portal with the sampler
// pinned exactly as double_score.rate sent it, and returns the raw content.
func doubleScoreRate(client *http.Client, portal, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": prompt, "temperature": 0.0, "n_predict": 24,
		"top_k": 0, "top_p": 1.0, "min_p": 0.05, "typical_p": 1.0,
		"repeat_penalty": 1.0, "seed": 0,
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
	return out.Content, nil
}
