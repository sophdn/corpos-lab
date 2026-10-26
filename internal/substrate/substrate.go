// Package substrate identifies the processor an inference server is actually
// running on.
//
// This exists because the substrate is the variable that went unrecorded
// through an entire study and nobody noticed: a run silently fell back to CPU
// at 5.1 tok/s against ~46 on GPU, was written up as a positive control, and
// the resulting CPU-vs-GPU comparison was certified as reproduced. Nothing in
// the record could have shown it, because nothing looked.
//
// No llama.cpp HTTP endpoint reports the processor — /props describes the model
// and the sampler defaults, not the device — so identification is a host-side
// probe rather than something the client can ask for. It is deliberately
// separate from the in-band throughput signal each row already carries: that
// number is a MEASUREMENT and is always present; this is an IDENTIFICATION and
// may not be available at all.
//
// The cardinal rule here is that an unsuccessful probe reports Unknown. It
// never guesses GPU, and it never fails a run. A wrong confident answer about
// the substrate is the failure being repaired; silence that says it is silent
// is strictly better than a plausible fiction.
package substrate

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Kind is the processor class an inference server is running on.
type Kind string

const (
	// KindGPU means the probe found the server holding GPU memory.
	KindGPU Kind = "gpu"
	// KindCPU means the probe ran and found no GPU allocation for the server.
	KindCPU Kind = "cpu"
	// KindUnknown means the probe could not tell. It is the default, and it is
	// an honest answer — never upgrade it to a guess.
	KindUnknown Kind = "unknown"
)

// Info is what a probe could establish about the processor.
type Info struct {
	// Kind is the processor class, or KindUnknown when the probe couldn't tell.
	Kind Kind `json:"kind"`
	// Device names the hardware (e.g. "NVIDIA GeForce RTX 3090"), empty when
	// unknown.
	Device string `json:"device,omitempty"`
	// VRAMMiB is the memory the server process holds on the device.
	VRAMMiB int `json:"vram_mib,omitempty"`
	// DetectedBy names the mechanism that established this, or "none". It is
	// recorded so a reader can weigh the claim instead of taking it on faith.
	DetectedBy string `json:"detected_by"`
	// Note carries why a probe came back Unknown.
	Note string `json:"note,omitempty"`
}

// Runner executes a host command and returns its stdout. Injected so the probe
// is testable without a GPU, an nvidia-smi, or any host at all.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// ExecRunner is the production Runner, shelling out for real.
func ExecRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// process is the substring identifying the inference server among the GPU's
// compute apps. llama-server reports as /app/llama-server inside the container.
const process = "llama-server"

// Probe asks the host which processor the inference server is using.
//
// It returns an Info and never an error: a probe that cannot answer reports
// KindUnknown with the reason in Note. Failing to describe a run is not a
// reason to refuse it — that reflex is the one being retired.
func Probe(ctx context.Context, run Runner) Info {
	devices, err := run(ctx, "nvidia-smi",
		"--query-compute-apps=process_name,used_memory", "--format=csv,noheader,nounits")
	if err != nil {
		// No nvidia-smi, no driver, or no permission. We do not know that this
		// means CPU — an AMD or Metal host would land here too — so we say so.
		return Info{
			Kind:       KindUnknown,
			DetectedBy: "none",
			Note:       fmt.Sprintf("nvidia-smi unavailable: %v", err),
		}
	}

	for _, line := range strings.Split(strings.TrimSpace(string(devices)), "\n") {
		name, mem, ok := strings.Cut(line, ",")
		if !ok || !strings.Contains(name, process) {
			continue
		}
		info := Info{
			Kind:       KindGPU,
			DetectedBy: "nvidia-smi",
			VRAMMiB:    parseMiB(mem),
		}
		info.Device = deviceName(ctx, run)
		return info
	}

	// nvidia-smi answered and the server is not among the GPU's compute apps.
	// That is real evidence of CPU execution, not an absence of evidence.
	return Info{
		Kind:       KindCPU,
		DetectedBy: "nvidia-smi",
		Note:       process + " holds no GPU memory",
	}
}

// deviceName reads the GPU's model name. A failure here downgrades the label,
// not the finding: we already know it is on a GPU.
func deviceName(ctx context.Context, run Runner) string {
	out, err := run(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.Split(strings.TrimSpace(string(out)), "\n")[0])
}

func parseMiB(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "MiB")))
	if err != nil {
		return 0
	}
	return n
}
