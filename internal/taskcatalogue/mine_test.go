package taskcatalogue

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"corpos-lab/internal/model"
)

func mineFixture(t *testing.T) (MineConfig, func()) {
	t.Helper()
	dir := t.TempDir()
	hist := filepath.Join(dir, "history.jsonl")
	if err := os.WriteFile(hist, []byte(historyFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	pastes := filepath.Join(dir, "pastes")
	if err := os.MkdirAll(pastes, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pastes, "abc.txt"), []byte("from the cache"), 0o600); err != nil {
		t.Fatal(err)
	}
	tax := filepath.Join(dir, "taxonomy.toml")
	if err := os.WriteFile(tax, []byte(taxFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	srv := fakeLedger{
		"FROM projects":   rows([]any{"widgets", "widgets"}),
		"FROM tool_calls": rows([]any{"s2", "widgets", 1}),
	}.server(t)
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return MineConfig{
		HistoryPath:  hist,
		PasteDir:     pastes,
		TaxonomyPath: tax,
		OutDir:       filepath.Join(dir, "out"),
		Toolkit:      Toolkit{BaseURL: srv.URL},
		Client:       &seqClient{},
		Params:       LabelParams(),
		Workers:      2,
		Commit:       map[string]string{"commit_sha": "abc"},
		Log:          &strings.Builder{},
		Now:          func() time.Time { clock = clock.Add(time.Second); return clock },
	}, srv.Close
}

func TestMineThenBuild(t *testing.T) {
	cfg, done := mineFixture(t)
	defer done()
	cfg.Limit = 2
	rec, err := Mine(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Prompts != 4 || rec.Pending != 2 || rec.Written != 2 || rec.Sessions != 1 || rec.Props.ModelAlias != "m" {
		t.Errorf("first run %+v", rec)
	}
	// A preview aggregate is allowed with -partial, refused without it.
	tax := mustTax(t)
	if _, err := Build(cfg.OutDir, tax, 0.8, false); err == nil {
		t.Error("unlabelled prompts must fail a non-partial build")
	}
	if s, err := Build(cfg.OutDir, tax, 0.8, true); err != nil || s.Unlabelled != 2 {
		t.Errorf("partial build: %+v %v", s, err)
	}

	// The second run resumes: only the two remaining prompts are labelled.
	cfg.Limit = 0
	rec, err = Mine(context.Background(), cfg)
	if err != nil || rec.Pending != 2 {
		t.Fatalf("resume %+v %v", rec, err)
	}
	s, err := Build(cfg.OutDir, tax, 0.8, false)
	if err != nil {
		t.Fatal(err)
	}
	if s.Total != 4 || s.ByRole["control"] != 1 || s.Openers != 3 {
		t.Errorf("summary %+v", s)
	}
	prompts, err := os.ReadFile(filepath.Join(cfg.OutDir, PromptsFile))
	if err != nil || !strings.Contains(string(prompts), "from the cache") || !strings.Contains(string(prompts), `"source":"session-toolkit"`) {
		t.Errorf("prompts file: %s %v", prompts, err)
	}
	runs, _ := os.ReadDir(filepath.Join(cfg.OutDir, RunsDir))
	if len(runs) != 2 {
		t.Errorf("want one run record per run, got %d", len(runs))
	}
}

// cancellingClient fails Props and cancels the run on its first Generate, the
// way a SIGINT lands mid-labelling.
type cancellingClient struct {
	seqClient
	cancel context.CancelFunc
}

func (*cancellingClient) Props(context.Context) (model.ServerProps, error) {
	return model.ServerProps{}, errors.New("no props")
}

func (c *cancellingClient) Generate(ctx context.Context, _ string, _ model.GenParams) (model.Response, error) {
	c.cancel()
	return model.Response{}, ctx.Err()
}

func TestMineRecordsPropsFailureAndInterruption(t *testing.T) {
	cfg, done := mineFixture(t)
	defer done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg.Client = &cancellingClient{cancel: cancel}
	cfg.Now = nil
	rec, err := Mine(ctx, cfg)
	if !errors.Is(err, context.Canceled) || rec.PropsErr != "no props" || rec.Interrupted == "" {
		t.Errorf("rec %+v err %v", rec, err)
	}
}

func TestMineErrors(t *testing.T) {
	cases := map[string]func(*MineConfig){
		"no history":  func(c *MineConfig) { c.HistoryPath = "/nonexistent/h" },
		"no taxonomy": func(c *MineConfig) { c.TaxonomyPath = "/nonexistent/t" },
		"bad taxonomy": func(c *MineConfig) {
			_ = os.WriteFile(c.TaxonomyPath, []byte("[[role]\n"), 0o600)
		},
		"ledger down":   func(c *MineConfig) { c.Toolkit.BaseURL = "http://127.0.0.1:1" },
		"out is a file": func(c *MineConfig) { _ = os.WriteFile(c.OutDir, []byte("x"), 0o600) },
		"bad labels": func(c *MineConfig) {
			_ = os.MkdirAll(c.OutDir, 0o750)
			_ = os.WriteFile(filepath.Join(c.OutDir, LabelsFile), []byte("{bad\n"), 0o600)
		},
		"labels is a dir": func(c *MineConfig) {
			_ = os.MkdirAll(filepath.Join(c.OutDir, LabelsFile), 0o750)
		},
		"prompts is a dir": func(c *MineConfig) {
			_ = os.MkdirAll(filepath.Join(c.OutDir, PromptsFile), 0o750)
		},
	}
	for name, mutate := range cases {
		cfg, done := mineFixture(t)
		mutate(&cfg)
		if _, err := Mine(context.Background(), cfg); err == nil {
			t.Errorf("%s: want error", name)
		}
		done()
	}
}

func TestMineSessionQueryFails(t *testing.T) {
	cfg, done := mineFixture(t)
	defer done()
	srv := fakeLedger{
		"FROM projects":   rows(),
		"FROM tool_calls": func(int) any { return map[string]any{"error": "nope"} },
	}.server(t)
	defer srv.Close()
	cfg.Toolkit.BaseURL = srv.URL
	if _, err := Mine(context.Background(), cfg); err == nil {
		t.Error("want error")
	}
}

func TestBuildErrors(t *testing.T) {
	tax := mustTax(t)
	dir := t.TempDir()
	if _, err := Build(dir, tax, 0.8, true); err == nil {
		t.Error("missing prompts: want error")
	}
	if err := os.WriteFile(filepath.Join(dir, PromptsFile), []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(dir, tax, 0.8, true); err == nil {
		t.Error("bad prompts: want error")
	}
	if err := os.WriteFile(filepath.Join(dir, PromptsFile), []byte(`{"id":1}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, LabelsFile), []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(dir, tax, 0.8, true); err == nil {
		t.Error("bad labels: want error")
	}
}

func TestDirPasteLookup(t *testing.T) {
	if DirPasteLookup("") != nil {
		t.Error("empty dir disables the lookup")
	}
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "h.txt"), []byte("hi"), 0o600)
	look := DirPasteLookup(dir)
	if s, ok := look("h"); !ok || s != "hi" {
		t.Errorf("hit: %q %v", s, ok)
	}
	if _, ok := look("missing"); ok {
		t.Error("miss")
	}
	if _, ok := look("../h"); ok {
		t.Error("a hash with a path must not escape the cache dir")
	}
}

func TestMaxf(t *testing.T) {
	if maxf(1, 2) != 2 || maxf(3, 2) != 3 {
		t.Error("maxf")
	}
}
