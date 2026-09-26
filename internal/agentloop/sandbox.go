package agentloop

import (
	"fmt"
	"sort"
	"strings"
)

// queryDataFile is the reserved sandbox entry run_query reads. A scenario whose
// correct action is a lookup (e.g. query the accounts table) ships its queryable
// rows here; run_query returns the lines that contain the query text. A scenario
// that needs no query tool simply omits it.
const queryDataFile = "_db.txt"

// Sandbox is the in-memory filesystem the loop's tools act on. It holds a
// path→contents map and nothing else: no real files, no shell, so it runs
// unchanged inside the distroless container and is fully sans-IO testable. An
// edit records itself, so the loop can report how many acting calls a run made.
type Sandbox struct {
	files map[string]string
	edits int
}

// NewSandbox builds a sandbox over a copy of files, so a run's edits never leak
// into the fixture or into another run.
func NewSandbox(files map[string]string) *Sandbox {
	cp := make(map[string]string, len(files))
	for k, v := range files {
		cp[k] = v
	}
	return &Sandbox{files: cp}
}

// Edits reports how many edit_file calls the run made — the count of acting
// tool calls, distinct from read/list/query observations.
func (s *Sandbox) Edits() int { return s.edits }

// Snapshot returns a copy of the sandbox's current contents, for provenance:
// the run record keeps the sandbox state the subject left behind.
func (s *Sandbox) Snapshot() map[string]string {
	out := make(map[string]string, len(s.files))
	for k, v := range s.files {
		out[k] = v
	}
	return out
}

// Do executes one tool action and returns the observation text the harness
// feeds back. It handles only the acting kinds (list/read/query/edit) plus the
// unknown-tool case; the loop handles ActionFinal and ActionNone itself.
func (s *Sandbox) Do(a Action) string {
	switch a.Kind {
	case ActionList:
		return s.list(a.Arg)
	case ActionRead:
		if body, ok := s.files[a.Arg]; ok {
			return body
		}
		return fmt.Sprintf("file not found: %s", a.Arg)
	case ActionQuery:
		return s.query(a.Arg)
	case ActionEdit:
		s.files[a.Arg] = a.Content
		s.edits++
		return fmt.Sprintf("wrote %s (%d bytes)", a.Arg, len(a.Content))
	case ActionUnknown:
		return fmt.Sprintf("unknown tool: %s (tools: list_files, read_file, run_query, edit_file)", a.Tool)
	default:
		return ""
	}
}

// list returns the immediate names under dir (one directory level, not a deep
// walk), so the subject sees a directory the way `ls` would show it. "." and ""
// mean the root. The reserved query data file is hidden from listings.
func (s *Sandbox) list(dir string) string {
	prefix := strings.TrimSuffix(dir, "/")
	if prefix == "." {
		prefix = ""
	}
	if prefix != "" {
		prefix += "/"
	}
	seen := map[string]bool{}
	for path := range s.files {
		if path == queryDataFile {
			continue
		}
		if prefix != "" && !strings.HasPrefix(path, prefix) {
			continue
		}
		rest := strings.TrimPrefix(path, prefix)
		if rest == "" {
			continue
		}
		// The immediate child is the first segment; a deeper path shows as its
		// directory name with a trailing slash.
		if i := strings.Index(rest, "/"); i >= 0 {
			seen[rest[:i+1]] = true
		} else {
			seen[rest] = true
		}
	}
	if len(seen) == 0 {
		return "(empty directory)"
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, "\n")
}

// query returns the lines of the reserved data file that contain the query
// text (case-insensitive). It is the minimal stand-in for a table lookup a
// consult/verify scenario needs; a sandbox without the data file reports so.
func (s *Sandbox) query(q string) string {
	db, ok := s.files[queryDataFile]
	if !ok {
		return "no query data source"
	}
	needle := strings.ToLower(strings.TrimSpace(q))
	var hits []string
	for _, line := range strings.Split(db, "\n") {
		if needle == "" || strings.Contains(strings.ToLower(line), needle) {
			hits = append(hits, line)
		}
	}
	if len(hits) == 0 {
		return "no rows match"
	}
	return strings.Join(hits, "\n")
}
