#!/usr/bin/env python3
"""Clustering-robust reanalysis at the duty level, plus a rater-free
corroboration of the coverage direction.

Reader-C review W1: the pooled per-call Fisher tests treat the 3 (or 7) hazard
calls within one duty as independent, which they are not. Here each duty is the
unit of analysis (n=30 per condition). W1 also asked for the registered
condition-by-class interaction to be tested, not just described: a permutation
test on the per-duty (first-principles rate minus non-first-principles rate),
annotated versus cartographer, is the difference-in-differences that IS the
interaction, with the duty as the independent unit.

Reader-C review W3: a deterministic keyword-presence pass corroborates the
coverage direction for the three non-first-principles hazards without any judge.
"""
import json
import math
import os
import random
import re
import sys

Z95 = 1.959963985


def wilson(k, n, z=Z95):
    if n == 0:
        return (0.0, 0.0)
    p = k / n
    d = 1 + z * z / n
    c = p + z * z / (2 * n)
    m = z * math.sqrt((p * (1 - p) + z * z / (4 * n)) / n)
    return ((c - m) / d, (c + m) / d)


def fisher_two_sided(a, b, c, d):
    r1, r2, c1, n = a + b, c + d, a + c, a + b + c + d
    def pmf(x):
        return (math.comb(r1, x) * math.comb(r2, c1 - x)) / math.comb(n, c1)
    p_obs = pmf(a)
    lo, hi = max(0, c1 - r2), min(r1, c1)
    return sum(pmf(x) for x in range(lo, hi + 1) if pmf(x) <= p_obs * (1 + 1e-7))


def perm_test(xs, ys, reps=100000, seed=20260912):
    """Two-sided permutation test on the difference of means of xs vs ys."""
    rng = random.Random(seed)
    obs = abs(sum(xs) / len(xs) - sum(ys) / len(ys))
    pool = xs + ys
    n = len(xs)
    hits = 0
    for _ in range(reps):
        rng.shuffle(pool)
        d = abs(sum(pool[:n]) / n - sum(pool[n:]) / len(ys))
        if d >= obs - 1e-12:
            hits += 1
    return (hits + 1) / (reps + 1)


def main():
    study = sys.argv[1] if len(sys.argv) > 1 else "."
    taboos = json.load(open(os.path.join(study, "taboo_set.json")))["taboos"]
    fp = [t["slug"] for t in taboos if t["class"] == "first_principles"]
    nonfp = [t["slug"] for t in taboos if t["class"] != "first_principles"]
    grid = json.load(open(os.path.join(study, "scores", "coverage_grid.json")))["grid"]

    # Per-duty rates by condition.
    duties = {}
    for row in grid.values():
        cond = row["condition"]
        cov = row["coverage"]
        fpr = sum(cov[s] for s in fp) / len(fp)
        nfr = sum(cov[s] for s in nonfp) / len(nonfp)
        duties.setdefault(cond, []).append((fpr, nfr))

    print("== per-duty mean coverage rate (n=30 duties/condition) ==")
    for cond, rows in duties.items():
        fpm = sum(r[0] for r in rows) / len(rows)
        nfm = sum(r[1] for r in rows) / len(rows)
        print(f"  {cond:<32} FP {fpm:.3f}  nonFP {nfm:.3f}")

    a, c, b = "annotated_instrument", "cartographer_instrument", "baseline"

    # Registered interaction, clustering-robust: per-duty (FP - nonFP), A vs C.
    da = [r[0] - r[1] for r in duties[a]]
    dc = [r[0] - r[1] for r in duties[c]]
    p_int = perm_test(da, dc)
    print("\n== registered interaction (duty-level, permutation) ==")
    print(f"  per-duty (FP-nonFP): annotated mean {sum(da)/len(da):+.3f}, "
          f"cartographer mean {sum(dc)/len(dc):+.3f}")
    print(f"  permutation p (annotated vs cartographer) = {p_int:.2g}")

    # non-FP contrast at duty level: duty covers >=1 non-FP hazard.
    def any_nonfp(rows):
        return sum(1 for r in rows if r[1] > 0)
    ak, ck = any_nonfp(duties[a]), any_nonfp(duties[c])
    p_nf = fisher_two_sided(ak, 30 - ak, ck, 30 - ck)
    print("\n== non-first-principles, duty-level (covers >=1 non-FP hazard) ==")
    print(f"  annotated {ak}/30, cartographer {ck}/30, Fisher p={p_nf:.2g}")

    # FP contrast at duty level: cartographer vs baseline, per-duty FP rate.
    ca = [r[0] for r in duties[c]]
    ba = [r[0] for r in duties[b]]
    p_fp = perm_test(ca, ba)
    print("\n== first-principles, duty-level (cartographer vs baseline) ==")
    print(f"  cartographer mean {sum(ca)/len(ca):.3f}, baseline mean {sum(ba)/len(ba):.3f}, "
          f"permutation p={p_fp:.2g}")

    # Per-taboo range for cartographer FP coverage.
    def taboo_count(cond, slug):
        return sum(1 for r in grid.values() if r["condition"] == cond and r["coverage"][slug])
    fp_counts = [(s, taboo_count(c, s)) for s in fp]
    print("\n== cartographer first-principles per-taboo spread ==")
    print("  " + ", ".join(f"{s.split('-')[-1]} {k}/30" for s, k in fp_counts))
    print(f"  range {min(k for _, k in fp_counts)}/30 to {max(k for _, k in fp_counts)}/30")

    # Rater-free keyword-presence corroboration for the 3 non-FP hazards.
    print("\n== rater-free keyword presence (responses mentioning the hazard) ==")
    KW = {
        "routing": re.compile(r"\brout(e|ing)\b", re.I),
        "durability(advisory)": re.compile(r"\b(advisory|durabilit|bypass)\w*", re.I),
        "constraint(discovered)": re.compile(r"\b(constraint|discover)\w*", re.I),
    }
    resp = os.path.join(study, "runs", "qwen38", "out", "responses")
    for cond in ["baseline", "annotated_instrument", "cartographer_instrument", "cartographer_scan_instrument"]:
        counts = {k: 0 for k in KW}
        for i in range(1, 31):
            fn = os.path.join(resp, f"{cond}_{i}.txt")
            if not os.path.exists(fn):
                continue
            text = open(fn, encoding="utf-8").read()
            for k, rx in KW.items():
                if rx.search(text):
                    counts[k] += 1
        print(f"  {cond:<32} " + "  ".join(f"{k} {counts[k]}/30" for k in KW))


if __name__ == "__main__":
    main()
