package image

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseDigestFile reads deploy/IMAGE_DIGESTS.txt, which scripts/build-lab-images.sh
// writes on every build: one "<name>[:tag] sha256:<64>" line per image, with blank
// lines and "#" comments skipped. It returns image name (tag dropped) → digest. A
// line with the wrong field count, an empty name, or a malformed digest is an
// error, so a corrupt file never becomes a bogus pin.
func ParseDigestFile(raw []byte) (map[string]string, error) {
	m := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("image: malformed digests line %q (want \"<tag> sha256:<digest>\")", line)
		}
		name := fields[0]
		if i := strings.IndexByte(name, ':'); i >= 0 {
			name = name[:i]
		}
		if name == "" || !isSHA256(fields[1]) {
			return nil, fmt.Errorf("image: malformed digests line %q (want \"<tag> sha256:<digest>\")", line)
		}
		m[name] = fields[1]
	}
	return m, nil
}

// pinnedRef matches a local digest-pinned image reference. The name class stops at
// "@", so "lab-x-v2@…" is its own name, never a prefix match of "lab-x".
var pinnedRef = regexp.MustCompile(`localhost/([A-Za-z0-9._-]+)@(sha256:[0-9a-f]{64})`)

// Repin rewrites every "localhost/<name>@sha256:<digest>" in src whose name has a
// recorded digest, so it pins that digest. It returns the new text and how many
// references moved; a reference already on the recorded digest is not counted.
// Names with no recorded digest, tag refs, and "digest:<name>" study references
// are left untouched. This is the re-pin step after an image rebuild: it is applied
// only to files the operator names, because a study that has already run keeps its
// literal pin (that pin is what makes its result reproducible).
func Repin(src string, digests map[string]string) (string, int) {
	changed := 0
	out := pinnedRef.ReplaceAllStringFunc(src, func(ref string) string {
		m := pinnedRef.FindStringSubmatch(ref)
		want, ok := digests[m[1]]
		if !ok || want == m[2] {
			return ref
		}
		changed++
		return "localhost/" + m[1] + "@" + want
	})
	return out, changed
}
