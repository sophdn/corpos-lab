#!/usr/bin/env bash
# scripts/worktree-new.sh <slug> — start a unit of work in a linked worktree.
#
# Creates ../<repo>-wt-<slug> on a fresh branch <slug> off the current main and
# runs worktree-setup.sh inside it so commits there are gate-only (full gate, no
# main-checkout guard, no shared-daemon restart). This is the ONLY sanctioned way
# to start code work — the main checkout stays on `main`. Run from the main
# checkout. (skill: worktree-workflow)
#
#   scripts/worktree-new.sh fix-thing
#   cd ../<repo>-wt-fix-thing && …work…
#
# Optional: --chain <chain-slug> records which chain this branch serves in the
# git-flow ledger, so chain-close reconciliation (gitflow/chain-reconcile.sh) can
# enumerate the chain's branches and the straggler warning can name the owner.
set -euo pipefail

slug=""
chain=""
while [ $# -gt 0 ]; do
    case "$1" in
        --chain) chain="${2:-}"; shift 2 ;;
        --chain=*) chain="${1#--chain=}"; shift ;;
        -*) echo "worktree-new: unknown flag '$1'" >&2; exit 2 ;;
        *) if [ -z "$slug" ]; then slug="$1"; shift; else echo "worktree-new: unexpected arg '$1'" >&2; exit 2; fi ;;
    esac
done
if [ -z "$slug" ]; then
    echo "usage: scripts/worktree-new.sh <slug> [--chain <chain-slug>]" >&2
    exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

# shellcheck source=gitflow-common.sh
# shellcheck disable=SC1091
. "$SCRIPT_DIR/gitflow-common.sh"
gitflow_load_config "$repo_root"

repo_name="$(basename "$repo_root")"
wt_path="$(dirname "$repo_root")/${repo_name}-wt-${slug}"
branch="$slug"

if git show-ref --verify --quiet "refs/heads/$branch"; then
    echo "worktree-new: branch '$branch' already exists — pick another slug or reap it first." >&2
    exit 1
fi
if [ -e "$wt_path" ]; then
    echo "worktree-new: path '$wt_path' already exists." >&2
    exit 1
fi

git worktree add -b "$branch" "$wt_path" "$GITFLOW_LANDING_BRANCH"
( cd "$wt_path" && "$SCRIPT_DIR/worktree-setup.sh" )

# Record the branch↔chain link so chain-close reconciliation can find it. A
# missing --chain records "-" (unattributed) rather than nothing, so the branch
# still shows up as needing attention at reconcile time.
gitflow_ledger_record "$branch" "$chain"

echo
if [ -n "$chain" ]; then
    echo "worktree-new: ready → cd $wt_path  (branch '$branch', chain '$chain', gate-only hooks)"
else
    echo "worktree-new: ready → cd $wt_path  (branch '$branch', gate-only hooks)"
fi
