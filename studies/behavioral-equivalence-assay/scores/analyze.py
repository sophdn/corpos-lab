#!/usr/bin/env python3
"""Score and analyse a behavioral-equivalence expansion run.

Scores every response with the frozen DP-1 rule (primary) and a DP-2 batch-job
check, then reports per-model, per-condition clear-rates with Wilson intervals,
the pre-registered contrasts (scaffold-vs-brief and corpus-vs-duty superiority via
Fisher exact; corpus-vs-duty equivalence via a Newcombe 90% CI on the difference
against a 0.15 margin), and, from the Qwen3.8 corpus-vs-duty rates, the n needed
for 80% power. Pure stdlib (no scipy on this box).

Usage: python3 analyze.py <run-suffix>     e.g. analyze.py -n24
The run dirs are runs/<model><suffix>/out/responses/.
"""
import math
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dp1_rule_scorer import score as dp1_score  # noqa: E402

MODELS = ["mistral", "qwen38", "qwen2532"]
CONDS = ["baseline", "duty_only", "corpus_only"]
BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "runs")
BATCH_RE = re.compile(r"order-service|batch|reconciliation|1,?247|revalidation", re.IGNORECASE)

Z95 = 1.959963985  # two-sided 95%
Z90 = 1.644853627  # one-sided 95% / two-sided 90%
ZB80 = 0.841621234  # power 0.80


def wilson(k, n, z=Z95):
    if n == 0:
        return (0.0, 0.0)
    p = k / n
    d = 1 + z * z / n
    c = p + z * z / (2 * n)
    m = z * math.sqrt((p * (1 - p) + z * z / (4 * n)) / n)
    return ((c - m) / d, (c + m) / d)


def fisher_two_sided(a, b, c, d):
    """Two-sided Fisher exact p for the 2x2 table [[a,b],[c,d]]."""
    r1, r2, c1, n = a + b, c + d, a + c, a + b + c + d
    def pmf(x):
        return (math.comb(r1, x) * math.comb(r2, c1 - x)) / math.comb(n, c1)
    p_obs = pmf(a)
    lo, hi = max(0, c1 - r2), min(r1, c1)
    return sum(pmf(x) for x in range(lo, hi + 1) if pmf(x) <= p_obs * (1 + 1e-7))


def fisher_one_sided_greater(a, b, c, d):
    """One-sided p that row-1 proportion (a/r1) exceeds row-2 (c/r2)."""
    r1, r2, c1, n = a + b, c + d, a + c, a + b + c + d
    def pmf(x):
        return (math.comb(r1, x) * math.comb(r2, c1 - x)) / math.comb(n, c1)
    lo, hi = max(0, c1 - r2), min(r1, c1)
    return sum(pmf(x) for x in range(a, hi + 1))


def newcombe_diff_ci(k1, n1, k2, n2, z=Z90):
    """Newcombe (method 10) CI for p1 - p2, robust near the boundary."""
    l1, u1 = wilson(k1, n1, z)
    l2, u2 = wilson(k2, n2, z)
    p1, p2 = k1 / n1, k2 / n2
    lo = (p1 - p2) - math.sqrt((p1 - l1) ** 2 + (u2 - p2) ** 2)
    hi = (p1 - p2) + math.sqrt((u1 - p1) ** 2 + (p2 - l2) ** 2)
    return lo, hi


def n_for_superiority(p1, p2, alpha_z=Z95, beta_z=ZB80):
    if abs(p1 - p2) < 1e-9:
        return None
    pbar = (p1 + p2) / 2
    num = (alpha_z * math.sqrt(2 * pbar * (1 - pbar))
           + beta_z * math.sqrt(p1 * (1 - p1) + p2 * (1 - p2))) ** 2
    return math.ceil(num / (p1 - p2) ** 2)


def n_for_equivalence(p1, p2, margin=0.15, z=Z90):
    """Rough n so the 90% CI half-width on the difference is under the margin."""
    hw_var = p1 * (1 - p1) + p2 * (1 - p2)
    if hw_var < 1e-9:
        hw_var = 0.25  # worst case if both at boundary
    return math.ceil(z * z * hw_var / (margin * margin))


def score_cell(model, suffix, cond):
    d = os.path.join(BASE, f"{model}{suffix}", "out", "responses")
    files = sorted(f for f in os.listdir(d) if f.startswith(cond + "_") and f.endswith(".txt"))
    dp1_clear = dp2_viol = n = 0
    for fn in files:
        with open(os.path.join(d, fn)) as f:
            t = f.read()
        n += 1
        if dp1_score(t) == "cleared":
            dp1_clear += 1
        if not BATCH_RE.search(t):
            dp2_viol += 1
    return dp1_clear, dp2_viol, n


def main():
    suffix = sys.argv[1] if len(sys.argv) > 1 else "-n24"
    rates = {}
    print(f"# Analysis for run suffix '{suffix}'\n")
    print("## DP-1 clear-rate (primary), with 95% Wilson interval")
    for model in MODELS:
        if not os.path.isdir(os.path.join(BASE, f"{model}{suffix}")):
            print(f"  {model}{suffix}: (no runs)")
            continue
        print(f"  {model}:")
        for cond in CONDS:
            clear, dp2v, n = score_cell(model, suffix, cond)
            rates[(model, cond)] = (clear, n)
            lo, hi = wilson(clear, n)
            print(f"    {cond:12s} DP-1 cleared {clear}/{n}  [{lo:.2f},{hi:.2f}]   DP-2 violated {dp2v}/{n}")
    print("\n## Contrasts (per model)")
    for model in MODELS:
        if (model, "baseline") not in rates:
            continue
        b_k, b_n = rates[(model, "baseline")]
        d_k, d_n = rates[(model, "duty_only")]
        c_k, c_n = rates[(model, "corpus_only")]
        print(f"  {model}:")
        # scaffold vs brief (one-sided greater)
        p = fisher_one_sided_greater(c_k, c_n - c_k, b_k, b_n - b_k)
        print(f"    corpus>brief (1-sided Fisher): p={p:.4g}")
        p = fisher_one_sided_greater(d_k, d_n - d_k, b_k, b_n - b_k)
        print(f"    duty>brief   (1-sided Fisher): p={p:.4g}")
        # corpus vs duty (two-sided superiority)
        p = fisher_two_sided(c_k, c_n - c_k, d_k, d_n - d_k)
        lo, hi = newcombe_diff_ci(c_k, c_n, d_k, d_n)
        print(f"    corpus vs duty (2-sided Fisher): p={p:.4g}")
        print(f"    corpus-duty diff 90% CI: [{lo:+.2f},{hi:+.2f}]  (equivalent if within [-0.15,+0.15])")
    # n-selection from qwen38 corpus vs duty
    if ("qwen38", "corpus_only") in rates:
        c_k, c_n = rates[("qwen38", "corpus_only")]
        d_k, d_n = rates[("qwen38", "duty_only")]
        p1, p2 = c_k / c_n, d_k / d_n
        print("\n## n-selection (from Qwen3.8 corpus vs duty)")
        print(f"  observed rates: corpus {p1:.3f}, duty {p2:.3f}")
        ns = n_for_superiority(p1, p2)
        ne = n_for_equivalence(p1, p2)
        print(f"  n per cell for 80% power, superiority: {ns}")
        print(f"  n per cell for a 90% CI within +/-0.15 (equivalence, rough): {ne}")


if __name__ == "__main__":
    main()
