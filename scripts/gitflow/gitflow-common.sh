#!/usr/bin/env bash
# gitflow/gitflow-common.sh — the git-flow service config loader.
#
# Source this from any gitflow script AFTER resolving SCRIPT_DIR:
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   . "$SCRIPT_DIR/gitflow-common.sh"
#   gitflow_load_config            # reads <repo-root>/.gitflow, applies defaults
#
# After gitflow_load_config the caller has these exported:
#   GITFLOW_LANDING_BRANCH   the protected/integration branch (default: main)
#   GITFLOW_GATE_CMD         gate command, or "" for none (default: auto-detect)
#   GITFLOW_COMMIT_GATE_CMD  gate command for the worktree pre-commit hook, or ""
#                            for none (default: GITFLOW_GATE_CMD). A repo whose
#                            gate has a fast tier sets this to that tier, so a
#                            worktree commit runs the fast gate while a merge runs
#                            the full GITFLOW_GATE_CMD. Unset means both run the
#                            same command.
#   GITFLOW_WORKTREE         on|off (default: on)
#   GITFLOW_GUARD            on|off (default: on)
#   GITFLOW_MIGRATIONS_GLOB  migration-collision glob, or "" to skip (default: "")
#   GITFLOW_POST_LAND_HOOK   repo-relative script run after a landing, or "" (default: "")
#
# The per-repo file <repo-root>/.gitflow is plain bash `KEY=value` lines and is
# the source of these values. A repo with no .gitflow gets the bare defaults —
# which is the whole point: most repos need no config.

# A sentinel that means "the caller/config did not set this" so an explicit
# empty string in .gitflow (e.g. GITFLOW_GATE_CMD="") is honoured as "no gate"
# and NOT overwritten by the auto-detect default.
_GITFLOW_UNSET='__gitflow_unset__'

gitflow_repo_root() {
    git rev-parse --show-toplevel 2>/dev/null
}

# gitflow_autodetect_gate <repo-root> — pick a gate command from what the repo
# carries: precommit.sh, then gate.sh, then a Go module, then none. This is the
# default ONLY; an explicit GITFLOW_GATE_CMD in .gitflow always wins.
gitflow_autodetect_gate() {
    local root="$1"
    if [[ -f "$root/scripts/precommit.sh" ]]; then
        echo 'scripts/precommit.sh'
    elif [[ -f "$root/scripts/gate.sh" ]]; then
        echo 'scripts/gate.sh'
    elif [[ -f "$root/go/go.mod" ]]; then
        echo 'make -C go build && make -C go test'
    else
        echo ''
    fi
}

# gitflow_load_config [repo-root] — load .gitflow and apply defaults.
gitflow_load_config() {
    local root="${1:-$(gitflow_repo_root)}"

    # Seed every key with the unset sentinel unless the environment already set
    # it (an env value is an intentional override, e.g. from a test).
    GITFLOW_LANDING_BRANCH="${GITFLOW_LANDING_BRANCH:-$_GITFLOW_UNSET}"
    GITFLOW_GATE_CMD="${GITFLOW_GATE_CMD:-$_GITFLOW_UNSET}"
    GITFLOW_COMMIT_GATE_CMD="${GITFLOW_COMMIT_GATE_CMD:-$_GITFLOW_UNSET}"
    GITFLOW_WORKTREE="${GITFLOW_WORKTREE:-$_GITFLOW_UNSET}"
    GITFLOW_GUARD="${GITFLOW_GUARD:-$_GITFLOW_UNSET}"
    GITFLOW_MIGRATIONS_GLOB="${GITFLOW_MIGRATIONS_GLOB:-$_GITFLOW_UNSET}"
    GITFLOW_POST_LAND_HOOK="${GITFLOW_POST_LAND_HOOK:-$_GITFLOW_UNSET}"

    # The repo file is authoritative: its plain `KEY=value` assignments override
    # the sentinels (and any ambient env). An explicit empty value there — e.g.
    # GITFLOW_GATE_CMD="" — survives as "" and is honoured as "no gate", because
    # only the sentinel triggers a default below.
    if [[ -n "$root" && -f "$root/.gitflow" ]]; then
        # shellcheck disable=SC1090
        # shellcheck disable=SC1091
        . "$root/.gitflow"
    fi

    # Apply defaults for anything the environment and the file both left unset.
    [[ "$GITFLOW_LANDING_BRANCH" == "$_GITFLOW_UNSET" ]] && GITFLOW_LANDING_BRANCH='main'
    [[ "$GITFLOW_GATE_CMD" == "$_GITFLOW_UNSET" ]] && GITFLOW_GATE_CMD="$(gitflow_autodetect_gate "$root")"
    # The commit-hook gate defaults to the merge gate, so a repo that does not
    # split tiers runs the same command on a worktree commit and a merge. This
    # resolves AFTER GITFLOW_GATE_CMD so the fallback picks up the resolved value.
    [[ "$GITFLOW_COMMIT_GATE_CMD" == "$_GITFLOW_UNSET" ]] && GITFLOW_COMMIT_GATE_CMD="$GITFLOW_GATE_CMD"
    [[ "$GITFLOW_WORKTREE" == "$_GITFLOW_UNSET" ]] && GITFLOW_WORKTREE='on'
    [[ "$GITFLOW_GUARD" == "$_GITFLOW_UNSET" ]] && GITFLOW_GUARD='on'
    [[ "$GITFLOW_MIGRATIONS_GLOB" == "$_GITFLOW_UNSET" ]] && GITFLOW_MIGRATIONS_GLOB=''
    [[ "$GITFLOW_POST_LAND_HOOK" == "$_GITFLOW_UNSET" ]] && GITFLOW_POST_LAND_HOOK=''

    export GITFLOW_LANDING_BRANCH GITFLOW_GATE_CMD GITFLOW_COMMIT_GATE_CMD \
           GITFLOW_WORKTREE GITFLOW_GUARD GITFLOW_MIGRATIONS_GLOB GITFLOW_POST_LAND_HOOK
}

# ── merged-detection ────────────────────────────────────────────────────────
# gitflow_branch_merged <branch> — exit 0 if <branch>'s tip is an ancestor of the
# landing branch (i.e. already merged), 1 otherwise. Shared by worktree-reap.sh
# (which branches are stragglers) and chain-reconcile.sh (which of a chain's
# branches still hold unmerged work). One definition so the two never disagree.
gitflow_branch_merged() {
    local branch="$1"
    git merge-base --is-ancestor "refs/heads/$branch" "$GITFLOW_LANDING_BRANCH" 2>/dev/null
}

# ── chain↔branch ledger ─────────────────────────────────────────────────────
# A per-repo, uncommitted record of which chain each worktree branch belongs to,
# so chain-close reconciliation can enumerate a chain's branches mechanically and
# the straggler warning can name the owning chain. It lives in the shared git dir
# (git-common-dir), so every linked worktree of the repo sees the same ledger.
# Format: TAB-separated `branch<TAB>chain<TAB>created_epoch<TAB>status[<TAB>note]`.
# status is one of: active | archived | discarded.
gitflow_ledger_path() {
    local common
    common="$(git rev-parse --git-common-dir 2>/dev/null)" || return 1
    # git-common-dir may be relative to CWD; make it absolute.
    case "$common" in /*) : ;; *) common="$(pwd)/$common" ;; esac
    printf '%s/gitflow-chain-branches\n' "$common"
}

# gitflow_ledger_record <branch> <chain> — append an active row. A blank chain is
# recorded as "-" so attribution-absent branches are still listed (and visible as
# unattributed at reconcile time).
gitflow_ledger_record() {
    local branch="$1" chain="${2:-}" path
    path="$(gitflow_ledger_path)" || return 0
    [ -n "$chain" ] || chain="-"
    printf '%s\t%s\t%s\tactive\n' "$branch" "$chain" "$(date +%s)" >> "$path"
}

# gitflow_ledger_set_status <branch> <status> [note] — rewrite the branch's row(s)
# with a new status and optional note. Used by chain-reconcile to mark a branch
# archived (parked on purpose) or discarded.
gitflow_ledger_set_status() {
    local branch="$1" status="$2" note="${3:-}" path tmp
    path="$(gitflow_ledger_path)" || return 0
    [ -f "$path" ] || return 0
    tmp="$(mktemp)"
    awk -v b="$branch" -v s="$status" -v n="$note" -F'\t' 'BEGIN{OFS="\t"}
        $1==b { $4=s; if (n!="") { $5=n } print; next }
        { print }' "$path" > "$tmp" && mv "$tmp" "$path"
}

# ── private-subtree declaration (shared with the canonical-home guard) ────────
# gitflow_private_subtrees [repo-root] — echo the private-subtree path prefixes a
# repo declares in its .canonical-home (one per line), or nothing. worktree-reap
# reuses this so the SAME declaration that blocks committing private data also
# blocks reaping a worktree that would strand it.
gitflow_private_subtrees() {
    local root="${1:-$(gitflow_repo_root)}" cfg line verb rest
    cfg="$root/.canonical-home"
    [ -f "$cfg" ] || return 0
    while IFS= read -r line || [ -n "$line" ]; do
        line="${line%%#*}"
        line="$(printf '%s' "$line" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"
        [ -z "$line" ] && continue
        verb="${line%%[[:space:]]*}"
        rest="$(printf '%s' "$line" | sed -E 's/^[^[:space:]]+[[:space:]]+//')"
        [ "$verb" = "private-subtree" ] && printf '%s\n' "$rest"
    done < "$cfg"
}
