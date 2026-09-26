#!/usr/bin/env bash
# scripts/canonical-home-guard.sh — read-only pre-commit content guard.
#
# Refuses a commit that (a) stages a file under a declared PRIVATE SUBTREE, or
# (b) authors a load-shaped reference to another repo's CANONICAL DATA. It is
# the generalized successor to seed-packet's retired corpus-boundary-guard
# (deleted 2026-09-13 when lab-app dissolved): that guard hard-coded one path
# literal; this one reads its rules from a per-repo `.canonical-home` file, so
# the same mechanism protects any repo that declares a boundary.
#
# The guard is a no-op where no `.canonical-home` exists, so it is safe to wire
# into a repo before that repo declares any rule.
#
# Config: `.canonical-home` at the repo root. One rule per line; blank lines and
# lines beginning with `#` are ignored. Three rule verbs:
#
#   private-subtree <path-prefix>
#       Refuse any STAGED FILE whose repo-relative path starts with <path-prefix>.
#       Defends a gitignored private area (e.g. corpos-lab's corpus/private/) in a
#       PUBLIC repo against a stray `git add -f`. <path-prefix> is matched literally
#       against the path; end it with `/` to mean "the directory and everything under".
#
#   foreign-ref <literal> [ext1,ext2,...]
#       Refuse a STAGED FILE that contains a load-shaped reference to <literal> —
#       the literal appearing inside a double-quoted string ("<literal>). This is
#       the "authors another repo's canonical data" boundary: it blocks code/config
#       here from loading a sibling repo's canonical path. The optional comma-list
#       restricts the scan to those extensions (no dot); default: rs,toml,json,go.
#       Prose (.md) is never scanned — a mention is a cross-reference, not a load.
#
#   exclude <path-prefix>
#       A staged path under <path-prefix> is skipped by EVERY rule. Use it for the
#       guard's own test fixtures, which necessarily contain the literals above.
#
# Output: one line per violation on stderr, then exit 1. `--dry-run` reports but
# always exits 0. With no rules (or no config) the guard prints nothing and exits 0.
#
# Usage: scripts/canonical-home-guard.sh [--dry-run]
#        The set of files scanned is the staged set (git diff --cached), unless
#        CANONICAL_HOME_FILES is set to a newline-separated list (used by tests).
set -euo pipefail

DRY_RUN=0
[ "${1:-}" = "--dry-run" ] && DRY_RUN=1

REPO_ROOT="$(git rev-parse --show-toplevel)"
CONFIG="$REPO_ROOT/.canonical-home"

if [ ! -f "$CONFIG" ]; then
    exit 0
fi

# Parse the config into three parallel arrays.
private_prefixes=()
foreign_literals=()
foreign_exts=()
exclude_prefixes=()

while IFS= read -r line || [ -n "$line" ]; do
    # Strip a trailing comment and surrounding whitespace; skip blanks/comments.
    line="${line%%#*}"
    line="$(printf '%s' "$line" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"
    [ -z "$line" ] && continue
    verb="${line%%[[:space:]]*}"
    rest="$(printf '%s' "$line" | sed -E 's/^[^[:space:]]+[[:space:]]+//')"
    case "$verb" in
        private-subtree) private_prefixes+=("$rest") ;;
        exclude)         exclude_prefixes+=("$rest") ;;
        foreign-ref)
            literal="${rest%%[[:space:]]*}"
            exts="rs,toml,json,go"
            if [ "$rest" != "$literal" ]; then
                exts="$(printf '%s' "$rest" | sed -E 's/^[^[:space:]]+[[:space:]]+//')"
            fi
            foreign_literals+=("$literal")
            foreign_exts+=("$exts")
            ;;
        *)
            echo "[canonical-home] config error in $CONFIG: unknown rule '$verb'" >&2
            exit 2
            ;;
    esac
done < "$CONFIG"

# Nothing declared -> nothing to enforce.
if [ "${#private_prefixes[@]}" -eq 0 ] && [ "${#foreign_literals[@]}" -eq 0 ]; then
    exit 0
fi

# The file set: staged files by default, or an override list for tests.
if [ -n "${CANONICAL_HOME_FILES:-}" ]; then
    mapfile -t files <<< "$CANONICAL_HOME_FILES"
else
    mapfile -t files < <(git diff --cached --name-only --diff-filter=ACMR)
fi

is_excluded() {
    local path="$1" pfx
    for pfx in "${exclude_prefixes[@]:-}"; do
        [ -n "$pfx" ] || continue
        case "$path" in "$pfx"*) return 0 ;; esac
    done
    return 1
}

violations=0

for path in "${files[@]}"; do
    [ -n "$path" ] || continue
    is_excluded "$path" && continue

    # (a) private-subtree: path-based.
    for pfx in "${private_prefixes[@]:-}"; do
        [ -n "$pfx" ] || continue
        case "$path" in
            "$pfx"*)
                echo "CANONICAL-HOME VIOLATION $path: [private-subtree] staged file under private subtree '$pfx' — this data must not be committed to a public repo." >&2
                violations=$((violations + 1))
                ;;
        esac
    done

    # (b) foreign-ref: content-based, only for existing regular files.
    [ -f "$REPO_ROOT/$path" ] || continue
    ext="${path##*.}"
    i=0
    while [ "$i" -lt "${#foreign_literals[@]}" ]; do
        literal="${foreign_literals[$i]}"
        exts="${foreign_exts[$i]}"
        i=$((i + 1))
        # Does this file's extension fall in the rule's scan set?
        case ",$exts," in *",$ext,"*) ;; *) continue ;; esac
        # Load-shaped reference = the literal inside a double-quoted string.
        if grep -nF "\"$literal" "$REPO_ROOT/$path" >/dev/null 2>&1; then
            hit="$(grep -nF "\"$literal" "$REPO_ROOT/$path" | head -1)"
            lineno="${hit%%:*}"
            echo "CANONICAL-HOME VIOLATION $path:$lineno: [foreign-ref] load-shaped reference to another repo's canonical data '$literal' — replace with an API call or move the caller." >&2
            violations=$((violations + 1))
        fi
    done
done

if [ "$violations" -gt 0 ]; then
    if [ "$DRY_RUN" -eq 1 ]; then
        echo "[canonical-home] $violations violation(s) — dry-run, not blocking." >&2
        exit 0
    fi
    echo "[canonical-home] $violations violation(s) — commit blocked." >&2
    exit 1
fi
exit 0
