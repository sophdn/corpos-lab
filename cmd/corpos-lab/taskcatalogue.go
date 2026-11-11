package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"corpos-lab/internal/model"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/taskcatalogue"
)

// runTaskCatalogue builds the catalogue of task kinds a principal asks a coding
// agent for, from the harness prompt log (corpos-lab chain
// popular-task-catalogue). "mine" attributes and labels the prompts; it is
// resumable, so an interrupted run continues where it stopped. "aggregate"
// turns the labels into the kind table.
func runTaskCatalogue(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "task-catalogue: want mine or aggregate")
		return 2
	}
	switch args[0] {
	case "mine":
		return runTaskCatalogueMine(args[1:])
	case "aggregate":
		return runTaskCatalogueAggregate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "task-catalogue: unknown verb %q (want mine or aggregate)\n", args[0])
		return 2
	}
}

func runTaskCatalogueMine(args []string) int {
	home, _ := os.UserHomeDir()
	fs := flag.NewFlagSet("task-catalogue mine", flag.ContinueOnError)
	history := fs.String("history", filepath.Join(home, ".claude", "history.jsonl"), "harness prompt log")
	pastes := fs.String("paste-cache", filepath.Join(home, ".claude", "paste-cache"), "harness paste cache (<hash>.txt); empty to skip")
	taxonomy := fs.String("taxonomy", "", "taxonomy TOML (roles, kinds, extra project aliases)")
	out := fs.String("out", "", "catalogue directory (private; never a tracked path)")
	toolkitURL := fs.String("toolkit-url", "http://localhost", "the toolkit service base URL")
	modelURL := fs.String("model-url", "http://localhost:8081/v1", "inference portal base URL")
	modelName := fs.String("model", "Qwen3.8-27B-Q4_K_M.gguf", "model to request from the portal")
	workers := fs.Int("workers", 4, "concurrent labelling calls (the portal serves 4 slots)")
	limit := fs.Int("limit", 0, "label at most this many pending prompts (0 = all); use a small value to smoke")
	retry := fs.Bool("retry-failed", false, "re-label prompts whose last label is invalid or errored")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *taxonomy == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "task-catalogue mine: -taxonomy and -out are required")
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := taskcatalogue.MineConfig{
		HistoryPath:  *history,
		PasteDir:     *pastes,
		TaxonomyPath: *taxonomy,
		OutDir:       *out,
		Toolkit:      taskcatalogue.Toolkit{BaseURL: *toolkitURL, Project: "corpos-lab", HTTP: &http.Client{Timeout: 30 * time.Second}},
		Client:       model.NewOpenAI(*modelURL, *modelName, *modelName, model.WithHTTPClient(&http.Client{Timeout: 5 * time.Minute})),
		Params:       taskcatalogue.LabelParams(),
		Workers:      *workers,
		Limit:        *limit,
		RetryFailed:  *retry,
		Log:          os.Stderr,
	}
	if wd, err := os.Getwd(); err == nil {
		if stamp, serr := provenance.Capture(ctx, wd); serr != nil {
			cfg.CommitErr = serr.Error()
		} else {
			cfg.Commit = stamp
		}
	}
	rec, err := taskcatalogue.Mine(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue mine: %v (wrote %d label(s); re-run to continue)\n", err, rec.Written)
		return 1
	}
	fmt.Printf("task-catalogue mine: wrote %d label(s) for %d pending prompt(s)\n", rec.Written, rec.Pending)
	return 0
}

func runTaskCatalogueAggregate(args []string) int {
	fs := flag.NewFlagSet("task-catalogue aggregate", flag.ContinueOnError)
	taxonomy := fs.String("taxonomy", "", "taxonomy TOML")
	out := fs.String("out", "", "catalogue directory")
	threshold := fs.Float64("threshold", 0.8, "opener share the catalogue set must cover")
	partial := fs.Bool("partial", false, "allow unlabelled prompts (a preview, not the catalogue)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *taxonomy == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "task-catalogue aggregate: -taxonomy and -out are required")
		return 2
	}
	src, err := os.ReadFile(*taxonomy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue aggregate: %v\n", err)
		return 1
	}
	tax, err := taskcatalogue.ParseTaxonomy(string(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue aggregate: %v\n", err)
		return 1
	}
	s, err := taskcatalogue.Build(*out, tax, *threshold, *partial)
	if err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue aggregate: %v\n", err)
		return 1
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue aggregate: %v\n", err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(*out, taskcatalogue.KindsFile), append(raw, '\n'), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "task-catalogue aggregate: %v\n", err)
		return 1
	}
	if err := taskcatalogue.WriteTable(os.Stdout, s); err != nil {
		return 1
	}
	return 0
}
