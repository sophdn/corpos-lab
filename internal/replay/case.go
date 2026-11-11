// Package replay runs one catalogue case through corpos in a sandbox, so the
// principal's real task mix can be replayed without touching live state
// (corpos-lab chain popular-task-catalogue, task build-replay-harness).
//
// A sandbox has three fences:
//
//   - Files: corpos runs under bubblewrap with the host root read-only. Each
//     repo and personal file the case touches is a throwaway copy, bound AT
//     its real path, so a prompt can name real paths and every write lands in
//     the copy. Any other write fails.
//   - Ledger: a throwaway the toolkit service container serves a copy of the
//     newest nightly ledger backup on its own port, and corpos's -mcp-url
//     points at it. The live ledger file is never opened.
//   - Git: each repo copy has its push URLs disabled.
//
// The verbs are up (build the sandbox), turn (send one principal message;
// the first starts a corpos session, later ones resume it) and down (run the
// leak checks, append the replay-ledger row, tear the sandbox down).
package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Turn is one principal message in a case.
type Turn struct {
	Text string `json:"text"`
	// Trigger says what the original message was answering: "opener",
	// "approve", "steer" or "report", plus free text on what prompted it.
	Trigger string `json:"trigger"`
	// Adapt says how the principal may adapt this turn to what corpos
	// actually did (the original reply it answered will not recur).
	Adapt string `json:"adapt,omitempty"`
}

// Case is one replayable task drawn from the catalogue.
type Case struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	SourcePrompts []int  `json:"source_prompts"`
	// Project is the ledger project corpos's tool calls are scoped to.
	Project string `json:"project"`
	// Cwd is where the principal was when they typed the opener.
	Cwd string `json:"cwd"`
	// Repos are git checkouts the case may read or change; each is copied.
	Repos []string `json:"repos,omitempty"`
	// Files are other paths (files or directories) the case touches; copied.
	Files []string `json:"files,omitempty"`
	// Writes says whether a correct run changes anything.
	Writes bool `json:"writes"`
	// Profile pins a corpos job-profile; empty lets corpos pick.
	Profile string `json:"profile,omitempty"`
	Turns   []Turn `json:"turns"`
	// Rubric is 2-5 checkable outcomes of a good run.
	Rubric []string `json:"rubric"`
	Notes  string   `json:"notes,omitempty"`
}

var caseIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

// Validate checks a case is complete and safe to sandbox: absolute paths, no
// path listed twice, nothing that would bind over a whole system tree.
func (c Case) Validate() error {
	if !caseIDPattern.MatchString(c.ID) {
		return fmt.Errorf("case id %q must be lowercase letters, digits and hyphens", c.ID)
	}
	if c.Kind == "" || c.Project == "" {
		return fmt.Errorf("case %s: kind and project are required", c.ID)
	}
	if len(c.Turns) == 0 || strings.TrimSpace(c.Turns[0].Text) == "" {
		return fmt.Errorf("case %s: needs at least one non-empty turn", c.ID)
	}
	if n := len(c.Rubric); n < 2 || n > 5 {
		return fmt.Errorf("case %s: rubric has %d outcomes, want 2-5", c.ID, n)
	}
	if !filepath.IsAbs(c.Cwd) {
		return fmt.Errorf("case %s: cwd %q must be absolute", c.ID, c.Cwd)
	}
	seen := map[string]bool{}
	for _, p := range append(append([]string{}, c.Repos...), c.Files...) {
		clean := filepath.Clean(p)
		if !filepath.IsAbs(p) || clean != p {
			return fmt.Errorf("case %s: path %q must be absolute and clean", c.ID, p)
		}
		if strings.Count(clean, string(filepath.Separator)) < 3 {
			return fmt.Errorf("case %s: path %q is too shallow to copy", c.ID, p)
		}
		if seen[clean] {
			return fmt.Errorf("case %s: path %q listed twice", c.ID, p)
		}
		seen[clean] = true
	}
	return nil
}

// LoadCase reads and validates a case JSON file.
func LoadCase(path string) (Case, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // the caller names the case file
	if err != nil {
		return Case{}, fmt.Errorf("read case: %w", err)
	}
	var c Case
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Case{}, fmt.Errorf("case %s: %w", path, err)
	}
	return c, c.Validate()
}
