// Package model defines the inference-client seam battery steps (and later
// assay runners) call through. Production implementation: an
// OpenAI-compatible chat-completions client for llama-server. Tests inject
// fakes — the seam is what makes inference steps sans-IO testable.
package model

import "context"

// GenParams are per-call generation parameters.
//
// It carries EVERY stage of llama.cpp's sampler chain, not just temperature.
// The chain is, in order (server-reported at /props as `samplers`):
//
//	penalties → dry → top_n_sigma → top_k → typ_p → top_p → min_p → xtc → temperature
//
// A param left nil is omitted from the wire and the server applies its own
// default. That is the hole this struct exists to close: temperature alone does
// not define a sampling distribution, so a study that declared only
// "temperature 0.8, seed N" had specified a fraction of its sampler and
// inherited the rest from whatever binary happened to be running. Upgrade
// llama.cpp, have it move a default, and the same study samples differently
// with nothing in the record to show it.
//
// Callers that care about knowing what they ran set every field; study-driven
// runs are validated for exactly that (see internal/study).
type GenParams struct {
	// Temperature is the final softmax scale. 0 is greedy decoding.
	Temperature *float64
	// MaxTokens caps the completion length.
	MaxTokens *int
	// Seed pins the sampler's RNG for this call. It is what makes a
	// temperature>0 run reproducible: without it, sampled runs cannot be
	// re-executed, and with temperature 0 it is irrelevant because decoding
	// is greedy. Nil leaves the server to pick.
	Seed *int

	// Truncation stages. Each is disabled at its neutral value: TopK 0,
	// TopP 1.0, MinP 0.0, TypicalP 1.0, TopNSigma -1.
	TopK      *int
	TopP      *float64
	MinP      *float64
	TypicalP  *float64
	TopNSigma *float64

	// Penalty stages, all neutral at the values noted: RepeatPenalty 1.0,
	// RepeatLastN 0, PresencePenalty 0, FrequencyPenalty 0.
	RepeatPenalty    *float64
	RepeatLastN      *int
	PresencePenalty  *float64
	FrequencyPenalty *float64

	// Disabling knobs for the remaining stages. Their sub-parameters
	// (xtc_threshold, dry_base, dry_allowed_length, …) are inert while these
	// are zero, so pinning the switch pins the stage.
	XTCProbability *float64
	DryMultiplier  *float64
}

// Timings is the server's own report of what a single generation cost. It is
// measured, not declared — and it is the in-band tripwire for the substrate.
// llama-server on GPU runs this model at ~40-60 tok/s; the same model on CPU
// ran at 5.1. A study that silently fell back to CPU is unmissable here, which
// is precisely what went unnoticed for a full study in July 2026.
type Timings struct {
	PromptN            int     `json:"prompt_n"`
	PredictedN         int     `json:"predicted_n"`
	PromptPerSecond    float64 `json:"prompt_per_second"`
	PredictedPerSecond float64 `json:"predicted_per_second"`
}

// Response is a model's reply to a single prompt, plus what the server said
// about itself while answering.
type Response struct {
	// RenderedPrompt is the exact prompt string the client put on the wire — for
	// raw /completion, the study-declared wrapper with the material substituted,
	// i.e. the literal bytes the subject received. Empty for the chat endpoint,
	// where the server applies its own template and the true input is not ours to
	// record. This is what makes a vanilla-path run self-describing down to its
	// input (reproducibility contract: studies/REPRODUCIBILITY.md).
	RenderedPrompt string
	// Text is the completion's answer content with any thinking removed: the
	// reasoning stripped into Reasoning below, and a leading inline
	// <think>…</think> block (llama.cpp emits this when no reasoning parser is
	// configured) trimmed off. This is what a caller parses — a verdict step
	// that expects a leading "PASS"/"FAIL" must not trip over a think block.
	Text string
	// Reasoning is the thinking a reasoning model produced, when the server
	// surfaced it — either as the response's reasoning_content field or as the
	// inline <think>…</think> block trimmed from Text. Empty for a non-thinking
	// model. Recorded as provenance; never parsed for the answer.
	Reasoning string
	// Model is the model the SERVER reports having used — not the one we
	// asked for. A mismatch against the requested model is the check that
	// catches a model swap behind the endpoint.
	Model string
	// SystemFingerprint is llama.cpp's build id (e.g. "b9445-af6528e6d").
	SystemFingerprint string
	// Truncated is true when generation stopped because it hit the token cap
	// (n_predict / max_tokens) rather than a natural stop. Recorded per run so a
	// response cut off mid-artifact is visible in the data instead of being
	// inferred by a scorer: on verbose classes a correct answer truncated before
	// its load-bearing step is otherwise indistinguishable from an incomplete one.
	// Sourced from /completion's stopped_limit and chat's finish_reason=="length".
	Truncated bool
	// Timings is the server's cost report for this generation.
	Timings Timings
}

// ServerProps is what the inference server reports about itself at GET /props.
//
// Note what is NOT here: the effective sampler. /props reports the server's
// LAUNCH defaults (default_generation_settings.params), not what any given
// request ran under — send temperature 0.0 to a server started at 0.8 and
// /props still says 0.8. Reading it back as "the config we ran" would
// manufacture a confident false record, which is the exact failure the
// record-what-ran invariant exists to prevent. Defaults is kept for what it
// honestly is: a cross-check on what an OMITTED param would have inherited.
type ServerProps struct {
	// ModelPath is the artifact the server actually loaded.
	ModelPath string `json:"model_path"`
	// ModelAlias is the served model's name.
	ModelAlias string `json:"model_alias"`
	// BuildInfo is the llama.cpp build id.
	BuildInfo string `json:"build_info"`
	// NCtx is the context window the server was launched with.
	NCtx int `json:"n_ctx"`
	// Defaults is default_generation_settings.params verbatim — the sampler
	// an unpinned param would silently inherit.
	Defaults map[string]any `json:"defaults,omitempty"`
}

// Client is the inference seam. Name and Version identify the model as the
// study DECLARED it; Props reports what the server says it is actually
// serving. The two are recorded side by side and never reconciled silently —
// a divergence is information about the run, not grounds for refusing it.
type Client interface {
	// Generate produces a completion for prompt under params.
	Generate(ctx context.Context, prompt string, params GenParams) (Response, error)
	// Name identifies the model (e.g. "qwen2.5-32b-instruct").
	Name() string
	// Version identifies the model artifact version for provenance.
	Version() string
	// Props reads the server's self-report. Callers record it with the
	// results; they do not gate on it.
	Props(ctx context.Context) (ServerProps, error)
}

// Float64 returns a pointer to v, for GenParams literals.
func Float64(v float64) *float64 { return &v }

// Int returns a pointer to v, for GenParams literals.
func Int(v int) *int { return &v }
