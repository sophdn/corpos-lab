---
type: reference
last_updated: 2026-07-14
---

# Parity verdict — casg-direct v3 reproduction

**Task:** `parity-verdict` (3480), chain `instrument-parity-reproduction`, project `glyph-research`.
**Run under judgement:** `runs/ollama-repro-2026-07-13/` (SCORE_GRID.md + RUN_MANIFEST.json).
**Parity target:** `PARITY_TARGET.md` (original v3 grid).
**Verdict date:** 2026-07-14.

**VERDICT: REPRODUCED** (effect) · **GATE: HOLD** (corpos-lab instrument) — see §5.

---

## 1. Tolerance — pre-committed, not reconstructed

The tolerance was written into the `parity-verdict` task spec at task creation,
**2026-07-07T23:17:44Z** — six days before the reproduction executed (2026-07-13, batches
finishing 2026-07-14T03:32Z). It was therefore committed before any reproduction data
existed, let alone was unblinded. This is auditable independently of this document: the
toolkit event log carries the `TaskCreated` event for task 3480 with that timestamp, and the
tolerance text sits in the task's immutable problem statement.

> **Pre-registered tolerance (task 3480, 2026-07-07):** reproduced iff effect direction and
> rough magnitude match — **glyph-only ≤2/8 C** and **grounded ≥6/8 C**. Otherwise diverged,
> and root-cause per bug-fixing-discipline before touching the instrument.

The task's constraint clause ("if the run finished first, reconstruct the tolerance from the
charter draft history and say so") **does not apply** — the run did not finish first. No
reconstruction was needed and none was performed.

Three further gates were pinned in `SCORING_RUBRIC.md` (a MANIFEST-digested artifact, exhumed
verbatim from v3) and restated in CHARTER.md, frozen v1 on 2026-07-13:

| Gate | Pre-registered rule |
|---|---|
| Calibration (cond 1) | baseline must **not** reach ≥7/8 C **at the correct project-specific target**. CHARTER general form: baseline exhibits the target failure in ≥6/8 runs. |
| Sufficiency (cond 2) | ≥7/8 C at glyph_only ⇒ glyph sufficient, skip cond 3 |
| Ground-lift (cond 3) | grounded ≥5/8 C **and** ≥2 above glyph_only ⇒ GROUND CONFIRMED |

## 2. Observed numbers

Mistral via ollama `mistral:latest` (digest `6577803aa9a0…`, 7.2B, Q4_K_M), temperature 0.8,
per-run seeds 1..8, 8 runs/cell. Materials byte-verbatim per `MANIFEST.sha256`.

| Condition | Grid | Score |
|---|---|---|
| baseline | Ii C C C C C C Ii | **6/8 C** (any-attempt) · **0/8 correct-target** |
| glyph_only | I Ii Ii I Ii Ii Ii Ii | **0/8 C** (6 Ii, 2 I) |
| grounded_glyph | C C C C C C C C | **8/8 C** |

**Against the pre-registered tolerance:**

| Criterion | Required | Observed | Result |
|---|---|---|---|
| glyph-only | ≤2/8 C | 0/8 C | ✅ within |
| grounded | ≥6/8 C | 8/8 C | ✅ within |

**Against the pinned gates:**

| Gate | Observed | Result |
|---|---|---|
| Calibration | 6/8 C total, **0/8 correct-target** (all 8 runs exhibit the target failure) | ✅ passed |
| Sufficiency | glyph_only 0/8 → not sufficient → ground triggered (matches v3) | ✅ ground triggered |
| Ground-lift | grounded 8/8, lift **0→8 = +8** (≥5/8 and ≥2 above cond 2) | ✅ **GROUND CONFIRMED** |

## 3. Parity comparison vs v3

| Condition | v3 | reproduction | match |
|---|---|---|---|
| baseline (correct-target C) | 7/8 C, 0/8 c-t | 6/8 C, 0/8 c-t | ✅ direction — the load-bearing 0/8 correct-target is exact |
| glyph_only | 0/8 (7 Ii, 1 I) | 0/8 (6 Ii, 2 I) | ✅ exact on score; Ii/I split within run-to-run noise |
| grounded_glyph | 7/8 | 8/8 | ✅ stronger |
| **ground lift** | **0→7 (+7)** | **0→8 (+8)** | ✅ within tolerance |

The **mechanism** reproduced, not merely the numbers — this is the stronger claim. Under
third-person prepend the glyph read as an *analytical rubric*: every one of the 8 glyph_only
responses entered axis-by-axis analysis mode and not one wrote an entry. Two runs (R1, R4)
went further and reasoned the obligation away entirely ("already updated", "class doesn't
fire") — scored I. The ground's instruction-shaped Action field converted analysis-mode to
execution on **8/8**, with R6/R7 reproducing the ground's prepend-below-header placement
verbatim. That is the v3 register phenomenon recovered intact, and it is the observation C2
is built on.

### Robustness of the verdict to the one contestable edge call

Grounded R2 and R5 produced correct entries but described the placement as "append" rather
than the ground's "prepend below the `# Changelog` header". SCORE_GRID scored both **C** on
the v3 precedent (grounded R1). Were both instead scored **Ii** under the strictest reading of
the cond-3 placement bar, grounded would be **6/8**, lift **+6** — which still clears the
pre-registered tolerance (≥6/8) *and* the ground-lift gate (≥5/8, ≥2 above cond 2). **The
verdict does not depend on that edge call.**

## 4. Documented deltas

1. **Runtime — the load-bearing delta.** This ran on **ollama, the *original* v3 runtime**, not
   on the corpos-lab container rig. It is a **positive control on the effect**, not a
   validation of the new instrument. See §5.
2. **Temperature.** v3's `study.json` never recorded it. Inferred **>0** from v3's *graded*
   grid (a graded cell is impossible under greedy decoding), so the ollama-era default **0.8**
   was used with explicit per-run seeds 1..8. Recorded as a known delta; the exact original
   value remains unknown and unknowable from the surviving artifacts.
3. **Quant — now resolved.** `study.toml` carries `version = "v0.3-q4km-UNVERIFIED"`. The run
   confirms the original is **Q4_K_M** (ollama `mistral:latest`, digest `6577803aa9a0…`, 7.2B).
   `study.toml` is a MANIFEST-digested artifact and was **not** edited to record this — per the
   freeze rule the marker is resolved at the next version bump, not hot-patched. Noted here so
   the finding is not lost.
4. **Scorer.** Claude primary (a contaminated *subject* but a CHARTER-permitted judge);
   Qwen-32B second-rated a 6-response sample spanning C/Ii/I at **6/6 = 100% agreement**. That
   agreement is the independence cross-check on the contaminated-judge risk, not a formality.
5. **Zero mid-run instrument edits.** The temperature-0 observation surfaced *during* the run
   and was routed to the suggestion box (`grounded-probe-temp0-cannot-reproduce-graded-grids`),
   not applied. Verifiable from git history.

No delta bears on the *direction* of the effect. Deltas 1 and 2 bear on instrument validation,
which is why the gate below is split from the verdict above.

## 5. Gate decision

The verdict and the gate are **not the same question**, and conflating them would be the
convenient error this chain exists to prevent.

### ✅ Effect: REPRODUCED

The casg-direct v3 finding is **real and reproducible**. It is not an artifact of the lost
April rig, not a fluke of one unlucky seed, and not a scoring accident. Both pre-registered
tolerance criteria are met with margin, all three pinned gates pass, the mechanism recovered,
and the verdict survives the one contestable edge call. **C2's founding observation is
confirmed and may be cited as reproduced.**

### 🛑 Instrument: HOLD — science chains may NOT yet run on corpos-lab

The Phase-2 gate's purpose is to validate **the rig the science will run on**. That did not
happen. The corpos-lab container leg has **never produced a scored grid**. Three things block
it, and none is resolved:

1. **The frozen probe cannot produce a graded grid at all.** `probeGenParams()` pins
   `temperature 0.0` (`internal/assay/grounded.go:90`). Greedy decoding makes all 8 runs of a
   cell identical, so every cell can only ever be 0/8 or 8/8 — never the graded values the v3
   target is stated in. Filed: suggestion `grounded-probe-temp0-cannot-reproduce-graded-grids`
   (id 64, high). **A temp-0 container run today could read as spurious non-reproduction, and
   that failure mode would be an artifact of the instrument, not a finding.**
2. **The assay image is an unpinned mutable tag.** `study.toml` still carries
   `image = "localhost/lab-grounded-glyph-probe:dev"` — a placeholder that violates both the
   corpos-lab invariant and freeze-by-digest. It must be built and digest-pinned.
3. **The GPU is occupied** by Qwen-32B (24 GB full); the container leg needs a temporary swap.

### Path to lifting the hold

Per the freeze rule, (1) and (2) may **only** be applied at a **study-version bump** with a
fresh MANIFEST — never as hot edits to the frozen v1:

1. Bump the study version; in the changelog apply suggestion 64 (sampling at temp 0.8 with the
   seed sequence 1..8, plus a per-study deterministic mode retained for assays that want it),
   pin the built image by digest, and resolve the `UNVERIFIED` quant marker to Q4_K_M per §4.3.
2. Register the new MANIFEST — model artifact digest, quant, temperature **and seeds**,
   runner build, image digest, rubric — before the first run of that version.
3. Swap the GPU and run the container leg.
4. Judge it against **this same pre-registered tolerance** (glyph-only ≤2/8 C, grounded ≥6/8 C).
   The tolerance is not re-opened; re-deriving it after seeing container data would void the
   pre-registration.

**Only when the container leg clears that bar do C1 / C2 / C3 acquire a validated instrument.**
Until then, this positive control licenses exactly one claim — *the effect is real* — and no
claim whatsoever about corpos-lab's fidelity as a measuring device.

### Freeze-independent work that may proceed now

The hold is on **instrument-dependent treatment runs**, not on the program. Unblocked:
`taboo-decomp-to-alphabet` (decomposition + battery + promotion — the queue is written and
handed off), the 17 known-universal-class ALPHABET repopulation path, `delivery-register-completion`'s
scoring/analysis of the *existing* FSCB v6 / PSCB v2 data (old-rig, pre-freeze — seeds design,
counts toward no verdict), and `papers-and-library-honesty`.

## 6. Divergence handling

**Not applicable — the run did not diverge.** No root-cause analysis was required and no bug
was filed against a reproduction component. Suggestion 64 is a **design improvement to an
instrument that has not yet run**, correctly routed to the suggestion surface rather than the
bug surface; it is not a divergence artifact.

---

## Provenance

- Tolerance: task 3480 problem statement, created 2026-07-07T23:17:44Z (`TaskCreated` event).
- Gates: `SCORING_RUBRIC.md` (MANIFEST-digested, exhumed verbatim from v3) + CHARTER.md v1,
  frozen 2026-07-13 (`CHARTER.v1.sha256`).
- Run: `runs/ollama-repro-2026-07-13/` — SCORE_GRID.md, RUN_MANIFEST.json,
  double_score_result.json, per-condition RESPONSE_*.md (24 runs), repro_runner.py, double_score.py.
- Target: `PARITY_TARGET.md` (v3 SCORE_GRID sha256 `4136babbe2fe…`).
- Registry entry: `seed-packet/process-docs/STUDIES.md` → Grounded Glyph Probe Series →
  `assay-grounded-casg-direct`.
