//go:build live

package battery

// Opt-in live smoke against a local llama-server:
//
//	go test -tags live -run TestLive -v ./internal/battery/
//
// Excluded from the gate (no build tag there) so commits stay hermetic;
// run it manually when llama-server is up to prove the inference items
// round-trip against the real OpenAI-compatible API.

import (
	"context"
	"os"
	"testing"
	"time"

	"corpos-lab/internal/model"
)

func TestLiveItem1AgainstLlamaServer(t *testing.T) {
	base := os.Getenv("CORPOS_LAB_LLAMA_URL")
	if base == "" {
		base = "http://localhost:8081/v1"
	}
	modelID := os.Getenv("CORPOS_LAB_LLAMA_MODEL")
	if modelID == "" {
		modelID = "Qwen2.5-32B-Instruct-Q4_K_M.gguf"
	}

	client := NewLiveClientForTest(base, modelID)
	content := fixture(t, "known_pass.md")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	st := &State{ItemID: "live-smoke", Content: content, Model: client}
	out := Item1XYZSpecificity(ctx, st)

	if out.Kind == OutcomeError {
		t.Fatalf("live item1 errored: %s", out.Message)
	}
	if out.Kind != OutcomeVerdict {
		t.Fatalf("unexpected outcome: %+v", out)
	}
	// The verdict itself is the model's call — the smoke asserts the
	// round-trip produced a typed verdict, not which one.
	t.Logf("live item1 verdict: %s", out.Verdict)
}

// NewLiveClientForTest builds the production OpenAI client for the live
// smoke; separated so the test body reads as the scenario.
func NewLiveClientForTest(base, modelID string) model.Client {
	return model.NewOpenAI(base, modelID, "live-smoke")
}
