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

// Every sampler stage must reach the wire. Temperature alone does not define a
// sampling distribution: a stage that is declared but silently dropped here
// inherits the server's own default, which is the failure this whole struct
// exists to prevent — and it fails invisibly, because the study file still says
// what the researcher intended.
func TestGenerateForwardsFullSamplerChain(t *testing.T) {
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "mistral", "v0.3-q4km", WithHTTPClient(srv.Client()))
	// Deliberately distinct values, none of them a llama.cpp default, so a
	// field crossed with its neighbour or left to default reads as a wrong
	// number rather than a plausible one.
	if _, err := c.Generate(context.Background(), "scenario", GenParams{
		Temperature:      Float64(0.7),
		MaxTokens:        Int(128),
		Seed:             Int(3),
		TopNSigma:        Float64(-1.0),
		TopK:             Int(11),
		TypicalP:         Float64(0.66),
		TopP:             Float64(0.88),
		MinP:             Float64(0.04),
		RepeatPenalty:    Float64(1.07),
		RepeatLastN:      Int(33),
		PresencePenalty:  Float64(0.22),
		FrequencyPenalty: Float64(0.11),
		XTCProbability:   Float64(0.15),
		DryMultiplier:    Float64(0.44),
	}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	for _, c := range []struct {
		field string
		got   any
		want  any
	}{
		{"temperature", derefF(got.Temperature), 0.7},
		{"max_tokens", derefI(got.MaxTokens), 128},
		{"seed", derefI(got.Seed), 3},
		{"top_n_sigma", derefF(got.TopNSigma), -1.0},
		{"top_k", derefI(got.TopK), 11},
		{"typical_p", derefF(got.TypicalP), 0.66},
		{"top_p", derefF(got.TopP), 0.88},
		{"min_p", derefF(got.MinP), 0.04},
		{"repeat_penalty", derefF(got.RepeatPenalty), 1.07},
		{"repeat_last_n", derefI(got.RepeatLastN), 33},
		{"presence_penalty", derefF(got.PresencePenalty), 0.22},
		{"frequency_penalty", derefF(got.FrequencyPenalty), 0.11},
		{"xtc_probability", derefF(got.XTCProbability), 0.15},
		{"dry_multiplier", derefF(got.DryMultiplier), 0.44},
	} {
		if c.got != c.want {
			t.Errorf("request %s = %v, want %v", c.field, c.got, c.want)
		}
	}
}

// A nil stage is omitted rather than sent as a zero — sending 0 for an unset
// min_p would silently DISABLE a stage the caller never spoke about, which is a
// different lie from the one we're fixing but a lie all the same.
func TestGenerateOmitsUnsetSamplerFields(t *testing.T) {
	var raw map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&raw)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "mistral", "v0.3-q4km", WithHTTPClient(srv.Client()))
	if _, err := c.Generate(context.Background(), "scenario", GenParams{
		Temperature: Float64(0.8),
	}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, absent := range []string{
		"top_k", "top_p", "min_p", "typical_p", "top_n_sigma", "repeat_penalty",
		"repeat_last_n", "presence_penalty", "frequency_penalty",
		"xtc_probability", "dry_multiplier", "seed", "max_tokens",
	} {
		if _, ok := raw[absent]; ok {
			t.Errorf("unset %s was sent as %v; nil must be omitted, not zeroed", absent, raw[absent])
		}
	}
}

// The server's account of what answered has to reach the caller, or a run
// cannot describe itself. Timings in particular is the substrate tripwire: the
// July 2026 CPU fallback ran at 5.1 tok/s against ~46 on GPU and nothing in the
// record could show it.
func TestGenerateCapturesServerReportedObservations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"role":"assistant","content":"body"}}],
			"model":"Mistral-7B-Instruct-v0.3.Q4_K_M.gguf",
			"system_fingerprint":"b9445-af6528e6d",
			"timings":{"prompt_n":31,"predicted_n":4,"prompt_per_second":330.29,"predicted_per_second":46.23}
		}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "asked-for-something-else", "v", WithHTTPClient(srv.Client()))
	resp, err := c.Generate(context.Background(), "p", GenParams{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Text != "body" {
		t.Fatalf("text = %q", resp.Text)
	}
	// Reported by the SERVER, not echoed from what we asked for.
	if resp.Model != "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf" {
		t.Errorf("model = %q, want the server's answer, not the request's", resp.Model)
	}
	if resp.SystemFingerprint != "b9445-af6528e6d" {
		t.Errorf("system_fingerprint = %q", resp.SystemFingerprint)
	}
	if resp.Timings.PredictedPerSecond != 46.23 || resp.Timings.PromptPerSecond != 330.29 {
		t.Errorf("timings rates = %+v", resp.Timings)
	}
	if resp.Timings.PromptN != 31 || resp.Timings.PredictedN != 4 {
		t.Errorf("timings counts = %+v", resp.Timings)
	}
}

// /props is served at the ROOT, not under /v1: llama-server answers /props with
// 200 and /v1/props with 404. Appending to baseURL would 404 against the real
// server while passing any test that used a naive fake, so the derivation is
// asserted directly.
func TestPropsURLIsDerivedFromTheServerRoot(t *testing.T) {
	for _, c := range []struct{ baseURL, want string }{
		{"http://llama-server:8081/v1", "http://llama-server:8081/props"},
		{"http://llama-server:8081/v1/", "http://llama-server:8081/props"},
		{"http://llama-server:8081", "http://llama-server:8081/props"},
		{"http://llama-server:8081/", "http://llama-server:8081/props"},
	} {
		got := NewOpenAI(c.baseURL, "m", "v").propsURL()
		if got != c.want {
			t.Errorf("propsURL(%q) = %q, want %q", c.baseURL, got, c.want)
		}
	}
}

func TestPropsReadsServerSelfReport(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		_, _ = w.Write([]byte(`{
			"model_path":"/models/Mistral-7B-Instruct-v0.3.Q4_K_M.gguf",
			"model_alias":"Mistral-7B-Instruct-v0.3.Q4_K_M.gguf",
			"build_info":"b9445-af6528e6d",
			"default_generation_settings":{"n_ctx":8192,"params":{"temperature":0.8,"top_k":40}}
		}`))
	}))
	defer srv.Close()

	c := NewOpenAI(srv.URL+"/v1", "m", "v", WithHTTPClient(srv.Client()))
	props, err := c.Props(context.Background())
	if err != nil {
		t.Fatalf("Props: %v", err)
	}
	if gotPath != "/props" || gotMethod != http.MethodGet {
		t.Fatalf("request = %s %s, want GET /props", gotMethod, gotPath)
	}
	if props.ModelAlias != "Mistral-7B-Instruct-v0.3.Q4_K_M.gguf" {
		t.Errorf("model_alias = %q", props.ModelAlias)
	}
	if props.ModelPath != "/models/Mistral-7B-Instruct-v0.3.Q4_K_M.gguf" {
		t.Errorf("model_path = %q", props.ModelPath)
	}
	if props.BuildInfo != "b9445-af6528e6d" {
		t.Errorf("build_info = %q", props.BuildInfo)
	}
	// n_ctx is nested under default_generation_settings, not top level.
	if props.NCtx != 8192 {
		t.Errorf("n_ctx = %d, want 8192", props.NCtx)
	}
	// Defaults is what an OMITTED param would have inherited — kept as a
	// cross-check, never as a claim about what this run used.
	if props.Defaults["top_k"] != float64(40) {
		t.Errorf("defaults[top_k] = %v, want 40", props.Defaults["top_k"])
	}
}

func TestPropsErrorPaths(t *testing.T) {
	t.Run("non-200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("props exploded"))
		}))
		defer srv.Close()
		_, err := NewOpenAI(srv.URL+"/v1", "m", "v", WithHTTPClient(srv.Client())).Props(context.Background())
		if err == nil || !strings.Contains(err.Error(), "500") {
			t.Fatalf("err = %v, want a 500", err)
		}
	})
	t.Run("garbage JSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{not json"))
		}))
		defer srv.Close()
		_, err := NewOpenAI(srv.URL+"/v1", "m", "v", WithHTTPClient(srv.Client())).Props(context.Background())
		if err == nil || !strings.Contains(err.Error(), "decode props") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("transport error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		c := NewOpenAI(srv.URL+"/v1", "m", "v", WithHTTPClient(srv.Client()))
		srv.Close()
		if _, err := c.Props(context.Background()); err == nil {
			t.Fatal("want a transport error against a closed server")
		}
	})
	t.Run("cancelled context", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		defer srv.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := NewOpenAI(srv.URL+"/v1", "m", "v", WithHTTPClient(srv.Client()))
		if _, err := c.Props(ctx); err == nil {
			t.Fatal("want a context error")
		}
	})
}

func derefF(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func derefI(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
