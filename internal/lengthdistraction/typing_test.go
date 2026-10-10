package lengthdistraction

import (
	"strings"
	"testing"

	"corpos-lab/internal/neutralprefix"
)

// buildKey makes a KeyFile whose ids encode (class, condition, run) and records
// them in the given order. Each entry is {id, cls, cond}.
func buildKey(entries [][3]string) *neutralprefix.KeyFile {
	kf := &neutralprefix.KeyFile{Meta: map[string]neutralprefix.KeyMeta{}}
	for i, e := range entries {
		id := e[0]
		kf.Order = append(kf.Order, id)
		kf.Meta[id] = neutralprefix.KeyMeta{
			Cls:       e[1],
			Condition: e[2],
			Scenario:  "1",
			Model:     "mistral",
			Run:       i + 1,
		}
	}
	return kf
}

func TestTallyStrictConsensus(t *testing.T) {
	key := buildKey([][3]string{
		{"a1", "cx", "neutral_prefix"},
		{"a2", "cx", "neutral_prefix"},
		{"a3", "cx", "neutral_prefix"},
		{"a4", "cx", "neutral_prefix"},
		{"b1", "cx", "baseline"},          // different condition, must be excluded
		{"z1", "other", "neutral_prefix"}, // different class, must be excluded
	})
	a := map[string]string{"a1": "N", "a2": "N", "a3": "C", "a4": "N", "b1": "C", "z1": "N"}
	b := map[string]string{"a1": "N", "a2": "Ii", "a3": "C", "a4": "", "b1": "C", "z1": "N"}

	d := Tally(key, a, b, true, "cx", "neutral_prefix")
	if d.Total != 4 {
		t.Fatalf("Total = %d, want 4 (only cx/neutral_prefix)", d.Total)
	}
	if d.N != 1 { // a1 both N; a2 split; a3 both C; a4 b blank -> split
		t.Errorf("N = %d, want 1", d.N)
	}
	if d.C != 1 {
		t.Errorf("C = %d, want 1", d.C)
	}
	if d.Split != 2 {
		t.Errorf("Split = %d, want 2 (a2 disagree, a4 blank)", d.Split)
	}
	if got := d.NRate(); got != 0.25 {
		t.Errorf("NRate = %v, want 0.25", got)
	}
	if got := d.CRate(); got != 0.25 {
		t.Errorf("CRate = %v, want 0.25", got)
	}
}

func TestTallySingleRater(t *testing.T) {
	key := buildKey([][3]string{
		{"a1", "cx", "neutral_prefix"},
		{"a2", "cx", "neutral_prefix"},
		{"a3", "cx", "neutral_prefix"},
	})
	a := map[string]string{"a1": "N", "a2": "Ii"} // a3 missing -> unscored "?"
	d := Tally(key, a, nil, false, "cx", "neutral_prefix")
	if d.Total != 3 {
		t.Fatalf("Total = %d, want 3", d.Total)
	}
	if d.N != 1 || d.Ii != 1 || d.Unscored != 1 {
		t.Errorf("got N=%d Ii=%d Unscored=%d, want 1/1/1", d.N, d.Ii, d.Unscored)
	}
}

func TestAllCodesFold(t *testing.T) {
	key := buildKey([][3]string{
		{"c1", "cx", "baseline"},
		{"ii1", "cx", "baseline"},
		{"ic1", "cx", "baseline"},
		{"i1", "cx", "baseline"},
		{"n1", "cx", "baseline"},
	})
	a := map[string]string{"c1": "C", "ii1": "Ii", "ic1": "Ic", "i1": "I", "n1": "N"}
	d := Tally(key, a, a, true, "cx", "baseline")
	if d.C != 1 || d.Ii != 1 || d.Ic != 1 || d.I != 1 || d.N != 1 || d.Total != 5 {
		t.Errorf("distribution wrong: %+v", d)
	}
}

func TestRateZeroTotal(t *testing.T) {
	var d Dist
	if d.NRate() != 0 || d.CRate() != 0 {
		t.Errorf("empty dist rates = %v/%v, want 0/0", d.NRate(), d.CRate())
	}
}

func TestType(t *testing.T) {
	hiN, loN := 0.30, 0.10
	cases := []struct {
		name       string
		neuN, impN float64
		want       Verdict
	}{
		{"derails-clean-imperative", 0.60, 0.05, IsLengthDistraction},
		{"at-thresholds", 0.30, 0.10, IsLengthDistraction},
		{"neutral-does-not-derail", 0.20, 0.00, NotLengthDistraction},
		{"both-derail", 0.60, 0.40, Inconclusive},
	}
	for _, c := range cases {
		neu := Dist{N: int(c.neuN * 100), Total: 100}
		imp := Dist{N: int(c.impN * 100), Total: 100}
		if got := Type(neu, imp, hiN, loN); got != c.want {
			t.Errorf("%s: Type = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestClassesOrderAndSorted(t *testing.T) {
	key := buildKey([][3]string{
		{"x1", "zebra", "baseline"},
		{"x2", "alpha", "baseline"},
		{"x3", "zebra", "neutral_prefix"},
		{"x4", "alpha", "neutral_prefix"},
	})
	got := Classes(key)
	if len(got) != 2 || got[0] != "zebra" || got[1] != "alpha" {
		t.Errorf("Classes first-seen = %v, want [zebra alpha]", got)
	}
	sorted := SortedClasses(key)
	if sorted[0] != "alpha" || sorted[1] != "zebra" {
		t.Errorf("SortedClasses = %v, want [alpha zebra]", sorted)
	}
}

func TestReport(t *testing.T) {
	// One length-distraction class and one comprehension class.
	var entries [][3]string
	add := func(cls, cond string, n int, id string) {
		for i := 0; i < n; i++ {
			entries = append(entries, [3]string{id + string(rune('a'+i)), cls, cond})
		}
	}
	add("ld", "neutral_prefix", 4, "ldn")
	add("ld", "imperative_only", 4, "ldi")
	add("comp", "neutral_prefix", 4, "cn")
	add("comp", "imperative_only", 4, "ci")
	key := buildKey(entries)

	a := map[string]string{}
	// ld: neutral all N (derails), imperative all Ii (clean) -> length-distraction
	for _, id := range key.Order {
		m := key.Meta[id]
		switch {
		case m.Cls == "ld" && m.Condition == "neutral_prefix":
			a[id] = "N"
		case m.Cls == "ld" && m.Condition == "imperative_only":
			a[id] = "Ii"
		case m.Cls == "comp" && m.Condition == "neutral_prefix":
			a[id] = "C" // does not derail
		case m.Cls == "comp" && m.Condition == "imperative_only":
			a[id] = "Ii"
		}
	}

	conds := []string{"neutral_prefix", "imperative_only"}
	rep := Report(key, a, a, true, []string{"ld", "comp"}, conds, 0.30, 0.10)

	if !strings.Contains(rep, "strict-consensus") {
		t.Errorf("report missing consensus marker:\n%s", rep)
	}
	if !strings.Contains(rep, "ld") || !strings.Contains(rep, "length-distraction") {
		t.Errorf("report missing ld verdict:\n%s", rep)
	}
	if !strings.Contains(rep, "not-length-distraction") {
		t.Errorf("report missing comp verdict:\n%s", rep)
	}
	// Single-rater header path.
	repA := Report(key, a, nil, false, []string{"ld"}, conds, 0.30, 0.10)
	if !strings.Contains(repA, "rater-a") {
		t.Errorf("single-rater report missing rater-a marker:\n%s", repA)
	}
}

func TestMajority(t *testing.T) {
	cases := []struct {
		name   string
		id     string
		raters []map[string]string
		want   string
	}{
		{"two-of-three agree", "x",
			[]map[string]string{{"x": "N"}, {"x": "N"}, {"x": "C"}}, "N"},
		{"all three agree", "x",
			[]map[string]string{{"x": "C"}, {"x": "C"}, {"x": "C"}}, "C"},
		{"three-way tie is split", "x",
			[]map[string]string{{"x": "C"}, {"x": "Ii"}, {"x": "N"}}, CodeSplit},
		{"two agree one missing", "x",
			[]map[string]string{{"x": "N"}, {"x": "N"}, {}}, "N"},
		{"one vote of three is not a majority", "x",
			[]map[string]string{{"x": "N"}, {}, {}}, CodeSplit},
		{"no raters", "x", nil, CodeSplit},
		{"two raters agree", "x",
			[]map[string]string{{"x": "C"}, {"x": "C"}}, "C"},
		{"two raters disagree is split", "x",
			[]map[string]string{{"x": "C"}, {"x": "N"}}, CodeSplit},
	}
	for _, c := range cases {
		if got := Majority(c.id, c.raters...); got != c.want {
			t.Errorf("%s: Majority = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMajorityMapThenTally(t *testing.T) {
	key := buildKey([][3]string{
		{"a1", "cx", "neutral_prefix"},
		{"a2", "cx", "neutral_prefix"},
		{"a3", "cx", "neutral_prefix"},
	})
	// a1: 3x N -> N; a2: N,N,C -> N; a3: C,Ii,N -> split.
	r1 := map[string]string{"a1": "N", "a2": "N", "a3": "C"}
	r2 := map[string]string{"a1": "N", "a2": "N", "a3": "Ii"}
	r3 := map[string]string{"a1": "N", "a2": "C", "a3": "N"}
	maj := MajorityMap(key, r1, r2, r3)
	d := Tally(key, maj, maj, true, "cx", "neutral_prefix")
	if d.N != 2 || d.Split != 1 || d.Total != 3 {
		t.Errorf("majority tally = %+v, want N=2 Split=1 Total=3", d)
	}
}

func TestLjustAndF42(t *testing.T) {
	if got := ljust("ab", 5); got != "ab   " {
		t.Errorf("ljust short = %q", got)
	}
	if got := ljust("abcdef", 3); got != "abcdef " {
		t.Errorf("ljust overflow = %q, want one trailing space", got)
	}
	if got := f42(0.5); got != "0.50" {
		t.Errorf("f42 = %q, want 0.50", got)
	}
}
