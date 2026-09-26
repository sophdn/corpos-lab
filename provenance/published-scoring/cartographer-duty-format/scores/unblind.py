#!/usr/bin/env python3
"""Unblind a judge's coverage calls into the analysis grid.

Takes the judge output (judge_raw.json: {blind_id: {slug: 0|1, ...}}) and the
blind map, and writes coverage_grid.json keyed by condition_run, the shape
analyze.py consumes. The judge never saw the condition; this step attaches it.

Usage: unblind.py <study-dir> [--judge judge_raw.json]
"""
import json
import os
import sys

SLUGS = [
    "triggers-routing", "explicit-prerequisite-gate",
    "investigation-fix-criterion-predeclaration", "investigation-fix-scope-boundary",
    "fix-locus-identification", "fix-attempt-root-cause-reassessment",
    "investigation-advisory-fix-durability", "completion-evidence-required",
    "document-claim-verification", "known-constraint-documentation",
]


def main():
    study = sys.argv[1]
    judge_file = "judge_raw.json"
    if "--judge" in sys.argv:
        judge_file = sys.argv[sys.argv.index("--judge") + 1]
    scores = os.path.join(study, "scores")
    mapping = json.load(open(os.path.join(scores, "blind_map.json")))
    judge = json.load(open(os.path.join(scores, judge_file)))

    grid = {}
    missing = []
    for did, meta in mapping.items():
        calls = judge.get(did)
        if calls is None:
            missing.append(did)
            continue
        cov = {s: int(bool(calls.get(s, 0))) for s in SLUGS}
        key = f"{meta['condition']}_{meta['run']}"
        grid[key] = {"condition": meta["condition"], "run": meta["run"], "coverage": cov}
    out = os.path.join(scores, "coverage_grid.json")
    with open(out, "w", encoding="utf-8") as f:
        json.dump({"judge": "claude-opus-4-8, blind to condition", "grid": grid}, f, indent=2)
    print(f"wrote {out} ({len(grid)} rows)")
    if missing:
        print(f"WARNING: {len(missing)} blind ids had no judge call: {missing[:10]}")


if __name__ == "__main__":
    main()
