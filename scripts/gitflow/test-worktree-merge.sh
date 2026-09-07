#!/usr/bin/env bash
# scripts/test-worktree-merge.sh — hermetic multi-"agent" dry-run for the
# worktree-merge.sh capstone (chain worktree-multi-agent-orchestration-support
# T8). Builds throwaway git repos in /tmp whose linked worktrees stand in for
# parallel subagents, then drives the full spawn → conflict-check → merge-back
# → cleanup cycle and asserts it completes with ZERO manual intervention:
# core.bare is auto-reset, disjoint work merges, the conflict surface is
# detected (overlapping files + duplicate migration numbers), and the
# worktrees/branches are reaped.
#
# Hermetic and repeatable: it never touches this repo, its main branch, or any
# remote. Exit 0 on all-pass, 1 on any failure.

set -uo pipefail

# A hermetic harness must not be able to reach the real repo, and inherited git
# env is the one channel that lets it. Inside a pre-commit hook git exports
# GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE, and the worktree gate-only commit path
# runs `git -c …`, which exports GIT_CONFIG_PARAMETERS/GIT_CONFIG_COUNT (bugs
# 921/937 — precommit.sh scrubs the same five for corpos-gate). Every git call
# below would then operate on THIS checkout instead of its throwaway /tmp repo.
# Observed 2026-08-11 the first time this harness was wired into the gate: the
# scenario-8 landing pushed the branch under test to the live origin.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT

# Prove the scrub worked before any scenario runs. If git still resolves a repo
# from a neutral directory, something in the environment still points at a real
# checkout — and these scenarios drive a script that pushes to origin. Refusing
# is the only safe response; the leak's first symptom was a live push.
_probe="$(mktemp -d)"
if _leak="$(git -C "$_probe" rev-parse --show-toplevel 2>/dev/null)"; then
    rm -rf "$_probe"
    echo "test-worktree-merge: REFUSING to run — git resolves '$_leak' from a neutral" >&2
    echo "  directory, so the inherited git environment is not clean and these" >&2
    echo "  scenarios would drive worktree-merge.sh against a real checkout." >&2
    exit 1
fi
rm -rf "$_probe"

GITFLOW_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELPER="$GITFLOW_DIR/worktree-merge.sh"
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

# new_repo creates a throwaway repo with a seed commit + the migrations dir,
# and prints its path.
new_repo() {
    local d; d="$(mktemp -d)"; SCRATCH+=("$d")
    git -C "$d" init -q -b main
    git -C "$d" config user.email t@example.com
    git -C "$d" config user.name test
    git -C "$d" config commit.gpgsign false
    mkdir -p "$d/go/internal/db/migrations"
    echo seed > "$d/README.md"
    git -C "$d" add -A
    git -C "$d" commit -q -m seed
    echo "$d"
}

# spawn_agent adds a worktree+branch and commits the given file::content pairs,
# simulating a parallel subagent's disjoint work.
spawn_agent() {
    local d="$1" b="$2"; shift 2
    git -C "$d" worktree add -q "$d/.wt/$b" -b "$b" >/dev/null 2>&1
    local fc f c
    for fc in "$@"; do
        f="${fc%%::*}"; c="${fc#*::}"
        mkdir -p "$(dirname "$d/.wt/$b/$f")"
        printf '%s\n' "$c" > "$d/.wt/$b/$f"
    done
    git -C "$d/.wt/$b" add -A
    git -C "$d/.wt/$b" commit -q -m "$b work"
}

echo "── Scenario 1: clean disjoint work (+ distinct migration numbers) ──"
d1="$(new_repo)"
spawn_agent "$d1" agent-a "featureA.txt::A" "go/internal/db/migrations/080_alpha.sql::-- a"
spawn_agent "$d1" agent-b "featureB.txt::B" "go/internal/db/migrations/081_beta.sql::-- b"
# Simulate the Agent-tool's core.bare flip on the shared config.
git -C "$d1" config core.bare true
rc=0
( cd "$d1" && "$HELPER" --no-gate agent-a agent-b ) > "$d1/out.log" 2>&1 || rc=$?
cat "$d1/out.log" | sed 's/^/    │ /'
assert "clean: helper exits 0" test "$rc" -eq 0
assert "clean: core.bare reset to false" test "$(git -C "$d1" config core.bare)" = "false"
assert "clean: featureA merged onto main" test -f "$d1/featureA.txt"
assert "clean: featureB merged onto main" test -f "$d1/featureB.txt"
assert "clean: 080 migration merged" test -f "$d1/go/internal/db/migrations/080_alpha.sql"
assert "clean: 081 migration merged" test -f "$d1/go/internal/db/migrations/081_beta.sql"
assert "clean: branch agent-a reaped" bash -c "! git -C '$d1' rev-parse --verify agent-a >/dev/null 2>&1"
assert "clean: branch agent-b reaped" bash -c "! git -C '$d1' rev-parse --verify agent-b >/dev/null 2>&1"
assert "clean: worktree dirs reaped" bash -c "! test -d '$d1/.wt/agent-a' && ! test -d '$d1/.wt/agent-b'"

echo "── Scenario 2: duplicate migration number is flagged when a glob is configured ──"
d2="$(new_repo)"
# The migration-collision check runs ONLY when GITFLOW_MIGRATIONS_GLOB is set.
printf 'GITFLOW_MIGRATIONS_GLOB="go/internal/db/migrations/*.sql"\n' > "$d2/.gitflow"
spawn_agent "$d2" agent-a "x.txt::A" "go/internal/db/migrations/080_alpha.sql::-- a"
spawn_agent "$d2" agent-b "y.txt::B" "go/internal/db/migrations/080_beta.sql::-- b"
rc=0
( cd "$d2" && "$HELPER" --check-only agent-a agent-b ) > "$d2/out.log" 2>&1 || rc=$?
cat "$d2/out.log" | sed 's/^/    │ /'
assert "dup-migration: --check-only exits 1" test "$rc" -eq 1
assert "dup-migration: report names number 080" grep -q "number 080" "$d2/out.log"
assert "dup-migration: nothing merged (x.txt absent on main)" bash -c "! test -f '$d2/x.txt'"

echo "── Scenario 2b: with NO migrations glob (default), a dup is NOT flagged ──"
# The generified default is an empty glob — a repo with no numbered migrations
# skips the check entirely. Same duplicate as 2, but no .gitflow, so the merge
# proceeds instead of aborting.
d2b="$(new_repo)"
spawn_agent "$d2b" agent-a "x.txt::A" "go/internal/db/migrations/080_alpha.sql::-- a"
spawn_agent "$d2b" agent-b "y.txt::B" "go/internal/db/migrations/080_beta.sql::-- b"
rc=0
( cd "$d2b" && "$HELPER" --check-only agent-a agent-b ) > "$d2b/out.log" 2>&1 || rc=$?
cat "$d2b/out.log" | sed 's/^/    │ /'
assert "no-glob: --check-only exits 0 (dup not checked)" test "$rc" -eq 0
assert "no-glob: did not run the migration check" bash -c "! grep -q 'duplicate migration' '$d2b/out.log'"

echo "── Scenario 3: overlapping file is flagged by --check-only ──"
d3="$(new_repo)"
spawn_agent "$d3" agent-a "shared.txt::from A"
spawn_agent "$d3" agent-b "shared.txt::from B"
rc=0
( cd "$d3" && "$HELPER" --check-only agent-a agent-b ) > "$d3/out.log" 2>&1 || rc=$?
cat "$d3/out.log" | sed 's/^/    │ /'
assert "overlap: --check-only exits 1" test "$rc" -eq 1
assert "overlap: report names shared.txt" grep -q "shared.txt" "$d3/out.log"

# ── Orphan landing-PR pre-flight (bug
# worktree-merge-automerge-failure-leaves-orphan-pr-that-nothing-detects) ──
#
# start_pr_stub writes a JSON array of open PRs and serves it from a throwaway
# HTTP server on an ephemeral port, standing in for the gitea pulls endpoint.
# Prints the base URL. The server answers EVERY path with the same payload,
# which is all the pre-flight needs.
STUBS=()
stub_cleanup() { for p in "${STUBS[@]:-}"; do [[ -n "$p" ]] && kill "$p" 2>/dev/null; done; }
trap 'stub_cleanup; cleanup' EXIT

start_pr_stub() {
    local payload="$1" d; d="$(mktemp -d)"; SCRATCH+=("$d")
    printf '%s' "$payload" > "$d/pulls.json"
    python3 - "$d" >/dev/null 2>&1 <<'PY' &
import http.server, socketserver, sys, threading
d = sys.argv[1]
body = open(d + "/pulls.json", "rb").read()
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *a): pass
srv = socketserver.TCPServer(("127.0.0.1", 0), H)
open(d + "/port", "w").write(str(srv.server_address[1]))
srv.serve_forever()
PY
    # NOTE the >/dev/null 2>&1 on the backgrounded python above: without it the
    # child inherits this function's stdout, and `$(start_pr_stub …)` blocks
    # forever waiting for that fd to close.
    STUBS+=("$!")
    local i
    for i in $(seq 1 50); do [[ -s "$d/port" ]] && break; sleep 0.1; done
    [[ -s "$d/port" ]] || { echo "stub failed to start" >&2; return 1; }
    echo "http://127.0.0.1:$(cat "$d/port")"
}

# repo_with_origin makes a repo whose origin LOOKS like gitea over https, so
# resolve_gitea_remote engages; the API base + token are then overridden to the
# stub. No network is touched.
repo_with_origin() {
    local d; d="$(new_repo)"
    git -C "$d" remote add origin https://gitea.invalid/git/owner/repo.git
    echo "$d"
}

echo "── Scenario 4: an open land/* PR blocks a new landing ──"
BASE="$(start_pr_stub '[{"number":21,"title":"land: docs: Run-61","head":{"ref":"land/main-20260805T015319"}}]')"
d4="$(repo_with_origin)"
spawn_agent "$d4" agent-a "featureA.txt::A"
rc=0
( cd "$d4" && WORKTREE_MERGE_API_BASE="$BASE" WORKTREE_MERGE_TOKEN=stub \
    "$HELPER" --check-only agent-a ) > "$d4/out.log" 2>&1 || rc=$?
cat "$d4/out.log" | sed 's/^/    │ /'
assert "orphan-pr: exits 1" test "$rc" -eq 1
assert "orphan-pr: names the open PR number" grep -q '#21' "$d4/out.log"
assert "orphan-pr: names the land branch" grep -q 'land/main-20260805T015319' "$d4/out.log"
assert "orphan-pr: explains the sweep hazard" grep -qi 'sweep' "$d4/out.log"

echo "── Scenario 5: --allow-open-land-prs overrides the pre-flight ──"
d5="$(repo_with_origin)"
spawn_agent "$d5" agent-a "featureA.txt::A"
rc=0
( cd "$d5" && WORKTREE_MERGE_API_BASE="$BASE" WORKTREE_MERGE_TOKEN=stub \
    "$HELPER" --check-only --allow-open-land-prs agent-a ) > "$d5/out.log" 2>&1 || rc=$?
cat "$d5/out.log" | sed 's/^/    │ /'
assert "override: exits 0" test "$rc" -eq 0
assert "override: no refusal printed" bash -c "! grep -q 'outstanding on origin' '$d5/out.log'"

echo "── Scenario 6: a non-land/* open PR does NOT block ──"
BASE2="$(start_pr_stub '[{"number":7,"title":"feat: unrelated","head":{"ref":"feature/xyz"}}]')"
d6="$(repo_with_origin)"
spawn_agent "$d6" agent-a "featureA.txt::A"
rc=0
( cd "$d6" && WORKTREE_MERGE_API_BASE="$BASE2" WORKTREE_MERGE_TOKEN=stub \
    "$HELPER" --check-only agent-a ) > "$d6/out.log" 2>&1 || rc=$?
cat "$d6/out.log" | sed 's/^/    │ /'
assert "non-land PR: exits 0" test "$rc" -eq 0
assert "non-land PR: no refusal printed" bash -c "! grep -q 'outstanding on origin' '$d6/out.log'"

echo "── Scenario 7: unreachable gitea degrades to a warning, not a block ──"
d7="$(repo_with_origin)"
spawn_agent "$d7" agent-a "featureA.txt::A"
rc=0
# Port 1 is reserved and never listening — curl fails fast.
( cd "$d7" && WORKTREE_MERGE_API_BASE="http://127.0.0.1:1" WORKTREE_MERGE_TOKEN=stub \
    "$HELPER" --check-only agent-a ) > "$d7/out.log" 2>&1 || rc=$?
cat "$d7/out.log" | sed 's/^/    │ /'
assert "unreachable: exits 0 (does not block local work)" test "$rc" -eq 0
assert "unreachable: warns that the pre-flight was skipped" grep -q 'pre-flight skipped' "$d7/out.log"

echo "── Scenario 8: the post-land hook runs after a landing, and gets --deploy ──"
# The generified script runs GITFLOW_POST_LAND_HOOK after a successful landing.
# The hook records its args; the toolkit's deploy-gap logic now lives in such a
# hook (scripts/gitflow-deploy.sh, tested by scripts/test-gitflow-deploy.sh).
d8="$(new_repo)"
mkdir -p "$d8/scripts"
cat > "$d8/scripts/hook.sh" <<'HOOK'
#!/usr/bin/env bash
echo "HOOK RAN args=[$*]"
printf '%s\n' "$*" > "$(git rev-parse --show-toplevel)/hook_args"
HOOK
chmod +x "$d8/scripts/hook.sh"
printf 'GITFLOW_POST_LAND_HOOK="scripts/hook.sh"\n' > "$d8/.gitflow"
git -C "$d8" add -A; git -C "$d8" commit -q -m "scripts: post-land hook"
spawn_agent "$d8" agent-a "featureA.txt::A"
rc=0
( cd "$d8" && "$HELPER" --no-gate --deploy agent-a ) > "$d8/out.log" 2>&1 || rc=$?
cat "$d8/out.log" | sed 's/^/    │ /'
assert "post-land-hook: exits 0" test "$rc" -eq 0
assert "post-land-hook: the hook ran" grep -q "HOOK RAN" "$d8/out.log"
assert "post-land-hook: --deploy was passed through" bash -c "grep -q -- '--deploy' '$d8/hook_args'"
assert "post-land-hook: worktree-merge names the hook" grep -q "post-land hook" "$d8/out.log"

echo "── Scenario 8b: no --deploy → the hook runs WITHOUT --deploy ──"
d8b="$(new_repo)"
mkdir -p "$d8b/scripts"
cp "$d8/scripts/hook.sh" "$d8b/scripts/hook.sh"; chmod +x "$d8b/scripts/hook.sh"
printf 'GITFLOW_POST_LAND_HOOK="scripts/hook.sh"\n' > "$d8b/.gitflow"
git -C "$d8b" add -A; git -C "$d8b" commit -q -m "scripts: post-land hook"
spawn_agent "$d8b" agent-a "featureA.txt::A"
rc=0
( cd "$d8b" && "$HELPER" --no-gate agent-a ) > "$d8b/out.log" 2>&1 || rc=$?
cat "$d8b/out.log" | sed 's/^/    │ /'
assert "no-deploy: exits 0" test "$rc" -eq 0
assert "no-deploy: the hook ran" grep -q "HOOK RAN" "$d8b/out.log"
assert "no-deploy: hook got no --deploy" bash -c "! grep -q -- '--deploy' '$d8b/hook_args'"

echo "── Scenario 8c: a failing post-land hook fails the run ──"
d8c="$(new_repo)"
mkdir -p "$d8c/scripts"
printf '#!/usr/bin/env bash\necho boom >&2\nexit 3\n' > "$d8c/scripts/hook.sh"
chmod +x "$d8c/scripts/hook.sh"
printf 'GITFLOW_POST_LAND_HOOK="scripts/hook.sh"\n' > "$d8c/.gitflow"
git -C "$d8c" add -A; git -C "$d8c" commit -q -m "scripts: failing hook"
spawn_agent "$d8c" agent-a "featureA.txt::A"
rc=0
( cd "$d8c" && "$HELPER" --no-gate agent-a ) > "$d8c/out.log" 2>&1 || rc=$?
cat "$d8c/out.log" | sed 's/^/    │ /'
assert "hook-fail: exits nonzero" test "$rc" -ne 0

echo "── Scenario 9: no post-land hook configured → nothing runs ──"
d9="$(new_repo)"
spawn_agent "$d9" agent-a "featureA.txt::A"
rc=0
( cd "$d9" && "$HELPER" --no-gate agent-a ) > "$d9/out.log" 2>&1 || rc=$?
cat "$d9/out.log" | sed 's/^/    │ /'
assert "no-hook: exits 0" test "$rc" -eq 0
assert "no-hook: no post-land hook line" bash -c "! grep -q 'post-land hook' '$d9/out.log'"
assert "no-hook: featureA still merged" test -f "$d9/featureA.txt"

echo "── Scenario 10: landing a change to the LANDING PATH itself (bug 1235) ──"
# Hazard 1, self-rewrite mid-run: bash reads a script incrementally by byte
# offset. When the merge rewrites worktree-merge.sh on disk underneath the
# running interpreter, a length change before the current offset can make
# bash resume mid-line and execute garbage. The fix re-execs from a temp
# snapshot, so the merge can rewrite the repo copy freely.
#
# The repo under test carries its own copies of the landing-path scripts and
# the run is driven through THAT copy — the same relationship the real repo
# has to its own scripts/. The branch then replaces it with something of a
# very different length.
#
# NOTE on what this proves. Whether the raw corruption fires on any given
# run depends on bash's read buffering and on where the merge falls relative
# to the interpreter's byte offset — this scenario does NOT reliably fail
# against the unfixed script, and a "landing completes" assertion alone
# would be vacuous. So the load-bearing assertion is the MECHANISM: the run
# executes from a snapshot outside every tree it can touch, which removes
# the dependency instead of betting on it. The outcome assertions below
# guard against the snapshot breaking the ordinary landing.
d10="$(new_repo)"
mkdir -p "$d10/scripts"
# All landing-path siblings must travel together — worktree-merge re-execs from
# a private snapshot of exactly these, and sources gitflow-common at load.
for _s in worktree-merge.sh gitea-pr-land.sh gitea-resolve-env.sh gitflow-common.sh; do
    cp "$GITFLOW_DIR/$_s" "$d10/scripts/$_s"
done
chmod +x "$d10/scripts/"*.sh
git -C "$d10" add -A
git -C "$d10" commit -q -m "scripts: landing path"
spawn_agent "$d10" agent-a "scripts/worktree-merge.sh::#!/usr/bin/env bash
echo 'a much shorter worktree-merge.sh'"
rc=0
( cd "$d10" && "$d10/scripts/worktree-merge.sh" --no-gate agent-a ) > "$d10/out.log" 2>&1 || rc=$?
cat "$d10/out.log" | sed 's/^/    │ /'
assert "self-rewrite: runs from a snapshot, not the repo copy it is merging into" \
    grep -q "running from a private snapshot" "$d10/out.log"
assert "self-rewrite: the snapshot lives outside the repo under test" \
    bash -c "! grep -q 'running from a private snapshot ($d10' '$d10/out.log'"
assert "self-rewrite: landing completes despite rewriting itself" test "$rc" -eq 0
assert "self-rewrite: reached the final line" grep -q "worktree-merge: done" "$d10/out.log"
assert "self-rewrite: the branch's version landed on main" \
    grep -q "a much shorter worktree-merge.sh" "$d10/scripts/worktree-merge.sh"
assert "self-rewrite: no bash syntax error from a shifted read offset" \
    bash -c "! grep -qiE 'syntax error|unexpected (end of file|token)' '$d10/out.log'"
# Scoped to THIS run's snapshot, read back from the path the run itself printed.
# Globbing ${TMPDIR:-/tmp}/worktree-merge-snapshot.* asserted on the SHARED /tmp
# namespace, so it saw every snapshot on the machine rather than only its own.
# gitea-pr-land waits up to GITEA_CI_TIMEOUT (default 1800s) for CI, so any
# landing anywhere held a snapshot for up to half an hour and failed this
# assertion — and with it the whole gate — for an unrelated commit that touched
# nothing near the landing path. Observed 2026-08-22 on a one-markdown-file
# commit (bug 1289). Deleting the foreign directory is not an option: it belongs
# to a live merge. The leak this guards against is real, so the check stays; only
# its scope narrows.
snapshot10="$(sed -n 's/.*running from a private snapshot (\([^)]*\)).*/\1/p' "$d10/out.log" | head -1)"
assert "self-rewrite: the run named the snapshot it used" test -n "$snapshot10"
assert "self-rewrite: the snapshot dir is cleaned up" \
    bash -c "! test -e \"\$1\"" _ "$snapshot10"

echo "── Scenario 11: refuses to run from inside a linked worktree (bug 1277) ──"
# Bug 1277: running from inside a linked worktree made the feature branch the
# landing target. The pre-flight now refuses before the gate.
d11="$(new_repo)"
spawn_agent "$d11" agent-a "featureA.txt::A"
rc=0
( cd "$d11/.wt/agent-a" && "$HELPER" --no-gate agent-a ) > "$d11/out.log" 2>&1 || rc=$?
cat "$d11/out.log" | sed 's/^/    │ /'
assert "linked-worktree: exits nonzero" test "$rc" -ne 0
assert "linked-worktree: says LINKED WORKTREE" grep -q "LINKED WORKTREE" "$d11/out.log"
assert "linked-worktree: names the main checkout to run from" grep -q "cd $d11 " "$d11/out.log"
assert "linked-worktree: nothing merged" bash -c "! test -f '$d11/featureA.txt'"

echo "── Scenario 12: refuses when the current branch is not main (bug 1277) ──"
d12="$(new_repo)"
git -C "$d12" checkout -b not-main 2>/dev/null
spawn_agent "$d12" agent-a "featureA.txt::A"
rc=0
( cd "$d12" && "$HELPER" --no-gate agent-a ) > "$d12/out.log" 2>&1 || rc=$?
cat "$d12/out.log" | sed 's/^/    │ /'
assert "wrong-branch: exits nonzero" test "$rc" -ne 0
assert "wrong-branch: names the branch" grep -q "not-main" "$d12/out.log"
assert "wrong-branch: nothing merged" bash -c "! test -f '$d12/featureA.txt'"

echo "── Scenario 13: a configured landing branch other than 'main' is honoured ──"
# .gitflow sets GITFLOW_LANDING_BRANCH=trunk; the repo's integration branch is
# trunk. worktree-merge must accept trunk as the target and merge onto it.
d13="$(new_repo)"
git -C "$d13" branch -m main trunk
printf 'GITFLOW_LANDING_BRANCH=trunk\n' > "$d13/.gitflow"
git -C "$d13" add -A; git -C "$d13" commit -q -m "config: landing branch trunk"
spawn_agent "$d13" agent-a "featureA.txt::A"
rc=0
( cd "$d13" && "$HELPER" --no-gate agent-a ) > "$d13/out.log" 2>&1 || rc=$?
cat "$d13/out.log" | sed 's/^/    │ /'
assert "landing-branch: exits 0" test "$rc" -eq 0
assert "landing-branch: integration target is trunk" grep -q "integration target = trunk" "$d13/out.log"
assert "landing-branch: featureA merged onto trunk" test -f "$d13/featureA.txt"

echo "── Scenario 13b: on 'main' with landing branch 'trunk' is refused ──"
# Same config, but the checkout sits on main — the branch guard must refuse and
# name the configured landing branch.
d13b="$(new_repo)"
printf 'GITFLOW_LANDING_BRANCH=trunk\n' > "$d13b/.gitflow"
git -C "$d13b" add -A; git -C "$d13b" commit -q -m "config: landing branch trunk"
spawn_agent "$d13b" agent-a "featureA.txt::A"
rc=0
( cd "$d13b" && "$HELPER" --no-gate agent-a ) > "$d13b/out.log" 2>&1 || rc=$?
cat "$d13b/out.log" | sed 's/^/    │ /'
assert "wrong-config-branch: exits nonzero" test "$rc" -ne 0
assert "wrong-config-branch: names 'trunk'" grep -q "trunk" "$d13b/out.log"
assert "wrong-config-branch: nothing merged" bash -c "! test -f '$d13b/featureA.txt'"

echo "── Scenario 14: local landing branch BEHIND origin is reconciled (bug 1297) ──"
# A bare origin advances past local main; worktree-merge must fetch + merge
# origin/main BEFORE the branch merge, so the later push is not rejected
# non-fast-forward. No gitea (path origin) — lands by direct push.
bare14="$(mktemp -d)"; SCRATCH+=("$bare14")
git init -q --bare -b main "$bare14"
d14="$(new_repo)"
git -C "$d14" remote add origin "$bare14"
git -C "$d14" push -q origin main
# A second clone advances origin/main with a non-conflicting commit.
other14="$(mktemp -d)"; SCRATCH+=("$other14")
git clone -q "$bare14" "$other14"
git -C "$other14" config user.email t@example.com; git -C "$other14" config user.name test
echo "from-origin" > "$other14/origin-only.txt"
git -C "$other14" add -A; git -C "$other14" commit -q -m "origin advances"
git -C "$other14" push -q origin main
# Now d14's local main is 1 behind origin/main. Add disjoint worktree work.
spawn_agent "$d14" agent-a "featureA.txt::A"
rc=0
( cd "$d14" && "$HELPER" --no-gate agent-a ) > "$d14/out.log" 2>&1 || rc=$?
cat "$d14/out.log" | sed 's/^/    │ /'
assert "non-ff: exits 0" test "$rc" -eq 0
assert "non-ff: reconciled with origin" grep -qi "reconcil" "$d14/out.log"
assert "non-ff: pushed to origin" grep -q "pushed main to origin" "$d14/out.log"
# origin now carries BOTH the origin-only commit and the agent's file.
git -C "$bare14" cat-file -e "$(git -C "$d14" rev-parse origin/main):origin-only.txt" 2>/dev/null
assert "non-ff: origin keeps its own commit" test "$?" -eq 0
assert "non-ff: origin got the agent work" bash -c "git -C '$d14' cat-file -e origin/main:featureA.txt 2>/dev/null"

echo "── Scenario 15: push-protected main lands via the gitea PR handshake (bug 1203) ──"
# The whole protected-main glue driven end-to-end through worktree-merge: a real
# git push to main is rejected 'protected branch' by the bare origin's
# pre-receive; land/* pushes are allowed; a local stub serves the gitea API. This
# is the seam test-gitea-pr-land.sh cannot reach (it drives gitea-pr-land alone).
bare15="$(mktemp -d)"; SCRATCH+=("$bare15")
git init -q --bare -b main "$bare15"
cat > "$bare15/hooks/pre-receive" <<'HOOK'
#!/usr/bin/env bash
while read -r _old _new ref; do
    if [ "$ref" = "refs/heads/main" ]; then
        echo "remote: pre-receive hook declined: protected branch" >&2
        exit 1
    fi
done
exit 0
HOOK
chmod +x "$bare15/hooks/pre-receive"
# Combined gitea API stub: list-pulls -> [], create-pull -> {number:42},
# branch protection -> no required checks (merge immediately), merge -> 200.
sd15="$(mktemp -d)"; SCRATCH+=("$sd15")
cat > "$sd15/stub.py" <<'PY'
import http.server, socketserver, json
class H(http.server.BaseHTTPRequestHandler):
    def reply(self, code, body):
        self.send_response(code); self.send_header("Content-Type","application/json")
        self.send_header("Content-Length", str(len(body))); self.end_headers(); self.wfile.write(body)
    def do_GET(self):
        p=self.path
        if "/branch_protections/" in p:
            return self.reply(200, b'{"branch_name":"main","enable_status_check":false,"status_check_contexts":[]}')
        if "/pulls" in p:
            return self.reply(200, b'[]')  # no orphan land PRs
        if "/status" in p:
            return self.reply(200, b'{"state":"success"}')
        return self.reply(200, b'[]')
    def do_POST(self):
        if self.path.endswith("/merge"):
            return self.reply(200, b'{}')
        if "/pulls" in self.path:
            return self.reply(201, b'{"number":42}')
        return self.reply(200, b'{}')
    def log_message(self,*a): pass
srv=socketserver.TCPServer(("127.0.0.1",0),H)
open("PORTFILE","w").write(str(srv.server_address[1])); srv.serve_forever()
PY
sed -i "s#PORTFILE#$sd15/port#" "$sd15/stub.py"
python3 "$sd15/stub.py" >/dev/null 2>&1 < /dev/null &
STUBS+=("$!")
for _ in $(seq 1 50); do [[ -s "$sd15/port" ]] && break; sleep 0.1; done
B15="http://127.0.0.1:$(cat "$sd15/port")"
d15="$(new_repo)"
git -C "$d15" remote add origin "$bare15"
# origin/main never gets seeded — the pre-receive rejects every main push, which
# is exactly the protected-main condition. worktree-merge tolerates a missing
# origin/main (its ahead/behind and reconcile guards skip) and reaches the PR
# landing path when the direct push is refused.
spawn_agent "$d15" agent-a "featureA.txt::A"
rc=0
( cd "$d15" && GITEA_API="$B15" GITEA_OWNER=owner GITEA_REPO=repo GITEA_TOKEN=stub \
    GITEA_CI_TIMEOUT=5 GITEA_CI_INTERVAL=1 "$HELPER" --no-gate agent-a ) > "$d15/out.log" 2>&1 || rc=$?
cat "$d15/out.log" | sed 's/^/    │ /'
assert "protected: detected the protection rejection" grep -q "push-protected" "$d15/out.log"
assert "protected: landed via a PR" grep -qi "landing via PR" "$d15/out.log"
assert "protected: the PR was merged through the handshake" grep -q "PR #42 merged" "$d15/out.log"

echo ""
echo "dry-run: $PASS pass, $FAIL fail"
[[ "$FAIL" -eq 0 ]]
