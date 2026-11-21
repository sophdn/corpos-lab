package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Ledger reads a toolkit ledger with one bounded SELECT at a time. It is the
// read side the leak check needs; taskcatalogue.Toolkit satisfies it.
type Ledger interface {
	Rows(ctx context.Context, sql string) ([][]string, error)
}

// Entity is a ledger item a replay touched.
type Entity struct {
	Kind    string `json:"kind"`
	Slug    string `json:"slug"`
	Project string `json:"project"`
}

// RepoChange is what a replay did to one repo copy.
type RepoChange struct {
	Repo    string `json:"repo"`
	Commits string `json:"commits,omitempty"`
	Status  string `json:"status,omitempty"`
}

// Row is one replay-ledger record: dogfood.jsonl's signals (cost, tokens,
// tools, answer) plus the case, its turns, what it changed, and the leak
// checks. Outcome stays empty until the case is judged.
type Row struct {
	TS          time.Time    `json:"ts"`
	Case        string       `json:"case"`
	Kind        string       `json:"kind"`
	Project     string       `json:"project"`
	RunID       string       `json:"run_id"`
	Turns       int          `json:"turns"`
	CostUSD     float64      `json:"cost_usd"`
	InTok       int          `json:"in_tok"`
	OutTok      int          `json:"out_tok"`
	Exits       []int        `json:"exits"`
	Tools       []string     `json:"tools,omitempty"`
	Answer      string       `json:"answer_excerpt"`
	RepoChanges []RepoChange `json:"repo_changes,omitempty"`
	Entities    []Entity     `json:"ledger_entities,omitempty"`
	LedgerLeaks []Entity     `json:"ledger_leaks,omitempty"`
	LiveMoved   []string     `json:"live_repos_moved,omitempty"`
	CorposVer   string       `json:"corpos_version,omitempty"`
	CorposSHA   string       `json:"corpos_commit"`
	Backup      string       `json:"ledger_backup"`
	Outcome     string       `json:"outcome"`
	MissingCap  string       `json:"missing_capability,omitempty"`
	Adaptations string       `json:"adaptations,omitempty"`
}

// Check runs the leak checks and builds the ledger row, without tearing
// anything down, so a dry run can be inspected before down.
//
// Ledger leak: every entity the sandbox ledger recorded an event for since
// the sandbox started must have no event in the live ledger in the same
// window. Repo leak: each live repo's HEAD and status must match the
// snapshot taken at up. A live repo can also move because the principal is
// working in it, so a moved repo is reported, not inferred to be a leak; the
// bubblewrap fence is what makes a write impossible.
func Check(ctx context.Context, r Runner, st State, sandbox, live Ledger) (Row, error) {
	row := Row{TS: st.StartedAt, Case: st.Case.ID, Kind: st.Case.Kind, Project: st.Case.Project,
		RunID: st.RunID, Turns: len(st.Turns), CorposVer: st.CorposVer, CorposSHA: st.CorposCommit, Backup: st.Backup}
	var answers []string
	for _, t := range st.Turns {
		row.CostUSD += t.CostUSD
		row.InTok += t.InTok
		row.OutTok += t.OutTok
		row.Exits = append(row.Exits, t.Exit)
		row.Tools = append(row.Tools, t.Tools...)
		answers = append(answers, t.Answer)
	}
	row.Answer = excerpt(strings.Join(answers, "\n---\n"), 1200)

	since := st.StartedAt.Format("2006-01-02T15:04:05")
	ents, err := sandbox.Rows(ctx, fmt.Sprintf(
		"SELECT DISTINCT entity_kind, entity_slug, entity_project_id FROM events WHERE ts >= '%s' AND entity_slug <> '' ORDER BY 1, 2, 3", since))
	if err != nil {
		return row, fmt.Errorf("read sandbox ledger: %w", err)
	}
	for _, e := range ents {
		ent := Entity{Kind: e[0], Slug: e[1], Project: e[2]}
		row.Entities = append(row.Entities, ent)
		hits, err := live.Rows(ctx, fmt.Sprintf(
			"SELECT COUNT(*) FROM events WHERE ts >= '%s' AND entity_kind = '%s' AND entity_slug = '%s' AND entity_project_id = '%s'",
			since, sqlQuote(ent.Kind), sqlQuote(ent.Slug), sqlQuote(ent.Project)))
		if err != nil {
			return row, fmt.Errorf("read live ledger: %w", err)
		}
		if len(hits) == 1 && len(hits[0]) == 1 && hits[0][0] != "0" {
			row.LedgerLeaks = append(row.LedgerLeaks, ent)
		}
	}

	after, err := SnapshotRepos(ctx, r, st.Case.Repos)
	if err != nil {
		return row, err
	}
	for i, b := range st.LiveBefore {
		if i < len(after) && after[i] != b {
			row.LiveMoved = append(row.LiveMoved, b.Path)
		}
	}
	for _, real := range sortedKeys(st.RepoCopies) {
		cp := st.RepoCopies[real]
		// Every local branch, so commits made on a worktree branch count too.
		commits, err := gitOut(ctx, r, cp, "log", "--oneline", "--branches", "--not", st.CopyHeads[real])
		if err != nil {
			return row, err
		}
		status, err := gitOut(ctx, r, cp, "status", "--porcelain")
		if err != nil {
			return row, err
		}
		if commits != "" || status != "" {
			row.RepoChanges = append(row.RepoChanges, RepoChange{Repo: real, Commits: commits, Status: status})
		}
	}
	return row, nil
}

func sqlQuote(s string) string { return strings.ReplaceAll(s, "'", "''") }

// AppendRow appends row to the replay ledger (JSONL).
func AppendRow(path string, row Row) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // caller-named ledger
	if err != nil {
		return fmt.Errorf("open replay ledger: %w", err)
	}
	if err := json.NewEncoder(f).Encode(row); err != nil {
		_ = f.Close()
		return fmt.Errorf("write replay ledger: %w", err)
	}
	return f.Close()
}

// Teardown stops the sandbox toolkit and deletes the copies (repos, files,
// ledger, vault, cache). It keeps state.json, the turn transcripts and
// corpos's session DBs, which the judge reads.
func Teardown(ctx context.Context, r Runner, st State) error {
	if _, errOut, err := r.Run(ctx, "", nil, "podman", "rm", "-f", st.Container); err != nil {
		return fmt.Errorf("stop sandbox toolkit: %w: %s", err, strings.TrimSpace(string(errOut)))
	}
	for _, sub := range []string{"repos", "files", "toolkit-data", "vault", "cache", "devroot"} {
		p := filepath.Join(st.Dir, sub)
		if filepath.Dir(p) != st.Dir || st.Dir == "" {
			return fmt.Errorf("refusing to remove %s", p)
		}
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return nil
}

// WaitReady polls the sandbox ledger until it answers a query or wait runs out.
func WaitReady(ctx context.Context, l Ledger, wait, every time.Duration) error {
	deadline := time.Now().Add(wait)
	for {
		_, err := l.Rows(ctx, "SELECT 1")
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("sandbox toolkit not ready after %s: %w", wait, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(every):
		}
	}
}

func sortStrings(s []string) { sort.Strings(s) }
