#!/usr/bin/env bash
# gitflow/test-worktree-reap.sh — hermetic regression net for
# gitflow/worktree-reap.sh, focused on the bug-1350 data-loss guard.
#
# A worktree whose branch ref is still at the base main SHA counts as "merged"
# by `git branch --merged` (it is not ahead of main). But it may hold staged or
# uncommitted work with a commit still in flight. Reap must NOT delete such a
# worktree; it must skip it and leave the work intact. A clean merged worktree
# is still reaped as before.
#
# Hermetic isolation. This harness drives worktree-reap.sh, which deletes
# worktrees and branches. When it runs at pre-push inside a real landing, the
# parent has exported its git env (GIT_DIR / GIT_WORK_TREE / GIT_INDEX_FILE —
# bugs 921/937), so an unscrubbed run would operate on the REAL checkout: reap
# the real worktrees and write config to the real repo (observed 2026-09-24: a
# `git config user.email` in a scenario leaked onto the real repo and tripped
# the repo-identity gate). Re-exec ONCE in a clean env with a minimal whitelist,
# then prove git resolves no repo from a neutral dir before any scenario runs.
set -uo pipefail

if [[ -z "${_WTR_HERMETIC:-}" ]]; then
    _self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    exec env -i HOME="${HOME:-/tmp}" PATH="$PATH" TMPDIR="${TMPDIR:-/tmp}" TERM="${TERM:-dumb}" \
        _WTR_HERMETIC=1 bash "$_self" "$@"
fi

_probe="$(mktemp -d)"
if _leak="$(git -C "$_probe" rev-parse --show-toplevel 2>/dev/null)"; then
    rm -rf "$_probe"
    echo "test-worktree-reap: REFUSING to run — git resolves '$_leak' from a neutral" >&2
    echo "  directory, so the inherited git environment is not clean and these" >&2
    echo "  scenarios would drive worktree-reap.sh against a real checkout." >&2
    exit 1
fi
rm -rf "$_probe"

GITFLOW_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REAP="$GITFLOW_DIR/worktree-reap.sh"
PASS=0
FAIL=0
SCRATCH=()

cleanup() {
    for d in "${SCRATCH[@]:-}"; do [[ -n "$d" && -d "$d" ]] && rm -rf "$d"; done
}
trap cleanup EXIT

assert() {
    local desc="$1"; shift
    if "$@"; then echo "  PASS  $desc"; PASS=$((PASS + 1));
    else echo "  FAIL  $desc"; FAIL=$((FAIL + 1)); fi
}

# new_env prints a parent dir; the repo is at <dir>/repo, on main with one commit.
new_env() {
    local d; d="$(mktemp -d)"; SCRATCH+=("$d")
    git init -q -b main "$d/repo"
    git -C "$d/repo" config user.email t@t
    git -C "$d/repo" config user.name t
    echo base > "$d/repo/base.txt"
    git -C "$d/repo" add -A
    git -C "$d/repo" commit -q -m init
    echo "$d"
}

run_reap() {
    local repo="$1"
    ( cd "$repo" && env GITFLOW_LANDING_BRANCH=main "$REAP" ) > "$repo/reap.log" 2>&1
}

echo "── Scenario 1: a clean merged worktree is reaped ──"
d1="$(new_env)"; r1="$d1/repo"
git -C "$r1" worktree add -q "$d1/wt-clean" -b clean main
run_reap "$r1"
sed 's/^/    │ /' "$r1/reap.log"
assert "clean: worktree removed" test ! -d "$d1/wt-clean"
assert "clean: branch deleted" bash -c "! git -C '$r1' show-ref --verify -q refs/heads/clean"

echo "── Scenario 2: a worktree with staged changes is NOT reaped (bug 1350) ──"
d2="$(new_env)"; r2="$d2/repo"
git -C "$r2" worktree add -q "$d2/wt-dirty" -b dirty main
echo change > "$d2/wt-dirty/newfile.txt"
git -C "$d2/wt-dirty" add newfile.txt   # staged, commit still in flight
run_reap "$r2"
sed 's/^/    │ /' "$r2/reap.log"
assert "dirty: worktree survived" test -d "$d2/wt-dirty"
assert "dirty: staged work intact" test -f "$d2/wt-dirty/newfile.txt"
assert "dirty: branch survived" git -C "$r2" show-ref --verify -q refs/heads/dirty
assert "dirty: reap reported skipping" grep -q "skipping branch dirty" "$r2/reap.log"

echo "── Scenario 3: a worktree with an in-flight commit (index.lock) is NOT reaped ──"
d3="$(new_env)"; r3="$d3/repo"
git -C "$r3" worktree add -q "$d3/wt-lock" -b lockbr main
lockdir="$(git -C "$d3/wt-lock" rev-parse --absolute-git-dir)"
: > "$lockdir/index.lock"
run_reap "$r3"
sed 's/^/    │ /' "$r3/reap.log"
assert "lock: worktree survived" test -d "$d3/wt-lock"
assert "lock: reap reported commit in progress" grep -q "commit in progress" "$r3/reap.log"
rm -f "$lockdir/index.lock"

echo "── Scenario 4: a merged worktree holding gitignored private-subtree data is NOT reaped ──"
d4="$(new_env)"; r4="$d4/repo"
printf 'corpus/private/\n' > "$r4/.gitignore"
printf 'private-subtree corpus/private/\n' > "$r4/.canonical-home"
git -C "$r4" add .gitignore .canonical-home
git -C "$r4" commit -q -m "declare private subtree"
git -C "$r4" worktree add -q "$d4/wt-priv" -b privbr main
mkdir -p "$d4/wt-priv/corpus/private"
echo "the only copy" > "$d4/wt-priv/corpus/private/run.json"   # gitignored -> status clean
run_reap "$r4"
sed 's/^/    │ /' "$r4/reap.log"
assert "private: worktree survived" test -d "$d4/wt-priv"
assert "private: the private file is intact" test -f "$d4/wt-priv/corpus/private/run.json"
assert "private: branch survived" git -C "$r4" show-ref --verify -q refs/heads/privbr
assert "private: reap reported refusing to strand" grep -q "strand private data" "$r4/reap.log"

echo "── Scenario 5: a stale UNMERGED worktree draws a warning naming its chain ──"
d5="$(new_env)"; r5="$d5/repo"
git -C "$r5" worktree add -q "$d5/wt-stale" -b stalebr main
echo work > "$d5/wt-stale/feature.txt"
git -C "$d5/wt-stale" add feature.txt
# An absolute ISO date well in the past (git's relative "N days ago" parser is
# unavailable under the scrubbed env). This is ~2 months old vs the default
# 14-day threshold, so it reads as stale regardless of the run date.
GIT_AUTHOR_DATE="2026-07-01T12:00:00" GIT_COMMITTER_DATE="2026-07-01T12:00:00" \
    git -C "$d5/wt-stale" commit -q -m "unmerged work"    # tip != main -> unmerged
# Record the branch↔chain link in the ledger the reap will read.
printf 'stalebr\tinjection-assay\t%s\tactive\n' "$(date +%s)" > "$r5/.git/gitflow-chain-branches"
run_reap "$r5"
sed 's/^/    │ /' "$r5/reap.log"
assert "stale: worktree survived (never reaped)" test -d "$d5/wt-stale"
assert "stale: branch survived" git -C "$r5" show-ref --verify -q refs/heads/stalebr
assert "stale: reap WARNED about unmerged work" grep -q "WARN.*stalebr.*UNMERGED" "$r5/reap.log"
assert "stale: warning named the owning chain" grep -q "chain injection-assay" "$r5/reap.log"

echo ""
echo "worktree-reap: $PASS pass, $FAIL fail"
[[ "$FAIL" -eq 0 ]]
