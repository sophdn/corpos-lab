#!/usr/bin/env python3
"""Build a blind, shuffled set of produced duties for coverage judging.

Reads every duty in a run's out/responses/, assigns each an opaque id, writes
the duty text to scores/blind/<id>.md with no condition label, and records the
id -> (condition, run) mapping in scores/blind_map.json so the coverage grid can
be unblinded after judging. The judge sees only the duty text and the taboo
definitions, never the condition, so a coverage call cannot be steered by which
arm produced the duty.

Usage: build_blind_set.py <run-dir> <study-dir> [--seed N]
"""
import json
import os
import random
import re
import sys

FNAME = re.compile(r"^(?P<cond>[a-z_]+)_(?P<run>\d+)\.txt$")


def main():
    run_dir, study = sys.argv[1], sys.argv[2]
    seed = 20260912
    if "--seed" in sys.argv:
        seed = int(sys.argv[sys.argv.index("--seed") + 1])
    resp_dir = os.path.join(run_dir, "out", "responses")
    items = []
    for fn in sorted(os.listdir(resp_dir)):
        m = FNAME.match(fn)
        if not m:
            continue
        items.append((m.group("cond"), int(m.group("run")), os.path.join(resp_dir, fn)))

    rng = random.Random(seed)
    rng.shuffle(items)
    blind_dir = os.path.join(study, "scores", "blind")
    os.makedirs(blind_dir, exist_ok=True)
    mapping = {}
    for i, (cond, run, path) in enumerate(items, 1):
        did = f"d{i:04d}"
        with open(path, encoding="utf-8") as f:
            text = f.read()
        with open(os.path.join(blind_dir, f"{did}.md"), "w", encoding="utf-8") as f:
            f.write(text)
        mapping[did] = {"condition": cond, "run": run}
    with open(os.path.join(study, "scores", "blind_map.json"), "w", encoding="utf-8") as f:
        json.dump(mapping, f, indent=2)
    print(f"wrote {len(items)} blind duties to {blind_dir}; map -> scores/blind_map.json")


if __name__ == "__main__":
    main()
