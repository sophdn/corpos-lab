package setupcompletion

import (
	"strings"
	"testing"
)

func TestTurn1PromptShape(t *testing.T) {
	p := Turn1Prompt()
	if !strings.HasPrefix(p, "<|im_start|>user\n") {
		t.Errorf("prompt does not start with the user turn opener")
	}
	if !strings.HasSuffix(p, "<|im_end|>\n<|im_start|>assistant\n<think>\n\n</think>\n\n") {
		t.Errorf("prompt does not end with the empty-think assistant opener")
	}
	if !strings.Contains(p, Turn1Preamble) {
		t.Errorf("prompt does not embed Turn1Preamble")
	}
	// The turn-1 edit_file line is the interactive-open form.
	if !strings.Contains(p, "edit_file <path>      - open a file to write (you will be prompted for content)") {
		t.Errorf("turn-1 edit_file line missing/altered")
	}
}

func TestTurn2PromptShape(t *testing.T) {
	p := Turn2Prompt()
	if !strings.Contains(p, Turn2Preamble) {
		t.Errorf("prompt does not embed Turn2Preamble")
	}
	// The turn-2 edit_file line is the inline-overwrite form.
	if !strings.Contains(p, "edit_file <path> ||| <new full contents>   - overwrite a file") {
		t.Errorf("turn-2 edit_file line missing/altered")
	}
	// The simulated transcript is present, with the out-of-date changelog.
	for _, want := range []string{
		"\nCALL list_files .",
		"\nOBSERVATION: CHANGELOG.md  filters.py  tests/  README.md",
		"\nCALL read_file CHANGELOG.md",
		"\nOBSERVATION:\n# Changelog\n\n## v1.4.0\n- Added BaseFilter\n",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("transcript missing segment %q", want)
		}
	}
}

func TestParamsPinned(t *testing.T) {
	if Turn1Params.NPredict != 200 || Turn1Params.Seed != 1 {
		t.Errorf("turn1 params = %+v", Turn1Params)
	}
	if Turn2Params.NPredict != 300 || Turn2Params.Seed != 3 {
		t.Errorf("turn2 params = %+v", Turn2Params)
	}
	// Shared sampler stages.
	for _, p := range []CompletionParams{Turn1Params, Turn2Params} {
		if p.Temperature != 0.8 || p.TopK != 0 || p.TopP != 1.0 || p.MinP != 0.05 ||
			p.RepeatPenalty != 1.0 || p.CachePrompt {
			t.Errorf("shared sampler drift: %+v", p)
		}
	}
}

func TestParseCompletionResponse(t *testing.T) {
	raw := []byte(`{"content":"hi there","stopped_limit":true,"timings":{"predicted_n":200,"predicted_per_second":42.34}}`)
	r, err := ParseCompletionResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Content != "hi there" {
		t.Errorf("content = %q", r.Content)
	}
	if r.StoppedLimit == nil || !*r.StoppedLimit {
		t.Errorf("stopped_limit = %v, want true", r.StoppedLimit)
	}
	if r.PredictedN != 200 {
		t.Errorf("predicted_n = %d", r.PredictedN)
	}
	if r.TokPerSec != 42.34 {
		t.Errorf("tok/s = %v", r.TokPerSec)
	}
}

func TestParseCompletionResponseAbsentFields(t *testing.T) {
	// No stopped_limit, no timings: pointer nil, ints/floats zero.
	r, err := ParseCompletionResponse([]byte(`{"content":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	if r.StoppedLimit != nil {
		t.Errorf("stopped_limit = %v, want nil", r.StoppedLimit)
	}
	if r.PredictedN != 0 || r.TokPerSec != 0 {
		t.Errorf("expected zero timings, got n=%d tps=%v", r.PredictedN, r.TokPerSec)
	}
}

func TestParseCompletionResponseError(t *testing.T) {
	if _, err := ParseCompletionResponse([]byte("not json")); err == nil {
		t.Errorf("expected an error on malformed JSON")
	}
}

func TestFormatTurn1Report(t *testing.T) {
	tru := true
	r := CompletionResult{Content: "the answer", StoppedLimit: &tru, PredictedN: 200, TokPerSec: 42.34}
	got := FormatTurn1Report(r)
	want := "=== CONTENT ===\nthe answer\n=== stopped_limit: True pred_n: 200 tok/s: 42.3\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestFormatTurn2Report(t *testing.T) {
	fls := false
	r := CompletionResult{Content: "done", StoppedLimit: &fls, PredictedN: 300}
	got := FormatTurn2Report(r)
	want := "=== CONTENT ===\ndone\n=== stopped_limit: False pred_n: 300\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestFormatTurnReportNilStopped(t *testing.T) {
	r := CompletionResult{Content: "c", StoppedLimit: nil, PredictedN: 1, TokPerSec: 5.0}
	if got := FormatTurn1Report(r); !strings.Contains(got, "stopped_limit: None") {
		t.Errorf("nil stopped_limit not rendered as None: %q", got)
	}
	if got := FormatTurn2Report(r); !strings.Contains(got, "stopped_limit: None") {
		t.Errorf("nil stopped_limit not rendered as None: %q", got)
	}
}

func TestPyRound1(t *testing.T) {
	cases := map[float64]string{0: "0.0", 5: "5.0", 42.34: "42.3", 42.36: "42.4"}
	for in, want := range cases {
		if got := pyRound1(in); got != want {
			t.Errorf("pyRound1(%v) = %q, want %q", in, got, want)
		}
	}
}
