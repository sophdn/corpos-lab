package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRawChatReturnsUnsplitContentAndReasoning(t *testing.T) {
	// A reply whose content carries an inline <think> block. Generate would move
	// the block into Reasoning and leave Text as the answer; RawChat must return
	// the content verbatim so a rater matches a code where the server put it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"served-m","system_fingerprint":"b9-abc",` +
			`"choices":[{"message":{"content":"<think>Ii</think>C","reasoning_content":"trace"},` +
			`"finish_reason":"length"}],` +
			`"usage":{"prompt_tokens":11,"completion_tokens":2,"total_tokens":13,` +
			`"prompt_cache_hit_tokens":8,"prompt_cache_miss_tokens":3}}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "qwen", "q4km", WithHTTPClient(srv.Client()))
	got, err := c.RawChat(context.Background(), "rate this", GenParams{Temperature: Float64(0)})
	if err != nil {
		t.Fatalf("RawChat: %v", err)
	}
	if got.Content != "<think>Ii</think>C" {
		t.Fatalf("Content = %q, want the verbatim reply (no think split)", got.Content)
	}
	if got.ReasoningContent != "trace" {
		t.Fatalf("ReasoningContent = %q, want %q", got.ReasoningContent, "trace")
	}
	if got.Model != "served-m" || got.ResponseID != "cmpl-1" || got.SystemFingerprint != "b9-abc" {
		t.Fatalf("provenance fields = %+v", got)
	}
	if !got.Truncated {
		t.Fatal("Truncated must be true when finish_reason is length")
	}
	want := Usage{PromptTokens: 11, CompletionTokens: 2, TotalTokens: 13, PromptCacheHitTokens: 8, PromptCacheMissTokens: 3}
	if got.Usage != want {
		t.Fatalf("Usage = %+v, want %+v", got.Usage, want)
	}
}

func TestRawChatSendsBearerTokenWhenSet(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"C"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "deepseek-flash", "", WithHTTPClient(srv.Client()), WithBearerToken("sk-secret"))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err != nil {
		t.Fatalf("RawChat: %v", err)
	}
	if gotAuth != "Bearer sk-secret" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer sk-secret")
	}
}

func TestRawChatSendsNoAuthHeaderWhenUnset(t *testing.T) {
	var hadAuth bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"C"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "local", "", WithHTTPClient(srv.Client()))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err != nil {
		t.Fatalf("RawChat: %v", err)
	}
	if hadAuth {
		t.Fatal("a local call must send no Authorization header")
	}
}

func TestRawChatErrorsOnNonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err == nil {
		t.Fatal("want an error on a 401 status")
	}
}

func TestRawChatErrorsOnMalformedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err == nil {
		t.Fatal("want a decode error on a malformed response body")
	}
}

func TestRawChatErrorsOnTransportFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // now nothing is listening: the request fails at the transport
	c := NewOpenAI(url+"/v1", "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err == nil {
		t.Fatal("want a transport error when the server is down")
	}
}

func TestRawChatErrorsOnNoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "m", "", WithHTTPClient(srv.Client()))
	if _, err := c.RawChat(context.Background(), "p", GenParams{}); err == nil {
		t.Fatal("want an error when the server returns no choices")
	}
}
