package restaxis

import (
	"strings"
	"testing"
)

func TestParseCorrectCompletions(t *testing.T) {
	md := "intro text\n\n## Scenario 1\nblock one\nmore\n\n### scenario_2 (title)\nblock two\n\n# SCENARIO 10\nblock ten\n"
	got := ParseCorrectCompletions(md)
	if len(got) != 3 {
		t.Fatalf("got %d scenarios, want 3: %+v", len(got), got)
	}
	if got[1] != "block one\nmore" {
		t.Errorf("scenario 1 block = %q", got[1])
	}
	if got[2] != "block two" {
		t.Errorf("scenario 2 block = %q", got[2])
	}
	if got[10] != "block ten" {
		t.Errorf("scenario 10 block = %q", got[10])
	}
}

func TestParseCorrectCompletionsEmpty(t *testing.T) {
	if got := ParseCorrectCompletions("no headings here"); len(got) != 0 {
		t.Errorf("want empty, got %+v", got)
	}
}

func TestBuildPhi4PromptGolden(t *testing.T) {
	got := BuildPhi4Prompt("  do the task  ", "the ground truth", "  the response  ")
	// Structural anchors that must hold byte-for-byte.
	if !strings.HasPrefix(got, "<|im_start|>user<|im_sep|>You are a strict, impartial grader") {
		t.Errorf("prompt prefix wrong: %q", got[:60])
	}
	if !strings.HasSuffix(got, "explaining it.<|im_end|><|im_start|>assistant<|im_sep|>") {
		t.Errorf("prompt suffix wrong: %q", got[len(got)-60:])
	}
	// scenario and response are trimmed; gt block inserted as-is.
	if !strings.Contains(got, "THE TASK GIVEN TO THE ASSISTANT:\ndo the task\n\n") {
		t.Errorf("scenario not trimmed/placed: %q", got)
	}
	if !strings.Contains(got, "THE ASSISTANT'S RESPONSE:\nthe response\n\n") {
		t.Errorf("response not trimmed/placed: %q", got)
	}
	if !strings.Contains(got, "watch for:\nthe ground truth\n\n") {
		t.Errorf("ground-truth block not placed: %q", got)
	}
	if !strings.Contains(got, "—") { // the em-dash in "GROUND TRUTH —"
		t.Errorf("em-dash missing from prompt")
	}
}

func TestParsePhi4Label(t *testing.T) {
	cases := []struct {
		name          string
		content       string
		wantLabel     string
		wantCoerced   bool
		wantReasonHas string
	}{
		{"ok", "OK\nit does the task", "OK", false, "it does the task"},
		{"of", "OF - unwarranted gate", "OF", false, ""},
		{"okc-normalizes", "OKc\nfine", "OK", false, "fine"},
		{"ofc-quirk-becomes-N", "OFc\nover-fires and fails", "N", true, "over-fires and fails"},
		{"invalid-becomes-N", "XYZ\ngibberish", "N", true, "gibberish"},
		{"empty", "", "N", true, ""},
		{"reason-joins-lines", "OK\nline one\nline two", "OK", false, "line one line two"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			label, reason, coerced := ParsePhi4Label(c.content)
			if label != c.wantLabel {
				t.Errorf("label = %q, want %q", label, c.wantLabel)
			}
			if coerced != c.wantCoerced {
				t.Errorf("coerced = %v, want %v", coerced, c.wantCoerced)
			}
			if c.wantReasonHas != "" && reason != c.wantReasonHas {
				t.Errorf("reason = %q, want %q", reason, c.wantReasonHas)
			}
		})
	}
}

func TestParsePhi4LabelReasonCap(t *testing.T) {
	long := strings.Repeat("x", 500)
	_, reason, _ := ParsePhi4Label("OK\n" + long)
	if len(reason) != 300 {
		t.Errorf("reason length = %d, want 300 (capped)", len(reason))
	}
}
