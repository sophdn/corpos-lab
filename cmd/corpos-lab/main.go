// Command corpos-lab is the headless lab controller CLI. It runs a study
// end-to-end from a single definition file — reproducible with no agent in the
// loop — and is the surface corpos drives as an external tool.
//
// Usage:
//
//	corpos-lab run-study <def.toml> [-work DIR]
//	corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]
//
// Exit 0 on a completed run, non-zero on any failure (the record is written
// regardless).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/batteryrun"
	"corpos-lab/internal/control"
	"corpos-lab/internal/disclosure"
	"corpos-lab/internal/image"
	"corpos-lab/internal/model"
	"corpos-lab/internal/persist"
	"corpos-lab/internal/power"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/study"
	"corpos-lab/internal/studyclose"
	"corpos-lab/internal/substrate"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

const usage = "usage:\n" +
	"  corpos-lab run-study <def.toml> [-work DIR]\n" +
	"  corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-all-items]\n" +
	"  corpos-lab rate (--slices-dir DIR | --slice FILE) --out-dir DIR --rater-id ID (--rubric R.md | --action) [--base URL] [--model M] [--api-key-env ENV] [--max-tokens N] [--jobs N] [--prov]\n" +
	"  corpos-lab action-conflict <score <runs-root> | slices <study-dir> | analyze <auto-dir>> [-out DIR]\n" +
	"  corpos-lab glyph-digest <slug> | --all | --file <path> [-candidates DIR]\n" +
	"  corpos-lab glyph-lint <candidate.md> [-registry ALPHABET.md]\n" +
	"  corpos-lab plan (-ci-half-width W | -effect-size D) [-p P] [-budget-runs N | -budget-seconds S -per-run-seconds R -grid-cells G] [-out FILE]\n" +
	"  corpos-lab plan -sequential (-ci-half-width W | -effect-size D) -k K -n N [-budget-runs N]\n" +
	"  corpos-lab study-close <study-dir>\n" +
	"  corpos-lab disclosure verify [-public-dir DIR] [-manifest FILE]\n" +
	"  corpos-lab disclosure promote [-canon DIR] [-registry FILE] [-public-dir DIR] [-manifest FILE]"

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "run-study":
		return runStudy(args[1:])
	case "battery":
		return runBattery(args[1:])
	case "rate":
		return runRate(args[1:])
	case "action-conflict":
		return runActionConflict(args[1:])
	case "glyph-digest":
		return runGlyphDigest(args[1:])
	case "glyph-lint":
		return runGlyphLint(args[1:])
	case "plan":
		return runPlan(args[1:])
	case "study-close":
		return runStudyClose(args[1:])
	case "disclosure":
		return runDisclosure(args[1:])
	default:
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
}

// runPlan sizes a study's per-cell n from a precision target and a budget, so an
// agent no longer hand-sets n. It prints the recommended n, which bound set it,
// and the rationale, and with -out records the plan as JSON to carry with the
// study. With -sequential it instead reports the stopping-rule verdict for the
// runs so far (-k successes in -n runs): STOP when the CI is tight enough or the
// budget cap is hit, else CONTINUE. All the math lives in internal/power.
func runPlan(args []string) int {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	ciHalfWidth := fs.Float64("ci-half-width", 0, "target 95% CI half-width in (0,1)")
	effectSize := fs.Float64("effect-size", 0, "proportion difference to call real; resolved to a half-width of delta/2")
	p := fs.Float64("p", 0, "worst-case proportion for planning (default 0.5)")
	budgetRuns := fs.Int("budget-runs", 0, "per-cell run cap (0 = none)")
	budgetSeconds := fs.Float64("budget-seconds", 0, "wall-clock/compute budget in seconds")
	perRunSeconds := fs.Float64("per-run-seconds", 0, "seconds per run, to turn the budget into a cap")
	gridCells := fs.Int("grid-cells", 1, "cells the budget is spread across")
	sequential := fs.Bool("sequential", false, "sequential-stop check: given -k and -n, print STOP/CONTINUE")
	k := fs.Int("k", 0, "successes so far (sequential mode)")
	n := fs.Int("n", 0, "runs so far (sequential mode)")
	out := fs.String("out", "", "write the plan as JSON to this path")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	target := *ciHalfWidth
	if target == 0 && *effectSize > 0 {
		w, err := power.HalfWidthForEffect(*effectSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: plan: %v\n", err)
			return 2
		}
		target = w
	}
	if target == 0 {
		fmt.Fprintln(os.Stderr, "corpos-lab: plan needs -ci-half-width or -effect-size")
		return 2
	}

	capN := *budgetRuns
	if capN == 0 && *budgetSeconds > 0 {
		c, err := power.BudgetCapPerCell(*budgetSeconds, *perRunSeconds, *gridCells)
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: plan: %v\n", err)
			return 2
		}
		capN = c
	}

	if *sequential {
		stop, reason, hw, err := power.SequentialStop(*k, *n, target, capN, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: plan: %v\n", err)
			return 2
		}
		verdict := "CONTINUE"
		if stop {
			verdict = "STOP"
		}
		fmt.Printf("%s (%s): k=%d n=%d, current 95%% CI half-width %.3f, target %.3f\n",
			verdict, reason, *k, *n, hw, target)
		return 0
	}

	plan, err := power.PlanN(target, *p, 0, capN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: plan: %v\n", err)
		return 2
	}
	fmt.Printf("recommended n = %d per cell (%s-bound)\n%s\n", plan.RecommendedN, plan.Bound, plan.Rationale)
	if *out != "" {
		b, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: plan: %v\n", err)
			return 1
		}
		if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: plan: write %s: %v\n", *out, err)
			return 1
		}
		fmt.Printf("plan recorded -> %s\n", *out)
	}
	return 0
}

// runStudyClose reports the mechanical study-close facts INQUIRY.md keeps as
// human discipline (predictions reconciled, a floor-cleared second rater,
// mechanism controls). It ADVISES — it prints findings and always exits 0; it
// never blocks a run or a commit (suggestion 165). It says nothing about
// scientific quality, only whether the declared mechanical steps happened.
func runStudyClose(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab study-close <study-dir>")
		return 2
	}
	dir := args[0]
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "corpos-lab: study-close needs a directory (%s): %v\n", dir, err)
		return 2
	}
	in, err := gatherStudyClose(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: study-close: %v\n", err)
		return 1
	}
	fmt.Printf("corpos-lab: study-close report for %s (advisory — nothing here blocks)\n", dir)
	for _, f := range studyclose.Report(in) {
		fmt.Printf("  [%-8s] %-20s %s\n", f.Status, f.Check, f.Detail)
	}
	return 0
}

// gatherStudyClose walks a study directory and collects the mechanical facts the
// study-close report checks. It is best-effort: an unreadable or unparseable
// entry is skipped rather than failing the report.
func gatherStudyClose(dir string) (studyclose.Inputs, error) {
	var in studyclose.Inputs
	condSet := map[string]bool{}
	raterSet := map[string]bool{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // an unreadable entry is skipped, not fatal: the report is best-effort
		}
		name := d.Name()
		if d.IsDir() {
			// rater ids: the immediate subdirectories of a scores/ dir.
			if filepath.Base(filepath.Dir(path)) == "scores" {
				raterSet[name] = true
			}
			return nil
		}
		inScores := filepath.Base(filepath.Dir(path)) == "scores"
		switch {
		case strings.HasSuffix(name, ".toml"):
			var def struct {
				Conditions []string `toml:"conditions"`
			}
			if raw, e := os.ReadFile(path); e == nil {
				if toml.Unmarshal(raw, &def) == nil {
					for _, c := range def.Conditions {
						condSet[c] = true
					}
				}
			}
		case name == "run-record.json":
			var rec struct {
				Status string `json:"status"`
			}
			if raw, e := os.ReadFile(path); e == nil {
				if json.Unmarshal(raw, &rec) == nil {
					in.RunStatuses = append(in.RunStatuses, rec.Status)
				}
			}
		case name == "RECONCILED.md" || name == "reconciliation.json":
			in.ReconciliationMarker = true
		case strings.HasPrefix(strings.ToLower(name), "predictions"):
			in.PredictionsFile = true
		case inScores && strings.HasSuffix(name, ".json"):
			// a loose scores/<rater>.json — the rater id is the stem before the first dot.
			raterSet[strings.SplitN(name, ".", 2)[0]] = true
		}
		return nil
	})
	if err != nil {
		return in, err
	}
	for c := range condSet {
		in.Conditions = append(in.Conditions, c)
	}
	for r := range raterSet {
		in.SecondRaters = append(in.SecondRaters, r)
	}
	sort.Strings(in.Conditions)
	sort.Strings(in.SecondRaters)
	return in, nil
}

const disclosureUsage = "usage:\n" +
	"  corpos-lab disclosure verify [-public-dir DIR] [-manifest FILE]\n" +
	"  corpos-lab disclosure promote [-canon DIR] [-registry FILE] [-public-dir DIR] [-manifest FILE]"

// runDisclosure drives the progressive-disclosure rule.
//
// "verify" is the public-repo CI gate: it confirms the published-glyph directory
// holds exactly the approved subset, needing no access to the private canon, so
// an unapproved glyph reaching the public tree fails the gate. "promote" copies
// the approved subset from the private canon into the public tree (default
// closed — a glyph is published only if it is both approved and certified, and
// only its AC-4 block, verified against the frozen registry digest). The rule
// logic lives in internal/disclosure (tested); this stays file glue.
func runDisclosure(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, disclosureUsage)
		return 2
	}
	sub := args[0]
	publicDir := "corpus/glyph-model/published"
	manifest := "corpus/glyph-model/PUBLIC_ALPHABET.txt"
	canonDir := "corpus/private/glyph-model"
	registryPath := "corpus/private/glyph-model/ALPHABET.md"
	var extra []string
	for i := 1; i < len(args); i++ {
		needsValue := func() (string, bool) {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "corpos-lab: %s needs a value\n", args[i])
				return "", false
			}
			i++
			return args[i], true
		}
		switch args[i] {
		case "-public-dir":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			publicDir = v
		case "-manifest":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			manifest = v
		case "-canon":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			canonDir = v
		case "-registry":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			registryPath = v
		default:
			extra = append(extra, args[i])
		}
	}
	if len(extra) != 0 {
		fmt.Fprintln(os.Stderr, disclosureUsage)
		return 2
	}

	switch sub {
	case "verify":
		if err := disclosure.VerifyPublished(publicDir, manifest); err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: disclosure verify FAILED: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "corpos-lab: disclosure verify PASS — public tree matches %s\n", manifest)
		return 0
	case "promote":
		regRaw, err := os.ReadFile(registryPath) //nolint:gosec // a registry path from argv, not a secret
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: read registry: %v\n", err)
			return 1
		}
		reg, err := disclosure.ParseRegistry(string(regRaw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		manRaw, err := os.ReadFile(manifest) //nolint:gosec // a manifest path from argv, not a secret
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: read approved manifest: %v\n", err)
			return 1
		}
		approved, err := disclosure.ParseApproved(string(manRaw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		res, err := disclosure.Promote(canonDir, publicDir, reg, approved)
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: disclosure promote FAILED: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "corpos-lab: disclosure promote — published %d, removed %d (into %s)\n",
			len(res.Published), len(res.Removed), publicDir)
		if len(res.Published) > 0 {
			fmt.Fprintf(os.Stderr, "  published: %s\n", strings.Join(res.Published, ", "))
		}
		if len(res.Removed) > 0 {
			fmt.Fprintf(os.Stderr, "  removed:   %s\n", strings.Join(res.Removed, ", "))
		}
		return 0
	default:
		fmt.Fprintln(os.Stderr, disclosureUsage)
		return 2
	}
}

// runBattery runs the mechanized ALPHABET battery against one candidate and
// writes the JSON result (the mechanized item verdicts, the Deferred stubs for
// the judged items, and run-level provenance). The judged items are resolved
// separately by an assessor; this surface produces only the measured half.
//
// It takes the candidate working-doc directly: batteryrun.Run extracts the
// assessable clean entry (the fallout line plus the AC-4 assembled-glyph block)
// from it, so there is no separate extract copy to keep in sync. Point
// <candidate.md> at corpus/private/glyph-model/candidates/CANDIDATE_<slug>_*.md.
//
// Item 12 is empirical: its default-alignment verdict is measured against an
// unguided-baseline run of the class scenario across the weak local treatment
// shelf, supplied as a JSON battery.BaselineOutcome via -baseline. With no
// -baseline the item defers with a recorded gap (it never guesses a verdict).
func runBattery(args []string) int {
	outPath := ""
	base := "http://localhost:8081/v1"
	modelName := "qwen3.6-27b"
	version := ""
	repoDir := "."
	registryPath := "corpus/private/glyph-model/ALPHABET.md"
	definitionPath := "corpus/glyph-model/GLYPH_DEFINITION.md"
	provenanceTypesPath := "corpus/glyph-model/GLYPH_PROVENANCE_TYPES.md"
	baselinePath := ""
	allItems := false
	rest := []string{}
	for i := 0; i < len(args); i++ {
		needsValue := func() (string, bool) {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "corpos-lab: %s needs a value\n", args[i])
				return "", false
			}
			i++
			return args[i], true
		}
		switch args[i] {
		case "-out":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			outPath = v
		case "-base":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			base = v
		case "-model":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			modelName = v
		case "-version":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			version = v
		case "-repo":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			repoDir = v
		case "-registry":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			registryPath = v
		case "-definition":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			definitionPath = v
		case "-provenance-types":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			provenanceTypesPath = v
		case "-baseline":
			v, ok := needsValue()
			if !ok {
				return 2
			}
			baselinePath = v
		case "-all-items":
			allItems = true
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab battery <candidate.md> [-out FILE] [-model NAME] [-base URL] [-repo DIR] [-registry ALPHABET.md] [-definition GLYPH_DEFINITION.md] [-provenance-types GLYPH_PROVENANCE_TYPES.md] [-baseline BASELINE.json] [-all-items]")
		return 2
	}
	candidatePath := rest[0]
	itemID := candidateSlug(candidatePath)

	// A thinking model producing reasoning for two verdict calls is slow; give
	// each generation room rather than clip it into a transport error.
	client := model.NewOpenAI(base, modelName, version,
		model.WithHTTPClient(&http.Client{Timeout: 5 * time.Minute}))

	res, err := batteryrun.Run(context.Background(), candidatePath, itemID, client, batteryrun.Deps{
		Substrate:  substrateProbe,
		Provenance: repoStamp(repoDir),
		Props:      client.Props,
		Registry:   batteryrun.FileRegistryReader{Path: registryPath},
		Reference: batteryrun.FileReferenceReader{
			DefinitionPath:      definitionPath,
			ProvenanceTypesPath: provenanceTypesPath,
		},
		Baseline: batteryrun.FileBaselineReader{Path: baselinePath},
	}, batteryrun.Options{AllItems: allItems})
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	encoded, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: marshal result: %v\n", err)
		return 1
	}
	if outPath == "" {
		fmt.Println(string(encoded))
	} else {
		if err := os.WriteFile(outPath, append(encoded, '\n'), 0o644); err != nil { //nolint:gosec // a result record, not a secret
			fmt.Fprintf(os.Stderr, "corpos-lab: write %s: %v\n", outPath, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "corpos-lab: battery result -> %s\n", outPath)
	}
	printBatterySummary(os.Stderr, res)
	return 0
}

// candidateSlug derives an item id from a candidate filename: it strips a
// leading "CANDIDATE_" and a trailing "_<date>.md", falling back to the base
// name without extension.
func candidateSlug(path string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name = strings.TrimPrefix(name, "CANDIDATE_")
	if i := strings.LastIndex(name, "_20"); i > 0 {
		name = name[:i]
	}
	return name
}

// runGlyphDigest prints the certification digest of one or every candidate's
// AC-4 glyph block — the Go form of the retired glyph_digest.py. The block
// extraction and hashing live in battery.GlyphDigest (tested); this stays file
// glue.
func runGlyphDigest(args []string) int {
	dir := "corpus/private/glyph-model/candidates"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "-candidates" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -candidates needs a value")
				return 2
			}
			i++
			dir = args[i]
			continue
		}
		rest = append(rest, args[i])
	}

	switch {
	case len(rest) == 1 && rest[0] == "--all":
		matches, err := filepath.Glob(filepath.Join(dir, "CANDIDATE_*.md"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		sort.Strings(matches)
		for _, p := range matches {
			d, err := fileDigest(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
				return 1
			}
			fmt.Printf("%s  %s\n", candidateSlug(p), d)
		}
		return 0
	case len(rest) == 2 && rest[0] == "--file":
		d, err := fileDigest(rest[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		fmt.Println(d)
		return 0
	case len(rest) == 1:
		matches, err := filepath.Glob(filepath.Join(dir, "CANDIDATE_"+rest[0]+"_*.md"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
			return 1
		}
		switch len(matches) {
		case 0:
			fmt.Fprintf(os.Stderr, "corpos-lab: no candidate file for slug %q under %s\n", rest[0], dir)
			return 1
		case 1:
			d, err := fileDigest(matches[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
				return 1
			}
			fmt.Println(d)
			return 0
		default:
			fmt.Fprintf(os.Stderr, "corpos-lab: ambiguous slug %q: %v\n", rest[0], matches)
			return 1
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: corpos-lab glyph-digest <slug> | --all | --file <path> [-candidates DIR]")
		return 2
	}
}

// fileDigest reads a candidate file and returns its AC-4 glyph-block digest.
func fileDigest(path string) (string, error) {
	b, err := os.ReadFile(path) //nolint:gosec // a candidate file path from argv, not a secret
	if err != nil {
		return "", err
	}
	return battery.GlyphDigest(string(b))
}

// runGlyphLint runs the deterministic structural subset of the battery against
// one candidate — a fast, model-free pre-battery shape check. It lints the
// extracted entry, not the raw doc, so the AC-3 specimen does not trip the
// intent scan. The item logic lives in internal/battery (tested); this stays
// file glue, and it is the surface suggestion 166's schema enforcement reaches
// through once that item joins battery.BuildStructuralLint.
func runGlyphLint(args []string) int {
	registryPath := "corpus/private/glyph-model/ALPHABET.md"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "-registry" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -registry needs a value")
				return 2
			}
			i++
			registryPath = args[i]
			continue
		}
		rest = append(rest, args[i])
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab glyph-lint <candidate.md> [-registry ALPHABET.md]")
		return 2
	}
	path := rest[0]
	content, err := os.ReadFile(path) //nolint:gosec // a candidate file path from argv, not a secret
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}
	entry, err := battery.ExtractEntry(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	res := battery.RunSequence(context.Background(), battery.BuildStructuralLint(), battery.Input{
		ItemID:         candidateSlug(path),
		Content:        entry,
		Registry:       batteryrun.FileRegistryReader{Path: registryPath},
		ContinueOnFail: true,
	})
	for _, sr := range res.StepResults {
		fmt.Fprintf(os.Stderr, "  %-28s %s\n", sr.StepName, verdictLabel(sr.Outcome))
	}
	if res.Passed {
		fmt.Fprintf(os.Stderr, "corpos-lab: glyph-lint %q PASS (structural items)\n", candidateSlug(path))
		return 0
	}
	fmt.Fprintf(os.Stderr, "corpos-lab: glyph-lint %q FAIL: %s\n", candidateSlug(path), res.FailureReason)
	return 1
}

// printBatterySummary writes a one-line-per-item human summary of the
// mechanized outcome so a reader does not have to parse the JSON to see where a
// run landed.
func printBatterySummary(w *os.File, res batteryrun.Result) {
	fmt.Fprintf(w, "corpos-lab: battery %q — model %q, substrate %s, commit %s\n",
		res.ItemID, res.Provenance.ModelID, res.Provenance.Substrate.Kind, shortSHA(res.Provenance.Provenance.CommitSHA))
	for _, sr := range res.Sequence.StepResults {
		fmt.Fprintf(w, "  %-28s %s\n", sr.StepName, verdictLabel(sr.Outcome))
	}
	switch {
	case res.Sequence.Passed:
		fmt.Fprintln(w, "  => mechanized items PASS (judged items still Deferred to an assessor)")
	case res.AllItems:
		fmt.Fprintf(w, "  => FAIL, all items run: %s\n", res.Sequence.FailureReason)
	default:
		fmt.Fprintf(w, "  => stopped at step %d: %s\n", res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}

	// Evidence state: fold the item verdicts into a battery RunVerdict and run the
	// candidate through the one promotion rule, so maturity is a fact the tooling
	// computes rather than prose a reader reconstructs (suggestion 164).
	verdict := battery.ComposeRunVerdict(res.Sequence.ItemVerdicts())
	if state, note, err := battery.PromoteOn(battery.StateCandidate, verdict); err == nil {
		fmt.Fprintf(w, "  => evidence: candidate -> %s (%s)\n", state, note)
	}
}

func verdictLabel(o battery.StepOutcome) string {
	switch o.Kind {
	case battery.OutcomeError:
		return "ERROR: " + o.Message
	case battery.OutcomeObservation:
		return "observation"
	default:
		if o.Verdict == nil {
			return "?"
		}
		return string(o.Verdict.Kind)
	}
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	if sha == "" {
		return "(unstamped)"
	}
	return sha
}

// repoStamp stamps the instrument repo (corpos-lab) that ran the battery.
// Defaults to the working directory, so run it from the checkout; -repo
// overrides. Records its own failure rather than inventing a commit.
func repoStamp(repoDir string) func(context.Context) (provenance.Stamp, error) {
	return func(ctx context.Context) (provenance.Stamp, error) {
		dir, err := filepath.Abs(repoDir)
		if err != nil {
			return provenance.Stamp{}, err
		}
		return provenance.Capture(ctx, dir)
	}
}

func runStudy(args []string) int {
	var defPath, workDir string
	toolkitURL := persist.DefaultToolkitURL
	project := "glyph-research"
	rest := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-work":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -work needs a value")
				return 2
			}
			i++
			workDir = args[i]
		case "-toolkit-url":
			// Empty value disables persistence to the toolkit.
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -toolkit-url needs a value")
				return 2
			}
			i++
			toolkitURL = args[i]
		case "-project":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "corpos-lab: -project needs a value")
				return 2
			}
			i++
			project = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "usage: corpos-lab run-study <def.toml> [-work DIR]")
		return 2
	}
	defPath = rest[0]

	def, err := study.LoadDef(defPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: %v\n", err)
		return 1
	}

	if workDir == "" {
		workDir = filepath.Join(filepath.Dir(defPath), "runs", def.Name)
	}

	fmt.Printf("corpos-lab: running study %q (%s, %d conditions × %d) on %s\n",
		def.Name, def.Assay, len(def.Conditions), def.RunsPerCell, def.Network)

	runResult, err := control.RunStudy(context.Background(), def, workDir, control.Deps{
		Launcher:   podmanLauncher{},
		DigestOf:   digestReader,
		Substrate:  substrateProbe,
		Provenance: provenanceReader(defPath),
		Preflight:  preflightProbe(defPath),
	})
	// Preflight warnings are recorded on the run; surface them to the operator
	// too, on success or failure.
	for _, w := range runResult.Warnings {
		fmt.Fprintf(os.Stderr, "corpos-lab: WARN: %s\n", w)
	}
	recordPath := filepath.Join(workDir, "run-record.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: study FAILED: %v\n", err)
		fmt.Fprintf(os.Stderr, "corpos-lab: failed-run record -> %s\n", recordPath)
		return 1
	}

	fmt.Printf("corpos-lab: study %s — %d rows, image %s\n",
		runResult.Status, len(runResult.Results.Rows), runResult.ImageDigest[:19])
	fmt.Printf("corpos-lab: run record -> %s\n", recordPath)

	// Persist to the toolkit for one queryable home (best-effort: the local
	// run record is the durable artifact, so a persist failure is a warning,
	// not a run failure). Raw responses stay on disk; only the pointer is sent.
	if toolkitURL != "" {
		responsesDir := filepath.Join(workDir, "out", "responses")
		client := persist.NewClient(toolkitURL, project)
		if err := client.Record(context.Background(), runResult, responsesDir); err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: WARN: persist to toolkit failed (run record on disk is unaffected): %v\n", err)
		} else {
			fmt.Printf("corpos-lab: persisted to toolkit (%s, project %s)\n", toolkitURL, project)
		}
	}
	return 0
}

// digestReader reads an image content digest via rootless podman.
//
// When the ref is digest-pinned and the read fails — the shape that means a
// rebuild orphaned the pinned digest — it returns an *image.AbsentPinError that
// names the missing digest and the digests currently present for the repo,
// instead of the opaque podman "exit status 125" the raw failure produces. The
// enrichment logic lives in internal/image (tested); this stays exec glue.
func digestReader(ctx context.Context, ref string) (string, error) {
	digest, err := image.Digest(ctx, podmanInspectDigest, ref)
	if err == nil {
		return digest, nil
	}
	repo, missing, ok := image.ParsePinnedDigest(ref)
	if !ok {
		return "", err
	}
	// Best-effort: list what IS present for the repo. A failure here (podman
	// missing, no such repo) just yields an empty Available, which the error
	// message renders as "rebuild it".
	var available []string
	if out, listErr := exec.CommandContext(ctx, "podman", "images", "--format", "{{.Digest}}", repo).Output(); listErr == nil {
		available = image.ParseDigestList(out)
	}
	return "", &image.AbsentPinError{Ref: ref, Repo: repo, Digest: missing, Available: available}
}

// podmanInspectDigest reads one image's content digest via rootless podman.
func podmanInspectDigest(ctx context.Context, ref string) ([]byte, error) {
	return exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Digest}}", ref).Output()
}

// preflightProbe warns about two "the code is not what you think" mismatches
// before a run, without ever blocking it:
//
//   - a pinned assay IMAGE older than the assay/runner code it should carry, so a
//     run on it silently executes stale code; and
//   - a host BINARY older than the source it was built from, the failure mode
//     from suggestion 169 (an old binary rejects a study's condition with a bare
//     "unknown condition").
//
// Every arm that cannot read a signal contributes no warning rather than
// guessing — observe, don't assert. The comparison logic lives in internal/image
// and internal/provenance (tested); this is exec glue.
func preflightProbe(defPath string) control.PreflightProbe {
	return func(ctx context.Context, def study.Def) []string {
		repoDir, err := filepath.Abs(filepath.Dir(defPath))
		if err != nil {
			return nil
		}
		var warnings []string

		// Image staleness: pinned image vs the code it runs.
		if built, err := image.Created(ctx, podmanInspectCreated, def.Image); err == nil {
			if changed, err := provenance.LastChange(ctx, repoDir, "internal/assay", "internal/runner"); err == nil {
				if msg, stale := image.StaleImageWarning(def.Image, built, changed); stale {
					warnings = append(warnings, msg)
				}
			}
		}

		// Binary staleness: this host binary vs the host-side source that defines
		// conditions and the run path.
		if bt := binaryBuildTime(); !bt.IsZero() {
			if changed, err := provenance.LastChange(ctx, repoDir, "internal", "cmd/corpos-lab"); err == nil {
				if msg, stale := provenance.BinaryStaleness(bt, changed); stale {
					warnings = append(warnings, msg)
				}
			}
		}

		// Scenario self-containment: a scenario that points at content absent from
		// the prompt makes its runs unscoreable (suggestion 170). Read the scenario
		// material and lint its text; a read failure yields no warning — the
		// materialize step reports a genuinely missing file loudly on its own.
		if sc := def.Materials.Scenario; sc != "" {
			p := sc
			if !filepath.IsAbs(p) {
				p = filepath.Join(filepath.Dir(defPath), p)
			}
			if raw, err := os.ReadFile(p); err == nil {
				warnings = append(warnings, study.ScenarioSelfContainmentWarnings(string(raw))...)
			}
		}

		return warnings
	}
}

// binaryBuildTime reads the running binary's VCS commit time from its embedded
// build info. It returns a zero time when the info is absent — a `go run` build
// or a test binary carries no vcs.time — so the caller treats it as "cannot
// tell" and stays silent.
func binaryBuildTime() time.Time {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return time.Time{}
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.time" {
			t, err := time.Parse(time.RFC3339, s.Value)
			if err != nil {
				return time.Time{}
			}
			return t
		}
	}
	return time.Time{}
}

// podmanInspectCreated reads one image's build time (epoch seconds) via rootless
// podman.
func podmanInspectCreated(ctx context.Context, ref string) ([]byte, error) {
	return exec.CommandContext(ctx, "podman", "image", "inspect", "--format", "{{.Created.Unix}}", ref).Output()
}

// substrateProbe identifies the processor the inference server is running on.
func substrateProbe(ctx context.Context) substrate.Info {
	return substrate.Probe(ctx, substrate.ExecRunner)
}

// provenanceReader stamps the repo the study definition lives in — the tree
// that carries both the instrument and the materials.
//
// A caveat worth stating plainly rather than leaving for someone to discover:
// this stamps the repo at defPath, which is the instrument's repo when the
// documented invocation is used (run from the checkout) and is NOT when the
// binary is run from elsewhere against a def somewhere else. The robust form is
// a build-time stamp (-ldflags -X main.commit=$(git rev-parse HEAD)), which
// describes the binary regardless of where it runs. This is the honest
// available approximation, and it records its own failure rather than
// inventing a commit.
func provenanceReader(defPath string) control.ProvenanceReader {
	return func(ctx context.Context) (provenance.Stamp, error) {
		dir, err := filepath.Abs(filepath.Dir(defPath))
		if err != nil {
			return provenance.Stamp{}, err
		}
		return provenance.Capture(ctx, dir)
	}
}

// podmanLauncher runs one disposable assay container with rootless podman.
type podmanLauncher struct{}

func (podmanLauncher) Launch(ctx context.Context, spec control.LaunchSpec) (control.LaunchResult, error) {
	if err := os.MkdirAll(spec.OutDir, 0o755); err != nil {
		return control.LaunchResult{}, fmt.Errorf("corpos-lab: create out dir: %w", err)
	}
	// Argv assembly lives in control.PodmanArgs, not here: cmd/ has no tests and
	// the gate's coverage floor scopes to ./internal/..., so an inline argv was
	// the one translation in the tree with nowhere to assert on it. This
	// launcher keeps only the exec concern.
	argv, err := control.PodmanArgs(spec)
	if err != nil {
		return control.LaunchResult{}, err
	}
	cmd := exec.CommandContext(ctx, "podman", argv...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		return control.LaunchResult{ExitCode: 0}, nil
	}
	// A non-zero container exit is a classified run failure, not a launcher
	// error — return the exit code so the controller records it.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return control.LaunchResult{ExitCode: exitErr.ExitCode(), Stderr: strings.TrimSpace(stderr.String())}, nil
	}
	// podman itself could not run (missing binary, bad network name, etc.).
	return control.LaunchResult{}, fmt.Errorf("corpos-lab: podman run: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
}
