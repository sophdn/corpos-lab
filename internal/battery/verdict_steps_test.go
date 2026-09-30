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

// The item-1 repair (task 3588): the prompt must tell the assessor to resolve
// X and Y from the entry's decision point and firing condition before judging
// them, and must say that the canonical "Taking X from Y" notation is not
// itself a failure. Without this the assessor reads the two symbols as the
// components and fails every entry in the corpus at step 0. The companion
// clause — an unresolvable component still FAILS — keeps the item from going
// vacuous; both are pinned so a prompt edit that drops either one is caught.
func TestItem1PromptDemandsResolvedReading(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	st := &State{Content: "entry", Model: f}
	Item1XYZSpecificity(context.Background(), st)
	prompt := f.prompts[0]
	for _, want := range []string{
		"RESOLVE BEFORE JUDGING",
		"Taking X from Y",
		"not the components themselves",
		"decision point and firing condition",
		"judge the RESOLVED X, Y, and Z",
		"no resolvable definition for a component anywhere",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("item-1 prompt must carry %q (task 3588 resolved-reading repair)", want)
		}
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

// --- Items 6, 11, 12: model-assessed verdict steps ---

// verdictStepCase drives the shared Pass/Fail/Error/Garbage assertions across
// the three new verdict steps so each item's polarity and item-number tagging is
// checked without repeating the boilerplate.
type verdictStepCase struct {
	name string
	item int
	fn   func(context.Context, *State) StepOutcome
	// framing is a phrase the item's prompt must carry (item identity).
	framing string
}

func TestNewVerdictSteps(t *testing.T) {
	cases := []verdictStepCase{
		{"item6", 6, Item6EntryCoherence, "retrosynthetic"},
		{"item11", 11, Item11SafetyClass, "safety-class boundaries"},
	}
	for _, c := range cases {
		t.Run(c.name+"/pass", func(t *testing.T) {
			out := c.fn(context.Background(), &State{Content: "entry", Model: &fakeClient{text: "PASS"}})
			if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
				t.Fatalf("expected pass, got %+v", out)
			}
		})
		t.Run(c.name+"/fail", func(t *testing.T) {
			out := c.fn(context.Background(), &State{Content: "entry", Model: &fakeClient{text: "FAIL because reasons"}})
			if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
				t.Fatalf("expected fail, got %+v", out)
			}
			if out.Verdict.Item == nil || *out.Verdict.Item != c.item {
				t.Fatalf("expected item %d, got %v", c.item, out.Verdict.Item)
			}
		})
		t.Run(c.name+"/error", func(t *testing.T) {
			out := c.fn(context.Background(), &State{Content: "entry", Model: &fakeClient{err: errors.New("timeout")}})
			if out.Kind != OutcomeError {
				t.Fatalf("expected error outcome, got %+v", out)
			}
			if !strings.Contains(out.Message, "model error:") {
				t.Fatalf("got %q", out.Message)
			}
		})
		t.Run(c.name+"/garbage", func(t *testing.T) {
			out := c.fn(context.Background(), &State{Content: "entry", Model: &fakeClient{text: "maybe"}})
			if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
				t.Fatalf("garbage should map to fail, got %+v", out)
			}
		})
		t.Run(c.name+"/prompt", func(t *testing.T) {
			f := &fakeClient{text: "PASS"}
			c.fn(context.Background(), &State{Content: "UNIQUE-SENTINEL-6-11-12", Model: f})
			if len(f.prompts) != 1 || !strings.Contains(f.prompts[0], "UNIQUE-SENTINEL-6-11-12") {
				t.Fatal("prompt should embed the entry content")
			}
			if !strings.Contains(f.prompts[0], c.framing) {
				t.Fatalf("prompt should carry the item framing %q", c.framing)
			}
		})
	}
}

// Item 12 is no longer a model-assessed verdict step — its empirical rebuild
// (bug 1332) and the tests for it live in baseline_test.go.

// --- Item 6: inline reference delivery (bug 1334 + suggestion 175) ---

// restAxisFormRule is the Rest-axis Form statement from GLYPH_DEFINITION.md — the
// rule the item-6 Qwen judge hallucinated as missing in bug 1334 ("no rule for
// the absent-preconditions / Y-neutral territory") while it was in fact present
// in the definition. A test reference doc must carry it so the judge can trace
// the Rest field against it.
const restAxisFormRule = "In Y-neutral, neither [Marker pull] nor [Aim channel] is active — [distinguishing condition]."

// testGlyphDefinition is a minimal stand-in for GLYPH_DEFINITION.md carrying the
// Rest-axis Form rule. It is inlined by tests (sans-IO) rather than read from the
// corpus tree.
const testGlyphDefinition = "# Glyph Definition\n\n" +
	"### Rest axis — the overhead release\n" +
	"> Form: " + restAxisFormRule + "\n"

// testProvenanceTypes is a minimal stand-in for GLYPH_PROVENANCE_TYPES.md.
const testProvenanceTypes = "# Glyph Provenance Types\n\nPROVENANCE-TAXONOMY-SENTINEL: the five types.\n"

// provenanceTaxonomySentinel appears only in testProvenanceTypes.
const provenanceTaxonomySentinel = "PROVENANCE-TAXONOMY-SENTINEL"

func TestItem6InlinesGlyphDefinition(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	st := &State{
		Content:   "entry",
		Model:     f,
		Reference: ReferenceMaterial{GlyphDefinition: testGlyphDefinition},
	}
	Item6EntryCoherence(context.Background(), st)
	prompt := f.prompts[0]
	for _, want := range []string{
		"=== BEGIN GLYPH DEFINITION ===",
		restAxisFormRule,
		"Trace each field against THIS TEXT",
		"do not report it as missing",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("item-6 prompt must carry %q when a reference is wired", want)
		}
	}
}

func TestItem6FallsBackWhenNoReference(t *testing.T) {
	f := &fakeClient{text: "PASS"}
	st := &State{Content: "SENTINEL-ENTRY", Model: f}
	Item6EntryCoherence(context.Background(), st)
	prompt := f.prompts[0]
	if strings.Contains(prompt, "=== BEGIN GLYPH DEFINITION ===") {
		t.Fatal("with no reference wired, item 6 must not claim to inline a definition")
	}
	// It still frames the item and carries the entry — the fallback is a working
	// prompt, not a broken one.
	if !strings.Contains(prompt, "retrosynthetic") || !strings.Contains(prompt, "SENTINEL-ENTRY") {
		t.Fatal("fallback prompt must keep the item framing and the entry content")
	}
}

func TestItem6DeliversProvenanceOnlyWhenCited(t *testing.T) {
	ref := ReferenceMaterial{
		GlyphDefinition: testGlyphDefinition,
		ProvenanceTypes: testProvenanceTypes,
	}
	t.Run("cited", func(t *testing.T) {
		f := &fakeClient{text: "PASS"}
		// The entry names a provenance type (Anachronicity).
		st := &State{Content: "Y-fire cites provenance type: Anachronicity.", Model: f, Reference: ref}
		Item6EntryCoherence(context.Background(), st)
		if !strings.Contains(f.prompts[0], provenanceTaxonomySentinel) {
			t.Fatal("an entry citing a provenance type must receive the taxonomy inline")
		}
	})
	t.Run("not-cited", func(t *testing.T) {
		f := &fakeClient{text: "PASS"}
		st := &State{Content: "a plain entry with no provenance reference", Model: f, Reference: ref}
		Item6EntryCoherence(context.Background(), st)
		if strings.Contains(f.prompts[0], provenanceTaxonomySentinel) {
			t.Fatal("an entry that cites no provenance type must not be handed the taxonomy (do not over-deliver)")
		}
	})
	t.Run("cited-but-no-provenance-doc", func(t *testing.T) {
		f := &fakeClient{text: "PASS"}
		st := &State{
			Content:   "Y-fire cites provenance type: Existence.",
			Model:     f,
			Reference: ReferenceMaterial{GlyphDefinition: testGlyphDefinition},
		}
		Item6EntryCoherence(context.Background(), st)
		if strings.Contains(f.prompts[0], "=== BEGIN PROVENANCE TYPES ===") {
			t.Fatal("no provenance block should appear when no provenance doc is wired")
		}
	})
}

// definitionAwareClient reproduces the bug-1334 failure shape: it FAILs item 6
// claiming the definition has no rule for the Y-neutral / absent-preconditions
// territory — UNLESS the Rest-axis Form rule is actually present in the prompt,
// in which case it traces the field and passes. Inlining the definition is the
// only thing that flips it from FAIL to PASS, so this test proves the fix closes
// the hallucination path rather than just adding text to the prompt.
type definitionAwareClient struct{ prompts []string }

func (c *definitionAwareClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.prompts = append(c.prompts, prompt)
	if strings.Contains(prompt, restAxisFormRule) {
		return model.Response{Text: "PASS"}, nil
	}
	return model.Response{Text: "FAIL the definition has no rule for the Y-neutral / absent-preconditions territory"}, nil
}

func (c *definitionAwareClient) Name() string    { return "definition-aware-fake" }
func (c *definitionAwareClient) Version() string { return "0.0.0" }
func (c *definitionAwareClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func TestItem6InlineReferenceClosesHallucination(t *testing.T) {
	const entry = "**Rest axis:** In Y-neutral territory the question does not arise.\n"
	t.Run("without-reference-hallucinates-fail", func(t *testing.T) {
		out := Item6EntryCoherence(context.Background(), &State{Content: entry, Model: &definitionAwareClient{}})
		if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
			t.Fatalf("without the definition inline, the judge hallucinates a FAIL; got %+v", out)
		}
	})
	t.Run("with-reference-passes", func(t *testing.T) {
		out := Item6EntryCoherence(context.Background(), &State{
			Content:   entry,
			Model:     &definitionAwareClient{},
			Reference: ReferenceMaterial{GlyphDefinition: testGlyphDefinition},
		})
		if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
			t.Fatalf("with the definition inline, the judge traces the Rest field and passes; got %+v", out)
		}
	})
}

func TestCitesProvenanceType(t *testing.T) {
	cases := []struct {
		content string
		want    bool
	}{
		{"provenance type: Anachronicity", true},
		{"the Existence gap here", true},
		{"a Recency verification", true},
		{"Authorization was not confirmed", true},
		{"Propagation to dependents", true},
		{"no provenance reference at all", false},
		{"", false},
	}
	for _, c := range cases {
		if got := citesProvenanceType(c.content); got != c.want {
			t.Errorf("citesProvenanceType(%q) = %v, want %v", c.content, got, c.want)
		}
	}
}

func TestVerdictGenParamsPinDeterministicSampling(t *testing.T) {
	p := verdictGenParams()
	if p.Temperature == nil || *p.Temperature != 0.0 {
		t.Fatalf("temperature should pin 0.0, got %v", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 8000 {
		t.Fatalf("max tokens should give a thinking model room (8000), got %v", p.MaxTokens)
	}
}
