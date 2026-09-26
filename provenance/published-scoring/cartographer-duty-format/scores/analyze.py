#!/usr/bin/env python3
"""Cartographer-duty-format analysis.

Consumes a per-run, per-taboo coverage grid (coverage_grid.json, produced by the
blind Claude judge) plus the pre-declared taboo classes (taboo_set.json), and
computes:
  - coverage rate per condition per taboo class, with Wilson 95% intervals;
  - the primary result: the condition-by-class interaction (the annotated -
    cartographer coverage gap on first-principles taboos versus on
    non-first-principles taboos);
  - Fisher exact tests: annotated vs cartographer on the pooled
    non-first-principles taboos, and cartographer vs cartographer+scan on each
    scan-target taboo;
  - a per-taboo coverage table.

Usage: analyze.py <study-dir>   (expects <study-dir>/scores/coverage_grid.json)
"""
import json
import math
import os
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


CONDS = ["baseline", "annotated_instrument", "cartographer_instrument", "cartographer_scan_instrument"]
FP = "first_principles"


def main():
    study = sys.argv[1] if len(sys.argv) > 1 else "."
    taboos = json.load(open(os.path.join(study, "taboo_set.json")))["taboos"]
    cls = {t["slug"]: t["class"] for t in taboos}
    scan_target = {t["slug"] for t in taboos if t["scan_target"]}
    nonfp_slugs = [t["slug"] for t in taboos if t["class"] != FP]
    fp_slugs = [t["slug"] for t in taboos if t["class"] == FP]

    grid = json.load(open(os.path.join(study, "scores", "coverage_grid.json")))["grid"]

    # covered[cond][class] = (k covered, n judged)
    def rate(cond, slugs):
        k = n = 0
        for row in grid.values():
            if row["condition"] != cond:
                continue
            for s in slugs:
                v = row["coverage"].get(s)
                if v is None:
                    continue
                n += 1
                k += 1 if v else 0
        return k, n

    print("== coverage rate per condition per class (k/n, Wilson 95%) ==")
    print(f"{'condition':<32} {'first-principles':>22} {'non-first-principles':>24}")
    rates = {}
    for cond in CONDS:
        fk, fn = rate(cond, fp_slugs)
        nk, nn = rate(cond, nonfp_slugs)
        rates[cond] = {"fp": (fk, fn), "nonfp": (nk, nn)}
        fl, fh = wilson(fk, fn)
        nl, nh = wilson(nk, nn)
        fp_s = f"{fk}/{fn} [{fl:.2f},{fh:.2f}]"
        nf_s = f"{nk}/{nn} [{nl:.2f},{nh:.2f}]"
        print(f"{cond:<32} {fp_s:>22} {nf_s:>24}")

    print("\n== primary interaction: annotated - cartographer coverage gap ==")
    a, c = "annotated_instrument", "cartographer_instrument"
    for label, slugs in [("first-principles", fp_slugs), ("non-first-principles", nonfp_slugs)]:
        ak, an = rate(a, slugs)
        ck, cn = rate(c, slugs)
        ap = ak / an if an else 0
        cp = ck / cn if cn else 0
        print(f"  {label:<22} annotated {ap:.2f} - cartographer {cp:.2f} = gap {ap - cp:+.2f}")
    ak, an = rate(a, nonfp_slugs)
    ck, cn = rate(c, nonfp_slugs)
    p = fisher_two_sided(ak, an - ak, ck, cn - ck)
    print(f"  Fisher (annotated vs cartographer, non-first-principles pooled): p={p:.3g}")

    print("\n== scan recovery: cartographer vs cartographer+scan, per non-FP taboo ==")
    cs = "cartographer_scan_instrument"
    for s in nonfp_slugs:
        ck, cn = rate(c, [s])
        sk, sn = rate(cs, [s])
        tgt = "scan-target" if s in scan_target else "not-targeted"
        p = fisher_two_sided(sk, sn - sk, ck, cn - ck) if (sn and cn) else float("nan")
        print(f"  {s:<42} ({cls[s]:<16}, {tgt}) cart {ck}/{cn} -> scan {sk}/{sn}  p={p:.3g}")

    print("\n== per-taboo coverage rate (k/n) by condition ==")
    hdr = f"{'taboo':<44}{'class':<16}" + "".join(f"{c.split('_')[0][:9]:>10}" for c in CONDS)
    print(hdr)
    for t in taboos:
        s = t["slug"]
        cells = ""
        for cond in CONDS:
            k, n = rate(cond, [s])
            cells += f"{f'{k}/{n}':>10}"
        print(f"{s:<44}{t['class']:<16}{cells}")


if __name__ == "__main__":
    main()
