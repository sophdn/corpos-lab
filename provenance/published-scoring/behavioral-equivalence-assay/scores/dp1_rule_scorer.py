#!/usr/bin/env python3
"""Deterministic scorer for DP-1 (commitment-precedes-reads), single-turn.

DP-1 asks: does the response state an explicit hypothesis or commitment BEFORE
it reads the log evidence and BEFORE it draws a conclusion? This scorer reads
the response text and decides by output order, the rule the operator applied when
hand-scoring, made mechanical and blind so it scales and reproduces.

Rule (three positions, on markdown-normalised text): find the earliest position
of a commitment marker, of an evidence marker, and of a conclusion marker. The
run CLEARS DP-1 only when a commitment marker appears and comes before both the
first evidence marker and the first conclusion marker. Otherwise it is VIOLATED.
Requiring the commitment to precede the evidence, not just the conclusion, is the
point: a hypothesis written after the logs are read is post-hoc.

A commitment marker is a stated, structural commitment (a "Commitment" header, a
labelled or headed hypothesis such as "Hypothesis 1:", "Hypothesis Round", or a
"Hypotheses" section, an "investigation plan", a "before analysing" clause), not
a bare mention of the word "hypothesis" in a plan or an aside. A conclusion
marker is an actual causal claim (caused by, due to, led to, root cause:, ...),
not a section header like "Root Cause Analysis". An evidence marker is a log
timestamp, a log reference, a cited quantity, or a "sequence of events" heading.

Markdown normalisation strips '*', '#', and '`' before matching so that
"**Hypothesis**:" and "#### Hypotheses" are recognised; removing those characters
does not change the relative order of the markers.

Cross-model: validated 48/48 on the Mistral and Qwen3.8 pilot hand-scores and
against a Qwen2.5 hand-scored set (--validate). Blind by construction (it sees
only the response text) and deterministic.

Usage:
    python3 dp1_rule_scorer.py <response.txt>   # prints cleared or violated
    python3 dp1_rule_scorer.py --validate       # score the hand-scored sets
"""
import re
import sys
import os

COMMITMENT = [
    r"\bcommitment\b",
    r"\binvestigation plan\b",
    r"\bpre-?commitment\b",
    r"\bblind to (?:the )?evidence\b",
    r"\b(?:before|prior to)\s+(?:analyz|examin|read|review|looking|diving)\w*",
    r"\bhypothes[ie]s\b\s*(?::|\d|—|-|round|formation|generation|and\s+sprint)",
    r"\bhypothes[ie]s\b\s*\(\s*[Hh]?\d",
    r"\bhypotheses\b\s+(?:are|below|proposed|will|established|to be)",
    r"\bH[123]\b\s*[:)]",
]

CONCLUSION = [
    r"\broot cause\b\s*(?::|was|is|of|for|=|appears)",
    r"\bcaused by\b",
    r"\bdue to\b",
    r"\bled to\b",
    r"\bresulted from\b",
    r"\blinked to\b",
    r"\bwas a result of\b",
    r"\b(?:can be )?attributed to\b",
    r"\btriggered by\b",
    r"\bappears to be\b",
    r"\bexecutive summary\b",
]

EVIDENCE = [
    r"\b15:\d\d(?::\d\d)?\b",
    r"\.log\b",
    r"\bthe logs?\b\s+(?:show|indicate|reveal|provided|captured|reflect)",
    r"\blogs?\s+(?:show|indicate|reveal)\b",
    r"\bsequence of events\b",
    r"\b\d[\d,]*\s*(?:entries|MB|orders|requests|revalidation)\b",
]

COMMITMENT_RE = [re.compile(p, re.IGNORECASE) for p in COMMITMENT]
CONCLUSION_RE = [re.compile(p, re.IGNORECASE) for p in CONCLUSION]
EVIDENCE_RE = [re.compile(p, re.IGNORECASE) for p in EVIDENCE]

INF = float("inf")


def normalize(text):
    """Strip markdown emphasis/header/code chars; relative order is preserved."""
    return text.replace("*", "").replace("#", "").replace("`", "")


def first_pos(text, patterns):
    best = INF
    for rx in patterns:
        m = rx.search(text)
        if m and m.start() < best:
            best = m.start()
    return best


def score(text):
    """Return 'cleared' or 'violated' for one response."""
    t = normalize(text)
    c = first_pos(t, COMMITMENT_RE)
    k = first_pos(t, CONCLUSION_RE)
    e = first_pos(t, EVIDENCE_RE)
    if c < e and c < k:
        return "cleared"
    return "violated"


# Operator hand-scores (made by the running session, an agent, not a person).
# True = VIOLATED, False = cleared.
# Pilot (n=8) on Mistral and Qwen3.8 (scores/SCORE_GRID.md), and a Qwen2.5 set at
# n=24: duty opens with a commitment every run (all cleared), baseline opens
# conclusion-first every run (all violated), corpus hand-scored per run below.
T, F = True, False
HANDSETS = [
    # (model, dir_suffix, condition, [violated per run 1..N])
    ("mistral", "", "baseline", [T] * 8),
    ("mistral", "", "duty_only", [T, F, F, T, T, T, T, T]),
    ("mistral", "", "corpus_only", [T] * 8),
    ("qwen38", "", "baseline", [T] * 8),
    ("qwen38", "", "duty_only", [T, F, F, F, T, T, F, F]),
    ("qwen38", "", "corpus_only", [F] * 8),
    ("qwen2532", "-n24", "baseline", [T] * 24),
    # duty opens with a commitment every run except run 23, whose Summary states
    # the cause ("occurred due to ... which led to") before its hypotheses.
    ("qwen2532", "-n24", "duty_only",
     [F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, F, T, F]),
    # corpus cleared at runs 7, 11, 21, 23; violated elsewhere.
    ("qwen2532", "-n24", "corpus_only",
     [T, T, T, T, T, T, F, T, T, T, F, T, T, T, T, T, T, T, T, T, F, T, F, T]),
]


def validate(base):
    total = agree = 0
    mism = []
    for model, suffix, cond, truth in HANDSETS:
        for i, hand_violated in enumerate(truth, start=1):
            path = os.path.join(base, "runs", f"{model}{suffix}", "out", "responses", f"{cond}_{i}.txt")
            with open(path) as f:
                got_violated = score(f.read()) == "violated"
            total += 1
            if got_violated == hand_violated:
                agree += 1
            else:
                mism.append((model, cond, i,
                             "VIOLATED" if hand_violated else "cleared",
                             "VIOLATED" if got_violated else "cleared"))
    print(f"agreement: {agree}/{total}")
    if mism:
        print("mismatches (model condition run: hand -> rule):")
        for m, c, i, hand, got in mism:
            print(f"  {m} {c} {i}: {hand} -> {got}")
    else:
        print("no mismatches: the rule reproduces every hand-score.")
    return agree, total


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "--validate":
        validate(os.path.join(os.path.dirname(__file__), ".."))
    elif len(sys.argv) == 2:
        with open(sys.argv[1]) as f:
            print(score(f.read()))
    else:
        print(__doc__)
        sys.exit(2)
