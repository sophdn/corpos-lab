#!/usr/bin/env bash
# corpos-lab gate — the single "can't launch broken code" entrypoint.
#
# Wired as the pre-commit hook (via core.hooksPath=.githooks) AND runnable
# standalone as the CI-equivalent (same checks, outside the hook). Family
# pattern inherited from corpos: gofmt, vet, golangci-lint, govulncheck,
# build, race-tested tests, a coverage floor on the logic packages, and the
# language-policy check (Go primary, shell sanctioned per .language-policy.toml).
#
# Dev tools are pinned by version and run via `go run <tool>@<ver>` so go.mod
# stays dependency-free and the versions are reproducible. The one exception is
# the language-policy stage: langpolicy lives in the private toolkit service's corpos-gate, so
# that stage builds it from a the private toolkit service checkout (as CI does) and skips with
# a warning when no checkout is reachable.
#
# No image stage yet — assay-container builds join the gate when the
# container-assay-model task lands them (they will be digest-pinned; see
# CLAUDE.md invariants).
set -euo pipefail
export PATH="$PATH:/usr/local/go/bin"

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

COVERAGE_MIN=95
GOLANGCI_VERSION=v1.62.2
GOVULNCHECK_VERSION=v1.3.0

fail() { printf '[gate] FAIL: %b\n' "$*" >&2; exit 1; }

echo "[gate] 1/7 gofmt drift (whole tree)…"
# Check every .go file in the tree, not just tracked files — so the manual
# gate matches the commit-time hook even for not-yet-staged files.
drift="$(gofmt -s -l .)"
[ -z "$drift" ] || fail "gofmt drift in:\n$drift\n(run: gofmt -s -w .)"

echo "[gate] 2/7 go vet…"
go vet ./... || fail "go vet"

echo "[gate] 3/7 golangci-lint ${GOLANGCI_VERSION}…"
go run "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}" run ./... || fail "golangci-lint"

echo "[gate] 4/7 govulncheck ${GOVULNCHECK_VERSION}…"
go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" ./... || fail "govulncheck"

echo "[gate] 5/7 go build…"
go build ./... || fail "go build"

echo "[gate] 6/7 go test -race + coverage floor…"
go test -race ./... || fail "go test -race"
# Coverage threshold is measured over the logic packages (./internal/...);
# cmd/ entrypoints are thin wrappers, integration-tested, not unit-covered.
go test -covermode=atomic -coverprofile=coverage.out ./internal/... >/dev/null || fail "coverage run"
cov="$(go tool cover -func=coverage.out | awk '/^total:/{gsub("%","",$3);print $3}')"
echo "[gate]   internal/ coverage: ${cov}% (min ${COVERAGE_MIN}%)"
awk -v c="$cov" -v m="$COVERAGE_MIN" 'BEGIN{exit !(c+0>=m+0)}' \
  || fail "coverage ${cov}% < ${COVERAGE_MIN}% floor"

echo "[gate] 7/7 language policy (Go primary; shell sanctioned per .language-policy.toml)…"
# langpolicy lives in the private toolkit service's corpos-gate, not this repo, so — like the CI
# language-policy job — build it from a the private toolkit service checkout and run the SAME
# `langpolicy --all` over this tree. A breach then fails here at commit time instead of
# only turning CI red later. The installed corpos-gate on PATH is deliberately NOT used:
# it can lag the source and miss the --all whole-tree flag. When no the private toolkit service checkout
# is reachable (a fresh clone, or CI's own gate job), the stage warns and skips rather than
# failing — CI's dedicated language-policy job still enforces it there.
lp_src="${CORPOS_TOOLKIT_SRC:-}"
if [ -z "$lp_src" ]; then
  for cand in "$ROOT/../the private toolkit service" "$HOME/dev/the private toolkit service"; do
    if [ -d "$cand/go/cmd/corpos-gate" ]; then lp_src="$cand"; break; fi
  done
fi
if [ -n "$lp_src" ] && [ -d "$lp_src/go/cmd/corpos-gate" ]; then
  lp_bin="$(mktemp -d)/corpos-gate"
  ( cd "$lp_src/go" && go build -o "$lp_bin" ./cmd/corpos-gate ) \
    || fail "building corpos-gate from $lp_src (language-policy stage)"
  "$lp_bin" langpolicy --all --root "$ROOT" \
    || fail "language policy — a non-Go, non-sanctioned file is tracked (see .language-policy.toml)"
else
  printf '[gate]   language policy SKIPPED — no the private toolkit service checkout found (set CORPOS_TOOLKIT_SRC, or place it at ../the private toolkit service or ~/dev/the private toolkit service). CI still enforces it.\n' >&2
fi

echo "[gate] PASS — can't launch broken code."
