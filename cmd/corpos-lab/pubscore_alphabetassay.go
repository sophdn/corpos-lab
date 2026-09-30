package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"corpos-lab/internal/alphabetassay"
)

// paAnalyze ports assay-scoring/analyze.py's IO: it reads key_<class>.json and the
// partials/rater_[AB]_<class>__*.json code files from the assay dir, computes the
// per-cell strict-consensus tallies and per-class agreement, writes analysis.json,
// and prints the table.
//
// Usage: corpos-lab pub-score alphabet-analyze <assay-dir>
func paAnalyze(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: corpos-lab pub-score alphabet-analyze <assay-dir>")
	}
	dir := args[0]
	inputs := map[string]alphabetassay.ClassInput{}
	for _, cls := range alphabetassay.Classes {
		keyf := filepath.Join(dir, "key_"+cls+".json")
		if _, err := os.Stat(keyf); err != nil {
			continue // no key file: Analyze records it as missing "no key"
		}
		var key map[string]alphabetassay.KeyEntry
		if err := readJSON(keyf, &key); err != nil {
			return err
		}
		a, err := aaLoadRater(dir, cls, "A")
		if err != nil {
			return err
		}
		b, err := aaLoadRater(dir, cls, "B")
		if err != nil {
			return err
		}
		inputs[cls] = alphabetassay.ClassInput{Key: key, ACodes: a, BCodes: b}
	}
	res := alphabetassay.Analyze(inputs)
	if err := writeJSON(filepath.Join(dir, "analysis.json"), res); err != nil {
		return err
	}
	fmt.Print(alphabetassay.FormatTable(res))
	return nil
}

// aaLoadRater merges all partials/rater_<r>_<class>__*.json files (each id -> code)
// for one rater of one class, matching analyze.py's load_rater.
func aaLoadRater(dir, cls, r string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "partials", "rater_"+r+"_"+cls+"__*.json"))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, f := range files {
		var codes map[string]string
		if err := readJSON(f, &codes); err != nil {
			return nil, err
		}
		for id, code := range codes {
			out[id] = code
		}
	}
	return out, nil
}

// paCollectAssay ports assay-scoring/collect_assay.py's IO: it walks the study's
// per-class response runs, builds the blind-rater packets (shuffled {id,text}) and
// the private keys, and writes them to the out dir.
//
// Usage: corpos-lab pub-score alphabet-collect-assay <study-dir> <out-dir>
func paCollectAssay(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: corpos-lab pub-score alphabet-collect-assay <study-dir> <out-dir>")
	}
	study, out := args[0], args[1]
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	byClass := map[string][]alphabetassay.RawResponse{}
	for _, cls := range alphabetassay.Classes {
		dirs, err := filepath.Glob(filepath.Join(study, cls, "runs", "asy-"+cls+"-s*", "out", "responses"))
		if err != nil {
			return err
		}
		sort.Strings(dirs)
		for _, rdir := range dirs {
			scen, model, ok := alphabetassay.ParseRunDir(cls, rdir)
			if !ok {
				continue
			}
			txts, err := filepath.Glob(filepath.Join(rdir, "*.txt"))
			if err != nil {
				return err
			}
			sort.Strings(txts)
			for _, f := range txts {
				base := filepath.Base(f)
				base = base[:len(base)-len(".txt")]
				cond, seed, ok := alphabetassay.ParseCondSeed(base)
				if !ok {
					continue
				}
				b, err := os.ReadFile(f) //nolint:gosec // a study response path
				if err != nil {
					return err
				}
				byClass[cls] = append(byClass[cls], alphabetassay.RawResponse{
					Scenario: scen, Model: model, Condition: cond, Seed: seed, Text: string(b),
				})
			}
		}
	}
	collected := alphabetassay.CollectAll(byClass, alphabetassay.DefaultCollectSeed)
	for _, cls := range alphabetassay.Classes {
		cc := collected[cls]
		rows := make([]map[string]any, 0, len(cc.Items))
		for _, it := range cc.Items {
			rows = append(rows, map[string]any{"id": it.ID, "text": it.Text})
		}
		if err := writeJSONL(filepath.Join(out, "responses_"+cls+".jsonl"), rows); err != nil {
			return err
		}
		if err := writeJSON(filepath.Join(out, "key_"+cls+".json"), cc.Keys); err != nil {
			return err
		}
		fmt.Printf("%s: %d responses\n", cls, len(cc.Items))
	}
	fmt.Println("done")
	return nil
}

func init() {
	registerPubScoreMode("alphabet-analyze", paAnalyze)
	registerPubScoreMode("alphabet-collect-assay", paCollectAssay)
}
