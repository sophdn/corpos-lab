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
	"encoding/json"
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

// ReferenceReader supplies the corpus reference material the model-assessed
// items deliver inline (item 6 traces fields against the glyph definition, and
// against the provenance taxonomy when the entry cites a provenance type). It is
// defined here — the consumer that assembles the battery Input — because the
// battery package only consumes the resulting text (battery.ReferenceMaterial),
// not the IO that produces it.
type ReferenceReader interface {
	Reference() (battery.ReferenceMaterial, error)
}

// FileReferenceReader satisfies ReferenceReader against the real filesystem: it
// reads the glyph-definition and provenance-type docs so item 6 can hand them to
// the judge. It is the IO edge kept out of the sans-IO battery package. An unset
// path reads as empty (that half of the reference is simply not delivered); a
// real read failure returns the error, which Run records as a gap without
// aborting — nothing here may fail a run.
type FileReferenceReader struct {
	DefinitionPath      string
	ProvenanceTypesPath string
}

// Reference reads the configured reference docs into a battery.ReferenceMaterial.
func (r FileReferenceReader) Reference() (battery.ReferenceMaterial, error) {
	var m battery.ReferenceMaterial
	if r.DefinitionPath != "" {
		b, err := os.ReadFile(r.DefinitionPath) //nolint:gosec // path is a corpus reference doc, operator-supplied
		if err != nil {
			return battery.ReferenceMaterial{}, fmt.Errorf("read glyph definition %s: %w", r.DefinitionPath, err)
		}
		m.GlyphDefinition = string(b)
	}
	if r.ProvenanceTypesPath != "" {
		b, err := os.ReadFile(r.ProvenanceTypesPath) //nolint:gosec // path is a corpus reference doc, operator-supplied
		if err != nil {
			return battery.ReferenceMaterial{}, fmt.Errorf("read provenance types %s: %w", r.ProvenanceTypesPath, err)
		}
		m.ProvenanceTypes = string(b)
	}
	return m, nil
}

// BaselineReader supplies the unguided-baseline measurement item 12 renders its
// empirical default-alignment verdict from: for the candidate's decision class,
// whether the class's target behavior fires when the scenario is run UNGUIDED
// (no glyph prefix) across the weak local treatment shelf. It is defined here —
// the consumer that assembles the battery Input — because the battery package
// only consumes the resulting measurement (battery.BaselineOutcome), not the
// inference that produces it. The baseline inference is a live control run that
// cannot happen inside the sans-IO battery package; it runs at this edge (or
// upstream, persisted to a file the reader loads).
type BaselineReader interface {
	Baseline() (battery.BaselineOutcome, error)
}

// FileBaselineReader satisfies BaselineReader by reading a JSON-encoded
// battery.BaselineOutcome from disk — the record a prior unguided-baseline run
// (the baseline/control arm across the weak shelf) wrote out. It is the IO edge
// kept out of the sans-IO battery package.
//
// An unset path reads as an empty outcome with no error: no baseline was
// supplied, so item 12 defers with a recorded gap rather than failing the run.
// A real read or JSON-parse failure returns the error, which Run records as a
// gap without aborting — nothing here may fail a run.
type FileBaselineReader struct{ Path string }

// Baseline reads the baseline-outcome JSON at the configured path.
func (r FileBaselineReader) Baseline() (battery.BaselineOutcome, error) {
	if r.Path == "" {
		return battery.BaselineOutcome{}, nil
	}
	b, err := os.ReadFile(r.Path) //nolint:gosec // path is the operator-supplied baseline record
	if err != nil {
		return battery.BaselineOutcome{}, fmt.Errorf("read baseline %s: %w", r.Path, err)
	}
	var out battery.BaselineOutcome
	if err := json.Unmarshal(b, &out); err != nil {
		return battery.BaselineOutcome{}, fmt.Errorf("parse baseline %s: %w", r.Path, err)
	}
	return out, nil
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
	// ReferenceError records why the model-assessed items ran WITHOUT their inline
	// reference material (a failed reference read). Item 6 then falls back to its
	// definition-from-memory prompt, whose verdicts are the ones the inline
	// delivery exists to make trustworthy — so the gap is recorded, not hidden.
	ReferenceError string `json:"reference_error,omitempty"`
	// Baseline is the unguided-baseline measurement item 12's empirical verdict was
	// rendered against: which weak-shelf models ran the class scenario UNGUIDED and
	// whether the class's target behavior fired without the glyph. Recorded so the
	// population-relative item-12 verdict is auditable and honest about the
	// population it is relative to. Empty when no baseline was supplied (item 12
	// then defers with a recorded gap).
	Baseline battery.BaselineOutcome `json:"baseline"`
	// BaselineError records why item 12 ran WITHOUT its unguided baseline (a failed
	// baseline read). Item 12 then defers rather than failing the run — an
	// empirical item with no measurement cannot render a verdict, and refusing the
	// run for want of it is the freeze reflex.
	BaselineError string `json:"baseline_error,omitempty"`
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
	// Reference supplies the inline reference material the model-assessed items
	// deliver to the judge. When nil the run proceeds with no reference — item 6
	// uses its definition-from-memory prompt — and no gap is recorded (the caller
	// chose not to wire it). When set but the read fails, the gap is recorded in
	// RunProvenance.ReferenceError and the run still proceeds.
	Reference ReferenceReader
	// Baseline supplies the unguided-baseline measurement item 12 renders its
	// empirical default-alignment verdict from. When nil the run proceeds with no
	// baseline — item 12 defers with a recorded gap — and no read is attempted
	// (the caller chose not to wire it). When set but the read fails, the gap is
	// recorded in RunProvenance.BaselineError and the run still proceeds.
	Baseline BaselineReader
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

	// Read the inline reference material before running the sequence. A failed
	// read is recorded as a gap and the run proceeds with an empty reference —
	// item 6 falls back to its definition-from-memory prompt rather than failing
	// the run for want of the reference.
	var reference battery.ReferenceMaterial
	referenceError := ""
	if deps.Reference != nil {
		ref, rerr := deps.Reference.Reference()
		if rerr != nil {
			referenceError = rerr.Error()
		} else {
			reference = ref
		}
	}

	// Read the unguided baseline item 12 renders its empirical verdict from. A
	// failed read is recorded as a gap and the run proceeds with an empty baseline
	// — item 12 defers rather than failing the run for want of the measurement.
	var baseline battery.BaselineOutcome
	baselineError := ""
	if deps.Baseline != nil {
		bl, berr := deps.Baseline.Baseline()
		if berr != nil {
			baselineError = berr.Error()
		} else {
			baseline = bl
		}
	}

	seq := battery.RunSequence(ctx, battery.BuildBattery(), battery.Input{
		ItemID:         itemID,
		Content:        content,
		Model:          client,
		EntryPath:      candidatePath,
		Profiles:       FileProfileReader{},
		Registry:       registry,
		Reference:      reference,
		Baseline:       baseline,
		ContinueOnFail: opts.AllItems,
	})

	prov := RunProvenance{
		Sampler:        battery.VerdictGenParams(),
		ModelID:        client.Name(),
		ModelVersion:   client.Version(),
		Substrate:      deps.Substrate(ctx),
		ReferenceError: referenceError,
		Baseline:       baseline,
		BaselineError:  baselineError,
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
