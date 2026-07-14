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
	if !strings.HasPrefix(digest, "sha256:") || len(digest) != len("sha256:")+64 {
		return "", fmt.Errorf("image: %s produced malformed digest %q", ref, digest)
	}
	return digest, nil
}
