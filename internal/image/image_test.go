package image

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func fakeRunner(out string, err error) Runner {
	return func(_ context.Context, _ string) ([]byte, error) {
		return []byte(out), err
	}
}

func TestDigestParsesValidSHA256(t *testing.T) {
	valid := "sha256:" + strings.Repeat("a", 64)
	got, err := Digest(context.Background(), fakeRunner(valid+"\n", nil), "localhost/lab-grounded-glyph-probe:dev")
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	if got != valid {
		t.Fatalf("got %q, want %q", got, valid)
	}
}

func TestDigestRejectsMalformed(t *testing.T) {
	cases := []string{
		"",
		"not-a-digest",
		"sha256:tooshort",
		"sha256:" + strings.Repeat("a", 63), // one short
		strings.Repeat("a", 64),             // missing prefix
	}
	for _, c := range cases {
		if _, err := Digest(context.Background(), fakeRunner(c, nil), "ref"); err == nil {
			t.Fatalf("expected error for malformed digest %q", c)
		}
	}
}

func TestDigestPropagatesRunnerError(t *testing.T) {
	_, err := Digest(context.Background(), fakeRunner("", errors.New("no such image")), "ref")
	if err == nil {
		t.Fatal("expected runner error to propagate")
	}
}

func TestParsePinnedDigest(t *testing.T) {
	valid := "sha256:" + strings.Repeat("a", 64)
	cases := []struct {
		ref       string
		repo, dig string
		wantOK    bool
	}{
		{"localhost/lab-grounded-glyph-probe@" + valid, "localhost/lab-grounded-glyph-probe", valid, true},
		{"localhost/lab-grounded-glyph-probe:dev", "", "", false},
		{"lab-base", "", "", false},
		{"repo@sha256:tooshort", "", "", false},
		{"@" + valid, "", "", false},
	}
	for _, c := range cases {
		repo, dig, ok := ParsePinnedDigest(c.ref)
		if ok != c.wantOK || repo != c.repo || dig != c.dig {
			t.Fatalf("ParsePinnedDigest(%q) = (%q,%q,%v), want (%q,%q,%v)",
				c.ref, repo, dig, ok, c.repo, c.dig, c.wantOK)
		}
	}
}

func TestParseDigestList(t *testing.T) {
	a := "sha256:" + strings.Repeat("a", 64)
	b := "sha256:" + strings.Repeat("b", 64)
	out := []byte(a + "\n\n<none>\n" + b + "\n" + a + "\n")
	got := ParseDigestList(out)
	want := []string{a, b} // sorted, de-duped, <none>/blank dropped
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("ParseDigestList = %v, want %v", got, want)
	}
	if ParseDigestList([]byte("\n<none>\n")) != nil {
		t.Fatal("expected nil for a list with no real digests")
	}
}

func TestAbsentPinErrorMessage(t *testing.T) {
	a := "sha256:" + strings.Repeat("a", 64)
	b := "sha256:" + strings.Repeat("b", 64)
	missing := "sha256:" + strings.Repeat("c", 64)
	base := &AbsentPinError{Ref: "repo@" + missing, Repo: "repo", Digest: missing}

	none := base.Error()
	if !strings.Contains(none, missing) || !strings.Contains(none, "rebuild") {
		t.Fatalf("none case must name the missing digest and say rebuild: %q", none)
	}

	one := (&AbsentPinError{Repo: "repo", Digest: missing, Available: []string{a}}).Error()
	if !strings.Contains(one, a) || !strings.Contains(one, "re-pin") {
		t.Fatalf("one case must name the current digest and say re-pin: %q", one)
	}

	many := (&AbsentPinError{Repo: "repo", Digest: missing, Available: []string{a, b}}).Error()
	if !strings.Contains(many, a) || !strings.Contains(many, b) {
		t.Fatalf("many case must name every present digest: %q", many)
	}
}
