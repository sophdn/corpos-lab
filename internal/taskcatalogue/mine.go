package taskcatalogue

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"corpos-lab/internal/model"
)

// Output file names inside the catalogue directory.
const (
	PromptsFile = "prompts.jsonl"
	LabelsFile  = "labels.jsonl"
	KindsFile   = "kinds.json"
	RunsDir     = "runs"
)

// MineConfig is everything one labelling run needs. The Toolkit and Client
// are the run's only reach into the world.
type MineConfig struct {
	HistoryPath string
	// PasteDir is the harness paste cache (<hash>.txt files); empty skips it.
	PasteDir     string
	TaxonomyPath string
	OutDir       string
	Toolkit      Toolkit
	Client       model.Client
	Params       model.GenParams
	Workers      int
	// Limit caps how many pending prompts this run labels (0 = all), for a
	// one-cell smoke before the full run.
	Limit       int
	RetryFailed bool
	// Commit identifies the code that ran; CommitErr records why it could not
	// be read. Neither stops a run.
	Commit    any
	CommitErr string
	Log       io.Writer
	Now       func() time.Time
}

// RunRecord is what a labelling run writes about itself, never edited after.
type RunRecord struct {
	Started      time.Time         `json:"started"`
	Finished     time.Time         `json:"finished"`
	ModelAsked   string            `json:"model_asked"`
	Props        model.ServerProps `json:"props"`
	PropsErr     string            `json:"props_err,omitempty"`
	Sampler      model.GenParams   `json:"sampler"`
	Commit       any               `json:"commit,omitempty"`
	CommitErr    string            `json:"commit_err,omitempty"`
	Prompts      int               `json:"prompts"`
	LexiconTerms int               `json:"lexicon_terms"`
	Sessions     int               `json:"sessions_with_toolkit_scope"`
	Pending      int               `json:"pending"`
	Written      int               `json:"written"`
	Workers      int               `json:"workers"`
	RetryFailed  bool              `json:"retry_failed"`
	Interrupted  string            `json:"interrupted,omitempty"`
}

// Mine parses the history, attributes every prompt to a project, rewrites
// prompts.jsonl, labels the prompts that still need it (appending to
// labels.jsonl), and writes a run record under runs/.
func Mine(ctx context.Context, cfg MineConfig) (RunRecord, error) {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	logf := func(format string, args ...any) {
		if cfg.Log != nil {
			_, _ = fmt.Fprintf(cfg.Log, format+"\n", args...)
		}
	}
	rec := RunRecord{Started: now().UTC(), ModelAsked: cfg.Client.Name(), Sampler: cfg.Params,
		Commit: cfg.Commit, CommitErr: cfg.CommitErr, Workers: cfg.Workers, RetryFailed: cfg.RetryFailed}

	prompts, err := readHistory(cfg.HistoryPath, cfg.PasteDir)
	if err != nil {
		return rec, err
	}
	rec.Prompts = len(prompts)
	taxSrc, err := os.ReadFile(cfg.TaxonomyPath)
	if err != nil {
		return rec, fmt.Errorf("read taxonomy: %w", err)
	}
	tax, err := ParseTaxonomy(string(taxSrc))
	if err != nil {
		return rec, err
	}
	terms, err := cfg.Toolkit.LedgerTerms(ctx)
	if err != nil {
		return rec, err
	}
	lex := NewLexicon(append(terms, tax.Aliases...), tax.Ignore...)
	rec.LexiconTerms = lex.Len()
	sessions, err := cfg.Toolkit.SessionProjects(ctx)
	if err != nil {
		return rec, err
	}
	rec.Sessions = len(sessions)
	attrs := Attribute(prompts, lex, sessions)
	if err := os.MkdirAll(filepath.Join(cfg.OutDir, RunsDir), 0o750); err != nil {
		return rec, fmt.Errorf("make out dir: %w", err)
	}
	if err := writePrompts(filepath.Join(cfg.OutDir, PromptsFile), prompts, attrs); err != nil {
		return rec, err
	}
	logf("task-catalogue: %d prompts, %d lexicon terms, %d sessions with a toolkit scope", len(prompts), lex.Len(), len(sessions))

	labelsPath := filepath.Join(cfg.OutDir, LabelsFile)
	done, err := readLabelsFile(labelsPath)
	if err != nil {
		return rec, err
	}
	pending := Pending(prompts, done, cfg.RetryFailed)
	if cfg.Limit > 0 && len(pending) > cfg.Limit {
		pending = pending[:cfg.Limit]
	}
	rec.Pending = len(pending)
	if props, perr := cfg.Client.Props(ctx); perr != nil {
		rec.PropsErr = perr.Error()
	} else {
		rec.Props = props
	}
	logf("task-catalogue: labelling %d prompt(s) with %d worker(s)", len(pending), cfg.Workers)

	f, err := os.OpenFile(labelsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return rec, fmt.Errorf("open labels: %w", err)
	}
	cls := Classifier{Client: cfg.Client, Tax: tax, Params: cfg.Params}
	started := now()
	written, runErr := LabelAll(ctx, cls, pending, cfg.Workers, f, func(n int) {
		if n%100 == 0 || n == len(pending) {
			el := now().Sub(started).Seconds()
			logf("task-catalogue: %d/%d labelled (%.1f/s)", n, len(pending), float64(n)/maxf(el, 1e-9))
		}
	})
	closeErr := f.Close()
	rec.Written = written
	rec.Finished = now().UTC()
	if runErr != nil {
		rec.Interrupted = runErr.Error()
	}
	if err := writeRunRecord(cfg.OutDir, rec); err != nil {
		return rec, err
	}
	if closeErr != nil {
		return rec, fmt.Errorf("close labels: %w", closeErr)
	}
	return rec, runErr
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func readHistory(path, pasteDir string) ([]Prompt, error) {
	f, err := os.Open(path) //nolint:gosec // the caller names the history file
	if err != nil {
		return nil, fmt.Errorf("open history: %w", err)
	}
	defer func() { _ = f.Close() }()
	return ParseHistory(f, DirPasteLookup(pasteDir))
}

// DirPasteLookup reads pastes from a cache directory of <hash>.txt files. An
// empty dir disables the lookup.
func DirPasteLookup(dir string) PasteLookup {
	if dir == "" {
		return nil
	}
	return func(hash string) (string, bool) {
		if hash != filepath.Base(hash) {
			return "", false
		}
		raw, err := os.ReadFile(filepath.Join(dir, hash+".txt")) //nolint:gosec // base name only, inside dir
		if err != nil {
			return "", false
		}
		return string(raw), true
	}
}

func readLabelsFile(path string) (map[int]PromptLabel, error) {
	f, err := os.Open(path) //nolint:gosec // path is inside the caller's out dir
	if os.IsNotExist(err) {
		return map[int]PromptLabel{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open labels: %w", err)
	}
	defer func() { _ = f.Close() }()
	return ReadLabels(f)
}

func writePrompts(path string, prompts []Prompt, attrs map[int]Attribution) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("write prompts: %w", err)
	}
	enc := json.NewEncoder(f)
	for _, p := range prompts {
		if err := enc.Encode(PromptRecord{Prompt: p, Attribution: attrs[p.ID]}); err != nil {
			_ = f.Close()
			return fmt.Errorf("write prompts: %w", err)
		}
	}
	return f.Close()
}

func writeRunRecord(outDir string, rec RunRecord) error {
	raw, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run record: %w", err)
	}
	name := rec.Started.Format("20060102T150405Z") + ".json"
	return os.WriteFile(filepath.Join(outDir, RunsDir, name), append(raw, '\n'), 0o600)
}

// ReadPromptRecords reads prompts.jsonl.
func ReadPromptRecords(r io.Reader) ([]PromptRecord, error) {
	dec := json.NewDecoder(r)
	var out []PromptRecord
	for {
		var p PromptRecord
		err := dec.Decode(&p)
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("prompts record %d: %w", len(out)+1, err)
		}
		out = append(out, p)
	}
}

// Build reads a catalogue directory and aggregates it. With partial unset, an
// unlabelled prompt is an error: the catalogue claims to cover the population.
func Build(outDir string, tax Taxonomy, threshold float64, partial bool) (Summary, error) {
	pf, err := os.Open(filepath.Join(outDir, PromptsFile)) //nolint:gosec // inside the caller's out dir
	if err != nil {
		return Summary{}, fmt.Errorf("open prompts: %w", err)
	}
	prompts, err := ReadPromptRecords(pf)
	_ = pf.Close()
	if err != nil {
		return Summary{}, err
	}
	labels, err := readLabelsFile(filepath.Join(outDir, LabelsFile))
	if err != nil {
		return Summary{}, err
	}
	s, err := Aggregate(prompts, labels, tax, threshold)
	if err != nil {
		return Summary{}, err
	}
	if !partial && s.Unlabelled > 0 {
		return s, fmt.Errorf("%d prompt(s) have no label; finish the labelling run or pass -partial", s.Unlabelled)
	}
	return s, nil
}

// WriteTable prints the kind table for a human: one line per opener kind,
// then the role totals.
func WriteTable(w io.Writer, s Summary) error {
	ew := &errWriter{w: w}
	ew.printf("prompts %d  openers %d  invalid %d  errored %d  unlabelled %d  catalogue covers %.1f%% (threshold %.0f%%)\n\n",
		s.Total, s.Openers, s.Invalid, s.Errored, s.Unlabelled, 100*s.Covered, 100*s.Threshold)
	ew.printf("%-3s %-28s %6s %7s %7s  %s\n", "cat", "kind", "count", "share", "cum", "top projects")
	for _, r := range s.Kinds {
		mark := ""
		if r.Catalogue {
			mark = "*"
		}
		ew.printf("%-3s %-28s %6d %6.1f%% %6.1f%%  %s\n", mark, r.Kind, r.Count, 100*r.Share, 100*r.Cumulative, topProjects(r.Projects, 3))
	}
	roles := make([]string, 0, len(s.ByRole))
	for r := range s.ByRole {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	ew.printf("\nroles:")
	for _, r := range roles {
		ew.printf(" %s=%d", r, s.ByRole[r])
	}
	ew.printf("\n")
	return ew.err
}

func topProjects(ps []ProjectCount, n int) string {
	out := ""
	for i, p := range ps {
		if i == n {
			break
		}
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s %d", p.Project, p.Count)
	}
	return out
}

type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, args ...any) {
	if e.err == nil {
		_, e.err = fmt.Fprintf(e.w, format, args...)
	}
}
