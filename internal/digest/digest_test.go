package digest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBytesKnownVector(t *testing.T) {
	// SHA-256("abc") is the NIST FIPS 180-2 test vector.
	got := Bytes([]byte("abc"))
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("Bytes(abc) = %s, want %s", got, want)
	}
}

func TestFileMatchesBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.md")
	content := []byte("# scenario v1\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := File(path)
	if err != nil {
		t.Fatalf("File: %v", err)
	}
	if want := Bytes(content); got != want {
		t.Fatalf("File = %s, want %s", got, want)
	}
}

func TestFileMissingNamesPath(t *testing.T) {
	_, err := File(filepath.Join(t.TempDir(), "absent.md"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "absent.md") {
		t.Fatalf("error should name the path: %v", err)
	}
}
