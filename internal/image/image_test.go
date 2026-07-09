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
