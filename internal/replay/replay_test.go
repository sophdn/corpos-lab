package replay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeRunner scripts commands by a prefix of "name arg arg…". The first
// matching rule answers; an unmatched command succeeds with empty output.
type fakeRunner struct {
	mu    sync.Mutex
	rules []rule
	calls []string
}

type rule struct {
	prefix string
	out    string
	errOut string
	err    error
	effect func(args []string)
}

func (f *fakeRunner) Run(_ context.Context, _ string, _ []string, name string, args ...string) ([]byte, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	line := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, line)
	for _, r := range f.rules {
		if strings.HasPrefix(line, r.prefix) {
			if r.effect != nil {
				r.effect(args)
			}
			return []byte(r.out), []byte(r.errOut), r.err
		}
	}
	return nil, nil, nil
}

func (f *fakeRunner) called(prefix string) bool {
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			return true
		}
	}
	return false
}

func validCase() Case {
	return Case{
		ID: "c-1", Kind: "build-thing", Project: "alpha", Cwd: "/home/u/dev/alpha",
		Repos: []string{"/home/u/dev/alpha"}, Files: []string{"/home/u/docs/notes.md"},
		Turns:  []Turn{{Text: "build it", Trigger: "opener"}},
		Rubric: []string{"a", "b"},
	}
}

func TestValidate(t *testing.T) {
	if err := validCase().Validate(); err != nil {
		t.Fatal(err)
	}
	bad := map[string]func(*Case){
		"id":           func(c *Case) { c.ID = "Bad ID" },
		"kind":         func(c *Case) { c.Kind = "" },
		"no turns":     func(c *Case) { c.Turns = nil },
		"empty turn":   func(c *Case) { c.Turns = []Turn{{Text: " "}} },
		"rubric":       func(c *Case) { c.Rubric = []string{"a"} },
		"cwd":          func(c *Case) { c.Cwd = "rel" },
		"relative":     func(c *Case) { c.Files = []string{"docs/x"} },
		"unclean":      func(c *Case) { c.Files = []string{"/home/u/docs/../x/y"} },
		"shallow":      func(c *Case) { c.Repos = []string{"/home/u"} },
		"listed twice": func(c *Case) { c.Files = []string{"/home/u/dev/alpha"} },
	}
	for name, mutate := range bad {
		c := validCase()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestLoadCase(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "c.json")
	_ = os.WriteFile(good, []byte(`{"id":"c-1","kind":"k","project":"p","cwd":"/a","turns":[{"text":"x","trigger":"opener"}],"rubric":["a","b"],"writes":false}`), 0o600)
	if c, err := LoadCase(good); err != nil || c.ID != "c-1" {
		t.Fatalf("%+v %v", c, err)
	}
	unknown := filepath.Join(dir, "u.json")
	_ = os.WriteFile(unknown, []byte(`{"id":"c-1","surprise":1}`), 0o600)
	if _, err := LoadCase(unknown); err == nil {
		t.Error("unknown field: want error")
	}
	if _, err := LoadCase(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("missing: want error")
	}
}

func testConfig(t *testing.T) Config {
	t.Helper()
	dir := t.TempDir()
	backups := filepath.Join(dir, "backups")
	_ = os.MkdirAll(backups, 0o750)
	for _, n := range []string{"toolkit-20260101T000000Z.db", "toolkit-20260102T000000Z.db", "other.db"} {
		_ = os.WriteFile(filepath.Join(backups, n), nil, 0o600)
	}
	return Config{Root: filepath.Join(dir, "root"), Home: "/home/u", BackupDir: backups, VaultDir: "/home/u/.claude/vault",
		Image: "img", Network: "net", Account: "acct", Port: 3999, CorposSrc: "/src/corpos", GoBin: "/go"}
}

func upRunner() *fakeRunner {
	return &fakeRunner{rules: []rule{
		{prefix: "git -C /home/u/dev/alpha rev-parse HEAD", out: "livehead\n"},
		{prefix: "git -C /home/u/dev/alpha status", out: ""},
		{prefix: "git -C /home/u/dev/alpha remote get-url origin", out: "https://forge/alpha.git\n"},
		{prefix: "git -C /home/u/dev/alpha remote", out: "origin\n"},
		{prefix: "git -C /src/corpos rev-parse", out: "corposhead\n"},
		{prefix: "git -C /src/corpos status", out: " M x.go\n"},
		{prefix: "git -C", out: "copyhead\n"},
		{prefix: "podman run", out: "cid123\n"},
	}}
}

func TestUp(t *testing.T) {
	cfg := testConfig(t)
	r := upRunner()
	st, err := Up(context.Background(), r, cfg, validCase(), time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	if st.ContainerID != "cid123" || st.CorposCommit != "corposhead+dirty" || !strings.HasSuffix(st.CorposBin, "c-1/bin/corpos") ||
		!strings.HasSuffix(st.Backup, "toolkit-20260102T000000Z.db") {
		t.Errorf("state %+v", st)
	}
	if st.RiskGate != "build-test" {
		t.Errorf("risk gate default %q", st.RiskGate)
	}
	if st.DevRoot != "" {
		t.Errorf("a missing dev root leaves the feature off, got %q", st.DevRoot)
	}
	if st.LiveBefore[0].Head != "livehead" || st.CopyHeads["/home/u/dev/alpha"] != "copyhead" {
		t.Errorf("heads %+v %+v", st.LiveBefore, st.CopyHeads)
	}
	copyDir := st.RepoCopies["/home/u/dev/alpha"]
	for _, want := range []string{
		"git clone --quiet --no-hardlinks -- /home/u/dev/alpha " + copyDir,
		"git -C " + copyDir + " remote remove origin",
		"git -C " + copyDir + " remote add origin https://forge/alpha.git",
		"git -C " + copyDir + " remote set-url --push origin " + disabledPush,
		"cp -a -- /home/u/docs/notes.md",
		"cp -a -- /home/u/.claude/vault",
		"/go build -o " + st.CorposBin + " ./cmd/corpos",
	} {
		if !r.called(want) {
			t.Errorf("missing call %q in %v", want, r.calls)
		}
	}
	loaded, err := LoadState(cfg.Root, "c-1")
	if err != nil || loaded.Container != "replay-tk-c-1" {
		t.Errorf("state file %+v %v", loaded, err)
	}
	if _, err := Up(context.Background(), r, cfg, validCase(), time.Unix(100, 0)); err == nil {
		t.Error("a second up for the same case must refuse")
	}
}

func TestUpListsDevRoot(t *testing.T) {
	cfg := testConfig(t)
	cfg.Home = t.TempDir()
	for _, n := range []string{"alpha", "beta"} {
		_ = os.MkdirAll(filepath.Join(cfg.Home, "dev", n), 0o750)
	}
	c := validCase()
	c.Repos, c.Files, c.Cwd = nil, nil, "/tmp"
	st, err := Up(context.Background(), upRunner(), cfg, c, time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if st.DevRoot != filepath.Join(cfg.Home, "dev") || !slices.Equal(st.DevEntries, []string{"alpha", "beta"}) {
		t.Errorf("dev root %q %v", st.DevRoot, st.DevEntries)
	}
}

func TestToolkitRunArgsStayOffTheLiveLedger(t *testing.T) {
	cfg := testConfig(t)
	args := strings.Join(toolkitRunArgs(cfg, State{Dir: "/sbx/c", Port: 3999, Container: "replay-tk-c"}), " ")
	for _, want := range []string{"-p 127.0.0.1:3999:3000", "/sbx/c/toolkit-data:/data", "/home/u/dev:/home/u/dev:ro", "/sbx/c/vault:"} {
		if !strings.Contains(args, want) {
			t.Errorf("missing %q in %s", want, args)
		}
	}
	for _, banned := range []string{".local/share/toolkit", "Secret", "TEAM_GITEA_TOKEN", ".claude/projects"} {
		if strings.Contains(args, banned) {
			t.Errorf("sandbox toolkit must not get %q: %s", banned, args)
		}
	}
}

func TestUpFailures(t *testing.T) {
	type tc struct {
		prefix string
		mutate func(*Config)
	}
	cases := map[string]tc{
		"invalid case":   {mutate: func(*Config) {}},
		"root is a file": {mutate: func(c *Config) { _ = os.WriteFile(c.Root, nil, 0o600) }},
		"live snapshot":  {prefix: "git -C /home/u/dev/alpha rev-parse"},
		"live status":    {prefix: "git -C /home/u/dev/alpha status"},
		"clone":          {prefix: "git clone"},
		"remote remove":  {prefix: "git -C /root/c-1/repos"},
		"remote list":    {prefix: "git -C /home/u/dev/alpha remote"},
		"get-url":        {prefix: "git -C /home/u/dev/alpha remote get-url"},
		"copy file":      {prefix: "cp -a -- /home/u/docs"},
		"no backups":     {mutate: func(c *Config) { c.BackupDir = "/nonexistent" }},
		"copy ledger":    {prefix: "cp -- "},
		"copy vault":     {prefix: "cp -a -- /home/u/.claude"},
		"podman":         {prefix: "podman run"},
		"corpos head":    {prefix: "git -C /src/corpos rev-parse"},
		"corpos status":  {prefix: "git -C /src/corpos status"},
		"corpos build":   {prefix: "/go build"},
	}
	for name, k := range cases {
		cfg := testConfig(t)
		if k.prefix == "git -C /root/c-1/repos" {
			k.prefix = "git -C " + cfg.Root + "/c-1/repos"
		}
		r := upRunner()
		if k.prefix != "" {
			r.rules = append([]rule{{prefix: k.prefix, err: errors.New("boom")}}, r.rules...)
		}
		if k.mutate != nil {
			k.mutate(&cfg)
		}
		c := validCase()
		if name == "invalid case" {
			c.Kind = ""
		}
		if _, err := Up(context.Background(), r, cfg, c, time.Unix(0, 0)); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestBwrapArgs(t *testing.T) {
	st := State{Case: validCase(), Dir: "/sbx/c", Home: "/home/u", Port: 3999, CorposBin: "/bin/corpos",
		RepoCopies: map[string]string{"/home/u/dev/alpha": "/sbx/c/repos/00-alpha"},
		FileCopies: map[string]string{"/home/u/docs/notes.md": "/sbx/c/files/00-notes.md"}}
	st.Case.Profile = "doc-filing"
	st.RiskGate = "build-test"
	st.DevRoot, st.DevEntries = "/home/u/dev", []string{"alpha", "beta"}
	got := strings.Join(BwrapArgs(st, "hello"), " ")
	if !strings.Contains(got, "--bind /sbx/c/devroot /home/u/dev --ro-bind-try /home/u/dev/beta /home/u/dev/beta --bind /sbx/c/repos/00-alpha /home/u/dev/alpha") {
		t.Errorf("dev root binds wrong: %s", got)
	}
	if strings.Contains(got, "--ro-bind-try /home/u/dev/alpha") {
		t.Error("a copied repo must not also be bound read-only")
	}
	for _, want := range []string{
		"--ro-bind / /", "--tmpfs /tmp", "--tmpfs /run/user",
		"--bind /sbx/c/corpos-state /home/u/.local/state/corpos",
		"--bind /sbx/c/repos/00-alpha /home/u/dev/alpha",
		"--bind /sbx/c/files/00-notes.md /home/u/docs/notes.md",
		"--chdir /home/u/dev/alpha -- /bin/corpos -mcp-url http://localhost:3999",
		"-risk-gate build-test", "-profile doc-filing", "-prompt hello",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "-resume") {
		t.Error("first turn must not resume")
	}
	st.RunID = "R1"
	if !strings.Contains(strings.Join(BwrapArgs(st, "x"), " "), "-resume R1 -prompt x") {
		t.Error("later turns resume the session")
	}
}

const stderrFixture = `corpos: session 01ABC (resume with -resume 01ABC)
  [tool work.bug_list -> ok]
  [tool fs.write -> denied]
corpos: session cost $0.0123
  qwen: 42816 in (31184 cached) + 670 out tok → $0.0000 [free]
  gemini: 100 in + 5 out tok → $0.0123
`

func TestParseTurn(t *testing.T) {
	var r TurnResult
	ParseTurn(stderrFixture, &r)
	if r.RunID != "01ABC" || r.CostUSD != 0.0123 || r.InTok != 42916 || r.OutTok != 675 ||
		!slices.Equal(r.Tools, []string{"work.bug_list:ok", "fs.write:denied"}) {
		t.Errorf("%+v", r)
	}
}

func TestRunTurn(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "turns"), 0o750)
	st := State{Case: validCase(), Dir: dir, Home: "/home/u", CorposBin: "/bin/corpos"}
	var resumed TurnResult
	ParseTurn("corpos: resumed session 01XYZ (3 prior messages)\n", &resumed)
	if resumed.RunID != "01XYZ" {
		t.Errorf("resumed run id %q", resumed.RunID)
	}
	clock := time.Unix(0, 0)
	now := func() time.Time { clock = clock.Add(3 * time.Second); return clock }

	r := &fakeRunner{rules: []rule{{prefix: "bwrap", out: "the answer", errOut: stderrFixture}}}
	res, err := RunTurn(context.Background(), r, &st, "go", now)
	if err != nil || res.N != 1 || res.Exit != 0 || res.Seconds != 3 || st.RunID != "01ABC" || res.Answer != "the answer" {
		t.Fatalf("%+v %v", res, err)
	}
	if b, _ := os.ReadFile(res.ErrFile); string(b) != stderrFixture {
		t.Error("stderr not saved")
	}

	exitErr := exec.Command("sh", "-c", "exit 3").Run()
	r = &fakeRunner{rules: []rule{{prefix: "bwrap", err: exitErr}}}
	res, err = RunTurn(context.Background(), r, &st, "again", now)
	if err != nil || res.Exit != 3 || res.N != 2 || len(st.Turns) != 2 {
		t.Errorf("exit code: %+v %v", res, err)
	}
	r = &fakeRunner{rules: []rule{{prefix: "bwrap", err: errors.New("no bwrap")}}}
	res, _ = RunTurn(context.Background(), r, &st, "x", now)
	if res.Exit != -1 || res.ErrorMsg != "no bwrap" {
		t.Errorf("start failure: %+v", res)
	}

	bad := State{Case: validCase(), Dir: filepath.Join(dir, "missing")}
	if _, err := RunTurn(context.Background(), &fakeRunner{}, &bad, "x", now); err == nil {
		t.Error("unwritable turns dir: want error")
	}
	if excerpt(strings.Repeat("a", 10), 4) != "aaaa…" {
		t.Error("excerpt")
	}
}

// fakeLedger answers by substring of the SQL.
type fakeLedger map[string][][]string

func (f fakeLedger) Rows(_ context.Context, sql string) ([][]string, error) {
	for k, v := range f {
		if strings.Contains(sql, k) {
			if v == nil {
				return nil, fmt.Errorf("query failed")
			}
			return v, nil
		}
	}
	return [][]string{{"0"}}, nil
}

func checkState(t *testing.T) State {
	t.Helper()
	return State{Case: validCase(), Dir: t.TempDir(), StartedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), RunID: "R",
		LiveBefore: []RepoState{{Path: "/home/u/dev/alpha", Head: "h", Status: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}},
		RepoCopies: map[string]string{"/home/u/dev/alpha": "/sbx/a"}, CopyHeads: map[string]string{"/home/u/dev/alpha": "c0"},
		Turns: []TurnResult{{CostUSD: 0.5, InTok: 10, OutTok: 2, Exit: 0, Tools: []string{"a:ok"}, Answer: "one"},
			{CostUSD: 0.25, InTok: 1, OutTok: 1, Exit: 1, Answer: "two"}}}
}

func TestCheck(t *testing.T) {
	st := checkState(t)
	r := &fakeRunner{rules: []rule{
		{prefix: "git -C /home/u/dev/alpha rev-parse", out: "moved\n"},
		{prefix: "git -C /sbx/a log --oneline --branches --not c0", out: "abc1 add sentinel\n"},
		{prefix: "git -C /sbx/a status", out: "?? docs/x.md\n"},
	}}
	sandbox := fakeLedger{"SELECT DISTINCT": {{"suggestion", "sentinel-a", "alpha"}, {"bug", "it's-fine", "alpha"}}}
	live := fakeLedger{"entity_slug = 'sentinel-a'": {{"1"}}}
	row, err := Check(context.Background(), r, st, sandbox, live)
	if err != nil {
		t.Fatal(err)
	}
	if row.CostUSD != 0.75 || row.InTok != 11 || !slices.Equal(row.Exits, []int{0, 1}) || !strings.Contains(row.Answer, "one\n---\ntwo") {
		t.Errorf("totals %+v", row)
	}
	if len(row.Entities) != 2 || len(row.LedgerLeaks) != 1 || row.LedgerLeaks[0].Slug != "sentinel-a" {
		t.Errorf("leaks %+v", row)
	}
	if !slices.Equal(row.LiveMoved, []string{"/home/u/dev/alpha"}) {
		t.Errorf("live moved %v", row.LiveMoved)
	}
	if len(row.RepoChanges) != 1 || row.RepoChanges[0].Commits != "abc1 add sentinel" {
		t.Errorf("changes %+v", row.RepoChanges)
	}
	if sqlQuote("it's") != "it''s" {
		t.Error("sqlQuote")
	}
}

func TestCheckErrors(t *testing.T) {
	st := checkState(t)
	ok := fakeLedger{"SELECT DISTINCT": {{"bug", "b", "p"}}}
	cases := map[string]struct {
		r       *fakeRunner
		sandbox fakeLedger
		live    fakeLedger
	}{
		"sandbox": {&fakeRunner{}, fakeLedger{"SELECT DISTINCT": nil}, fakeLedger{}},
		"live":    {&fakeRunner{}, ok, fakeLedger{"COUNT": nil}},
		"snapshot": {&fakeRunner{rules: []rule{{prefix: "git -C /home/u", err: errors.New("x")}}},
			fakeLedger{"SELECT DISTINCT": {}}, fakeLedger{}},
		"copy log": {&fakeRunner{rules: []rule{{prefix: "git -C /sbx/a log", err: errors.New("x")}}},
			fakeLedger{"SELECT DISTINCT": {}}, fakeLedger{}},
		"copy status": {&fakeRunner{rules: []rule{{prefix: "git -C /sbx/a status", err: errors.New("x")}}},
			fakeLedger{"SELECT DISTINCT": {}}, fakeLedger{}},
	}
	for name, c := range cases {
		if _, err := Check(context.Background(), c.r, st, c.sandbox, c.live); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestAppendRowAndTeardown(t *testing.T) {
	dir := t.TempDir()
	ledger := filepath.Join(dir, "l.jsonl")
	for i := 0; i < 2; i++ {
		if err := AppendRow(ledger, Row{Case: "c"}); err != nil {
			t.Fatal(err)
		}
	}
	if b, _ := os.ReadFile(ledger); strings.Count(string(b), "\n") != 2 {
		t.Errorf("ledger %q", b)
	}
	if err := AppendRow(filepath.Join(dir, "no", "x"), Row{}); err == nil {
		t.Error("bad path: want error")
	}

	st := State{Dir: filepath.Join(dir, "sbx"), Container: "replay-tk-c"}
	for _, sub := range []string{"repos", "files", "turns"} {
		_ = os.MkdirAll(filepath.Join(st.Dir, sub), 0o750)
	}
	r := &fakeRunner{}
	if err := Teardown(context.Background(), r, st); err != nil {
		t.Fatal(err)
	}
	if !r.called("podman rm -f replay-tk-c") {
		t.Error("container not stopped")
	}
	if _, err := os.Stat(filepath.Join(st.Dir, "repos")); !os.IsNotExist(err) {
		t.Error("repos copy not removed")
	}
	if _, err := os.Stat(filepath.Join(st.Dir, "turns")); err != nil {
		t.Error("turn transcripts must be kept")
	}
	if err := Teardown(context.Background(), &fakeRunner{rules: []rule{{prefix: "podman", err: errors.New("x")}}}, st); err == nil {
		t.Error("podman failure: want error")
	}
	if err := Teardown(context.Background(), r, State{}); err == nil {
		t.Error("empty dir must refuse")
	}
}

type flakyLedger struct{ fails int }

func (f *flakyLedger) Rows(context.Context, string) ([][]string, error) {
	if f.fails > 0 {
		f.fails--
		return nil, errors.New("not yet")
	}
	return nil, nil
}

func TestWaitReady(t *testing.T) {
	if err := WaitReady(context.Background(), &flakyLedger{fails: 2}, time.Second, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := WaitReady(context.Background(), &flakyLedger{fails: 1 << 30}, 5*time.Millisecond, time.Millisecond); err == nil {
		t.Error("timeout: want error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := WaitReady(ctx, &flakyLedger{fails: 1 << 30}, time.Second, time.Second); !errors.Is(err, context.Canceled) {
		t.Errorf("cancel: %v", err)
	}
}

func TestStateFiles(t *testing.T) {
	if _, err := LoadState(t.TempDir(), "none"); err == nil {
		t.Error("missing state: want error")
	}
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "c"), 0o750)
	_ = os.WriteFile(filepath.Join(dir, "c", StateFile), []byte("{bad"), 0o600)
	if _, err := LoadState(dir, "c"); err == nil {
		t.Error("bad state: want error")
	}
	if err := SaveState(State{Dir: filepath.Join(dir, "nope")}); err == nil {
		t.Error("unwritable: want error")
	}
}

func TestExecRunner(t *testing.T) {
	out, _, err := ExecRunner{}.Run(context.Background(), t.TempDir(), []string{"REPLAY_X=1"}, "sh", "-c", "echo $REPLAY_X")
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		t.Errorf("%q %v", out, err)
	}
}
