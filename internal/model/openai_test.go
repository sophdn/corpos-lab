package model

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIGenerateSendsChatCompletionShape(t *testing.T) {
	var got chatRequest
	var gotPath, gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"PASS"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "qwen2.5-32b", "q4km-2026", WithHTTPClient(srv.Client()))
	resp, err := c.Generate(context.Background(), "evaluate this", GenParams{
		Temperature: Float64(0.0),
		MaxTokens:   Int(256),
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Text != "PASS" {
		t.Fatalf("Text = %q, want PASS", resp.Text)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}
	if got.Model != "qwen2.5-32b" {
		t.Fatalf("model = %q", got.Model)
	}
	if len(got.Messages) != 1 || got.Messages[0].Role != "user" || got.Messages[0].Content != "evaluate this" {
		t.Fatalf("messages = %+v", got.Messages)
	}
	if got.Temperature == nil || *got.Temperature != 0.0 {
		t.Fatalf("temperature = %v", got.Temperature)
	}
	if got.MaxTokens == nil || *got.MaxTokens != 256 {
		t.Fatalf("max_tokens = %v", got.MaxTokens)
	}
	// No seed was asked for, so none is sent — omitempty leaves the server's
	// default alone rather than silently pinning seed 0.
	if got.Seed != nil {
		t.Fatalf("seed should be absent when unset, got %v", *got.Seed)
	}
	if c.Name() != "qwen2.5-32b" || c.Version() != "q4km-2026" {
		t.Fatal("identity accessors")
	}
}

func TestOpenAIGenerateOmitsNilParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := json.Marshal(map[string]any{})
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		if _, present := body["temperature"]; present {
			t.Error("temperature should be omitted when nil")
		}
		if _, present := body["max_tokens"]; present {
			t.Error("max_tokens should be omitted when nil")
		}
		_ = raw
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL, "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.Generate(context.Background(), "p", GenParams{}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
}

func TestOpenAIGenerateNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("model exploded"))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL, "m", "", WithHTTPClient(srv.Client()))
	_, err := c.Generate(context.Background(), "p", GenParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "500") || !strings.Contains(err.Error(), "model exploded") {
		t.Fatalf("error should carry status and body: %v", err)
	}
}

func TestOpenAIGenerateGarbageJSONIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL, "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.Generate(context.Background(), "p", GenParams{}); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestOpenAIGenerateEmptyChoicesIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL, "m", "", WithHTTPClient(srv.Client()))
	_, err := c.Generate(context.Background(), "p", GenParams{})
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("expected no-choices error, got %v", err)
	}
}

func TestOpenAIGenerateConnectionErrorIsWrapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // immediately closed: connection refused

	c := NewOpenAI(srv.URL, "downed-model", "")
	_, err := c.Generate(context.Background(), "p", GenParams{})
	if err == nil || !strings.Contains(err.Error(), "downed-model") {
		t.Fatalf("expected wrapped transport error naming the model, got %v", err)
	}
}

func TestOpenAIGenerateContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL, "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.Generate(ctx, "p", GenParams{}); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestTruncateBoundsErrorBodies(t *testing.T) {
	if got := truncate("short", 200); got != "short" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("x", 300)
	got := truncate(long, 200)
	if len([]rune(got)) != 201 || !strings.HasSuffix(got, "…") {
		t.Fatalf("truncate shape: len %d", len(got))
	}
}

// The seed must reach the wire: it is what makes a temperature>0 run
// reproducible, and a study that declares seeds but never sends them would
// produce varied-but-unrepeatable grids while appearing correctly pinned.
func TestGenerateForwardsSeed(t *testing.T) {
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "mistral", "v0.3-q4km", WithHTTPClient(srv.Client()))
	if _, err := c.Generate(context.Background(), "scenario", GenParams{
		Temperature: Float64(0.8),
		MaxTokens:   Int(512),
		Seed:        Int(7),
	}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got.Seed == nil || *got.Seed != 7 {
		t.Fatalf("seed = %v, want 7", got.Seed)
	}
	if got.Temperature == nil || *got.Temperature != 0.8 {
		t.Fatalf("temperature = %v", got.Temperature)
	}
}
