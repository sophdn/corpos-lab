//go:build live

package model

// Opt-in live smoke against a local llama-server:
//
//	go test -tags live -run TestLive -v ./internal/model/
//
// Excluded from the gate (no build tag there) so commits stay hermetic. This
// file exists because the unit tests assert against a fake llama-server that we
// wrote, and a fake agrees with whatever we believed when we wrote it. Every
// assumption below was WRONG in at least one earlier draft of this code, and
// only the real server said so:
//
//   - /props is served at the ROOT; /v1/props is a 404.
//   - /props reports the server's LAUNCH DEFAULTS, not the effective sampler.
//   - the chat-completions response carries model / system_fingerprint /
//     timings, none of which we were reading.
//
// Run it when llama-server is up, and believe it over the fakes.

import (
	"context"
	"os"
	"testing"
	"time"
)

func liveURL() string {
	if base := os.Getenv("CORPOS_LAB_LLAMA_URL"); base != "" {
		return base
	}
	return "http://localhost:8081/v1"
}

// Props must decode the real nested shape. n_ctx lives under
// default_generation_settings, not at the top level.
func TestLivePropsDecodesTheRealServerShape(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	props, err := NewOpenAI(liveURL(), "probe", "v").Props(ctx)
	if err != nil {
		t.Fatalf("Props: %v (is llama-server up on %s?)", err, liveURL())
	}
	t.Logf("model_alias=%q build_info=%q n_ctx=%d", props.ModelAlias, props.BuildInfo, props.NCtx)

	if props.ModelAlias == "" || props.ModelPath == "" {
		t.Error("model identity empty — this is the field that catches a wrong-model swap")
	}
	if props.NCtx == 0 {
		t.Error("n_ctx = 0 — the nested default_generation_settings decode is wrong")
	}
	if props.BuildInfo == "" {
		t.Error("build_info empty — a server upgrade mid-study would be invisible")
	}

	// The trap, asserted so it cannot be quietly forgotten: these are the
	// server's LAUNCH defaults. They are what an UNPINNED param inherits — not
	// what any request ran under. Reading them back as "the config we ran"
	// would manufacture a confident false record.
	if len(props.Defaults) == 0 {
		t.Error("defaults empty — the inherited-sampler cross-check is gone")
	}
	t.Logf("defaults (what an UNPINNED param would inherit): top_k=%v temperature=%v",
		props.Defaults["top_k"], props.Defaults["temperature"])
}

// The real server must return the observations each row records. A fake that
// returns them proves nothing; this proves the field names are right.
func TestLiveGenerateReturnsServerObservations(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := NewOpenAI(liveURL(), "probe", "v").Generate(ctx, "Say the single word: ok", GenParams{
		Temperature: Float64(0.8), MaxTokens: Int(8), Seed: Int(3),
		TopNSigma: Float64(-1.0), TopK: Int(0), TypicalP: Float64(1.0),
		TopP: Float64(1.0), MinP: Float64(0.05),
		RepeatPenalty: Float64(1.0), RepeatLastN: Int(0),
		PresencePenalty: Float64(0), FrequencyPenalty: Float64(0),
		XTCProbability: Float64(0), DryMultiplier: Float64(0),
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	t.Logf("model=%q fingerprint=%q tok/s=%.1f", resp.Model, resp.SystemFingerprint,
		resp.Timings.PredictedPerSecond)

	if resp.Text == "" {
		t.Error("no text")
	}
	if resp.Model == "" {
		t.Error("server-reported model empty — declared-vs-served can't be compared")
	}
	if resp.SystemFingerprint == "" {
		t.Error("system_fingerprint empty")
	}
	// The substrate tripwire. On this box a 32B on GPU runs ~70 tok/s; the CPU
	// fallback that went unnoticed for a full study ran at 5.1. We assert only
	// that it is NON-ZERO — the number is a measurement, not a target, and
	// asserting a threshold here would be inventing a parity gate.
	if resp.Timings.PredictedPerSecond == 0 {
		t.Error("throughput 0 — the substrate tripwire is dead and a CPU fallback would be invisible again")
	}
	if resp.Timings.PredictedN == 0 {
		t.Error("predicted_n 0 — timings decode is wrong")
	}
}

// The load-bearing claim behind "sent == effective": llama.cpp's OpenAI
// endpoint honors the sampler extensions we send. It is not in the OpenAI spec,
// so nothing but the real server can establish it.
//
// This asserts the weaker, honest thing the endpoint can prove on its own: that
// sending the full chain is ACCEPTED rather than rejected or ignored into an
// error. The direct confirmation that each value lands on the slot came from
// polling /slots mid-request during design; see CLAUDE.md.
func TestLiveServerAcceptsTheFullSamplerChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Deliberately non-default values across every stage.
	if _, err := NewOpenAI(liveURL(), "probe", "v").Generate(ctx, "Say: ok", GenParams{
		Temperature: Float64(0.7), MaxTokens: Int(8), Seed: Int(999),
		TopNSigma: Float64(-1.0), TopK: Int(7), TypicalP: Float64(0.66),
		TopP: Float64(0.88), MinP: Float64(0.04),
		RepeatPenalty: Float64(1.05), RepeatLastN: Int(33),
		PresencePenalty: Float64(0.1), FrequencyPenalty: Float64(0.1),
		XTCProbability: Float64(0.22), DryMultiplier: Float64(0.44),
	}); err != nil {
		t.Fatalf("the server rejected the full sampler chain: %v", err)
	}
}
