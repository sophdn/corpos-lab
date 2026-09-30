package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"corpos-lab/internal/pubshared"
)

func init() {
	registerPubScoreMode("shared-scramble", raSharedScramble)
	registerPubScoreMode("shared-build-app", raSharedBuildApp)
	registerPubScoreMode("shared-anchor-eval", raSharedAnchorEval)
}

// raSharedScramble ports scramble.py: it reads a glyph markdown file, scrambles
// or vocab-swaps the descriptive words while preserving structure, and writes the
// result. In shuffle mode the reorder is a deterministic seeded Go shuffle
// (structural parity with the python original, not byte-identical).
//
// Usage: corpos-lab pub-score shared-scramble <glyph.md> <out.md> [seed] [--vocab-swap] [--neutralize-title] [--strength <f>]
// --strength <f> (0..1) selects the graded dial (task 4258): f=0 is the weak
// shuffle, f=1 the strong vocab-swap, and the middle vocab-swaps a fraction f of
// words. When absent, the published shuffle/vocab-swap modes are used unchanged.
func raSharedScramble(args []string) error {
	strength := -1.0 // sentinel: not graded
	var rest []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--strength" && i+1 < len(args):
			v, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return fmt.Errorf("--strength must be a number: %w", err)
			}
			strength = v
			i++
		case strings.HasPrefix(args[i], "--strength="):
			v, err := strconv.ParseFloat(strings.TrimPrefix(args[i], "--strength="), 64)
			if err != nil {
				return fmt.Errorf("--strength must be a number: %w", err)
			}
			strength = v
		default:
			rest = append(rest, args[i])
		}
	}
	args = rest
	var positional []string
	flags := map[string]bool{}
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			flags[a] = true
		} else {
			positional = append(positional, a)
		}
	}
	if len(positional) < 2 {
		return fmt.Errorf("usage: shared-scramble <glyph.md> <out.md> [seed] [--vocab-swap] [--neutralize-title]")
	}
	seed := 42
	if len(positional) > 2 {
		s, err := strconv.Atoi(positional[2])
		if err != nil {
			return fmt.Errorf("seed must be an integer: %w", err)
		}
		seed = s
	}
	mode := "shuffle"
	if flags["--vocab-swap"] {
		mode = "vocab-swap"
	}
	neutralize := flags["--neutralize-title"]

	b, err := os.ReadFile(positional[0]) //nolint:gosec // an operator-supplied glyph path
	if err != nil {
		return err
	}
	rng := rand.New(rand.NewSource(int64(seed))) //nolint:gosec // not cryptographic: a reproducible scramble control
	var out string
	if strength >= 0 {
		out = pubshared.TransformGraded(string(b), strength, neutralize, rng)
	} else {
		out = pubshared.Transform(string(b), mode, neutralize, rng.Shuffle)
	}
	return os.WriteFile(positional[1], []byte(out), 0o600)
}

// raSharedBuildApp ports build_app.py: it validates the blind items and labels,
// fills the app template, and writes the single-file HTML.
//
// Usage: corpos-lab pub-score shared-build-app --items <f|-> --labels <f> --title <t> --store-key <k> [--rubric <f>] [--out <path>] --template <f>
func raSharedBuildApp(args []string) error {
	items, args := flagValue(args, "--items")
	labels, args := flagValue(args, "--labels")
	title, args := flagValue(args, "--title")
	storeKey, args := flagValue(args, "--store-key")
	rubricPath, args := flagValue(args, "--rubric")
	outPath, args := flagValue(args, "--out")
	templatePath, args := flagValue(args, "--template")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if items == "" || labels == "" || title == "" || storeKey == "" || templatePath == "" {
		return fmt.Errorf("need --items, --labels, --title, --store-key and --template")
	}

	itemsData, err := loadJSONBytes(items)
	if err != nil {
		return err
	}
	labelsData, err := loadJSONBytes(labels)
	if err != nil {
		return err
	}
	itemsVal, err := pubshared.ParseOrdered(itemsData)
	if err != nil {
		return err
	}
	labelsVal, err := pubshared.ParseOrdered(labelsData)
	if err != nil {
		return err
	}
	var rubric *string
	if rubricPath != "" {
		rb, rerr := os.ReadFile(rubricPath) //nolint:gosec // an operator-supplied rubric path
		if rerr != nil {
			return rerr
		}
		s := string(rb)
		rubric = &s
	}
	tpl, err := os.ReadFile(templatePath) //nolint:gosec // an operator-supplied template path
	if err != nil {
		return err
	}
	out := outPath
	if out == "" {
		home, herr := os.UserHomeDir()
		if herr != nil {
			return herr
		}
		out = filepath.Join(home, "blind-scoring", storeKey+".html")
	}

	html, err := pubshared.BuildApp(string(tpl), itemsVal, labelsVal, title, storeKey, rubric)
	if err != nil {
		// A validation/blindness failure aborts with a non-zero exit, the same as
		// python's build_app; the caller (runPubScore) reports the message.
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(html), 0o600); err != nil {
		return err
	}

	if pubshared.MissingCriteria(labelsVal, rubric) {
		fmt.Fprintln(os.Stderr, "build_app: NOTE the label set has more than two codes but carries no per-code "+
			"`criteria` and no --rubric panel. A rater may not be able to tell borderline codes "+
			"apart (e.g. C vs Ic). Add the discriminating criteria per code, or a --rubric "+
			"panel with the canonical target form.")
	}
	if strings.HasPrefix(out, "/tmp") {
		fmt.Fprintf(os.Stderr, "build_app: WARNING %s is under /tmp — a snap-confined browser "+
			"(Ubuntu's default Firefox) cannot open it; write under $HOME instead\n", out)
	}
	nItems := jsonArrayLen(itemsVal)
	nLabels := jsonArrayLen(labelsVal)
	fmt.Printf("build_app: wrote %s — %d blind items, %d labels (%d KB). Open it with: xdg-open %s\n",
		out, nItems, nLabels, len(html)/1024, out)
	return nil
}

// raSharedAnchorEval ports anchor_eval.py: it scores a rater against the human
// anchor and the Claude consensus, printing the report to stdout.
//
// Usage: corpos-lab pub-score shared-anchor-eval --anchor-dir <dir> [--rater <file>] [--label <name>]
func raSharedAnchorEval(args []string) error {
	anchorDir, args := flagValue(args, "--anchor-dir")
	raterPath, args := flagValue(args, "--rater")
	label, args := flagValue(args, "--label")
	if len(args) != 0 {
		return fmt.Errorf("unexpected args: %v", args)
	}
	if anchorDir == "" {
		return fmt.Errorf("need --anchor-dir")
	}
	rids, err := orderedRids(filepath.Join(anchorDir, "anchor_slice.jsonl"))
	if err != nil {
		return err
	}
	var humanByH map[string]string
	if err := readJSON(filepath.Join(anchorDir, "human_codes.json"), &humanByH); err != nil {
		return err
	}
	var ra, rb map[string]string
	if err := readJSON(filepath.Join(anchorDir, "anchor_raterA.json"), &ra); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(anchorDir, "anchor_raterB.json"), &rb); err != nil {
		return err
	}
	human := pubshared.BuildHuman(rids, humanByH)
	claude := pubshared.BuildClaude(rids, ra, rb)

	var rater map[string]string
	raterLabel := label
	if raterPath != "" {
		if err := readJSON(raterPath, &rater); err != nil {
			return err
		}
		if raterLabel == "" {
			raterLabel = strings.TrimSuffix(filepath.Base(raterPath), filepath.Ext(raterPath))
		}
	}
	fmt.Print(pubshared.AnchorEvalReport(filepath.Base(anchorDir), rids, human, claude, rater, raterLabel))
	return nil
}

// loadJSONBytes reads a JSON spec: a path, or "-" for stdin.
func loadJSONBytes(spec string) ([]byte, error) {
	if spec == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(spec) //nolint:gosec // an operator-supplied JSON path
}

// jsonArrayLen returns the element count of a parsed JSON array, or 0.
func jsonArrayLen(v any) int {
	if arr, ok := v.([]any); ok {
		return len(arr)
	}
	return 0
}

// orderedRids reads anchor_slice.jsonl and returns each line's "id" in order.
func orderedRids(path string) ([]string, error) {
	b, err := os.ReadFile(path) //nolint:gosec // a study anchor path
	if err != nil {
		return nil, err
	}
	var rids []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, err
		}
		rids = append(rids, rec.ID)
	}
	return rids, nil
}
