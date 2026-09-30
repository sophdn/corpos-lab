#!/usr/bin/env bash
# scripts/fetch-corpos-gitflow.sh — download the published corpos-gitflow binary
# and put it on PATH. The CONSUMER side of the git-flow binary-fetch model (chain
# cannibalize-git-flow-service-to-go), mirroring scripts/fetch-corpos-gate.sh.
#
# This file is CANONICAL here and COPIED into each consuming repo (the same
# discipline as fetch-corpos-gate.sh: the copy is real because a consumer's CI
# can't reach this tree). the private toolkit service itself does NOT fetch — it OWNS the
# binary and builds it locally (`make -C go corpos-gitflow-install`). This script
# is what the other four repos use once they are rolled onto the binary.
#
# Usage:
#   scripts/fetch-corpos-gitflow.sh [dest_dir]     # default: $HOME/.local/bin
#
# Env:
#   CORPOS_GITFLOW_TOKEN   gitea token with READ scope on the toolkit repo (required)
#   CORPOS_GITFLOW_HOST    gitea host, e.g. git.example.internal (required)
#   CORPOS_GITFLOW_REPO    owner/repo to fetch from (default: shared/the private toolkit service)
#   CORPOS_GITFLOW_SHA     if set, REQUIRE the downloaded binary to report this SHA
#
# CORPOS_GITFLOW_HOST has no default ON PURPOSE: baking this fleet's gitea
# hostname into a public repo is what the pii-scan rejects (and did, for
# fetch-corpos-gate.sh). A consumer's workflow supplies the host with the token.
#
# WHY A TOKEN IS REQUIRED. Every repo on this gitea instance is private, so an
# unauthenticated fetch fails. Publishing the binary as a release asset removes
# the TOOLCHAIN from a consumer's CI (no Go, no module downloads); it does not
# remove authentication.
#
# The download is checksum-verified and then executed to confirm it reports a
# version, so a truncated or wrong-arch artifact fails HERE with a clear message.
set -euo pipefail

DEST="${1:-$HOME/.local/bin}"
REPO="${CORPOS_GITFLOW_REPO:-shared/the private toolkit service}"
TAG="corpos-gitflow-latest"
ASSET="corpos-gitflow-linux-amd64"

if [ -z "${CORPOS_GITFLOW_HOST:-}" ]; then
  echo "ERROR: CORPOS_GITFLOW_HOST is not set (e.g. git.example.internal)." >&2
  echo "  It has no default so this repo carries no infrastructure hostname;" >&2
  echo "  set it in the consuming workflow next to CORPOS_GITFLOW_TOKEN." >&2
  exit 1
fi
HOST="$CORPOS_GITFLOW_HOST"
API="https://$HOST/git/api/v1"

if [ -z "${CORPOS_GITFLOW_TOKEN:-}" ]; then
  cat >&2 <<'EOF'
ERROR: CORPOS_GITFLOW_TOKEN is not set.

corpos-gitflow is published as a release asset on a PRIVATE gitea repo, so
downloading it needs a token with read scope on that repo. In a workflow, add
one to the repository's Actions secrets and pass it through:

    env:
      CORPOS_GITFLOW_TOKEN: ${{ secrets.CORPOS_GITFLOW_TOKEN }}

Locally you do not need this script — build it from a the private toolkit service checkout
instead:

    make -C go corpos-gitflow-install
EOF
  exit 1
fi

api() { curl -sk -m 120 -H "Authorization: token $CORPOS_GITFLOW_TOKEN" "$@"; }

echo "[fetch-corpos-gitflow] resolving release '$TAG' on $REPO"
assets_json="$(api "$API/repos/$REPO/releases/tags/$TAG")"

read -r bin_url sum_url <<EOF
$(printf '%s' "$assets_json" | python3 -c '
import json, sys
d = json.load(sys.stdin)
if "assets" not in d:
    sys.exit("no assets on the release (is the publish job green?)")
by = {a["name"]: a.get("browser_download_url") or a.get("url") for a in d["assets"]}
name = "'"$ASSET"'"
missing = [n for n in (name, name + ".sha256") if not by.get(n)]
if missing:
    sys.exit("release is missing asset(s): " + ", ".join(missing))
print(by[name], by[name + ".sha256"])
')
EOF

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "[fetch-corpos-gitflow] downloading $ASSET"
api -L -o "$tmp/$ASSET" "$bin_url"
api -L -o "$tmp/$ASSET.sha256" "$sum_url"

echo "[fetch-corpos-gitflow] verifying checksum"
( cd "$tmp" && sha256sum -c "$ASSET.sha256" >/dev/null ) || {
  echo "ERROR: checksum mismatch on $ASSET — refusing to install it." >&2
  exit 1
}

chmod +x "$tmp/$ASSET"

reported="$("$tmp/$ASSET" version)" || {
  echo "ERROR: the downloaded binary did not run (wrong arch? truncated?)." >&2
  exit 1
}
echo "[fetch-corpos-gitflow] $reported"

if [ -n "${CORPOS_GITFLOW_SHA:-}" ]; then
  case "$reported" in
    *" $CORPOS_GITFLOW_SHA "*) : ;;
    *)
      echo "ERROR: expected build $CORPOS_GITFLOW_SHA, got: $reported" >&2
      exit 1
      ;;
  esac
  echo "[fetch-corpos-gitflow] pinned build $CORPOS_GITFLOW_SHA confirmed"
fi

mkdir -p "$DEST"
install -m 0755 "$tmp/$ASSET" "$DEST/corpos-gitflow"
echo "[fetch-corpos-gitflow] installed to $DEST/corpos-gitflow"

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "[fetch-corpos-gitflow] NOTE: $DEST is not on PATH; add it before running corpos-gitflow." ;;
esac
