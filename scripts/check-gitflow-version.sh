#!/usr/bin/env bash
# scripts/check-gitflow-version.sh — consumer-side version-pin gate for the owned
# corpos-gitflow binary (chain cannibalize-git-flow-service-to-go).
#
# This is CANONICAL here and copied into each consumer's scripts/ when the repo is
# rolled onto the binary. It replaces that repo's consumer-side gitflow-drift text
# check: once a consumer's worktree shims exec `corpos-gitflow` instead of a synced
# scripts/gitflow/ copy, "is my copy of the service current?" becomes "is the
# corpos-gitflow binary I run current?".
#
# "Current" means the installed binary was built AT OR AFTER the last change to
# corpos-gitflow's own source (go/cmd/corpos-gitflow, go/internal/gitflow). The
# binary is stamped with the the private toolkit service commit it was built from, so the test
# is: is the last service-source commit an ANCESTOR of the binary's build commit?
# This deliberately does NOT compare against the private toolkit service HEAD — an unrelated
# the private toolkit service commit must not flag the binary stale, only a change to the
# service's own source.
#
# It mirrors gitflow-drift's loud-SKIP: without a the private toolkit service checkout (CI,
# containers) there is nothing to pin against, so it skips rather than fails.
set -euo pipefail

# The canonical the private toolkit service checkout to pin against. Overridable, defaulting to
# the conventional location, exactly like check-gitflow-drift.sh's GITFLOW_CANONICAL_DIR.
CANON="${GITFLOW_CANONICAL_REPO:-$HOME/dev/the private toolkit service}"

if [ ! -d "$CANON/.git" ]; then
    echo "[gitflow-version] no the private toolkit service checkout at $CANON — SKIP (set GITFLOW_CANONICAL_REPO to enforce)."
    exit 0
fi

if ! command -v corpos-gitflow >/dev/null 2>&1; then
    echo "[gitflow-version] ✖ corpos-gitflow is not on PATH — the worktree shims cannot run." >&2
    echo "  Install it: make -C \"$CANON/go\" corpos-gitflow-install   (or scripts/fetch-corpos-gitflow.sh)" >&2
    exit 1
fi

build_sha="$(corpos-gitflow version | awk '{print $2}')"
last_src="$(git -C "$CANON" log -1 --format=%H -- go/cmd/corpos-gitflow go/internal/gitflow 2>/dev/null || echo '')"

if [ -z "$last_src" ]; then
    echo "[gitflow-version] could not resolve the last corpos-gitflow source commit in $CANON — SKIP."
    exit 0
fi

if git -C "$CANON" merge-base --is-ancestor "$last_src" "$build_sha" 2>/dev/null; then
    echo "[gitflow-version] OK — corpos-gitflow $build_sha includes the latest service source."
    exit 0
fi

echo "[gitflow-version] ✖ STALE: corpos-gitflow ($build_sha) predates the latest change to its source." >&2
echo "  Re-install the current binary: make -C \"$CANON/go\" corpos-gitflow-install" >&2
exit 1
