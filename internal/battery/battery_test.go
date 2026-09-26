package battery

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Characterization tests ported from sequences/battery.rs (3 tests) and
// tests/battery_integration.rs (2 tests + fixtures, copied verbatim to
// testdata/).

var allStepNames = []string{
	"item1-xyz-specificity",
	"item2-intent-language-scan",
	"item3-duplicate-check",
	"item4-na",
	"item5-y-not-fire",
	"item6-entry-coherence",
	"item7-sister-mirror",
	"item8-phenomenological",
	"item9-universality",
	"item10-axis-presence",
	"item11-safety-class",
	"item12-default-alignment",
	"item13-contamination-radius",
	"item14-globality-demand",
	"item15-fallout-profile",
}

func TestBatteryHas15ItemsMatchingSourceDoc(t *testing.T) {
	// ALPHABET_ENTRY_BATTERY.md declares 15 items across four strata.
	// Mismatch here means the sequence drifted from the source doc.
	if got := BuildBattery().StepCount(); got != 15 {
		t.Fatalf("expected 15 steps, got %d", got)
	}
}

func TestBatteryNameIsBattery(t *testing.T) {
	if got := BuildBattery().Name; got != "battery" {
		t.Fatalf("expected name battery, got %q", got)
	}
}

func TestEveryRegisteredStepHasAVersion(t *testing.T) {
	// Every registered step name must resolve to a non-default version;
	// "0.0.0" is reserved for unknown steps.
	for _, name := range allStepNames {
		if v := StepVersion(name); v == "0.0.0" {
			t.Errorf("step %s is missing a version entry", name)
		}
	}
	if v := StepVersion("not-a-step"); v != "0.0.0" {
		t.Fatalf("unknown step should be 0.0.0, got %q", v)
	}
	if v := StepVersion("item2-intent-language-scan"); v != "0.1.0" {
		t.Fatalf("implemented step should be 0.1.0, got %q", v)
	}
	// Items 3, 11, 13 mechanized under port-structural-prober-items now carry a
	// real 0.1.0 version, no longer "0.0.0-deferred". (Item 6 left this cohort
	// under the model-assessed hardening, and item 12 left it under the empirical
	// rebuild — see the 0.2.0 group below.)
	for _, name := range []string{
		"item3-duplicate-check",
		"item11-safety-class",
		"item13-contamination-radius",
	} {
		if v := StepVersion(name); v != "0.1.0" {
			t.Fatalf("mechanized step %s should be 0.1.0, got %q", name, v)
		}
	}
	// Only items 5, 7, 8, 14 remain deferred (a separate task decides them).
	for _, name := range []string{
		"item5-y-not-fire",
		"item7-sister-mirror",
		"item8-phenomenological",
		"item14-globality-demand",
	} {
		if v := StepVersion(name); v != "0.0.0-deferred" {
			t.Fatalf("deferred step %s should be 0.0.0-deferred, got %q", name, v)
		}
	}
	// The item-9/15 repairs (task 3586), the item-1 resolved-reading repair
	// (task 3588), the item-6 inline-reference repair (bug 1334 + suggestion
	// 175), and the item-12 empirical rebuild (bug 1332) bump those past the
	// 0.1.0 cohort so a post-hoc query can separate the repaired/rebuilt-logic
	// runs from the prior runs whose verdicts on those items cannot be trusted.
	for _, name := range []string{
		"item1-xyz-specificity",
		"item6-entry-coherence",
		"item9-universality",
		"item12-default-alignment",
		"item15-fallout-profile",
	} {
		if v := StepVersion(name); v != "0.2.0" {
			t.Fatalf("repaired step %s should be 0.2.0, got %q", name, v)
		}
		if StepVersion(name) == StepVersion("item2-intent-language-scan") {
			t.Fatalf("repaired step %s must be distinguishable from the 0.1.0 cohort", name)
		}
	}
}

func fixture(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return string(raw)
}

// knownPassProfiles wires the fake fallout-profile reader the full-battery
// fixtures need now that Item 15 opens and dimensionally checks the referent.
// known_pass.md references FALLOUT_test-known-pass.md; the reader returns a
// profile covering all five dimensions.
func knownPassProfiles() *fakeProfiles {
	return &fakeProfiles{docs: map[string]string{
		"FALLOUT_test-known-pass.md": completeBoldProfile,
	}}
}

func TestKnownPassItemPassesFullBattery(t *testing.T) {
	content := fixture(t, "known_pass.md")
	result := RunSequence(context.Background(), BuildBattery(), Input{
		ItemID:   "known-pass",
		Content:  content,
		Model:    &fakeClient{text: "PASS"},
		Profiles: knownPassProfiles(),
		Registry: &fakeRegistry{},
	})

	if !result.Passed {
		t.Fatalf("known-pass item should pass all battery steps, but failed at step %d with reason: %q",
			result.ExitIndex, result.FailureReason)
	}
	if len(result.StepResults) != 15 {
		t.Fatalf("expected 15 step results, got %d", len(result.StepResults))
	}
	for i, sr := range result.StepResults {
		if sr.StepName != allStepNames[i] {
			t.Fatalf("step %d name = %q, want %q (stratum order parity)", i, sr.StepName, allStepNames[i])
		}
	}
}

func TestKnownFailItem2FailsAtIntentLanguage(t *testing.T) {
	content := fixture(t, "known_fail_item2.md")
	result := RunSequence(context.Background(), BuildBattery(), Input{
		ItemID:  "known-fail",
		Content: content,
		Model:   &fakeClient{text: "PASS"},
	})

	if result.Passed {
		t.Fatal("known-fail item should fail the battery")
	}
	// Item 2 is step index 1 (fail-fast ordering parity).
	if result.ExitIndex != 1 {
		t.Fatalf("expected exit at step index 1 (item2), got %d", result.ExitIndex)
	}
	failing := result.StepResults[result.ExitIndex]
	if failing.Outcome.Kind != OutcomeVerdict || failing.Outcome.Verdict.Kind != KindFail {
		t.Fatalf("expected Fail verdict at item2, got %+v", failing.Outcome)
	}
	if failing.Outcome.Verdict.Item == nil || *failing.Outcome.Verdict.Item != 2 {
		t.Fatalf("item2 is the expected failing step, got %v", failing.Outcome.Verdict.Item)
	}
	if !strings.Contains(failing.Outcome.Verdict.Reason, "intent-modeling language") {
		t.Fatalf("reason should mention intent-modeling language, got: %q", failing.Outcome.Verdict.Reason)
	}
}

func TestFullBatteryComposeOnKnownPassIsDefer(t *testing.T) {
	// The 15-item run on a passing entry with no baseline wired carries 5 Deferred
	// verdicts: the 4 stubs (items 5, 7, 8, 14) plus item 12, which now defers on
	// an absent unguided baseline rather than being model-assessed. So the composed
	// run verdict is Defer (not Promote) until those land / a baseline is supplied.
	content := fixture(t, "known_pass.md")
	result := RunSequence(context.Background(), BuildBattery(), Input{
		ItemID:   "known-pass",
		Content:  content,
		Model:    &fakeClient{text: "PASS"},
		Profiles: knownPassProfiles(),
		Registry: &fakeRegistry{},
	})
	composed := ComposeRunVerdict(result.ItemVerdicts())
	if composed.Kind != RunDefer {
		t.Fatalf("expected Defer, got %+v", composed)
	}
	if len(composed.Pending) != 5 {
		t.Fatalf("expected 5 pending reasons (4 stubs + item 12's absent baseline), got %d", len(composed.Pending))
	}
}
