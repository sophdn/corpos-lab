#!/usr/bin/env python3
"""Isolated, resumable, straggler-proof rater runner for blind grid scoring.

WHY THIS EXISTS (suggestion isolated-deterministic-rater-runner-for-lab-grid-
scoring). During the alphabet-assay scoring, blind-rater subagents self-forked and
wrote into a shared temp namespace: concurrent raters collided, and a 28-minute
straggler landed after analysis and overwrote a clean result. The isolation rules
were prompt-enforced (RATER_INSTRUCTIONS.md), not structural.

This runner makes them structural:

  1. NO nested subagents. A rater is an ordinary subprocess (an argv), not a
     dispatched agent, so it cannot spawn its own sub-forks. Concurrency is a
     fixed worker pool with an explicit size, never recursive fan-out.
  2. ISOLATED scratch. Each rater invocation runs in its own fresh temp directory,
     with TMPDIR and the working directory pointed there, so no two raters share a
     scratch namespace.
  3. UNIQUE, un-clobberable output. A rater writes only to a private path inside
     its own scratch. A validated result is then promoted to the slice's canonical
     output with an atomic, create-only move — so once a slice has a fresh result,
     a late straggler cannot overwrite it.
  4. COMPLETION BY CONTENT. A slice is done when its canonical output exists and
     its JSON keys are exactly the slice's ids — not because a job "finished". A
     killed run resumes and re-runs only the slices without a valid result.

RATER CONTRACT (from RATER_INSTRUCTIONS.md). A rater reads a slice — a JSONL file,
one {"id": "...", "text": "..."} per line — applies one rubric, and writes a single
JSON object mapping every id to its code. The runner passes the slice path and the
output path into the rater command via {slice} and {out} placeholders.

Usage:
    rate.py --slices-dir slices/ --out-dir scores/ --rater-id phi4 \
        --rater-cmd 'python3 score_phi4.py --in {slice} --out {out}' [--jobs N] [--dry-run]

    --jobs defaults to 1 (sequential — the safest, fully deterministic order).
"""
import argparse
import concurrent.futures as cf
import json
import os
import shlex
import subprocess
import sys
import tempfile
from pathlib import Path


def slice_ids(slice_path):
    """The set of ids a slice declares, read from its JSONL lines."""
    ids = []
    for line in Path(slice_path).read_text().splitlines():
        line = line.strip()
        if not line:
            continue
        obj = json.loads(line)
        ids.append(obj["id"])
    return ids


def is_complete(out_path, expected_ids):
    """A canonical output is complete when it parses and its keys are exactly the
    slice's ids — no missing id, no extra id. This is the content check that makes
    completion independent of whether a job process 'finished'."""
    try:
        got = json.loads(Path(out_path).read_text())
    except (OSError, json.JSONDecodeError):
        return False
    return isinstance(got, dict) and set(got) == set(expected_ids)


def run_one(slice_path, out_dir, rater_id, rater_argv):
    """Run one rater on one slice in isolation. Returns (slice_name, status)."""
    slice_path = Path(slice_path)
    expected = slice_ids(slice_path)
    canonical = Path(out_dir) / rater_id / f"{slice_path.stem}.json"
    canonical.parent.mkdir(parents=True, exist_ok=True)

    if is_complete(canonical, expected):
        return (slice_path.name, "skip-complete")

    # Isolated scratch: TMPDIR + cwd both point here, so the rater cannot reach a
    # shared namespace, and its private output path is unique per invocation.
    with tempfile.TemporaryDirectory(prefix=f"rate-{rater_id}-{slice_path.stem}-") as scratch:
        priv_out = Path(scratch) / "out.json"
        argv = [a.replace("{slice}", str(slice_path)).replace("{out}", str(priv_out)) for a in rater_argv]
        env = dict(os.environ, TMPDIR=scratch)
        proc = subprocess.run(argv, cwd=scratch, env=env, capture_output=True, text=True)
        if proc.returncode != 0:
            return (slice_path.name, f"rater-exit-{proc.returncode}: {proc.stderr.strip()[:200]}")
        if not is_complete(priv_out, expected):
            # Keep the bad output next to the canonical dir for inspection; it can
            # never become the canonical result.
            bad = canonical.with_suffix(".invalid.json")
            try:
                bad.write_text(Path(priv_out).read_text())
            except OSError:
                pass
            return (slice_path.name, "invalid-output (keys != slice ids)")
        # Promote create-only: if a fresh canonical appeared meanwhile (a straggler
        # or a parallel run), keep it and drop this one.
        try:
            fd = os.open(str(canonical), os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o644)
        except FileExistsError:
            return (slice_path.name, "already-had-result (kept existing)")
        with os.fdopen(fd, "w") as fh:
            fh.write(Path(priv_out).read_text())
        return (slice_path.name, "scored")


def discover_slices(args):
    if args.slice:
        return [Path(args.slice)]
    return sorted(Path(args.slices_dir).glob("*.jsonl"))


def main(argv=None):
    ap = argparse.ArgumentParser(description="Isolated, resumable blind rater runner.")
    src = ap.add_mutually_exclusive_group(required=True)
    src.add_argument("--slices-dir", help="directory of *.jsonl slices")
    src.add_argument("--slice", help="a single slice JSONL file")
    ap.add_argument("--out-dir", required=True, help="output root; results land under <out-dir>/<rater-id>/")
    ap.add_argument("--rater-id", required=True, help="rater name (namespaces the output dir)")
    ap.add_argument("--rater-cmd", required=True,
                    help="rater command with {slice} and {out} placeholders")
    ap.add_argument("--jobs", type=int, default=1, help="worker pool size (default 1, sequential)")
    ap.add_argument("--dry-run", action="store_true", help="report done vs pending, run nothing")
    args = ap.parse_args(argv)

    if args.jobs < 1:
        print("rate: --jobs must be >= 1", file=sys.stderr)
        return 2
    rater_argv = shlex.split(args.rater_cmd)
    if "{slice}" not in args.rater_cmd or "{out}" not in args.rater_cmd:
        print("rate: --rater-cmd must contain both {slice} and {out} placeholders", file=sys.stderr)
        return 2

    slices = discover_slices(args)
    if not slices:
        print("rate: no slices found", file=sys.stderr)
        return 2

    done, pending = [], []
    for s in slices:
        canonical = Path(args.out_dir) / args.rater_id / f"{s.stem}.json"
        (done if is_complete(canonical, slice_ids(s)) else pending).append(s)

    print(f"rate: {len(slices)} slices — {len(done)} complete, {len(pending)} pending "
          f"(rater {args.rater_id}, jobs {args.jobs})")
    if args.dry_run:
        for s in pending:
            print(f"  PENDING  {s.name}")
        return 0

    results = []
    if args.jobs == 1:
        for s in pending:
            results.append(run_one(s, args.out_dir, args.rater_id, rater_argv))
    else:
        with cf.ThreadPoolExecutor(max_workers=args.jobs) as pool:
            futs = {pool.submit(run_one, s, args.out_dir, args.rater_id, rater_argv): s for s in pending}
            for fut in cf.as_completed(futs):
                results.append(fut.result())

    scored = sum(1 for _, st in results if st == "scored")
    failed = [(n, st) for n, st in results if st not in ("scored", "skip-complete")]
    for n, st in results:
        print(f"  {st:28s} {n}")
    print(f"rate: scored {scored}, {len(failed)} failed/invalid, "
          f"{len(done)} already complete")
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
