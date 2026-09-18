#!/usr/bin/env python3
"""Build blind rating slices + a held-out key from the setup-vs-agentic-loop runs.

One slice per glyph, MIXING every condition and both setups, so a rater scores
against the class correct-target bar without the slice grouping revealing the
condition. Ids are opaque content hashes; the key.json maps each id to its
(glyph, condition, setup, run) for de-anonymising after scoring.

Blindness note (honest): a loop transcript and a raw-completion response differ
in structure, so a rater can see which SETUP produced a response. The blindness
that holds is to the CONDITION (baseline / glyph_only / imperative_only) and to
the predicted direction; the correct-target bar is identical across setups.

Usage:
    build_slices.py --study-root <dir> --out <dir> [--glyph <name> ...]
"""
import argparse
import hashlib
import json
from pathlib import Path

# study-name -> (setup, glyph-dir). The run dirs live under <glyph-dir>/runs/<study-name>.
STUDIES = {
    "setup-raw-casg-direct-qwen38": ("raw", "casg-direct"),
    "setup-loop-casg-direct-qwen38": ("loop", "casg-direct"),
    "setup-raw-parent-state-qwen38": ("raw", "parent-state-check-bypass"),
    "setup-loop-parent-state-qwen38": ("loop", "parent-state-check-bypass"),
}
GLYPH_OF = {
    "casg-direct": "casg-direct",
    "parent-state-check-bypass": "parent-state-check-bypass",
}
CONDITIONS = ["baseline", "glyph_only", "imperative_only"]


def opaque_id(study, cond, run, text):
    """A content hash that leaks neither condition nor setup."""
    h = hashlib.sha256(f"{study}\x00{cond}\x00{run}\x00{text}".encode()).hexdigest()
    return h[:16]


def collect(study_root):
    """Yield (glyph, setup, condition, run, text) for every run response found."""
    for study, (setup, glyph_dir) in STUDIES.items():
        resp_dir = Path(study_root) / glyph_dir / "runs" / study / "out" / "responses"
        if not resp_dir.is_dir():
            continue
        for cond in CONDITIONS:
            for txt in sorted(resp_dir.glob(f"{cond}_*.txt")):
                run = int(txt.stem.rsplit("_", 1)[1])
                yield glyph_dir, setup, cond, run, txt.read_text()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--study-root", required=True)
    ap.add_argument("--out", required=True)
    ap.add_argument("--glyph", action="append", help="limit to these glyph dirs")
    args = ap.parse_args()

    out = Path(args.out)
    (out / "slices").mkdir(parents=True, exist_ok=True)

    per_glyph = {}
    key = {}
    for glyph, setup, cond, run, text in collect(args.study_root):
        if args.glyph and glyph not in args.glyph:
            continue
        rid = opaque_id(setup + glyph, cond, run, text)
        per_glyph.setdefault(glyph, []).append({"id": rid, "text": text})
        key[rid] = {"glyph": glyph, "condition": cond, "setup": setup, "run": run}

    for glyph, items in sorted(per_glyph.items()):
        items.sort(key=lambda o: o["id"])  # stable, id-sorted; order leaks nothing
        slice_path = out / "slices" / f"{glyph}.jsonl"
        with slice_path.open("w") as fh:
            for obj in items:
                fh.write(json.dumps(obj) + "\n")
        print(f"{glyph}: {len(items)} responses -> {slice_path}")

    (out / "key.json").write_text(json.dumps(key, indent=2) + "\n")
    print(f"key: {len(key)} ids -> {out / 'key.json'}")


if __name__ == "__main__":
    main()
