// Package provenance captures the git identity of the tree a run executed
// from. The Rust source stamped commit SHA + dirty flag at build time via
// build.rs; here they are captured at runtime, which is fresher — a
// worktree-based run stamps the tree it actually used, not the tree the
// binary was compiled from.
//
// Called from control.RunStudy via the Provenance seam, and recorded on every
// StudyRun. That is worth stating because it was not true until 2026-07-14:
// this package sat here fully written and fully tested with ZERO callers,
// while CLAUDE.md asserted that every result row carried the repo commit. The
// tests passed, the coverage counted, and the invariant was false. If you are
// about to remove the last caller, remove the claim in the same commit.
package provenance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Stamp identifies the repo state a run executed from.
type Stamp struct {
	// CommitSHA is the full HEAD commit hash.
	CommitSHA string `json:"commit_sha"`
	// Dirty is true when the working tree had uncommitted changes.
	Dirty bool `json:"dirty"`
}

// Capture reads the git identity of repoDir. It fails loudly when repoDir is
// not a git repository — an unstampable run must not silently pass as stamped.
// Loudly to the CALLER, which records the failure and proceeds: a run that
// cannot say which tree it came from is worse than one that can, but it is not
// void, and refusing it would be the freeze reflex.
func Capture(ctx context.Context, repoDir string) (Stamp, error) {
	sha, err := gitOutput(ctx, repoDir, "rev-parse", "HEAD")
	if err != nil {
		return Stamp{}, fmt.Errorf("provenance: rev-parse HEAD in %s: %w", repoDir, err)
	}

	status, err := gitOutput(ctx, repoDir, "status", "--porcelain")
	if err != nil {
		return Stamp{}, fmt.Errorf("provenance: status in %s: %w", repoDir, err)
	}

	return Stamp{
		CommitSHA: sha,
		Dirty:     status != "",
	}, nil
}

// LastChange returns the commit time of the most recent commit that touched any
// of paths, where paths are repo-root-relative (e.g. "internal/assay"). repoDir
// may be any directory inside the repo: the paths are matched against the repo
// root via git's top-level magic pathspec, so a caller passing a study
// subdirectory still resolves them correctly. It fails when repoDir is not a git
// repo; it returns a zero time and no error when no commit touches paths, so a
// caller treats "cannot tell" as "do not warn" rather than as a failure.
func LastChange(ctx context.Context, repoDir string, paths ...string) (time.Time, error) {
	args := []string{"log", "-1", "--format=%ct", "--"}
	for _, p := range paths {
		args = append(args, ":/"+p)
	}
	out, err := gitOutput(ctx, repoDir, args...)
	if err != nil {
		return time.Time{}, fmt.Errorf("provenance: git log for %v in %s: %w", paths, repoDir, err)
	}
	if out == "" {
		return time.Time{}, nil
	}
	secs, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("provenance: unparseable commit time %q: %w", out, err)
	}
	return time.Unix(secs, 0).UTC(), nil
}

// BinaryStaleness reports whether a running binary predates the source it was
// built from. buildTime is the commit time the binary embeds (its vcs.time
// build setting); lastSourceChange is the commit time of the newest commit that
// touched the source that matters. It returns an advisory message and true when
// the binary is behind. A zero buildTime or a zero lastSourceChange means
// "cannot tell" and returns no warning — observe, don't assert. It never blocks
// a run; the caller records the message as a preflight warning.
//
// This turns the failure mode from suggestion 169 — an old host binary rejecting
// a study's condition with a bare "unknown condition" error — into a plain
// binary-is-stale signal, surfaced as itself.
func BinaryStaleness(buildTime, lastSourceChange time.Time) (string, bool) {
	if buildTime.IsZero() || lastSourceChange.IsZero() {
		return "", false
	}
	if buildTime.Before(lastSourceChange) {
		return fmt.Sprintf("host corpos-lab binary was built from source dated %s, but the "+
			"source changed at %s — rebuild it (go build ./cmd/corpos-lab); a validation error "+
			"like \"unknown condition\" may be a stale binary, not a bad study",
			buildTime.UTC().Format(time.RFC3339), lastSourceChange.UTC().Format(time.RFC3339)), true
	}
	return "", false
}

func gitOutput(ctx context.Context, repoDir string, args ...string) (string, error) {
	full := append([]string{"-C", repoDir}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	cmd.Env = scrubbedGitEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// scrubbedGitEnv strips inherited GIT_* variables so the stamp always
// describes repoDir. A caller running inside a git hook inherits GIT_DIR /
// GIT_INDEX_FILE pointing at the hook's repo, which would silently override
// -C and stamp the WRONG repo — surfaced when this package's own tests ran
// under the pre-commit gate.
func scrubbedGitEnv() []string {
	env := os.Environ()
	kept := env[:0]
	for _, kv := range env {
		if !strings.HasPrefix(kv, "GIT_") {
			kept = append(kept, kv)
		}
	}
	return kept
}
