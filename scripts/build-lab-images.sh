#!/usr/bin/env bash
# Build the corpos-lab assay images with rootless podman and record their
# content digests. Base first, then every Containerfile.<variant> (globbed, so
# a new variant is zero-script-edit), then a report-mode smoke on each variant.
#
# This is NOT part of the pre-commit gate (an Ubuntu-free distroless build is
# still minutes and needs podman); it runs standalone and in CI. The Go gate
# (scripts/gate.sh) stays fast and hermetic.
#
#   scripts/build-lab-images.sh              build base + variants + smoke + record digests
#   scripts/build-lab-images.sh --digests    just print current image digests
#   scripts/build-lab-images.sh --check-pins verify the study scaffold's pinned probe
#                                            digests match the recorded build (no rebuild)
#   scripts/build-lab-images.sh --repin [toml...]
#                                            build, then re-pin the scaffold constants and
#                                            each named study TOML to the new digests
#   scripts/build-lab-images.sh --repin-only [toml...]
#                                            re-pin from the recorded digests (no rebuild)
#
# A re-pin rewrites every localhost/<image>@sha256:<digest> pin in the scaffold
# (internal/setupcompletion/studies.go) and in each named file, then runs
# --check-pins and the scaffold oracle test. Only named TOMLs change: a study that
# has already run keeps its literal pin, which is what makes its result reproducible.
# The oracle hash masks the image pins, so a re-pin never needs an oracle edit.
#
# A full build ends with the --check-pins check, so a probe rebuild that changes a
# digest fails loudly here until the study scaffold is re-pinned, rather than seeding
# new studies with a stale image (bug 1425 follow-up; suggestion
# drift-check-scaffold-probe-image-digests-against-latest-build).
set -euo pipefail
export PATH="$PATH:/usr/local/go/bin"

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

DEPLOY="$ROOT/deploy"
DIGEST_FILE="$DEPLOY/IMAGE_DIGESTS.txt"
BASE_TAG="lab-base:dev"

fail() { printf '[lab-images] FAIL: %b\n' "$*" >&2; exit 1; }

# The study-scaffold generator (internal/setupcompletion/studies.go) hard-codes a
# probe-image pin per constant and stamps it into every generated study. Nothing
# else ties those constants to the latest built image, so they drift silently when
# a probe rebuild records a new digest and no one re-pins the generator. SCAFFOLD_PINS
# maps each built image variant to the Go constant that must pin it. Extend this map
# when a new pinned image is added to studies.go.
SCAFFOLD="$ROOT/internal/setupcompletion/studies.go"
declare -A SCAFFOLD_PINS=(
    [lab-agentic-loop-probe]=LoopImg
    [lab-grounded-glyph-probe]=RawImg
)

# check_pins compares each scaffold constant's pinned digest against the digest the
# build recorded for the same image in DIGEST_FILE. It reads files only (no podman),
# so it is safe to run standalone after a re-pin. A drift, a missing pin, or a missing
# recorded digest fails loudly and names the constant to fix.
check_pins() {
    [ -f "$DIGEST_FILE" ] || fail "no $DIGEST_FILE — run scripts/build-lab-images.sh first to record the built digests"
    [ -f "$SCAFFOLD" ] || fail "study scaffold not found at $SCAFFOLD"
    local drift=0 image const pinned built
    for image in "${!SCAFFOLD_PINS[@]}"; do
        const="${SCAFFOLD_PINS[$image]}"
        pinned="$(sed -n "s|.*${image}@\(sha256:[0-9a-f]\{64\}\).*|\1|p" "$SCAFFOLD" | head -1)"
        built="$(awk -v n="${image}:dev" '$1==n{print $2}' "$DIGEST_FILE")"
        if [ -z "$pinned" ]; then
            printf '[lab-images]   check-pins: no %s pin for %s found in the scaffold\n' "$const" "$image" >&2
            drift=1
        elif [ -z "$built" ]; then
            printf '[lab-images]   check-pins: %s has no recorded digest in %s\n' "$image" "$DIGEST_FILE" >&2
            drift=1
        elif [ "$pinned" != "$built" ]; then
            printf '[lab-images]   check-pins: DRIFT %s\n    scaffold pins:  %s\n    built image is: %s\n' "$const" "$pinned" "$built" >&2
            drift=1
        else
            echo "[lab-images]   $const matches built $image ($built)"
        fi
    done
    [ "$drift" -eq 0 ] || fail "study-scaffold image pins have drifted from the built images.\n    Run scripts/build-lab-images.sh --repin-only [study TOMLs...] to re-pin them."
}

# repin rewrites the scaffold constants and each named file to the recorded
# digests (corpos-lab repin), then proves the result: the pins match the build and
# the scaffold oracle test passes.
repin() {
    echo "[lab-images] re-pinning the scaffold${*:+ and $# named file(s)} to the recorded digests…"
    go run ./cmd/corpos-lab repin -digests "$DIGEST_FILE" "$SCAFFOLD" "$@" || fail "repin"
    check_pins
    go test ./internal/setupcompletion/ || fail "scaffold oracle test after re-pin"
}

if [ "${1:-}" = "--check-pins" ]; then
    check_pins
    echo "[lab-images] check-pins PASS — the scaffold pins match the recorded build."
    exit 0
fi

if [ "${1:-}" = "--repin-only" ]; then
    shift
    repin "$@"
    echo "[lab-images] repin PASS"
    exit 0
fi

REPIN=0
if [ "${1:-}" = "--repin" ]; then
    REPIN=1
    shift
    REPIN_FILES=("$@")
fi

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

if [ "$REPIN" -eq 1 ]; then
    repin "${REPIN_FILES[@]}"
else
    echo "[lab-images] checking the study scaffold's image pins against the built digests…"
    check_pins
fi
echo "[lab-images] PASS"
