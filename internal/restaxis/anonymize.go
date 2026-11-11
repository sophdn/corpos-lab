package restaxis

import (
	"fmt"
	"strings"
)

// AnonResponse is one arm/seed response for a scenario, as read from the runs tree.
type AnonResponse struct {
	Arm  string
	Seed int
	Text string
}

// AnonScenario is one scenario's material for a blind bundle: the task prompt, its
// ground-truth completion, and the responses present (in read order). A caller
// includes a scenario exactly when its responses directory exists, mirroring
// anonymize.py's `if not rdir.exists(): continue` — the responses slice may be
// empty when the directory exists but holds no arm files.
type AnonScenario struct {
	Scenario    int
	Task        string
	GroundTruth string
	Responses   []AnonResponse
}

// AnonEntry is the operator-held de-anonymization record for one response id. It
// mirrors the python map.json value {scenario, arm, seed}.
type AnonEntry struct {
	Scenario int    `json:"scenario"`
	Arm      string `json:"arm"`
	Seed     int    `json:"seed"`
}

// AnonBundle is an arm-blind scoring bundle plus the separate id->arm keymap.
// Markdown is the rater-facing sheet (responses by opaque id, no arm named); Map is
// the operator-held id -> {scenario, arm, seed} that never enters the sheet.
type AnonBundle struct {
	Markdown string
	Map      map[string]AnonEntry
}

// BuildAnonBundle assembles an arm-blind scoring bundle for one (model, glyph).
// Within each scenario the responses are shuffled (via the injected shuffle) so
// their arm and seed order carries no signal, then assigned sequential ids
// R0000, R0001… across scenarios in the order the scenarios are supplied. The
// id->arm mapping is returned separately and never appears in the markdown, so a
// rater scoring by id cannot recover the arm.
//
// shuffle has math/rand's Shuffle signature and is injected so the draw is
// deterministic in a test and random in production. This does NOT reproduce
// anonymize.py's exact ordering — Go's RNG is not byte-compatible with python's
// Mersenne Twister — but the item set, the blinding, and the id->arm keymap are
// structurally identical. (Same tradeoff as actionconflict.Pilot.)
func BuildAnonBundle(scenarios []AnonScenario, shuffle func(n int, swap func(i, j int))) AnonBundle {
	mapping := map[string]AnonEntry{}
	lines := []string{
		"# Scoring bundle — score every response by its ID.\n",
		"For each RESPONSE, output exactly one label: OK, OF, OFc, I, or N.\n",
	}
	idn := 0
	for _, sc := range scenarios {
		bucket := append([]AnonResponse(nil), sc.Responses...)
		shuffle(len(bucket), func(i, j int) { bucket[i], bucket[j] = bucket[j], bucket[i] })
		lines = append(lines,
			fmt.Sprintf("\n---\n## Task group %d\n", sc.Scenario),
			"**The task given to the assistant:**\n\n"+sc.Task+"\n",
			"\n**Ground truth (correct completion + the over-fire error to watch for):**\n\n"+sc.GroundTruth+"\n",
			"\n**Responses to score:**\n",
		)
		for _, r := range bucket {
			rid := fmt.Sprintf("R%04d", idn)
			idn++
			mapping[rid] = AnonEntry{Scenario: sc.Scenario, Arm: r.Arm, Seed: r.Seed}
			lines = append(lines, fmt.Sprintf("\n### %s\n```\n%s\n```\n", rid, r.Text))
		}
	}
	return AnonBundle{Markdown: strings.Join(lines, "\n"), Map: mapping}
}
