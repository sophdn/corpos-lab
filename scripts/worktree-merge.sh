#!/usr/bin/env bash
# scripts/worktree-merge.sh — merge linked-worktree branches back into the
# current checkout AND land the result on origin. Run from the MAIN checkout
# (on main):
#   conflict-surface check → ordered merge → gate (scripts/gate.sh) → LAND on
#   origin (direct push, or branch+PR via the gitea API when main is
#   push-protected — bug 1203) → reap the merged worktrees + branches.
#   Aborts cleanly on a real conflict.
# (skill: worktree-workflow)
#
# Usage: scripts/worktree-merge.sh [--check-only|--no-gate|--no-reap] <branch|path>...
set -euo pipefail

DO_GATE=1
DO_REAP=1
CHECK_ONLY=0
BRANCHES=()
while [ $# -gt 0 ]; do
    case "$1" in
        --check-only) CHECK_ONLY=1 ;;
        --no-gate)    DO_GATE=0 ;;
        --no-reap)    DO_REAP=0 ;;
        --*) echo "worktree-merge: unknown option: $1" >&2; exit 2 ;;
        *) BRANCHES+=("$1") ;;
    esac
    shift
done
if [ ${#BRANCHES[@]} -eq 0 ]; then
    echo "usage: scripts/worktree-merge.sh [--check-only|--no-gate|--no-reap] <branch|path>..." >&2
    exit 2
fi

# core.bare auto-reset (the Agent-tool's isolation=worktree flips it). Idempotent.
if [ "$(git config core.bare 2>/dev/null || echo false)" = "true" ]; then
    git config core.bare false
    echo "worktree-merge: reset core.bare true → false"
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"
target="$(git branch --show-current 2>/dev/null || echo DETACHED)"
echo "worktree-merge: integration target = $target ($REPO_ROOT)"

# Divergence guard (bug 1203): unpushed commits on local main are exactly what
# a later `git reset --hard origin/main` silently destroys (observed 2026-07-22,
# five commits). Say so up front; the landing step below ships them together
# with this merge.
if git rev-parse --verify -q "origin/$target" >/dev/null 2>&1; then
    ahead="$(git rev-list --count "origin/$target..$target" 2>/dev/null || echo 0)"
    if [ "$ahead" -gt 0 ]; then
        echo "worktree-merge: ⚠ local $target is AHEAD of origin/$target by $ahead unpushed commit(s)."
        echo "  They land together with this merge. NEVER 'git reset --hard origin/$target' while ahead — that destroys them."
    fi
fi

# Resolve each arg to a branch; remember its worktree path for reaping.
declare -a MB=()
declare -A WT_OF=()
for arg in "${BRANCHES[@]}"; do
    if [ -d "$arg" ]; then
        b="$(git -C "$arg" branch --show-current 2>/dev/null || true)"
        [ -n "$b" ] || { echo "worktree-merge: $arg not on a branch — skipping" >&2; continue; }
        MB+=("$b"); WT_OF["$b"]="$(cd "$arg" && pwd)"
    else
        MB+=("$arg")
        wt="$(git worktree list --porcelain | awk -v r="refs/heads/$arg" '
            /^worktree /{p=$2} /^branch /{ if($2==r) print p }')"
        [ -n "$wt" ] && WT_OF["$arg"]="$wt"
    fi
done
[ ${#MB[@]} -gt 0 ] || { echo "worktree-merge: no mergeable branches resolved" >&2; exit 2; }

# Conflict-surface check: files touched by 2+ branches vs their merge-base.
echo "worktree-merge: conflict-surface check across ${#MB[@]} branch(es)…"
tmp="$(mktemp)"; : > "$tmp"
for b in "${MB[@]}"; do
    base="$(git merge-base HEAD "$b" 2>/dev/null || echo HEAD)"
    git diff --name-only "$base" "$b" | sed "s|\$|\t$b|"
done > "$tmp"
overlap="$(cut -f1 "$tmp" | sort | uniq -d || true)"
conflicts=0
if [ -n "$overlap" ]; then
    echo "  ⚠ files modified by more than one branch (textual merge may conflict):"
    while IFS= read -r f; do
        [ -z "$f" ] && continue
        echo "      $f →$(awk -F'\t' -v f="$f" '$1==f{printf " %s",$2}' "$tmp")"
        conflicts=1
    done <<< "$overlap"
else
    echo "  ✓ no file touched by more than one branch"
fi
rm -f "$tmp"

if [ "$CHECK_ONLY" -eq 1 ]; then
    [ "$conflicts" -eq 1 ] && { echo "worktree-merge: --check-only → CONFLICTS PRESENT"; exit 1; }
    echo "worktree-merge: --check-only → clean (safe to merge-back)"; exit 0
fi
if [ "$conflicts" -eq 1 ]; then
    echo "worktree-merge: conflict surface non-empty — refusing to auto-merge." >&2
    echo "  Resolve the overlaps above (or merge by hand) and re-run." >&2
    exit 1
fi

# Ordered merge: fast-forward the first when possible, --no-ff the rest.
first=1
for b in "${MB[@]}"; do
    if [ "$first" -eq 1 ] && git merge --ff-only "$b" 2>/dev/null; then
        echo "worktree-merge: fast-forwarded to $b"
    else
        if ! git merge --no-ff -m "merge(worktree): integrate $b" "$b"; then
            echo "worktree-merge: MERGE CONFLICT on $b — aborting this merge." >&2
            git merge --abort 2>/dev/null || true
            echo "  Resolve manually; already-merged branches were kept." >&2
            exit 1
        fi
        echo "worktree-merge: merged $b (--no-ff)"
    fi
    first=0
done

# Gate the merged result.
if [ "$DO_GATE" -eq 1 ]; then
    echo ""
    echo "worktree-merge: gate → scripts/gate.sh"
    ./scripts/gate.sh
else
    echo "worktree-merge: --no-gate — skipping gate."
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
    echo "worktree-merge: landing $target on origin…"
    push_err="$(mktemp)"
    if git push origin "$target" 2>"$push_err"; then
        echo "worktree-merge: pushed $target to origin"
        rm -f "$push_err"
    elif grep -qi 'protected branch' "$push_err"; then
        rm -f "$push_err"
        echo "worktree-merge: origin/$target is push-protected — landing via PR"
        remote_url="$(git remote get-url origin)"
        host_path="${remote_url#https://}"; host_path="${host_path%.git}"
        api_host="${host_path%%/*}"
        owner="$(printf '%s' "$host_path" | awk -F/ '{print $(NF-1)}')"
        repo="$(printf '%s' "$host_path" | awk -F/ '{print $NF}')"
        api="https://${api_host}/git/api/v1"
        tok="$(sed -n "s#https://[^:]*:\([0-9a-f]\{40\}\)@${api_host}.*#\1#p" "$HOME/.git-credentials" | head -1)"
        if [ -z "$tok" ]; then
            echo "worktree-merge: no gitea token for $api_host in ~/.git-credentials." >&2
            echo "  Work is safe but UNPUSHED on local $target — push it as a branch + open a PR by hand." >&2
            exit 1
        fi
        land="land/${target}-$(date +%Y%m%dT%H%M%S)"
        git push origin "$target:refs/heads/$land"
        echo "worktree-merge: merged state pushed as origin/$land"
        title="land: $(git log -1 --format=%s "$target")"
        pr_idx="$(curl -sk -X POST -H "Authorization: token $tok" -H 'Content-Type: application/json' \
            "$api/repos/$owner/$repo/pulls" \
            --data "$(jq -nc --arg h "$land" --arg t "$title" '{head:$h, base:"main", title:$t}')" \
            | jq -r '.number // empty')"
        if [ -z "$pr_idx" ]; then
            echo "worktree-merge: PR creation failed — work is safe on origin/$land; open the PR by hand." >&2
            exit 1
        fi
        # Branch protection requires the CI status check (ci / gate*) to pass
        # before a merge is accepted; merging earlier returns 405 "Please try
        # again later" (observed live on PR #17). Poll the commit status,
        # bounded, then merge. A pending state after the wait still attempts
        # the merge — a rejection leaves the PR open with instructions.
        sha="$(git rev-parse "$target")"
        echo "worktree-merge: waiting for required CI status on ${sha:0:9} (up to 15 min)…"
        ci_state=pending
        for _ in $(seq 1 45); do
            ci_state="$(curl -sk -H "Authorization: token $tok" \
                "$api/repos/$owner/$repo/commits/$sha/status" | jq -r '.state // "pending"')"
            case "$ci_state" in success|failure|error) break ;; esac
            sleep 20
        done
        echo "worktree-merge: CI status = $ci_state"
        if [ "$ci_state" = "failure" ] || [ "$ci_state" = "error" ]; then
            echo "worktree-merge: CI FAILED on the landing branch — PR #$pr_idx left open; fix and re-land." >&2
            exit 1
        fi
        merge_resp="$(mktemp)"
        merge_code="$(curl -sk -o "$merge_resp" -w '%{http_code}' -X POST \
            -H "Authorization: token $tok" -H 'Content-Type: application/json' \
            "$api/repos/$owner/$repo/pulls/$pr_idx/merge" --data '{"Do":"merge"}')"
        if [ "$merge_code" = "200" ]; then
            echo "worktree-merge: PR #$pr_idx merged on origin"
            git fetch origin
            if git diff --quiet "$target" "origin/$target"; then
                git reset --hard "origin/$target"
                echo "worktree-merge: local $target re-synced to origin (content verified identical first)"
            else
                echo "worktree-merge: PR merged but origin/$target content differs (concurrent changes?)." >&2
                echo "  NOT resetting local $target — reconcile manually; everything is on origin." >&2
            fi
            git push origin ":refs/heads/$land" 2>/dev/null || true
        else
            echo "worktree-merge: PR #$pr_idx created but auto-merge failed (HTTP $merge_code):" >&2
            cat "$merge_resp" >&2 || true
            echo "  Merge it in the gitea UI; local $target stays ahead until then (work is safe on origin/$land)." >&2
        fi
        rm -f "$merge_resp"
    else
        echo "worktree-merge: push to origin/$target failed (not a protection rejection):" >&2
        cat "$push_err" >&2
        rm -f "$push_err"
        echo "  Local merge kept; land it manually." >&2
        exit 1
    fi
else
    echo "worktree-merge: no origin remote — local merge only."
fi

# Reap merged worktrees + branches.
if [ "$DO_REAP" -eq 1 ]; then
    echo ""
    for b in "${MB[@]}"; do
        wt="${WT_OF[$b]:-}"
        if [ -n "$wt" ] && [ -d "$wt" ]; then
            git worktree remove --force "$wt" 2>/dev/null \
                && echo "worktree-merge: removed worktree $wt" \
                || echo "worktree-merge: could not remove worktree $wt (left in place)" >&2
        fi
        git branch -d "$b" 2>/dev/null \
            && echo "worktree-merge: deleted merged branch $b" \
            || echo "worktree-merge: branch $b not fully merged or in use — left in place" >&2
    done
    git worktree prune
else
    echo "worktree-merge: --no-reap — worktrees/branches left in place."
fi

echo ""
echo "worktree-merge: done — ${#MB[@]} branch(es) integrated into $target."
