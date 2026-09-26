#!/usr/bin/env python3
"""Aggregate rater codes into per-cell correct-action (C) rates for the
neutral-prefix control grid. Reads key.json (id -> metadata) and one or more
rater score dirs/files. Reports per (class, condition) pooled C-rate and per
(class, scenario, model, condition) C-rate, plus rater agreement and
strict-consensus C when two raters are given.

Usage:
  aggregate.py --rater-a scores/claude --rater-b scores/phi4   # dirs (merge *.json) or single files
  aggregate.py --rater-a scores/phi4                            # single rater
"""
import argparse, glob, json, os
from collections import defaultdict

CONDS = ["baseline", "neutral_prefix", "glyph_only", "imperative_only"]
CLASSES = ["casg-direct", "formal-step-context-bypass",
           "parent-state-check-bypass", "conditional-gate-uniform-default"]


def load_scores(path):
    codes = {}
    files = glob.glob(os.path.join(path, "*.json")) if os.path.isdir(path) else [path]
    for f in files:
        codes.update(json.load(open(f)))
    return codes


def crate(items):  # items: list of codes
    n = len(items)
    c = sum(1 for x in items if x == "C")
    return c, n, (c / n if n else float("nan"))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--key", default=os.path.join(os.path.dirname(__file__), "key.json"))
    ap.add_argument("--rater-a", required=True)
    ap.add_argument("--rater-b", default=None)
    a = ap.parse_args()
    key = json.load(open(a.key))
    A = load_scores(a.rater_a)
    B = load_scores(a.rater_b) if a.rater_b else None

    print(f"rater-a: {len(A)} ids" + (f" | rater-b: {len(B)} ids" if B else ""))
    if B:
        common = [i for i in key if i in A and i in B]
        agree = sum(1 for i in common if A[i] == B[i])
        cagree = sum(1 for i in common if (A[i] == "C") == (B[i] == "C"))
        print(f"agreement (exact code): {agree}/{len(common)} = {agree/len(common):.3f}")
        print(f"agreement (C vs not-C): {cagree}/{len(common)} = {cagree/len(common):.3f}")

    def code_for(i):
        if B:  # strict-consensus C: both must code C, else not-C
            return "C" if (A.get(i) == "C" and B.get(i) == "C") else "x"
        return A.get(i, "?")

    # pooled per class x condition
    print("\n== pooled C-rate per class x condition ==" + ("  [strict-consensus]" if B else "  [rater-a]"))
    hdr = "class".ljust(34) + "".join(c[:9].ljust(11) for c in CONDS)
    print(hdr)
    pooled = defaultdict(list)
    for i, m in key.items():
        pooled[(m["cls"], m["condition"])].append(code_for(i))
    for cls in CLASSES:
        row = cls.ljust(34)
        for cond in CONDS:
            c, n, r = crate(pooled[(cls, cond)])
            row += f"{c:>2}/{n:<3}={r:4.2f}".ljust(11)
        print(row)

    # per class x model x scenario x condition (C count out of 16)
    print("\n== C-count/16 per class x model x scenario x condition ==")
    cells = defaultdict(list)
    for i, m in key.items():
        cells[(m["cls"], m["model"], m["scenario"], m["condition"])].append(code_for(i))
    for cls in CLASSES:
        for model in ["mistral", "phi4", "qwen38"]:
            for sc in ["1", "2"]:
                parts = []
                for cond in CONDS:
                    c, n, r = crate(cells[(cls, model, sc, cond)])
                    parts.append(f"{cond[:4]}={c:>2}/{n}")
                print(f"  {cls[:22]:22} {model:8} s{sc}  " + "  ".join(parts))


if __name__ == "__main__":
    main()
