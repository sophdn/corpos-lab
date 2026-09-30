package rater

import "testing"

// The characterization net for the pure scoring floor. Every case that the
// reference Python raters (and their regression tests) distinguished is
// reproduced here, plus the word-boundary and alternation-order cases the
// reference regex implied but did not test.

func TestGroundedCode(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		reasoning string
		want      string
	}{
		{"first token of content wins", "Ii — recognizes but no action", "", "Ii"},
		{"first match over a later one", "C. Not Ii.", "", "C"},
		{"empty content falls back to last reasoning match",
			"", "I need to decide. It could be Ii, but the action is performed. Final: C", "C"},
		{"missing content uses reasoning", "", "... therefore N", "N"},
		{"both empty is blank", "", "", ""},
		{"no code anywhere is blank", "no verdict here", "", ""},
		{"content with a code beats reasoning", "C", "N", "C"},
		{"content no code, reasoning last match", "prose only", "first C then finally N", "N"},
		{"bare I", "I", "", "I"},
		{"bare N", "N", "", "N"},
		{"Ic whole word", "Ic", "", "Ic"},
		{"code inside a longer word does not match", "Icon of the aid", "", ""},
		{"reasoning-only single code", "", "N", "N"},
		{"content code found before a reasoning fallback would run", "the code is I here", "C", "I"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := GroundedCode(tc.content, tc.reasoning); got != tc.want {
				t.Fatalf("GroundedCode(%q, %q) = %q, want %q", tc.content, tc.reasoning, got, tc.want)
			}
		})
	}
}

func TestActionVerdict(t *testing.T) {
	cases := []struct {
		name  string
		reply string
		want  string
	}{
		{"empty is unscoreable", "", "unscoreable"},
		{"bare A_local", "A_local", "A_local"},
		{"lowercase a_canon normalizes", "a_canon", "A_canon"},
		{"uppercase NEITHER normalizes", "NEITHER", "neither"},
		{"explicit unscoreable", "unscoreable", "unscoreable"},
		{"verdict inside a sentence", "The model edits v2, so A_canon.", "A_canon"},
		{"no verdict token is unscoreable", "it does something else", "unscoreable"},
		{"first verdict wins", "A_local, not A_canon", "A_local"},
		{"mixed case a_LoCaL", "a_LoCaL", "A_local"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ActionVerdict(tc.reply); got != tc.want {
				t.Fatalf("ActionVerdict(%q) = %q, want %q", tc.reply, got, tc.want)
			}
		})
	}
}

func TestEffectiveMaxTokens(t *testing.T) {
	n16, n32 := 16, 32
	cases := []struct {
		name     string
		explicit *int
		hosted   bool
		want     int
	}{
		{"explicit wins local", &n16, false, 16},
		{"explicit wins hosted", &n32, true, 32},
		{"local default is tight", nil, false, 16},
		{"hosted default is generous", nil, true, 2048},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EffectiveMaxTokens(tc.explicit, tc.hosted); got != tc.want {
				t.Fatalf("EffectiveMaxTokens(%v, %v) = %d, want %d", tc.explicit, tc.hosted, got, tc.want)
			}
		})
	}
}
