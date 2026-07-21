package battery

import (
	"context"
	"strings"
	"testing"
)

// Characterization tests ported from steps/battery/static_steps.rs (11 tests).

func staticState(content string) *State {
	return &State{ItemID: "test", Content: content, Model: &fakeClient{text: "mock"}}
}

func TestItem2PassesCleanEntry(t *testing.T) {
	out := Item2IntentLanguageScan(context.Background(), staticState("A firing condition based on observable artifacts."))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem2FailsOnIntentLanguage(t *testing.T) {
	out := Item2IntentLanguageScan(context.Background(), staticState("Fires when the agent decides to skip the step."))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 2 {
		t.Fatalf("expected item 2, got %v", out.Verdict.Item)
	}
	if !strings.Contains(out.Verdict.Reason, "intent-modeling language") {
		t.Fatalf("reason should mention intent-modeling language, got %q", out.Verdict.Reason)
	}
	if !strings.Contains(out.Verdict.Reason, `"when the agent decides"`) {
		t.Fatalf("reason should quote the matched phrase, got %q", out.Verdict.Reason)
	}
}

func TestItem2CaseInsensitive(t *testing.T) {
	out := Item2IntentLanguageScan(context.Background(), staticState("Fires WHEN THE AGENT BELIEVES the state is valid."))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
}

func TestItem4AlwaysEmitsNotApplicable(t *testing.T) {
	out := Item4NA(context.Background(), staticState("any content"))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindNotApplicable {
		t.Fatalf("expected not_applicable, got %+v", out)
	}
	if out.Verdict.Note != "item 4 does not apply to glyph entries" {
		t.Fatalf("note parity, got %q", out.Verdict.Note)
	}
}

func TestDeferredItem3EmitsDeferredWithPendingReason(t *testing.T) {
	out := DeferredStep(3)(context.Background(), staticState("ignored"))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindDeferred {
		t.Fatalf("expected deferred, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Pending, "item 3") || !strings.Contains(out.Verdict.Pending, "corpus") {
		t.Fatalf("pending should mention item 3 and corpus, got %q", out.Verdict.Pending)
	}
}

func TestEveryDeferredStubEmitsDeferred(t *testing.T) {
	for _, item := range []int{3, 5, 6, 7, 8, 11, 12, 13, 14} {
		out := DeferredStep(item)(context.Background(), staticState("ignored"))
		if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindDeferred {
			t.Fatalf("stub %d emitted non-Deferred: %+v", item, out)
		}
		wantTag := "item " + itoa(item)
		if !strings.Contains(out.Verdict.Pending, wantTag) {
			t.Fatalf("pending for item %d should mention %q, got: %q", item, wantTag, out.Verdict.Pending)
		}
	}
}

func TestDeferredPendingUnknownItemFallsBack(t *testing.T) {
	if got := DeferredPending(99); got != "item deferred pending integration" {
		t.Fatalf("got %q", got)
	}
}

func TestItem10PassesWithAllAxes(t *testing.T) {
	content := "**Y — Decision terrain**\n" +
		"Some description.\n" +
		"**Marker axis:** The failure direction.\n" +
		"**Aim axis:** The correct path.\n" +
		"**Rest axis:** Irrelevant territory.\n"
	out := Item10AxisPresence(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem10FailsMissingYMarker(t *testing.T) {
	out := Item10AxisPresence(context.Background(), staticState("Some text without the expected section heading."))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "no Y marker") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem10FailsMissingAxes(t *testing.T) {
	content := "**Y — Decision terrain**\nSome description.\n**Marker axis:** Only marker."
	out := Item10AxisPresence(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "Aim") || !strings.Contains(out.Verdict.Reason, "Rest") {
		t.Fatalf("reason should name Aim and Rest, got %q", out.Verdict.Reason)
	}
	if strings.Contains(out.Verdict.Reason, "Marker") {
		t.Fatalf("Marker is present and should not be listed, got %q", out.Verdict.Reason)
	}
}

// Item 15 fallout-profile fixtures. Both heading forms the restored profiles
// use; the missing-suppression fixture carries the full five-item definition
// preamble but omits the Suppression *section heading*, so it proves the check
// keys on headings rather than the vacuous substring the preamble would
// satisfy. The entry references the profile by the relative path the real
// candidate files use.
const (
	entryWithFallout = "Entry text.\n\n" +
		"**Fallout profile:** ../fallout-profiles/TASK_fallout-x.md\n"
	falloutRef = "../fallout-profiles/TASK_fallout-x.md"

	completeBoldProfile = "## Analysis\n" +
		"**Attentional shift.** primes scanning.\n" +
		"**Over-application.** could misfire.\n" +
		"**Meta-awareness effects.** frames done as co-presence.\n" +
		"**Scope creep.** the boundary is fuzzy.\n" +
		"**Suppression effects.** operation-close hesitation.\n"

	completeATXProfile = "## Analysis\n" +
		"### Attentional shift\nprimes scanning.\n" +
		"### Over-application\ncould misfire.\n" +
		"### Meta-awareness effects\nframes done.\n" +
		"### Scope creep\nthe boundary is fuzzy.\n" +
		"### Suppression effects\noperation-close hesitation.\n"

	missingSuppressionProfile = "## Analysis\n" +
		"Dimensions checked:\n" +
		"- **Attentional shift:** does loading prime attention?\n" +
		"- **Over-application:** could the action misfire?\n" +
		"- **Meta-awareness effects:** does it frame self-behavior?\n" +
		"- **Scope creep:** false-positive pattern-matching?\n" +
		"- **Suppression effects:** does it prime unwanted hesitation?\n\n" +
		"**Attentional shift.** primes scanning.\n" +
		"**Over-application.** could misfire.\n" +
		"**Meta-awareness effects.** frames done.\n" +
		"**Scope creep.** the boundary is fuzzy.\n"
	// no **Suppression effects.** section heading — only the preamble list item
)

func TestItem15FailsWhenFieldAbsent(t *testing.T) {
	st := &State{Content: "Entry text with no fallout field.", Profiles: &fakeProfiles{}}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "Fallout profile") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem15FailsWhenFieldPresentButEmpty(t *testing.T) {
	// A marker with no path is treated as no reference at all.
	st := &State{Content: "Entry text.\n\n**Fallout profile:**\n", Profiles: &fakeProfiles{}}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "Fallout profile") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem15FailsWhenReferentMissing(t *testing.T) {
	st := &State{Content: entryWithFallout, Profiles: &fakeProfiles{docs: map[string]string{}}}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "missing or unreadable") {
		t.Fatalf("reason should report an unreadable referent, got %q", out.Verdict.Reason)
	}
}

func TestItem15FailsWhenNoReaderConfigured(t *testing.T) {
	// A referenced field with no reader must fail closed, never pass: the
	// battery cannot verify a profile it cannot open.
	st := &State{Content: entryWithFallout, Profiles: nil}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "no profile reader") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem15FailsWhenDimensionUnaddressed(t *testing.T) {
	// The preamble lists all five dimensions; only the Suppression *section*
	// is missing. A substring scan would pass this vacuously — the heading
	// check must fail it and name the missing dimension.
	fp := &fakeProfiles{docs: map[string]string{falloutRef: missingSuppressionProfile}}
	st := &State{Content: entryWithFallout, Profiles: fp}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "suppression effects") {
		t.Fatalf("reason should name the unaddressed dimension, got %q", out.Verdict.Reason)
	}
	if strings.Contains(out.Verdict.Reason, "attentional shift") {
		t.Fatalf("addressed dimensions should not be listed, got %q", out.Verdict.Reason)
	}
}

func TestItem15PassesWhenReferentCompleteBoldHeadings(t *testing.T) {
	fp := &fakeProfiles{docs: map[string]string{falloutRef: completeBoldProfile}}
	st := &State{Content: entryWithFallout, Profiles: fp}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem15PassesWhenReferentCompleteATXHeadings(t *testing.T) {
	fp := &fakeProfiles{docs: map[string]string{falloutRef: completeATXProfile}}
	st := &State{Content: entryWithFallout, Profiles: fp}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem15ResolvesReferenceRelativeToEntryFile(t *testing.T) {
	// The `../` reference must resolve against the entry file's directory, not
	// be opened verbatim: a candidate in candidates/ points up-and-over into
	// fallout-profiles/.
	entry := "corpus/glyph-model/candidates/CANDIDATE_x.md"
	wantPath := "corpus/glyph-model/fallout-profiles/TASK_fallout-x.md"
	fp := &fakeProfiles{docs: map[string]string{wantPath: completeBoldProfile}}
	st := &State{Content: entryWithFallout, EntryPath: entry, Profiles: fp}
	out := Item15FalloutProfile(context.Background(), st)
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
	if len(fp.seen) != 1 || fp.seen[0] != wantPath {
		t.Fatalf("expected reader to receive %q, got %v", wantPath, fp.seen)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
