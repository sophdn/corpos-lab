#!/usr/bin/env python3
"""Build a scrambled-glyph control material from a real glyph.

A scrambled glyph is the structure/comprehension control for the matched-content
program: it preserves a glyph's three-axis SHAPE and length but destroys the
decision content, so — read against the real glyph — it separates the effect of
the glyph's format-structure from the effect of comprehending its content. See
INQUIRY.md "How we measure" -> "Mechanism controls".

Two modes and a title switch (suggestion refine-scrambled-glyph-control-title-and-
vocabulary-leak):

- mode=shuffle (default): word order shuffled within each prose span, seeded. This
  removes syntax/comprehensibility but KEEPS the class's own topical vocabulary in
  shuffled order — so keywords (milestone, changelog, provenance) still leak. This
  is the original control; the default output is unchanged for reproducibility.
- mode=vocab-swap: each content word is replaced with a neutral lorem-style token
  (seeded, word count preserved), removing BOTH syntax AND topical vocabulary while
  keeping the three-axis shape and length. The stronger structure-only control.
- --neutralize-title: replace the "# <class-slug>" title with a neutral "# glyph"
  so a descriptive class name (e.g. parent-state-check-bypass) does not name the
  decision. Recommended for new scrambled controls; off by default so existing
  materials reproduce byte-for-byte.

In every mode the structure is preserved verbatim: blank lines, `---` rules,
markdown headers, and each line's leading prefix (a `>` blockquote marker, a `-`
bullet, and a `**Label:**` bold label). Only the descriptive words are touched.

Usage:
    scripts/scramble.py <glyph.md> <out.md> [seed] [--mode shuffle|vocab-swap] [--neutralize-title]
"""
import re
import random
import sys

# prefix = optional blockquote '> ', optional bullet '- ', optional bold label ending ':**'
PREFIX_RE = re.compile(r"^(\s*>\s*)?(-\s+)?(\*\*[^*]+?:\*\*\s*)?")
TITLE_RE = re.compile(r"^#\s+\S")

# A fixed neutral filler pool (lorem-ipsum family): coherent-looking tokens with
# no topical meaning, so vocab-swap removes keyword leakage.
LOREM = (
    "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor "
    "incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud "
    "exercitation ullamco laboris nisi aliquip ex ea commodo consequat duis aute "
    "irure reprehenderit voluptate velit esse cillum fugiat nulla pariatur excepteur "
    "sint occaecat cupidatat non proident sunt culpa qui officia deserunt mollit anim"
).split()


def _neutralize_title(line: str) -> str:
    return "# glyph"


def _transform_line(line: str, mode: str, rng: random.Random, lorem_i: list) -> str:
    m = PREFIX_RE.match(line)
    prefix = m.group(0) if m else ""
    words = line[len(prefix):].split()
    if mode == "vocab-swap":
        swapped = []
        for _ in words:
            swapped.append(LOREM[lorem_i[0] % len(LOREM)])
            lorem_i[0] += 1
        words = swapped
    elif len(words) > 1:
        rng.shuffle(words)
    return prefix + " ".join(words)


def transform(src_text: str, seed: int = 42, mode: str = "shuffle", neutralize_title: bool = False) -> str:
    rng = random.Random(seed)
    lorem_i = [0]  # mutable cursor into LOREM for vocab-swap
    out = []
    for line in src_text.split("\n"):
        if line.strip() == "" or line.strip() == "---":
            out.append(line)
        elif TITLE_RE.match(line.lstrip()) or line.lstrip().startswith("#"):
            out.append(_neutralize_title(line) if (neutralize_title and TITLE_RE.match(line.lstrip())) else line)
        else:
            out.append(_transform_line(line, mode, rng, lorem_i))
    return "\n".join(out)


# Back-compat alias: the original shuffle-only entry point.
def scramble(src_text: str, seed: int = 42) -> str:
    return transform(src_text, seed, mode="shuffle", neutralize_title=False)


def main(argv: list) -> int:
    args = [a for a in argv[1:] if not a.startswith("--")]
    flags = [a for a in argv[1:] if a.startswith("--")]
    if len(args) < 2:
        print("usage: scramble.py <glyph.md> <out.md> [seed] [--vocab-swap] [--neutralize-title]", file=sys.stderr)
        return 2
    seed = int(args[2]) if len(args) > 2 else 42
    mode = "vocab-swap" if "--vocab-swap" in flags else "shuffle"
    neutralize = "--neutralize-title" in flags
    with open(args[0]) as fh:
        text = fh.read()
    with open(args[1], "w") as fh:
        fh.write(transform(text, seed, mode=mode, neutralize_title=neutralize))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
