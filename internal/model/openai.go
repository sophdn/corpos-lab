package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
}

// OpenAIOption configures an OpenAI client.
type OpenAIOption func(*OpenAI)

// WithHTTPClient injects the HTTP transport (tests pass the httptest
// server's client; production uses a timeout-configured client).
func WithHTTPClient(c *http.Client) OpenAIOption {
	return func(o *OpenAI) { o.httpc = c }
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

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Seed        *int          `json:"seed,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Generate sends prompt as a single user message and returns the first
// choice's content.
func (o *OpenAI) Generate(ctx context.Context, prompt string, params GenParams) (Response, error) {
	payload := chatRequest{
		Model:       o.modelID,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: params.Temperature,
		MaxTokens:   params.MaxTokens,
		Seed:        params.Seed,
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
	return Response{Text: parsed.Choices[0].Message.Content}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
