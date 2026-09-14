// Package batteryrun is the host-side caller that runs the mechanized ALPHABET
// battery against a candidate entry. It is where the battery reaches the world:
// the battery package is sans-IO, so the filesystem read of the entry and its
// fallout-profile referent, and the capture of run provenance, live here.
//
// It runs the mechanized items (1, 2, 3, 4, 6, 9, 10, 11, 12, 13, 15); items 5,
// 7, 8, and 14 register as Deferred stubs in the sequence and are resolved by an
// assessor working the prose spec, in a separate context. The sequence fails
// fast on the first mechanized FAIL, so a candidate that fails early carries a
// recorded FAIL at that item rather than a full sweep.
package batteryrun

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
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

// FileRegistryReader satisfies battery.RegistryReader against the real
// filesystem: it reads the ALPHABET registry file and extracts the identity of
// every promoted entry, so Item 3 can dedupe the candidate against it. It is the
// IO edge Item 3 needs, kept out of the sans-IO battery package.
//
// An ABSENT registry file (Path unset, or ALPHABET.md not created yet) is a
// verified-empty registry: it returns no identities and no error, so Item 3
// passes a clean candidate. Only a real read failure returns an error, which
// Item 3 fails closed on. It reads the file's bytes mechanically to list
// identities — it does not load entry terrain into an agent's context.
type FileRegistryReader struct{ Path string }

// RegistryIdentities returns the promoted-entry identities in the registry file,
// parsed with battery.GlyphIdentity's rule so they compare against the candidate
// identically.
func (r FileRegistryReader) RegistryIdentities() ([]string, error) {
	if r.Path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(r.Path) //nolint:gosec // path is the operator-supplied registry location
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return battery.RegistryIdentitiesFrom(string(b)), nil
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
	// Registry is the reader Item 3 dedupes the candidate against. When nil it
	// defaults to an empty FileRegistryReader — a verified-empty registry, so a
	// clean candidate passes Item 3. A caller points it at the real ALPHABET.md to
	// get a real duplicate check.
	Registry battery.RegistryReader
}

// Run executes the mechanized battery against the candidate at candidatePath.
// It reads the candidate working-doc, extracts the assessable clean entry from
// it (battery.ExtractEntry: the fallout line plus the AC-4 assembled-glyph
// block, dropping the metadata header body, the AC-1/AC-2 derivation, and the
// project-scoped AC-3 specimen), and assesses that. It passes candidatePath as
// EntryPath, so Item 15 resolves the fallout reference relative to the
// candidate's own directory. This is the one-step path: the battery reads the
// canonical candidate definition directly, with no separate extract copy.
//
// The hard errors are failing to read the candidate and a candidate with no
// AC-4 glyph block; a failed provenance or props readback is recorded as a gap,
// never a reason to abort — a run that cannot fully describe itself is still a
// run.
func Run(ctx context.Context, candidatePath, itemID string, client model.Client, deps Deps, opts Options) (Result, error) {
	raw, err := os.ReadFile(candidatePath) //nolint:gosec // candidatePath is the corpus entry under assessment
	if err != nil {
		return Result{}, fmt.Errorf("batteryrun: read candidate %s: %w", candidatePath, err)
	}
	content, err := battery.ExtractEntry(string(raw))
	if err != nil {
		return Result{}, fmt.Errorf("batteryrun: extract entry from %s: %w", candidatePath, err)
	}

	registry := deps.Registry
	if registry == nil {
		registry = FileRegistryReader{}
	}

	seq := battery.RunSequence(ctx, battery.BuildBattery(), battery.Input{
		ItemID:         itemID,
		Content:        content,
		Model:          client,
		EntryPath:      candidatePath,
		Profiles:       FileProfileReader{},
		Registry:       registry,
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
