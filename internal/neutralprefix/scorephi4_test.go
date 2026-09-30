package neutralprefix

import "testing"

// labelBattery is expected output captured from score_phi4.py's CODE_RE:
//
//	CODE_RE = re.compile(r"\b(Ii|Ic|C|I|N)\b")
//	m = CODE_RE.search(out); (m.group(1) if m else "N")
var labelBattery = []struct {
	in, want string
}{
	{"C", "C"},
	{"C first line\nreason", "C"},
	{"Ii", "Ii"},
	{"Ic", "Ic"},
	{"I", "I"},
	{"N", "N"},
	{"OK", "N"},
	{"", "N"},
	{"  C  ", "C"},
	{"\nC", "C"},
	{"the code is C.", "C"},
	{"label: Ii\nbecause", "Ii"},
	{"Icc", "N"},
	{"Iii", "N"},
	{"II", "N"},
	{"CC", "N"},
	{"C.", "C"},
	{"Ii.", "Ii"},
	{"xIx", "N"},
	{"x I x", "I"},
	{"score N/A", "N"},
	{"Nope", "N"},
	{"None", "N"},
	{"Incorrect", "N"},
	{"Correct", "N"},
	{"Iii Ic", "Ic"},
	{"prefix Ic done", "Ic"},
	{"1. C", "C"},
	{"- I", "I"},
	{"```\nC\n```", "C"},
	{"Ic then C", "Ic"},
	{"I i", "I"},
	{"cI", "N"},
	{"C\tI", "C"},
	{"The label is: Ic (over-fire).", "Ic"},
	{"N — not scoreable", "N"},
	{"aIi", "N"},
	{"Ii\n", "Ii"},
	{"no code here", "N"},
	{"Ci", "N"},
	{"iC", "N"},
	{"lowercase c", "N"},
	{"🎯 C", "C"},
	{"C😀", "C"},
	// N followed by a non-word boundary is a real match, not the default:
	{"verdict: N.", "N"},
	{"N first", "N"},
}

func TestParseScoreLabelBattery(t *testing.T) {
	for _, tc := range labelBattery {
		if got := ParseScoreLabel(tc.in); got != tc.want {
			t.Errorf("ParseScoreLabel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildScorePrompt(t *testing.T) {
	got := BuildScorePrompt("RUBRIC BODY", "the response")
	want := "<|im_start|>user<|im_sep|>RUBRIC BODY" +
		"\n\n---\nScore the RESPONSE below against the rubric. Output ONLY the code as " +
		"the first token: one of C, Ii, Ic, I, N. No explanation.\n\nRESPONSE:\n" +
		"the response<|im_end|><|im_start|>assistant<|im_sep|>"
	if got != want {
		t.Fatalf("BuildScorePrompt mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestClassOf(t *testing.T) {
	cases := map[string]string{
		"casg-direct__00.jsonl":                       "casg-direct",
		"/x/y/parent-state-check-bypass__11.jsonl":    "parent-state-check-bypass",
		"conditional-gate-uniform-default__07.jsonl":  "conditional-gate-uniform-default",
		"noSuffix.jsonl":                              "noSuffix.jsonl",
		`C:\win\formal-step-context-bypass__03.jsonl`: "formal-step-context-bypass",
	}
	for in, want := range cases {
		if got := ClassOf(in); got != want {
			t.Errorf("ClassOf(%q) = %q, want %q", in, got, want)
		}
	}
}
