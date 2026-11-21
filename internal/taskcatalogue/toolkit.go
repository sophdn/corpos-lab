package taskcatalogue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Toolkit reads the work ledger through the toolkit service's read-only
// admin.db_query action over HTTP. corpos-lab never opens the toolkit DB
// directly; the ledger stays owned by the toolkit service.
type Toolkit struct {
	BaseURL string
	Project string
	HTTP    *http.Client
}

// pageSize is db_query's row cap per call.
const pageSize = 1000

type queryEnvelope struct {
	Action    string      `json:"action"`
	Project   string      `json:"project"`
	Rationale string      `json:"rationale"`
	Params    queryParams `json:"params"`
}

type queryParams struct {
	SQL      string `json:"sql"`
	Args     []any  `json:"args,omitempty"`
	Limit    int    `json:"limit"`
	MaxBytes int    `json:"max_bytes"`
}

type queryResult struct {
	Rows            [][]any `json:"rows"`
	Truncated       bool    `json:"truncated"`
	TruncatedReason string  `json:"truncated_reason"`
	Error           string  `json:"error"`
}

// Rows runs sql (which must not carry its own LIMIT/OFFSET and must ORDER BY a
// stable key) page by page and returns every row as strings. A page cut short
// by the byte cap is an error, because a silently short page would drop terms.
func (t Toolkit) Rows(ctx context.Context, sql string) ([][]string, error) {
	var out [][]string
	for offset := 0; ; offset += pageSize {
		page, err := t.page(ctx, fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, pageSize, offset))
		if err != nil {
			return nil, err
		}
		out = append(out, page...)
		if len(page) < pageSize {
			return out, nil
		}
	}
}

func (t Toolkit) page(ctx context.Context, sql string) ([][]string, error) {
	body, err := json.Marshal(queryEnvelope{
		Action:    "db_query",
		Project:   t.Project,
		Rationale: "task catalogue: read project names and session scopes for prompt attribution",
		Params:    queryParams{SQL: sql, Limit: pageSize, MaxBytes: 262144},
	})
	if err != nil {
		return nil, fmt.Errorf("toolkit: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(t.BaseURL, "/")+"/mcp/admin", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("toolkit: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	httpc := t.HTTP
	if httpc == nil {
		httpc = http.DefaultClient
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("toolkit: POST /mcp/admin: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("toolkit: read answer: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("toolkit: status %d: %s", resp.StatusCode, capRunes(string(raw), 300))
	}
	var r queryResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("toolkit: non-JSON answer: %s", capRunes(string(raw), 300))
	}
	if r.Error != "" {
		return nil, fmt.Errorf("toolkit: db_query refused: %s", r.Error)
	}
	// A full page reports "row limit N reached", which paging expects. Any
	// other cut (the byte cap) means this page lost rows.
	if r.Truncated && !strings.HasPrefix(r.TruncatedReason, "row limit") {
		return nil, fmt.Errorf("toolkit: page cut short (%s); narrow the query", r.TruncatedReason)
	}
	out := make([][]string, len(r.Rows))
	for i, row := range r.Rows {
		cells := make([]string, len(row))
		for j, c := range row {
			if c != nil {
				cells[j] = fmt.Sprint(c)
			}
		}
		out[i] = cells
	}
	return out, nil
}

// minSlugHyphens is how many hyphens a work-item slug needs before it counts
// as a project name. Short slugs ("fix-gate") are ordinary phrases; a slug with
// two or more hyphens is specific enough that seeing it in a prompt names its
// project.
const minSlugHyphens = 2

// LedgerTerms reads every name the ledger knows that can tie a prompt to a
// project: project ids and names, canonical-name aliases resolved to their
// replacement, and the slugs of chains, bugs, suggestions and proposals.
func (t Toolkit) LedgerTerms(ctx context.Context) ([]Term, error) {
	var terms []Term
	projects, err := t.Rows(ctx, "SELECT id, name FROM projects ORDER BY id")
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, r := range projects {
		known[r[0]] = true
		terms = append(terms, Term{Text: r[0], Project: r[0]}, Term{Text: r[1], Project: r[0]})
	}
	aliases, err := t.Rows(ctx, "SELECT a.alias, COALESCE(NULLIF(e.replacement, ''), e.slug) FROM canon_aliases a JOIN canon_entries e ON e.slug = a.entry_slug WHERE e.kind IN ('project', 'repo') ORDER BY a.alias")
	if err != nil {
		return nil, err
	}
	for _, r := range aliases {
		// A replacement can be prose ("the private toolkit service (gitea …)"); its
		// first word is the project. An alias that resolves to no
		// registered project is dropped rather than invented.
		if f := strings.Fields(r[1]); len(f) > 0 && known[f[0]] {
			terms = append(terms, Term{Text: r[0], Project: f[0]})
		}
	}
	for _, table := range []string{"proj_chain_status", "proj_current_bugs", "proj_current_suggestions", "proj_current_proposals"} {
		rows, err := t.Rows(ctx, "SELECT slug, project_id FROM "+table+" ORDER BY id")
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if strings.Count(r[0], "-") >= minSlugHyphens {
				terms = append(terms, Term{Text: r[0], Project: r[1]})
			}
		}
	}
	return terms, nil
}

// SessionProjects maps each harness session to the project its toolkit calls
// named most often (ties alphabetical).
func (t Toolkit) SessionProjects(ctx context.Context) (map[string]string, error) {
	rows, err := t.Rows(ctx, "SELECT session_id, project_id, COUNT(*) FROM tool_calls WHERE project_id <> '' AND session_id <> '' GROUP BY session_id, project_id ORDER BY session_id, project_id")
	if err != nil {
		return nil, err
	}
	counts := map[string]map[string]int{}
	for _, r := range rows {
		var n int
		if _, err := fmt.Sscan(r[2], &n); err != nil {
			return nil, fmt.Errorf("toolkit: session count %q: %w", r[2], err)
		}
		m, ok := counts[r[0]]
		if !ok {
			m = map[string]int{}
			counts[r[0]] = m
		}
		m[r[1]] += n
	}
	out := make(map[string]string, len(counts))
	for s, c := range counts {
		out[s] = top(c)
	}
	return out, nil
}
