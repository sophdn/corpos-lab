#!/usr/bin/env python3
"""Reconcile the tightened-rubric blind re-rate for the three remaining
neutral-prefix classes against the old single-rater codes.

For each class it reports:
  - inter-rater agreement between the two new blind raters (A vs B), over 384.
  - new strict-consensus C-rate per condition (both raters code C, / 96).
  - old single-rater C-rate per condition (old union code == C, / 96).
  - the full new consensus-code distribution per condition.

Run after the six rater subagents write their codes.json files.
"""
import json
from pathlib import Path
from collections import Counter
HERE = Path(__file__).resolve().parent

SD = Path("/home/sophi/dev/corpos-lab-wt-comprehension-as-compliance-rev/"
          "studies/neutral-prefix-control")
KEY = json.loads((SD / "scoring/key.json").read_text())

CLASSES = {
    "parent-state-check-bypass": (
        "retrated_parent-state-check-bypass_raterA.json",
        "retrated_parent-state-check-bypass_raterB.json",
    ),
    "formal-step-context-bypass": (
        "retrated_formal-step-context-bypass_raterA.json",
        "retrated_formal-step-context-bypass_raterB.json",
    ),
    "conditional-gate-uniform-default": (
        "retrated_conditional-gate-uniform-default_raterA.json",
        "retrated_conditional-gate-uniform-default_raterB.json",
    ),
}
CONDS = ["baseline", "neutral_prefix", "glyph_only", "imperative_only"]


def old_codes(cls):
    d = json.loads((SD / f"scoring/scores/claude/{cls}__A.json").read_text())
    d.update(json.loads((SD / f"scoring/scores/claude/{cls}__B.json").read_text()))
    return d


def main():
    for cls, (pa, pb) in CLASSES.items():
        a = json.loads((HERE/pa).read_text())
        b = json.loads((HERE/pb).read_text())
        old = old_codes(cls)
        ids = [rid for rid, m in KEY.items() if m["cls"] == cls]
        assert len(ids) == 384, f"{cls}: {len(ids)} ids"

        # inter-rater agreement A vs B
        agree = sum(1 for i in ids if a.get(i) == b.get(i))
        print(f"\n=== {cls} ===")
        print(f"inter-rater A/B agreement: {agree}/{len(ids)} = {agree/len(ids):.3f}")

        for cond in CONDS:
            cids = [i for i in ids if KEY[i]["condition"] == cond]
            n = len(cids)
            new_c = sum(1 for i in cids if a.get(i) == "C" and b.get(i) == "C")
            old_c = sum(1 for i in cids if old.get(i) == "C")
            dist = Counter(
                (a.get(i), b.get(i)) if a.get(i) == b.get(i) else "SPLIT"
                for i in cids
            )
            # consensus distribution (agreed code, else SPLIT)
            cons = Counter(a.get(i) if a.get(i) == b.get(i) else "SPLIT" for i in cids)
            print(f"  {cond:16s} n={n:2d}  "
                  f"old C-rate={old_c/n:.3f} ({old_c})  "
                  f"new strict-C={new_c/n:.3f} ({new_c})  "
                  f"consensus={dict(cons)}")


if __name__ == "__main__":
    main()
