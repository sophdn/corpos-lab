package lengthdistraction

import (
	"reflect"
	"testing"

	"corpos-lab/internal/neutralprefix"
)

func keyFrom(meta map[string]neutralprefix.KeyMeta, order []string) *neutralprefix.KeyFile {
	return &neutralprefix.KeyFile{Order: order, Meta: meta}
}

func TestUnreportedConditionsFindsDropped(t *testing.T) {
	key := keyFrom(map[string]neutralprefix.KeyMeta{
		"r0": {Cls: "a", Condition: "baseline"},
		"r1": {Cls: "a", Condition: "scrambled_glyph"},
		"r2": {Cls: "a", Condition: "off_target_glyph"},
		"r3": {Cls: "a", Condition: "neutral_prefix"},
	}, []string{"r0", "r1", "r2", "r3"})
	conds := []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}
	got := UnreportedConditions(key, conds)
	want := []string{"off_target_glyph", "scrambled_glyph"} // sorted
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestUnreportedConditionsAllReported(t *testing.T) {
	key := keyFrom(map[string]neutralprefix.KeyMeta{
		"r0": {Cls: "a", Condition: "baseline"},
		"r1": {Cls: "a", Condition: "neutral_prefix"},
	}, []string{"r0", "r1"})
	conds := []string{"baseline", "neutral_prefix", "glyph_only", "imperative_only"}
	if got := UnreportedConditions(key, conds); len(got) != 0 {
		t.Fatalf("want no unreported conditions, got %v", got)
	}
}
