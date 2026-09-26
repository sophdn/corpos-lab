#!/usr/bin/env bash
# gitflow/test-chain-reconcile.sh — hermetic regression net for
# gitflow/chain-reconcile.sh.
#
# Asserts the disposition contract: a merged branch is resolved, an unmerged
# undisposed branch blocks (exit 1), --archive parks it (ledger marker + branch
# kept), --discard deletes branch + worktree, and a chain with no branches is a
# clean exit 0.
#
# Hermetic isolation. The script creates worktrees and branches; at pre-commit the
# parent's git env leaks in (bugs 921/937), so an unscrubbed run would drive the
# real repo. Re-exec ONCE in a clean env, then prove git resolves no repo from a
# neutral dir before any scenario runs.
set -uo pipefail

if [[ -z "${_CR_HERMETIC:-}" ]]; then
    _self="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/$(basename "${BASH_SOURCE[0]}")"
    exec env -i HOME="${HOME:-/tmp}" PATH="$PATH" TMPDIR="${TMPDIR:-/tmp}" TERM="${TERM:-dumb}" \
        _CR_HERMETIC=1 bash "$_self" "$@"
fi

_probe="$(mktemp -d)"
if _leak="$(git -C "$_probe" rev-parse --show-toplevel 2>/dev/null)"; then
    rm -rf "$_probe"
    echo "test-chain-reconcile: REFUSING to run — git resolves '$_leak' from a neutral dir." >&2
    exit 1
fi
rm -rf "$_probe"

GITFLOW_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RECONCILE="$GITFLOW_DIR/chain-reconcile.sh"
PASS=0
FAIL=0
SCRATCH=()
cleanup() { for d in "${SCRATCH[@]:-}"; do [[ -n "$d" && -d "$d" ]] && rm -rf "$d"; done; }
trap cleanup EXIT
assert() { local d="$1"; shift; if "$@"; then echo "  PASS  $d"; PASS=$((PASS+1)); else echo "  FAIL  $d"; FAIL=$((FAIL+1)); fi; }

# new_repo prints a repo path on main with one commit.
new_repo() {
    local d; d="$(mktemp -d)"; SCRATCH+=("$d")
    git init -q -b main "$d/repo"
    git -C "$d/repo" config user.email t@t
    git -C "$d/repo" config user.name t
    echo base > "$d/repo/base.txt"; git -C "$d/repo" add -A; git -C "$d/repo" commit -q -m init
    echo "$d/repo"
}
ledger_of() { echo "$1/.git/gitflow-chain-branches"; }
record() { printf '%s\t%s\t%s\tactive\n' "$2" "$3" "$(date +%s)" >> "$(ledger_of "$1")"; }
# run <repo> <args...> -> echoes exit code
run() { local repo="$1"; shift; local rc=0; ( cd "$repo" && env GITFLOW_LANDING_BRANCH=main "$RECONCILE" "$@" ) > "$repo/rec.log" 2>&1 || rc=$?; echo "$rc"; }

# 1. A merged branch (tip == main) is resolved -> exit 0.
r="$(new_repo)"; git -C "$r" branch merged1; record "$r" merged1 chainA
rc="$(run "$r" chainA)"
assert "merged branch -> exit 0" test "$rc" = 0
assert "merged branch reported resolved" grep -q "merged1 — merged" "$r/rec.log"

# 2. An unmerged, undisposed branch blocks -> exit 1, named.
r="$(new_repo)"; git -C "$r" worktree add -q "$r/../wt-feat1" -b feat1 main
echo x > "$r/../wt-feat1/f.txt"; git -C "$r/../wt-feat1" add f.txt; git -C "$r/../wt-feat1" commit -q -m w
record "$r" feat1 chainB
rc="$(run "$r" chainB)"
assert "unmerged undisposed -> exit 1" test "$rc" = 1
assert "unmerged branch named" grep -q "feat1 — UNMERGED" "$r/rec.log"

# 3. --archive parks the branch -> exit 0, ledger marks it archived, branch kept.
rc="$(run "$r" chainB --archive feat1 --note 'parked: only copy of the run')"
assert "archive -> exit 0" test "$rc" = 0
assert "archive kept the branch" git -C "$r" show-ref --verify -q refs/heads/feat1
assert "ledger marks feat1 archived" grep -qP "^feat1\tchainB\t[0-9]+\tarchived" "$(ledger_of "$r")"

# 4. --discard deletes the branch and its worktree -> exit 0.
r2="$(new_repo)"; git -C "$r2" worktree add -q "$r2/../wt-feat2" -b feat2 main
echo x > "$r2/../wt-feat2/f.txt"; git -C "$r2/../wt-feat2" add f.txt; git -C "$r2/../wt-feat2" commit -q -m w
record "$r2" feat2 chainC
rc="$(run "$r2" chainC --discard feat2)"
assert "discard -> exit 0" test "$rc" = 0
assert "discard deleted the branch" bash -c "! git -C '$r2' show-ref --verify -q refs/heads/feat2"
assert "discard removed the worktree" test ! -d "$r2/../wt-feat2"

# 5. A chain with no recorded branches is a clean exit 0.
r3="$(new_repo)"; record "$r3" other otherchain
rc="$(run "$r3" chainEmpty)"
assert "chain with no branches -> exit 0" test "$rc" = 0

echo ""
echo "chain-reconcile: $PASS pass, $FAIL fail"
[[ "$FAIL" -eq 0 ]]
