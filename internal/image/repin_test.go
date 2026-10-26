package image

import (
	"strings"
	"testing"
)

const (
	oldLoop = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	newLoop = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	oldRaw  = "sha256:3333333333333333333333333333333333333333333333333333333333333333"
	newRaw  = "sha256:4444444444444444444444444444444444444444444444444444444444444444"
)

const digestFile = `# corpos-lab assay image digests — recorded by scripts/build-lab-images.sh
# Study records pin the variant digest; a verify mismatch is a hard refusal.
lab-base:dev sha256:5555555555555555555555555555555555555555555555555555555555555555

lab-agentic-loop-probe:dev ` + newLoop + `
lab-grounded-glyph-probe:dev ` + newRaw + `
`

func TestParseDigestFile(t *testing.T) {
	got, err := ParseDigestFile([]byte(digestFile))
	if err != nil {
		t.Fatalf("ParseDigestFile: %v", err)
	}
	want := map[string]string{
		"lab-base":                 "sha256:5555555555555555555555555555555555555555555555555555555555555555",
		"lab-agentic-loop-probe":   newLoop,
		"lab-grounded-glyph-probe": newRaw,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries %v, want %d", len(got), got, len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("digest[%s] = %q, want %q", k, got[k], v)
		}
	}
}

func TestParseDigestFileUntaggedName(t *testing.T) {
	got, err := ParseDigestFile([]byte("lab-x " + newLoop + "\n"))
	if err != nil {
		t.Fatalf("ParseDigestFile: %v", err)
	}
	if got["lab-x"] != newLoop {
		t.Errorf("untagged name not indexed: %v", got)
	}
}

func TestParseDigestFileRejects(t *testing.T) {
	cases := map[string]string{
		"wrong field count": "one two three\n",
		"malformed digest":  "lab-x:dev sha256:abc\n",
		"empty name":        ":dev " + newLoop + "\n",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseDigestFile([]byte(in)); err == nil {
				t.Errorf("expected an error for %q", in)
			}
		})
	}
}

func TestRepinRewritesEveryRecordedImage(t *testing.T) {
	digests, err := ParseDigestFile([]byte(digestFile))
	if err != nil {
		t.Fatal(err)
	}
	src := `name = "x"
image = "localhost/lab-agentic-loop-probe@` + oldLoop + `"
# also a raw one: localhost/lab-grounded-glyph-probe@` + oldRaw + `
other = "localhost/lab-agentic-loop-probe@` + oldLoop + `"
`
	got, n := Repin(src, digests)
	if n != 3 {
		t.Errorf("changed = %d, want 3", n)
	}
	if strings.Contains(got, oldLoop) || strings.Contains(got, oldRaw) {
		t.Errorf("old digest survived:\n%s", got)
	}
	if strings.Count(got, "lab-agentic-loop-probe@"+newLoop) != 2 ||
		strings.Count(got, "lab-grounded-glyph-probe@"+newRaw) != 1 {
		t.Errorf("new digests not stamped:\n%s", got)
	}
}

func TestRepinCurrentPinIsNotAChange(t *testing.T) {
	digests := map[string]string{"lab-agentic-loop-probe": newLoop}
	src := `image = "localhost/lab-agentic-loop-probe@` + newLoop + `"`
	got, n := Repin(src, digests)
	if n != 0 || got != src {
		t.Errorf("Repin on a current pin = (%q, %d), want unchanged, 0", got, n)
	}
}

func TestRepinLeavesUnrecordedImagesAlone(t *testing.T) {
	digests := map[string]string{"lab-agentic-loop-probe": newLoop}
	src := `a = "localhost/lab-other-probe@` + oldLoop + `"
b = "localhost/lab-agentic-loop-probe-v2@` + oldLoop + `"
c = "digest:lab-agentic-loop-probe"
d = "localhost/lab-agentic-loop-probe:dev"`
	got, n := Repin(src, digests)
	if n != 0 || got != src {
		t.Errorf("Repin touched an unrecorded or non-digest ref: (%q, %d)", got, n)
	}
}
