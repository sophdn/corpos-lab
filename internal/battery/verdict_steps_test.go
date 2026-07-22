package battery

import (
	"context"
	"errors"
	"strings"
	"testing"

	"corpos-lab/internal/model"
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

// labelledCalibrationEntry is an entry whose structural fields are otherwise
// project-agnostic, with its ONLY project-specific vocabulary ("parent
// initiative document") sitting inside a violation-signal calibration instance
// labelled "recognition illustration — does not define scope". The 2026-04-23
// ruling voided the reading that such a label exempts the instance from
// Sub-check A: this entry must be assessed FAIL, not PASS.
const labelledCalibrationEntry = "**Y — Decision terrain**\n" +
	"The agent closes an operation while a dependent state check remains undone.\n" +
	"**Violation signal:** the dependent state is never verified before close.\n" +
	"Calibration instance (recognition illustration — does not define scope): " +
	"the agent proceeds without the parent initiative document loaded.\n"

// calibrationVocabSentinel appears in labelledCalibrationEntry only inside the
// labelled calibration instance.
const calibrationVocabSentinel = "parent initiative document"

// rulingOperativePhrase is the load-bearing sentence the repaired prompt must
// carry; if it is dropped the runner reverts to the interpretation the ruling
// voided.
const rulingOperativePhrase = "does not exempt the instance from this scan"

// rulingAwareClient stands in for a model that applies Sub-check A exactly as
// far as the prompt instructs. It fails an entry carrying project-specific
// vocabulary inside a labelled calibration instance only when the prompt tells
// it such labels grant no exemption — so if the ruling sentence is removed from
// the prompt it reverts to PASS, flipping the FAIL assertion below.
type rulingAwareClient struct{ prompts []string }

func (c *rulingAwareClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.prompts = append(c.prompts, prompt)
	rulingPresent := strings.Contains(prompt, rulingOperativePhrase)
	entryHasCalibrationVocab := strings.Contains(prompt, calibrationVocabSentinel)
	if rulingPresent && entryHasCalibrationVocab {
		return model.Response{Text: "FAIL project-specific vocabulary in a labelled calibration instance"}, nil
	}
	return model.Response{Text: "PASS"}, nil
}

func (c *rulingAwareClient) Name() string    { return "ruling-aware-fake" }
func (c *rulingAwareClient) Version() string { return "0.0.0" }
func (c *rulingAwareClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func TestItem9PromptCarriesCalibrationInstanceRuling(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	st := &State{Content: labelledCalibrationEntry, Model: f}
	Item9Universality(context.Background(), st)
	if len(f.prompts) != 1 {
		t.Fatalf("expected one prompt, got %d", len(f.prompts))
	}
	prompt := f.prompts[0]
	if !strings.Contains(prompt, calibrationVocabSentinel) {
		t.Fatal("prompt should embed the entry under evaluation")
	}
	for _, want := range []string{"recognition illustration", rulingOperativePhrase, "still a FAIL"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt should carry the 2026-04-23 ruling; missing %q", want)
		}
	}
}

func TestItem9PromptScopesOutFalloutProfileMetadata(t *testing.T) {
	// The entry carries a **Fallout profile:** reference (Item 15's referent);
	// it is a metadata pointer, not one of Item 9's structural fields, so the
	// prompt must tell the assessor to ignore its path — otherwise the model
	// flags corpus plumbing as a project-specific reference.
	f := &fakeClient{text: "PASS"}
	st := &State{Content: "**Fallout profile:** ../fallout-profiles/x.md\n**Glyph:** g\n", Model: f}
	Item9Universality(context.Background(), st)
	if len(f.prompts) != 1 {
		t.Fatalf("expected one prompt, got %d", len(f.prompts))
	}
	if !strings.Contains(f.prompts[0], "**Fallout profile:**") ||
		!strings.Contains(f.prompts[0], "metadata pointer, not a structural field") {
		t.Fatal("prompt should instruct the assessor to ignore the fallout-profile metadata reference")
	}
}

func TestItem9AssessesLabelledCalibrationInstanceAsFail(t *testing.T) {
	st := &State{Content: labelledCalibrationEntry, Model: &rulingAwareClient{}}
	out := Item9Universality(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("labelled calibration instance must be assessed FAIL, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 9 {
		t.Fatalf("expected item 9, got %v", out.Verdict.Item)
	}
}

func TestVerdictGenParamsPinDeterministicSampling(t *testing.T) {
	p := verdictGenParams()
	if p.Temperature == nil || *p.Temperature != 0.0 {
		t.Fatalf("temperature should pin 0.0, got %v", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 2048 {
		t.Fatalf("max tokens should give a thinking model room (2048), got %v", p.MaxTokens)
	}
}
