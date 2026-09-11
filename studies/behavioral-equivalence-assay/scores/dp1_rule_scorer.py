#!/usr/bin/env python3
"""Deterministic scorer for DP-1 (commitment-precedes-reads), single-turn.

DP-1 asks: does the response state an explicit hypothesis or commitment BEFORE
it reads the log evidence and BEFORE it draws a conclusion? This scorer reads
the response text and decides by output order, the rule the primary human judge
applied, made mechanical and blind so it scales and reproduces.

Rule (three positions): find the earliest position of a commitment marker, of an
evidence marker, and of a conclusion marker. The run CLEARS DP-1 only when a
commitment marker appears and comes before both the first evidence marker and
the first conclusion marker. Otherwise it is VIOLATED. Requiring the commitment
to precede the evidence, not just the conclusion, is the point of the pattern:
a hypothesis written after the logs are read is post-hoc.

A commitment marker is a stated, structural commitment (a "Commitment" header, a
labelled hypothesis such as "Hypothesis 1:" or "H1:", an "investigation plan",
or a "before analysing" clause), not a bare mention of the word "hypothesis" in
a plan or an aside. A conclusion marker is an actual causal claim, not a section
header like "Root Cause Analysis". An evidence marker is a log timestamp, a log
reference, a cited quantity, or a "sequence of events" heading.

Blind by construction: the scorer sees only the response text, never the
condition. Deterministic: same text yields the same code anywhere.

Usage:
    python3 dp1_rule_scorer.py <response.txt>   # prints cleared or violated
    python3 dp1_rule_scorer.py --validate       # score the committed 48 and
                                                # compare to the hand-scores
"""
import re
import sys
import os

# Stated, structural commitments. A bare "hypothesis" is deliberately NOT here:
# a plan ("then propose a hypothesis") or an aside ("a confirmed hypothesis
# survives elimination") is not a commitment authored before the evidence.
COMMITMENT = [
    r"\bcommitment\b",
    r"\binvestigation plan\b",
    r"\bpre-?commitment\b",
    r"\bblind to (?:the )?evidence\b",
    r"\b(?:before|prior to)\s+(?:analyz|examin|read|review|looking|diving)\w*",
    r"\bhypothes[ie]s\b\s*(?:\d|:|—|\bformulation\b|\bgeneration\b)",
    r"\bhypothes[ie]s\b\s*\(\s*[Hh]?\d",
    r"\bhypotheses\b\s+(?:are|below|proposed|will|established|to be)",
    r"\bH[123]\b\s*[:)]",
]

# An actual causal claim. "root cause" counts only as a claim (colon or copula
# after it), so "Root Cause Analysis" and "Root Cause Identification" do not fire.
CONCLUSION = [
    r"\broot cause\b\s*(?::|was|is|of|for|=|appears)",
    r"\bcaused by\b",
    r"\bwas a result of\b",
    r"\bcan be attributed to\b",
    r"\battributed to\b",
    r"\btriggered by\b",
    r"\bappears to be\b",
    r"\bthe (?:payment[ -]processing )?failure (?:was|is|appears|resulted|can be|stemmed)\b",
    r"\bexecutive summary\b",
]

# Log evidence being read: a timestamp, a log-file reference, a cited quantity,
# or a "sequence of events" heading.
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


def first_pos(text, patterns):
    best = INF
    for rx in patterns:
        m = rx.search(text)
        if m and m.start() < best:
            best = m.start()
    return best


def score(text):
    """Return 'cleared' or 'violated' for one response."""
    c = first_pos(text, COMMITMENT_RE)
    k = first_pos(text, CONCLUSION_RE)
    e = first_pos(text, EVIDENCE_RE)
    if c < e and c < k:
        return "cleared"
    return "violated"


# The primary human judge's DP-1 hand-scores for the n=8 pilot (scores/SCORE_GRID.md).
# True = VIOLATED, False = cleared. Runs 1..8.
HANDSCORES = {
    ("mistral", "baseline"): [True] * 8,
    ("mistral", "duty_only"): [True, False, False, True, True, True, True, True],
    ("mistral", "corpus_only"): [True] * 8,
    ("qwen38", "baseline"): [True] * 8,
    ("qwen38", "duty_only"): [True, False, False, False, True, True, False, False],
    ("qwen38", "corpus_only"): [False] * 8,
}


def validate(base):
    total = agree = 0
    mism = []
    for (model, cond), truth in HANDSCORES.items():
        for i, hand_violated in enumerate(truth, start=1):
            path = os.path.join(base, "runs", model, "out", "responses", f"{cond}_{i}.txt")
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
