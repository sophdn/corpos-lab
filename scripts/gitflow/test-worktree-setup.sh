#!/usr/bin/env bash
# gitflow/test-worktree-setup.sh — hermetic test for the gate-only pre-commit
# hook that worktree-setup.sh writes into a linked worktree. Asserts the hook
# runs the COMMIT gate (GITFLOW_COMMIT_GATE_CMD — the fast tier for a tiered
# repo), falls back to GITFLOW_GATE_CMD when no commit gate is set, and is a
# no-op when no commit gate is configured. (chain 475 T13.)
#
# Hermetic and repeatable: throwaway git repos in /tmp, no network, no touch of
# this repo. It scrubs the inherited env ONCE via env -i so an ambient GITFLOW_*
# value cannot leak into a fixture's config resolution. Exit 0 on all-pass.

set -uo pipefail

if [[ -z "${_WTS_HERMETIC:-}" ]]; then
    _self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    exec env -i HOME="${HOME:-/tmp}" PATH="$PATH" TMPDIR="${TMPDIR:-/tmp}" TERM="${TERM:-dumb}" \
        _WTS_HERMETIC=1 bash "$_self" "$@"
fi

GITFLOW_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SETUP="$GITFLOW_DIR/worktree-setup.sh"
PASS=0
FAIL=0
SCRATCH=()

cleanup() { for d in "${SCRATCH[@]:-}"; do [[ -n "$d" && -d "$d" ]] && rm -rf "$d"; done; }
trap cleanup EXIT

assert() {
    local desc="$1"; shift
    if "$@"; then echo "  PASS  $desc"; PASS=$((PASS + 1));
    else echo "  FAIL  $desc"; FAIL=$((FAIL + 1)); fi
}

# make_wt <gitflow-contents> — build a repo + one linked worktree, commit the
# given .gitflow (empty arg = no .gitflow), run worktree-setup INSIDE the
# worktree, and print the path to the generated gate-only pre-commit hook.
make_wt() {
    local gitflow="$1"
    local d; d="$(mktemp -d)"; SCRATCH+=("$d")
    git -C "$d" init -q -b main
    git -C "$d" config user.email t@example.com
    git -C "$d" config user.name test
    git -C "$d" config commit.gpgsign false
    echo seed > "$d/README.md"
    [[ -n "$gitflow" ]] && printf '%s\n' "$gitflow" > "$d/.gitflow"
    git -C "$d" add -A
    git -C "$d" commit -q -m seed
    git -C "$d" worktree add -q "$d/.wt/b" -b b >/dev/null 2>&1
    ( cd "$d/.wt/b" && bash "$SETUP" ) >/dev/null 2>&1
    echo "$(git -C "$d/.wt/b" rev-parse --absolute-git-dir)/gate-only-hooks/pre-commit"
}

echo "── Scenario 1: the commit hook runs the fast tier, not the merge gate ──"
hook="$(make_wt 'GITFLOW_GATE_CMD="FULL_MERGE_GATE"
GITFLOW_COMMIT_GATE_CMD="FAST_COMMIT_GATE"')"
assert "commit-tier: hook written" test -f "$hook"
assert "commit-tier: runs the fast commit gate" grep -q "FAST_COMMIT_GATE" "$hook"
assert "commit-tier: does NOT run the full merge gate" bash -c "! grep -q 'FULL_MERGE_GATE' '$hook'"

echo "── Scenario 2: no commit gate set → falls back to the merge gate ──"
hook="$(make_wt 'GITFLOW_GATE_CMD="ONLY_GATE"')"
assert "fallback: runs GITFLOW_GATE_CMD" grep -q "ONLY_GATE" "$hook"

echo "── Scenario 3: explicit empty commit gate → no-op hook ──"
hook="$(make_wt 'GITFLOW_GATE_CMD="SOME_GATE"
GITFLOW_COMMIT_GATE_CMD=""')"
assert "empty-commit-gate: hook written" test -f "$hook"
assert "empty-commit-gate: runs no gate" bash -c "! grep -q 'SOME_GATE' '$hook'"
assert "empty-commit-gate: is a no-op (exit 0)" grep -q "exit 0" "$hook"

echo "── Scenario 4: repo configures no gate at all → no-op hook ──"
hook="$(make_wt '')"
assert "no-gate-repo: is a no-op (exit 0)" grep -q "exit 0" "$hook"

echo ""
echo "worktree-setup: $PASS pass, $FAIL fail"
[[ "$FAIL" -eq 0 ]]
