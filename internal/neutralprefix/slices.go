package neutralprefix

import (
	"crypto/sha1" //nolint:gosec // sha1 is the committed id scheme, not a security primitive
	"encoding/hex"
	"fmt"
	"strings"
)

// slicesClasses is build_slices.py's CLASSES, in output order. per_class is
// initialized for exactly these classes, so slices are emitted class by class in
// this order.
var slicesClasses = []string{"casg-direct", "formal-step-context-bypass",
	"parent-state-check-bypass", "conditional-gate-uniform-default"}

// SliceRecord is one discovered response, in the deterministic scan order the
// caller must preserve (sorted results.json paths, then row order). Scenario, Cls,
// Model, Condition are strings and Run is the results row's integer run.
type SliceRecord struct {
	Cls       string
	Scenario  string
	Model     string
	Condition string
	Run       int
	Text      string
}

// SliceFile is one emitted blind slice: a base name (<class>__NN.jsonl) and its
// JSONL lines (each a compact {"id","text"}).
type SliceFile struct {
	Name  string
	Lines []string
}

// SlicesResult holds everything build_slices.py writes plus its stdout summary.
// Key and Manifest are ready for PyDumps (key.json is indent=0, MANIFEST.json is
// indent=1, both ensure_ascii=True).
type SlicesResult struct {
	Key      *OMap
	Manifest *OMap
	Slices   []SliceFile
	Summary  string
}

type sliceItem struct {
	oid  string
	text string
}

// oidFor computes the opaque response id: "r" + first 12 hex chars of
// sha1("<cls>|<sc>|<model>|<cond>|<run>"), matching build_slices.py.
func oidFor(cls, sc, model, cond string, run int) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", cls, sc, model, cond, run))) //nolint:gosec // committed id scheme
	return "r" + hex.EncodeToString(h[:])[:12]
}

// BuildSlices reproduces build_slices.py from an already-ordered record list, over
// the chain-548 class set (slicesClasses). It is BuildSlicesFor pinned to that set,
// kept so the committed chain-548 slices reproduce byte-identically.
func BuildSlices(records []SliceRecord, perSlice, seed int) *SlicesResult {
	return BuildSlicesFor(records, perSlice, seed, slicesClasses)
}

// BuildSlicesFor is BuildSlices over an explicit class-emission order, so a study
// with classes other than the chain-548 set (the length-distraction-typing
// extension) can build its own blind slices. It assigns each record an opaque id,
// builds the un-blinding key (in record order) and the per-class/condition manifest
// (over every class present), then emits condition-blind slices for the classes in
// `classes`, shuffled per class by one PyRandom(seed) drawn in that class order and
// chunked at perSlice. The class order is load-bearing: it drives the RNG draw
// sequence, so two runs match only when `classes` matches.
func BuildSlicesFor(records []SliceRecord, perSlice, seed int, classes []string) *SlicesResult {
	res := &SlicesResult{Key: NewOMap(), Manifest: NewOMap()}

	perClass := map[string][]sliceItem{}
	var manifestClasses []string
	manCount := map[string]*OCounter{}
	for _, rec := range records {
		oid := oidFor(rec.Cls, rec.Scenario, rec.Model, rec.Condition, rec.Run)
		res.Key.Set(oid, NewOMap().
			Set("cls", rec.Cls).
			Set("scenario", rec.Scenario).
			Set("model", rec.Model).
			Set("condition", rec.Condition).
			Set("run", rec.Run))
		perClass[rec.Cls] = append(perClass[rec.Cls], sliceItem{oid, rec.Text})
		if _, ok := manCount[rec.Cls]; !ok {
			manCount[rec.Cls] = NewOCounter()
			manifestClasses = append(manifestClasses, rec.Cls)
		}
		manCount[rec.Cls].Inc(rec.Condition)
	}
	for _, cls := range manifestClasses {
		res.Manifest.Set(cls, manCount[cls].ToOMap())
	}

	rng := NewPyRandom(uint64(seed))
	total := 0
	nClasses := 0
	for _, cls := range classes {
		items := perClass[cls]
		if len(items) > 0 {
			nClasses++
		}
		rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		for i := 0; i < len(items); i += perSlice {
			end := i + perSlice
			if end > len(items) {
				end = len(items)
			}
			chunk := items[i:end]
			sf := SliceFile{Name: fmt.Sprintf("%s__%02d.jsonl", cls, i/perSlice)}
			for _, it := range chunk {
				sf.Lines = append(sf.Lines, PyDumps(NewOMap().Set("id", it.oid).Set("text", it.text), -1, true))
			}
			res.Slices = append(res.Slices, sf)
			total += len(chunk)
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "built slices for %d responses across %d classes\n", total, nClasses)
	for _, cls := range classes {
		if oc, ok := manCount[cls]; ok {
			sum := 0
			for _, k := range oc.Keys {
				sum += oc.Get(k)
			}
			fmt.Fprintf(&sb, "  %s: %d responses %s\n", cls, sum, oc.PyDictRepr())
		}
	}
	res.Summary = sb.String()
	return res
}
