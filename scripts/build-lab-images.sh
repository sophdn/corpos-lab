#!/usr/bin/env bash
# Build the corpos-lab assay images with rootless podman and record their
# content digests. Base first, then every Containerfile.<variant> (globbed, so
# a new variant is zero-script-edit), then a report-mode smoke on each variant.
#
# This is NOT part of the pre-commit gate (an Ubuntu-free distroless build is
# still minutes and needs podman); it runs standalone and in CI. The Go gate
# (scripts/gate.sh) stays fast and hermetic.
#
#   scripts/build-lab-images.sh            build base + variants + smoke + record digests
#   scripts/build-lab-images.sh --digests  just print current image digests
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

DEPLOY="$ROOT/deploy"
DIGEST_FILE="$DEPLOY/IMAGE_DIGESTS.txt"
BASE_TAG="lab-base:dev"

fail() { printf '[lab-images] FAIL: %b\n' "$*" >&2; exit 1; }
command -v podman >/dev/null 2>&1 || fail "podman not found (rootless podman is mandatory)"

# Every Containerfile.<name> except .base is a variant.
mapfile -t VARIANTS < <(cd "$DEPLOY" && ls Containerfile.* 2>/dev/null | grep -v '\.base$' | sed 's/^Containerfile\.//')

record_digest() {
    local ref="$1"
    podman image inspect "$ref" --format '{{.Digest}}' 2>/dev/null
}

if [ "${1:-}" = "--digests" ]; then
    echo "[lab-images] current digests:"
    echo "  $BASE_TAG -> $(record_digest "$BASE_TAG")"
    for v in "${VARIANTS[@]}"; do
        echo "  lab-$v:dev -> $(record_digest "lab-$v:dev")"
    done
    exit 0
fi

echo "[lab-images] building $BASE_TAG …"
podman build -f "$DEPLOY/Containerfile.base" -t "$BASE_TAG" "$ROOT" || fail "base build"

# Assert non-root (distroless has no shell — read the image config).
user="$(podman image inspect "$BASE_TAG" --format '{{.Config.User}}')"
case "$user" in
    nonroot:nonroot|nonroot|65532*) echo "[lab-images]   base runs as non-root: $user" ;;
    ""|root|0|0:0) fail "base image runs as root (User=${user:-<empty>}) — non-root invariant broken" ;;
    *) echo "[lab-images]   WARN: unexpected base User=$user" ;;
esac

{
    echo "# corpos-lab assay image digests — recorded by scripts/build-lab-images.sh"
    echo "# Study records pin the variant digest; a verify mismatch is a hard refusal."
    echo "$BASE_TAG $(record_digest "$BASE_TAG")"
} > "$DIGEST_FILE"

for v in "${VARIANTS[@]}"; do
    ref="lab-$v:dev"
    echo "[lab-images] building $ref …"
    podman build -f "$DEPLOY/Containerfile.$v" -t "$ref" "$ROOT" || fail "$v build"

    echo "[lab-images]   smoke: $ref report"
    podman run --rm "$ref" report || fail "$v report-mode smoke"

    echo "$ref $(record_digest "$ref")" >> "$DIGEST_FILE"
done

echo "[lab-images] digests recorded -> $DIGEST_FILE"
cat "$DIGEST_FILE"
echo "[lab-images] PASS"
