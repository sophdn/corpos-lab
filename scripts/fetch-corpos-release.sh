#!/usr/bin/env bash
# scripts/fetch-corpos-release.sh — download a published corpos binary
# (corpos-gate or corpos-gitflow) and put it on PATH. The consumer side of
# scripts/publish-corpos-gate.sh and the git-flow binary-fetch model; called
# through the thin scripts/fetch-corpos-gate.sh and scripts/fetch-corpos-gitflow.sh.
#
# CANONICAL in the private toolkit service and copied into each consuming repo (a consumer's
# CI cannot reach this tree, which is the problem it solves). Edit it in
# the private toolkit service and run scripts/sync-distributed-scripts.sh; the
# distributed-scripts-drift guard fails while a sibling checkout's copy differs.
#
# Usage:
#   scripts/fetch-corpos-release.sh <corpos-gate|corpos-gitflow> [dest_dir]
#                                                   # default: $HOME/.local/bin
#
# Env (PFX is CORPOS_GATE or CORPOS_GITFLOW, from the tool name):
#   PFX_TOKEN   gitea token with READ scope on the toolkit repo (required)
#   PFX_HOST    gitea host, e.g. git.example.internal (required)
#   PFX_REPO    owner/repo to fetch from (default: shared/the private toolkit service)
#   PFX_SHA     if set, REQUIRE the downloaded binary to report this SHA
#
# PFX_HOST has no default ON PURPOSE. Baking this fleet's gitea hostname into
# the script would put real infrastructure into a repo that publishes to a
# public mirror — the pii-scan gate rejected exactly that on the first cut, and
# it was right to. A consumer's workflow supplies the host alongside the token,
# which also makes this script portable to any gitea rather than to one box.
#
# WHY A TOKEN IS REQUIRED. Every repo on this gitea instance is private, and an
# unauthenticated fetch fails asking for a username. Publishing the binary as a
# release asset removes the TOOLCHAIN from a consumer's CI (no Go, no module
# downloads, no CGO build); it does not remove the authentication. A consuming
# workflow needs a read-scoped token in its Actions secrets.
#
# The download is checksum-verified and then executed to confirm it reports a
# version, so a truncated or wrong-arch artifact fails HERE with a clear message
# rather than later with a confusing one.

set -euo pipefail

TOOL="${1:-}"
case "$TOOL" in
  corpos-gate | corpos-gitflow) ;;
  *)
    echo "usage: fetch-corpos-release.sh <corpos-gate|corpos-gitflow> [dest_dir]" >&2
    exit 2
    ;;
esac
PFX="$(printf '%s' "$TOOL" | tr '[:lower:]-' '[:upper:]_')"
LOG="[fetch-$TOOL]"

DEST="${2:-$HOME/.local/bin}"
repo_var="${PFX}_REPO"
host_var="${PFX}_HOST"
token_var="${PFX}_TOKEN"
sha_var="${PFX}_SHA"
REPO="${!repo_var:-shared/the private toolkit service}"
TAG="$TOOL-latest"
ASSET="$TOOL-linux-amd64"

if [ -z "${!host_var:-}" ]; then
  echo "ERROR: $host_var is not set (e.g. git.example.internal)." >&2
  echo "  It has no default so this repo carries no infrastructure hostname;" >&2
  echo "  set it in the consuming workflow next to $token_var." >&2
  exit 1
fi
API="https://${!host_var}/git/api/v1"

if [ -z "${!token_var:-}" ]; then
  cat >&2 <<EOF
ERROR: $token_var is not set.

$TOOL is published as a release asset on a PRIVATE gitea repo, so
downloading it needs a token with read scope on that repo. In a workflow, add
one to the repository's Actions secrets and pass it through:

    env:
      $token_var: \${{ secrets.$token_var }}

Locally you do not need this script at all — build it from a the private toolkit service
checkout instead:

    make -C go $TOOL-install
EOF
  exit 1
fi
TOKEN="${!token_var}"

api() { curl -sk -m 120 -H "Authorization: token $TOKEN" "$@"; }

echo "$LOG resolving release '$TAG' on $REPO"
assets_json="$(api "$API/repos/$REPO/releases/tags/$TAG")"

# A checked substitution: a release without the asset or its checksum stops
# here, instead of running on to a download of an empty URL.
urls="$(printf '%s' "$assets_json" | python3 -c '
import json, sys
d = json.load(sys.stdin)
if "assets" not in d:
    sys.exit("no assets on the release (is the publish job green?)")
by = {a["name"]: a.get("browser_download_url") or a.get("url") for a in d["assets"]}
name = sys.argv[1]
missing = [n for n in (name, name + ".sha256") if not by.get(n)]
if missing:
    sys.exit("release is missing asset(s): " + ", ".join(missing))
print(by[name], by[name + ".sha256"])
' "$ASSET")" || exit 1
read -r bin_url sum_url <<<"$urls"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "$LOG downloading $ASSET"
api -L -o "$tmp/$ASSET" "$bin_url"
api -L -o "$tmp/$ASSET.sha256" "$sum_url"

echo "$LOG verifying checksum"
(cd "$tmp" && sha256sum -c "$ASSET.sha256" >/dev/null) || {
  echo "ERROR: checksum mismatch on $ASSET — refusing to install it." >&2
  exit 1
}

chmod +x "$tmp/$ASSET"

# Execute it. A wrong-arch or truncated binary fails here, at fetch time, with
# a clear message — not later, mid-gate, with a confusing one.
reported="$("$tmp/$ASSET" version)" || {
  echo "ERROR: the downloaded binary did not run (wrong arch? truncated?)." >&2
  exit 1
}
echo "$LOG $reported"

if [ -n "${!sha_var:-}" ]; then
  case "$reported" in
    *" ${!sha_var} "*) : ;;
    *)
      echo "ERROR: expected build ${!sha_var}, got: $reported" >&2
      exit 1
      ;;
  esac
  echo "$LOG pinned build ${!sha_var} confirmed"
fi

mkdir -p "$DEST"
install -m 0755 "$tmp/$ASSET" "$DEST/$TOOL"
echo "$LOG installed to $DEST/$TOOL"

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "$LOG NOTE: $DEST is not on PATH; add it before running $TOOL." ;;
esac
