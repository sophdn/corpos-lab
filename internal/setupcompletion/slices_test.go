package setupcompletion

import (
	"sort"
	"testing"
)

func TestOpaqueIDReference(t *testing.T) {
	// Anchored to the python oracle: opaque_id("rawcasg-direct","baseline",1,"hello world\n").
	got := OpaqueID("rawcasg-direct", "baseline", 1, "hello world\n")
	want := "473e777cc6b3206a"
	if got != want {
		t.Errorf("OpaqueID = %q, want %q", got, want)
	}
	if len(got) != 16 {
		t.Errorf("id length = %d, want 16", len(got))
	}
}

func TestOpaqueIDFieldsMatter(t *testing.T) {
	base := OpaqueID("rawg", "baseline", 1, "t")
	if OpaqueID("loopg", "baseline", 1, "t") == base {
		t.Errorf("setup+glyph change did not change id")
	}
	if OpaqueID("rawg", "glyph_only", 1, "t") == base {
		t.Errorf("condition change did not change id")
	}
	if OpaqueID("rawg", "baseline", 2, "t") == base {
		t.Errorf("run change did not change id")
	}
	if OpaqueID("rawg", "baseline", 1, "u") == base {
		t.Errorf("text change did not change id")
	}
}

func TestBuildSlices(t *testing.T) {
	recs := []RunRecord{
		{Glyph: "b", Setup: "raw", Condition: "baseline", Run: 1, Text: "t1"},
		{Glyph: "a", Setup: "loop", Condition: "glyph_only", Run: 2, Text: "t2"},
		{Glyph: "a", Setup: "raw", Condition: "imperative_only", Run: 3, Text: "t3"},
	}
	glyphs, perGlyph, key := BuildSlices(recs)
	if len(glyphs) != 2 || glyphs[0] != "a" || glyphs[1] != "b" {
		t.Errorf("glyphs = %v, want sorted [a b]", glyphs)
	}
	if len(perGlyph["a"]) != 2 || len(perGlyph["b"]) != 1 {
		t.Errorf("per-glyph counts wrong: %v", perGlyph)
	}
	// key maps each id back to full coordinates.
	if len(key) != 3 {
		t.Errorf("key has %d entries, want 3", len(key))
	}
	for _, it := range perGlyph["a"] {
		e, ok := key[it.ID]
		if !ok || e.Glyph != "a" {
			t.Errorf("key entry for %q wrong: %+v ok=%v", it.ID, e, ok)
		}
	}
	// items within a glyph are sorted by id.
	as := perGlyph["a"]
	if !sort.SliceIsSorted(as, func(i, j int) bool { return as[i].ID < as[j].ID }) {
		t.Errorf("glyph 'a' items not id-sorted: %+v", as)
	}
}

func TestBuildSlicesEmpty(t *testing.T) {
	glyphs, perGlyph, key := BuildSlices(nil)
	if len(glyphs) != 0 || len(perGlyph) != 0 || len(key) != 0 {
		t.Errorf("empty input should yield empty output")
	}
}

func TestSetupFromStudy(t *testing.T) {
	cases := []struct {
		study string
		setup string
		ok    bool
	}{
		{"setup-raw-casg-direct-qwen38", "raw", true},
		{"setup-loop-parent-state-check-bypass-qwen38", "loop", true},
		{"setup-other-x-qwen38", "", false},
		{"random", "", false},
	}
	for _, c := range cases {
		s, ok := SetupFromStudy(c.study)
		if s != c.setup || ok != c.ok {
			t.Errorf("SetupFromStudy(%q) = %q,%v; want %q,%v", c.study, s, ok, c.setup, c.ok)
		}
	}
}

func TestIsSmokeStudy(t *testing.T) {
	for _, s := range []string{"setup-raw-x-smoke-qwen38", "sm-anything"} {
		if !IsSmokeStudy(s) {
			t.Errorf("IsSmokeStudy(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"setup-raw-casg-direct-qwen38", "setup-loop-x-qwen38"} {
		if IsSmokeStudy(s) {
			t.Errorf("IsSmokeStudy(%q) = true, want false", s)
		}
	}
}

func TestRunFromStem(t *testing.T) {
	cases := []struct {
		stem string
		run  int
	}{
		{"baseline_3", 3},
		{"glyph_only_5", 5},
		{"imperative_only_16", 16},
	}
	for _, c := range cases {
		got, err := RunFromStem(c.stem)
		if err != nil || got != c.run {
			t.Errorf("RunFromStem(%q) = %d,%v; want %d", c.stem, got, err, c.run)
		}
	}
	if _, err := RunFromStem("nounderscore"); err == nil {
		t.Errorf("expected error for stem with no underscore")
	}
	if _, err := RunFromStem("baseline_x"); err == nil {
		t.Errorf("expected error for non-integer run")
	}
}
