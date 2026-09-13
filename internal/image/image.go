// Package image reads content digests of assay images from rootless podman.
// The digest is what a run record names — image tags are mutable and never
// trusted, so a tag would say nothing about what actually executed. Recorded,
// not enforced: a digest differing from a prior run's is information about the
// two runs. The podman invocation is injectable so the parse/validate logic is
// testable without a container runtime.
package image

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Runner executes an image-inspect command and returns its stdout. The
// production runner shells out to podman (see cmd/lab); tests inject a fake.
// The podman exec itself lives in the command layer so this package stays
// sans-IO and fully unit-tested.
type Runner func(ctx context.Context, ref string) ([]byte, error)

// Digest returns the content digest of ref (e.g.
// "sha256:1a2b…") using the supplied runner. It validates the shape so a
// malformed or empty reply becomes an error rather than a bogus pin.
func Digest(ctx context.Context, run Runner, ref string) (string, error) {
	out, err := run(ctx, ref)
	if err != nil {
		return "", err
	}
	digest := strings.TrimSpace(string(out))
	if !isSHA256(digest) {
		return "", fmt.Errorf("image: %s produced malformed digest %q", ref, digest)
	}
	return digest, nil
}

// isSHA256 reports whether s has the exact shape "sha256:" + 64 chars. It does
// not validate the 64 chars are hex — the callers only need to tell a digest
// from a tag or an error string, and podman never emits a non-hex digest.
func isSHA256(s string) bool {
	return strings.HasPrefix(s, "sha256:") && len(s) == len("sha256:")+64
}

// ParsePinnedDigest splits a digest-pinned image reference "<repo>@sha256:<64>"
// into its repo and digest halves. ok is false when ref is not digest-pinned
// (a bare name or a ":tag" ref), so callers can leave tag refs on their
// existing path and only enrich the digest-pin case.
func ParsePinnedDigest(ref string) (repo, digest string, ok bool) {
	at := strings.LastIndex(ref, "@")
	if at < 0 {
		return "", "", false
	}
	repo, digest = ref[:at], ref[at+1:]
	if repo == "" || !isSHA256(digest) {
		return "", "", false
	}
	return repo, digest, true
}

// ParseDigestList reads the output of `podman images <repo> --format
// {{.Digest}}` — one digest per line — into a sorted, de-duplicated slice.
// Blank lines and podman's "<none>" placeholder for a dangling image are
// dropped, so the result names only real, referenceable digests.
func ParseDigestList(out []byte) []string {
	seen := map[string]bool{}
	var digests []string
	for _, line := range strings.Split(string(out), "\n") {
		d := strings.TrimSpace(line)
		if d == "" || d == "<none>" || seen[d] {
			continue
		}
		seen[d] = true
		digests = append(digests, d)
	}
	sort.Strings(digests)
	return digests
}

// AbsentPinError reports that a study definition pins an assay image by a
// content digest that is not present locally, so the run cannot launch.
//
// It exists because the raw failure is an opaque podman "exit status 125": a
// digest pin is permanent, but the image it names is not, so every rebuild of
// the assay image orphans the digest every prior study.toml pinned. The message
// names the missing digest and the digests that ARE present, so the operator can
// rebuild and re-pin, or pin to a digest that exists. It never substitutes a
// different image on its own — running a study on an image it did not name is
// exactly the silent-swap failure the lab is scarred by (see CLAUDE.md).
type AbsentPinError struct {
	Ref       string   // the full pinned reference from the study def
	Repo      string   // the repo half of Ref
	Digest    string   // the missing digest half of Ref
	Available []string // digests currently present locally for Repo
}

func (e *AbsentPinError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "image: pinned digest %s is not present locally for %s", e.Digest, e.Repo)
	switch len(e.Available) {
	case 0:
		fmt.Fprintf(&b, "; no image for %s is present — rebuild it (scripts/build-lab-images.sh) and re-pin the study", e.Repo)
	case 1:
		fmt.Fprintf(&b, "; the current build is %s — re-pin the study to it, or rebuild for the pinned digest", e.Available[0])
	default:
		fmt.Fprintf(&b, "; digests present are [%s] — re-pin the study to one, or rebuild for the pinned digest", strings.Join(e.Available, ", "))
	}
	return b.String()
}
