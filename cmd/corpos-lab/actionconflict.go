package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"corpos-lab/internal/actionconflict"
)

// runActionConflict is the thin IO wiring over internal/actionconflict, the
// deterministic content-priority-under-conflict pipeline. Modes:
//
//	action-conflict score   <runs-root> [-out DIR]   -> per_response.jsonl + tally.json
//	action-conflict loop-score <runs-root> [-out DIR] -> per_response.jsonl + tally.json (agentic loop)
//	action-conflict slices  <study-dir> [-out DIR]   -> lowconf.jsonl + pilot.jsonl + det_map.json
//	action-conflict ingest-blind <verdicts.jsonl> [-out DIR] [-family NAME]
//	                                                 -> lowconf.<family>.json
//	action-conflict analyze <auto-dir>               -> final_verdicts.jsonl + splits.json + OVERRIDE_TABLE.md
func runActionConflict(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab action-conflict <score|loop-score|slices|ingest-blind|analyze> …")
		return 2
	}
	mode, rest := args[0], args[1:]
	var err error
	switch mode {
	case "score":
		err = acScore(rest)
	case "loop-score":
		err = acLoopScore(rest)
	case "slices":
		err = acSlices(rest)
	case "ingest-blind":
		err = acIngestBlind(rest)
	case "analyze":
		err = acAnalyze(rest)
	default:
		fmt.Fprintln(os.Stderr, "action-conflict: unknown mode "+mode)
		return 2
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: action-conflict %s: %v\n", mode, err)
		return 1
	}
	return 0
}

func flagValue(args []string, name string) (string, []string) {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1], append(append([]string{}, args[:i]...), args[i+2:]...)
		}
	}
	return "", args
}

func legPrecision(runName string) string {
	switch {
	case strings.HasSuffix(runName, "-cc-medium"):
		return "medium"
	case strings.HasSuffix(runName, "-cc-strong"):
		return "strong"
	default:
		return "weak"
	}
}

func scenarioOf(cell string) string {
	for _, s := range []string{"config-target", "api-version", "record-location"} {
		if strings.Contains(cell, s) {
			return s
		}
	}
	return ""
}

// verdictFn scores one response's text under a scenario. The single-turn path
// uses actionconflict.Classify (text markers); the loop path uses
// actionconflict.LoopClassify (executed-action order), which ignores truncation.
type verdictFn func(text, scen string, truncated bool) (verdict, confidence, note string)

// acScore scores single-turn runs with the deterministic text classifier.
func acScore(args []string) error {
	return acScoreWith(args, actionconflict.Classify)
}

// acLoopScore scores agentic-loop runs by executed-action order read from the
// transcript's OBSERVATION lines. It is fully deterministic — no rater slice.
func acLoopScore(args []string) error {
	return acScoreWith(args, func(text, scen string, _ bool) (string, string, string) {
		return actionconflict.LoopClassify(text, scen)
	})
}

// acScoreWith walks the run dirs under runs-root and scores every response.
func acScoreWith(args []string, classify verdictFn) error {
	out, rest := flagValue(args, "-out")
	if len(rest) != 1 {
		return fmt.Errorf("need a runs-root")
	}
	root := rest[0]
	if out == "" {
		out = filepath.Join(filepath.Dir(strings.TrimRight(root, "/")), "scoring", "auto")
	}
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}

	var rows []map[string]any
	runDirs, err := findRunDirs(root)
	if err != nil {
		return err
	}
	for _, rd := range runDirs {
		name := filepath.Base(rd)
		doc, results, err := readResults(filepath.Join(rd, "out", "results.json"))
		if err != nil {
			continue // no results for this cell
		}
		// The scenario key is the recorded item_id — authoritative and decoupled
		// from the run-dir naming. Fall back to the legacy name-guess only when a
		// run predates the item_id field.
		scen, _ := doc["item_id"].(string)
		if scen == "" {
			scen = scenarioOf(name)
		}
		if scen == "" {
			continue
		}
		modelID, _ := doc["model_id"].(string)
		prec := legPrecision(name)
		for _, r := range results {
			cond := firstString(r, "condition", "Condition")
			run := r["run"]
			if run == nil {
				run = r["Run"]
			}
			obs, _ := firstMap(r, "observed", "Observed")
			trunc := boolField(obs, "truncated", "Truncated")
			text := readResponse(rd, cond, run)
			verdict, conf, note := classify(text, scen, trunc)
			precision := ""
			if cond == "canon_conflict" {
				precision = prec
			}
			rows = append(rows, map[string]any{
				"cell": name, "scenario": scen, "model": modelID, "condition": cond,
				"precision": precision, "run": run, "truncated": trunc,
				"verdict": verdict, "confidence": conf, "note": note, "chars": len(text),
			})
		}
	}
	if err := writeJSONL(filepath.Join(out, "per_response.jsonl"), rows); err != nil {
		return err
	}

	tally := map[string]map[string]int{}
	for _, r := range rows {
		key := fmt.Sprintf("%v|%v|%v|%v", r["scenario"], r["model"], r["condition"], r["precision"])
		d := tally[key]
		if d == nil {
			d = map[string]int{"A_local": 0, "A_canon": 0, "neither": 0, "malformed": 0, "unscoreable": 0, "n": 0, "low_conf": 0}
			tally[key] = d
		}
		d[r["verdict"].(string)]++
		d["n"]++
		if r["confidence"] == "low" {
			d["low_conf"]++
		}
	}
	if err := writeJSON(filepath.Join(out, "tally.json"), tally); err != nil {
		return err
	}
	n := len(rows)
	low := 0
	for _, r := range rows {
		if r["confidence"] == "low" {
			low++
		}
	}
	fmt.Fprintf(os.Stderr, "scored %d responses; deterministic-high=%d, needs-rater-low=%d -> %s\n",
		n, n-low, low, out)
	return nil
}

// acSlices reads per_response.jsonl and builds the rater slices + det map.
func acSlices(args []string) error {
	out, rest := flagValue(args, "-out")
	if len(rest) != 1 {
		return fmt.Errorf("need a study-dir")
	}
	studyDir := rest[0]
	if out == "" {
		out = filepath.Join(studyDir, "scoring", "auto")
	}
	rows, err := readRows(filepath.Join(out, "per_response.jsonl"))
	if err != nil {
		return err
	}
	textOf := func(r actionconflict.Row) string {
		name := fmt.Sprintf("%s_%d.txt", r.Condition, r.Run)
		// Two study layouts: a parent dir with per-scenario subdirs
		// (studyDir/<scenario>/runs/<cell>, chain 543), and a single-class study
		// dir whose runs sit directly under it (studyDir/runs/<cell>, this chain).
		for _, p := range []string{
			filepath.Join(studyDir, r.Scenario, "runs", r.Cell, "out", "responses", name),
			filepath.Join(studyDir, "runs", r.Cell, "out", "responses", name),
		} {
			if b, err := os.ReadFile(p); err == nil { //nolint:gosec // a study response path
				return string(b)
			}
		}
		return ""
	}

	low := actionconflict.Lowconf(rows)
	if err := writeSlice(filepath.Join(out, "lowconf.jsonl"), low, textOf); err != nil {
		return err
	}
	// The blind bundle for the MCP-free blind-action-scorer subagent: opaque ids
	// (no condition), shuffled order, plus a secret token -> Rid keymap the
	// operator holds. The Rid itself names the condition, so the rater must never
	// see it. See tools/blind-scorer/README.md.
	if err := writeBlindBundle(out, low, textOf); err != nil {
		return err
	}
	pilot := actionconflict.Pilot(rows, func(items []actionconflict.Row) {
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	})
	if err := writeSlice(filepath.Join(out, "pilot.jsonl"), pilot, textOf); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(out, "det_map.json"), actionconflict.DetMap(rows)); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "lowconf %d, pilot %d -> %s\n", len(low), len(pilot), out)
	return nil
}

// writeBlindBundle writes the anonymized bundle (blind_bundle.jsonl) the blind
// subagent scores, plus the secret token -> Rid keymap (blind_keymap.json) the
// operator keeps. The bundle carries opaque ids only, in shuffled order, so
// neither the id nor the ordering reveals a response's condition.
func writeBlindBundle(out string, rows []actionconflict.Row, textOf func(actionconflict.Row) string) error {
	keymap := make(map[string]string, len(rows))
	items := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		rid := actionconflict.Rid(r)
		tok := actionconflict.AnonID(rid)
		keymap[tok] = rid
		items = append(items, map[string]any{"id": tok, "scenario": r.Scenario, "text": textOf(r)})
	}
	rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	if err := writeJSONL(filepath.Join(out, "blind_bundle.jsonl"), items); err != nil {
		return err
	}
	return writeJSON(filepath.Join(out, "blind_keymap.json"), keymap)
}

// knownRaterFamily is the set of families the analyze step reads. ingest-blind
// refuses to write any other name, so a typo cannot produce a file analyze
// silently ignores.
var knownRaterFamily = map[string]bool{"deepseek": true, "devstral": true, "claude": true}

// acIngestBlind converts the blind-action-scorer subagent's JSON-Lines verdicts
// into lowconf.<family>.json — the rater file the analyze step reads. The blind
// subagent is the Claude family, so the family defaults to "claude".
func acIngestBlind(args []string) error {
	family, rest := flagValue(args, "-family")
	if family == "" {
		family = "claude"
	}
	if !knownRaterFamily[family] {
		return fmt.Errorf("family %q is not one analyze reads (deepseek, devstral, claude)", family)
	}
	out, rest := flagValue(rest, "-out")
	keymapPath, rest := flagValue(rest, "-keymap")
	if len(rest) != 1 {
		return fmt.Errorf("need a verdicts.jsonl from the blind-action-scorer subagent")
	}
	verdictsPath := rest[0]
	if out == "" {
		out = filepath.Dir(verdictsPath)
	}
	if keymapPath == "" {
		keymapPath = filepath.Join(out, "blind_keymap.json")
	}
	verdicts, err := readBlindVerdicts(verdictsPath)
	if err != nil {
		return err
	}
	// Translate the subagent's opaque ids back to Rids via the secret keymap, so
	// the rater file is keyed the way analyze reads it. The rater never held this
	// map; the operator does.
	var keymap map[string]string
	if err := readJSON(keymapPath, &keymap); err != nil {
		return fmt.Errorf("read blind keymap %s: %w", keymapPath, err)
	}
	for i, v := range verdicts {
		rid, ok := keymap[v.ID]
		if !ok {
			return fmt.Errorf("verdict %d: opaque id %q not in keymap %s", i, v.ID, keymapPath)
		}
		verdicts[i].ID = rid
	}
	m, err := actionconflict.IngestBlind(verdicts)
	if err != nil {
		return err
	}
	dst := filepath.Join(out, "lowconf."+family+".json")
	if err := writeJSON(dst, m); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "ingested %d blind verdicts (%s) -> %s\n", len(m), family, dst)
	return nil
}

func readBlindVerdicts(path string) ([]actionconflict.BlindVerdict, error) {
	f, err := os.Open(path) //nolint:gosec // a study scoring path
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var out []actionconflict.BlindVerdict
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := sc.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var v actionconflict.BlindVerdict
		if err := json.Unmarshal(line, &v); err != nil {
			return nil, fmt.Errorf("bad blind verdict line: %w", err)
		}
		out = append(out, v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// acAnalyze resolves final verdicts and renders the override table.
func acAnalyze(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("need an auto-dir")
	}
	auto := args[0]
	var det map[string]actionconflict.DetRow
	if err := readJSON(filepath.Join(auto, "det_map.json"), &det); err != nil {
		return err
	}
	ds, dv, cl := map[string]string{}, map[string]string{}, map[string]string{}
	_ = readJSON(filepath.Join(auto, "lowconf.deepseek.json"), &ds)
	_ = readJSON(filepath.Join(auto, "lowconf.devstral.json"), &dv)
	_ = readJSON(filepath.Join(auto, "lowconf.claude.json"), &cl)

	final, splits := actionconflict.FinalVerdict(det, ds, dv, cl)
	ids := make([]string, 0, len(final))
	for id := range final {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var fr []map[string]any
	for _, id := range ids {
		d := final[id]
		fr = append(fr, map[string]any{
			"id": id, "verdict": d.Verdict, "confidence": d.Confidence,
			"scenario": d.Scenario, "condition": d.Condition, "precision": d.Precision,
			"model": d.Model, "final": d.Final, "source": d.Source,
		})
	}
	if err := writeJSONL(filepath.Join(auto, "final_verdicts.jsonl"), fr); err != nil {
		return err
	}
	if splits == nil {
		splits = []actionconflict.Split{}
	}
	if err := writeJSON(filepath.Join(auto, "splits.json"), splits); err != nil {
		return err
	}
	report := actionconflict.OverrideReport(final, len(splits))
	if err := os.WriteFile(filepath.Join(auto, "OVERRIDE_TABLE.md"), []byte(report), 0o600); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, report)
	return nil
}

// --- small IO helpers ---

func findRunDirs(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		_, recErr := os.Stat(filepath.Join(p, "run-record.json"))
		info, outErr := os.Stat(filepath.Join(p, "out"))
		if recErr == nil && outErr == nil && info.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func readResults(path string) (map[string]any, []map[string]any, error) {
	b, err := os.ReadFile(path) //nolint:gosec // a study results path
	if err != nil {
		return nil, nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err == nil {
		if rawRows, ok := doc["rows"].([]any); ok {
			return doc, toMaps(rawRows), nil
		}
	}
	var list []any
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, nil, err
	}
	return map[string]any{}, toMaps(list), nil
}

func toMaps(raw []any) []map[string]any {
	out := make([]map[string]any, 0, len(raw))
	for _, r := range raw {
		if m, ok := r.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func readResponse(rd, cond string, run any) string {
	p := filepath.Join(rd, "out", "responses", fmt.Sprintf("%s_%v.txt", cond, run))
	b, err := os.ReadFile(p) //nolint:gosec // a study response path
	if err != nil {
		return ""
	}
	return string(b)
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok {
			return s
		}
	}
	return ""
}

func firstMap(m map[string]any, keys ...string) (map[string]any, bool) {
	for _, k := range keys {
		if v, ok := m[k].(map[string]any); ok {
			return v, true
		}
	}
	return nil, false
}

func boolField(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if b, ok := m[k].(bool); ok {
			return b
		}
	}
	return false
}

func readRows(path string) ([]actionconflict.Row, error) {
	f, err := os.Open(path) //nolint:gosec // a study auto-dir path
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var rows []actionconflict.Row
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if len(strings.TrimSpace(sc.Text())) == 0 {
			continue
		}
		var r actionconflict.Row
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, err
		}
		rows = append(rows, r)
	}
	return rows, sc.Err()
}

func writeSlice(path string, rows []actionconflict.Row, textOf func(actionconflict.Row) string) error {
	var lines []map[string]any
	for _, r := range rows {
		lines = append(lines, map[string]any{"id": actionconflict.Rid(r), "text": textOf(r), "scenario": r.Scenario})
	}
	return writeJSONL(path, lines)
}

func writeJSONL(path string, rows []map[string]any) error {
	f, err := os.Create(path) //nolint:gosec // a study auto-dir path
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	w := bufio.NewWriter(f)
	for _, r := range rows {
		b, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return w.Flush()
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path) //nolint:gosec // a study auto-dir path
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
