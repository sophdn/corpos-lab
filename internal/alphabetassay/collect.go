package alphabetassay

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// DefaultCollectSeed is collect_assay.py's random.seed(20260916).
const DefaultCollectSeed = 20260916

// Item is one blind-rater packet line: an opaque id and the response text.
type Item struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// RawResponse is one parsed grid response, in the deterministic (sorted) order
// the caller discovered it. The caller supplies the raw file text; CollectAll
// applies the strip.
type RawResponse struct {
	Scenario  int
	Model     string
	Condition string
	Seed      int
	Text      string
}

// ClassCollect is one class's packet: the private key (id -> metadata) and the
// shuffled blind items.
type ClassCollect struct {
	Keys  map[string]KeyEntry
	Items []Item
}

// runDirRe and condSeedRe are compiled per call in the parse helpers below; the
// run-dir pattern depends on the class name, so it cannot be a package var.

// ParseRunDir extracts the scenario and model from a
// asy-<class>-s<N>-<model>/out/responses directory path, matching
// collect_assay.py's re.search. ok is false when the path does not match.
func ParseRunDir(class, dirPath string) (scenario int, model string, ok bool) {
	re := regexp.MustCompile(`asy-` + regexp.QuoteMeta(class) + `-s(\d+)-(\w+)/out/responses`)
	m := re.FindStringSubmatch(dirPath)
	if m == nil {
		return 0, "", false
	}
	sc, _ := strconv.Atoi(m[1]) // m[1] is \d+, so Atoi cannot fail
	return sc, m[2], true
}

var condSeedRe = regexp.MustCompile(`^(.+)_(\d+)$`)

// ParseCondSeed splits a response filename base (no ".txt") into its condition
// and seed, matching collect_assay.py's re.match(r"(.+)_(\d+)$", base).
func ParseCondSeed(base string) (cond string, seed int, ok bool) {
	m := condSeedRe.FindStringSubmatch(base)
	if m == nil {
		return "", 0, false
	}
	sd, _ := strconv.Atoi(m[2]) // m[2] is \d+, so Atoi cannot fail
	return m[1], sd, true
}

// CollectAll ports collect_assay.py's per-class packet build. byClass maps each
// class to its responses in discovery order; ids are assigned in that order as
// "<class[:5]>-<NNNNN>" before a single seeded shuffle of the items (the RNG is
// seeded once and consumed across classes in Classes order, mirroring the python).
//
// Because Go's math/rand is not python's Mersenne Twister, the shuffled item
// ORDER differs from the python output; the item SET and the key mapping are
// identical, which is the structural-parity contract for this randomized tool.
func CollectAll(byClass map[string][]RawResponse, seed int64) map[string]ClassCollect {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic study shuffle, not security
	out := make(map[string]ClassCollect, len(Classes))
	for _, cls := range Classes {
		resps := byClass[cls]
		keys := make(map[string]KeyEntry, len(resps))
		items := make([]Item, 0, len(resps))
		for _, r := range resps {
			rid := ridFor(cls, len(keys))
			keys[rid] = KeyEntry{
				Class:     cls,
				Scenario:  r.Scenario,
				Model:     r.Model,
				Condition: r.Condition,
				Seed:      r.Seed,
			}
			items = append(items, Item{ID: rid, Text: strings.TrimSpace(r.Text)})
		}
		rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		out[cls] = ClassCollect{Keys: keys, Items: items}
	}
	return out
}

// ridFor builds the opaque id "<class[:5]>-<NNNNN>" (collect_assay.py's
// f"{cls[:5]}-{len(key):05d}"). Class names are ASCII, so a 5-byte prefix is a
// 5-rune prefix.
func ridFor(class string, n int) string {
	prefix := class
	if len(prefix) > 5 {
		prefix = prefix[:5]
	}
	return prefix + "-" + pad5(n)
}

func pad5(n int) string {
	s := strconv.Itoa(n)
	if len(s) >= 5 {
		return s
	}
	return strings.Repeat("0", 5-len(s)) + s
}
