#!/usr/bin/env bash
# scripts/worktree-merge.sh — the canonical multi-agent merge-back helper.
#
# CAPSTONE of chain worktree-multi-agent-orchestration-support (T8). Codifies
# the spawn → conflict-check → merge-back → cleanup ritual so a multi-subagent
# run integrates with ZERO manual core.bare / migration / registry babysitting:
#
#   1. Reset core.bare (the Agent-tool's isolation=worktree flips it to true on
#      the shared .git/config; left flipped, the main checkout can't merge —
#      "this operation must be run in a work tree"). Idempotent.
#   2. Pre-merge CONFLICT-SURFACE CHECK across the named branches: files touched
#      by 2+ branches, and duplicate migration NNN numbers. (T3 makes migration
#      collisions improbable and T7 made the registry guards conflict-free, so
#      this is a fast safety net, not the primary defense.)
#   3. Ordered MERGE into the current checkout: fast-forward to the first branch
#      when possible, --no-ff the rest. Aborts cleanly on a real conflict.
#   4. Build/test GATE on the merged result (the configured GITFLOW_GATE_CMD,
#      auto-detected to scripts/precommit.sh, scripts/gate.sh, the Go build+test,
#      or none).
#   5. REAP the merged worktrees + branches (handles the Agent-tool's locks).
#
# Usage:
#   scripts/worktree-merge.sh [options] <branch|worktree-path>...
#
# Options:
#   --check-only    Run steps 1-2 (core.bare reset + conflict surface) and stop.
#                   The "dry-run" pre-spawn/pre-merge check. Exit 1 if conflicts.
#   --no-gate       Skip the build/test gate (step 4).
#   --no-reap       Merge but leave the worktrees/branches in place (step 5 off).
#   --gate-cmd CMD  Override the gate command (default: the configured
#                   GITFLOW_GATE_CMD).
#   --deploy        Pass --deploy to the repo's post-land hook
#                   (GITFLOW_POST_LAND_HOOK), so a repo that ships a deployed
#                   artifact can rebuild it. Without --deploy the hook typically
#                   only REPORTS the gap (landing is not deploying). A repo with
#                   no post-land hook ignores this flag.
#   --allow-open-land-prs
#                   Skip the orphan landing-PR pre-flight. The pre-flight
#                   refuses to start a landing while an earlier land/* PR is
#                   still open on origin, because landing again would sweep the
#                   stranded PR's commits in under a new PR and conceal it.
#                   Override only when you know the open PR is unrelated.
#
# Merges into whatever the current checkout has checked out — run it from the
# integration target (normally the main checkout), then LANDS the result on
# origin (direct push, or branch+PR via the gitea API when main is
# push-protected — bug 1203). It never
# touches a remote.

set -euo pipefail

# ── Step 0: run from a private snapshot of the landing path ────────────
# Bug 1235. Landing a change to the landing path itself has two hazards,
# and neither announces itself:
#
#   1. SELF-REWRITE MID-RUN. bash does not slurp a script; it reads it
#      incrementally, tracking a byte offset into the open file. This
#      script runs from the main checkout, merges a branch into main, and
#      that merge REWRITES this file on disk underneath the running
#      interpreter. If the new file differs in length before the current
#      offset, bash resumes at the wrong byte and executes garbage — a
#      syntax error, or half a line from the middle of the file.
#   2. SELF-REAP. Step 5 deletes the merged worktrees. Run from inside one
#      of them (the natural move when you want the fixed version's
#      behavior), the script deletes its own directory mid-run and any
#      later sibling-script invocation through SCRIPT_DIR fails.
#
# Both were previously avoided only by an undocumented trick: staging a
# copy of the two landing-path scripts in a scratch dir and driving the
# landing from there. This makes the trick the default, for every run.
#
# We copy this script plus its landing-path sibling into a temp dir and
# re-exec from there, so the text bash is reading and the helper it
# resolves through SCRIPT_DIR both live outside every tree the run can
# touch. Cost is two file copies of a few KB — the ordinary landing is not
# measurably slower and behaves identically. WORKTREE_MERGE_SNAPSHOT
# marks the re-executed pass so it happens exactly once.
LANDING_PATH_SCRIPTS=(worktree-merge.sh gitea-pr-land.sh gitea-resolve-env.sh gitflow-common.sh)

if [[ -z "${WORKTREE_MERGE_SNAPSHOT:-}" ]]; then
    _src_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    _snapshot="$(mktemp -d "${TMPDIR:-/tmp}/worktree-merge-snapshot.XXXXXX")"
    for _f in "${LANDING_PATH_SCRIPTS[@]}"; do
        if [[ -f "$_src_dir/$_f" ]]; then
            cp "$_src_dir/$_f" "$_snapshot/$_f"
            chmod +x "$_snapshot/$_f"
        fi
    done
    # Exported so the re-executed pass knows where it came from (for
    # messages) and can clean the snapshot up on exit.
    export WORKTREE_MERGE_SNAPSHOT="$_snapshot"
    export WORKTREE_MERGE_SOURCE_DIR="$_src_dir"
    exec "$_snapshot/worktree-merge.sh" "$@"
fi
# From here on we ARE the snapshot copy. Drop it on the way out, however
# the run ends — the snapshot is disposable and the source of truth is
# still $WORKTREE_MERGE_SOURCE_DIR.
trap '[[ "${WORKTREE_MERGE_SNAPSHOT:-}" == */worktree-merge-snapshot.* ]] && rm -rf "$WORKTREE_MERGE_SNAPSHOT"' EXIT

# One line, so an operator reading the log knows why the repo's copy of this
# script can change underneath the run without consequence — and so the
# regression net has something to assert on. Whether the raw corruption
# would fire on any given run depends on bash's read buffering and on where
# the merge lands relative to the interpreter's byte offset; the snapshot
# removes the dependency rather than betting on it.
echo "worktree-merge: running from a private snapshot ($WORKTREE_MERGE_SNAPSHOT) of ${WORKTREE_MERGE_SOURCE_DIR:-the landing path} — bug 1235."

CHECK_ONLY=0
DO_GATE=1
DO_REAP=1
GATE_CMD=""
ALLOW_OPEN_LAND_PRS=0
DO_DEPLOY=0
BRANCHES=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --check-only) CHECK_ONLY=1 ;;
        --no-gate)    DO_GATE=0 ;;
        --no-reap)    DO_REAP=0 ;;
        --gate-cmd)   shift; GATE_CMD="${1:-}" ;;
        --allow-open-land-prs) ALLOW_OPEN_LAND_PRS=1 ;;
        --deploy)     DO_DEPLOY=1 ;;
        --*)          echo "worktree-merge: unknown option: $1" >&2; exit 2 ;;
        *)            BRANCHES+=("$1") ;;
    esac
    shift
done

if [[ ${#BRANCHES[@]} -eq 0 ]]; then
    echo "worktree-merge: name at least one branch or worktree path to merge" >&2
    echo "  usage: scripts/worktree-merge.sh [--check-only|--no-gate|--no-reap] <branch|path>..." >&2
    exit 2
fi

# ── Step 1: core.bare auto-reset (must run FIRST) ──────────────────────
# The Agent tool's isolation=worktree flips core.bare=true on the shared
# config; left flipped, `git rev-parse --show-toplevel` (and the merge) fail
# with "this operation must be run in a work tree". `git config` is plumbing
# that works regardless, so reset it BEFORE any work-tree-requiring command.
# Idempotent.
if [[ "$(git config core.bare 2>/dev/null || echo false)" == "true" ]]; then
    git config core.bare false
    echo "worktree-merge: reset core.bare true → false (Agent-tool flip)"
fi

# Resolved from BASH_SOURCE, not REPO_ROOT: the sibling helpers must come from
# the same scripts/ dir as the worktree-merge.sh actually running, whichever
# checkout that is.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# ── git-flow service config ────────────────────────────────────────────
# Read <repo>/.gitflow (or the bare defaults) so the three repo-specific bits —
# the landing branch, the gate command, the migration-collision glob, and the
# post-land hook — come from config, not hardcoded literals. gitflow-common.sh
# is a landing-path sibling, so the snapshot copy is sourced when running from a
# snapshot.
# shellcheck source=gitflow-common.sh
. "$SCRIPT_DIR/gitflow-common.sh"
gitflow_load_config "$REPO_ROOT"
LANDING_BRANCH="$GITFLOW_LANDING_BRANCH"

target_branch="$(git branch --show-current 2>/dev/null || echo 'DETACHED')"
echo "worktree-merge: integration target = $target_branch ($REPO_ROOT)"

# ── Pre-flight: refuse to run from a linked worktree or the wrong branch ──
# Bug 1277. The header says "run from the MAIN checkout (on main)" but nothing
# enforced it; running from inside a linked worktree silently made the feature
# branch the landing target, pushed the wrong thing, then deleted its own cwd.
# This guard makes the comment the law.
_git_dir="$(git rev-parse --git-dir 2>/dev/null)"
_git_common="$(git rev-parse --git-common-dir 2>/dev/null)"
if [[ "$_git_dir" != "$_git_common" ]]; then
    _main_checkout="$(cd "$_git_common/.." && pwd)"
    echo "worktree-merge: ✖ this is a LINKED WORKTREE, not the main checkout." >&2
    echo "  Run from the main checkout instead:" >&2
    echo "    cd $_main_checkout && scripts/worktree-merge.sh $*" >&2
    exit 1
fi
if [[ "$target_branch" != "$LANDING_BRANCH" ]]; then
    echo "worktree-merge: ✖ current branch is '$target_branch', not '$LANDING_BRANCH'." >&2
    echo "  worktree-merge.sh must run from the main checkout on the '$LANDING_BRANCH' branch." >&2
    echo "  Switch to '$LANDING_BRANCH' first, or run from a checkout that is on it." >&2
    exit 1
fi

# Divergence guard (bug 1203): unpushed commits on local main are exactly what
# a later `git reset --hard origin/main` silently destroys (observed 2026-07-22,
# five commits). Say so up front; the landing step below ships them together
# with this merge.
if git rev-parse --verify -q "origin/$target_branch" >/dev/null 2>&1; then
    ahead="$(git rev-list --count "origin/$target_branch..$target_branch" 2>/dev/null || echo 0)"
    if [[ "$ahead" -gt 0 ]]; then
        echo "worktree-merge: ⚠ local $target_branch is AHEAD of origin/$target_branch by $ahead unpushed commit(s)."
        echo "  They land together with this merge. NEVER 'git reset --hard origin/$target_branch' while ahead — that destroys them."
    fi
fi

# ── gitea remote resolution ────────────────────────────────────────────
# Shared by the orphan-PR pre-flight below and the PR landing path further
# down. Sets GITEA_API/OWNER/REPO/TOKEN and returns 0, or returns 1 when this
# checkout has no https gitea origin or no token for it. WORKTREE_MERGE_API_BASE
# and WORKTREE_MERGE_TOKEN override the derived values (used by
# scripts/test-worktree-merge.sh to point at a local stub; not for daily use).
#
# The resolver lives in gitea-resolve-env.sh (bug 1301) and is shared with
# gitea-pr-land.sh. SCRIPT_DIR resolves to the snapshot dir when running from
# a snapshot, so the copied resolver is used.
# Preserve a fully pre-set gitea env (all four) as an intentional override; a
# real session sets none of these, so this is inert there and only the resolver
# derives them. It lets a caller point the whole handshake at a stub (the
# regression net's protected-main path) without an https origin. A partial set
# is cleared so a stray single var cannot half-configure the handshake.
if [[ -n "${GITEA_API:-}" && -n "${GITEA_OWNER:-}" && -n "${GITEA_REPO:-}" && -n "${GITEA_TOKEN:-}" ]]; then
    : # keep the caller's full override
else
    GITEA_API=""; GITEA_OWNER=""; GITEA_REPO=""; GITEA_TOKEN=""
fi
resolve_gitea_remote() {
    # shellcheck source=gitea-resolve-env.sh
    . "$SCRIPT_DIR/gitea-resolve-env.sh" 2>/dev/null || return 1
    return 0
}

# ── Orphan landing-PR pre-flight ───────────────────────────────────────
# A previous run can open the landing PR and then fail to merge it — the API
# merge returns non-200, or the session dies inside the bounded CI poll. That
# leaves an open land/* PR, an origin/land/* branch, and local main still
# AHEAD.
#
# The state is SELF-CONCEALING, which is why it needs a pre-flight rather than
# a note in the docs: because the commit stays on local main, the NEXT landing
# branches its land/* from that same still-ahead main and sweeps the orphaned
# commit to origin inside an unrelated PR. The content lands, `git log
# origin/main` shows it, and every later session concludes the landing
# succeeded while the stranded PR sits open unnoticed. Observed live on corpos
# PR #21 (open 4 days, swept in by PR #22). So "check git log on main" CANNOT
# detect this — only "no open land/* PR on origin" can.
#
# Refuse to start a new landing while one is outstanding. Network problems
# degrade to a warning: an unreachable gitea must not block local merge work.
check_orphan_land_prs() {
    if ! resolve_gitea_remote; then
        return 0  # no gitea origin / no token — nothing to check against
    fi
    local resp orphans
    resp="$(curl -sk -m 15 -H "Authorization: token $GITEA_TOKEN" \
        "$GITEA_API/repos/$GITEA_OWNER/$GITEA_REPO/pulls?state=open&limit=50" 2>/dev/null || true)"
    if [[ -z "$resp" ]]; then
        echo "worktree-merge: ⚠ orphan-PR pre-flight skipped — gitea unreachable." >&2
        return 0
    fi
    orphans="$(printf '%s' "$resp" \
        | jq -r 'if type=="array" then (.[] | select((.head.ref // "") | startswith("land/")) | "  #\(.number)  \(.head.ref)  \(.title)") else empty end' 2>/dev/null || true)"
    [[ -n "$orphans" ]] || return 0
    {
        echo "worktree-merge: ✖ open land/* PR(s) already outstanding on origin:"
        printf '%s\n' "$orphans"
        echo ""
        echo "  A previous landing did not complete. Landing again NOW would branch from a"
        echo "  local $target_branch that still carries those commits and sweep them to origin"
        echo "  inside this PR — hiding the stranded one permanently."
        echo ""
        echo "  Resolve first: merge the PR in the gitea UI (or close it if its commits are"
        echo "  already on origin/$target_branch — check with"
        echo "  'git merge-base --is-ancestor <pr-head> origin/$target_branch'), then delete"
        echo "  the stale land/* branch. Re-run with --allow-open-land-prs to override."
    } >&2
    return 1
}

if [[ "$ALLOW_OPEN_LAND_PRS" -eq 0 ]]; then
    check_orphan_land_prs || exit 1
fi

# ── Fetch origin and reconcile local main (bug 1297) ──────────────────────
# After an operator merges an orphan land/* PR in the gitea UI, origin/main has
# a merge commit that local main does not. Without this reconcile the gate runs
# to completion and the push is rejected non-fast-forward — wasting several
# minutes of gate work. Fetch and merge BEFORE any gate or branch merge.
if git remote get-url origin >/dev/null 2>&1; then
    git fetch origin 2>/dev/null || true
    if git rev-parse --verify -q "origin/$target_branch" >/dev/null 2>&1; then
        behind="$(git rev-list --count "$target_branch..origin/$target_branch" 2>/dev/null || echo 0)"
        if [[ "$behind" -gt 0 ]]; then
            echo "worktree-merge: local $target_branch is behind origin/$target_branch by $behind commit(s) — reconciling."
            if ! git merge "origin/$target_branch" --no-edit; then
                echo "worktree-merge: ✖ could not reconcile local $target_branch with origin/$target_branch." >&2
                echo "  Resolve the conflict manually, then re-run." >&2
                exit 1
            fi
            echo "worktree-merge: reconciled local $target_branch with origin/$target_branch."
        fi
    fi
fi

# Resolve each argument to a branch name. A path that is a registered worktree
# resolves to that worktree's checked-out branch; anything else is treated as a
# branch name directly. Records the worktree path (when known) for reaping.
declare -a MERGE_BRANCHES=()
declare -A WORKTREE_OF=()
for arg in "${BRANCHES[@]}"; do
    if [[ -d "$arg" ]]; then
        b="$(git -C "$arg" branch --show-current 2>/dev/null || true)"
        if [[ -z "$b" ]]; then
            echo "worktree-merge: $arg is not on a branch (detached?) — skipping" >&2
            continue
        fi
        MERGE_BRANCHES+=("$b")
        WORKTREE_OF["$b"]="$(cd "$arg" && pwd)"
    else
        MERGE_BRANCHES+=("$arg")
        # Discover the worktree hosting this branch (for reaping), if any.
        wt="$(git worktree list --porcelain | awk -v b="refs/heads/$arg" '
            /^worktree /{p=$2} /^branch /{ if($2==b) print p }')"
        [[ -n "$wt" ]] && WORKTREE_OF["$arg"]="$wt"
    fi
done

if [[ ${#MERGE_BRANCHES[@]} -eq 0 ]]; then
    echo "worktree-merge: no mergeable branches resolved" >&2
    exit 2
fi

# ── Step 2: pre-merge conflict-surface check ───────────────────────────
echo ""
echo "worktree-merge: conflict-surface check across ${#MERGE_BRANCHES[@]} branch(es)…"
conflicts=0

# (a) Files touched by 2+ branches (vs the merge base with the target).
tmp_files="$(mktemp)"
: > "$tmp_files"
for b in "${MERGE_BRANCHES[@]}"; do
    base="$(git merge-base HEAD "$b" 2>/dev/null || echo HEAD)"
    git diff --name-only "$base" "$b" | sed "s|\$|\t$b|"
done > "$tmp_files"
overlap="$(cut -f1 "$tmp_files" | sort | uniq -d || true)"
if [[ -n "$overlap" ]]; then
    echo "  ⚠ files modified by more than one branch (textual merge may conflict):"
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        branches_for_f="$(awk -F'\t' -v f="$f" '$1==f{printf " %s",$2}' "$tmp_files")"
        echo "      $f →$branches_for_f"
        conflicts=1
    done <<< "$overlap"
else
    echo "  ✓ no file touched by more than one branch"
fi
rm -f "$tmp_files"

# (b) Duplicate migration numbers across branches (NNN_ prefix collision).
# Only when the repo configures a migrations glob (GITFLOW_MIGRATIONS_GLOB in
# .gitflow); a repo with no numbered migrations skips this check entirely.
if [[ -n "$GITFLOW_MIGRATIONS_GLOB" ]]; then
    mig_tmp="$(mktemp)"
    : > "$mig_tmp"
    for b in "${MERGE_BRANCHES[@]}"; do
        base="$(git merge-base HEAD "$b" 2>/dev/null || echo HEAD)"
        git diff --name-only --diff-filter=A "$base" "$b" -- "$GITFLOW_MIGRATIONS_GLOB" \
          | sed -E 's|.*/([0-9]+)_.*\.sql$|\1|' | grep -E '^[0-9]+$' | sed "s|\$|\t$b|" || true
    done > "$mig_tmp"
    dup_mig="$(cut -f1 "$mig_tmp" | sort | uniq -d || true)"
    if [[ -n "$dup_mig" ]]; then
        echo "  ⚠ duplicate migration number(s) across branches — renumber before merge:"
        while IFS= read -r n; do
            [[ -z "$n" ]] && continue
            echo "      number $n added by:$(awk -F'\t' -v n="$n" '$1==n{printf " %s",$2}' "$mig_tmp")"
            conflicts=1
        done <<< "$dup_mig"
    else
        echo "  ✓ no duplicate migration numbers"
    fi
    rm -f "$mig_tmp"
fi

if [[ "$CHECK_ONLY" -eq 1 ]]; then
    echo ""
    if [[ "$conflicts" -eq 1 ]]; then
        echo "worktree-merge: --check-only → CONFLICTS PRESENT (resolve before merge-back)"
        exit 1
    fi
    echo "worktree-merge: --check-only → clean (safe to merge-back)"
    exit 0
fi

if [[ "$conflicts" -eq 1 ]]; then
    echo ""
    echo "worktree-merge: conflict surface non-empty — refusing to auto-merge." >&2
    echo "  Resolve the overlaps above (or renumber migrations) and re-run, or merge by hand." >&2
    exit 1
fi

# ── Step 3: ordered merge ──────────────────────────────────────────────
echo ""
first=1
for b in "${MERGE_BRANCHES[@]}"; do
    if [[ "$first" -eq 1 ]] && git merge --ff-only "$b" 2>/dev/null; then
        echo "worktree-merge: fast-forwarded to $b"
    else
        merge_out="$(git merge --no-ff -m "merge(worktree): integrate $b" "$b" 2>&1)" || {
            git merge --abort 2>/dev/null || true
            if grep -q "untracked working tree files would be overwritten" <<<"$merge_out"; then
                # Untracked-collision shape (bug 1294): the main checkout has
                # untracked files that the incoming branch also adds.
                {
                    echo "worktree-merge: UNTRACKED FILE COLLISION on $b"
                    echo ""
                    echo "  The main checkout has untracked files that the branch"
                    echo "  also carries. Git refuses to merge because it would"
                    echo "  overwrite them."
                    echo ""
                    echo "  The incoming branch's copies are the committed versions."
                    echo "  The untracked files in the main checkout are the ones"
                    echo "  to remove (after verifying they are duplicates or"
                    echo "  already backed up)."
                    echo ""
                    echo "  Convention: write harness results to a scratch dir,"
                    echo "  then copy them into the worktree that commits them."
                    echo "  Never write directly into the main checkout."
                    echo ""
                    echo "  Files named by git:"
                    grep "	" <<<"$merge_out" | head -20
                    echo ""
                    echo "  After removing the main-checkout copies:"
                    echo "    scripts/worktree-merge.sh $b"
                } >&2
            else
                echo "worktree-merge: MERGE CONFLICT on $b — aborting this merge." >&2
                echo "$merge_out" >&2
                echo "  Resolve manually; already-merged branches were kept." >&2
            fi
            exit 1
        }
        echo "worktree-merge: merged $b (--no-ff)"
    fi
    first=0
done

# ── Step 4: build/test gate ────────────────────────────────────────────
if [[ "$DO_GATE" -eq 1 ]]; then
    # --gate-cmd wins; otherwise the configured GITFLOW_GATE_CMD (auto-detected
    # to scripts/precommit.sh, scripts/gate.sh, the Go build+test, or none).
    eff_gate="${GATE_CMD:-$GITFLOW_GATE_CMD}"
    if [[ -n "$eff_gate" ]]; then
        echo ""
        echo "worktree-merge: gate → $eff_gate"
        bash -c "$eff_gate"
    else
        echo "worktree-merge: no gate command (GITFLOW_GATE_CMD empty, no --gate-cmd) — skipping gate."
    fi
else
    echo "worktree-merge: --no-gate — skipping build/test gate."
fi

# ── Land on origin (bug 1203) ────────────────────────────────────────────────
# main is push-protected on gitea for the shared/* repos, so a merge that stops
# at local main strands unpushable commits — and the recovery a reasonable
# agent reaches for (reset --hard origin/main) silently destroys them. Landing
# is therefore part of the merge: push directly when the remote allows it;
# otherwise ship the merged state as a branch, open + merge a PR via the gitea
# API, and re-sync local main only after verifying origin carries
# content-identical state.
if git remote get-url origin >/dev/null 2>&1; then
    echo ""
    echo "worktree-merge: landing $target_branch on origin…"
    push_err="$(mktemp)"
    if git push origin "$target_branch" 2>"$push_err"; then
        echo "worktree-merge: pushed $target_branch to origin"
        rm -f "$push_err"
    elif grep -qi 'protected branch' "$push_err"; then
        rm -f "$push_err"
        echo "worktree-merge: origin/$target_branch is push-protected — landing via PR"
        if ! resolve_gitea_remote; then
            echo "worktree-merge: no gitea token for origin in ~/.git-credentials." >&2
            echo "  Work is safe but UNPUSHED on local $target_branch — push it as a branch + open a PR by hand." >&2
            exit 1
        fi
        api="$GITEA_API"; owner="$GITEA_OWNER"; repo="$GITEA_REPO"; tok="$GITEA_TOKEN"
        land="land/${target_branch}-$(date +%Y%m%dT%H%M%S)"
        git push origin "$target_branch:refs/heads/$land"
        echo "worktree-merge: merged state pushed as origin/$land"
        title="land: $(git log -1 --format=%s "$target_branch")"
        pr_idx="$(curl -sk -X POST -H "Authorization: token $tok" -H 'Content-Type: application/json' \
            "$api/repos/$owner/$repo/pulls" \
            --data "$(jq -nc --arg h "$land" --arg t "$title" --arg b "$LANDING_BRANCH" '{head:$h, base:$b, title:$t}')" \
            | jq -r '.number // empty')"
        if [ -z "$pr_idx" ]; then
            echo "worktree-merge: PR creation failed — work is safe on origin/$land; open the PR by hand." >&2
            exit 1
        fi
        # Branch protection requires its status checks to pass before a merge
        # is accepted; merging earlier returns 405 "Please try again later"
        # (observed live on PR #17). scripts/gitea-pr-land.sh owns that
        # handshake — it waits on the contexts branch protection actually
        # REQUIRES (not the whole combined commit status, which any unrelated
        # queued run pins at pending — bug 1224) and retries a rejected merge
        # once. Its regression net is scripts/test-gitea-pr-land.sh.
        sha="$(git rev-parse "$target_branch")"
        echo "worktree-merge: landing PR #$pr_idx via the gitea CI handshake on ${sha:0:9}…"
        land_rc=0
        GITEA_API="$api" GITEA_OWNER="$owner" GITEA_REPO="$repo" GITEA_TOKEN="$tok" \
            "$SCRIPT_DIR/gitea-pr-land.sh" --pr "$pr_idx" --sha "$sha" --base "$LANDING_BRANCH" \
            || land_rc=$?
        if [ "$land_rc" -eq 0 ]; then
            git fetch origin
            if git diff --quiet "$target_branch" "origin/$target_branch"; then
                git reset --hard "origin/$target_branch"
                echo "worktree-merge: local $target_branch re-synced to origin (content verified identical first)"
            else
                echo "worktree-merge: PR merged but origin/$target_branch content differs (concurrent changes?)." >&2
                echo "  NOT resetting local $target_branch — reconcile manually; everything is on origin." >&2
            fi
            git push origin ":refs/heads/$land" 2>/dev/null || true
        else
            echo "worktree-merge: PR #$pr_idx created but auto-merge failed (see the gitea-pr-land output above)." >&2
            echo "" >&2
            echo "  Work is safe on origin/$land, and local $target_branch stays ahead until PR #$pr_idx merges." >&2
            echo "  MERGE OR CLOSE PR #$pr_idx BEFORE THE NEXT LANDING. Until you do, the next run's" >&2
            echo "  land branch would be cut from this same still-ahead $target_branch and would sweep" >&2
            echo "  these commits to origin inside an unrelated PR, hiding #$pr_idx permanently." >&2
            echo "  (The pre-flight at the top of this script now refuses that; don't --allow-open-land-prs past it.)" >&2
            # EXIT NON-ZERO. Falling through to the reap printed a final
            # "done — N branch(es) integrated" success line over a landing that
            # did not land, which is how corpos PR #21 went unnoticed for 4 days.
            exit 1
        fi
    else
        echo "worktree-merge: push to origin/$target_branch failed (not a protection rejection):" >&2
        cat "$push_err" >&2
        rm -f "$push_err"
        echo "  Local merge kept; land it manually." >&2
        exit 1
    fi
else
    echo "worktree-merge: no origin remote — local merge only."
fi

# ── Step 4b: post-land hook ────────────────────────────────────────────
# A repo may run a script after a successful landing — for example to reconcile
# a deployed artifact with the newly landed code. The hook is configured as
# GITFLOW_POST_LAND_HOOK (a repo-relative path) in .gitflow; empty means none.
# The hook receives --deploy when the run itself was invoked with --deploy, so
# it can tell "report the gap" from "actually redeploy". A non-zero hook exit
# fails the run — the landing is already safe on origin, but the operator must
# know the post-land step did not finish.
#
# corpos-toolkit's hook (scripts/gitflow-deploy.sh) carries the container
# image-rebuild + deploy-gap report that used to live inline here. Landing is
# not deploying: the canonical toolkit-server is the container, which advances
# only on an explicit image rebuild (bug
# worktree-merge-to-main-does-not-rebuild-main-checkout-binary).
if [[ -n "$GITFLOW_POST_LAND_HOOK" ]]; then
    hook_path="$REPO_ROOT/$GITFLOW_POST_LAND_HOOK"
    if [[ -x "$hook_path" ]]; then
        echo ""
        echo "worktree-merge: post-land hook → $GITFLOW_POST_LAND_HOOK"
        if [[ "$DO_DEPLOY" -eq 1 ]]; then
            "$hook_path" --deploy || exit 1
        else
            "$hook_path" || exit 1
        fi
    else
        echo "worktree-merge: ⚠ post-land hook '$GITFLOW_POST_LAND_HOOK' is not executable — skipping." >&2
    fi
fi

# ── Step 5: reap merged worktrees + branches ───────────────────────────
if [[ "$DO_REAP" -eq 1 ]]; then
    echo ""
    for b in "${MERGE_BRANCHES[@]}"; do
        wt="${WORKTREE_OF[$b]:-}"
        if [[ -n "$wt" && -d "$wt" ]]; then
            # Bug 1235 hazard 2 (self-reap): never delete the directory this
            # run is standing in. The snapshot above keeps the SCRIPT text
            # safe, but a removed cwd still breaks every relative path the
            # rest of the run uses, and leaves the operator's shell in a
            # deleted directory afterwards. Skip loudly with the recipe —
            # the worktree is merged, so reaping it is safe from anywhere
            # else.
            if [[ "$PWD" == "$wt" || "$PWD" == "$wt"/* ]]; then
                echo "worktree-merge: NOT reaping $wt — this run's working directory is inside it." >&2
                echo "  $b is merged; reap it from the main checkout with:  git worktree remove --force $wt && git branch -d $b" >&2
                continue
            fi
            # --force handles the Agent-tool's worktree lock; safe because the
            # branch's commits are now merged into the integration target.
            if git worktree remove --force "$wt" 2>/dev/null; then
                echo "worktree-merge: removed worktree $wt"
            else
                echo "worktree-merge: could not remove worktree $wt (left in place)" >&2
            fi
        fi
        if git branch -d "$b" 2>/dev/null; then
            echo "worktree-merge: deleted merged branch $b"
        else
            echo "worktree-merge: branch $b not fully merged or in use — left in place" >&2
        fi
    done
    git worktree prune
else
    echo "worktree-merge: --no-reap — worktrees/branches left in place."
fi

echo ""
echo "worktree-merge: done — ${#MERGE_BRANCHES[@]} branch(es) integrated into $target_branch."
