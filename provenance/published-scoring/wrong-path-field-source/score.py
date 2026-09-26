#!/usr/bin/env python3
"""Score the wrong-path-field-source runs, per SCORING_RUBRIC.md.

Deterministic layer only (the headline): parse VERDICT and FIELD SOURCE from each
response, classify the field, and compute per cell the verdict accuracy, the
scope-citation rate among correct verdicts, and the wrong-path rate. Then the
routing classification (from the base arm) and the brittleness differential
(base vs ablated). No judge here; observable/evidence quality is a separate pass.

Run from the study directory:  python3 score.py
Reads runs/<arm>/<model>/<glyph>-<ab>/out/{results.json,responses/*.txt}.
"""
from __future__ import annotations

import collections
import glob
import json
import os
import re
import statistics

HERE = os.path.dirname(os.path.abspath(__file__))
RUNS = os.path.join(HERE, "runs")

MODELS = ["mistral", "qwen2532", "qwen38"]
ARMS = ["base", "ablated"]
GLYPHS = ["cas", "cgu", "fsb", "gop", "psc", "scb"]
AB = ["a", "b"]

# Ground truth: a-scenario fires (yes), b-scenario is a carve-out (no).
GT = {"a": "yes", "b": "no"}

# Field-source families, checked by earliest occurrence in the cited string.
# "rest" resolves the Rest axis, which the ablation leaves in place.
FAMILIES = ["scope", "marker", "aim", "pull", "rest"]


def parse_verdict(text: str) -> str | None:
    m = re.search(r"VERDICT:\s*(yes|no)", text, re.IGNORECASE)
    return m.group(1).lower() if m else None


def classify_field(text: str) -> str:
    m = re.search(r"FIELD SOURCE:\s*(.+)", text, re.IGNORECASE)
    if not m:
        return "none"
    s = m.group(1).lower()
    hits = [(s.find(f), f) for f in FAMILIES if f in s]
    if not hits:
        return "other"
    return min(hits)[1]  # earliest family keyword in reading order


def cell_dir(arm: str, model: str, glyph: str, ab: str) -> str:
    return os.path.join(RUNS, arm, model, f"{glyph}-{ab}", "out")


def score_cell(arm: str, model: str, glyph: str, ab: str) -> dict | None:
    out = cell_dir(arm, model, glyph, ab)
    if not os.path.isfile(os.path.join(out, "results.json")):
        return None
    gt = GT[ab]
    runs = sorted(glob.glob(os.path.join(out, "responses", "grounded_glyph_*.txt")))
    correct = 0
    correct_scope = 0
    n = 0
    for rf in runs:
        n += 1
        text = open(rf).read()
        v = parse_verdict(text)
        fam = classify_field(text)
        if v == gt:
            correct += 1
            if fam == "scope":
                correct_scope += 1
    return {
        "n": n,
        "correct": correct,
        "correct_scope": correct_scope,   # correct verdicts citing Scope (intended)
        "correct_shortcut": correct - correct_scope,  # correct via non-scope route
    }


def main() -> None:
    cells: dict = {}
    for arm in ARMS:
        for model in MODELS:
            for glyph in GLYPHS:
                for ab in AB:
                    c = score_cell(arm, model, glyph, ab)
                    if c:
                        cells[(arm, model, glyph, ab)] = c

    print(f"scored {len(cells)} cells\n")

    # H1 — dissociation, per model (base arm). Verdict accuracy vs scope-citation.
    print("== H1: verdict accuracy vs scope engagement (BASE arm) ==")
    print(f"{'model':10} {'acc':>6} {'scope-cite/correct':>20} {'wrong-path/correct':>20}")
    for model in MODELS:
        cc = [cells[(a, m, g, ab)] for (a, m, g, ab) in cells if a == "base" and m == model]
        n = sum(c["n"] for c in cc)
        corr = sum(c["correct"] for c in cc)
        cs = sum(c["correct_scope"] for c in cc)
        acc = corr / n if n else 0
        scope_rate = cs / corr if corr else float("nan")
        wp_rate = 1 - scope_rate if corr else float("nan")
        print(f"{model:10} {acc:6.2f} {scope_rate:20.2f} {wp_rate:20.2f}   (correct {corr}/{n}, scope {cs})")

    # Same split by scenario polarity (a = firing, b = carve-out), base arm.
    print("\n== H1 by scenario polarity (BASE arm), all models pooled ==")
    for ab in AB:
        cc = [cells[(a, m, g, x)] for (a, m, g, x) in cells if a == "base" and x == ab]
        n = sum(c["n"] for c in cc); corr = sum(c["correct"] for c in cc); cs = sum(c["correct_scope"] for c in cc)
        sr = cs / corr if corr else float("nan")
        print(f"  {ab} ({GT[ab]}): acc {corr/n:.2f}  scope-cite/correct {sr:.2f}  (correct {corr}/{n})")

    # Routing classification per base-arm cell, and brittleness differential.
    print("\n== H2: brittleness (BASE vs ABLATED accuracy drop, by base-arm routing) ==")
    for model in MODELS:
        shortcut_drops = []
        scope_drops = []
        detail = []
        for glyph in GLYPHS:
            for ab in AB:
                b = cells.get(("base", model, glyph, ab))
                a = cells.get(("ablated", model, glyph, ab))
                if not b or not a:
                    continue
                if b["correct"] < 2:
                    routing = "uninform"  # too few correct to classify
                else:
                    routing = "shortcut" if b["correct_shortcut"] > b["correct_scope"] else "scope"
                drop = (b["correct"] - a["correct"]) / b["n"]
                detail.append((glyph, ab, routing, b["correct"], a["correct"], drop))
                if routing == "shortcut":
                    shortcut_drops.append(drop)
                elif routing == "scope":
                    scope_drops.append(drop)
        sc = statistics.mean(shortcut_drops) if shortcut_drops else float("nan")
        scp = statistics.mean(scope_drops) if scope_drops else float("nan")
        diff = sc - scp if shortcut_drops and scope_drops else float("nan")
        print(f"\n  {model}: shortcut-cell mean drop {sc:.2f} (n={len(shortcut_drops)}), "
              f"scope-cell mean drop {scp:.2f} (n={len(scope_drops)}), differential {diff:+.2f}")
        for g, ab, r, bc, ac, d in detail:
            print(f"    {g}-{ab:1} {r:8} base {bc}/8 -> ablated {ac}/8  drop {d:+.2f}")

    # The mechanism behind the H2 null: where routing GOES when the shortcut is
    # removed. If ablation re-routes to Scope with accuracy held, the shortcut was
    # a preference, not a crutch — the scope path was available underneath.
    print("\n== Field-source class distribution per model/arm (all runs) ==")
    for model in MODELS:
        for arm in ARMS:
            c = collections.Counter()
            for rf in glob.glob(os.path.join(RUNS, arm, model, "*", "out", "responses", "grounded_glyph_*.txt")):
                c[classify_field(open(rf).read())] += 1
            print(f"  {arm:8} {model:9} {dict(c)}")

    # Per-model, per-polarity field-source among CORRECT verdicts (base arm). The
    # non-scope route differs by model: a recognition axis for the Qwen models,
    # the Pull character for Mistral.
    print("\n== Field-source among correct verdicts, by model and polarity (BASE arm) ==")
    for model in MODELS:
        for ab in AB:
            c = collections.Counter()
            for glyph in GLYPHS:
                out = cell_dir("base", model, glyph, ab)
                for rf in glob.glob(os.path.join(out, "responses", "grounded_glyph_*.txt")):
                    text = open(rf).read()
                    if parse_verdict(text) == GT[ab]:
                        c[classify_field(text)] += 1
            print(f"  {model:9} {ab} ({GT[ab]:3}) {dict(c)}")

    # W1/W2: does the thinking model's reasoning engage scope even when it cites a
    # non-scope field? Among Qwen3.8 base runs with a non-scope FIELD SOURCE, the
    # fraction whose reasoning trace mentions the scope check. A self-report that
    # names a non-scope field over reasoning that worked the scope condition is an
    # unfaithful self-report, not an unengaged field.
    print("\n== Qwen3.8 base: scope worked in the trace despite a non-scope citation ==")
    nonscope = mention = 0
    for rf in glob.glob(os.path.join(RUNS, "base", "qwen38", "*", "out", "responses", "grounded_glyph_*.txt")):
        if classify_field(open(rf).read()) == "scope":
            continue
        nonscope += 1
        reas = rf.replace("/responses/", "/reasoning/")
        if os.path.isfile(reas) and ("scope" in open(reas).read().lower() or "operative" in open(reas).read().lower()):
            mention += 1
    print(f"  non-scope-cited base runs: {nonscope}; trace mentions scope/operative: {mention}")


if __name__ == "__main__":
    main()
