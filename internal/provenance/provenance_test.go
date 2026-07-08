package provenance

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// Scrub inherited GIT_* (the pre-commit hook exports GIT_DIR/
	// GIT_INDEX_FILE, which would override -C), then add author identity.
	env := []string{
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test",
	}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func tempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func TestCaptureCleanRepo(t *testing.T) {
	dir := tempRepo(t)
	stamp, err := Capture(context.Background(), dir)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(stamp.CommitSHA) {
		t.Fatalf("CommitSHA = %q, want 40-hex", stamp.CommitSHA)
	}
	if stamp.Dirty {
		t.Fatal("fresh commit should not be dirty")
	}
}

func TestCaptureDirtyRepo(t *testing.T) {
	dir := tempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp, err := Capture(context.Background(), dir)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !stamp.Dirty {
		t.Fatal("modified tree should be dirty")
	}
}

func TestCaptureUntrackedFileCountsAsDirty(t *testing.T) {
	dir := tempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp, err := Capture(context.Background(), dir)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if !stamp.Dirty {
		t.Fatal("untracked file should count as dirty (porcelain parity)")
	}
}

func TestCaptureFailsLoudlyOutsideGit(t *testing.T) {
	if _, err := Capture(context.Background(), t.TempDir()); err == nil {
		t.Fatal("expected error outside a git repo")
	}
}

func TestCaptureFailsOnMissingDir(t *testing.T) {
	if _, err := Capture(context.Background(), filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("expected error for missing dir")
	}
}
