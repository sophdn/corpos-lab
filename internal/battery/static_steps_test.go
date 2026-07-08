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

func TestItem15PassesWithField(t *testing.T) {
	out := Item15FalloutProfile(context.Background(), staticState("Entry text.\n\n**Fallout profile:** See FALLOUT_casg-direct.md"))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem15FailsWithoutField(t *testing.T) {
	out := Item15FalloutProfile(context.Background(), staticState("Entry text with no fallout field."))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected fail, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "Fallout profile") {
		t.Fatalf("got %q", out.Verdict.Reason)
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
