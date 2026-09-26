#!/usr/bin/env bash
# check-gitflow-drift.sh — the git-flow service drift gate. Mirrors
# check-skills-embed-drift.sh: a consumer's copy of the service must match
# corpos-toolkit's canonical, and drift cannot go unnoticed.
#
# Two roles, auto-detected:
#
#   CONSUMER-SIDE (hard). Run from a consumer repo (one that carries
#   scripts/gitflow/). Compares each manifest file in scripts/gitflow/ against
#   corpos-toolkit's canonical gitflow/ on disk. Fails naming the drifted
#   file(s) and the re-sync command. Loud-SKIPs when the corpos-toolkit checkout
#   is absent (CI, the container) — exactly like the skills gate skips without a
#   sibling corpos.
#
#   OWNER-SIDE (advisory). Run from corpos-toolkit itself (the repo that owns
#   gitflow/). Reports which registered consumers' scripts/gitflow/ has drifted
#   from canonical, so a canonical edit's fan-out debt is visible. NEVER fails
#   the owner's commit — a hard owner gate across every consumer would wedge each
#   canonical edit behind several live protected-main landings, the exact
#   fragile machinery this service exists to make safe.
#
# Usage: [GITFLOW_CANONICAL_DIR=/path/to/corpos-toolkit/gitflow] check-gitflow-drift.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo "")"

# Canonical gitflow/ dir. When this very script is the canonical copy (owner
# mode) SCRIPT_DIR already IS the canonical dir. Otherwise default to the
# conventional corpos-toolkit checkout, overridable.
DEFAULT_CANON="${GITFLOW_CANONICAL_DIR:-$HOME/dev/corpos-toolkit/gitflow}"

read_manifest() {
    local mf="$1"
    [ -f "$mf" ] || return 1
    grep -vE '^[[:space:]]*#|^[[:space:]]*$' "$mf"
}

# ── Owner mode: this repo IS the canonical owner ──────────────────────────
# Detected when SCRIPT_DIR holds the owner-only installer + consumer registry.
if [ -f "$SCRIPT_DIR/install-gitflow.sh" ] && [ -f "$SCRIPT_DIR/consumers.txt" ]; then
    CANON="$SCRIPT_DIR"
    manifest="$(read_manifest "$SCRIPT_DIR/manifest.txt")" || {
        echo "[gitflow-drift] no manifest.txt in $SCRIPT_DIR — cannot check." >&2
        exit 0
    }
    dev_root="$(dirname "$(dirname "$SCRIPT_DIR")")"   # ~/dev (parent of corpos-toolkit)
    any_drift=0
    while read -r name relpath; do
        [ -z "$name" ] && continue
        consumer="$dev_root/$relpath"
        [ -d "$consumer/scripts/gitflow" ] || { echo "[gitflow-drift] NOTE: $name not yet onboarded (no scripts/gitflow/)."; continue; }
        drift=()
        while IFS= read -r f; do
            [ -z "$f" ] && continue
            if [ ! -f "$consumer/scripts/gitflow/$f" ] || ! cmp -s "$CANON/$f" "$consumer/scripts/gitflow/$f"; then
                drift+=("$f")
            fi
        done <<< "$manifest"
        if [ "${#drift[@]}" -gt 0 ]; then
            any_drift=1
            echo "[gitflow-drift] ADVISORY: $name ($relpath) has drifted from canonical:"
            for f in "${drift[@]}"; do echo "    - $f"; done
        fi
    done <<< "$(read_manifest "$SCRIPT_DIR/consumers.txt")"
    if [ "$any_drift" -eq 1 ]; then
        echo "[gitflow-drift] re-sync each drifted consumer: gitflow/install-gitflow.sh <repo>"
        echo "[gitflow-drift] (advisory only — the owner commit is NOT blocked)"
    else
        echo "[gitflow-drift] OK — every onboarded consumer matches canonical."
    fi
    exit 0
fi

# ── Consumer mode: this repo carries scripts/gitflow/ ─────────────────────
if [ -z "$REPO_ROOT" ] || [ ! -d "$REPO_ROOT/scripts/gitflow" ]; then
    echo "[gitflow-drift] no scripts/gitflow/ here — nothing to check."
    exit 0
fi

CANON="$DEFAULT_CANON"
if [ ! -d "$CANON" ]; then
    echo "[gitflow-drift] canonical gitflow/ not found at $CANON — SKIP (set GITFLOW_CANONICAL_DIR to enforce)." >&2
    exit 0
fi

manifest="$(read_manifest "$CANON/manifest.txt")" || {
    echo "[gitflow-drift] no manifest.txt in $CANON — SKIP." >&2
    exit 0
}

drift=()
missing=()
while IFS= read -r f; do
    [ -z "$f" ] && continue
    if [ ! -f "$REPO_ROOT/scripts/gitflow/$f" ]; then
        missing+=("$f")
    elif ! cmp -s "$CANON/$f" "$REPO_ROOT/scripts/gitflow/$f"; then
        drift+=("$f")
    fi
done <<< "$manifest"

rc=0
if [ "${#missing[@]}" -gt 0 ]; then
    echo "[gitflow-drift] MISSING from scripts/gitflow/:" >&2
    for f in "${missing[@]}"; do echo "    - $f" >&2; done
    rc=1
fi
if [ "${#drift[@]}" -gt 0 ]; then
    echo "[gitflow-drift] STALE vs canonical ($CANON):" >&2
    for f in "${drift[@]}"; do echo "    - $f" >&2; done
    rc=1
fi
if [ "$rc" -ne 0 ]; then
    echo "[gitflow-drift] Fix: re-sync from corpos-toolkit: gitflow/install-gitflow.sh $(basename "$REPO_ROOT")" >&2
    exit 1
fi
echo "[gitflow-drift] OK — scripts/gitflow/ matches canonical ($CANON)."
exit 0
