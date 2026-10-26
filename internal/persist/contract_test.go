package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/control"
	"corpos-lab/internal/model"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/substrate"
)

// contractFixture is the committed sample study_run_record params body. It is
// the shared conformance fixture for the wire contract between this package
// (recordParams / scoreRow, the producer) and the private toolkit service's
// go/internal/measure/study_run.go (studyRunInput / studyRunRowInput, the
// consumer). The two structs are separate implementations in separate Go
// modules; this file is what binds them.
//
// The canonical copy lives in the private toolkit service at
// go/internal/measure/testdata/study_run_record.contract.json, and that repo's
// study-run-contract-drift gate fails while this copy differs from it on
// corpos-lab origin/main. Change the contract here and there together: land
// this side first, then the toolkit side.
const contractFixture = "testdata/study_run_record.contract.json"

// freeFormMaps are paths whose value is a map with caller-chosen keys, not a
// struct. Their key sets are data, not contract, so the walk stops there.
var freeFormMaps = map[string]bool{
	"materials_hashes": true,
	"server.defaults":  true,
}

// keyPaths returns every object key path in v ("sampler.top_k",
// "rows[].verdict_kind"), stopping at freeFormMaps. Array elements share one
// "[]" segment, so the result is the union of their key sets.
func keyPaths(prefix string, v any, out map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			out[p] = true
			if !freeFormMaps[p] {
				keyPaths(p, child, out)
			}
		}
	case []any:
		for _, child := range x {
			keyPaths(prefix+"[]", child, out)
		}
	}
}

func keySet(t *testing.T, raw []byte) []string {
	t.Helper()
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	set := map[string]bool{}
	keyPaths("", v, set)
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// fullRun sets every field paramsFrom reads to a non-zero value, so every
// omitempty key is present on the wire. The fixture is the complete key set;
// a run that leaves an optional field empty sends a subset of it.
func fullRun() control.StudyRun {
	run := sampleRun()
	run.Status = control.StatusFailed
	run.Error = "container exited 1"
	run.ProvenanceError = "stamp partial: describe failed"
	run.Provenance = provenance.Stamp{CommitSHA: "abc123", Dirty: true}
	run.Substrate = substrate.Info{
		Kind: substrate.KindGPU, Device: "NVIDIA GeForce RTX 3090",
		VRAMMiB: 21184, DetectedBy: "nvidia-smi", Note: "probe note",
	}
	run.Results.Sampler = assay.Sampling{
		Temperature: 0.8, Seeds: []int{11, 12}, MaxTokens: 1024,
		TopNSigma: -1, TopK: 40, TypicalP: 1, TopP: 0.95, MinP: 0.05,
		RepeatPenalty: 1.1, RepeatLastN: 64, PresencePenalty: 0.1, FrequencyPenalty: 0.2,
		XTCProbability: 0.3, DryMultiplier: 0.4, Stop: []string{"OBSERVATION:"},
	}
	run.Results.Server = model.ServerProps{
		ModelPath: "/models/q.gguf", ModelAlias: "q", BuildInfo: "b1", NCtx: 32768,
		Defaults: map[string]any{"temperature": 0.8},
	}
	run.Results.ModelMismatch = "declared X; served Y"
	return run
}

// TestParamsWireMatchesContractFixture pins the producer half of the
// study_run_record contract: a fully populated run marshals to exactly the
// fixture's key set (names and nesting). A field renamed, added, or dropped
// here fails this test until the fixture, and with it the private toolkit service, is
// updated to match.
func TestParamsWireMatchesContractFixture(t *testing.T) {
	fixture, err := os.ReadFile(filepath.FromSlash(contractFixture))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	wire, err := json.Marshal(paramsFrom(fullRun(), "/runs/responses"))
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	want := keySet(t, fixture)
	got := keySet(t, wire)
	wantSet := map[string]bool{}
	for _, k := range want {
		wantSet[k] = true
	}
	gotSet := map[string]bool{}
	for _, k := range got {
		gotSet[k] = true
	}
	for _, k := range got {
		if !wantSet[k] {
			t.Errorf("lab sends %q, which the contract fixture does not have", k)
		}
	}
	for _, k := range want {
		if !gotSet[k] {
			t.Errorf("contract fixture has %q, which lab does not send", k)
		}
	}
}

// TestCompletedRunOmitsEmptyError pins the agreed `error,omitempty` tag: a run
// with no error sends no error key. The toolkit treats an absent error and an
// empty one the same (both store ""), so the key carries information only
// when it is present.
func TestCompletedRunOmitsEmptyError(t *testing.T) {
	wire, err := json.Marshal(paramsFrom(sampleRun(), "/runs/responses"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(wire, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["error"]; ok {
		t.Errorf("completed run sent an error key: %s", m["error"])
	}
}
