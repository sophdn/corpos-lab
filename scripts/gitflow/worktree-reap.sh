#!/usr/bin/env bash
# scripts/worktree-reap.sh — delete local branches already merged into main and
# prune their worktrees. Run from the main checkout. NEVER touches main or any
# unmerged branch (uses `git branch -d`, which refuses unmerged). The standalone
# cleanup for stragglers that worktree-merge.sh didn't reap. (skill: worktree-workflow)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=gitflow-common.sh
# shellcheck disable=SC1091
. "$SCRIPT_DIR/gitflow-common.sh"
gitflow_load_config

# The Agent-tool's isolation=worktree can flip core.bare=true on the shared
# config; reset it so work-tree commands work. Idempotent.
if [ "$(git config core.bare 2>/dev/null || echo false)" = "true" ]; then
    git config core.bare false
fi

repo_root="$(git rev-parse --show-toplevel)"
# A worktree older than this (by its branch-tip commit date) that still holds
# unmerged work draws a stalled-chain warning. Override with GITFLOW_STALE_DAYS.
GITFLOW_STALE_DAYS="${GITFLOW_STALE_DAYS:-14}"
# The private-subtree prefixes this repo declares (via .canonical-home). Reaping a
# worktree that holds such files absent from main would strand them — the exact
# data-loss mode that motivated this gate (corpus/private/, 2026-09-13).
mapfile -t _private_prefixes < <(gitflow_private_subtrees "$repo_root")

# worktree_strands_private <wt> — exit 0 if the worktree holds a file under a
# declared private subtree that is absent from the main checkout.
worktree_strands_private() {
    local wt="$1" pfx f rel
    for pfx in "${_private_prefixes[@]:-}"; do
        [ -n "$pfx" ] || continue
        [ -d "$wt/$pfx" ] || continue
        while IFS= read -r f; do
            rel="${f#"$wt/"}"
            [ -e "$repo_root/$rel" ] || return 0
        done < <(find "$wt/$pfx" -type f 2>/dev/null)
    done
    return 1
}

git worktree prune

reaped=0
while IFS= read -r b; do
    [ -z "$b" ] && continue
    [ "$b" = "$GITFLOW_LANDING_BRANCH" ] && continue
    wt="$(git worktree list --porcelain | awk -v r="refs/heads/$b" '
        /^worktree /{p=$2} /^branch /{ if($2==r) print p }')"
    if [ -n "$wt" ] && [ -d "$wt" ]; then
        # Bug 1350: a worktree whose branch ref is still at the base main SHA
        # counts as "merged" by `git branch --merged` (it is not ahead of main),
        # but it may hold STAGED or uncommitted changes with a commit still in
        # flight — its pre-commit gate running, the branch ref not yet advanced.
        # Reaping it with --force then deletes unfinalized work (observed
        # 2026-09-22: a mid-gate commit's worktree was removed and the staged fix
        # was lost). Never --force past a dirty or mid-commit worktree; only a
        # clean one is safe to remove.
        if [ -n "$(git -C "$wt" status --porcelain 2>/dev/null)" ]; then
            echo "reap: worktree $wt has uncommitted/staged changes — skipping branch $b (refusing to delete unfinalized work)" >&2
            continue
        fi
        wt_gitdir="$(git -C "$wt" rev-parse --absolute-git-dir 2>/dev/null || true)"
        if [ -n "$wt_gitdir" ] && [ -e "$wt_gitdir/index.lock" ]; then
            echo "reap: worktree $wt has a commit in progress (index.lock present) — skipping branch $b" >&2
            continue
        fi
        # Data-loss guard: a gitignored private subtree built inside a worktree is
        # not carried by the merge and would vanish with the worktree dir. Refuse.
        if worktree_strands_private "$wt"; then
            echo "reap: worktree $wt holds gitignored private-subtree files absent from main — skipping branch $b (refusing to strand private data; author private content in the MAIN checkout, not a worktree)" >&2
            continue
        fi
        if git worktree remove --force "$wt" 2>/dev/null; then
            echo "reap: removed worktree $wt"
        else
            echo "reap: could not remove worktree $wt — skipping branch $b" >&2
            continue
        fi
    fi
    if git branch -d "$b" 2>/dev/null; then
        echo "reap: deleted merged branch $b"
        reaped=$((reaped + 1))
    fi
done < <(git branch --merged "$GITFLOW_LANDING_BRANCH" --format='%(refname:short)')

git worktree prune

# Warn on stale worktrees still holding UNMERGED work — a stalled-chain signal.
# Reap never touches unmerged branches, so these are invisible to the loop above;
# without a warning they sit forever, indistinguishable from parked-on-purpose.
# Non-destructive: this only prints. It names the owning chain from the ledger
# when known, and points at the two ways to resolve it.
now_epoch="$(date +%s)"
stale_secs=$(( GITFLOW_STALE_DAYS * 86400 ))
ledger="$(gitflow_ledger_path 2>/dev/null || true)"
warned=0
while IFS=$'\t' read -r wt b; do
    [ -z "$b" ] && continue
    [ "$b" = "$GITFLOW_LANDING_BRANCH" ] && continue
    gitflow_branch_merged "$b" && continue
    tip_epoch="$(git log -1 --format=%ct "refs/heads/$b" 2>/dev/null || echo "$now_epoch")"
    age=$(( now_epoch - tip_epoch ))
    [ "$age" -lt "$stale_secs" ] && continue
    days=$(( age / 86400 ))
    chain=""
    if [ -n "$ledger" ] && [ -f "$ledger" ]; then
        chain="$(awk -F'\t' -v br="$b" '$1==br{c=$2} END{if(c!="-")print c}' "$ledger")"
    fi
    echo "reap: WARN worktree $wt (branch $b${chain:+, chain $chain}) holds UNMERGED work, last commit ${days}d ago — land it (worktree-merge.sh $b) or reconcile it (chain-reconcile.sh <chain>)." >&2
    warned=$((warned + 1))
done < <(git worktree list --porcelain | awk '
    /^worktree /{p=$2; b=""}
    /^branch /{b=$2; sub("refs/heads/","",b); print p"\t"b}')

if [ "$warned" -gt 0 ]; then
    echo "reap: $warned stale unmerged worktree(s) warned (older than ${GITFLOW_STALE_DAYS}d)." >&2
fi
echo "reap: done ($reaped branch(es) reaped)."
