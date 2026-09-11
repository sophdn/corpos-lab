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
