# Study materials manifest — matched-content experiment (Q1)

Chain 423 (glyph-research), task 3486 (author-matched-materials). Authored
2026-09-07.

Records each material's content digest so a run record can say exactly what the
model saw. **Recorded, not frozen** (INQUIRY.md "Record what ran, automatically";
PROTOCOL section on digests). A digest that differs from a prior run's is
information about the two runs — never grounds for refusing to run. Post-audit
edits reopen the audit for the affected pair (PROTOCOL 2.4); they do not lock the
file.

## Conditions (per class)

- **T0** baseline — `scenario.md` only.
- **T1** glyph — `glyph.md` (ALPHABET entry verbatim) + `scenario.md`.
- **T2** imperative — `imperative.md` (information-matched directive) + `scenario.md`.

## Content-parity audit

Each T1/T2 pair audited by an independent session that authored neither text
(PROTOCOL 2.4). Records in each class's `PARITY_AUDIT.md`.

| Class | Verdict | Note |
|-------|---------|------|
| casg-direct | SIGN-OFF | 13 propositions matched, first pass |
| conditional-gate-uniform-default | SIGN-OFF | 17 propositions matched, first pass |
| formal-step-context-bypass | SIGN-OFF | 19 propositions matched, first pass |
| parent-state-check-bypass | SIGN-OFF | flagged 1 (missing "already checked this session" exemption), fixed, re-audited by a second independent session |

## Digests (sha256)

| Class | Material | sha256 |
|-------|----------|--------|
| casg-direct | scenario | `fdf92ed205d0eb1982062195e01c94450d7680062546215206d8c4bdbb1310b0` |
| casg-direct | glyph | `78b28398bc603370fbbe7ff61e57a8040162f5acd4a4419055a437512af73af7` |
| casg-direct | imperative | `f4a5e0eb5fee557bec0e05cfcaf414cc8314fbb1bb4c60613875048a9b693466` |
| parent-state-check-bypass | scenario | `c24185feb932f1f3315140ae336d2d4c5d3e671045e50fadaa527e43be4cf0e3` |
| parent-state-check-bypass | glyph | `5de449a5cd2674fb205ad2aacb8bb90d85d82f949262dfc050e267d2b0cb5e9d` |
| parent-state-check-bypass | imperative | `5bb17d6549507dd1a58745db688d3f7a597a3fad32ae6a7eab5a24088d927387` |
| conditional-gate-uniform-default | scenario | `7e8b5f5c7e51dbcfff03ca731d7c78cf182237eb953268d3800b5482286ecae6` |
| conditional-gate-uniform-default | glyph | `02f2c7c404cdb108ffd9241949c24945dbc656944609a0592ef4eb9545835d1b` |
| conditional-gate-uniform-default | imperative | `0f40d168c56730c73db9b0cd176b16afbe1ec14d4d8611c44636ebc6bbce5a38` |
| formal-step-context-bypass | scenario | `667dcae94f6e226bbccad9139847cb2c4a6924e635b91f790922a5b51da0a96c` |
| formal-step-context-bypass | glyph | `13438e52d4c161debf655c134ce02312eb6d628b813106d5162dfb124597a03f` |
| formal-step-context-bypass | imperative | `12d67bb5f7758382d34e0792b149856c55c9cfd34cb61e753a35f466483e41ff` |

Regenerate: `sha256sum <class>/materials/<material>.md`.

## Word-count ratio (T1 glyph : T2 imperative)

PROTOCOL 2.3 flags a ratio above 3:1. All within budget — the difference is the
glyph's structural overhead (axis headers, invariant framing, Y-not-fire), which
is part of the format under study and not padded into T2.

| Class | T1 words | T2 words | ratio |
|-------|----------|----------|-------|
| casg-direct | 820 | 389 | 2.11 |
| parent-state-check-bypass | 943 | 405 | 2.33 |
| conditional-gate-uniform-default | 952 | 408 | 2.33 |
| formal-step-context-bypass | 942 | 442 | 2.13 |

## Open items for the execute task (3487)

- **Runner condition support.** `internal/assay/grounded.go` supports
  `baseline / glyph_only / grounded_glyph` with glyph + ground material slots. The
  T2 condition — imperative + scenario, no glyph — has no slot. The execute task
  needs a condition mapping (baseline→T0, glyph_only→T1, and a new imperative
  condition→T2) or a new assay before the grid can run.
- **T0 calibration** culls non-calibrating cells; keep >= 3 classes (PROTOCOL 3,
  11). Four classes are authored to leave headroom.
- **PREDICTIONS.md** written before the first run, never in a subject-visible file.
- **study.toml** per class (or combined), digest-pinned image, sampler chain from
  PROTOCOL 6.
