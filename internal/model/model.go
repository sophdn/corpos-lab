// Package model defines the inference-client seam battery steps (and later
// assay runners) call through. Production implementation: an
// OpenAI-compatible chat-completions client for llama-server. Tests inject
// fakes — the seam is what makes inference steps sans-IO testable.
package model

import "context"

// GenParams are per-call generation parameters. Nil fields mean "use the
// server's default"; battery verdict steps pin temperature 0.0 / 256 tokens.
type GenParams struct {
	Temperature *float64
	MaxTokens   *int
	// Seed pins the sampler's RNG for this call. It is what makes a
	// temperature>0 run reproducible: without it, sampled runs cannot be
	// re-executed, and with temperature 0 it is irrelevant because decoding
	// is greedy. Nil leaves the server to pick.
	Seed *int
}

// Response is a model's reply to a single prompt.
type Response struct {
	// Text is the raw completion text.
	Text string
}

// Client is the inference seam. Name and Version identify the model for
// provenance stamping (charter freeze-by-digest: sampling params and model
// identity are part of every persisted result).
type Client interface {
	// Generate produces a completion for prompt under params.
	Generate(ctx context.Context, prompt string, params GenParams) (Response, error)
	// Name identifies the model (e.g. "qwen2.5-32b-instruct").
	Name() string
	// Version identifies the model artifact version for provenance.
	Version() string
}

// Float64 returns a pointer to v, for GenParams literals.
func Float64(v float64) *float64 { return &v }

// Int returns a pointer to v, for GenParams literals.
func Int(v int) *int { return &v }
