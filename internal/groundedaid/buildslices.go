package groundedaid

import (
	"crypto/sha1" //nolint:gosec // sha1 is the study's committed opaque-id scheme, not security
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// DefaultPerSlice and DefaultSeed are build_slices.py's --per-slice and --seed
// defaults.
const (
	DefaultPerSlice = 96
	DefaultSeed     = 12345
)

// Item is one blind slice line: an opaque id and the response text.
type Item struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// RawRow is one completed run row paired with its response text, in the
// deterministic (sorted) order the caller discovered it.
type RawRow struct {
	Cls   string
	Sc    string
	Model string
	Cond  string
	Run   int
	Text  string
}

// Slice is one output slice: a class, its 0-based index, and its items.
type Slice struct {
	Class string
	Index int
	Items []Item
}

// BuildResult holds everything build_slices.py emits: the id->key map, the
// coverage manifest, the shuffled slices, and the counts for the summary line.
type BuildResult struct {
	Key           map[string]KeyEntry
	Manifest      map[string]map[string]int
	Slices        []Slice
	Total         int
	PerClassCount map[string]int

	// condOrder records each class's condition first-appearance order, so the
	// summary reproduces python's dict-repr ordering.
	condOrder map[string][]string
}

var runDirRe = regexp.MustCompile(`^gnp-(.+)-s(\d+)-([^-]+)$`)

// ParseRunDir splits a gnp-<class>-s<n>-<model> run-dir name, matching
// build_slices.py's parse_run_dir. ok is false when the name does not match.
func ParseRunDir(name string) (cls, sc, model string, ok bool) {
	m := runDirRe.FindStringSubmatch(name)
	if m == nil {
		return "", "", "", false
	}
	return m[1], m[2], m[3], true
}

// Oid is build_slices.py's opaque id: "r" + the first 12 hex chars of
// sha1("<cls>|<sc>|<model>|<cond>|<run>").
func Oid(cls, sc, model, cond string, run int) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", cls, sc, model, cond, run))) //nolint:gosec // committed id scheme
	return "r" + hex.EncodeToString(h[:])[:12]
}

// BuildSlices ports build_slices.py's key/manifest/slice build. rows are the
// completed run rows in discovery order; ids are the deterministic Oid, and each
// class's items are shuffled with a single seeded RNG (classes processed in
// Classes order) before being chunked into perSlice-sized slices.
//
// The key map and manifest are deterministic and match the python by key; the
// per-slice ORDER differs (Go RNG != python Mersenne Twister), so slice parity is
// structural (same item set).
func BuildSlices(rows []RawRow, perSlice int, seed int64) BuildResult {
	if perSlice < 1 {
		perSlice = 1
	}
	res := BuildResult{
		Key:           map[string]KeyEntry{},
		Manifest:      map[string]map[string]int{},
		PerClassCount: map[string]int{},
		condOrder:     map[string][]string{},
	}
	perClass := map[string][]Item{}
	for _, cls := range Classes {
		perClass[cls] = nil
		res.PerClassCount[cls] = 0
	}

	for _, r := range rows {
		oid := Oid(r.Cls, r.Sc, r.Model, r.Cond, r.Run)
		res.Key[oid] = KeyEntry{Cls: r.Cls, Scenario: r.Sc, Model: r.Model, Condition: r.Cond, Run: r.Run}
		perClass[r.Cls] = append(perClass[r.Cls], Item{ID: oid, Text: r.Text})
		if res.Manifest[r.Cls] == nil {
			res.Manifest[r.Cls] = map[string]int{}
		}
		if _, seen := res.Manifest[r.Cls][r.Cond]; !seen {
			res.condOrder[r.Cls] = append(res.condOrder[r.Cls], r.Cond)
		}
		res.Manifest[r.Cls][r.Cond]++
	}

	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic study shuffle, not security
	for _, cls := range Classes {
		items := perClass[cls]
		res.PerClassCount[cls] = len(items)
		rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		for i := 0; i < len(items); i += perSlice {
			end := i + perSlice
			if end > len(items) {
				end = len(items)
			}
			chunk := make([]Item, end-i)
			copy(chunk, items[i:end])
			res.Slices = append(res.Slices, Slice{Class: cls, Index: i / perSlice, Items: chunk})
			res.Total += len(chunk)
		}
	}
	return res
}

// FormatSummary reproduces build_slices.py's printed summary, including python's
// dict-repr of each class's manifest (single quotes, first-appearance order).
func FormatSummary(res BuildResult) string {
	var b strings.Builder
	across := 0
	total := 0
	for _, cls := range Classes {
		total += res.PerClassCount[cls]
	}
	if total > 0 {
		across = len(Classes)
	}
	b.WriteString(fmt.Sprintf("wrote %d items across %d classes\n", res.Total, across))
	for _, cls := range Classes {
		b.WriteString(fmt.Sprintf("  %s: %d items — %s\n", cls, res.PerClassCount[cls], pyDictRepr(res, cls)))
	}
	return b.String()
}

// pyDictRepr renders one class's manifest as python's repr(dict) would.
func pyDictRepr(res BuildResult, cls string) string {
	order := res.condOrder[cls]
	m := res.Manifest[cls]
	if len(order) == 0 || m == nil {
		return "{}"
	}
	parts := make([]string, 0, len(order))
	for _, cond := range order {
		parts = append(parts, "'"+cond+"': "+strconv.Itoa(m[cond]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
