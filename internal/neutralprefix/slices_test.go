package neutralprefix

import (
	"strings"
	"testing"
)

func TestOidFor(t *testing.T) {
	// Values computed with Python's hashlib, matching build_slices.py.
	if got := oidFor("casg-direct", "1", "mistral", "baseline", 1); got != "r8d70816448d3" {
		t.Fatalf("oid = %q, want r8d70816448d3", got)
	}
	if got := oidFor("parent-state-check-bypass", "2", "phi4", "glyph_only", 5); got != "r5aad7292d4f6" {
		t.Fatalf("oid = %q, want r5aad7292d4f6", got)
	}
}

func makeRecords(cls string, conds []string, per int) []SliceRecord {
	var recs []SliceRecord
	for _, cond := range conds {
		for run := 1; run <= per; run++ {
			recs = append(recs, SliceRecord{
				Cls: cls, Scenario: "1", Model: "mistral", Condition: cond, Run: run,
				Text: cls + "-" + cond + "-resp",
			})
		}
	}
	return recs
}

func TestBuildSlicesStructure(t *testing.T) {
	recs := makeRecords("casg-direct", []string{"baseline", "neutral_prefix"}, 40)
	res := BuildSlices(recs, 32, 12345)

	// key has one entry per record, in record order.
	if res.Key.Len() != 80 {
		t.Fatalf("key len %d, want 80", res.Key.Len())
	}
	// manifest: casg-direct -> {baseline:40, neutral_prefix:40}, first-seen order.
	man := res.Manifest.vals[0].(*OMap)
	if res.Manifest.keys[0] != "casg-direct" || man.keys[0] != "baseline" || man.vals[0].(int) != 40 {
		t.Fatalf("manifest wrong: %s", PyDumps(res.Manifest, 1, true))
	}
	// 80 items at per-slice 32 -> chunks of 32, 32, 16 = 3 slice files.
	if len(res.Slices) != 3 {
		t.Fatalf("slice files %d, want 3", len(res.Slices))
	}
	names := []string{res.Slices[0].Name, res.Slices[1].Name, res.Slices[2].Name}
	want := []string{"casg-direct__00.jsonl", "casg-direct__01.jsonl", "casg-direct__02.jsonl"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("slice name[%d]=%q want %q", i, names[i], want[i])
		}
	}
	if len(res.Slices[0].Lines) != 32 || len(res.Slices[2].Lines) != 16 {
		t.Fatalf("chunk sizes wrong: %d,%d", len(res.Slices[0].Lines), len(res.Slices[2].Lines))
	}
	// every line is a compact {"id","text"} object.
	for _, l := range res.Slices[0].Lines {
		if !strings.HasPrefix(l, `{"id": "r`) || !strings.Contains(l, `"text":`) {
			t.Fatalf("bad slice line: %q", l)
		}
	}
	// summary reports total and per-class histogram.
	if !strings.Contains(res.Summary, "built slices for 80 responses across 1 classes") {
		t.Fatalf("summary total wrong: %q", res.Summary)
	}
	if !strings.Contains(res.Summary, "casg-direct: 80 responses {'baseline': 40, 'neutral_prefix': 40}") {
		t.Fatalf("summary histogram wrong: %q", res.Summary)
	}
}

func TestBuildSlicesForCustomClasses(t *testing.T) {
	// A class outside the chain-548 set: BuildSlices emits nothing for it, but
	// BuildSlicesFor with an explicit class order does. This is the exact gap the
	// length-distraction-typing study hit.
	recs := makeRecords("casg-delegate", []string{"baseline", "neutral_prefix"}, 40)

	pinned := BuildSlices(recs, 32, 12345)
	if len(pinned.Slices) != 0 {
		t.Fatalf("BuildSlices emitted %d slices for a non-548 class, want 0", len(pinned.Slices))
	}
	// key and manifest still cover the records even when no slice is emitted.
	if pinned.Key.Len() != 80 {
		t.Fatalf("key len %d, want 80 even with no slices", pinned.Key.Len())
	}

	res := BuildSlicesFor(recs, 32, 12345, []string{"casg-delegate"})
	if len(res.Slices) != 3 {
		t.Fatalf("BuildSlicesFor slice files %d, want 3", len(res.Slices))
	}
	if res.Slices[0].Name != "casg-delegate__00.jsonl" {
		t.Fatalf("slice name %q, want casg-delegate__00.jsonl", res.Slices[0].Name)
	}
	if !strings.Contains(res.Summary, "built slices for 80 responses across 1 classes") {
		t.Fatalf("summary wrong: %q", res.Summary)
	}
}

func TestBuildSlicesDeterministicShuffle(t *testing.T) {
	recs := makeRecords("casg-direct", []string{"baseline"}, 50)
	a := BuildSlices(recs, 32, 12345)
	b := BuildSlices(recs, 32, 12345)
	if a.Slices[0].Lines[0] != b.Slices[0].Lines[0] {
		t.Fatal("same seed produced different shuffle")
	}
	c := BuildSlices(recs, 32, 999)
	// A different seed almost certainly reorders; assert the first lines differ to
	// confirm the seed actually drives the shuffle (probability of collision ~1/50).
	if a.Slices[0].Lines[0] == c.Slices[0].Lines[0] && a.Slices[0].Lines[1] == c.Slices[0].Lines[1] {
		t.Fatal("different seed produced identical leading order")
	}
}

func TestBuildSlicesEmpty(t *testing.T) {
	res := BuildSlices(nil, 32, 12345)
	if res.Key.Len() != 0 || len(res.Slices) != 0 {
		t.Fatalf("empty input should yield no key/slices")
	}
	if !strings.Contains(res.Summary, "built slices for 0 responses across 0 classes") {
		t.Fatalf("empty summary wrong: %q", res.Summary)
	}
}
