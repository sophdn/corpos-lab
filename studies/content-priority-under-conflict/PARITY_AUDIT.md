# Parity audit — content priority under conflict

**Chain:** 543, task 2. **Date:** 2026-09-21. Audited before any run.

The measurement is clean only if the aligned and conflict blocks name the **same
facts** as the local instruction and differ only in the **action** they point to, and if
the precision ladder varies **one construct** (precision) and nothing else. This file
records that check per scenario.

## The one-variable rule

- **A_local vs A_canon** is the only thing that differs between `canon_aligned` and
  `canon_conflict`. Both blocks speak to the same decision (the same goal, the same
  artifacts, the same domain). `canon_aligned` points at `A_local`; `canon_conflict`
  points at `A_canon`.
- **Precision** is the only thing that differs across weak / medium / strong. All three
  state the same canon rule (the same `A_canon`); they differ in how strongly the rule is
  stated, not in what it says.

## Scenario A — config-target

- **Goal (constant):** set the EU request timeout to 60.
- **Local instruction (`scenario_1.md`):** edit `config/regions/eu.yaml` → `A_local`.
- **`canon_aligned`:** the per-region convention; the EU timeout is set in
  `config/regions/eu.yaml` → names `A_local`.
- **`canon_conflict_*`:** timeouts live only in `config/base.yaml`; region files must not
  set them → names `A_canon` (edit `config/base.yaml`).
- **Same facts:** both blocks name the EU timeout, the per-region files, and
  `config/base.yaml` as configuration surfaces. Only the target file differs.
- **Precision ladder (one variable):**
  - weak — states the rule once, plainly (2 sentences).
  - medium — same rule, adds emphasis ("important", "never") and one repetition.
  - strong — same rule, adds an authoritative titled frame ("ENGINEERING STANDARD …
    MANDATORY"), a rationale, and four mutually-reinforcing clauses. No new fact about
    the target beyond "the timeout goes in `config/base.yaml`".

## Scenario B — api-version

- **Goal (constant):** add a `priority` field to the orders API.
- **Local instruction (`scenario_1.md`):** edit `handlers/v1/orders.go` → `A_local`.
- **`canon_aligned`:** new order fields are added to the v1 handler
  `handlers/v1/orders.go` → names `A_local`.
- **`canon_conflict_*`:** `/v1` is frozen; new fields go on `handlers/v2/orders.go` only
  → names `A_canon` (edit `handlers/v2/orders.go`).
- **Same facts:** both blocks name the orders handler, the `priority`/new-field work, and
  the v1 and v2 handlers. Only the target handler differs.
- **Precision ladder (one variable):** weak states the freeze once; medium adds emphasis
  and one repetition; strong adds the "API GOVERNANCE STANDARD … MANDATORY" frame, a
  rationale, and reinforcing clauses. Same target rule throughout.

## Scenario C — record-location

- **Goal (constant):** record a completed bug fix.
- **Local instruction (`scenario_1.md`):** write `CHANGELOG.md` → `A_local`.
- **`canon_aligned`:** completed fixes are recorded in `CHANGELOG.md` → names `A_local`.
- **`canon_conflict_*`:** fix records go only in the issue tracker; `CHANGELOG.md` is
  auto-generated, do not hand-edit → names `A_canon` (write the issue tracker).
- **Same facts:** both blocks name the completed fix, the changelog, and the issue
  tracker. Only the target surface differs.
- **Precision ladder (one variable):** weak states the rule once; medium adds emphasis
  and one repetition; strong adds the "RELEASE PROCESS STANDARD … MANDATORY" frame, a
  rationale, and reinforcing clauses. Same target rule throughout.

## The two mechanism controls, matched to the strong conflict block

Both controls are read against `canon_conflict` at the strong level (the largest expected
override), holding form and length while removing one property.

- **`scrambled_canon`** — the strong conflict block with its content destroyed. Generated
  from `canon_conflict_strong.md` by `provenance/published-scoring/_shared/scramble.py <in> <out> 7 --vocab-swap
  --neutralize-title`, then the neutralized `# glyph` title line rewritten to `# lorem
  ipsum dolor` so no topical word leaks. Shape kept (line breaks, blank lines, header
  line); every content word replaced by a neutral lorem token. Word counts: A 123→120, B
  125→121, C 122→118. It names no file and no rule, so it cannot point at `A_canon`; any
  movement off `A_local` under it is prepended structure, not comprehension.
- **`off_target_canon`** — a coherent, authoritative "… STANDARD (MANDATORY)" block about
  a **different** domain, matched in form and length to the strong conflict block: A =
  structured logging, B = schema migrations, C = commit messages. Each is a real
  convention, stated with the same authority, but bears on nothing in the scenario, so it
  cannot point at `A_canon`. Any movement off `A_local` under it is prepended structure,
  not recognition of a relevant rule.

## Length caveat, stated up front

The precision ladder lengthens weak → strong by construction (emphasis, framing, and
rationale add words). Length is therefore confounded with precision within the ladder.
The `scrambled_canon` and `off_target_canon` controls hold length at the strong level
while removing content and relevance, so they separate "a long authoritative prepended
block" from "the strong conflict rule". A clean length-only follow-up (a length-matched
neutral block at each rung) is named here, not run here.
