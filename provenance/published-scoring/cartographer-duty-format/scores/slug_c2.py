#!/usr/bin/env python3
"""C2 — deterministic taboo-slug citation scorer.

Reads every produced duty in a run's out/responses/ and counts how many of the
ten registered taboo slugs it cites. This is the mechanism half of the finding:
the annotated method cites slugs (canon is a runtime dependency of the produced
duty); the cartographer method cites none. The count is exact and reproducible —
no judgment — so it separates the formats without a rater.

Usage: slug_c2.py <run-dir> [<run-dir> ...]
Prints a per-condition summary and writes slug_c2.json next to the first run.
"""
import json
import os
import re
import sys

SLUGS = [
    "triggers-routing",
    "explicit-prerequisite-gate",
    "investigation-fix-criterion-predeclaration",
    "investigation-fix-scope-boundary",
    "fix-locus-identification",
    "fix-attempt-root-cause-reassessment",
    "investigation-advisory-fix-durability",
    "completion-evidence-required",
    "document-claim-verification",
    "known-constraint-documentation",
]
# A slug is "cited" only as a backticked or bare kebab token, not as ordinary
# prose containing the words. Match the exact slug string.
PATTERNS = {s: re.compile(re.escape(s)) for s in SLUGS}
FNAME = re.compile(r"^(?P<cond>[a-z_]+)_(?P<run>\d+)\.txt$")


def cited_slugs(text):
    return [s for s, p in PATTERNS.items() if p.search(text)]


def main():
    run_dirs = sys.argv[1:]
    if not run_dirs:
        print(__doc__)
        sys.exit(1)
    grid = {}
    by_cond = {}
    for rd in run_dirs:
        resp_dir = os.path.join(rd, "out", "responses")
        for fn in sorted(os.listdir(resp_dir)):
            m = FNAME.match(fn)
            if not m:
                continue
            cond, run = m.group("cond"), int(m.group("run"))
            with open(os.path.join(resp_dir, fn), encoding="utf-8") as f:
                text = f.read()
            hits = cited_slugs(text)
            key = f"{cond}_{run}"
            grid[key] = {"condition": cond, "run": run, "n_slugs": len(hits), "slugs": hits}
            by_cond.setdefault(cond, []).append(len(hits))

    print(f"{'condition':<32} {'runs':>4} {'runs citing >=1':>16} {'mean slugs':>11}")
    summary = {}
    for cond in sorted(by_cond):
        counts = by_cond[cond]
        n = len(counts)
        cited = sum(1 for c in counts if c > 0)
        mean = sum(counts) / n if n else 0.0
        print(f"{cond:<32} {n:>4} {cited:>16} {mean:>11.2f}")
        summary[cond] = {"runs": n, "runs_citing_ge1": cited, "mean_slugs": mean}

    out = os.path.join(run_dirs[0], "slug_c2.json")
    with open(out, "w", encoding="utf-8") as f:
        json.dump({"summary": summary, "grid": grid}, f, indent=2)
    print(f"\nwrote {out}")


if __name__ == "__main__":
    main()
