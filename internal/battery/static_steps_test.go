package battery

import (
	"context"
	"errors"
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

func TestEveryDeferredStubEmitsDeferred(t *testing.T) {
	// Only items 5, 7, 8, 14 remain deferred; 3, 6, 11, 12, 13 are mechanized.
	for _, item := range []int{5, 7, 8, 14} {
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

// --- Item 3: duplicate check (deterministic) ---

func regState(content string, reg RegistryReader) *State {
	return &State{ItemID: "test", Content: content, Registry: reg}
}

func TestItem3PassesAgainstEmptyRegistry(t *testing.T) {
	out := Item3DuplicateCheck(context.Background(),
		regState("**Glyph:** `unique-slug`\n", &fakeRegistry{}))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("clean candidate against empty registry should pass, got %+v", out)
	}
}

func TestItem3PassesWhenIdentityNotInPopulatedRegistry(t *testing.T) {
	reg := &fakeRegistry{identities: []string{"casg-delegate", "casg-direct"}}
	out := Item3DuplicateCheck(context.Background(), regState("**Glyph:** `new-glyph`\n", reg))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("non-colliding identity should pass, got %+v", out)
	}
}

func TestItem3PassesOnSelfCertification(t *testing.T) {
	// Re-certifying an already-promoted glyph: its own identity is the only match
	// in the registry, so self-exclusion makes item 3 pass. A glyph is never a
	// duplicate of itself.
	reg := &fakeRegistry{identities: []string{"casg-direct"}}
	out := Item3DuplicateCheck(context.Background(), regState("**Glyph:** `casg-direct`\n", reg))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("re-certification should pass via self-exclusion, got %+v", out)
	}
}

func TestItem3FailsOnDistinctDuplicateCaseInsensitive(t *testing.T) {
	// The identity appears twice: one occurrence is the candidate's own entry
	// (self-excluded), the second is a distinct entry sharing the identity, which
	// is a genuine duplicate. Case-insensitive matching is covered here too.
	reg := &fakeRegistry{identities: []string{"CASG-Delegate", "casg-delegate"}}
	out := Item3DuplicateCheck(context.Background(), regState("**Glyph:** `casg-delegate`\n", reg))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("expected duplicate fail, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 3 {
		t.Fatalf("expected item 3, got %v", out.Verdict.Item)
	}
	if !strings.Contains(out.Verdict.Reason, "duplicate identity") {
		t.Fatalf("reason should report a duplicate, got %q", out.Verdict.Reason)
	}
}

func TestItem3FailsWhenEntryHasNoIdentity(t *testing.T) {
	out := Item3DuplicateCheck(context.Background(),
		regState("Some entry text with no glyph field or candidate heading.\n", &fakeRegistry{}))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("an entry with no identity must fail the duplicate check, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "no glyph identity") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem3FailsClosedOnNilRegistry(t *testing.T) {
	// Parity with Item 15: a check the battery cannot run must not pass.
	out := Item3DuplicateCheck(context.Background(), regState("**Glyph:** `x`\n", nil))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("nil registry must fail closed, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "no registry reader") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem3FailsWhenRegistryUnreadable(t *testing.T) {
	reg := &fakeRegistry{err: errors.New("permission denied")}
	out := Item3DuplicateCheck(context.Background(), regState("**Glyph:** `x`\n", reg))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("an unreadable registry must fail closed, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "registry unreadable") {
		t.Fatalf("got %q", out.Verdict.Reason)
	}
}

func TestItem3UsesCandidateHeadingWhenNoGlyphField(t *testing.T) {
	// The identity comes from the candidate heading; it appears twice in the
	// registry (own entry + a distinct duplicate), so after self-exclusion the
	// second occurrence fails the item — proving the heading identity was read.
	reg := &fakeRegistry{identities: []string{"test", "test"}}
	out := Item3DuplicateCheck(context.Background(),
		regState("# Glyph Candidate: test\n\nbody\n", reg))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("heading identity should be read and deduped, got %+v", out)
	}
}

func TestGlyphIdentity(t *testing.T) {
	if id, ok := GlyphIdentity("**Glyph:** `alpha-beta`\n"); !ok || id != "alpha-beta" {
		t.Fatalf("glyph field: got %q,%v", id, ok)
	}
	if id, ok := GlyphIdentity("# Glyph Candidate: gamma\n"); !ok || id != "gamma" {
		t.Fatalf("heading fallback: got %q,%v", id, ok)
	}
	// The glyph field wins over a candidate heading when both are present.
	both := "# Glyph Candidate: heading-slug\n\n**Glyph:** `field-slug`\n"
	if id, ok := GlyphIdentity(both); !ok || id != "field-slug" {
		t.Fatalf("glyph field should win: got %q,%v", id, ok)
	}
	if _, ok := GlyphIdentity("no identity here\n"); ok {
		t.Fatal("expected no identity")
	}
}

func TestRegistryIdentitiesFrom(t *testing.T) {
	doc := "# ALPHABET\n\n**Glyph:** `slug`\n\n**Glyph:** `casg-delegate`\n\n**Glyph:** `casg-direct`\n"
	got := RegistryIdentitiesFrom(doc)
	if len(got) != 3 || got[1] != "casg-delegate" || got[2] != "casg-direct" {
		t.Fatalf("got %v", got)
	}
	if len(RegistryIdentitiesFrom("no entries here")) != 0 {
		t.Fatal("expected no identities from an entry-free doc")
	}
}

// --- Item 13: contamination radius (deterministic) ---

func TestItem13PassesWithNoDependencySurface(t *testing.T) {
	out := Item13ContaminationRadius(context.Background(),
		staticState("**Glyph:** `standalone`\n\nNo other glyph references here.\n"))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestItem13PassesAtThresholdBoundary(t *testing.T) {
	// Exactly five referenced glyphs is at the threshold, not over it.
	content := "**Glyph:** `subject`\n" +
		"Depends on `a-one`, `b-two`, `c-three`, `d-four`, `e-five`.\n"
	out := Item13ContaminationRadius(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("five deps is at threshold, should pass, got %+v", out)
	}
}

func TestItem13FailsAboveThreshold(t *testing.T) {
	content := "**Glyph:** `subject`\n" +
		"Depends on `a-one`, `b-two`, `c-three`, `d-four`, `e-five`, `f-six`.\n"
	out := Item13ContaminationRadius(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFail {
		t.Fatalf("six deps exceeds threshold, should fail, got %+v", out)
	}
	if out.Verdict.Item == nil || *out.Verdict.Item != 13 {
		t.Fatalf("expected item 13, got %v", out.Verdict.Item)
	}
	if !strings.Contains(out.Verdict.Reason, "dependency surface has 6") {
		t.Fatalf("reason should report the count, got %q", out.Verdict.Reason)
	}
}

func TestItem13ExcludesOwnIdentityFromCount(t *testing.T) {
	// The entry's own slug repeated many times is not a dependency on another
	// glyph and must not push the count over threshold.
	content := "**Glyph:** `self-ref`\n" +
		"`self-ref` `self-ref` `self-ref` `self-ref` `self-ref` `self-ref` `self-ref`\n"
	out := Item13ContaminationRadius(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("own-identity references must be excluded, got %+v", out)
	}
}

func TestItem13FlagsHighRiskCategory(t *testing.T) {
	content := "**Glyph:** `guarded`\n\nThis governs a trust-class decision point.\n"
	out := Item13ContaminationRadius(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindFlag {
		t.Fatalf("high-risk category should FLAG, got %+v", out)
	}
	if !strings.Contains(out.Verdict.Reason, "trust-class") {
		t.Fatalf("flag should name the category, got %q", out.Verdict.Reason)
	}
}

func TestItem13IgnoresBacktickedFilePaths(t *testing.T) {
	// A backtick-quoted file path is Item 9's concern, not a glyph dependency:
	// six of them must not be counted as six referenced glyphs.
	content := "**Glyph:** `subject`\n" +
		"See `a/b.md`, `c/d.md`, `e/f.md`, `g/h.md`, `i/j.md`, `k/l.md`.\n"
	out := Item13ContaminationRadius(context.Background(), staticState(content))
	if out.Kind != OutcomeVerdict || out.Verdict.Kind != KindPass {
		t.Fatalf("file paths must not count as dependency references, got %+v", out)
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
