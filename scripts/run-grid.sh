#!/usr/bin/env bash
# scripts/run-grid.sh — resumable, chunked batch runner for a multi-hour grid of
# study definitions. (glyph-research chain study-chain-process-hardening,
# suggestion 179.)
#
# WHY THIS EXISTS
#   A grid is many study.toml files (one per cell — see studies/MATRIX_STUDIES.md).
#   Running them one-by-one is a babysitting exercise: a foreground tool timeout or
#   a background memory-guard kill loses the whole leg and forces a redo from
#   scratch. This runner makes a grid RESTARTABLE:
#     * idempotent  — a cell whose run-record already says "completed" is skipped,
#                     so a killed leg resumes with zero rework.
#     * chunked     — run at most N cells then exit, so an invocation stays under
#                     the foreground timeout and the memory guard; re-run to
#                     continue.
#     * model-filtered — run only the cells for one served model, because the one
#                     inference portal serves one model at a time (swap between
#                     legs, never a second server).
#     * dry-run     — report completed vs remaining without running anything.
#
#   It is a THIN wrapper over `corpos-lab run-study`. It changes nothing about the
#   assay, the provenance, or persistence — each cell is an ordinary run-study
#   invocation, recorded exactly as a hand-run cell would be.
#
# USAGE
#   scripts/run-grid.sh --defs <dir|glob> [options]
#
#     --defs <dir|glob>   Directory searched recursively for *.toml study defs, or
#                         a glob of def files. Required.
#     --model <substr>    Run only defs whose model_id contains <substr> (e.g.
#                         "primary", "Mistral"). Default: all models.
#     --chunk <N>         Run at most N pending cells this invocation, then exit.
#                         Default 0 = all remaining.
#     --work-root <dir>   Put each cell's run dir at <dir>/<study-name>. Default:
#                         mirror run-study — <def-dir>/runs/<study-name>.
#     --toolkit-url <url> Passed through to run-study. Empty string disables
#                         persistence. Default: run-study's own default.
#     --dry-run           Report the plan (completed / pending / filtered) and exit.
#     -h, --help          Show this help.
#
#   Env: CORPOS_LAB_BIN overrides the binary (default: a fresh `go build` of
#   ./cmd/corpos-lab, so the grid never runs on a stale binary).
#
# EXIT CODES
#   0  the chunk finished with no cell failing (grid may have cells remaining)
#   1  at least one cell failed this invocation
#   2  bad args / setup error

set -euo pipefail

DEFS=""
MODEL=""
CHUNK=0
WORK_ROOT=""
TOOLKIT_URL="__unset__"
DRY_RUN=0

usage() { sed -n '2,48p' "$0"; }

die() { echo "run-grid: $*" >&2; exit 2; }

while [ $# -gt 0 ]; do
  case "$1" in
    --defs)        DEFS="${2:-}"; shift 2 ;;
    --model)       MODEL="${2:-}"; shift 2 ;;
    --chunk)       CHUNK="${2:-}"; shift 2 ;;
    --work-root)   WORK_ROOT="${2:-}"; shift 2 ;;
    --toolkit-url) TOOLKIT_URL="${2:-}"; shift 2 ;;
    --dry-run)     DRY_RUN=1; shift ;;
    -h|--help)     usage; exit 0 ;;
    *) die "unknown arg: $1" ;;
  esac
done

[ -n "$DEFS" ] || die "--defs is required"
case "$CHUNK" in ''|*[!0-9]*) die "--chunk must be a non-negative integer, got '$CHUNK'" ;; esac

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# --- resolve the binary: a fresh build by default, so the grid never runs on a
# stale binary (the sister failure mode to suggestion 169). ------------------
BIN="${CORPOS_LAB_BIN:-}"
# shellcheck disable=SC2317  # invoked indirectly via the EXIT trap
# return 0 so a false final test does not become the script's exit status.
cleanup() { [ -n "${BUILT_BIN:-}" ] && rm -f "$BUILT_BIN"; return 0; }
trap cleanup EXIT
if [ -z "$BIN" ]; then
  BUILT_BIN="$(mktemp)"
  echo "run-grid: building corpos-lab from source ..." >&2
  ( cd "$repo_root" && go build -o "$BUILT_BIN" ./cmd/corpos-lab ) || die "go build failed"
  BIN="$BUILT_BIN"
fi

# --- enumerate the defs ----------------------------------------------------
defs=()
if [ -d "$DEFS" ]; then
  mapfile -t defs < <(find "$DEFS" -type f -name '*.toml' | sort)
else
  # Treat DEFS as a glob. Nullglob so a no-match is an empty list, not the
  # literal pattern.
  shopt -s nullglob
  # shellcheck disable=SC2206  # intentional word-split + glob expansion of DEFS
  defs=($DEFS)
  shopt -u nullglob
  if [ "${#defs[@]}" -gt 0 ]; then
    mapfile -t defs < <(printf '%s\n' "${defs[@]}" | sort)
  fi
fi
[ "${#defs[@]}" -gt 0 ] || die "no *.toml study defs found under: $DEFS"

# --- helpers ---------------------------------------------------------------
# study_field reads a top-level `key = "value"` string from a def file.
study_field() {
  local key="$1" file="$2"
  sed -n "s/^${key}[[:space:]]*=[[:space:]]*\"\\(.*\\)\"[[:space:]]*$/\\1/p" "$file" | head -1
}

# work_dir_for echoes the run dir a def resolves to, matching run-study's default.
work_dir_for() {
  local file="$1" name="$2"
  if [ -n "$WORK_ROOT" ]; then
    echo "${WORK_ROOT}/${name}"
  else
    echo "$(dirname "$file")/runs/${name}"
  fi
}

# is_completed is true when a cell's run-record exists AND records a completed
# run. A failed run also writes a run-record, so existence alone is not enough —
# a failed cell must re-run.
is_completed() {
  local wd="$1" rec="$1/run-record.json"
  [ -f "$rec" ] || return 1
  grep -Eq '"status"[[:space:]]*:[[:space:]]*"completed"' "$rec"
}

# --- classify every def ----------------------------------------------------
pending=()
completed=0
filtered=0
declare -A NAME_OF WORKDIR_OF
for f in "${defs[@]}"; do
  name="$(study_field name "$f")"
  [ -n "$name" ] || die "def has no name field: $f"
  model="$(study_field model_id "$f")"
  if [ -n "$MODEL" ] && [[ "$model" != *"$MODEL"* ]]; then
    filtered=$((filtered + 1))
    continue
  fi
  wd="$(work_dir_for "$f" "$name")"
  NAME_OF["$f"]="$name"
  WORKDIR_OF["$f"]="$wd"
  if is_completed "$wd"; then
    completed=$((completed + 1))
  else
    pending+=("$f")
  fi
done

echo "run-grid: ${#defs[@]} defs — ${completed} completed, ${#pending[@]} pending, ${filtered} filtered out by --model '${MODEL}'"

# --- dry run: show the plan and stop ---------------------------------------
if [ "$DRY_RUN" -eq 1 ]; then
  for f in "${pending[@]:-}"; do
    [ -n "$f" ] || continue
    echo "  PENDING  ${NAME_OF[$f]}  ($f)"
  done
  echo "run-grid: dry-run — nothing executed"
  exit 0
fi

# --- run the pending cells, up to the chunk --------------------------------
ran=0
failed=0
for f in "${pending[@]:-}"; do
  [ -n "$f" ] || continue
  if [ "$CHUNK" -gt 0 ] && [ "$ran" -ge "$CHUNK" ]; then
    break
  fi
  name="${NAME_OF[$f]}"
  wd="${WORKDIR_OF[$f]}"
  echo "run-grid: [$((ran + 1))] run-study $name ..."
  args=("run-study" "$f" "-work" "$wd")
  [ "$TOOLKIT_URL" != "__unset__" ] && args+=("-toolkit-url" "$TOOLKIT_URL")
  if "$BIN" "${args[@]}"; then
    ran=$((ran + 1))
  else
    ran=$((ran + 1))
    failed=$((failed + 1))
    echo "run-grid: WARN cell FAILED: $name (record left in $wd; it will re-run next invocation)" >&2
  fi
done

remaining=$(( ${#pending[@]} - ran ))
echo "run-grid: chunk done — ran ${ran} (${failed} failed), ${remaining} still pending"
if [ "$remaining" -gt 0 ]; then
  echo "run-grid: re-run the same command to continue; completed cells are skipped"
fi
[ "$failed" -eq 0 ] || exit 1
exit 0
