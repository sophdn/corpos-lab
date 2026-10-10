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
# DURABLE PROGRESS LOG + DIED-VS-DONE SIGNAL
#   A long sweep run in the background can be killed with no warning — the
#   memory-guard SIGKILL that took a 20-cell grid after cell 12 was invisible until
#   someone read the run dirs by hand (bug
#   run-grid-dies-silently-with-no-durable-progress-log). So this runner leaves two
#   durable, self-describing traces beside the defs (or under --work-root):
#     * run-grid-progress.jsonl — an APPEND-ONLY log, one JSON line per event
#         (sweep_start / cell_start / cell_complete / cell_fail / sweep_end). Each
#         line is written with its own open→write→close and flushed, so it survives
#         the process being killed (even SIGKILL, which no trap can catch). A cell
#         with a cell_start and no matching cell_complete/cell_fail is a cell that
#         was interrupted mid-run.
#     * run-grid-status.json — a SENTINEL, rewritten at sweep start to
#         {"status":"running",...} and on exit to done / failed / interrupted with
#         the exit reason. Written from a bash trap on EXIT/INT/TERM. A SIGKILL
#         cannot run the trap, so the sentinel stays "running" while the pid is
#         gone — that mismatch is itself the "it died" signal.
#   Detect a died-vs-done run from either trace: a clean sweep ends with a
#   sweep_end line (reason=done|failed) and a non-"running" sentinel; a sweep that
#   died mid-run has a dangling cell_start (no terminal line for that cell) and a
#   sentinel still reading "running".
#
#   RESUME is driven by the durable log as well as the per-cell run-record: a cell
#   is skipped when its run-record says "completed" OR the log recorded a
#   cell_complete for it. A cell that only ever got a cell_start (killed mid-run) is
#   RE-RUN, never silently skipped.
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
#     --progress-log <f>  Path to the durable append-only progress log. Default:
#                         run-grid-progress.jsonl under --work-root, else the --defs
#                         directory, else the repo root. The status sentinel
#                         (run-grid-status.json) sits beside it.
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
#   130 interrupted by a signal (SIGINT/SIGTERM); a partial log is left behind

set -euo pipefail

DEFS=""
MODEL=""
CHUNK=0
WORK_ROOT=""
TOOLKIT_URL="__unset__"
PROGRESS_LOG_ARG=""
DRY_RUN=0

usage() { sed -n '2,79p' "$0"; }

die() { echo "run-grid: $*" >&2; exit 2; }

while [ $# -gt 0 ]; do
  case "$1" in
    --defs)         DEFS="${2:-}"; shift 2 ;;
    --model)        MODEL="${2:-}"; shift 2 ;;
    --chunk)        CHUNK="${2:-}"; shift 2 ;;
    --work-root)    WORK_ROOT="${2:-}"; shift 2 ;;
    --toolkit-url)  TOOLKIT_URL="${2:-}"; shift 2 ;;
    --progress-log) PROGRESS_LOG_ARG="${2:-}"; shift 2 ;;
    --dry-run)      DRY_RUN=1; shift ;;
    -h|--help)      usage; exit 0 ;;
    *) die "unknown arg: $1" ;;
  esac
done

[ -n "$DEFS" ] || die "--defs is required"
case "$CHUNK" in ''|*[!0-9]*) die "--chunk must be a non-negative integer, got '$CHUNK'" ;; esac

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# --- durable progress log + died-vs-done signal ----------------------------
# The log is append-only and self-describing; the sentinel is the quick "is the
# last sweep still running?" read. Both live beside the defs so a resume (or a
# human) finds them from --defs alone. See the header for the detection recipe.
if [ -n "$PROGRESS_LOG_ARG" ]; then
  PROGRESS_LOG="$PROGRESS_LOG_ARG"
elif [ -n "$WORK_ROOT" ]; then
  PROGRESS_LOG="${WORK_ROOT%/}/run-grid-progress.jsonl"
elif [ -d "$DEFS" ]; then
  PROGRESS_LOG="${DEFS%/}/run-grid-progress.jsonl"
else
  PROGRESS_LOG="${repo_root}/run-grid-progress.jsonl"
fi
SENTINEL="$(dirname "$PROGRESS_LOG")/run-grid-status.json"
mkdir -p "$(dirname "$PROGRESS_LOG")" 2>/dev/null || true

# state the traps and the run loop read/write
SWEEP_STARTED=0     # 1 once we log sweep_start — gates terminal records
SWEEP_SIGNAL=""     # set by the signal traps to the caught signal name
ran=0
failed=0
remaining="unknown"

# json_escape emits a value safe to drop between JSON double-quotes. Paths and
# study names here carry no control chars, so backslash + double-quote cover it.
json_escape() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  printf '%s' "$s"
}

# log_event <event> [key=value ...] — append ONE self-describing JSON line to the
# durable log and flush it. A fresh redirection per call means no in-process buffer
# can swallow the record if the process is killed; the fdatasync guards a crash.
log_event() {
  [ "$DRY_RUN" -eq 1 ] && return 0
  [ -n "$PROGRESS_LOG" ] || return 0
  local event="$1"; shift
  local ts json kv key val
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  json="$(printf '{"ts":"%s","event":"%s"' "$ts" "$event")"
  for kv in "$@"; do
    key="${kv%%=*}"; val="${kv#*=}"
    json="${json}$(printf ',"%s":"%s"' "$key" "$(json_escape "$val")")"
  done
  json="${json}}"
  printf '%s\n' "$json" >> "$PROGRESS_LOG"
  sync "$PROGRESS_LOG" 2>/dev/null || true
}

# write_sentinel <status> <reason> [exit_code] — overwrite the quick-glance status
# file. status is running | done | failed | interrupted.
write_sentinel() {
  [ "$DRY_RUN" -eq 1 ] && return 0
  [ -n "$SENTINEL" ] || return 0
  local status="$1" reason="$2" ec="${3:-}" ts
  ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '{"status":"%s","reason":"%s","exit_code":"%s","pid":"%s","ts":"%s","progress_log":"%s"}\n' \
    "$(json_escape "$status")" "$(json_escape "$reason")" "$(json_escape "$ec")" \
    "$$" "$ts" "$(json_escape "$PROGRESS_LOG")" > "$SENTINEL"
  sync "$SENTINEL" 2>/dev/null || true
}

# finish is the single EXIT trap: it cleans up the built binary AND leaves the
# terminal signal (a sweep_end line + a settled sentinel) so a reader can tell a
# clean finish from a mid-run death. It must NOT change the exit status.
# shellcheck disable=SC2317  # invoked indirectly via the EXIT trap
finish() {
  local ec=$?
  [ -n "${BUILT_BIN:-}" ] && rm -f "$BUILT_BIN"
  if [ "$SWEEP_STARTED" -eq 1 ]; then
    local status reason
    if [ -n "$SWEEP_SIGNAL" ]; then
      status="interrupted"; reason="signal:${SWEEP_SIGNAL}"
    elif [ "$ec" -eq 0 ]; then
      status="done"; reason="complete"
    else
      status="failed"; reason="exit:${ec}"
    fi
    log_event sweep_end reason="$reason" status="$status" exit_code="$ec" \
      ran="$ran" failed="$failed" remaining="$remaining"
    write_sentinel "$status" "$reason" "$ec"
  fi
  return 0
}

# A catchable signal (memory guard SIGTERM, Ctrl-C SIGINT) records the reason then
# exits, which runs finish. SIGKILL cannot be trapped — there the dangling
# cell_start and the still-"running" sentinel are the signal instead.
# shellcheck disable=SC2317  # invoked indirectly via signal traps
on_signal() {
  SWEEP_SIGNAL="$1"
  echo "run-grid: caught SIG$1 — leaving durable progress + sentinel and exiting" >&2
  exit 130
}
trap finish EXIT
trap 'on_signal INT' INT
trap 'on_signal TERM' TERM
trap 'on_signal HUP' HUP

# --- resolve the binary: a fresh build by default, so the grid never runs on a
# stale binary (the sister failure mode to suggestion 169). ------------------
BIN="${CORPOS_LAB_BIN:-}"
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

# log_completed is true when the durable log recorded a cell_complete for <name>.
# This makes resume driven by the durable log too, not only the per-cell record:
# either source alone is enough to skip a cell.
log_completed() {
  local name="$1"
  [ -f "$PROGRESS_LOG" ] || return 1
  grep -F "\"cell\":\"$name\"" "$PROGRESS_LOG" 2>/dev/null \
    | grep -Fq '"event":"cell_complete"'
}

# log_interrupted is true when the log shows more cell_start events for <name>
# than terminal (cell_complete + cell_fail) events — i.e. a start with no matching
# finish, the fingerprint of a cell killed mid-run. Such a cell is RE-RUN, and we
# say so in the plan rather than skipping it silently.
log_interrupted() {
  local name="$1" starts ends
  [ -f "$PROGRESS_LOG" ] || return 1
  starts="$(grep -F "\"cell\":\"$name\"" "$PROGRESS_LOG" 2>/dev/null \
    | grep -Fc '"event":"cell_start"' || true)"
  ends="$(grep -F "\"cell\":\"$name\"" "$PROGRESS_LOG" 2>/dev/null \
    | grep -Fc -e '"event":"cell_complete"' -e '"event":"cell_fail"' || true)"
  [ "${starts:-0}" -gt "${ends:-0}" ]
}

# --- classify every def ----------------------------------------------------
pending=()
completed=0
filtered=0
interrupted_seen=0
declare -A NAME_OF WORKDIR_OF INTERRUPTED_OF
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
  if is_completed "$wd" || log_completed "$name"; then
    completed=$((completed + 1))
  else
    pending+=("$f")
    if log_interrupted "$name"; then
      INTERRUPTED_OF["$f"]=1
      interrupted_seen=$((interrupted_seen + 1))
    fi
  fi
done

echo "run-grid: ${#defs[@]} defs — ${completed} completed, ${#pending[@]} pending, ${filtered} filtered out by --model '${MODEL}'"
if [ "$interrupted_seen" -gt 0 ]; then
  echo "run-grid: ${interrupted_seen} pending cell(s) were interrupted mid-run by an earlier sweep and will re-run (see $PROGRESS_LOG)"
fi

# --- dry run: show the plan and stop ---------------------------------------
if [ "$DRY_RUN" -eq 1 ]; then
  for f in "${pending[@]:-}"; do
    [ -n "$f" ] || continue
    if [ -n "${INTERRUPTED_OF[$f]:-}" ]; then
      echo "  PENDING  ${NAME_OF[$f]}  ($f)  [interrupted mid-run — will retry]"
    else
      echo "  PENDING  ${NAME_OF[$f]}  ($f)"
    fi
  done
  echo "run-grid: dry-run — nothing executed"
  exit 0
fi

# --- open the sweep: the first durable marker + the "running" sentinel -------
SWEEP_STARTED=1
log_event sweep_start defs="${#defs[@]}" completed="$completed" \
  pending="${#pending[@]}" filtered="$filtered" model="$MODEL" chunk="$CHUNK"
write_sentinel running started ""

# --- run the pending cells, up to the chunk --------------------------------
remaining="${#pending[@]}"
for f in "${pending[@]:-}"; do
  [ -n "$f" ] || continue
  if [ "$CHUNK" -gt 0 ] && [ "$ran" -ge "$CHUNK" ]; then
    break
  fi
  name="${NAME_OF[$f]}"
  wd="${WORKDIR_OF[$f]}"
  echo "run-grid: [$((ran + 1))] run-study $name ..."
  log_event cell_start cell="$name" def="$f" work_dir="$wd"
  args=("run-study" "$f" "-work" "$wd")
  [ "$TOOLKIT_URL" != "__unset__" ] && args+=("-toolkit-url" "$TOOLKIT_URL")
  if "$BIN" "${args[@]}"; then
    ran=$((ran + 1))
    log_event cell_complete cell="$name" work_dir="$wd"
  else
    cell_ec=$?
    ran=$((ran + 1))
    failed=$((failed + 1))
    log_event cell_fail cell="$name" work_dir="$wd" exit_code="$cell_ec"
    echo "run-grid: WARN cell FAILED: $name (record left in $wd; it will re-run next invocation)" >&2
  fi
  remaining=$(( ${#pending[@]} - ran ))
done

remaining=$(( ${#pending[@]} - ran ))
echo "run-grid: chunk done — ran ${ran} (${failed} failed), ${remaining} still pending"
if [ "$remaining" -gt 0 ]; then
  echo "run-grid: re-run the same command to continue; completed cells are skipped"
fi
[ "$failed" -eq 0 ] || exit 1
exit 0
