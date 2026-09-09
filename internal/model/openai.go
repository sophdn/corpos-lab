package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAI is a chat-completions client for any OpenAI-compatible server —
// in this lab, llama-server at http://localhost:8081/v1. It implements
// Client behind an injectable *http.Client so tests run sans-IO against
// httptest servers.
type OpenAI struct {
	baseURL string
	modelID string
	version string
	httpc   *http.Client
	// completionTemplate, when non-empty, switches the client to raw /completion
	// (the least-opinionated path, no server-side chat template): the prompt is
	// substituted into this wrapper at its "{prompt}" placeholder and sent
	// verbatim to the server root's /completion. Empty means the OpenAI
	// chat-completions endpoint. See studies/REPRODUCIBILITY.md.
	completionTemplate string
}

// OpenAIOption configures an OpenAI client.
type OpenAIOption func(*OpenAI)

// WithHTTPClient injects the HTTP transport (tests pass the httptest
// server's client; production uses a timeout-configured client).
func WithHTTPClient(c *http.Client) OpenAIOption {
	return func(o *OpenAI) { o.httpc = c }
}

// WithCompletion switches the client to the raw /completion endpoint. Each
// prompt is substituted into template at its "{prompt}" placeholder and sent
// verbatim — the model sees exactly the study-declared wrapper and nothing a
// chat template would inject. The rendered string is returned on the Response so
// a run records the literal bytes the subject received. See
// studies/REPRODUCIBILITY.md.
func WithCompletion(template string) OpenAIOption {
	return func(o *OpenAI) { o.completionTemplate = template }
}

// NewOpenAI builds a client for the OpenAI-compatible server at baseURL
// (e.g. "http://localhost:8081/v1"). modelID is sent as the model field and
// reported by Name(); version is the artifact version reported by Version()
// for provenance stamping (llama-server does not report one).
func NewOpenAI(baseURL, modelID, version string, opts ...OpenAIOption) *OpenAI {
	o := &OpenAI{
		baseURL: baseURL,
		modelID: modelID,
		version: version,
		httpc:   http.DefaultClient,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Name reports the model identifier.
func (o *OpenAI) Name() string { return o.modelID }

// Version reports the configured artifact version.
func (o *OpenAI) Version() string { return o.version }

// chatRequest is llama-server's chat-completions body. Beyond the OpenAI
// fields it carries llama.cpp's sampler extensions (top_k, min_p, typical_p,
// repeat_*, xtc_*, dry_*, top_n_sigma) — verified honored by polling /slots
// mid-request: every field below comes back applied to the slot exactly as
// sent. omitempty + pointer means an unset param is dropped from the wire and
// silently inherits the server default, which is what GenParams documents.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Seed        *int          `json:"seed,omitempty"`

	TopK      *int     `json:"top_k,omitempty"`
	TopP      *float64 `json:"top_p,omitempty"`
	MinP      *float64 `json:"min_p,omitempty"`
	TypicalP  *float64 `json:"typical_p,omitempty"`
	TopNSigma *float64 `json:"top_n_sigma,omitempty"`

	RepeatPenalty    *float64 `json:"repeat_penalty,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	XTCProbability *float64 `json:"xtc_probability,omitempty"`
	DryMultiplier  *float64 `json:"dry_multiplier,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// ReasoningContent is llama.cpp's separated thinking, present when the
	// server was launched with a reasoning parser (e.g. --reasoning-format).
	// When absent, a thinking model instead emits the reasoning inline as a
	// leading <think>…</think> block in Content.
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	// Model and SystemFingerprint are the server's account of what answered.
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Timings           struct {
		PromptN            int     `json:"prompt_n"`
		PredictedN         int     `json:"predicted_n"`
		PromptPerSecond    float64 `json:"prompt_per_second"`
		PredictedPerSecond float64 `json:"predicted_per_second"`
	} `json:"timings"`
}

// propsResponse mirrors GET /props. The nested shape is llama.cpp's:
// default_generation_settings carries n_ctx and the launch-default params.
type propsResponse struct {
	ModelPath  string `json:"model_path"`
	ModelAlias string `json:"model_alias"`
	BuildInfo  string `json:"build_info"`
	DefaultGen struct {
		NCtx   int            `json:"n_ctx"`
		Params map[string]any `json:"params"`
	} `json:"default_generation_settings"`
}

// Generate produces a completion for prompt. It dispatches to the raw
// /completion endpoint when the client was built WithCompletion (the
// least-opinionated subject path), otherwise to chat-completions.
func (o *OpenAI) Generate(ctx context.Context, prompt string, params GenParams) (Response, error) {
	if o.completionTemplate != "" {
		return o.generateCompletion(ctx, prompt, params)
	}
	return o.generateChat(ctx, prompt, params)
}

// generateChat sends prompt as a single user message and returns the first
// choice's content.
func (o *OpenAI) generateChat(ctx context.Context, prompt string, params GenParams) (Response, error) {
	payload := chatRequest{
		Model:       o.modelID,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: params.Temperature,
		MaxTokens:   params.MaxTokens,
		Seed:        params.Seed,

		TopK:      params.TopK,
		TopP:      params.TopP,
		MinP:      params.MinP,
		TypicalP:  params.TypicalP,
		TopNSigma: params.TopNSigma,

		RepeatPenalty:    params.RepeatPenalty,
		RepeatLastN:      params.RepeatLastN,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,

		XTCProbability: params.XTCProbability,
		DryMultiplier:  params.DryMultiplier,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("model: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("model: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpc.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("model: %s: %w", o.modelID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("model: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Response{}, fmt.Errorf("model: %s returned %d: %s",
			o.modelID, resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Response{}, fmt.Errorf("model: decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Response{}, fmt.Errorf("model: %s returned no choices", o.modelID)
	}
	msg := parsed.Choices[0].Message
	text, inlineReasoning := splitThinking(msg.Content)
	reasoning := msg.ReasoningContent
	if reasoning == "" {
		reasoning = inlineReasoning
	}
	return Response{
		Text:              text,
		Reasoning:         reasoning,
		Model:             parsed.Model,
		SystemFingerprint: parsed.SystemFingerprint,
		Truncated:         parsed.Choices[0].FinishReason == "length",
		Timings: Timings{
			PromptN:            parsed.Timings.PromptN,
			PredictedN:         parsed.Timings.PredictedN,
			PromptPerSecond:    parsed.Timings.PromptPerSecond,
			PredictedPerSecond: parsed.Timings.PredictedPerSecond,
		},
	}, nil
}

// completionRequest is llama.cpp's raw /completion body. The full sampler chain
// rides here exactly as it does for chat; n_predict is /completion's name for
// the token cap. cache_prompt is pinned false so a run is not shaped by a warm
// prompt cache left by a prior run.
type completionRequest struct {
	Prompt      string   `json:"prompt"`
	NPredict    *int     `json:"n_predict,omitempty"`
	Seed        *int     `json:"seed,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`

	TopK      *int     `json:"top_k,omitempty"`
	TopP      *float64 `json:"top_p,omitempty"`
	MinP      *float64 `json:"min_p,omitempty"`
	TypicalP  *float64 `json:"typical_p,omitempty"`
	TopNSigma *float64 `json:"top_n_sigma,omitempty"`

	RepeatPenalty    *float64 `json:"repeat_penalty,omitempty"`
	RepeatLastN      *int     `json:"repeat_last_n,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	XTCProbability *float64 `json:"xtc_probability,omitempty"`
	DryMultiplier  *float64 `json:"dry_multiplier,omitempty"`

	CachePrompt bool `json:"cache_prompt"`
}

// completionResponse is llama.cpp's raw /completion reply. It carries the same
// timings block as chat, but no system_fingerprint (build id) — that is read
// separately from /props and recorded per run, so its absence here is a small,
// honest gap, not a failure.
type completionResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	// StoppedLimit is llama.cpp's flag that generation stopped at n_predict (the
	// token cap) rather than an EOS or stop word — i.e. the reply was truncated.
	StoppedLimit bool `json:"stopped_limit"`
	Timings      struct {
		PromptN            int     `json:"prompt_n"`
		PredictedN         int     `json:"predicted_n"`
		PromptPerSecond    float64 `json:"prompt_per_second"`
		PredictedPerSecond float64 `json:"predicted_per_second"`
	} `json:"timings"`
}

// generateCompletion sends the study-declared wrapper (prompt substituted) to
// the raw /completion endpoint and returns the reply. The rendered wrapper is
// returned on the Response as the literal input the subject received.
func (o *OpenAI) generateCompletion(ctx context.Context, prompt string, params GenParams) (Response, error) {
	rendered := strings.Replace(o.completionTemplate, "{prompt}", prompt, 1)
	payload := completionRequest{
		Prompt:      rendered,
		NPredict:    params.MaxTokens,
		Seed:        params.Seed,
		Temperature: params.Temperature,

		TopK:      params.TopK,
		TopP:      params.TopP,
		MinP:      params.MinP,
		TypicalP:  params.TypicalP,
		TopNSigma: params.TopNSigma,

		RepeatPenalty:    params.RepeatPenalty,
		RepeatLastN:      params.RepeatLastN,
		PresencePenalty:  params.PresencePenalty,
		FrequencyPenalty: params.FrequencyPenalty,

		XTCProbability: params.XTCProbability,
		DryMultiplier:  params.DryMultiplier,

		CachePrompt: false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("model: marshal completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.completionURL(), bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("model: build completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpc.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("model: %s: %w", o.modelID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("model: read completion response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Response{}, fmt.Errorf("model: %s completion returned %d: %s",
			o.modelID, resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed completionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Response{}, fmt.Errorf("model: decode completion response: %w", err)
	}
	text, inlineReasoning := splitThinking(parsed.Content)
	return Response{
		RenderedPrompt:    rendered,
		Text:              text,
		Reasoning:         inlineReasoning,
		Model:             parsed.Model,
		SystemFingerprint: "",
		Truncated:         parsed.StoppedLimit,
		Timings: Timings{
			PromptN:            parsed.Timings.PromptN,
			PredictedN:         parsed.Timings.PredictedN,
			PromptPerSecond:    parsed.Timings.PromptPerSecond,
			PredictedPerSecond: parsed.Timings.PredictedPerSecond,
		},
	}, nil
}

// completionURL derives the /completion endpoint from baseURL. Like /props,
// /completion is served at the ROOT, not under /v1, so this trims the OpenAI
// path segment rather than appending to it.
func (o *OpenAI) completionURL() string {
	root := strings.TrimSuffix(o.baseURL, "/")
	root = strings.TrimSuffix(root, "/v1")
	return strings.TrimSuffix(root, "/") + "/completion"
}

// splitThinking separates a leading inline <think>…</think> block from the
// answer. llama.cpp emits reasoning inline in the content when no reasoning
// parser is configured; a caller that parses the answer (e.g. a verdict step
// keyed on a leading PASS/FAIL) must see the content past the think block. It
// returns the trimmed answer and the reasoning (without the tags). Content
// with no think block is returned unchanged with empty reasoning. An unclosed
// <think> (a truncated generation) yields an empty answer and the partial
// reasoning, so the caller fails the parse rather than reading think text as
// the answer.
func splitThinking(content string) (answer, reasoning string) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "<think>") {
		return content, ""
	}
	rest := trimmed[len("<think>"):]
	end := strings.Index(rest, "</think>")
	if end < 0 {
		return "", strings.TrimSpace(rest)
	}
	reasoning = strings.TrimSpace(rest[:end])
	answer = strings.TrimSpace(rest[end+len("</think>"):])
	return answer, reasoning
}

// propsURL derives the /props endpoint from baseURL.
//
// /props is served at the ROOT, not under /v1 — llama-server answers /props
// with 200 and /v1/props with 404 — so this trims the OpenAI path segment
// rather than appending to it. baseURL is conventionally
// "http://llama-server:8081/v1"; a baseURL already at the root is left alone.
func (o *OpenAI) propsURL() string {
	root := strings.TrimSuffix(o.baseURL, "/")
	root = strings.TrimSuffix(root, "/v1")
	return strings.TrimSuffix(root, "/") + "/props"
}

// Props reads the server's self-report: which artifact it loaded, which build
// is serving it, the context window, and the sampler defaults an unpinned
// param would inherit.
//
// This does NOT report the effective sampler for any request — see ServerProps.
// Its load-bearing use is model identity: the study declares a model_id, and
// until now nothing ever checked it against what the server was actually
// serving.
func (o *OpenAI) Props(ctx context.Context) (ServerProps, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.propsURL(), nil)
	if err != nil {
		return ServerProps{}, fmt.Errorf("model: build props request: %w", err)
	}

	resp, err := o.httpc.Do(req)
	if err != nil {
		return ServerProps{}, fmt.Errorf("model: props: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return ServerProps{}, fmt.Errorf("model: read props response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ServerProps{}, fmt.Errorf("model: props returned %d: %s",
			resp.StatusCode, truncate(string(raw), 200))
	}

	var parsed propsResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ServerProps{}, fmt.Errorf("model: decode props response: %w", err)
	}
	return ServerProps{
		ModelPath:  parsed.ModelPath,
		ModelAlias: parsed.ModelAlias,
		BuildInfo:  parsed.BuildInfo,
		NCtx:       parsed.DefaultGen.NCtx,
		Defaults:   parsed.DefaultGen.Params,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
