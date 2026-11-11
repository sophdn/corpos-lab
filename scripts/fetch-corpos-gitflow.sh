#!/usr/bin/env bash
# scripts/fetch-corpos-gitflow.sh — download the published corpos-gitflow binary and put
# it on PATH. A thin wrapper over scripts/fetch-corpos-release.sh, which holds the
# download, checksum, and version checks for both corpos binaries.
#
# CANONICAL in the private toolkit service and copied into each consuming repo, together with
# fetch-corpos-release.sh. Edit it in the private toolkit service and run
# scripts/sync-distributed-scripts.sh.
#
# Usage:
#   scripts/fetch-corpos-gitflow.sh [dest_dir]     # default: $HOME/.local/bin
#
# Env: CORPOS_GITFLOW_TOKEN (required), CORPOS_GITFLOW_HOST (required),
#      CORPOS_GITFLOW_REPO, CORPOS_GITFLOW_SHA. See fetch-corpos-release.sh.
exec bash "$(dirname "${BASH_SOURCE[0]}")/fetch-corpos-release.sh" corpos-gitflow "$@"
