#!/usr/bin/env bash
# corpos-lab gate — the single "can't launch broken code" entrypoint.
#
# Wired as the pre-commit hook (via core.hooksPath=.githooks) AND runnable
# standalone as the CI-equivalent (same checks, outside the hook). Family
# pattern inherited from corpos: gofmt, vet, golangci-lint, govulncheck,
# build, race-tested tests, and a coverage floor on the logic packages.
#
# Dev tools are pinned by version and run via `go run <tool>@<ver>` so go.mod
# stays dependency-free and the versions are reproducible.
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

echo "[gate] 1/6 gofmt drift (whole tree)…"
# Check every .go file in the tree, not just tracked files — so the manual
# gate matches the commit-time hook even for not-yet-staged files.
drift="$(gofmt -s -l .)"
[ -z "$drift" ] || fail "gofmt drift in:\n$drift\n(run: gofmt -s -w .)"

echo "[gate] 2/6 go vet…"
go vet ./... || fail "go vet"

echo "[gate] 3/6 golangci-lint ${GOLANGCI_VERSION}…"
go run "github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_VERSION}" run ./... || fail "golangci-lint"

echo "[gate] 4/6 govulncheck ${GOVULNCHECK_VERSION}…"
go run "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}" ./... || fail "govulncheck"

echo "[gate] 5/6 go build…"
go build ./... || fail "go build"

echo "[gate] 6/6 go test -race + coverage floor…"
go test -race ./... || fail "go test -race"
# Coverage threshold is measured over the logic packages (./internal/...);
# cmd/ entrypoints are thin wrappers, integration-tested, not unit-covered.
go test -covermode=atomic -coverprofile=coverage.out ./internal/... >/dev/null || fail "coverage run"
cov="$(go tool cover -func=coverage.out | awk '/^total:/{gsub("%","",$3);print $3}')"
echo "[gate]   internal/ coverage: ${cov}% (min ${COVERAGE_MIN}%)"
awk -v c="$cov" -v m="$COVERAGE_MIN" 'BEGIN{exit !(c+0>=m+0)}' \
  || fail "coverage ${cov}% < ${COVERAGE_MIN}% floor"

echo "[gate] PASS — can't launch broken code."
