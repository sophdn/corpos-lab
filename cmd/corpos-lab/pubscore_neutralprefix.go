package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"corpos-lab/internal/neutralprefix"
)

// The neutral-prefix-control provenance scoring tools, ported from
// provenance/published-scoring/neutral-prefix-control/*.py (chain
// eradicate-python-rewrite-in-go). Pure logic lives in internal/neutralprefix;
// these handlers are the IO seams. Each registers its pub-score mode from init().
func init() {
	registerPubScoreMode("neutral-build-anchor", npBuildAnchor)
	registerPubScoreMode("neutral-reconcile-rerate", npReconcileRerate)
	registerPubScoreMode("neutral-aggregate", npAggregate)
	registerPubScoreMode("neutral-build-slices", npBuildSlices)
	registerPubScoreMode("neutral-score-phi4", npScorePhi4)
}

// npBuildAnchor ports build_anchor.py: build the blind human-anchor item set and
// its held key from the casg-direct scenario-1 responses. The python derived every
// path from __file__; here --study is the neutral-prefix-control dir and --out
// defaults to <study>/human-anchor.
//
//	pub-score neutral-build-anchor --study <dir> [--out <dir>]
func npBuildAnchor(args []string) error {
	study, args := flagValue(args, "--study")
	out, args := flagValue(args, "--out")
	_ = args
	if study == "" {
		return fmt.Errorf("need --study <neutral-prefix-control dir>")
	}
	if out == "" {
		out = filepath.Join(study, "human-anchor")
	}

	var keyMeta map[string]neutralprefix.KeyMeta
	if err := readJSON(filepath.Join(study, "scoring", "key.json"), &keyMeta); err != nil {
		return err
	}
	claude := map[string]string{}
	for _, half := range []string{"casg-direct__A.json", "casg-direct__B.json"} {
		var m map[string]string
		if err := readJSON(filepath.Join(study, "scoring", "scores", "claude", half), &m); err != nil {
			return err
		}
		for k, v := range m {
			claude[k] = v
		}
	}
	scenB, err := os.ReadFile(filepath.Join(study, "casg-direct", "materials", "scenario_1.md")) //nolint:gosec // study path
	if err != nil {
		return err
	}

	respText := func(model, cond string, run int) (string, error) {
		p := filepath.Join(study, "casg-direct", "runs",
			fmt.Sprintf("npc-casg-direct-s1-%s", model), "out", "responses",
			fmt.Sprintf("%s_%d.txt", cond, run))
		b, err := os.ReadFile(p) //nolint:gosec // study path
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	res, err := neutralprefix.BuildAnchor(keyMeta, claude, string(scenB), respText)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "items.json"),
		[]byte(neutralprefix.PyDumps(res.Items, 1, false)), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "key.json"),
		[]byte(neutralprefix.PyDumps(res.Held, 1, true)), 0o600); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "built %d blind items -> %s\n", res.Held.Len(), out)
	fmt.Fprintf(os.Stderr, "hidden strata: %s\n", res.Strata.PyDictRepr())
	fmt.Fprintf(os.Stderr, "hidden Claude-code mix: %s\n", res.ClaudeMix.PyDictRepr())
	return nil
}

// npReconcileRerate ports reconcile_rerate.py: report the tightened-rubric blind
// re-rate against the old single-rater codes for the three remaining classes.
// --study is the neutral-prefix-control dir; --anchor (default <study>/human-anchor)
// holds the retrated_*_rater{A,B}.json files. Output is printed to stdout.
//
//	pub-score neutral-reconcile-rerate --study <dir> [--anchor <dir>]
func npReconcileRerate(args []string) error {
	study, args := flagValue(args, "--study")
	anchor, args := flagValue(args, "--anchor")
	_ = args
	if study == "" {
		return fmt.Errorf("need --study <neutral-prefix-control dir>")
	}
	if anchor == "" {
		anchor = filepath.Join(study, "human-anchor")
	}

	keyB, err := os.ReadFile(filepath.Join(study, "scoring", "key.json")) //nolint:gosec // study path
	if err != nil {
		return err
	}
	key, err := neutralprefix.ParseKeyFile(keyB)
	if err != nil {
		return err
	}

	names := []string{"parent-state-check-bypass", "formal-step-context-bypass", "conditional-gate-uniform-default"}
	var classes []neutralprefix.ReconcileClass
	for _, name := range names {
		var ra, rb map[string]string
		if err := readJSON(filepath.Join(anchor, "retrated_"+name+"_raterA.json"), &ra); err != nil {
			return err
		}
		if err := readJSON(filepath.Join(anchor, "retrated_"+name+"_raterB.json"), &rb); err != nil {
			return err
		}
		old, err := loadClaudeUnion(study, name)
		if err != nil {
			return err
		}
		classes = append(classes, neutralprefix.ReconcileClass{Name: name, RaterA: ra, RaterB: rb, Old: old})
	}

	report, err := neutralprefix.Reconcile(key, classes)
	if err != nil {
		return err
	}
	fmt.Print(report)
	return nil
}

// npAggregate ports aggregate.py: pooled and per-cell C-rates for the grid, with
// rater agreement and strict-consensus when two raters are given. --rater-a and
// --rater-b are dirs (merge *.json) or single files; --key is scoring/key.json.
// Output is printed to stdout.
//
//	pub-score neutral-aggregate --key <key.json> --rater-a <path> [--rater-b <path>]
func npAggregate(args []string) error {
	keyPath, args := flagValue(args, "--key")
	raterA, args := flagValue(args, "--rater-a")
	raterB, args := flagValue(args, "--rater-b")
	_ = args
	if keyPath == "" || raterA == "" {
		return fmt.Errorf("need --key <key.json> and --rater-a <dir|file>")
	}

	keyB, err := os.ReadFile(keyPath) //nolint:gosec // study path
	if err != nil {
		return err
	}
	key, err := neutralprefix.ParseKeyFile(keyB)
	if err != nil {
		return err
	}
	a, err := loadScores(raterA)
	if err != nil {
		return err
	}
	var b map[string]string
	hasB := raterB != ""
	if hasB {
		if b, err = loadScores(raterB); err != nil {
			return err
		}
	}
	fmt.Print(neutralprefix.Aggregate(key, a, b, hasB))
	return nil
}

var npStatusRe = regexp.MustCompile(`"status"\s*:\s*"completed"`)

// npBuildSlices ports build_slices.py: scan <root>/*/runs/npc-*/ for completed
// cells, pair each result row with its response text, and write condition-blind
// slices, the un-blinding key, and a coverage manifest under <root>/scoring/.
//
//	pub-score neutral-build-slices [--root <dir>] [--per-slice 32] [--seed 12345] [--prefix npc]
//
// --prefix is the run-dir name prefix to scan (default "npc" for the chain-548
// neutral-prefix-control study). The length-distraction-typing study reuses this
// builder with --prefix ldt --root studies/length-distraction-typing.
func npBuildSlices(args []string) error {
	root, args := flagValue(args, "--root")
	perSliceStr, args := flagValue(args, "--per-slice")
	seedStr, args := flagValue(args, "--seed")
	prefix, args := flagValue(args, "--prefix")
	_ = args
	if root == "" {
		root = "studies/neutral-prefix-control"
	}
	if prefix == "" {
		prefix = "npc"
	}
	perSlice, err := intOr(perSliceStr, 32)
	if err != nil {
		return fmt.Errorf("--per-slice: %w", err)
	}
	seed, err := intOr(seedStr, 12345)
	if err != nil {
		return fmt.Errorf("--seed: %w", err)
	}

	records, err := npScanRecords(root, prefix)
	if err != nil {
		return err
	}
	// The chain-548 study (prefix npc) keeps its pinned class order so its committed
	// slices reproduce byte-identically; any other study emits slices for the classes
	// its records contain, in first-seen (sorted-path) order.
	var res *neutralprefix.SlicesResult
	if prefix == "npc" {
		res = neutralprefix.BuildSlices(records, perSlice, seed)
	} else {
		res = neutralprefix.BuildSlicesFor(records, perSlice, seed, firstSeenClasses(records))
	}

	outDir := filepath.Join(root, "scoring")
	slicesDir := filepath.Join(outDir, "slices")
	if err := os.MkdirAll(slicesDir, 0o750); err != nil {
		return err
	}
	for _, sf := range res.Slices {
		body := strings.Join(sf.Lines, "\n")
		if len(sf.Lines) > 0 {
			body += "\n"
		}
		if err := os.WriteFile(filepath.Join(slicesDir, sf.Name), []byte(body), 0o600); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(outDir, "key.json"),
		[]byte(neutralprefix.PyDumps(res.Key, 0, true)), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outDir, "MANIFEST.json"),
		[]byte(neutralprefix.PyDumps(res.Manifest, 1, true)), 0o600); err != nil {
		return err
	}
	fmt.Print(res.Summary)
	return nil
}

// firstSeenClasses returns the distinct class names in record order, so a study's
// slice-emission order is deterministic (records arrive in sorted-path order).
func firstSeenClasses(records []neutralprefix.SliceRecord) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range records {
		if !seen[r.Cls] {
			seen[r.Cls] = true
			out = append(out, r.Cls)
		}
	}
	return out
}

// npScanRecords walks the study runs tree in build_slices.py's exact order (sorted
// results.json paths, then row order), skipping cells whose run-record is not
// "completed" and rows whose response file is missing. prefix selects the run-dir
// name prefix to match ("npc" for chain 548, "ldt" for length-distraction-typing).
func npScanRecords(root, prefix string) ([]neutralprefix.SliceRecord, error) {
	runDirRe := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-(.+)-s(\d+)-([^-]+)$`)
	matches, err := filepath.Glob(filepath.Join(root, "*", "runs", prefix+"-*", "out", "results.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var records []neutralprefix.SliceRecord
	for _, results := range matches {
		runDir := filepath.Dir(filepath.Dir(results))
		m := runDirRe.FindStringSubmatch(filepath.Base(runDir))
		if m == nil {
			continue
		}
		cls, sc := m[1], m[2]
		rec := filepath.Join(runDir, "run-record.json")
		recB, err := os.ReadFile(rec) //nolint:gosec // study path
		if err != nil || !npStatusRe.Match(recB) {
			continue
		}
		var res struct {
			Rows []struct {
				Condition string `json:"condition"`
				Run       int    `json:"run"`
			} `json:"rows"`
		}
		if err := readJSON(results, &res); err != nil {
			return nil, err
		}
		for _, r := range res.Rows {
			resp := filepath.Join(runDir, "out", "responses", fmt.Sprintf("%s_%d.txt", r.Condition, r.Run))
			text, err := os.ReadFile(resp) //nolint:gosec // study path
			if err != nil {
				continue
			}
			records = append(records, neutralprefix.SliceRecord{
				Cls: cls, Scenario: sc, Model: m[3], Condition: r.Condition, Run: r.Run, Text: string(text),
			})
		}
	}
	return records, nil
}

// npScorePhi4 ports score_phi4.py: the phi-4 blind rater. It reads a slice JSONL,
// grades each response against its class rubric over the llama.cpp portal, and
// writes {id: code} JSON. The prompt build and label parse are pure
// (internal/neutralprefix); the model call is the IO seam here.
//
//	pub-score neutral-score-phi4 --in <slice.jsonl> --out <codes.json> [--rubric-dir <dir>] [--portal URL]
func npScorePhi4(args []string) error {
	inp, args := flagValue(args, "--in")
	outp, args := flagValue(args, "--out")
	rubricDir, args := flagValue(args, "--rubric-dir")
	portal, args := flagValue(args, "--portal")
	_ = args
	if inp == "" || outp == "" {
		return fmt.Errorf("need --in <slice.jsonl> and --out <codes.json>")
	}
	if rubricDir == "" {
		rubricDir = filepath.Join(filepath.Dir(inp), "..", "rubrics")
	}
	if portal == "" {
		portal = "http://localhost:8081/completion"
	}

	rubricB, err := os.ReadFile(filepath.Join(rubricDir, neutralprefix.ClassOf(inp)+".md")) //nolint:gosec // study path
	if err != nil {
		return err
	}
	rubric := string(rubricB)

	f, err := os.Open(inp) //nolint:gosec // study path
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	client := &http.Client{Timeout: 130 * time.Second}
	scores := neutralprefix.NewOMap()
	n := 0
	dec := json.NewDecoder(f)
	for dec.More() {
		var o struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		if err := dec.Decode(&o); err != nil {
			return err
		}
		prompt := neutralprefix.BuildScorePrompt(rubric, o.Text)
		content, err := npPhi4Complete(client, portal, prompt)
		if err != nil {
			return fmt.Errorf("%s: %w", o.ID, err)
		}
		scores.Set(o.ID, neutralprefix.ParseScoreLabel(content))
		n++
	}
	if err := os.WriteFile(outp, []byte(neutralprefix.PyDumps(scores, 0, true)), 0o600); err != nil {
		return err
	}
	fmt.Printf("scored %d from %s\n", n, filepath.Base(inp))
	return nil
}

// npPhi4Complete posts one grader prompt with the sampler pinned exactly as
// score_phi4.py sent it (temperature 0, n_predict 6, top_k 1, cache_prompt true,
// seed 1) and returns the raw content.
func npPhi4Complete(client *http.Client, portal, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"prompt": prompt, "temperature": 0.0, "n_predict": 6,
		"top_k": 1, "cache_prompt": true, "seed": 1,
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

// intOr parses s as an int, returning def when s is empty.
func intOr(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}

// loadClaudeUnion reads scoring/scores/claude/<cls>__A.json unioned with __B.json
// (rid -> code), matching reconcile_rerate.old_codes.
func loadClaudeUnion(study, cls string) (map[string]string, error) {
	out := map[string]string{}
	for _, half := range []string{"__A.json", "__B.json"} {
		var m map[string]string
		if err := readJSON(filepath.Join(study, "scoring", "scores", "claude", cls+half), &m); err != nil {
			return nil, err
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out, nil
}

// loadScores merges rater codes from a directory (all *.json) or a single file,
// matching aggregate.load_scores. Provenance sidecars (*.prov.json, written by
// `rate --prov`) are skipped: they hold token-usage numbers, not id->code maps, so
// unmarshalling one as codes fails.
func loadScores(path string) (map[string]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var files []string
	if info.IsDir() {
		matches, err := filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, f := range matches {
			if strings.HasSuffix(f, ".prov.json") {
				continue
			}
			files = append(files, f)
		}
		sort.Strings(files)
	} else {
		files = []string{path}
	}
	out := map[string]string{}
	for _, f := range files {
		var m map[string]string
		if err := readJSON(f, &m); err != nil {
			return nil, err
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out, nil
}
