package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("build-calib-spa", raBuildCalibSpa) }

// calibMeta is one calib_map.json entry: which grid cell a calibration id points
// at. Only glyph/scenario are shown blind; model/arm/seed locate the response.
type calibMeta struct {
	Glyph    string `json:"glyph"`
	Scenario int    `json:"scenario"`
	Model    string `json:"model"`
	Arm      string `json:"arm"`
	Seed     int    `json:"seed"`
}

// raBuildCalibSpa injects the calibration items into the blind-scoring app
// template and writes the single-file HTML. Ported from build_calib_spa.py.
//
//	corpos-lab rest-axis build-calib-spa \
//	  --calib-map <calib_map.json> --template <template.html> \
//	  --study <study-dir> [--out <out.html>] [--answers-key v1]
//
// It reads calib_map.json (id -> {glyph, scenario, model, arm, seed}), pulls the
// ground truth from each glyph's CORRECT_COMPLETIONS.md, the task from the
// scenario materials, and the response text, then fills the template. The
// embedded items are blind: id, task, ground truth, response only.
func raBuildCalibSpa(args []string) error {
	calibMap, args := flagValue(args, "--calib-map")
	template, args := flagValue(args, "--template")
	study, args := flagValue(args, "--study")
	out, args := flagValue(args, "--out")
	answersKey, args := flagValue(args, "--answers-key")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if calibMap == "" || template == "" || study == "" {
		return fmt.Errorf("need --calib-map, --template and --study")
	}
	if answersKey == "" {
		answersKey = "v1"
	}
	if out == "" {
		out = filepath.Join(filepath.Dir(calibMap), "calibration.html")
	}

	var m map[string]calibMeta
	if err := readJSONFile(calibMap, &m); err != nil {
		return err
	}
	tpl, err := os.ReadFile(template) //nolint:gosec // an operator-supplied template path
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids) // C00, C01, ... stable order, matching python sorted(m.items())

	completions := map[string]map[int]string{} // glyph -> parsed CORRECT_COMPLETIONS.md
	items := make([]restaxis.CalibItem, 0, len(ids))
	for _, id := range ids {
		meta := m[id]
		gc, ok := completions[meta.Glyph]
		if !ok {
			b, err := os.ReadFile(filepath.Join(study, meta.Glyph, "CORRECT_COMPLETIONS.md")) //nolint:gosec // a study ground-truth path
			if err != nil {
				return err
			}
			gc = restaxis.ParseCompletions(string(b))
			completions[meta.Glyph] = gc
		}
		gt := gc[meta.Scenario] // python .get(s, "") -> "" when absent

		taskB, err := os.ReadFile(filepath.Join(study, meta.Glyph, "materials", fmt.Sprintf("scenario_%d.md", meta.Scenario))) //nolint:gosec // a study materials path
		if err != nil {
			return err
		}
		task := restaxis.StripText(string(taskB))

		respPath := filepath.Join(study, "runs", meta.Model, meta.Glyph, "s"+strconv.Itoa(meta.Scenario), "out", "responses", meta.Arm+"_"+strconv.Itoa(meta.Seed)+".txt")
		resp := "(missing)"
		if b, err := os.ReadFile(respPath); err == nil { //nolint:gosec // a study response path
			resp = restaxis.StripText(string(b))
		}

		items = append(items, restaxis.CalibItem{ID: id, Task: task, GT: gt, Resp: resp})
	}

	html := restaxis.BuildCalibSPA(string(tpl), items, answersKey)
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(html), 0o600); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "built %s with %d blind items (%d KB)\n", out, len(items), len(html)/1024)
	return nil
}
