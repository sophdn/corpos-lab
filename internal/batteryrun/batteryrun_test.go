package batteryrun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

// completeCandidate satisfies every mechanized static item: a Y marker with all
// three axes (item 10), no intent-modeling language (item 2), and a fallout
// field (item 15) pointing at profile.md alongside it.
const completeCandidate = "# Glyph Candidate: test\n\n" +
	"**Y — Decision terrain**\nSome terrain description.\n" +
	"**Marker axis:** the failure direction.\n" +
	"**Aim axis:** the correct path.\n" +
	"**Rest axis:** irrelevant territory.\n" +
	"**Fallout profile:** profile.md\n"

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
