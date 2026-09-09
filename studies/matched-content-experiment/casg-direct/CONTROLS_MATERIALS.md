# Control materials — casg-direct (mechanism-controls, chain 522)

Two added conditions for the mechanism controls. Both pair with the SAME casg-direct
scenario (materials/scenario.md) and run under the SAME no-tools notice as T0/T1/T2.

## scrambled_glyph (structure vs comprehension)
- File: materials/scrambled_glyph.md
- Method: word-order shuffle within each prose span of materials/glyph.md, seeded
  (scripts/scramble.py, random.Random(42)). Structure lines preserved verbatim:
  the title, Y-fire / Y-not-fire labels, the Marker / Aim / Rest axis headers, and
  every bold label (Invariant, Firing condition, Does not fire on, Violation signal,
  Recognition signal, Characterization, Distinguishing condition), blockquote and
  bullet markers, and the --- rules. Only the descriptive words are reordered.
- Parity: identical word count to glyph.md (820 = 820) and identical three-axis
  shape; content is incomprehensible (the decision logic cannot be recovered).
- Reads against glyph_only: if it reproduces the glyph's register effect (Ii),
  the effect is structural to the format; if it collapses toward baseline, the
  effect requires comprehension of the content.

## off_target_glyph (recognition in the loop)
- File: materials/off_target_glyph.md
- Content: the conditional-gate-uniform-default glyph.md VERBATIM (fixed rotation:
  each class takes the next class's glyph; casg-direct <- conditional-gate).
  A coherent glyph about withholding a deploy pending a condition — unrelated to
  the release/changelog companion-artifact decision in the scenario.
- Reads against glyph_only: equal register effect on- and off-target implicates
  glyph structure over recognition; effect only on the matching glyph implicates
  recognition of the scenario as in the loop.

## Digests (sha256)
- glyph.md              78b28398bc603370fbbe7ff61e57a8040162f5acd4a4419055a437512af73af7
- scrambled_glyph.md    a8749b1542703ec3260a66627e50026cc3059df2b9e2e372f614b186458c2413
- off_target_glyph.md   02f2c7c404cdb108ffd9241949c24945dbc656944609a0592ef4eb9545835d1b
- (off-target source: conditional-gate glyph.md 02f2c7c404cdb108ffd9241949c24945dbc656944609a0592ef4eb9545835d1b)

## Parity + comprehension audit (independent, 2026-09-08)
Independent auditor (blind to authoring intent), verdict **SIGN-OFF**:
- scrambled_glyph.md: same three-axis shape as glyph.md; length identical (820=820
  words); incomprehensible — word-salad, the decision cannot be recovered (a reader
  can only pattern-match repeated nouns, not the choice/trigger/outcomes).
- off_target_glyph.md: coherent, correctly structured; describes conditional-branch-
  vs-uniform-default (a deploy-gate mechanism) — no companion artifact, no release,
  no changelog. Clearly does NOT match the release/changelog scenario.
