package baseline

import (
	"context"
	"errors"
	"testing"
	"time"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/model"
)

// fakeClient is a scripted model.Client: it returns queued responses in order
// and reports a fixed name and version.
type fakeClient struct {
	name, version string
	replies       []model.Response
	err           error
	calls         int
}

func (f *fakeClient) Generate(_ context.Context, _ string, _ model.GenParams) (model.Response, error) {
	if f.err != nil {
		return model.Response{}, f.err
	}
	r := f.replies[f.calls]
	f.calls++
	return r, nil
}
func (f *fakeClient) Name() string    { return f.name }
func (f *fakeClient) Version() string { return f.version }
func (f *fakeClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, nil
}

func fixedNow() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }

func TestRunCaptureRecordsReportedModelAndRuns(t *testing.T) {
	c := &fakeClient{
		name:    "role-anchor",
		version: "v0.3-q4km",
		replies: []model.Response{
			{Text: "reply one", Model: "Mistral-7B", SystemFingerprint: "b1"},
			{Text: "reply two", Model: "Mistral-7B", SystemFingerprint: "b1"},
		},
	}
	got, err := RunCapture(context.Background(), c, "post-write-verification", "SCENARIO", 2, assay.Sampling{}, fixedNow)
	if err != nil {
		t.Fatalf("RunCapture: %v", err)
	}
	if got.ModelID != "Mistral-7B" {
		t.Errorf("ModelID = %q, want the reported model Mistral-7B", got.ModelID)
	}
	if got.Requested != "role-anchor" || got.Version != "v0.3-q4km" {
		t.Errorf("Requested/Version = %q/%q, want role-anchor/v0.3-q4km", got.Requested, got.Version)
	}
	if got.RunDate != "2026-09-29T12:00:00Z" {
		t.Errorf("RunDate = %q, want the injected clock", got.RunDate)
	}
	if len(got.Runs) != 2 {
		t.Fatalf("Runs = %d, want 2", len(got.Runs))
	}
	if got.Runs[0].ReplyText != "reply one" || got.Runs[1].ReplyText != "reply two" {
		t.Errorf("reply text not captured in order: %+v", got.Runs)
	}
	if got.Runs[0].SliceID == got.Runs[1].SliceID {
		t.Errorf("slice ids must differ per run, both %q", got.Runs[0].SliceID)
	}
	if got.Runs[0].Observed.Model != "Mistral-7B" || got.Runs[0].Observed.BuildInfo != "b1" {
		t.Errorf("Observed not captured: %+v", got.Runs[0].Observed)
	}
}

func TestRunCaptureNilNowUsesWallClock(t *testing.T) {
	c := &fakeClient{name: "m", replies: []model.Response{{Text: "x", Model: "m"}}}
	got, err := RunCapture(context.Background(), c, "item", "S", 1, assay.Sampling{}, nil)
	if err != nil {
		t.Fatalf("RunCapture: %v", err)
	}
	if got.RunDate == "" {
		t.Error("RunDate empty with nil now; want a wall-clock stamp")
	}
}

func TestRunCaptureEmptyReportedModelFallsBackToRequested(t *testing.T) {
	c := &fakeClient{name: "role-primary", replies: []model.Response{{Text: "x"}}}
	got, err := RunCapture(context.Background(), c, "item", "S", 1, assay.Sampling{}, fixedNow)
	if err != nil {
		t.Fatalf("RunCapture: %v", err)
	}
	if got.ModelID != "" {
		t.Errorf("ModelID = %q, want empty (server reported none)", got.ModelID)
	}
	if got.Runs[0].SliceID != "item|role-primary|run1" {
		t.Errorf("SliceID = %q, want the requested-model fallback", got.Runs[0].SliceID)
	}
}

func TestRunCaptureRejectsEmptyScenario(t *testing.T) {
	c := &fakeClient{name: "m"}
	if _, err := RunCapture(context.Background(), c, "item", "", 1, assay.Sampling{}, fixedNow); err == nil {
		t.Fatal("want an error for an empty scenario")
	}
}

func TestRunCaptureRejectsNonPositiveRuns(t *testing.T) {
	c := &fakeClient{name: "m"}
	if _, err := RunCapture(context.Background(), c, "item", "S", 0, assay.Sampling{}, fixedNow); err == nil {
		t.Fatal("want an error for runs < 1")
	}
}

func TestRunCapturePropagatesGenerateError(t *testing.T) {
	c := &fakeClient{name: "m", err: errors.New("boom")}
	if _, err := RunCapture(context.Background(), c, "item", "S", 1, assay.Sampling{}, fixedNow); err == nil {
		t.Fatal("want the Generate error surfaced")
	}
}

func TestSliceIDUnknownWhenNoModelNamed(t *testing.T) {
	if got := SliceID("item", "", "", 3); got != "item|unknown-model|run3" {
		t.Errorf("SliceID = %q, want the unknown-model fallback", got)
	}
}

func TestSliceLinesMapIDAndText(t *testing.T) {
	c := Capture{Runs: []CaptureRun{
		{SliceID: "a", ReplyText: "one"},
		{SliceID: "b", ReplyText: "two"},
	}}
	lines := c.SliceLines()
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	if lines[0].ID != "a" || lines[0].Text != "one" || lines[0].Scenario != "" {
		t.Errorf("line 0 = %+v, want id a / text one / no scenario", lines[0])
	}
	if lines[1].ID != "b" || lines[1].Text != "two" {
		t.Errorf("line 1 = %+v", lines[1])
	}
}
