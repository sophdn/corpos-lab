package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"corpos-lab/internal/model"
	"corpos-lab/internal/rater"
)

// localChatBase is the one local inference portal (llama-server). A base URL
// that differs from it, or an api-key-env, marks a hosted endpoint.
const localChatBase = "http://localhost:8081/v1"

// runRate scores blind grid slices with one model and one rubric kind. It is the
// thin wiring over internal/rater: it builds the chat client (local or a hosted
// endpoint with a bearer token), builds the grounded or action rater, discovers
// the slices, and runs them resumably with an atomic, create-only promotion so a
// late straggler cannot overwrite a fresh result.
func runRate(args []string) int {
	fs := flag.NewFlagSet("rate", flag.ContinueOnError)
	slicesDir := fs.String("slices-dir", "", "directory of *.jsonl slices")
	slice := fs.String("slice", "", "a single slice JSONL file")
	outDir := fs.String("out-dir", "", "output root; results land under <out-dir>/<rater-id>/")
	raterID := fs.String("rater-id", "", "rater name (namespaces the output dir)")
	rubricPath := fs.String("rubric", "", "class rubric .md for the grounded rater")
	action := fs.Bool("action", false, "score the committed-action verdict instead of the grounded codes")
	base := fs.String("base", localChatBase, "OpenAI-compatible base URL")
	modelID := fs.String("model", "", "model id to request (empty for local)")
	apiKeyEnv := fs.String("api-key-env", "", "env var holding a bearer token for a hosted endpoint")
	maxTokens := fs.Int("max-tokens", -1, "per-rating token budget; <0 uses the default (16 local, 2048 hosted)")
	jobs := fs.Int("jobs", 1, "worker pool size (default 1, sequential)")
	prov := fs.Bool("prov", false, "write a provenance sidecar next to each result")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *outDir == "" || *raterID == "" {
		fmt.Fprintln(os.Stderr, "corpos-lab: rate needs --out-dir and --rater-id")
		return 2
	}
	if (*slicesDir == "") == (*slice == "") {
		fmt.Fprintln(os.Stderr, "corpos-lab: rate needs exactly one of --slices-dir or --slice")
		return 2
	}
	if !*action && *rubricPath == "" {
		fmt.Fprintln(os.Stderr, "corpos-lab: rate needs --rubric (or --action)")
		return 2
	}

	hosted := *apiKeyEnv != "" || *base != localChatBase
	token := ""
	if *apiKeyEnv != "" {
		token = os.Getenv(*apiKeyEnv)
		if token == "" {
			fmt.Fprintf(os.Stderr, "corpos-lab: rate: env var %s is empty\n", *apiKeyEnv)
			return 2
		}
	}
	var budget *int
	if *maxTokens >= 0 {
		budget = maxTokens
	}

	client := model.NewOpenAI(*base, *modelID, "",
		model.WithHTTPClient(&http.Client{Timeout: 3 * time.Minute}),
		model.WithBearerToken(token))

	cfg := rater.Config{
		Seam:           client,
		RaterID:        *raterID,
		Endpoint:       *base,
		RequestedModel: *modelID,
		MaxTokens:      budget,
		Hosted:         hosted,
	}
	var r *rater.Rater
	if *action {
		r = rater.NewAction(cfg)
	} else {
		rubric, err := os.ReadFile(*rubricPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "corpos-lab: rate: read rubric: %v\n", err)
			return 1
		}
		r = rater.NewGrounded(cfg, string(rubric))
	}

	slices, err := rater.DiscoverSlices(*slicesDir, *slice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: rate: %v\n", err)
		return 1
	}
	if len(slices) == 0 {
		fmt.Fprintln(os.Stderr, "corpos-lab: rate: no slices found")
		return 2
	}

	sum, err := rater.RunSlices(context.Background(), rater.RunConfig{
		Rater: r, OutDir: *outDir, Jobs: *jobs, WriteProvenance: *prov,
	}, slices)
	if err != nil {
		fmt.Fprintf(os.Stderr, "corpos-lab: rate: %v\n", err)
		return 1
	}
	for _, o := range sum.Outcomes {
		fmt.Fprintf(os.Stderr, "  %-14s %s\n", o.Status, o.Slice)
	}
	fmt.Fprintf(os.Stderr, "rate: scored %d, %d already complete, %d kept-existing (rater %s)\n",
		sum.Scored, sum.AlreadyComplete, sum.KeptExisting, *raterID)
	return 0
}
