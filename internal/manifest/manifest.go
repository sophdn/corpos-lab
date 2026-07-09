// Package manifest implements the freeze-by-digest rule for containerized
// assays (CHARTER.md): a study run is pinned to exact content digests of its
// executor-visible inputs — the assay image and every material file — and any
// divergence at verify time is a hard refusal, not a warning. A result whose
// manifest cannot be reproduced is an anecdote, not data.
package manifest

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"corpos-lab/internal/digest"
	"corpos-lab/internal/runner"
)

// ImagePin identifies an assay image by immutable content digest. The ref is
// recorded for humans; the digest is what verification compares — tags are
// mutable and never trusted.
type ImagePin struct {
	Ref    string `json:"ref"`
	Digest string `json:"digest"`
}

// RunManifest pins every executor-visible input of one assay run: the image,
// study.json, and each material file, plus the model identity. It is written
// beside the study record at pin time and re-checked before every launch.
type RunManifest struct {
	Assay        string            `json:"assay"`
	Image        ImagePin          `json:"image"`
	StudyDigest  string            `json:"study_digest"`
	Materials    map[string]string `json:"materials"`
	ModelID      string            `json:"model_id"`
	ModelVersion string            `json:"model_version"`
}

// Compute builds the manifest for the study laid out in inDir, pinning it to
// image. It hashes study.json and every material file the spec names.
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
	for _, n := range []string{spec.Materials.Scenario, spec.Materials.Glyph, spec.Materials.Ground} {
		if n != "" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// MismatchError is the hard-refusal error Verify returns when the live inputs
// diverge from the pinned manifest. It lists every divergence so an operator
// sees the full picture, not just the first.
type MismatchError struct {
	Divergences []string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("manifest: verification FAILED — %d divergence(s) from pinned manifest:\n  - %s",
		len(e.Divergences), strings.Join(e.Divergences, "\n  - "))
}

// Verify recomputes the manifest for inDir under the actually-observed image
// digest and compares it field-by-field against pinned. Any divergence —
// image digest, study.json content, a material's content, an added or removed
// material — is collected into a MismatchError. A nil return means the run is
// faithful to what was pinned and may proceed.
func Verify(inDir string, actualImage ImagePin, pinned RunManifest) error {
	current, err := Compute(inDir, actualImage)
	if err != nil {
		return err
	}

	var divergences []string

	if pinned.Image.Digest != current.Image.Digest {
		divergences = append(divergences, fmt.Sprintf(
			"image digest: pinned %s, got %s", short(pinned.Image.Digest), short(current.Image.Digest)))
	}
	if pinned.Assay != current.Assay {
		divergences = append(divergences, fmt.Sprintf(
			"assay: pinned %q, got %q", pinned.Assay, current.Assay))
	}
	if pinned.StudyDigest != current.StudyDigest {
		divergences = append(divergences, fmt.Sprintf(
			"study.json digest: pinned %s, got %s", short(pinned.StudyDigest), short(current.StudyDigest)))
	}
	if pinned.ModelID != current.ModelID {
		divergences = append(divergences, fmt.Sprintf(
			"model_id: pinned %q, got %q", pinned.ModelID, current.ModelID))
	}
	if pinned.ModelVersion != current.ModelVersion {
		divergences = append(divergences, fmt.Sprintf(
			"model_version: pinned %q, got %q", pinned.ModelVersion, current.ModelVersion))
	}
	divergences = append(divergences, diffMaterials(pinned.Materials, current.Materials)...)

	if len(divergences) > 0 {
		sort.Strings(divergences)
		return &MismatchError{Divergences: divergences}
	}
	return nil
}

// diffMaterials reports every material that was added, removed, or changed
// between the pinned and current manifests.
func diffMaterials(pinned, current map[string]string) []string {
	var out []string
	for name, pd := range pinned {
		cd, ok := current[name]
		switch {
		case !ok:
			out = append(out, fmt.Sprintf("material %q: pinned but missing from run", name))
		case pd != cd:
			out = append(out, fmt.Sprintf("material %q content: pinned %s, got %s", name, short(pd), short(cd)))
		}
	}
	for name := range current {
		if _, ok := pinned[name]; !ok {
			out = append(out, fmt.Sprintf("material %q: present in run but not pinned", name))
		}
	}
	return out
}

func short(sha string) string {
	if len(sha) <= 12 {
		return sha
	}
	return sha[:12]
}
