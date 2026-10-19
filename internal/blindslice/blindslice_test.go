package blindslice

import (
	"fmt"
	"testing"
)

func TestOid(t *testing.T) {
	// Values computed with python's hashlib, matching build_slices.py; the first
	// is a committed key.json id.
	cases := []struct {
		cls, sc, model, cond string
		run                  int
		want                 string
	}{
		{"governed-operation-protocol-bypass", "1", "mistral", "baseline", 1, "r28e2ffec7842"},
		{"casg-direct", "1", "mistral", "baseline", 1, "r8d70816448d3"},
		{"parent-state-check-bypass", "2", "phi4", "glyph_only", 5, "r5aad7292d4f6"},
	}
	for _, c := range cases {
		if got := Oid(c.cls, c.sc, c.model, c.cond, c.run); got != c.want {
			t.Errorf("Oid(%s,%s,%s,%s,%d) = %q, want %q", c.cls, c.sc, c.model, c.cond, c.run, got, c.want)
		}
	}
}

func TestChunk(t *testing.T) {
	var got []string
	Chunk([]int{1, 2, 3, 4, 5}, 2, func(i int, c []int) { got = append(got, fmt.Sprint(i, c)) })
	if fmt.Sprint(got) != "[0 [1 2] 1 [3 4] 2 [5]]" {
		t.Errorf("Chunk = %v", got)
	}
	Chunk([]int{}, 2, func(int, []int) { t.Error("Chunk called fn for an empty slice") })
}
