package substrate

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeRunner answers per-command, so a test can make nvidia-smi succeed for the
// compute-apps query and fail for the device-name one.
type fakeRunner struct {
	computeApps string
	computeErr  error
	gpuName     string
	gpuNameErr  error
	calls       []string
}

func (f *fakeRunner) run(_ context.Context, name string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, name+" "+strings.Join(args, " "))
	joined := strings.Join(args, " ")
	switch {
	case strings.Contains(joined, "query-compute-apps"):
		return []byte(f.computeApps), f.computeErr
	case strings.Contains(joined, "query-gpu"):
		return []byte(f.gpuName), f.gpuNameErr
	}
	return nil, errors.New("unexpected command")
}

func TestProbeIdentifiesGPUFromComputeApps(t *testing.T) {
	f := &fakeRunner{
		// Real nvidia-smi shape: the server plus unrelated processes.
		computeApps: "/usr/lib/xorg/Xorg, 9\n/app/llama-server, 21184\n",
		gpuName:     "NVIDIA GeForce RTX 3090\n",
	}
	got := Probe(context.Background(), f.run)

	if got.Kind != KindGPU {
		t.Fatalf("kind = %q, want gpu", got.Kind)
	}
	if got.Device != "NVIDIA GeForce RTX 3090" {
		t.Errorf("device = %q", got.Device)
	}
	if got.VRAMMiB != 21184 {
		t.Errorf("vram = %d, want 21184", got.VRAMMiB)
	}
	if got.DetectedBy != "nvidia-smi" {
		t.Errorf("detected_by = %q — the mechanism must be named so a reader can weigh it", got.DetectedBy)
	}
}

// nvidia-smi answered and the server is not on it. That is evidence of CPU
// execution — the exact condition that went unnoticed for a full study.
func TestProbeIdentifiesCPUWhenServerHoldsNoGPUMemory(t *testing.T) {
	f := &fakeRunner{computeApps: "/usr/lib/xorg/Xorg, 9\n"}
	got := Probe(context.Background(), f.run)

	if got.Kind != KindCPU {
		t.Fatalf("kind = %q, want cpu", got.Kind)
	}
	if got.DetectedBy != "nvidia-smi" {
		t.Errorf("detected_by = %q", got.DetectedBy)
	}
	if !strings.Contains(got.Note, "llama-server") {
		t.Errorf("note should say what was looked for and not found: %q", got.Note)
	}
}

// The cardinal rule: a probe that cannot answer says so. An absent nvidia-smi
// does NOT mean CPU — an AMD or Metal host lands here too — and a confident
// wrong answer about the substrate is the failure being repaired.
func TestProbeReportsUnknownRatherThanGuessing(t *testing.T) {
	f := &fakeRunner{computeErr: errors.New("exec: nvidia-smi: not found")}
	got := Probe(context.Background(), f.run)

	if got.Kind != KindUnknown {
		t.Fatalf("kind = %q — a missing nvidia-smi must not be read as cpu or gpu", got.Kind)
	}
	if got.DetectedBy != "none" {
		t.Errorf("detected_by = %q, want none", got.DetectedBy)
	}
	if !strings.Contains(got.Note, "nvidia-smi") {
		t.Errorf("note must carry why it could not tell: %q", got.Note)
	}
	if got.Device != "" || got.VRAMMiB != 0 {
		t.Errorf("an unknown probe must invent nothing: %+v", got)
	}
}

// Knowing it is on a GPU but not which one is a partial answer, not a failure.
func TestProbeKeepsGPUFindingWhenDeviceNameIsUnavailable(t *testing.T) {
	f := &fakeRunner{
		computeApps: "/app/llama-server, 21184\n",
		gpuNameErr:  errors.New("query failed"),
	}
	got := Probe(context.Background(), f.run)

	if got.Kind != KindGPU {
		t.Fatalf("kind = %q — a missing device NAME must not lose the device FINDING", got.Kind)
	}
	if got.Device != "" {
		t.Errorf("device = %q, want empty rather than invented", got.Device)
	}
	if got.VRAMMiB != 21184 {
		t.Errorf("vram = %d", got.VRAMMiB)
	}
}

// Unparseable memory must not take the GPU finding down with it.
func TestProbeToleratesUnparseableMemory(t *testing.T) {
	f := &fakeRunner{computeApps: "/app/llama-server, [N/A]\n", gpuName: "NVIDIA GeForce RTX 3090"}
	got := Probe(context.Background(), f.run)

	if got.Kind != KindGPU {
		t.Fatalf("kind = %q", got.Kind)
	}
	if got.VRAMMiB != 0 {
		t.Errorf("vram = %d, want 0 for an unparseable figure", got.VRAMMiB)
	}
}

func TestProbeHandlesEmptyAndMalformedOutput(t *testing.T) {
	for name, out := range map[string]string{
		"empty":            "",
		"whitespace only":  "   \n  ",
		"no comma":         "/app/llama-server 21184\n",
		"unrelated procs":  "python, 512\nnode, 64\n",
		"blank final line": "/usr/lib/xorg/Xorg, 9\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeRunner{computeApps: out}
			// Every one of these is "nvidia-smi answered, server not on it".
			if got := Probe(context.Background(), f.run); got.Kind != KindCPU {
				t.Fatalf("kind = %q, want cpu", got.Kind)
			}
		})
	}
}

// The probe must not shell out at all until asked, and must query by process
// rather than assuming a slot.
func TestProbeQueriesComputeAppsFirst(t *testing.T) {
	f := &fakeRunner{computeApps: "/app/llama-server, 100\n", gpuName: "GPU"}
	Probe(context.Background(), f.run)

	if len(f.calls) == 0 || !strings.Contains(f.calls[0], "query-compute-apps") {
		t.Fatalf("first call = %v, want the compute-apps query", f.calls)
	}
}

// ExecRunner is the production seam; exercise it against a command that exists
// everywhere so the wiring itself is covered without needing a GPU.
func TestExecRunnerRunsAndReportsFailure(t *testing.T) {
	out, err := ExecRunner(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("ExecRunner: %v", err)
	}
	if strings.TrimSpace(string(out)) != "hello" {
		t.Errorf("out = %q", out)
	}
	if _, err := ExecRunner(context.Background(), "definitely-not-a-real-binary-xyz"); err == nil {
		t.Error("a missing binary must surface as an error, not an empty success")
	}
}
