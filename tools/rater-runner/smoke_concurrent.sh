#!/usr/bin/env bash
# End-to-end smoke for the hardened blind-rating pathway (chain
# blind-rating-pathway-hardening, task verify-pathway-end-to-end).
#
# It proves three things on a REAL slice, without running a model grid:
#   1. rate.py runs to completion when invoked with RELATIVE --slices-dir /
#      --out-dir and a relative script in --rater-cmd (bug 1338).
#   2. Two raters run CONCURRENTLY on the same slice with no shared-scratch
#      collision. The mechanical rater deliberately writes a fixed-name scratch
#      file (all_items.txt — the chain 549 hazard) into its cwd and checks it is
#      still its own; rate.py's isolated scratch cwd keeps it private.
#   3. Each output is valid {id: code} JSON whose keys are exactly the slice ids.
#
# The rater here is a deterministic mechanical stand-in, NOT a real scorer. Its
# codes are meaningless; the point is the pathway, not the scores. All output goes
# to a throwaway temp workspace that is removed on exit — nothing is written into
# the repo or a real scores/ dir.
#
# Usage: smoke_concurrent.sh [path/to/slice.jsonl]
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RATE="$HERE/rate.py"
REPO="$(git -C "$HERE" rev-parse --show-toplevel)"
SLICE="${1:-$REPO/studies/grounded-non-prescriptive-aid/scoring/slices/governed-operation-protocol-bypass__00.jsonl}"

[ -f "$SLICE" ] || { echo "smoke: slice not found: $SLICE" >&2; exit 2; }

WORK="$(mktemp -d "${TMPDIR:-/tmp}/rater-smoke-XXXXXX")"
trap 'rm -rf "$WORK"' EXIT

mkdir -p "$WORK/slices"
STEM="$(basename "$SLICE" .jsonl)"
cp "$SLICE" "$WORK/slices/$STEM.jsonl"

# Mechanical rater with a deliberate fixed-name scratch hazard.
cat > "$WORK/mech_rater.py" <<'PY'
import json, os, random, sys, time
a = sys.argv[1:]
inp = a[a.index("--in") + 1]
out = a[a.index("--out") + 1]
# Fixed-name scratch file in the CURRENT directory. Under rate.py the cwd is a
# private scratch dir, so this is private. Without isolation two concurrent raters
# would share it and clobber each other (the chain 549 all_items.txt race).
token = f"{os.getpid()}-{random.random()}"
with open("all_items.txt", "w") as f:
    f.write(token)
time.sleep(0.3)  # widen the collision window
with open("all_items.txt") as f:
    if f.read() != token:
        sys.stderr.write("SCRATCH COLLISION: all_items.txt changed under me\n")
        sys.exit(3)
codes = {}
for line in open(inp):
    line = line.strip()
    if line:
        obj = json.loads(line)
        codes[obj["id"]] = "C" if "verif" in obj["text"].lower() else "I"
json.dump(codes, open(out, "w"))
PY

# Run two raters concurrently, from a cwd where every path is RELATIVE.
cd "$WORK"
rc_a=0; rc_b=0
python3 "$RATE" --slices-dir slices --out-dir out_a --rater-id claude_a_sim \
    --rater-cmd 'python3 mech_rater.py --in {slice} --out {out}' > log_a.txt 2>&1 &
pid_a=$!
python3 "$RATE" --slices-dir slices --out-dir out_b --rater-id claude_b_sim \
    --rater-cmd 'python3 mech_rater.py --in {slice} --out {out}' > log_b.txt 2>&1 &
pid_b=$!
wait $pid_a || rc_a=$?
wait $pid_b || rc_b=$?

OUT_A="out_a/claude_a_sim/$STEM.json"
OUT_B="out_b/claude_b_sim/$STEM.json"

# Validate: both exited 0, both outputs exist, keys == slice ids, no collision.
python3 - "$SLICE" "$OUT_A" "$OUT_B" "$rc_a" "$rc_b" <<'PY'
import json, sys
slice_path, out_a, out_b, rc_a, rc_b = sys.argv[1:6]
fail = []
if rc_a != "0":
    fail.append(f"rater A exited {rc_a}")
if rc_b != "0":
    fail.append(f"rater B exited {rc_b}")
ids = set()
for line in open(slice_path):
    line = line.strip()
    if line:
        ids.add(json.loads(line)["id"])
for name, path in (("A", out_a), ("B", out_b)):
    try:
        got = json.loads(open(path).read())
    except OSError:
        fail.append(f"rater {name}: no output at {path}")
        continue
    if set(got) != ids:
        fail.append(f"rater {name}: keys != slice ids "
                    f"(missing {ids - set(got)}, extra {set(got) - ids})")
if fail:
    print("SMOKE FAIL:")
    for f in fail:
        print(f"  - {f}")
    sys.exit(1)
print(f"SMOKE PASS: two concurrent raters, {len(ids)} ids each, "
      f"keys match the slice, no shared-scratch collision.")
PY