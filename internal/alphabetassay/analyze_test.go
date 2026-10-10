package alphabetassay

import (
	"math"
	"strings"
	"testing"
)

func TestAnalyzeBasic(t *testing.T) {
	inputs := map[string]ClassInput{
		"casg-direct": {
			Key: map[string]KeyEntry{
				"a": {Model: "qwen38", Condition: "baseline"},
				"b": {Model: "qwen38", Condition: "baseline"},
				"c": {Model: "mistral", Condition: "glyph_only"},
				"d": {Model: "qwen38", Condition: "baseline"}, // rated by A only -> unrated
			},
			ACodes: map[string]string{"a": "C", "b": "C", "c": "N", "d": "C"},
			BCodes: map[string]string{"a": "C", "b": "N", "c": "N"},
		},
	}
	res := Analyze(inputs)

	// missing: 8 other classes "no key", plus the unrated note for casg-direct.
	if got := res.Missing["casg-direct"]; got != "1 ids unrated (of 4)" {
		t.Errorf("missing note = %q", got)
	}
	if got := res.Missing["casg-delegate"]; got != "no key" {
		t.Errorf("absent class note = %q, want 'no key'", got)
	}

	cells := res.PerCell["casg-direct"]
	qb := cells["qwen38|baseline"]
	if qb.N != 2 || qb.ConsC != 1 || qb.AC != 2 || qb.BC != 1 {
		t.Errorf("qwen38|baseline cell = %+v", qb)
	}
	mg := cells["mistral|glyph_only"]
	if mg.N != 1 || mg.ConsC != 0 || mg.AC != 0 || mg.BC != 0 {
		t.Errorf("mistral|glyph_only cell = %+v", mg)
	}

	ag := res.Agreement["casg-direct"]
	if ag.N != 3 {
		t.Fatalf("agreement n = %d, want 3", ag.N)
	}
	if ag.RawAgree == nil || *ag.RawAgree != 0.667 {
		t.Errorf("raw_agree = %v, want 0.667", ag.RawAgree)
	}
	if ag.Kappa == nil || *ag.Kappa != 0.4 {
		t.Errorf("kappa = %v, want 0.4", ag.Kappa)
	}
}

func TestAnalyzeEmptyIDs(t *testing.T) {
	inputs := map[string]ClassInput{
		"casg-direct": {
			Key:    map[string]KeyEntry{"x": {Model: "mistral", Condition: "baseline"}},
			ACodes: map[string]string{},
			BCodes: map[string]string{},
		},
	}
	res := Analyze(inputs)
	ag := res.Agreement["casg-direct"]
	if ag.N != 0 || ag.RawAgree != nil || ag.Kappa != nil {
		t.Errorf("empty-ids agreement = %+v, want n=0 and nil floats", ag)
	}
	if len(res.PerCell["casg-direct"]) != 0 {
		t.Errorf("per_cell should be empty for a class with no jointly-rated ids")
	}
	if res.Missing["casg-direct"] != "1 ids unrated (of 1)" {
		t.Errorf("missing = %q", res.Missing["casg-direct"])
	}
}

func TestAnalyzeKappaUndefined(t *testing.T) {
	// Both raters code the single label C for every id: pe == 1, kappa undefined (NaN).
	inputs := map[string]ClassInput{
		"casg-direct": {
			Key:    map[string]KeyEntry{"a": {Model: "m", Condition: "baseline"}, "b": {Model: "m", Condition: "baseline"}},
			ACodes: map[string]string{"a": "C", "b": "C"},
			BCodes: map[string]string{"a": "C", "b": "C"},
		},
	}
	res := Analyze(inputs)
	ag := res.Agreement["casg-direct"]
	if ag.RawAgree == nil || *ag.RawAgree != 1.0 {
		t.Errorf("raw_agree = %v, want 1.0", ag.RawAgree)
	}
	if ag.Kappa == nil || !math.IsNaN(*ag.Kappa) {
		t.Errorf("kappa = %v, want NaN", ag.Kappa)
	}
}

func TestKappaZeroIDs(t *testing.T) {
	if got := kappa(map[string]string{}, map[string]string{}, nil); !math.IsNaN(got) {
		t.Errorf("kappa(no ids) = %v, want NaN", got)
	}
}

func TestFormatTableNormalAndMissing(t *testing.T) {
	inputs := map[string]ClassInput{
		"casg-direct": {
			Key:    map[string]KeyEntry{"a": {Model: "qwen38", Condition: "baseline"}, "b": {Model: "qwen38", Condition: "baseline"}},
			ACodes: map[string]string{"a": "C", "b": "N"},
			BCodes: map[string]string{"a": "C", "b": "N"},
		},
	}
	out := FormatTable(Analyze(inputs))
	if !strings.HasPrefix(out, "STRICT-CONSENSUS C  (both raters C)   consC/N per condition\n\n") {
		t.Errorf("missing header:\n%s", out)
	}
	if !strings.Contains(out, "## casg-direct   agree=1.0 kappa=1.0 n=2") {
		t.Errorf("class heading wrong:\n%s", out)
	}
	// qwen38 row: baseline shows 1/2 (one consensus C of two).
	if !strings.Contains(out, "  qwen38            1/2") {
		t.Errorf("qwen38 row wrong:\n%s", out)
	}
	// A class with no key prints MISSING.
	if !strings.Contains(out, "## casg-delegate: MISSING") {
		t.Errorf("expected MISSING line:\n%s", out)
	}
	// The missing/unrated summary line lists the no-key classes in Classes order.
	if !strings.Contains(out, `MISSING/UNRATED: {"formal-step-context-bypass": "no key"`) {
		t.Errorf("missing/unrated line wrong:\n%s", out)
	}
}

func TestFormatTableNoneAgreement(t *testing.T) {
	inputs := map[string]ClassInput{
		"casg-direct": {
			Key: map[string]KeyEntry{"x": {Model: "mistral", Condition: "baseline"}},
		},
	}
	out := FormatTable(Analyze(inputs))
	if !strings.Contains(out, "## casg-direct   agree=None kappa=None n=0") {
		t.Errorf("None agreement not rendered:\n%s", out)
	}
}

func TestPyFloatRepr(t *testing.T) {
	cases := map[float64]string{
		0.948: "0.948", 0.9: "0.9", 1.0: "1.0", 0.0: "0.0", -0.5: "-0.5",
		math.NaN(): "nan", math.Inf(1): "inf", math.Inf(-1): "-inf",
	}
	for in, want := range cases {
		if got := pyFloatRepr(in); got != want {
			t.Errorf("pyFloatRepr(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestRoundN(t *testing.T) {
	cases := []struct {
		in   float64
		n    int
		want float64
	}{
		{0.6666666, 3, 0.667},
		{0.9999, 3, 1.0},
		{0.125, 2, 0.12}, // round half to even
	}
	for _, c := range cases {
		if got := roundN(c.in, c.n); got != c.want {
			t.Errorf("roundN(%v,%d) = %v, want %v", c.in, c.n, got, c.want)
		}
	}
	if got := roundN(math.NaN(), 3); !math.IsNaN(got) {
		t.Errorf("roundN(NaN) = %v, want NaN", got)
	}
}

func TestRateStrAndPad(t *testing.T) {
	if got := rateStr(Cell{N: 0}); got != "-" {
		t.Errorf("rateStr(n=0) = %q, want -", got)
	}
	if got := rateStr(Cell{N: 8, ConsC: 5}); got != "5/8" {
		t.Errorf("rateStr = %q, want 5/8", got)
	}
	if got := trunc("abcdef", 3); got != "abc" {
		t.Errorf("trunc = %q", got)
	}
	if got := trunc("ab", 3); got != "ab" {
		t.Errorf("trunc short = %q", got)
	}
	if got := ljust("x", 3); got != "x  " {
		t.Errorf("ljust = %q", got)
	}
	if got := ljust("xyz", 2); got != "xyz" {
		t.Errorf("ljust no-op = %q", got)
	}
	if got := rjust("x", 3); got != "  x" {
		t.Errorf("rjust = %q", got)
	}
	if got := rjust("xyz", 2); got != "xyz" {
		t.Errorf("rjust no-op = %q", got)
	}
}

func TestOptFloatNil(t *testing.T) {
	if got := optFloat(nil); got != "None" {
		t.Errorf("optFloat(nil) = %q", got)
	}
	v := 0.5
	if got := optFloat(&v); got != "0.5" {
		t.Errorf("optFloat(0.5) = %q", got)
	}
}
