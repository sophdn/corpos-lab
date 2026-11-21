// Package persist ships a completed study run to the private toolkit service for one
// queryable home. It maps the local control.StudyRun into the toolkit's
// `study_run_record` measure action and POSTs it over the MCP HTTP surface —
// the only channel to the toolkit (the DB is never touched directly). Raw
// model responses stay on disk; only their directory pointer is sent.
package persist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"corpos-lab/internal/assay"
	"corpos-lab/internal/control"
	"corpos-lab/internal/model"
	"corpos-lab/internal/substrate"
)

// DefaultToolkitURL is the canonical toolkit HTTP daemon (post-flip).
const DefaultToolkitURL = "http://localhost"

// DefaultTimeout bounds one persist request. The toolkit is a local daemon and
// the write is one event, so a healthy call takes well under a second; the
// slack covers a contended write lock. Without a bound a wedged daemon would
// hang run-study after the run itself had finished.
const DefaultTimeout = 30 * time.Second

// scoreRow is the flattened per-condition score the toolkit persists. The
// toolkit's verdict_kind is a free-form cell, so a probe ships its rubric code
// (or "unscored") there rather than a pass/fail it has no basis to claim.
// verdict_reason stays empty for probe rows: an unjudged capture has no reason
// to give, and stuffing the response text into it is what previously disguised
// unscored runs as failures.
type scoreRow struct {
	Item          string `json:"item"`
	Condition     string `json:"condition"`
	Run           int    `json:"run"`
	VerdictKind   string `json:"verdict_kind"`
	VerdictReason string `json:"verdict_reason"`
	Rationale     string `json:"rationale"`
}

// recordParams is the `study_run_record` action's params: the flattened,
// self-contained run the toolkit folds into its projections.
//
// The wire shape is a contract with the private toolkit service's studyRunInput
// (go/internal/measure/study_run.go), bound by the shared fixture
// testdata/study_run_record.contract.json: contract_test.go asserts this
// struct marshals to exactly the fixture's key set, and the toolkit decodes the
// same fixture strictly. Change a field here and the fixture (canonical in
// the private toolkit service) together.
type recordParams struct {
	Name            string            `json:"name"`
	Assay           string            `json:"assay"`
	ItemID          string            `json:"item_id"`
	Image           string            `json:"image"`
	ImageDigest     string            `json:"image_digest"`
	Status          string            `json:"status"`
	Error           string            `json:"error,omitempty"`
	StudyDigest     string            `json:"study_digest"`
	MaterialsHashes map[string]string `json:"materials_hashes"`
	ModelID         string            `json:"model_id"`
	ModelVersion    string            `json:"model_version"`
	ResponsesDir    string            `json:"responses_dir"`
	RunAt           string            `json:"run_at"`
	Rows            []scoreRow        `json:"rows"`

	// What the run OBSERVED, as distinct from what it declared. ModelID above
	// is an echo of the study definition; these are the server's and the host's
	// own accounts, and they are why the toolkit copy can answer "what ran"
	// rather than only "what was asked for".
	Sampler         assay.Sampling    `json:"sampler"`
	Server          model.ServerProps `json:"server"`
	ModelMismatch   string            `json:"model_mismatch,omitempty"`
	Substrate       substrate.Info    `json:"substrate"`
	CommitSHA       string            `json:"commit_sha,omitempty"`
	CommitDirty     bool              `json:"commit_dirty"`
	ProvenanceError string            `json:"provenance_error,omitempty"`
}

// mcpEnvelope is the `POST /mcp/<surface>` request body.
type mcpEnvelope struct {
	Action    string       `json:"action"`
	Project   string       `json:"project"`
	Rationale string       `json:"rationale"`
	Params    recordParams `json:"params"`
}

// paramsFrom maps a StudyRun into the action params. responsesDir is the
// on-disk pointer to the run's raw responses; runAt is the run timestamp
// (taken from the container's extraction manifest when present).
func paramsFrom(run control.StudyRun, responsesDir string) recordParams {
	p := recordParams{
		Name:            run.Name,
		Assay:           run.Assay,
		ItemID:          run.ItemID,
		Image:           run.Image,
		ImageDigest:     run.ImageDigest,
		Status:          string(run.Status),
		Error:           run.Error,
		StudyDigest:     run.Manifest.StudyDigest,
		MaterialsHashes: run.Manifest.Materials,
		ModelID:         run.Manifest.ModelID,
		ModelVersion:    run.Manifest.ModelVersion,
		ResponsesDir:    responsesDir,
		Rows:            []scoreRow{},

		Substrate:       run.Substrate,
		CommitSHA:       run.Provenance.CommitSHA,
		CommitDirty:     run.Provenance.Dirty,
		ProvenanceError: run.ProvenanceError,
	}
	if run.Results != nil {
		p.Sampler = run.Results.Sampler
		p.Server = run.Results.Server
		p.ModelMismatch = run.Results.ModelMismatch
	}
	if run.Extraction != nil {
		p.RunAt = run.Extraction.FinishedAt
	}
	if run.Results != nil {
		for _, r := range run.Results.Rows {
			p.Rows = append(p.Rows, scoreRow{
				Item:      r.Item,
				Condition: string(r.Condition),
				Run:       r.Run,
				// An empty Score would ship as "" and read as a blank cell;
				// send the explicit sentinel so the DB says "unscored" out loud.
				VerdictKind: string(scoreOrUnscored(r.Score)),
				Rationale:   r.Rationale,
			})
		}
	}
	if p.MaterialsHashes == nil {
		p.MaterialsHashes = map[string]string{}
	}
	return p
}

// scoreOrUnscored maps a probe's rubric code to the wire value, defaulting a
// zero-value Score to the explicit Unscored sentinel. A row that reached the
// toolkit as an empty string would be indistinguishable from a dropped field.
func scoreOrUnscored(s assay.ProbeScore) assay.ProbeScore {
	if s == "" {
		return assay.Unscored
	}
	return s
}

// Client posts study runs to the toolkit's MCP HTTP surface.
type Client struct {
	baseURL string
	project string
	httpc   *http.Client
}

// Option configures a Client.
type Option func(*Client)

// NewClient builds a persistence client for the toolkit at baseURL, scoping
// writes to project. Each request is bounded by DefaultTimeout.
func NewClient(baseURL, project string, opts ...Option) *Client {
	c := &Client{baseURL: baseURL, project: project, httpc: &http.Client{Timeout: DefaultTimeout}}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Record persists run to the toolkit via the study_run_record measure action.
// responsesDir is the on-disk responses pointer. A transport error, a non-2xx
// status, or a 200 whose body is not a study_run_record success (see
// checkRecorded) is returned so the caller can decide (the CLI treats it as a warning — the
// local run record is the durable artifact).
func (c *Client) Record(ctx context.Context, run control.StudyRun, responsesDir string) error {
	env := mcpEnvelope{
		Action:    "study_run_record",
		Project:   c.project,
		Rationale: fmt.Sprintf("persist study run %s (%s)", run.Name, run.Status),
		Params:    paramsFrom(run, responsesDir),
	}
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("persist: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/mcp/measure", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("persist: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("persist: POST /mcp/measure: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("persist: toolkit returned %d: %s", resp.StatusCode, truncate(string(raw)))
	}
	return checkRecorded(raw)
}

// recordResponse is the part of a POST /mcp/measure answer that says whether
// the run was stored. The toolkit answers 200 for every dispatched call, so
// the status code says nothing: a dispatch-level rejection (rationale_required,
// action not implemented, retired_default_project, a handler error) arrives as
// a top-level "error" string, and so does study_run_record's own param error.
// Only a body that carries the run id it was stored under is a success.
type recordResponse struct {
	RunID string `json:"run_id"`
	Error string `json:"error"`
}

// checkRecorded returns nil only when raw is a study_run_record success.
func checkRecorded(raw []byte) error {
	var r recordResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return fmt.Errorf("persist: toolkit answered with a non-JSON body: %s", truncate(string(raw)))
	}
	if r.Error != "" {
		return fmt.Errorf("persist: toolkit rejected the run: %s", truncate(string(raw)))
	}
	if r.RunID == "" {
		return fmt.Errorf("persist: toolkit answer has no run_id: %s", truncate(string(raw)))
	}
	return nil
}

// maxBodyInError caps how much of a toolkit answer an error message quotes.
const maxBodyInError = 300

func truncate(s string) string {
	if len(s) <= maxBodyInError {
		return s
	}
	return s[:maxBodyInError] + "…"
}
