package batteryrun

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/model"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/substrate"
)

type fakeClient struct {
	text string
	err  error
}

func (f fakeClient) Generate(_ context.Context, _ string, _ model.GenParams) (model.Response, error) {
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{Text: f.text}, nil
}
func (f fakeClient) Name() string    { return "fake-model" }
func (f fakeClient) Version() string { return "v0" }
func (f fakeClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

// completeCandidate is a candidate working-doc, not a pre-cleaned entry. Its
// AC-3 specimen carries item-2 intent language and a project path that WOULD
// trip the intent and universality items if assessed; extraction must drop it.
// The AC-4 block it extracts to satisfies every mechanized static item: a Y
// marker with all three axes (item 10), no intent-modeling language (item 2),
// and a fallout field (item 15) pointing at profile.md alongside it. Its
// identity is "test" (**Glyph:** `test`).
const completeCandidate = "# Glyph Candidate: test\n" +
	"\n" +
	"**Fallout profile:** profile.md\n" +
	"\n" +
	"## AC-3 — Decomp source material\n" +
	"\n" +
	"Specimen: when the agent decides, at /home/foo/project/x.md.\n" +
	"\n" +
	"## AC-4 — Assembled glyph\n" +
	"\n" +
	"[Executor: fill the template below.]\n" +
	"\n" +
	"**Glyph:** `test`\n" +
	"**Y — Decision terrain**\nSome terrain description.\n" +
	"**Marker axis:** the failure direction.\n" +
	"**Aim axis:** the correct path.\n" +
	"**Rest axis:** irrelevant territory.\n"

const completeProfile = "## Analysis\n" +
	"**Attentional shift.** a.\n" +
	"**Over-application.** b.\n" +
	"**Meta-awareness effects.** c.\n" +
	"**Scope creep.** d.\n" +
	"**Suppression effects.** e.\n"

func writeCandidate(t *testing.T, entry, profile string) string {
	t.Helper()
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "candidate.md")
	if err := os.WriteFile(entryPath, []byte(entry), 0o600); err != nil {
		t.Fatal(err)
	}
	if profile != "" {
		if err := os.WriteFile(filepath.Join(dir, "profile.md"), []byte(profile), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return entryPath
}

func okDeps() Deps {
	return Deps{
		Substrate: func(_ context.Context) substrate.Info {
			return substrate.Info{Kind: substrate.KindGPU, DetectedBy: "test"}
		},
		Provenance: func(_ context.Context) (provenance.Stamp, error) {
			return provenance.Stamp{CommitSHA: "deadbeef"}, nil
		},
		Props: func(_ context.Context) (model.ServerProps, error) {
			return model.ServerProps{ModelAlias: "served-model"}, nil
		},
	}
}

// capturingClient records every prompt it is sent and always passes, so a test
// can assert what the model-assessed items were handed.
type capturingClient struct{ prompts []string }

func (c *capturingClient) Generate(_ context.Context, prompt string, _ model.GenParams) (model.Response, error) {
	c.prompts = append(c.prompts, prompt)
	return model.Response{Text: "PASS"}, nil
}
func (c *capturingClient) Name() string    { return "capturing-fake" }
func (c *capturingClient) Version() string { return "v0" }
func (c *capturingClient) Props(_ context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

// erroringReference is a ReferenceReader whose read always fails, so a test can
// assert the run records the gap and proceeds.
type erroringReference struct{}

func (erroringReference) Reference() (battery.ReferenceMaterial, error) {
	return battery.ReferenceMaterial{}, errors.New("reference unreadable")
}

func TestFileReferenceReaderReadsAndErrors(t *testing.T) {
	dir := t.TempDir()
	defPath := filepath.Join(dir, "GLYPH_DEFINITION.md")
	provPath := filepath.Join(dir, "GLYPH_PROVENANCE_TYPES.md")
	if err := os.WriteFile(defPath, []byte("DEF-SENTINEL"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(provPath, []byte("PROV-SENTINEL"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Both paths set: both halves read.
	m, err := FileReferenceReader{DefinitionPath: defPath, ProvenanceTypesPath: provPath}.Reference()
	if err != nil {
		t.Fatalf("Reference: %v", err)
	}
	if m.GlyphDefinition != "DEF-SENTINEL" || m.ProvenanceTypes != "PROV-SENTINEL" {
		t.Fatalf("read material = %+v", m)
	}

	// Empty paths: an empty, error-free material — that half is simply not delivered.
	if m, err := (FileReferenceReader{}).Reference(); err != nil || m.GlyphDefinition != "" || m.ProvenanceTypes != "" {
		t.Fatalf("empty reader should yield empty material, got %+v / %v", m, err)
	}

	// A missing definition file is a real read failure.
	if _, err := (FileReferenceReader{DefinitionPath: filepath.Join(dir, "missing.md")}).Reference(); err == nil {
		t.Fatal("expected an error for a missing definition doc")
	}
	// A missing provenance file is likewise a real read failure.
	if _, err := (FileReferenceReader{ProvenanceTypesPath: filepath.Join(dir, "missing.md")}).Reference(); err == nil {
		t.Fatal("expected an error for a missing provenance-types doc")
	}
}

func TestRunDeliversReferenceToModelAssessedItems(t *testing.T) {
	dir := t.TempDir()
	defPath := filepath.Join(dir, "GLYPH_DEFINITION.md")
	if err := os.WriteFile(defPath, []byte("DEFINITION-DELIVERED-SENTINEL"), 0o600); err != nil {
		t.Fatal(err)
	}
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	client := &capturingClient{}
	deps := okDeps()
	deps.Reference = FileReferenceReader{DefinitionPath: defPath}
	res, err := Run(context.Background(), entryPath, "test", client, deps, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Provenance.ReferenceError != "" {
		t.Fatalf("no gap expected, got %q", res.Provenance.ReferenceError)
	}
	found := false
	for _, p := range client.prompts {
		if strings.Contains(p, "DEFINITION-DELIVERED-SENTINEL") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("the item-6 judge must be handed the inlined glyph definition")
	}
}

func TestRunRecordsReferenceGapWithoutAborting(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Reference = erroringReference{}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("a failed reference read must not abort the run: %v", err)
	}
	if res.Provenance.ReferenceError != "reference unreadable" {
		t.Errorf("reference gap should be recorded, got %q", res.Provenance.ReferenceError)
	}
	// The run still completed its sequence.
	if len(res.Sequence.StepResults) == 0 {
		t.Fatal("the run should still have executed the sequence")
	}
}

// erroringBaseline is a BaselineReader whose read always fails, so a test can
// assert the run records the gap and item 12 defers rather than failing.
type erroringBaseline struct{}

func (erroringBaseline) Baseline() (battery.BaselineOutcome, error) {
	return battery.BaselineOutcome{}, errors.New("baseline unreadable")
}

func TestFileBaselineReaderReadsEmptyAndErrors(t *testing.T) {
	dir := t.TempDir()

	// A well-formed baseline JSON round-trips into the outcome.
	good := battery.BaselineOutcome{Models: []battery.BaselineModelResult{
		{ModelID: "Mistral-7B", Version: "v0.3", Fired: true, Runs: 8},
	}}
	encoded, err := json.Marshal(good)
	if err != nil {
		t.Fatal(err)
	}
	goodPath := filepath.Join(dir, "baseline.json")
	if err := os.WriteFile(goodPath, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := FileBaselineReader{Path: goodPath}.Baseline()
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if len(got.Models) != 1 || got.Models[0].ModelID != "Mistral-7B" || !got.Models[0].Fired {
		t.Fatalf("round-trip = %+v", got)
	}

	// An unset path is an empty, error-free outcome — no baseline supplied.
	if out, err := (FileBaselineReader{}).Baseline(); err != nil || len(out.Models) != 0 {
		t.Fatalf("empty path should yield empty/no-error, got %+v / %v", out, err)
	}

	// A missing file is a real read failure.
	if _, err := (FileBaselineReader{Path: filepath.Join(dir, "missing.json")}).Baseline(); err == nil {
		t.Fatal("expected an error for a missing baseline file")
	}

	// Malformed JSON is a parse failure, distinct from a missing file.
	badPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badPath, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (FileBaselineReader{Path: badPath}).Baseline(); err == nil {
		t.Fatal("expected a parse error for malformed baseline JSON")
	}
}

func writeBaseline(t *testing.T, out battery.BaselineOutcome) string {
	t.Helper()
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "baseline.json")
	if err := os.WriteFile(p, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunFailsItem12WhenBaselineFiresAcrossShelf(t *testing.T) {
	// The class's target behavior fired unguided on every shelf model → item 12
	// is a trained default → the run fails at item 12.
	baselinePath := writeBaseline(t, battery.BaselineOutcome{Models: []battery.BaselineModelResult{
		{ModelID: "Mistral-7B", Version: "v0.3", Fired: true, Runs: 8},
		{ModelID: "phi-4", Version: "phi4", Fired: true, Runs: 8},
	}})
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Baseline = FileBaselineReader{Path: baselinePath}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Sequence.Passed {
		t.Fatal("expected the run to fail on the empirical item-12 verdict")
	}
	failing := res.Sequence.StepResults[res.Sequence.ExitIndex]
	if failing.StepName != "item12-default-alignment" {
		t.Fatalf("expected the run to stop at item 12, stopped at %q", failing.StepName)
	}
	if failing.Outcome.Verdict == nil || failing.Outcome.Verdict.Kind != battery.KindFail {
		t.Fatalf("item 12 should be a FAIL, got %+v", failing.Outcome)
	}
	// The verdict's baseline is recorded in provenance so it is auditable.
	if len(res.Provenance.Baseline.Models) != 2 {
		t.Fatalf("run provenance must record the baseline population, got %+v", res.Provenance.Baseline)
	}
	if res.Provenance.BaselineError != "" {
		t.Fatalf("no baseline gap expected, got %q", res.Provenance.BaselineError)
	}
}

func TestRunItem12PassesWhenBaselineMissedOnAModel(t *testing.T) {
	// Missed unguided on one shelf model → the glyph does real work → item 12
	// passes and the run completes.
	baselinePath := writeBaseline(t, battery.BaselineOutcome{Models: []battery.BaselineModelResult{
		{ModelID: "Mistral-7B", Version: "v0.3", Fired: false, Runs: 8},
		{ModelID: "phi-4", Version: "phi4", Fired: true, Runs: 8},
	}})
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Baseline = FileBaselineReader{Path: baselinePath}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Sequence.Passed {
		t.Fatalf("the run should pass; failed at step %d: %q", res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}
	item12 := res.Sequence.StepResults[11]
	if item12.StepName != "item12-default-alignment" || item12.Outcome.Verdict.Kind != battery.KindPassWithCondition {
		t.Fatalf("item 12 = %+v", item12.Outcome)
	}
}

func TestRunRecordsBaselineGapWithoutAborting(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Baseline = erroringBaseline{}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("a failed baseline read must not abort the run: %v", err)
	}
	if res.Provenance.BaselineError != "baseline unreadable" {
		t.Errorf("baseline gap should be recorded, got %q", res.Provenance.BaselineError)
	}
	// Item 12 defers on the gap; the sequence still runs every step and passes.
	if len(res.Sequence.StepResults) != 15 {
		t.Fatalf("the run should still execute all 15 steps, got %d", len(res.Sequence.StepResults))
	}
	item12 := res.Sequence.StepResults[11]
	if item12.StepName != "item12-default-alignment" || item12.Outcome.Verdict.Kind != battery.KindDeferred {
		t.Fatalf("item 12 should defer on a baseline gap, got %+v", item12.Outcome)
	}
}

func TestFileProfileReaderReadsAndErrors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.md")
	if err := os.WriteFile(p, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader := FileProfileReader{}
	got, err := reader.ReadProfile(p)
	if err != nil || got != "hello" {
		t.Fatalf("ReadProfile = %q, %v", got, err)
	}
	if _, err := reader.ReadProfile(filepath.Join(dir, "missing.md")); err == nil {
		t.Fatal("expected error for a missing profile")
	}
}

func TestFileRegistryReaderReadsIdentitiesAbsentAndEmpty(t *testing.T) {
	dir := t.TempDir()
	// Absent file: a verified-empty registry, not an error.
	absent := FileRegistryReader{Path: filepath.Join(dir, "no-alphabet.md")}
	if ids, err := absent.RegistryIdentities(); err != nil || len(ids) != 0 {
		t.Fatalf("absent registry should be empty/no-error, got %v,%v", ids, err)
	}
	// Empty path: also empty.
	if ids, err := (FileRegistryReader{}).RegistryIdentities(); err != nil || len(ids) != 0 {
		t.Fatalf("empty path should be empty/no-error, got %v,%v", ids, err)
	}
	// Populated file: identities extracted with the shared rule.
	p := filepath.Join(dir, "ALPHABET.md")
	if err := os.WriteFile(p, []byte("# ALPHABET\n\n**Glyph:** `casg-delegate`\n\n**Glyph:** `casg-direct`\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ids, err := (FileRegistryReader{Path: p}).RegistryIdentities()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(ids) != 2 || ids[0] != "casg-delegate" || ids[1] != "casg-direct" {
		t.Fatalf("got %v", ids)
	}
}

func TestRunFailsItem3OnRegistryDuplicate(t *testing.T) {
	// completeCandidate's identity is "test" (# Glyph Candidate: test). The
	// registry carries that identity TWICE — one is the candidate's own entry
	// (self-excluded), the second is a distinct duplicate — so Item 3 fails and
	// the run stops there.
	dir := t.TempDir()
	regPath := filepath.Join(dir, "ALPHABET.md")
	if err := os.WriteFile(regPath, []byte("**Glyph:** `test`\n\n**Glyph:** `test`\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Registry = FileRegistryReader{Path: regPath}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Sequence.Passed {
		t.Fatal("expected the run to fail on the item-3 duplicate")
	}
	failing := res.Sequence.StepResults[res.Sequence.ExitIndex]
	if failing.StepName != "item3-duplicate-check" {
		t.Fatalf("expected the run to stop at item 3, stopped at %q", failing.StepName)
	}
}

func TestRunPassesItem3OnSelfCertification(t *testing.T) {
	// Re-certification: the registry holds the candidate's own identity once. Self
	// exclusion makes Item 3 pass — a glyph is never a duplicate of itself. This
	// is the false-fail the fix targets (verified with casg-direct against the
	// real ALPHABET registry).
	dir := t.TempDir()
	regPath := filepath.Join(dir, "ALPHABET.md")
	if err := os.WriteFile(regPath, []byte("**Glyph:** `test`\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := okDeps()
	deps.Registry = FileRegistryReader{Path: regPath}
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Sequence.Passed {
		t.Fatalf("re-certification should pass, failed at step %d: %q",
			res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}
}

func TestRunPassesMechanizedItemsAndCapturesProvenance(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	res, err := Run(context.Background(), entryPath, "test-item", fakeClient{text: "PASS"}, okDeps(), Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Sequence.Passed {
		t.Fatalf("expected pass, failed at step %d: %q", res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}
	if len(res.Sequence.StepResults) != 15 {
		t.Fatalf("expected 15 steps, got %d", len(res.Sequence.StepResults))
	}
	// Item 15 (last) must be a real PASS, not the old vacuous check: the reader
	// resolved profile.md next to the entry and found all five dimensions.
	item15 := res.Sequence.StepResults[14]
	if item15.StepName != "item15-fallout-profile" || item15.Outcome.Verdict.Kind != battery.KindPass {
		t.Fatalf("item 15 = %+v", item15.Outcome)
	}
	// Provenance captured.
	if res.Provenance.Provenance.CommitSHA != "deadbeef" {
		t.Errorf("commit = %q", res.Provenance.Provenance.CommitSHA)
	}
	if res.Provenance.Props.ModelAlias != "served-model" {
		t.Errorf("props alias = %q", res.Provenance.Props.ModelAlias)
	}
	if res.Provenance.Substrate.Kind != substrate.KindGPU {
		t.Errorf("substrate = %+v", res.Provenance.Substrate)
	}
	if res.Provenance.ModelID != "fake-model" || res.Provenance.ModelVersion != "v0" {
		t.Errorf("declared model = %q/%q", res.Provenance.ModelID, res.Provenance.ModelVersion)
	}
	if res.Provenance.Sampler.Temperature == nil || *res.Provenance.Sampler.Temperature != 0.0 {
		t.Errorf("sampler should record the pinned verdict temperature, got %+v", res.Provenance.Sampler)
	}
}

func TestRunFailsFastOnMechanizedFail(t *testing.T) {
	// Item 15 fails: no profile on disk. The sequence stops there; provenance
	// is still captured.
	entryPath := writeCandidate(t, completeCandidate, "")
	res, err := Run(context.Background(), entryPath, "test-item", fakeClient{text: "PASS"}, okDeps(), Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Sequence.Passed {
		t.Fatal("expected the run to fail at item 15 (missing referent)")
	}
	if res.Provenance.Provenance.CommitSHA != "deadbeef" {
		t.Errorf("provenance should still be captured on a failing run")
	}
}

// Options.AllItems carries through to the sequence: a run whose first
// model-assessed item fails still measures every later item, including item 15.
// This is what a recertification pass needs — six of the eight candidates fail
// at item 9, and item 15 is half of what the pass exists to measure.
func TestRunAllItemsMeasuresEveryItemDespiteFailure(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	res, err := Run(context.Background(), entryPath, "test-item",
		fakeClient{text: "FAIL not specific enough"}, okDeps(), Options{AllItems: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Sequence.Passed {
		t.Fatal("a failing item must still fail the run")
	}
	if !res.AllItems {
		t.Error("the result must record that it ran under all-items")
	}
	if got, want := len(res.Sequence.StepResults), battery.BuildBattery().StepCount(); got != want {
		t.Fatalf("expected all %d items measured, got %d", want, got)
	}
	last := res.Sequence.StepResults[len(res.Sequence.StepResults)-1]
	if last.StepName != "item15-fallout-profile" {
		t.Fatalf("expected item 15 to have run, last step was %q", last.StepName)
	}
}

func TestRunErrorsWhenCandidateUnreadable(t *testing.T) {
	_, err := Run(context.Background(), filepath.Join(t.TempDir(), "nope.md"), "x", fakeClient{text: "PASS"}, okDeps(), Options{})
	if err == nil {
		t.Fatal("expected an error reading a missing candidate")
	}
}

// The battery assesses the extracted clean entry, not the whole working-doc.
// completeCandidate's AC-3 specimen carries item-2 intent language ("when the
// agent decides") and a project path; both sit outside the AC-4 block. The run
// passes because extraction drops them. Without extraction, item 2 would fail
// on that text — this test is the regression guard for the one-step path.
func TestRunAssessesExtractedEntryNotWholeDoc(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	res, err := Run(context.Background(), entryPath, "test", fakeClient{text: "PASS"}, okDeps(), Options{AllItems: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Sequence.Passed {
		t.Fatalf("the run must pass on the extracted entry; failed at step %d: %q",
			res.Sequence.ExitIndex, res.Sequence.FailureReason)
	}
}

func TestRunErrorsWhenCandidateHasNoGlyphBlock(t *testing.T) {
	// A file with no AC-4 **Glyph:** line is not a candidate working-doc.
	entryPath := writeCandidate(t, "# Just a note\n\nNo glyph block here.\n", "")
	_, err := Run(context.Background(), entryPath, "x", fakeClient{text: "PASS"}, okDeps(), Options{})
	if err == nil {
		t.Fatal("expected an extraction error for a doc with no glyph block")
	}
}

func TestRunRecordsProvenanceGapsWithoutAborting(t *testing.T) {
	entryPath := writeCandidate(t, completeCandidate, completeProfile)
	deps := Deps{
		Substrate: func(_ context.Context) substrate.Info {
			return substrate.Info{Kind: substrate.KindUnknown, DetectedBy: "none"}
		},
		Provenance: func(_ context.Context) (provenance.Stamp, error) {
			return provenance.Stamp{}, errors.New("not a git repo")
		},
		Props: func(_ context.Context) (model.ServerProps, error) {
			return model.ServerProps{}, errors.New("props unreachable")
		},
	}
	res, err := Run(context.Background(), entryPath, "test-item", fakeClient{text: "PASS"}, deps, Options{})
	if err != nil {
		t.Fatalf("a failed readback must not abort the run: %v", err)
	}
	if res.Provenance.PropsError != "props unreachable" {
		t.Errorf("props error = %q", res.Provenance.PropsError)
	}
	if res.Provenance.ProvenanceError != "not a git repo" {
		t.Errorf("provenance error = %q", res.Provenance.ProvenanceError)
	}
}
