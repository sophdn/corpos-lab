package agentloop

import (
	"strings"
	"testing"
)

func fixture() map[string]string {
	return map[string]string{
		"CHANGELOG.md":    "# Changelog\n\n## v1.4.0\n- Added BaseFilter\n",
		"filters.py":      "class BaseFilter: ...",
		"README.md":       "# filters",
		"tests/test_x.py": "def test(): ...",
		"tests/test_y.py": "def test(): ...",
		queryDataFile:     "id=1 region=EU\nid=2 region=US\nid=3 region=EU",
	}
}

func TestSandboxListRootShowsImmediateChildren(t *testing.T) {
	sb := NewSandbox(fixture())
	got := sb.list(".")
	// The reserved query file is hidden; tests/ collapses to one dir entry.
	want := "CHANGELOG.md\nREADME.md\nfilters.py\ntests/"
	if got != want {
		t.Fatalf("list(.) =\n%q\nwant\n%q", got, want)
	}
}

func TestSandboxListSubdir(t *testing.T) {
	sb := NewSandbox(fixture())
	got := sb.list("tests/")
	if got != "test_x.py\ntest_y.py" {
		t.Fatalf("list(tests/) = %q", got)
	}
}

func TestSandboxListEmpty(t *testing.T) {
	sb := NewSandbox(fixture())
	if got := sb.list("nonexistent/"); got != "(empty directory)" {
		t.Fatalf("list(nonexistent/) = %q, want empty-directory note", got)
	}
}

func TestSandboxReadHitAndMiss(t *testing.T) {
	sb := NewSandbox(fixture())
	if got := sb.Do(Action{Kind: ActionRead, Arg: "README.md"}); got != "# filters" {
		t.Fatalf("read README.md = %q", got)
	}
	if got := sb.Do(Action{Kind: ActionRead, Arg: "missing.md"}); !strings.Contains(got, "file not found") {
		t.Fatalf("read missing = %q, want not-found", got)
	}
}

func TestSandboxEditWritesAndCounts(t *testing.T) {
	sb := NewSandbox(fixture())
	obs := sb.Do(Action{Kind: ActionEdit, Arg: "CHANGELOG.md", Content: "# Changelog\n## v1.5.0\n"})
	if !strings.Contains(obs, "wrote CHANGELOG.md") {
		t.Fatalf("edit observation = %q", obs)
	}
	if sb.Edits() != 1 {
		t.Fatalf("edits = %d, want 1", sb.Edits())
	}
	if got := sb.Snapshot()["CHANGELOG.md"]; !strings.Contains(got, "v1.5.0") {
		t.Fatalf("snapshot CHANGELOG.md = %q, want the new v1.5.0 body", got)
	}
}

func TestSandboxQueryFiltersRows(t *testing.T) {
	sb := NewSandbox(fixture())
	got := sb.Do(Action{Kind: ActionQuery, Arg: "region=EU"})
	if got != "id=1 region=EU\nid=3 region=EU" {
		t.Fatalf("query region=EU = %q", got)
	}
	if miss := sb.Do(Action{Kind: ActionQuery, Arg: "region=ZZ"}); miss != "no rows match" {
		t.Fatalf("query miss = %q", miss)
	}
}

func TestSandboxQueryWithoutDataSource(t *testing.T) {
	sb := NewSandbox(map[string]string{"a.md": "x"})
	if got := sb.Do(Action{Kind: ActionQuery, Arg: "anything"}); got != "no query data source" {
		t.Fatalf("query without db = %q", got)
	}
}

func TestSandboxUnknownTool(t *testing.T) {
	sb := NewSandbox(fixture())
	if got := sb.Do(Action{Kind: ActionUnknown, Tool: "git_status"}); !strings.Contains(got, "unknown tool: git_status") {
		t.Fatalf("unknown tool obs = %q", got)
	}
}

// A run's edits must not mutate the fixture map the caller passed.
func TestSandboxIsolatesFixture(t *testing.T) {
	f := fixture()
	sb := NewSandbox(f)
	sb.Do(Action{Kind: ActionEdit, Arg: "CHANGELOG.md", Content: "changed"})
	if f["CHANGELOG.md"] == "changed" {
		t.Fatal("edit leaked into the caller's fixture map")
	}
}

// Do on a control kind (Final/None) returns no observation.
func TestSandboxDoOnControlKindsIsEmpty(t *testing.T) {
	sb := NewSandbox(fixture())
	if got := sb.Do(Action{Kind: ActionFinal}); got != "" {
		t.Fatalf("Do(final) = %q, want empty", got)
	}
	if got := sb.Do(Action{Kind: ActionNone}); got != "" {
		t.Fatalf("Do(none) = %q, want empty", got)
	}
}

// bug 1441: an edit with an empty path wrote 387 bytes to the path "", and an
// edit with no content wiped a real file to 0 bytes, both reported as success.
// Each is a lost call, so the sandbox now refuses it with an error observation,
// writes nothing, and does not count an edit.
func TestSandboxEditRefusesEmptyPathAndEmptyContent(t *testing.T) {
	sb := NewSandbox(map[string]string{"config/services.yaml": "a: 1\n"})
	for name, a := range map[string]Action{
		"empty path":    {Kind: ActionEdit, Arg: "", Content: "x"},
		"blank path":    {Kind: ActionEdit, Arg: "  ", Content: "x"},
		"empty content": {Kind: ActionEdit, Arg: "config/services.yaml", Content: ""},
	} {
		obs := sb.Do(a)
		if !strings.HasPrefix(obs, "error: ") {
			t.Errorf("%s: obs = %q, want an error observation", name, obs)
		}
	}
	if sb.Edits() != 0 {
		t.Errorf("edits = %d, want 0", sb.Edits())
	}
	snap := sb.Snapshot()
	if len(snap) != 1 || snap["config/services.yaml"] != "a: 1\n" {
		t.Errorf("sandbox changed: %v", snap)
	}
}

// bug 1441: the unknown-tool error listed run_query even when the run offers no
// query data, steering the subject to a tool it cannot use.
func TestSandboxUnknownToolListsOnlyOfferedTools(t *testing.T) {
	plain := NewSandbox(map[string]string{"a.md": "x"}).Do(Action{Kind: ActionUnknown, Tool: "cat"})
	if strings.Contains(plain, "run_query") || !strings.Contains(plain, "list_files, read_file, edit_file") {
		t.Errorf("no-query sandbox obs = %q, want run_query left out", plain)
	}
	withDB := NewSandbox(map[string]string{queryDataFile: "row"}).Do(Action{Kind: ActionUnknown, Tool: "cat"})
	if !strings.Contains(withDB, "list_files, read_file, run_query, edit_file") {
		t.Errorf("query sandbox obs = %q, want run_query listed", withDB)
	}
}
