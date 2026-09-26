#!/usr/bin/env python3
"""Score a rater against a human validity anchor and against the Claude raters.

The anchor (studies/*/human-anchor) holds a small stratified slice the author
blind-scored. Machine raters score anchor_slice.jsonl (opaque r-ids, one per line);
the human codes are keyed H01..HNN in file order. Alignment is POSITIONAL: the kth
line of anchor_slice.jsonl is item H{k}. This is how the anchor was built
(build_anchor.py writes items.json and anchor_slice.jsonl in one pass) and is
self-validated below: run with no --rater and the Claude two-rater consensus is
scored against the human, which must reproduce the AGREEMENT.md headline
(off-task N 13/13, C-vs-not-C 37/39 for neutral-prefix).

Reports, for the rater under test:
  - exact-code agreement vs human and vs Claude consensus,
  - C-vs-not-C agreement (the correct-action call) vs both,
  - off-task N recall (of the human-N items, how many the rater also called N) — the
    metric phi-4 failed (it under-detected off-task),
  - the direction of the C-boundary disagreements vs human.

Usage:
  anchor_eval.py --anchor-dir studies/neutral-prefix-control/human-anchor            # Claude self-check
  anchor_eval.py --anchor-dir <dir> --rater <dir>/roster/granite30b.json --label granite30b
"""
import argparse
import json
from pathlib import Path


def load_json(p):
    return json.loads(Path(p).read_text())


def ordered_rids(anchor_dir):
    rids = []
    for line in (Path(anchor_dir) / "anchor_slice.jsonl").read_text().splitlines():
        line = line.strip()
        if line:
            rids.append(json.loads(line)["id"])
    return rids


def agreement(pred, gold, ids):
    """(exact, c_vs_notc) over ids where both have a code."""
    common = [i for i in ids if pred.get(i) and gold.get(i)]
    if not common:
        return 0, 0, 0
    exact = sum(1 for i in common if pred[i] == gold[i])
    cnc = sum(1 for i in common if (pred[i] == "C") == (gold[i] == "C"))
    return exact, cnc, len(common)


def report(label, pred, human, claude, ids):
    print(f"\n=== {label} ===")
    for name, gold in (("human", human), ("claude-consensus", claude)):
        e, c, n = agreement(pred, gold, ids)
        print(f"  vs {name:16s} exact {e:2d}/{n} = {e/n:.3f}   "
              f"C-vs-notC {c:2d}/{n} = {c/n:.3f}")
    # off-task N recall against the human
    human_n = [i for i in ids if human.get(i) == "N"]
    got_n = sum(1 for i in human_n if pred.get(i) == "N")
    over_n = sum(1 for i in ids if pred.get(i) == "N" and human.get(i) not in ("N", None))
    print(f"  off-task N recall (human-N -> rater-N): {got_n}/{len(human_n)}"
          f"   over-called N (rater-N, human-not-N): {over_n}")
    # C-boundary disagreement direction vs human
    lenient = sum(1 for i in ids if pred.get(i) == "C" and human.get(i) not in ("C", None))
    strict = sum(1 for i in ids if human.get(i) == "C" and pred.get(i) not in ("C", None))
    print(f"  C-boundary: rater-C / human-not-C = {lenient} (lenient)   "
          f"human-C / rater-not-C = {strict} (strict)")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--anchor-dir", required=True)
    ap.add_argument("--rater", default=None, help="rater output {r-id: code}; omit for Claude self-check")
    ap.add_argument("--label", default=None)
    a = ap.parse_args()
    ad = Path(a.anchor_dir)

    rids = ordered_rids(ad)
    human_by_h = load_json(ad / "human_codes.json")
    human = {rid: human_by_h.get(f"H{k + 1:02d}") for k, rid in enumerate(rids)}
    ra = load_json(ad / "anchor_raterA.json")
    rb = load_json(ad / "anchor_raterB.json")
    claude = {rid: (ra.get(rid) if ra.get(rid) == rb.get(rid) else "SPLIT") for rid in rids}

    print(f"anchor {ad.name}: {len(rids)} items  |  human-N={sum(1 for v in human.values() if v == 'N')}"
          f"  claude-splits={sum(1 for v in claude.values() if v == 'SPLIT')}")

    # Self-check: Claude consensus vs human must reproduce AGREEMENT.md.
    report("SELF-CHECK: claude-consensus vs human", claude, human, claude, rids)

    if a.rater:
        pred = load_json(a.rater)
        report(a.label or Path(a.rater).stem, pred, human, claude, rids)


if __name__ == "__main__":
    main()
