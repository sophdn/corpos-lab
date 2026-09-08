#!/usr/bin/env bash
# scripts/test-gitea-pr-land.sh — hermetic regression net for
# scripts/gitea-pr-land.sh, the gitea landing handshake (wait for the PR's
# REQUIRED CI checks, then merge, with a retry).
#
# Bug 1224 (worktree-merge-ci-poll-too-short-under-runner-contention-and-stale-
# queued-run-holds-combined-status). The handshake used to wait on the COMBINED
# commit status, which is strictly stricter than gitea's own merge gate: gitea
# merges on the contexts named in branch_protections.status_check_contexts
# alone, so any unrelated queued run (ci / mirror-publish, a stale push-event
# run) pinned the combined state at "pending" and the landing had to be
# finished by hand. Scenario 1 is that exact shape and is the assertion that
# fails against the old combined-status logic.
#
# Hermetic: every gitea endpoint is served by a throwaway local HTTP stub. No
# network, no repo state, no remote. Exit 0 on all-pass, 1 on any failure.

set -uo pipefail

GITFLOW_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELPER="$GITFLOW_DIR/gitea-pr-land.sh"
PASS=0
FAIL=0
SCRATCH=()
STUBS=()

cleanup() {
    for p in "${STUBS[@]:-}"; do [[ -n "$p" ]] && kill "$p" 2>/dev/null; done
    for d in "${SCRATCH[@]:-}"; do [[ -n "$d" && -d "$d" ]] && rm -rf "$d"; done
}
trap cleanup EXIT

assert() {
    local desc="$1"; shift
    if "$@"; then echo "  PASS  $desc"; PASS=$((PASS + 1));
    else echo "  FAIL  $desc"; FAIL=$((FAIL + 1)); fi
}

# start_gitea_stub <dir> — serves the gitea endpoints the handshake touches,
# reading its payloads from <dir> so each scenario can shape the responses:
#
#   protections.json   GET  …/branch_protections/<branch>   (absent -> 404)
#   statuses.json      GET  …/statuses/<sha>
#   statuses2.json     …served from the 2nd statuses request on (optional,
#                      for "the required check goes green while we poll")
#   combined.json      GET  …/commits/<sha>/status
#   merge_code         POST …/pulls/<n>/merge   (HTTP code to return)
#   merge_code2        …returned from the 2nd merge attempt on (optional)
#
# and records every merge attempt in <dir>/merge_calls.
#
# The stub program is written out by new_stub_dir and run by start_gitea_stub,
# which prints the base URL.
_stub_py() { cat <<'PY'
import http.server, socketserver, sys, os

d = sys.argv[1]

def read(name, default=None):
    p = os.path.join(d, name)
    if not os.path.exists(p):
        return default
    with open(p, "rb") as f:
        return f.read()

def bump(name):
    p = os.path.join(d, name)
    with open(p, "a") as f:
        f.write("x\n")
    with open(p) as f:
        return len(f.read().split())

class H(http.server.BaseHTTPRequestHandler):
    def reply(self, code, body):
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        p = self.path
        if "/branch_protections/" in p:
            body = read("protections.json")
            if body is None:
                return self.reply(404, b'{"errors":["not found"],"message":"branch protection does not exist"}')
            return self.reply(200, body)
        if "/statuses/" in p:
            n = bump("statuses_calls")
            body = read("statuses2.json") if n >= 2 and read("statuses2.json") else read("statuses.json", b"[]")
            return self.reply(200, body)
        if p.endswith("/status"):
            return self.reply(200, read("combined.json", b'{"state":"pending"}'))
        # GET …/pulls/<n>  (the single-PR object pr_is_merged reads). NOT the
        # list, and NOT the /merge POST. Default: not merged.
        if "/pulls/" in p and not p.endswith("/merge"):
            return self.reply(200, read("pull.json", b'{"merged":false}'))
        return self.reply(200, b"[]")

    def do_POST(self):
        if "/merge" in self.path:
            n = bump("merge_calls")
            code = read("merge_code", b"200").decode().strip()
            second = read("merge_code2")
            if n >= 2 and second:
                code = second.decode().strip()
            code = int(code)
            return self.reply(code, b"{}" if code == 200 else b'{"message":"Please try again later"}')
        return self.reply(200, b"{}")

    def log_message(self, *a):
        pass

srv = socketserver.TCPServer(("127.0.0.1", 0), H)
with open(os.path.join(d, "port"), "w") as f:
    f.write(str(srv.server_address[1]))
srv.serve_forever()
PY
}

# new_stub_dir prints a scratch dir preloaded with the stub program.
new_stub_dir() {
    local d; d="$(mktemp -d)"; SCRATCH+=("$d")
    _stub_py > "$d/stub.py"
    echo "$d"
}

start_gitea_stub() {
    local d="$1"
    : > "$d/merge_calls"
    : > "$d/statuses_calls"
    python3 "$d/stub.py" "$d" >/dev/null 2>&1 < /dev/null &
    STUBS+=("$!")
    local _
    for _ in $(seq 1 50); do [[ -s "$d/port" ]] && break; sleep 0.1; done
    [[ -s "$d/port" ]] || { echo "stub failed to start" >&2; return 1; }
    echo "http://127.0.0.1:$(cat "$d/port")"
}

SHA=0123456789abcdef0123456789abcdef01234567

# run_land <stub-dir> <base-url> [extra env assignments…] — drives the helper
# with fast timings so a scenario never costs more than a second.
run_land() {
    local d="$1" base="$2"; shift 2
    env GITEA_API="$base" GITEA_OWNER=owner GITEA_REPO=repo GITEA_TOKEN=stub \
        GITEA_CI_TIMEOUT="${GITEA_CI_TIMEOUT:-1}" \
        GITEA_CI_INTERVAL="${GITEA_CI_INTERVAL:-1}" \
        GITEA_MERGE_RETRY_DELAY="${GITEA_MERGE_RETRY_DELAY:-0}" \
        "$@" \
        "$HELPER" --pr 99 --sha "$SHA" --base main > "$d/out.log" 2>&1
}

echo "── Scenario 1: a queued NON-required run must not pin the landing (bug 1224) ──"
# The exact live shape: branch protection requires 'ci / precommit*', which is
# green, while 'ci / mirror-publish (push)' is still queued — so the COMBINED
# status is pending. Gitea itself merges here (verified live on PR #24), so the
# handshake must too, and must not burn the deadline first.
d1="$(new_stub_dir)"
cat > "$d1/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d1/statuses.json" <<'JSON'
[{"id":3,"context":"ci / mirror-publish (push)","status":"pending"},
 {"id":2,"context":"ci / precommit (push)","status":"success"}]
JSON
echo '{"state":"pending"}' > "$d1/combined.json"
B1="$(start_gitea_stub "$d1")"
start=$SECONDS
rc=0; GITEA_CI_TIMEOUT=30 run_land "$d1" "$B1" || rc=$?
elapsed=$((SECONDS - start))
sed 's/^/    │ /' "$d1/out.log"
assert "required-green: exits 0" test "$rc" -eq 0
assert "required-green: merged the PR" grep -q "merged" "$d1/out.log"
assert "required-green: merged on the FIRST attempt" test "$(wc -w < "$d1/merge_calls")" -eq 1
assert "required-green: did not wait out the deadline" test "$elapsed" -lt 10
assert "required-green: names the required context it waited on" grep -q "ci / precommit\*" "$d1/out.log"

echo "── Scenario 2: a pending REQUIRED context is waited on, then merges ──"
d2="$(new_stub_dir)"
cat > "$d2/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d2/statuses.json" <<'JSON'
[{"id":1,"context":"ci / precommit (push)","status":"pending"}]
JSON
cat > "$d2/statuses2.json" <<'JSON'
[{"id":2,"context":"ci / precommit (push)","status":"success"}]
JSON
echo '{"state":"pending"}' > "$d2/combined.json"
B2="$(start_gitea_stub "$d2")"
rc=0; GITEA_CI_TIMEOUT=30 GITEA_CI_INTERVAL=1 run_land "$d2" "$B2" || rc=$?
sed 's/^/    │ /' "$d2/out.log"
assert "required-pending: exits 0 once it goes green" test "$rc" -eq 0
assert "required-pending: polled more than once" test "$(wc -w < "$d2/statuses_calls")" -ge 2
assert "required-pending: merged the PR" test "$(wc -w < "$d2/merge_calls")" -eq 1

echo "── Scenario 3: a FAILED required context refuses the merge ──"
d3="$(new_stub_dir)"
cat > "$d3/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d3/statuses.json" <<'JSON'
[{"id":1,"context":"ci / precommit (push)","status":"failure"}]
JSON
echo '{"state":"success"}' > "$d3/combined.json"
B3="$(start_gitea_stub "$d3")"
rc=0; run_land "$d3" "$B3" || rc=$?
sed 's/^/    │ /' "$d3/out.log"
assert "required-failed: exits 1" test "$rc" -eq 1
assert "required-failed: never attempted the merge" test "$(wc -w < "$d3/merge_calls")" -eq 0
assert "required-failed: says CI failed" grep -qi "ci failed" "$d3/out.log"

echo "── Scenario 4: a rejected merge is retried at least once (bug 1224 AC) ──"
d4="$(new_stub_dir)"
cat > "$d4/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d4/statuses.json" <<'JSON'
[{"id":1,"context":"ci / precommit (push)","status":"success"}]
JSON
echo '405' > "$d4/merge_code"
echo '200' > "$d4/merge_code2"
B4="$(start_gitea_stub "$d4")"
rc=0; run_land "$d4" "$B4" || rc=$?
sed 's/^/    │ /' "$d4/out.log"
assert "merge-retry: exits 0 after the retry succeeds" test "$rc" -eq 0
assert "merge-retry: attempted the merge twice" test "$(wc -w < "$d4/merge_calls")" -eq 2
assert "merge-retry: says it is retrying" grep -qi "retry" "$d4/out.log"

echo "── Scenario 5: a merge that never succeeds still exits 1, loudly ──"
d5="$(new_stub_dir)"
cat > "$d5/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d5/statuses.json" <<'JSON'
[{"id":1,"context":"ci / precommit (push)","status":"success"}]
JSON
echo '405' > "$d5/merge_code"
B5="$(start_gitea_stub "$d5")"
rc=0; run_land "$d5" "$B5" || rc=$?
sed 's/^/    │ /' "$d5/out.log"
assert "merge-stuck: exits 1" test "$rc" -eq 1
assert "merge-stuck: retried before giving up" test "$(wc -w < "$d5/merge_calls")" -ge 2
assert "merge-stuck: reports the HTTP code" grep -q "405" "$d5/out.log"

echo "── Scenario 6: no branch protection falls back to the combined status ──"
# Backwards compatibility: a repo with no protection rule (the stub 404s
# branch_protections) must keep the old, stricter behavior rather than merging
# blind.
d6="$(new_stub_dir)"
cat > "$d6/statuses.json" <<'JSON'
[{"id":1,"context":"ci / whatever","status":"pending"}]
JSON
echo '{"state":"success"}' > "$d6/combined.json"
B6="$(start_gitea_stub "$d6")"
rc=0; run_land "$d6" "$B6" || rc=$?
sed 's/^/    │ /' "$d6/out.log"
assert "no-protection: exits 0 on a green combined status" test "$rc" -eq 0
assert "no-protection: says it fell back" grep -qi "combined" "$d6/out.log"
assert "no-protection: merged the PR" test "$(wc -w < "$d6/merge_calls")" -eq 1

echo "── Scenario 7: a required context that never reports is pending, not success ──"
# The glob matches nothing in /statuses — the check has not started. Treating
# "absent" as success would merge before CI ran at all.
d7="$(new_stub_dir)"
cat > "$d7/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d7/statuses.json" <<'JSON'
[{"id":1,"context":"ci / mirror-publish (push)","status":"success"}]
JSON
echo '{"state":"success"}' > "$d7/combined.json"
B7="$(start_gitea_stub "$d7")"
start=$SECONDS
rc=0; GITEA_CI_TIMEOUT=2 GITEA_CI_INTERVAL=1 run_land "$d7" "$B7" || rc=$?
elapsed=$((SECONDS - start))
sed 's/^/    │ /' "$d7/out.log"
assert "absent-required: waited rather than merging immediately" test "$elapsed" -ge 2
assert "absent-required: warns the deadline expired" grep -qi "still pending after" "$d7/out.log"

echo "── Scenario 8: protection with status checks disabled merges immediately ──"
d8="$(new_stub_dir)"
cat > "$d8/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":false,"status_check_contexts":[]}
JSON
echo '{"state":"pending"}' > "$d8/combined.json"
B8="$(start_gitea_stub "$d8")"
start=$SECONDS
rc=0; GITEA_CI_TIMEOUT=30 run_land "$d8" "$B8" || rc=$?
elapsed=$((SECONDS - start))
sed 's/^/    │ /' "$d8/out.log"
assert "no-checks-required: exits 0" test "$rc" -eq 0
assert "no-checks-required: did not wait" test "$elapsed" -lt 10
assert "no-checks-required: merged the PR" test "$(wc -w < "$d8/merge_calls")" -eq 1

echo "── Scenario 9: an unreachable API exits 1 rather than merging blind ──"
d9="$(new_stub_dir)"
rc=0; GITEA_CI_TIMEOUT=2 GITEA_CI_INTERVAL=1 run_land "$d9" "http://127.0.0.1:1" || rc=$?
sed 's/^/    │ /' "$d9/out.log"
assert "unreachable: exits 1" test "$rc" -eq 1

echo "── Scenario 10: null per-check state + success status merges (locks the field) ──"
# The exact live shape on this Gitea: every status entry carries a populated
# `status` but a NULL per-check `state`. The handshake keys on `.status`, so it
# must merge. This fixture pins that field choice against a future edit that
# switches to the null `.state` and reintroduces the silent hang.
d10="$(new_stub_dir)"
cat > "$d10/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d10/statuses.json" <<'JSON'
[{"id":2,"context":"ci / precommit (push)","status":"success","state":null}]
JSON
B10="$(start_gitea_stub "$d10")"
rc=0; run_land "$d10" "$B10" || rc=$?
sed 's/^/    │ /' "$d10/out.log"
assert "null-state: exits 0" test "$rc" -eq 0
assert "null-state: merged the PR" grep -q "merged" "$d10/out.log"
assert "null-state: merged on the first attempt" test "$(wc -w < "$d10/merge_calls")" -eq 1

echo "── Scenario 11: ALL matching event-variants must pass before merge (strict) ──"
# ci / precommit (pull_request) is green but ci / precommit (push) is still
# pending. Gitea requires EVERY matching context, so the handshake must WAIT,
# not merge early — merging early returns 405 "Not all required status checks
# successful" (the 2026-09-07 PR #73 regression). When the push variant goes
# green on the next poll (statuses2), it merges.
d11="$(new_stub_dir)"
cat > "$d11/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d11/statuses.json" <<'JSON'
[{"id":3,"context":"ci / precommit (pull_request)","status":"success","state":null},
 {"id":2,"context":"ci / precommit (push)","status":"pending","state":null}]
JSON
cat > "$d11/statuses2.json" <<'JSON'
[{"id":5,"context":"ci / precommit (pull_request)","status":"success","state":null},
 {"id":4,"context":"ci / precommit (push)","status":"success","state":null}]
JSON
B11="$(start_gitea_stub "$d11")"
rc=0; GITEA_CI_TIMEOUT=30 GITEA_CI_INTERVAL=1 run_land "$d11" "$B11" || rc=$?
sed 's/^/    │ /' "$d11/out.log"
assert "strict-multi-variant: exits 0 once BOTH variants are green" test "$rc" -eq 0
assert "strict-multi-variant: waited (polled more than once)" test "$(wc -w < "$d11/statuses_calls")" -ge 2
assert "strict-multi-variant: merged the PR exactly once" test "$(wc -w < "$d11/merge_calls")" -eq 1

echo "── Scenario 12: an already-merged PR (405 on merge) is success, not failure ──"
# A human or a concurrent run merged the PR first; the merge POST then returns a
# non-200 (observed HTTP 405). pr_is_merged confirms it and the handshake exits 0.
d12="$(new_stub_dir)"
cat > "$d12/protections.json" <<'JSON'
{"branch_name":"main","enable_status_check":true,"status_check_contexts":["ci / precommit*"]}
JSON
cat > "$d12/statuses.json" <<'JSON'
[{"id":1,"context":"ci / precommit (push)","status":"success","state":null}]
JSON
echo '405' > "$d12/merge_code"
echo '{"merged":true}' > "$d12/pull.json"
B12="$(start_gitea_stub "$d12")"
rc=0; run_land "$d12" "$B12" || rc=$?
sed 's/^/    │ /' "$d12/out.log"
assert "already-merged: exits 0" test "$rc" -eq 0
assert "already-merged: says already merged" grep -qi "already merged" "$d12/out.log"

echo "── Scenario 13: a PR already merged before we poll short-circuits ──"
# pr_is_merged is checked up front, so a re-run against an already-merged PR
# exits immediately without polling CI or attempting a merge.
d13="$(new_stub_dir)"
echo '{"merged":true}' > "$d13/pull.json"
B13="$(start_gitea_stub "$d13")"
rc=0; run_land "$d13" "$B13" || rc=$?
sed 's/^/    │ /' "$d13/out.log"
assert "pre-merged: exits 0" test "$rc" -eq 0
assert "pre-merged: says nothing to do" grep -qi "nothing to do" "$d13/out.log"
assert "pre-merged: never attempted a merge" test "$(wc -w < "$d13/merge_calls")" -eq 0

echo ""
echo "gitea-pr-land: $PASS pass, $FAIL fail"
[[ "$FAIL" -eq 0 ]]
