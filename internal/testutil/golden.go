// Package testutil holds helpers shared by the module's tests. Only _test.go
// files import it, so strict deadcode (where tests are not roots) exempts it by
// path in scripts/gate.sh and gate.yml.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// Golden compares got with testdata/<name>, byte for byte. With
// UPDATE_GOLDEN=1 it first rewrites the file from got. On a mismatch it fails
// with "<what> drifted from <path>:" and the new output.
func Golden(t testing.TB, name, what, got string) {
	t.Helper()
	golden := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("%s drifted from %s:\n%s", what, golden, got)
	}
}
