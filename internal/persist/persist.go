// Package persist ships a completed study run to corpos-toolkit for one
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

	"corpos-lab/internal/control"
)

// DefaultToolkitURL is the canonical toolkit HTTP daemon (post-flip).
const DefaultToolkitURL = "http://localhost:3001"

// scoreRow is the flattened per-condition score the toolkit persists — the
// controller's nested verdict is flattened to kind/reason here.
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
	}
	if run.Extraction != nil {
		p.RunAt = run.Extraction.FinishedAt
	}
	if run.Results != nil {
		for _, r := range run.Results.Rows {
			p.Rows = append(p.Rows, scoreRow{
				Item:          r.Item,
				Condition:     string(r.Condition),
				Run:           r.Run,
				VerdictKind:   string(r.Verdict.Kind),
				VerdictReason: r.Verdict.Reason,
				Rationale:     r.Rationale,
			})
		}
	}
	if p.MaterialsHashes == nil {
		p.MaterialsHashes = map[string]string{}
	}
	return p
}

// Client posts study runs to the toolkit's MCP HTTP surface.
type Client struct {
	baseURL string
	project string
	httpc   *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient injects the HTTP transport (tests pass an httptest client).
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpc = c }
}

// NewClient builds a persistence client for the toolkit at baseURL, scoping
// writes to project.
func NewClient(baseURL, project string, opts ...Option) *Client {
	c := &Client{baseURL: baseURL, project: project, httpc: http.DefaultClient}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Record persists run to the toolkit via the study_run_record measure action.
// responsesDir is the on-disk responses pointer. A non-2xx or transport error
// is returned so the caller can decide (the CLI treats it as a warning — the
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
		return fmt.Errorf("persist: toolkit returned %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
