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
	"strings"
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
