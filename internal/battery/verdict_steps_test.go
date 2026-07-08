package battery

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Characterization tests ported from steps/battery/verdict_steps.rs
// (11 tests): the parse seam and items 1/9 against Pass/Fail/Error/Garbage
// fakes.

func TestParsePass(t *testing.T) {
	for _, text := range []string{"PASS", "  PASS\n", "pass"} {
		if v := ParseModelVerdict(text, 1); v.Kind != KindPass {
			t.Fatalf("ParseModelVerdict(%q) = %+v, want pass", text, v)
		}
	}
}

func TestParseFailWithReason(t *testing.T) {
	v := ParseModelVerdict("FAIL missing X", 3)
	if v.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", v)
	}
	if v.Item == nil || *v.Item != 3 {
		t.Fatalf("expected item 3, got %v", v.Item)
	}
	if !strings.Contains(v.Reason, "Item 3") || !strings.Contains(v.Reason, "missing X") {
		t.Fatalf("got %q", v.Reason)
	}
}

func TestParseFailWithoutReason(t *testing.T) {
	v := ParseModelVerdict("FAIL", 5)
	if v.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", v)
	}
	if v.Item == nil || *v.Item != 5 {
		t.Fatalf("expected item 5, got %v", v.Item)
	}
	if !strings.Contains(v.Reason, "unspecified") {
		t.Fatalf("got %q", v.Reason)
	}
}

func TestParseGarbageResponseBecomesFail(t *testing.T) {
	v := ParseModelVerdict("maybe", 7)
	if v.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", v)
	}
	if v.Item == nil || *v.Item != 7 {
		t.Fatalf("expected item 7, got %v", v.Item)
	}
	if !strings.Contains(v.Reason, "did not start with PASS or FAIL") {
		t.Fatalf("got %q", v.Reason)
	}
}

func TestItem1PassesWhenModelSaysPass(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{text: "PASS"}}
	out := Item1XYZSpecificity(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem1FailsWhenModelSaysFail(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{text: "FAIL X is a placeholder"}}
	out := Item1XYZSpecificity(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 1 {
		t.Fatalf("expected item 1, got %v", out.Verdict.Item)
	}
	if !strings.Contains(out.Verdict.Reason, "Item 1") || !strings.Contains(out.Verdict.Reason, "placeholder") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem1ReturnsErrorOnModelFailure(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{err: errors.New("timeout")}}
	out := Item1XYZSpecificity(context.Background(), st)
	if out.Kind != OutcomeError {
		t.Fatalf("expected error outcome, got %+v", out)
	}
	if !strings.Contains(out.Message, "Item 1: model error:") {
		t.Fatalf("got %q", out.Message)
	}
}

func TestItem1GarbageResponseMapsToFail(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{text: "maybe"}}
	out := Item1XYZSpecificity(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
}

func TestItem1PromptCarriesEntryContent(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	st := &State{Content: "UNIQUE-SENTINEL-CONTENT", Model: f}
	Item1XYZSpecificity(context.Background(), st)
	if len(f.prompts) != 1 || !strings.Contains(f.prompts[0], "UNIQUE-SENTINEL-CONTENT") {
		t.Fatal("prompt should embed the entry content")
	}
	if !strings.Contains(f.prompts[0], "structural specificity") {
		t.Fatal("prompt should carry the item-1 framing")
	}
}

func TestItem9PassesWhenModelSaysPass(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{text: "PASS"}}
	out := Item9Universality(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem9FailsWhenModelSaysFail(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{text: "FAIL contains process-docs/ALPHABET.md"}}
	out := Item9Universality(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 9 {
		t.Fatalf("expected item 9, got %v", out.Verdict.Item)
	}
	if !strings.Contains(out.Verdict.Reason, "Item 9") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem9ReturnsErrorOnModelFailure(t *testing.T) {
	st := &State{Content: "some entry", Model: &fakeClient{err: errors.New("timeout")}}
	out := Item9Universality(context.Background(), st)
	if out.Kind != OutcomeError {
		t.Fatalf("expected error outcome, got %+v", out)
	}
	if !strings.Contains(out.Message, "Item 9: model error:") {
		t.Fatalf("got %q", out.Message)
	}
}

func TestVerdictGenParamsPinDeterministicSampling(t *testing.T) {
	p := verdictGenParams()
	if p.Temperature == nil || *p.Temperature != 0.0 {
		t.Fatalf("temperature should pin 0.0, got %v", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 256 {
		t.Fatalf("max tokens should pin 256, got %v", p.MaxTokens)
	}
}
