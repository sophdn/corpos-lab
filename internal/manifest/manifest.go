// Package manifest records the content digests of a run's inputs — the assay
// image, study.json, and every material file — so a run record says exactly
// what was fed in.
//
// It RECORDS; it does not enforce. The verify-or-refuse arm (Verify,
// MismatchError) was deleted 2026-07-14 along with the freeze-by-digest rule it
// served (INQUIRY.md §What changed the method). That arm was never called on the
// run path anyway — control.RunStudy only ever called Compute — so the freeze
// had the authority of a mechanism with none of the checking.
//
// The deeper reason it went: these digests cover the STIMULUS (scenario, glyph,
// ground, rubric) which by design never changes. The variables that actually
// moved between runs — temperature, the sampler, and the processor — were not
// captured here at all. Pinning the stimulus and refusing on its drift protected
// nothing while certifying results as frozen. What replaces it is observation:
// capture the effective config the server actually used, and the substrate.
//
// A digest that differs from a prior run is information about the two runs. It
// is not grounds for refusing to run.
package manifest

import (
	"fmt"
	"path/filepath"
	"sort"

	"corpos-lab/internal/digest"
	"corpos-lab/internal/runner"
)

// ImagePin identifies an assay image by immutable content digest. The ref is
// recorded for humans; the digest is what identifies the bytes that ran — tags
// are mutable and say nothing about content.
type ImagePin struct {
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

// RunManifest records the content digests of one assay run's inputs: the image,
// study.json, and each material file, plus the model identity. It is captured at
// launch and stored with the run.
//
// It is a record of what was fed in, not a promise about what may run. Nothing
// re-checks it; nothing refuses on a mismatch.
type RunManifest struct {
	Assay        string            `json:"assay"`
	Image        ImagePin          `json:"image"`
	StudyDigest  string            `json:"study_digest"`
	Materials    map[string]string `json:"materials"`
	ModelID      string            `json:"model_id"`
	ModelVersion string            `json:"model_version"`
}

// Compute captures the input digests for the study laid out in inDir under the
// image it ran with. It hashes study.json and every material file the spec names.
func Compute(inDir string, image ImagePin) (RunManifest, error) {
	spec, err := runner.LoadSpec(inDir)
	if err != nil {
		return RunManifest{}, err
	}

	studyDigest, err := digest.File(filepath.Join(inDir, "study.json"))
	if err != nil {
		return RunManifest{}, fmt.Errorf("manifest: digest study.json: %w", err)
	}

	materials := map[string]string{}
	for _, name := range namedMaterials(spec) {
		d, err := digest.File(filepath.Join(inDir, name))
		if err != nil {
			return RunManifest{}, fmt.Errorf("manifest: digest material %q: %w", name, err)
		}
		materials[name] = d
	}

	return RunManifest{
		Assay:        spec.Assay,
		Image:        image,
		StudyDigest:  studyDigest,
		Materials:    materials,
		ModelID:      spec.Model.ModelID,
		ModelVersion: spec.Model.Version,
	}, nil
}

// namedMaterials returns the material filenames the spec references, in a
// stable order, skipping the optional ones that are unset.
func namedMaterials(spec runner.StudySpec) []string {
	var names []string
	for _, n := range []string{spec.Materials.Scenario, spec.Materials.Glyph, spec.Materials.Ground, spec.Materials.Imperative} {
		if n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}
