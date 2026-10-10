package groundedaid

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseRunDir(t *testing.T) {
	cls, sc, model, ok := ParseRunDir("gnp-post-write-verification-absent-s1-mistral")
	if !ok || cls != "post-write-verification-absent" || sc != "1" || model != "mistral" {
		t.Errorf("ParseRunDir = (%q,%q,%q,%v)", cls, sc, model, ok)
	}
	if _, _, _, ok := ParseRunDir("not-a-run-dir"); ok {
		t.Errorf("ParseRunDir(nonmatch) ok = true, want false")
	}
}

func TestOidMatchesCommittedScheme(t *testing.T) {
	// Cross-check against a committed key.json id: r28e2ffec7842 is
	// governed-operation-protocol-bypass | 1 | mistral | baseline | run 1.
	got := Oid("governed-operation-protocol-bypass", "1", "mistral", "baseline", 1)
	if got != "r28e2ffec7842" {
		t.Errorf("Oid = %q, want r28e2ffec7842", got)
	}
	if len(got) != 13 {
		t.Errorf("oid length = %d, want 13", len(got))
	}
}

func TestBuildSlices(t *testing.T) {
	rows := []RawRow{
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "baseline", Run: 1, Text: "t1"},
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "baseline", Run: 2, Text: "t2"},
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "ground_only", Run: 1, Text: "t3"},
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "ground_only", Run: 2, Text: "t4"},
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "ground_only", Run: 3, Text: "t5"},
	}
	res := BuildSlices(rows, 2, DefaultSeed)

	if res.Total != 5 {
		t.Errorf("total = %d, want 5", res.Total)
	}
	if len(res.Key) != 5 {
		t.Errorf("key entries = %d, want 5", len(res.Key))
	}
	oid := Oid("post-write-verification-absent", "1", "mistral", "baseline", 1)
	if got := res.Key[oid]; got != (KeyEntry{Cls: "post-write-verification-absent", Scenario: "1", Model: "mistral", Condition: "baseline", Run: 1}) {
		t.Errorf("key[%s] = %+v", oid, got)
	}
	if res.Manifest["post-write-verification-absent"]["baseline"] != 2 ||
		res.Manifest["post-write-verification-absent"]["ground_only"] != 3 {
		t.Errorf("manifest = %+v", res.Manifest["post-write-verification-absent"])
	}
	if res.PerClassCount["post-write-verification-absent"] != 5 {
		t.Errorf("per-class count = %d, want 5", res.PerClassCount["post-write-verification-absent"])
	}

	// 5 items at per-slice 2 -> slices of sizes 2,2,1 for this class.
	var sizes []int
	for _, s := range res.Slices {
		if s.Class == "post-write-verification-absent" {
			sizes = append(sizes, len(s.Items))
		}
	}
	if !reflect.DeepEqual(sizes, []int{2, 2, 1}) {
		t.Errorf("slice sizes = %v, want [2 2 1]", sizes)
	}

	// Deterministic for a fixed seed.
	res2 := BuildSlices(rows, 2, DefaultSeed)
	if !reflect.DeepEqual(res.Slices, res2.Slices) {
		t.Errorf("slice order not deterministic for a fixed seed")
	}
}

func TestBuildSlicesPerSliceClamp(t *testing.T) {
	rows := []RawRow{
		{Cls: "parent-state-check-bypass", Sc: "1", Model: "phi4", Cond: "baseline", Run: 1, Text: "a"},
		{Cls: "parent-state-check-bypass", Sc: "1", Model: "phi4", Cond: "baseline", Run: 2, Text: "b"},
	}
	res := BuildSlices(rows, 0, DefaultSeed) // clamped to 1 -> one item per slice
	n := 0
	for _, s := range res.Slices {
		if len(s.Items) != 1 {
			t.Errorf("clamped slice has %d items, want 1", len(s.Items))
		}
		n++
	}
	if n != 2 {
		t.Errorf("slice count = %d, want 2", n)
	}
}

func TestFormatSummary(t *testing.T) {
	rows := []RawRow{
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "baseline", Run: 1, Text: "a"},
		{Cls: "post-write-verification-absent", Sc: "1", Model: "mistral", Cond: "ground_only", Run: 1, Text: "b"},
	}
	res := BuildSlices(rows, 96, DefaultSeed)
	out := FormatSummary(res)
	if !strings.Contains(out, "wrote 2 items across 3 classes\n") {
		t.Errorf("summary head = %q", out)
	}
	// dict-repr preserves condition first-appearance order.
	if !strings.Contains(out, "  post-write-verification-absent: 2 items — {'baseline': 1, 'ground_only': 1}\n") {
		t.Errorf("summary manifest repr wrong:\n%s", out)
	}
	// An empty class prints {}.
	if !strings.Contains(out, "  parent-state-check-bypass: 0 items — {}\n") {
		t.Errorf("empty class repr wrong:\n%s", out)
	}
}

func TestFormatSummaryEmpty(t *testing.T) {
	res := BuildSlices(nil, 96, DefaultSeed)
	out := FormatSummary(res)
	if !strings.Contains(out, "wrote 0 items across 0 classes\n") {
		t.Errorf("empty summary = %q", out)
	}
}
