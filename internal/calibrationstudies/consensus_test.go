package calibrationstudies

import "testing"

func TestConsensusTable(t *testing.T) {
	key := map[string]KeyEntry{
		// Cell (alpha, baseline, m1): mixed agreement, n=4.
		"id1": {Cls: "alpha", Condition: "baseline", Model: "m1"},
		"id2": {Cls: "alpha", Condition: "baseline", Model: "m1"},
		"id3": {Cls: "alpha", Condition: "baseline", Model: "m1"},
		"id4": {Cls: "alpha", Condition: "baseline", Model: "m1"},
		// Cell (alpha, baseline, m2): devstral absent but claude+deepseek agree.
		"id7": {Cls: "alpha", Condition: "baseline", Model: "m2"},
		// Cell (alpha, glyph_only, m1): unanimous C.
		"id5": {Cls: "alpha", Condition: "glyph_only", Model: "m1"},
		// Cell (beta, baseline, m1): unanimous I.
		"id6": {Cls: "beta", Condition: "baseline", Model: "m1"},
		// Cell (delta, baseline, m1): majority C without the validated pair.
		"d1": {Cls: "delta", Condition: "baseline", Model: "m1"},
		// Cell (gamma, baseline, m1): 7 of 8 C, to exercise the 0.87 floor.
		"g1": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g2": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g3": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g4": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g5": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g6": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g7": {Cls: "gamma", Condition: "baseline", Model: "m1"},
		"g8": {Cls: "gamma", Condition: "baseline", Model: "m1"},
	}
	claude := map[string]string{
		"id1": "C", "id2": "C", "id3": "C", "id4": "I",
		"id7": "C", "id5": "C", "id6": "I", "d1": "I",
		"g1": "C", "g2": "C", "g3": "C", "g4": "C", "g5": "C", "g6": "C", "g7": "C", "g8": "I",
	}
	deepseek := map[string]string{
		"id1": "C", "id2": "C", "id3": "I", "id4": "Ii",
		"id7": "C", "id5": "C", "id6": "I", "d1": "C",
		"g1": "C", "g2": "C", "g3": "C", "g4": "C", "g5": "C", "g6": "C", "g7": "C", "g8": "I",
	}
	devstral := map[string]string{
		"id1": "C", "id2": "I", "id3": "I", "id4": "N",
		// id7 absent on purpose.
		"id5": "C", "id6": "I", "d1": "C",
		"g1": "C", "g2": "C", "g3": "C", "g4": "C", "g5": "C", "g6": "C", "g7": "C", "g8": "I",
	}

	want := "cls\tcond\tmodel\tn\tC_maj\trate_maj\tC_cd\trate_cd\n" +
		"alpha\tbaseline\tm1\t4\t2\t0.5\t2\t0.5\n" +
		"alpha\tbaseline\tm2\t1\t1\t1\t1\t1\n" +
		"alpha\tglyph_only\tm1\t1\t1\t1\t1\t1\n" +
		"beta\tbaseline\tm1\t1\t0\t0\t0\t0\n" +
		"delta\tbaseline\tm1\t1\t1\t1\t0\t0\n" +
		"gamma\tbaseline\tm1\t8\t7\t0.87\t7\t0.87\n"

	got := ConsensusTable(key, claude, deepseek, devstral)
	if got != want {
		t.Errorf("ConsensusTable mismatch:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFloorRate(t *testing.T) {
	cases := []struct {
		num, denom int
		want       string
	}{
		{0, 8, "0"},
		{8, 8, "1"},
		{4, 8, "0.5"},
		{7, 8, "0.87"},
		{1, 8, "0.12"},
		{3, 8, "0.37"},
		{5, 0, "0"}, // guard: no division by zero
	}
	for _, c := range cases {
		if got := floorRate(c.num, c.denom); got != c.want {
			t.Errorf("floorRate(%d,%d)=%q, want %q", c.num, c.denom, got, c.want)
		}
	}
}
