#!/usr/bin/env bash
# scripts/gitea-pr-land.sh — the gitea landing handshake: wait for a pull
# request's CI to be acceptable, then merge it.
#
# Split out of scripts/worktree-merge.sh so the wait-then-merge logic has a
# regression net that does not need a real remote (scripts/test-gitea-pr-land.sh
# drives it entirely against a local HTTP stub). worktree-merge.sh owns the
# push / PR-create / re-sync / reap around it; this script owns only the
# handshake.
#
# Usage:
#   gitea-pr-land.sh --pr <number> --sha <commit> [--base <branch>]
#
# Required environment (worktree-merge.sh exports these from its
# resolve_gitea_remote):
#   GITEA_API      API base, e.g. https://host/git/api/v1
#   GITEA_OWNER    repo owner
#   GITEA_REPO     repo name
#   GITEA_TOKEN    token with repo scope
#
# Optional knobs:
#   GITEA_CI_TIMEOUT        seconds to wait for required CI (default 1800)
#   GITEA_CI_INTERVAL       seconds between polls (default 20)
#   GITEA_MERGE_RETRIES     extra merge attempts after a rejection (default 1)
#   GITEA_MERGE_RETRY_DELAY seconds before a merge retry (default 30)
#
# Exit 0 = the PR is merged on origin. Exit 1 = it is not, with the reason on
# stderr; the caller is expected to leave the PR open and say so loudly.

set -euo pipefail

PR=""
SHA=""
BASE="main"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --pr)   PR="$2"; shift 2 ;;
        --sha)  SHA="$2"; shift 2 ;;
        --base) BASE="$2"; shift 2 ;;
        -h|--help) sed -n '2,30p' "$0"; exit 0 ;;
        *) echo "gitea-pr-land: unknown argument: $1" >&2; exit 2 ;;
    esac
done

[[ -n "$PR"  ]] || { echo "gitea-pr-land: --pr is required" >&2; exit 2; }
[[ -n "$SHA" ]] || { echo "gitea-pr-land: --sha is required" >&2; exit 2; }
# Auto-resolve gitea env vars when not already set (bug 1301).
# worktree-merge.sh exports these before calling us; standalone callers
# get them derived from the git remote and ~/.git-credentials.
_need_resolve=0
for v in GITEA_API GITEA_OWNER GITEA_REPO GITEA_TOKEN; do
    [[ -n "${!v:-}" ]] || _need_resolve=1
done
if [[ "$_need_resolve" -eq 1 ]]; then
    _resolve="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/gitea-resolve-env.sh"
    if [[ -f "$_resolve" ]]; then
        # shellcheck source=gitea-resolve-env.sh
        # shellcheck disable=SC1091
        . "$_resolve"
    fi
fi
for v in GITEA_API GITEA_OWNER GITEA_REPO GITEA_TOKEN; do
    [[ -n "${!v:-}" ]] || { echo "gitea-pr-land: $v is required (set it, or run from a repo with an https gitea origin)" >&2; exit 2; }
done

CI_TIMEOUT="${GITEA_CI_TIMEOUT:-1800}"
CI_INTERVAL="${GITEA_CI_INTERVAL:-20}"
MERGE_RETRIES="${GITEA_MERGE_RETRIES:-1}"
MERGE_RETRY_DELAY="${GITEA_MERGE_RETRY_DELAY:-30}"

REPO_API="$GITEA_API/repos/$GITEA_OWNER/$GITEA_REPO"

# ── Signal handling (bug 1298) ────────────────────────────────────────────────
# Session cleanup can send SIGTERM or SIGHUP between the CI poll and the merge
# POST, killing the process silently and leaving an unmerged land/* PR. Trap
# those signals so the death is visible and exits nonzero.
_signal_exit() {
    echo "gitea-pr-land: KILLED by signal before PR #$PR could be merged — PR is still open on origin ($GITEA_OWNER/$GITEA_REPO)." >&2
    exit 130
}
trap _signal_exit TERM HUP PIPE

say() { echo "gitea-pr-land: $*"; }

api_get() { curl -sk -m 15 -H "Authorization: token $GITEA_TOKEN" "$REPO_API/$1"; }

# pr_is_merged returns 0 when gitea reports the PR already merged — a human or a
# concurrent landing got there first. That is SUCCESS, never a failure: the merge
# POST on an already-merged PR returns a non-200 (observed HTTP 405 on this
# instance), which must not be read as "auto-merge failed".
pr_is_merged() {
    local body
    body="$(api_get "pulls/$PR" || true)"
    [[ "$(printf '%s' "$body" | jq -r 'if type=="object" then (.merged // false) else false end' 2>/dev/null)" == "true" ]]
}

# combined_state prints the whole-commit combined status. This is gitea's
# roll-up across EVERY context that reported on the sha, required or not, so it
# is strictly stricter than the server's own merge gate — used only as the
# fallback when branch protection can't be read.
combined_state() {
    local body
    body="$(api_get "commits/$SHA/status" || true)"
    [[ -n "$body" ]] || { echo unknown; return; }
    printf '%s' "$body" | jq -r '.state // "unknown"' 2>/dev/null || echo unknown
}

# REQUIRED_CONTEXTS is the branch-protection glob list, resolved once. The three
# outcomes are distinguished because they mean different things:
#   resolved, non-empty -> gate on exactly these globs
#   resolved, empty     -> the branch requires no checks; merge now
#   unresolved          -> no protection rule readable; fall back to combined
REQUIRED_CONTEXTS=()
CONTEXTS_RESOLVED=0
resolve_required_contexts() {
    local body enabled
    body="$(api_get "branch_protections/$BASE" || true)"
    # A 404 body has no branch_name; anything unparseable lands here too.
    [[ "$(printf '%s' "$body" | jq -r '.branch_name // empty' 2>/dev/null)" == "$BASE" ]] || return 1
    enabled="$(printf '%s' "$body" | jq -r '.enable_status_check // false' 2>/dev/null)"
    CONTEXTS_RESOLVED=1
    [[ "$enabled" == "true" ]] || return 0
    mapfile -t REQUIRED_CONTEXTS < <(printf '%s' "$body" | jq -r '.status_check_contexts[]?' 2>/dev/null)
    return 0
}

# required_state prints success | pending | failure for the REQUIRED contexts
# only. Gitea merges on these alone — verified live against PR #24, where the
# combined status was pending (pinned by the unrequired 'ci / mirror-publish
# (push)') while 'ci / precommit (push)' was green and POST …/merge returned
# 200. Waiting on the combined status is therefore stricter than the server's
# own gate, which is what stalled every contended landing (bug 1224).
#
# A required glob that matches NO reported context is pending, not success: the
# check has not started yet, and treating absent as green would merge before CI
# ran at all.
required_state() {
    local body latest overall=success
    body="$(api_get "statuses/$SHA" || true)"
    latest="$(printf '%s' "$body" | jq -r '
        if type == "array" then
            group_by(.context) | map(sort_by(.id // 0) | last)
            | .[] | "\(.context)\t\(.status)"
        else empty end' 2>/dev/null || true)"

    local glob ctx st found
    for glob in "${REQUIRED_CONTEXTS[@]}"; do
        found=""
        while IFS=$'\t' read -r ctx st; do
            [[ -n "$ctx" ]] || continue
            # Intentional glob match: status_check_contexts entries are globs
            # ('ci / precommit*'), matched the same way gitea matches them.
            # shellcheck disable=SC2053
            [[ "$ctx" == $glob ]] || continue
            # A glob can match several contexts — notably the same job's `push`
            # and `pull_request` event variants (ci / precommit (push) AND
            # ci / precommit (pull_request)). Gitea's merge gate requires EVERY
            # matching context to be successful: a merge POST while one variant
            # is still pending returns 405 "Not all required status checks
            # successful" (observed 2026-09-07 on PR #73). So a pending sibling
            # HOLDS the glob — success is provisional and any pending entry drops
            # the glob back to pending. Waiting for both variants is correct, not
            # over-strict; the ~26-min waits are the slower push-event run queued
            # behind the single-parallel runner (a CI-config issue, not this).
            case "$st" in
                failure|error) found=failure; break ;;
                success)       [[ -n "$found" ]] || found=success ;;
                *)             found=pending ;;
            esac
        done <<<"$latest"
        case "$found" in
            failure) echo failure; return ;;
            success) ;;
            *)       overall=pending ;;
        esac
    done
    echo "$overall"
}

ci_state() {
    if (( CONTEXTS_RESOLVED == 0 )); then
        combined_state
    elif (( ${#REQUIRED_CONTEXTS[@]} == 0 )); then
        echo success
    else
        required_state
    fi
}

# If the PR is already merged (a re-run, or a human/concurrent merge got there
# first), there is nothing to wait for or merge.
if pr_is_merged; then
    say "PR #$PR is already merged on origin — nothing to do."
    exit 0
fi

# ── Wait for CI ──────────────────────────────────────────────────────────────
if resolve_required_contexts; then
    if (( ${#REQUIRED_CONTEXTS[@]} == 0 )); then
        say "branch protection on '$BASE' requires no status checks — merging without a wait."
    else
        say "waiting on the required context(s) for '$BASE': ${REQUIRED_CONTEXTS[*]}"
    fi
else
    say "no readable branch protection for '$BASE' — falling back to the combined commit status."
fi

deadline=$((SECONDS + CI_TIMEOUT))
poll_start=$SECONDS
state=""
while :; do
    state="$(ci_state)"
    case "$state" in
        success)
            break
            ;;
        failure|error)
            echo "gitea-pr-land: CI FAILED on ${SHA:0:9} (state=$state) — PR #$PR left open; fix and re-land." >&2
            exit 1
            ;;
    esac
    elapsed=$(( SECONDS - poll_start ))
    if (( SECONDS >= deadline )); then
        say "CI still pending after ${elapsed}s (state=$state) — attempting the merge anyway."
        break
    fi
    # Suggestion 95: print elapsed time so a queued job is distinguishable from
    # a hung one. The runner has capacity 1 and jobs can pend 30+ minutes
    # legitimately, but a stale pending for longer than any green run should
    # prompt a look at the runner.
    elapsed_min=$(( elapsed / 60 ))
    elapsed_rem=$(( elapsed % 60 ))
    if (( elapsed > 0 )); then
        if (( elapsed_min > 0 )); then
            say "CI pending for ${elapsed_min}m${elapsed_rem}s (${elapsed}s total) on ${SHA:0:9} — state=$state"
        else
            say "CI pending for ${elapsed}s on ${SHA:0:9} — state=$state"
        fi
        if (( elapsed_min >= 10 )); then
            say "  the runner is maxParallel=1. If another repo's job is ahead in the queue,"
            say "  nothing will start until it finishes. Check from the runner host:"
            say "    ssh <user>@<runner-host> 'docker ps -a --format \"{{.Names}}\\t{{.Status}}\"'"
        fi
    fi
    sleep "$CI_INTERVAL"
done
say "CI state = $state (waited ${elapsed:-0}s)"

# ── Merge, with a retry ──────────────────────────────────────────────────────
attempt=0
while :; do
    resp="$(mktemp)"
    # `|| true`: a transport failure (host down, timeout) must not abort under
    # set -e — it has to reach the retry and then the loud exit 1, the same as
    # any other non-200. curl writes 000 to %{http_code} in that case.
    code="$(curl -sk -m 30 -o "$resp" -w '%{http_code}' -X POST \
        -H "Authorization: token $GITEA_TOKEN" -H 'Content-Type: application/json' \
        "$REPO_API/pulls/$PR/merge" --data '{"Do":"merge"}' || true)"
    if [[ "$code" == "200" ]]; then
        rm -f "$resp"
        say "PR #$PR merged on origin"
        exit 0
    fi
    # A non-200 can simply mean the PR is already merged — a human or a
    # concurrent run beat us to it. That is success, not a failure.
    if pr_is_merged; then
        rm -f "$resp"
        say "PR #$PR already merged on origin (confirmed after HTTP $code)"
        exit 0
    fi
    say "merge of PR #$PR rejected (HTTP $code): $(tr -d '\n' < "$resp")"
    rm -f "$resp"
    attempt=$((attempt + 1))
    if (( attempt > MERGE_RETRIES )); then
        echo "gitea-pr-land: PR #$PR could not be merged after $attempt attempt(s) (last HTTP $code)." >&2
        exit 1
    fi
    say "retrying the merge in ${MERGE_RETRY_DELAY}s (attempt $((attempt + 1)) of $((MERGE_RETRIES + 1)))…"
    sleep "$MERGE_RETRY_DELAY"
    state="$(ci_state)"
    if [[ "$state" == "failure" || "$state" == "error" ]]; then
        echo "gitea-pr-land: CI went $state while retrying — PR #$PR left open." >&2
        exit 1
    fi
done
