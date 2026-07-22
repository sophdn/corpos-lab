// Package batteryrun is the host-side caller that runs the mechanized ALPHABET
// battery against a candidate entry. It is where the battery reaches the world:
// the battery package is sans-IO, so the filesystem read of the entry and its
// fallout-profile referent, and the capture of run provenance, live here.
//
// It runs only the mechanized items (1, 2, 4, 9, 10, 15); the nine judged items
// register as Deferred stubs in the sequence and are resolved by an assessor
// working the prose spec, in a separate context. The sequence fails fast on the
// first mechanized FAIL, so a candidate that fails early carries a recorded FAIL
// at that item rather than a full sweep.
package batteryrun

import (
	"context"
	"fmt"
	"os"

	"corpos-lab/internal/battery"
	"corpos-lab/internal/model"
	"corpos-lab/internal/provenance"
	"corpos-lab/internal/substrate"
)

// FileProfileReader satisfies battery.ProfileReader against the real
// filesystem. It is the IO edge Item 15 needs, kept out of the sans-IO battery
// package: the step resolves the referenced path (pure string math) and this
// opens it.
type FileProfileReader struct{}

// ReadProfile reads the fallout-profile document at the already-resolved path.
func (FileProfileReader) ReadProfile(path string) (string, error) {
	b, err := os.ReadFile(path) //nolint:gosec // path is a corpus doc reference, resolved by the caller
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RunProvenance is the run-level record of what the mechanized battery ran
// under. Every model-assessed item in a run hits the same server with the same
// pinned sampler, so this is captured once and applies to each of them —
// faithful because the server, build, substrate, and sampler do not change
// mid-run. Nothing here may fail a run: a readback that could not be taken is
// recorded as the gap it is.
type RunProvenance struct {
	// Sampler is what the model-assessed steps send (pinned, deterministic).
	Sampler model.GenParams `json:"sampler"`
	// ModelID and ModelVersion are what the run DECLARED (the client's Name and
	// Version); Props.ModelAlias/ModelPath is what the server says it served.
	ModelID      string `json:"model_id"`
	ModelVersion string `json:"model_version"`
	// Props is the server's /props self-report; PropsError records why it is
	// empty when the readback failed.
	Props      model.ServerProps `json:"props"`
	PropsError string            `json:"props_error,omitempty"`
	// Substrate is the processor the inference server ran on.
	Substrate substrate.Info `json:"substrate"`
	// Provenance is the instrument's repo state; ProvenanceError records why it
	// is empty when the stamp could not be taken.
	Provenance      provenance.Stamp `json:"provenance"`
	ProvenanceError string           `json:"provenance_error,omitempty"`
}

// Result is one candidate's mechanized battery pass: the sequence result (the
// per-item verdicts for the mechanized items, Deferred for the judged stubs,
// fail-fast on the first mechanized FAIL unless Options.AllItems suspends it)
// plus the run-level provenance. AllItems records which of the two the run
// executed under, so a reader can tell an unassessed item from a measured one
// without inferring it from where the verdicts stop.
type Result struct {
	CandidatePath string         `json:"candidate_path"`
	ItemID        string         `json:"item_id"`
	AllItems      bool           `json:"all_items"`
	Sequence      battery.Result `json:"sequence"`
	Provenance    RunProvenance  `json:"provenance"`
}

// Options are the run's execution choices, as opposed to Deps' world-seams.
type Options struct {
	// AllItems runs every battery item even after one fails, instead of
	// stopping at the first failure. The run verdict is unchanged; the later
	// items get measured rather than left unassessed. See battery.Input's
	// ContinueOnFail for why the default is still fail-fast.
	AllItems bool
}

// Deps are the seams a run reaches the world through, injected so Run is
// testable without a GPU, a git repo, or a live server. Each is expected to
// report a gap rather than fail: Substrate returns KindUnknown, and Props and
// Provenance return an error that Run records without aborting.
type Deps struct {
	Substrate  func(ctx context.Context) substrate.Info
	Provenance func(ctx context.Context) (provenance.Stamp, error)
	Props      func(ctx context.Context) (model.ServerProps, error)
}

// Run executes the mechanized battery against the candidate at candidatePath.
// It reads the candidate as the entry content and passes its path as EntryPath,
// so Item 15 resolves the fallout reference relative to it. The only hard error
// is failing to read the candidate itself; a failed provenance or props
// readback is recorded as a gap, never a reason to abort — a run that cannot
// fully describe itself is still a run.
func Run(ctx context.Context, candidatePath, itemID string, client model.Client, deps Deps, opts Options) (Result, error) {
	content, err := os.ReadFile(candidatePath) //nolint:gosec // candidatePath is the corpus entry under assessment
	if err != nil {
		return Result{}, fmt.Errorf("batteryrun: read candidate %s: %w", candidatePath, err)
	}

	seq := battery.RunSequence(ctx, battery.BuildBattery(), battery.Input{
		ItemID:         itemID,
		Content:        string(content),
		Model:          client,
		EntryPath:      candidatePath,
		Profiles:       FileProfileReader{},
		ContinueOnFail: opts.AllItems,
	})

	prov := RunProvenance{
		Sampler:      battery.VerdictGenParams(),
		ModelID:      client.Name(),
		ModelVersion: client.Version(),
		Substrate:    deps.Substrate(ctx),
	}
	if props, perr := deps.Props(ctx); perr != nil {
		prov.PropsError = perr.Error()
	} else {
		prov.Props = props
	}
	if stamp, serr := deps.Provenance(ctx); serr != nil {
		prov.ProvenanceError = serr.Error()
	} else {
		prov.Provenance = stamp
	}

	return Result{
		CandidatePath: candidatePath,
		ItemID:        itemID,
		AllItems:      opts.AllItems,
		Sequence:      seq,
		Provenance:    prov,
	}, nil
}
