package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"corpos-lab/internal/restaxis"
)

func init() { registerRestAxisMode("anonymize", raAnonymize) }

// scenarioHdrRe matches a "## Scenario N …" heading in a CORRECT_COMPLETIONS.md
// file. Ported from the python parse_completions regex.
var scenarioHdrRe = regexp.MustCompile(`(?im)^#{1,3}\s*scenario[ _]?(\d+)\b.*$`)

// parseRestAxisCompletions maps a scenario number to its ground-truth body. It
// splits the file on the scenario headings and trims each following section, the
// same way python's re.split + strip does. Shared by the anonymize and calib modes.
func parseRestAxisCompletions(text string) map[int]string {
	out := map[int]string{}
	locs := scenarioHdrRe.FindAllStringSubmatchIndex(text, -1)
	for i, loc := range locs {
		n, err := strconv.Atoi(text[loc[2]:loc[3]])
		if err != nil {
			continue
		}
		bodyEnd := len(text)
		if i+1 < len(locs) {
			bodyEnd = locs[i+1][0]
		}
		out[n] = strings.TrimSpace(text[loc[1]:bodyEnd])
	}
	return out
}

// raAnonymize builds an arm-blind scoring bundle for one (model, glyph): it groups
// each scenario's responses, shuffles them so the arm/seed order is hidden, assigns
// opaque ids, and writes bundle.md (given to the rater) plus map.json (the
// operator-held id -> {scenario, arm, seed}). Ported from anonymize.py.
//
// Usage: rest-axis anonymize <study-dir> <model> <glyph> --out <dir> [--seed <n>]
func raAnonymize(args []string) error {
	out, args := flagValue(args, "--out")
	seedStr, args := flagValue(args, "--seed")
	if out == "" {
		return fmt.Errorf("need --out <dir> (where to write <model>.<glyph>/{bundle.md,map.json})")
	}
	if len(args) != 3 {
		return fmt.Errorf("need <study-dir> <model> <glyph>")
	}
	study, model, glyph := args[0], args[1], args[2]
	if seedStr == "" {
		seedStr = "1234" // anonymize.py's default seed_rng
	}

	comp, err := os.ReadFile(filepath.Join(study, glyph, "CORRECT_COMPLETIONS.md")) //nolint:gosec // a study path
	if err != nil {
		return err
	}
	gt := parseRestAxisCompletions(string(comp))

	root := filepath.Join(study, "runs", model, glyph)
	scens, err := scenarioDirs(root)
	if err != nil {
		return err
	}
	var scenarios []restaxis.AnonScenario
	for _, snum := range scens {
		rdir := filepath.Join(root, fmt.Sprintf("s%d", snum), "out", "responses")
		if _, err := os.Stat(rdir); err != nil {
			continue // mirrors python's `if not rdir.exists(): continue`
		}
		task := ""
		if b, err := os.ReadFile(filepath.Join(study, glyph, "materials", fmt.Sprintf("scenario_%d.md", snum))); err == nil { //nolint:gosec // a study path
			task = strings.TrimSpace(string(b))
		}
		sc := restaxis.AnonScenario{Scenario: snum, Task: task, GroundTruth: gt[snum]}
		for _, arm := range restaxis.Arms {
			for seed := 1; seed <= 3; seed++ {
				p := filepath.Join(rdir, fmt.Sprintf("%s_%d.txt", arm, seed))
				b, err := os.ReadFile(p) //nolint:gosec // a study response path
				if err != nil {
					continue
				}
				sc.Responses = append(sc.Responses, restaxis.AnonResponse{
					Arm: arm, Seed: seed, Text: strings.TrimSpace(string(b)),
				})
			}
		}
		scenarios = append(scenarios, sc)
	}

	rng := rand.New(rand.NewSource(restAxisSeed(model, glyph, seedStr))) //nolint:gosec // not cryptographic: a reproducible blinding shuffle
	bundle := restaxis.BuildAnonBundle(scenarios, rng.Shuffle)

	d := filepath.Join(out, model+"."+glyph)
	if err := os.MkdirAll(d, 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(d, "bundle.md"), []byte(bundle.Markdown), 0o600); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(d, "map.json"), bundle.Map); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s/%s: bundle with %d responses -> %s\n", model, glyph, len(bundle.Map), filepath.Join(d, "bundle.md"))
	return nil
}

// scenarioDirs lists the scenario numbers under a runs/<model>/<glyph> tree, from
// directory names like "s14", sorted ascending. Mirrors anonymize.py's
// `sorted(..., key=lambda p: int(p.name[1:]))`.
func scenarioDirs(root string) ([]int, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var nums []int
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "s") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "s"))
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	sort.Ints(nums)
	return nums, nil
}

// restAxisSeed derives a stable int64 seed from the model, glyph, and seed string.
// anonymize.py seeds its RNG from the string "<model>/<glyph>/<seed>" so each
// (model, glyph) shuffles differently but reproducibly; this keeps that property
// without claiming byte parity with python's Mersenne Twister.
func restAxisSeed(model, glyph, seed string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(model + "/" + glyph + "/" + seed))
	return int64(h.Sum64()) //nolint:gosec // deliberate wraparound to seed math/rand
}
