#!/usr/bin/env python3
"""Produce the ablated terrain set from the base terrain set.

The perturbation arm removes the Marker axis and the Aim axis from each glyph
terrain, leaving the header, the Y-Decision block, and the Rest axis unchanged.
The transformation is purely structural: everything from the "### Marker axis"
line up to (but not including) the "### Rest axis" line is deleted, and nothing
else is touched. The output files are byte-clean terrain the model reads, so no
provenance note is added inside them — provenance lives in this script and in
the per-file diffs.

Run from the study directory:  python3 materials/make_ablated.py
"""
from __future__ import annotations

import pathlib
import sys

HERE = pathlib.Path(__file__).resolve().parent
BASE = HERE / "base"
ABLATED = HERE / "ablated"

MARKER = "### Marker axis"
REST = "### Rest axis"

GLYPHS = ["cas", "cgu", "fsb", "gop", "psc", "scb"]


def ablate(text: str, name: str) -> str:
    lines = text.splitlines(keepends=True)
    marker_idx = _find(lines, MARKER, name)
    rest_idx = _find(lines, REST, name)
    if not marker_idx < rest_idx:
        sys.exit(f"{name}: Marker axis does not precede Rest axis")
    # Keep header + Y-Decision (through the separator before Marker), then the
    # Rest axis to end. The '---' that preceded Marker now precedes Rest.
    return "".join(lines[:marker_idx] + lines[rest_idx:])


def _find(lines: list[str], header: str, name: str) -> int:
    hits = [i for i, ln in enumerate(lines) if ln.strip() == header]
    if len(hits) != 1:
        sys.exit(f"{name}: expected exactly one '{header}', found {len(hits)}")
    return hits[0]


def main() -> None:
    ABLATED.mkdir(exist_ok=True)
    for g in GLYPHS:
        src = BASE / f"GLYPH_{g}-terrain.md"
        dst = ABLATED / f"GLYPH_{g}-terrain.md"
        out = ablate(src.read_text(), g)
        if MARKER in out or "### Aim axis" in out:
            sys.exit(f"{g}: ablation left a Marker or Aim axis behind")
        if REST not in out:
            sys.exit(f"{g}: ablation dropped the Rest axis")
        dst.write_text(out)
        print(f"{g}: {len(src.read_text().splitlines())} -> {len(out.splitlines())} lines")


if __name__ == "__main__":
    main()
