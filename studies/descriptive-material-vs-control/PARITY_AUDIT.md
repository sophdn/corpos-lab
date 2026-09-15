# Content-parity audit — H1 (descriptive material vs control)

The 2×2 varies **form** (descriptive vs directive) and **grounding** (domain-free
vs domain-specific), with the propositional content held fixed, plus a baseline.
Two parity claims must hold, per class and per scenario:

1. **Domain-free pair — form only.** `glyph.md` (descriptive) and `imperative.md`
   (directive) state the same propositions; they differ only in mood. Both are
   domain-free (abstract roles, no named tool, file, or scenario).
2. **Domain-specific pair — mood only.** `ground_N.md` (descriptive) and
   `domain_imperative_N.md` (directive) state the same propositions; they differ
   only in mood.
3. **Non-copy.** Neither `ground_N.md` nor `domain_imperative_N.md` hands over the
   finished artifact. The model must perform the target action itself.

The glyph is the certified ALPHABET AC-4 block (frozen by digest). The imperative
was authored to match its content in directive mood.

---

## parent-state-check-bypass

**Domain-free pair (glyph = imperative).** Shared propositions:

- A work item belongs to a parent context that carries a required pre-execution
  state check for a new session.
- The check has not been run this session and feels like overhead unrelated to
  the work item.
- The correct configuration is a parent state check record preceding the
  work-item action in the session.
- Beginning the work item with no such record leaves the parent context's state
  unverified.

The glyph states these as decision-point description (Y-fire, Marker/Aim/Rest
invariants). The imperative states them as commands ("run the parent state check
before you begin the work item"). Both name abstract roles only.

**Domain-specific pair (ground = domain-directive), per scenario.** Shared
propositions:

- P1. The parent context's current state — [s1: in-scope tickets, milestone
  status, scope/priority changes | s2: latest upstream load, partition watermark,
  backfills/schema changes | s3: approval state, freeze windows, applied ordered
  steps] — is recorded in [s1: `milestones/release-2026-Q2.md` | s2:
  `manifests/sales_curated.yaml` | s3: `changes/CR-4471.md`].
- P2. The validity of the work item rests on that current parent state.
- P3. That state is known for this session/run only once the parent record has
  been read this session/run.
- P4. The work item is well-founded only when the parent record has been read
  before the work begins.

`ground_N.md` states P1–P4 declaratively; `domain_imperative_N.md` commands the
read ("Read [record] and confirm the current state before you begin"). Mood only.

**Non-copy:** neither aid reads the parent record for the model or supplies its
contents. The read is the model's to perform.

---

## post-write-verification-absent

**Domain-free pair (glyph = imperative).** Shared propositions:

- A sequence of state-mutating operations against a target has just completed;
  each returned a success acknowledgment; completion is about to be declared.
- The acknowledgment confirms the operation was issued and accepted, not that the
  persisted content matches intent; success and silent failure are
  indistinguishable from it.
- The authoritative source for the resulting state is a read of the target.
- The correct configuration is a read of the target's resulting state after the
  final operation and before the completion statement.

The glyph states these as decision-point description; the imperative as commands
("read the target back and confirm before declaring done"). Both domain-free.

**Domain-specific pair (ground = domain-directive), per scenario.** Shared
propositions:

- P1. The acknowledgment ([s1: `Applied edit: 1 change` | s2: `UPDATE — statement
  OK` | s3: `Modified`]) confirms the operation was issued and accepted, not the
  resulting content; success and silent failure return the same acknowledgment.
- P2. The authoritative source for the resulting state is a read of [s1:
  `server/config.py` | s2: the `accounts` table | s3: `api/reference.md`].
- P3. The task is done only when that target has been read back and shows the
  intended content ([s1: timeout 60 | s2: `region = 'EU'` for DE/FR/NL | s3: no
  `/v1/orders` remain]).

`ground_N.md` states P1–P3 declaratively; `domain_imperative_N.md` commands the
read-back. Mood only.

**Non-copy:** neither aid performs the read or supplies the target's resulting
content. The read is the model's to perform.

---

## initiative-task-preexistence-gate

**Domain-free pair (glyph = imperative).** Shared propositions:

- An author is assigning a readiness/actionable status to a tracking entry whose
  backing artifact does not exist at any discoverable location.
- A readiness status is a present-state claim a consumer acts on; planning intent
  is not a created artifact.
- The correct configuration is a readiness status only for a confirmed-existing
  artifact; otherwise an honest planning status.

The glyph states these as decision-point description; the imperative as commands
("confirm the artifact exists before assigning a readiness status; else an honest
planning status"). Both domain-free.

**Domain-specific pair (ground = domain-directive), per scenario.** Shared
propositions:

- P1. Readiness rests on whether [s1: `deploy/rollback.sh` | s2:
  `api/handlers/export_csv.py` | s3: `tests/integration/payment_webhook_test.py`]
  exists at a discoverable location — not on [s1: approach worked out | s2: design
  agreed | s3: cases specified] and not on implementation steps having been taken.
- P2. The status a consumer reads is a present-state claim; planning progress is
  not a created artifact.
- P3. The entry is ready only when the artifact is confirmed to exist; otherwise
  the honest status is a planning status that makes no readiness claim.

`ground_N.md` states P1–P3 declaratively; `domain_imperative_N.md` commands the
existence check before a readiness status. Mood only. **Neither asserts whether
the artifact exists** — that is the model's to check.

**Non-copy:** neither aid creates the artifact or supplies its contents.

---

## Cross-level parity

Each class carries one core proposition at two grounding levels. Domain-free
(glyph, imperative): "verify the [parent state | resulting state | backing
artifact] before [starting | completing | claiming ready]." Domain-specific
(ground, domain-directive): the same, instantiated in the scenario's named
record, target, or artifact. Content is held constant; only form and grounding
vary.

## Residual the design cannot remove

The glyph is a three-axis structured block; the other three aids are prose. The
glyph-vs-ground contrast mixes grounding with block structure. The
imperative-vs-domain-directive contrast (prose vs prose) is the cleaner grounding
read; the glyph-vs-imperative contrast (both domain-free) is the cleaner form
read. The optional scrambled and off-target controls bound the structure and
recognition questions separately.
