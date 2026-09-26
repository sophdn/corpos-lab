package persist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/control"
	"corpos-lab/internal/extract"
	"corpos-lab/internal/manifest"
	"corpos-lab/internal/model"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/runner"
	"corpos-lab/internal/substrate"
)

func sampleRun() control.StudyRun {
	return control.StudyRun{
		Name:        "casg-direct-v3-smoke",
		Assay:       "grounded-glyph-probe",
		ItemID:      "casg-direct",
		Image:       "localhost/lab-grounded-glyph-probe:dev",
		ImageDigest: "sha256:abc123",
		Status:      control.StatusCompleted,
		Manifest: manifest.RunManifest{
			StudyDigest:  "studyhash",
			Materials:    map[string]string{"scenario.md": "s-hash", "glyph.md": "g-hash"},
			ModelID:      "qwen",
			ModelVersion: "q4",
		},
		Results: &runner.Results{
			Assay: "grounded-glyph-probe", ItemID: "casg-direct", ModelID: "qwen",
			Rows: []assay.ScoreRow{
				// As a container emits them: captured, not yet judged.
				{Item: "casg-direct", Condition: assay.Baseline, Run: 1, Score: assay.Unscored, Rationale: "rat-b"},
				// As a judge leaves them once the rubric has been applied.
				{Item: "casg-direct", Condition: assay.GroundedGlyph, Run: 1, Score: assay.ScoreC, Rationale: "rat-g"},
			},
		},
		Extraction: &extract.Manifest{FinishedAt: "2026-07-09T00:33:10Z"},
	}
}

func TestParamsFromFlattensRunAndScores(t *testing.T) {
	p := paramsFrom(sampleRun(), "/out/responses")
	if p.Name != "casg-direct-v3-smoke" || p.Assay != "grounded-glyph-probe" || p.Status != "completed" {
		t.Fatalf("parent fields: %+v", p)
	}
	if p.StudyDigest != "studyhash" || p.ModelID != "qwen" || p.ModelVersion != "q4" {
		t.Fatalf("manifest-derived fields: %+v", p)
	}
	if p.MaterialsHashes["scenario.md"] != "s-hash" {
		t.Fatalf("materials hashes: %+v", p.MaterialsHashes)
	}
	if p.ResponsesDir != "/out/responses" {
		t.Fatalf("responses pointer: %q", p.ResponsesDir)
	}
	if p.RunAt != "2026-07-09T00:33:10Z" {
		t.Fatalf("run_at from extraction: %q", p.RunAt)
	}
	if len(p.Rows) != 2 {
		t.Fatalf("expected 2 flattened rows, got %d", len(p.Rows))
	}
	// An unjudged capture must reach the canonical DB saying so. It previously
	// shipped as verdict_kind "fail" with the response text as the reason,
	// which is the defect this asserts against: a run nobody has scored is not
	// a run that failed.
	if p.Rows[0].VerdictKind != "unscored" || p.Rows[0].Condition != "baseline" {
		t.Fatalf("row0 flatten: %+v", p.Rows[0])
	}
	if p.Rows[0].VerdictReason != "" {
		t.Fatalf("unscored row must carry no failure reason, got %q", p.Rows[0].VerdictReason)
	}
	if p.Rows[1].VerdictKind != "C" || p.Rows[1].Condition != "grounded_glyph" {
		t.Fatalf("row1 flatten: %+v", p.Rows[1])
	}
}

func TestParamsFromShipsZeroScoreAsExplicitUnscored(t *testing.T) {
	// A row whose Score was never set must not reach the toolkit as a blank
	// cell — blank is indistinguishable from a dropped field.
	run := sampleRun()
	run.Results.Rows = []assay.ScoreRow{
		{Item: "casg-direct", Condition: assay.Baseline, Run: 1, Rationale: "r"},
	}
	p := paramsFrom(run, "/out/responses")
	if p.Rows[0].VerdictKind != "unscored" {
		t.Fatalf("zero-value score should ship as the explicit sentinel, got %q", p.Rows[0].VerdictKind)
	}
}

func TestParamsFromHandlesNilResultsAndExtraction(t *testing.T) {
	// A failed run may carry no results and no extraction manifest.
	run := control.StudyRun{Name: "n", Assay: "grounded-glyph-probe", ItemID: "i", Status: control.StatusFailed, Error: "boom"}
	p := paramsFrom(run, "/out/responses")
	if p.Status != "failed" || p.Error != "boom" {
		t.Fatalf("failed run: %+v", p)
	}
	if p.Rows == nil || len(p.Rows) != 0 {
		t.Fatalf("rows should be empty non-nil, got %v", p.Rows)
	}
	if p.MaterialsHashes == nil {
		t.Fatal("materials hashes should be non-nil")
	}
	if p.RunAt != "" {
		t.Fatalf("run_at should be empty with no extraction, got %q", p.RunAt)
	}
}

func TestRecordPostsMCPEnvelope(t *testing.T) {
	var gotPath string
	var gotEnv mcpEnvelope
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotEnv); err != nil {
			t.Errorf("decode: %v", err)
		}
		_, _ = w.Write([]byte(`{"ok":true,"run_id":"sr-1"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "glyph-research", WithHTTPClient(srv.Client()))
	if err := c.Record(context.Background(), sampleRun(), "/out/responses"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if gotPath != "/mcp/measure" {
		t.Fatalf("path = %q, want /mcp/measure", gotPath)
	}
	if gotEnv.Action != "study_run_record" || gotEnv.Project != "glyph-research" {
		t.Fatalf("envelope: %+v", gotEnv)
	}
	if gotEnv.Rationale == "" {
		t.Fatal("rationale required by toolkit dispatch policy")
	}
	if gotEnv.Params.Name != "casg-direct-v3-smoke" || len(gotEnv.Params.Rows) != 2 {
		t.Fatalf("params: %+v", gotEnv.Params)
	}
}

func TestRecordNon2xxIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad params"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "glyph-research", WithHTTPClient(srv.Client()))
	err := c.Record(context.Background(), sampleRun(), "/out/responses")
	if err == nil {
		t.Fatal("expected error on 400")
	}
}

func TestRecordTruncatesLongErrorBody(t *testing.T) {
	long := make([]byte, 500)
	for i := range long {
		long[i] = 'x'
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(long)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "glyph-research", WithHTTPClient(srv.Client()))
	err := c.Record(context.Background(), sampleRun(), "/out/responses")
	if err == nil {
		t.Fatal("expected error")
	}
	// The 500-char body must be truncated with an ellipsis in the message.
	msg := err.Error()
	if len(msg) > 400 {
		t.Fatalf("error body should be truncated, got %d chars", len(msg))
	}
	if got := truncate("short", 300); got != "short" {
		t.Fatalf("truncate short = %q", got)
	}
}

func TestRecordTransportErrorIsWrapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // connection refused
	c := NewClient(srv.URL, "glyph-research")
	if err := c.Record(context.Background(), sampleRun(), "/out/responses"); err == nil {
		t.Fatal("expected transport error")
	}
}

// The toolkit copy has to carry the observations, not just the declarations.
// ModelID is an echo of the study definition; without these fields the queryable
// home could only ever answer "what was asked for", which is the question that
// was never in doubt.
func TestParamsFromShipsObservationsNotJustDeclarations(t *testing.T) {
	run := sampleRun()
	run.Substrate = substrate.Info{
		Kind: substrate.KindGPU, Device: "NVIDIA GeForce RTX 3090",
		VRAMMiB: 21184, DetectedBy: "nvidia-smi",
	}
	run.Provenance = provenance.Stamp{CommitSHA: "abc123", Dirty: true}
	run.Results.Sampler = assay.Sampling{Temperature: 0.8, MinP: 0.05, TopK: 0, RepeatPenalty: 1.0}
	run.Results.Server = model.ServerProps{
		ModelAlias: "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf",
		BuildInfo:  "b9445-af6528e6d",
		NCtx:       8192,
	}
	run.Results.ModelMismatch = "study declared X; server reports Y"

	p := paramsFrom(run, "/runs/responses")

	if p.Substrate.Kind != substrate.KindGPU || p.Substrate.Device != "NVIDIA GeForce RTX 3090" {
		t.Errorf("substrate not forwarded: %+v", p.Substrate)
	}
	if p.CommitSHA != "abc123" || !p.CommitDirty {
		t.Errorf("provenance not forwarded: sha=%q dirty=%v", p.CommitSHA, p.CommitDirty)
	}
	if p.Sampler.MinP != 0.05 || p.Sampler.RepeatPenalty != 1.0 {
		t.Errorf("sampler not forwarded: %+v", p.Sampler)
	}
	if p.Server.BuildInfo != "b9445-af6528e6d" || p.Server.NCtx != 8192 {
		t.Errorf("server readback not forwarded: %+v", p.Server)
	}
	if p.ModelMismatch != "study declared X; server reports Y" {
		t.Errorf("model mismatch not forwarded: %q", p.ModelMismatch)
	}

	// And it has to survive JSON — the toolkit reads the wire, not the struct.
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"substrate"`, `"commit_sha":"abc123"`, `"commit_dirty":true`,
		`"sampler"`, `"min_p":0.05`, `"server"`, `"build_info":"b9445-af6528e6d"`,
		`"model_mismatch"`,
	} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("wire body missing %s", want)
		}
	}
}

// A failed provenance stamp travels as the recorded failure it is, rather than
// arriving as a plausible-looking empty commit.
func TestParamsFromForwardsProvenanceFailure(t *testing.T) {
	run := sampleRun()
	run.ProvenanceError = "not a git repository"
	p := paramsFrom(run, "/d")
	if p.ProvenanceError != "not a git repository" {
		t.Errorf("provenance error not forwarded: %q", p.ProvenanceError)
	}
	if p.CommitSHA != "" {
		t.Errorf("commit should stay empty on a failed stamp, got %q", p.CommitSHA)
	}
}
