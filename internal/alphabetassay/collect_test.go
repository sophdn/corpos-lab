package alphabetassay

import (
	"reflect"
	"testing"
)

func TestParseRunDir(t *testing.T) {
	sc, model, ok := ParseRunDir("casg-direct", "x/casg-direct/runs/asy-casg-direct-s3-qwen38/out/responses")
	if !ok || sc != 3 || model != "qwen38" {
		t.Errorf("ParseRunDir = (%d,%q,%v), want (3,qwen38,true)", sc, model, ok)
	}
	if _, _, ok := ParseRunDir("casg-direct", "nope"); ok {
		t.Errorf("ParseRunDir(nonmatch) ok = true, want false")
	}
}

func TestParseCondSeed(t *testing.T) {
	cases := []struct {
		base string
		cond string
		seed int
		ok   bool
	}{
		{"baseline_10", "baseline", 10, true},
		{"domain_imperative_only_3", "domain_imperative_only", 3, true},
		{"noseed", "", 0, false},
	}
	for _, c := range cases {
		cond, seed, ok := ParseCondSeed(c.base)
		if cond != c.cond || seed != c.seed || ok != c.ok {
			t.Errorf("ParseCondSeed(%q) = (%q,%d,%v), want (%q,%d,%v)", c.base, cond, seed, ok, c.cond, c.seed, c.ok)
		}
	}
}

func TestCollectAll(t *testing.T) {
	byClass := map[string][]RawResponse{
		"casg-direct": {
			{Scenario: 1, Model: "mistral", Condition: "baseline", Seed: 1, Text: "  hello  \n"},
			{Scenario: 1, Model: "mistral", Condition: "baseline", Seed: 2, Text: "world"},
			{Scenario: 2, Model: "phi4", Condition: "glyph_only", Seed: 1, Text: "z"},
		},
	}
	out := CollectAll(byClass, DefaultCollectSeed)
	cc := out["casg-direct"]

	// Keys: ids assigned in input order as "casg--0000N" before the shuffle.
	if len(cc.Keys) != 3 {
		t.Fatalf("keys = %d, want 3", len(cc.Keys))
	}
	first := cc.Keys["casg--00000"]
	if first != (KeyEntry{Class: "casg-direct", Scenario: 1, Model: "mistral", Condition: "baseline", Seed: 1}) {
		t.Errorf("id casg--00000 = %+v", first)
	}
	if _, ok := cc.Keys["casg--00002"]; !ok {
		t.Errorf("expected id casg--00002")
	}

	// Items: same set as keys, text stripped.
	gotIDs := map[string]string{}
	for _, it := range cc.Items {
		gotIDs[it.ID] = it.Text
	}
	if gotIDs["casg--00000"] != "hello" {
		t.Errorf("text not stripped: %q", gotIDs["casg--00000"])
	}
	if len(gotIDs) != 3 {
		t.Errorf("items = %d, want 3", len(gotIDs))
	}

	// Deterministic: same seed -> identical item order.
	out2 := CollectAll(byClass, DefaultCollectSeed)
	if !reflect.DeepEqual(cc.Items, out2["casg-direct"].Items) {
		t.Errorf("shuffle not deterministic for a fixed seed")
	}

	// Classes with no responses get empty packets.
	if len(out["casg-delegate"].Items) != 0 {
		t.Errorf("absent class should yield no items")
	}
}

func TestRidFor(t *testing.T) {
	if got := ridFor("casg-direct", 18); got != "casg--00018" {
		t.Errorf("ridFor long class = %q", got)
	}
	if got := ridFor("abc", 5); got != "abc-00005" {
		t.Errorf("ridFor short class = %q", got)
	}
	if got := pad5(123456); got != "123456" {
		t.Errorf("pad5 overflow = %q", got)
	}
}
