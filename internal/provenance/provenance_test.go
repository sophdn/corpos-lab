package provenance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
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

// gitCommitTime returns the committer time of the last commit touching path,
// read straight from git, so a test can compare LastChange against ground truth
// without pinning a wall clock.
func gitCommitTime(t *testing.T, dir, path string) time.Time {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "log", "-1", "--format=%ct", "--", ":/"+path)
	env := []string{}
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	var secs int64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &secs); err != nil {
		t.Fatalf("parse %q: %v", out, err)
	}
	return time.Unix(secs, 0).UTC()
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

func TestLastChangeReturnsCommitTimeOfTouchingPath(t *testing.T) {
	dir := tempRepo(t) // init commit touches f.txt only
	sub := filepath.Join(dir, "internal", "assay")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.go"), []byte("package assay\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "assay change")

	got, err := LastChange(context.Background(), dir, "internal/assay", "internal/runner")
	if err != nil {
		t.Fatalf("LastChange: %v", err)
	}
	// Compare against the commit time git itself reports for that path, so the
	// test asserts LastChange reads the right commit, not a pinned clock.
	want := gitCommitTime(t, dir, "internal/assay")
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	// A path no commit ever touched yields a zero time and no error — the
	// caller reads that as "cannot tell", not as a failure.
	zero, err := LastChange(context.Background(), dir, "internal/never")
	if err != nil {
		t.Fatalf("LastChange(untouched): %v", err)
	}
	if !zero.IsZero() {
		t.Fatalf("expected zero time for an untouched path, got %v", zero)
	}
}

func TestLastChangeResolvesFromSubdir(t *testing.T) {
	dir := tempRepo(t)
	sub := filepath.Join(dir, "studies", "x")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	// f.txt exists at the repo root from the init commit; a path matched from a
	// subdir must still resolve against the repo root via the :/ magic pathspec.
	got, err := LastChange(context.Background(), sub, "f.txt")
	if err != nil {
		t.Fatalf("LastChange from subdir: %v", err)
	}
	if got.IsZero() {
		t.Fatal("expected the init commit time, got zero")
	}
}

func TestLastChangeFailsOutsideGit(t *testing.T) {
	if _, err := LastChange(context.Background(), t.TempDir(), "internal/assay"); err == nil {
		t.Fatal("expected error outside a git repo")
	}
}
