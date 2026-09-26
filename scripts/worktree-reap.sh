#!/usr/bin/env bash
# Thin shim → the git-flow service (scripts/gitflow/). Canonical source:
# corpos-toolkit/gitflow. Do not edit; re-synced by install-gitflow.sh and
# drift-gated by scripts/gitflow/check-gitflow-drift.sh.
exec "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/gitflow/worktree-reap.sh" "$@"
