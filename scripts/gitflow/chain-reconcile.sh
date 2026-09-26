#!/usr/bin/env bash
# gitflow/chain-reconcile.sh <chain-slug> [--archive <branch>]... [--discard <branch>]... [--note <text>]
#
# Reconcile a chain's worktree branches at chain close. A chain can close while a
# branch still holds real, unmerged work (chain 536's high-floor arm sat on
# branch injection-assay-run, the only copy of a 960-completion run). Reap never
# surfaces such a branch — it is not merged — so a later session cannot tell
# parked-on-purpose from stranded. This makes the disposition explicit and
# machine-readable instead of a free-text line in the closure summary.
#
# It reads the git-flow ledger (branch↔chain, recorded by worktree-new --chain),
# lists the chain's branches, and for each UNMERGED one demands a disposition:
#
#   land     — run `worktree-merge.sh <branch>` first (landing goes through the
#              gate; this script never lands). Reported, not applied here.
#   --archive <branch>  — park on purpose: mark the branch archived in the ledger
#              (a durable, machine-readable marker) and leave it in place.
#   --discard <branch>  — delete the branch and its worktree (abandoned work).
#
# With no unresolved branch left it exits 0. While any unmerged branch is neither
# landed, archived, nor discarded it exits non-zero, naming each one — so a chain
# cannot be quietly closed over stranded work.
#
# ATTRIBUTION LIMIT: chain_close is a toolkit-server action with no git access, so
# it cannot invoke this itself. Run this as the close ritual. The straggler
# warning in worktree-reap.sh is the backstop: it surfaces any unmerged branch
# that slipped past, naming the owning chain from the same ledger.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=gitflow-common.sh
# shellcheck disable=SC1091
. "$SCRIPT_DIR/gitflow-common.sh"
gitflow_load_config

chain=""
archives=()
discards=()
note=""
while [ $# -gt 0 ]; do
    case "$1" in
        --archive) archives+=("${2:-}"); shift 2 ;;
        --discard) discards+=("${2:-}"); shift 2 ;;
        --note) note="${2:-}"; shift 2 ;;
        -*) echo "chain-reconcile: unknown flag '$1'" >&2; exit 2 ;;
        *) if [ -z "$chain" ]; then chain="$1"; shift; else echo "chain-reconcile: unexpected arg '$1'" >&2; exit 2; fi ;;
    esac
done
if [ -z "$chain" ]; then
    echo "usage: chain-reconcile.sh <chain-slug> [--archive <branch>]... [--discard <branch>]... [--note <text>]" >&2
    exit 2
fi

ledger="$(gitflow_ledger_path 2>/dev/null || true)"
if [ -z "$ledger" ] || [ ! -f "$ledger" ]; then
    echo "chain-reconcile: no ledger yet ($ledger) — no branches recorded for any chain." >&2
    echo "chain-reconcile: chain '$chain' has no attributed branches; nothing to reconcile."
    exit 0
fi

# Apply dispositions first, so the scan below sees their effect.
for b in "${archives[@]:-}"; do
    [ -n "$b" ] || continue
    gitflow_ledger_set_status "$b" archived "${note:-parked at chain close}"
    echo "chain-reconcile: archived (parked) branch '$b' — left in place, marked in the ledger."
done
for b in "${discards[@]:-}"; do
    [ -n "$b" ] || continue
    wt="$(git worktree list --porcelain | awk -v r="refs/heads/$b" '/^worktree /{p=$2} /^branch /{ if($2==r) print p }')"
    if [ -n "$wt" ] && [ -d "$wt" ]; then
        git worktree remove --force "$wt" 2>/dev/null && echo "chain-reconcile: removed worktree $wt"
    fi
    git branch -D "$b" 2>/dev/null && echo "chain-reconcile: discarded branch '$b'."
    gitflow_ledger_set_status "$b" discarded "${note:-discarded at chain close}"
done

# Scan the chain's branches: latest status per branch from the ledger.
unresolved=()
while IFS=$'\t' read -r b status; do
    [ -n "$b" ] || continue
    if ! git show-ref --verify --quiet "refs/heads/$b"; then
        echo "  $b — gone (landed+reaped or discarded); resolved."
        continue
    fi
    case "$status" in
        archived)  echo "  $b — ARCHIVED (parked on purpose); resolved."; continue ;;
        discarded) echo "  $b — marked discarded but branch still present; re-run --discard $b." ; unresolved+=("$b"); continue ;;
    esac
    if gitflow_branch_merged "$b"; then
        echo "  $b — merged into $GITFLOW_LANDING_BRANCH; resolved (reap will clean the worktree)."
    else
        echo "  $b — UNMERGED, no disposition."
        unresolved+=("$b")
    fi
done < <(awk -F'\t' -v c="$chain" '$2==c{st[$1]=$4} END{for(b in st) print b"\t"st[b]}' "$ledger")

echo
if [ "${#unresolved[@]}" -gt 0 ]; then
    {
        echo "chain-reconcile: chain '$chain' has ${#unresolved[@]} unresolved branch(es):"
        for b in "${unresolved[@]}"; do echo "  - $b"; done
        echo "Resolve each before closing the chain:"
        echo "  land it:     scripts/worktree-merge.sh <branch>"
        echo "  park it:     scripts/../gitflow/chain-reconcile.sh $chain --archive <branch> --note '<why>'"
        echo "  discard it:  scripts/../gitflow/chain-reconcile.sh $chain --discard <branch>"
    } >&2
    exit 1
fi
echo "chain-reconcile: chain '$chain' reconciled — every attributed branch is landed, archived, or discarded."
exit 0
