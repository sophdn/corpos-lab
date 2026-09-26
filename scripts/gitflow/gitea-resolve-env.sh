#!/usr/bin/env bash
# scripts/gitea-resolve-env.sh — derive GITEA_API, GITEA_OWNER, GITEA_REPO,
# GITEA_TOKEN from the current repo's origin remote and ~/.git-credentials.
#
# Source this from any git checkout with a gitea https origin:
#   . scripts/gitea-resolve-env.sh
#   curl -H "Authorization: token $GITEA_TOKEN" "$GITEA_API/repos/$GITEA_OWNER/$GITEA_REPO"
#
# Or call it as a command — it prints the exports for eval:
#   eval "$(scripts/gitea-resolve-env.sh)"
#
# Test overrides (for scripts/test-worktree-merge.sh and friends):
#   GITEA_RESOLVE_API_BASE   overrides the derived API base URL
#   GITEA_RESOLVE_TOKEN      overrides the credential-store lookup
#
# Exit / return 0 on success, 1 when the origin is not an https gitea
# remote or no token is found.

_gitea_resolve() {
    # Full override: when all four are already set, trust them and skip the
    # origin-derivation. A caller that fully specifies the gitea target (the
    # regression net pointing at a local stub with a non-https origin) is
    # honoured, and a real caller never has all four set by accident.
    if [[ -n "${GITEA_API:-}" && -n "${GITEA_OWNER:-}" && -n "${GITEA_REPO:-}" && -n "${GITEA_TOKEN:-}" ]]; then
        export GITEA_API GITEA_OWNER GITEA_REPO GITEA_TOKEN
        return 0
    fi
    local remote_url host_path api_host
    remote_url="$(git remote get-url origin 2>/dev/null || true)"
    [[ "$remote_url" == https://* ]] || return 1
    host_path="${remote_url#https://}"; host_path="${host_path%.git}"
    api_host="${host_path%%/*}"
    GITEA_OWNER="$(printf '%s' "$host_path" | awk -F/ '{print $(NF-1)}')"
    GITEA_REPO="$(printf '%s' "$host_path" | awk -F/ '{print $NF}')"
    GITEA_API="${GITEA_RESOLVE_API_BASE:-${WORKTREE_MERGE_API_BASE:-https://${api_host}/git/api/v1}}"
    GITEA_TOKEN="${GITEA_RESOLVE_TOKEN:-${WORKTREE_MERGE_TOKEN:-$(sed -n "s#https://[^:]*:\([0-9a-f]\{40\}\)@${api_host}.*#\1#p" "$HOME/.git-credentials" 2>/dev/null | head -1)}}"
    [[ -n "$GITEA_TOKEN" ]] || return 1
    export GITEA_API GITEA_OWNER GITEA_REPO GITEA_TOKEN
    return 0
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    # Called as a command — print exports.
    set -euo pipefail
    GITEA_API="" GITEA_OWNER="" GITEA_REPO="" GITEA_TOKEN=""
    if _gitea_resolve; then
        printf 'export GITEA_API=%q\n' "$GITEA_API"
        printf 'export GITEA_OWNER=%q\n' "$GITEA_OWNER"
        printf 'export GITEA_REPO=%q\n' "$GITEA_REPO"
        printf 'export GITEA_TOKEN=%q\n' "$GITEA_TOKEN"
    else
        echo "gitea-resolve-env: no https origin or no token found" >&2
        exit 1
    fi
else
    # Sourced — resolve into the caller's environment.
    GITEA_API="${GITEA_API:-}" GITEA_OWNER="${GITEA_OWNER:-}" GITEA_REPO="${GITEA_REPO:-}" GITEA_TOKEN="${GITEA_TOKEN:-}"
    _gitea_resolve || { echo "gitea-resolve-env: no https origin or no token found" >&2; }
fi
