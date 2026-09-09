#!/usr/bin/env python3
"""Build a scrambled-glyph control material from a real glyph.

A scrambled glyph is the structure/comprehension control for the matched-content
program: it preserves a glyph's three-axis SHAPE and exact length but shuffles the
word order within each prose span so the decision logic cannot be recovered. Read
against the real glyph, it separates the effect of the glyph's format-structure
(if the scramble reproduces the effect) from the effect of comprehending its
content (if it collapses toward baseline). See INQUIRY.md "How we measure" ->
"Mechanism controls".

Method: structure lines are preserved verbatim — the title, blank lines, `---`
rules, markdown headers (`#`), and the leading structural prefix of each line
(a `>` blockquote marker, a `-` bullet, and a `**Label:**` bold label). Only the
descriptive words after that prefix are reordered, with a FIXED seed so the output
is reproducible and can be re-derived from the source glyph.

Usage:
    scripts/scramble.py <glyph.md> <scrambled_glyph.md> [seed]

Default seed is 42. Word count is preserved exactly (a within-span permutation),
so the scrambled control is length-matched to its source by construction.
"""
import re
import random
import sys

# prefix = optional blockquote '> ', optional bullet '- ', optional bold label ending ':**'
PREFIX_RE = re.compile(r"^(\s*>\s*)?(-\s+)?(\*\*[^*]+?:\*\*\s*)?")


def scramble(src_text: str, seed: int = 42) -> str:
    rng = random.Random(seed)
    out = []
    for line in src_text.split("\n"):
        if line.strip() == "" or line.lstrip().startswith("#") or line.strip() == "---":
            out.append(line)
            continue
        m = PREFIX_RE.match(line)
        prefix = m.group(0) if m else ""
        words = line[len(prefix):].split()
        if len(words) > 1:
            rng.shuffle(words)
        out.append(prefix + " ".join(words))
    return "\n".join(out)


def main(argv: list[str]) -> int:
    if len(argv) < 3:
        print("usage: scramble.py <glyph.md> <scrambled_glyph.md> [seed]", file=sys.stderr)
        return 2
    seed = int(argv[3]) if len(argv) > 3 else 42
    with open(argv[1]) as fh:
        text = fh.read()
    with open(argv[2], "w") as fh:
        fh.write(scramble(text, seed))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
