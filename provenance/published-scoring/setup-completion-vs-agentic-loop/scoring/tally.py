#!/usr/bin/env python3
"""Tally two blind raters into per-cell code counts for setup-vs-agentic-loop.

Reads key.json and two rater outputs (id -> code), joins on id, and prints, per
glyph, a (setup x condition) table with strict-consensus C (both raters code C)
and the recognition-without-execution rate (consensus Ii). Also prints overall
raw agreement and Cohen's kappa per glyph.

Usage:
    tally.py --key <key.json> --rater-a <a.json> --rater-b <b.json>
"""
import argparse
import json
from collections import defaultdict

CODES = ["C", "Ii", "Ic", "I", "N"]
CONDITIONS = ["baseline", "glyph_only", "imperative_only"]
SETUPS = ["raw", "loop"]


def kappa(pairs):
    """Cohen's kappa over (a, b) code pairs."""
    n = len(pairs)
    if n == 0:
        return float("nan")
    po = sum(1 for a, b in pairs if a == b) / n
    a_marg = defaultdict(int)
    b_marg = defaultdict(int)
    for a, b in pairs:
        a_marg[a] += 1
        b_marg[b] += 1
    pe = sum((a_marg[c] / n) * (b_marg[c] / n) for c in set(a_marg) | set(b_marg))
    return (po - pe) / (1 - pe) if pe != 1 else 1.0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--key", required=True)
    ap.add_argument("--rater-a", required=True)
    ap.add_argument("--rater-b", required=True)
    args = ap.parse_args()

    key = json.load(open(args.key))
    a = json.load(open(args.rater_a))
    b = json.load(open(args.rater_b))

    # cell -> list of (codeA, codeB)
    cells = defaultdict(list)
    per_glyph_pairs = defaultdict(list)
    for rid, meta in key.items():
        if rid not in a or rid not in b:
            continue
        pair = (a[rid], b[rid])
        cells[(meta["glyph"], meta["setup"], meta["condition"])].append(pair)
        per_glyph_pairs[meta["glyph"]].append(pair)

    for glyph in sorted({m["glyph"] for m in key.values()}):
        print(f"\n=== {glyph} ===")
        print(f"{'setup':5} {'condition':16} {'n':>3} {'consC':>6} {'consIi':>7}  per-rater-C(A/B)  per-rater-Ii(A/B)")
        for setup in SETUPS:
            for cond in CONDITIONS:
                pairs = cells.get((glyph, setup, cond), [])
                n = len(pairs)
                if n == 0:
                    continue
                cons_c = sum(1 for x, y in pairs if x == "C" and y == "C")
                cons_ii = sum(1 for x, y in pairs if x == "Ii" and y == "Ii")
                a_c = sum(1 for x, _ in pairs if x == "C")
                b_c = sum(1 for _, y in pairs if y == "C")
                a_ii = sum(1 for x, _ in pairs if x == "Ii")
                b_ii = sum(1 for _, y in pairs if y == "Ii")
                print(f"{setup:5} {cond:16} {n:>3} {cons_c:>6} {cons_ii:>7}  {a_c:>6}/{b_c:<6}  {a_ii:>6}/{b_ii}")
        gp = per_glyph_pairs[glyph]
        agree = sum(1 for x, y in gp if x == y) / len(gp) if gp else float("nan")
        print(f"  raw agreement: {agree:.3f}   kappa: {kappa(gp):.3f}   (n={len(gp)})")


if __name__ == "__main__":
    main()
