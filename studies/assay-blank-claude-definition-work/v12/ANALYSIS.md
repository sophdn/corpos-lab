# Analysis — assay-blank-claude-definition-work-v12

**Study:** assay-blank-claude-definition-work-v12
**Date:** 2026-04-01
**Analyst:** Claude (claude-sonnet-4-6)

---

## Step 1 — Verdict scores

| Scenario | Ground truth | v10 | v11a | v11b | v12 |
|----------|-------------|-----|------|------|-----|
| gop-a | yes | — | 4/4 | 4/4 | 4/4 |
| gop-b | no | 1/4 | 4/4 | 1/4 | **4/4** |
| cas-a | yes | 3/4 | 0/4 | 0/4 | **3/4** |
| cas-b | no | 4/4 | 2/4 | 4/4 | **3/4** |

---

## Step 2 — Carve-out field source (b-scenarios)

**gop-b** (4/4 correct):

| Run | Verdict | Field source | Classification |
|-----|---------|-------------|----------------|
| R1 | C | Artifact coupling | Coupling-cited |
| R2 | C | Artifact coupling | Coupling-cited |
| R3 | C | Scope - Downstream-of-consultation | Scope-cited (discriminator state name from coupling block) |
| R4 | C | Scope — not operative when | Scope-cited |

R3 cites the discriminator state name "Downstream-of-consultation," which is defined inside the coupling block rather than the base scope field — the boundary between Scope-cited and Coupling-cited is blurred here. R3 and R4 cite scope-field labels; R1 and R2 cite the artifact coupling block header directly.

**cas-b** (3/4 correct):

| Run | Verdict | Field source | Classification |
|-----|---------|-------------|----------------|
| R1 | I | Scope - operative when | Scope-cited (wrong verdict) |
| R2 | C | Artifact coupling | Coupling-cited |
| R3 | C | Artifact coupling | Coupling-cited |
| R4 | C | Scope — not operative when | Scope-cited |

No Navigate-cited field sources in any cas-b run.

---

## Step 3 — Aggregate observable and evidence quality

| Scenario | Verdict | Observable | Evidence |
|----------|---------|-----------|---------|
| gop-a | 4/4 | C | C |
| gop-b | 4/4 | C | P |
| cas-a | 3/4 | P | P |
| cas-b | 3/4 | P | P |

**gop-b observable (C):** All four runs cite the discriminating condition correctly — the operation is traceable to a prior consultation output and is not independently initiated. The carve-out condition is named precisely.

**gop-b evidence (P):** R1 and R2 cite specific trace steps (step 4 prescribed by step 1). R3 cites "Trace steps 1 to 4" and R4 cites "Trace Step 4" — true but not specific enough to distinguish the prescription chain. Aggregate P.

**cas-a observable and evidence (P):** R1–R3 reach correct verdicts through partially correct reasoning. R1 cites "No companion artifact update" without checking for delegation action — incomplete. R2 cites "delegation action is absent" (correct for this scenario but citing an Obligation-held discriminator state, which is correct). R3 cites the coupling block boundary type header rather than the discriminating condition. R4 is I (hallucination, see below).

**cas-b observable and evidence (P):** R2 and R3 cite the coupling block boundary type label ("executor → companion-update-assigned role") rather than the "Delegation complete" discriminator state name. The reasoning is directionally correct but does not name the state the coupling block defines. R4 cites the scope-not-operative condition fully and correctly. R1 is I.

---

## Step 4 — Comparison against baselines

**gop-b:** Full recovery. v12 matches v11a's 4/4 while the v12 format also includes the coupling block. The Navigate block is sufficient to route the verdict correctly; the coupling block additionally characterizes the state before routing.

**cas-b:** Partial regression from v11b. v11b achieved 4/4; v12 achieves 3/4. One run (R1) misread the delegation action in the trace and reached a wrong verdict via the scope-operative condition. The regression is trace-reading, not format.

**cas-a:** Improvement from v11a/v11b (both 0/4) to 3/4. This was not a target of v12 and was not predicted. See Step 8 for interpretation.

---

## Step 5 — Block interaction analysis

**gop-b:** In R1 and R2, the coupling block is the primary field source; Navigate is not cited. In R3, the "Downstream-of-consultation" discriminator state — which is defined in the coupling block's scope discriminator section — is cited as the field source; this is a coupling block output routed without Navigate being named. In R4, the base scope-not-operative field is cited directly. In no run does Navigate appear as the explicit field source, but the coupling block's characterization of the consultation-prescription relationship is load-bearing in R1–R3. Observable and Evidence confirm the model is reasoning through the coupling chain (consultation at step 1 → prescribed artifact at step 4) rather than short-cutting to a verdict. The blocks are operating in sequence for gop-b.

**cas-b:** In R2 and R3, the coupling block boundary type is cited, but the discriminator state name ("Delegation complete") is not explicitly named — the model recognizes the coupling boundary structure without explicitly stepping through the discriminator. The EVIDENCE in R2 and R3 is specific to the trace (step 5: Registry Coordinator change request), so the trace read is correct even if the state name is not cited. R4 routes correctly through the scope-not-operative field. R1 fails at trace reading before the coupling block's discriminator state can fire.

---

## Step 6 — cas-b mechanism check

Pre-check: does any correct cas-b verdict route through Navigate as the primary field source?

- R2: Coupling-cited (Artifact coupling header). Navigate not cited. ✓
- R3: Coupling-cited (Artifact coupling header). Navigate not cited. ✓
- R4: Scope-cited (Scope — not operative when). Navigate not cited. ✓

No correct cas-b verdict routes through Navigate as primary field source. No wrong-path successes. The coupling block or base scope field is the active lever for all three correct runs. The cas-b concern from v11a — that Navigate would override the coupling block's scope determination — does not appear in v12.

---

## Step 7 — Navigate ordering as between-glyph confound

- gop-Navigate leads with the yes-branch (independently initiated → yes).
- cas-Navigate leads with the no-branch (delegation complete → no).

gop-b: 4/4 despite leading with yes-branch. Navigate ordering is not suppressing the no verdict for gop.
cas-b: 3/4 with leading no-branch. The single miss (R1) is a trace-reading failure that occurs before Navigate is reached — the model incorrectly evaluates the delegation action's presence and routes through the scope-operative condition. Navigate ordering is not the cause of R1's failure.

Ordering is not a contributing factor to the differential results in v12.

---

## Step 8 — Additivity assessment

**gop-b:** Additive. The coupling block characterizes the consultation-prescription relationship; the discriminator state "Downstream-of-consultation" is cited in R3 as the field source, and the coupling-to-prescription chain is present in the Observable and Evidence of R1 and R2. Navigate is present in the format but is not explicitly cited — the coupling block's state characterization appears sufficient to route the verdict in all four runs. The Navigate block may be reinforcing rather than routing, but it is not overriding.

**cas-b:** Partially additive. The coupling block is the primary active lever for R2 and R3. The base scope field is the lever for R4. Navigate is not cited in any run. The failure in R1 is pre-coupling: the model misreads step 5 as not being a delegation action and routes through scope-operative before the coupling discriminator fires. The coupling block does not prevent this failure because the trace read is wrong before the discriminator state can be evaluated.

**Implication:** The two blocks are not interfering with each other (the concern from the null-result conditions). Navigate appears inert as an explicit routing mechanism in v12 — the coupling block and base scope field are doing the resolving work. Navigate may still be providing a structural anchor (a chain endpoint to route toward) without being cited as the field source.

---

## cas-a improvement — mechanism analysis

cas-a improved from 0/4 (v11a and v11b) to 3/4 (v12). The PROBE designated cas-a as a diagnostic scenario with an expected result of 0/4; improvement was flagged as evidence that positive-framing is doing more than expected.

**What happened in the 3 correct runs (R1–R3):**
- R1: Cites Scope field, observes "No companion artifact update" — partially correct, routes through scope-operative without delegation action check. Verdict correct.
- R2: Cites Artifact coupling, observes "The delegation action is absent in the trace" — correctly applies the Obligation-held state. Verdict correct.
- R3: Cites Artifact coupling boundary type — recognizes the coupling structure and applies it to reach a correct yes verdict. Verdict correct.

**What happened in R4 (wrong):**
- R4 cites "Scope — not operative when" and claims "Delegation action ('Documentation Coordinator') appears in the trace." No Documentation Coordinator appears in cas-a's trace. The model imported the calibration instance from the glyph file (which mentions a Documentation Coordinator as an illustration), hallucinated it into the scenario trace, and routed to a wrong no verdict.

**Source of improvement (R1–R3):** The positive-framing rewrite named "Obligation held" as an affirmative state. The coupling block's presence adds a named artifact class (the delegation action) that the model checks for. When the delegation action is absent, this absence is now labeled with a positive state name rather than as a negation. R2's Observable ("The delegation action is absent in the trace") suggests the model is checking for the delegation action artifact class specifically, not just checking for companion update presence. This is consistent with the positive-framing hypothesis.

**Source of failure (R4):** Calibration-instance leakage. The Documentation Coordinator example in the glyph is a real contamination risk when the trace contains no delegation action — the model imports the example as trace evidence. This is a glyph authoring issue, not a format issue.

**Assessment:** cas-a at 3/4 is above the v12 predicted floor of 0/4. The positive-framing rewrite is a plausible explanation for R1–R3's improvement. R4's failure mechanism (calibration-instance leakage) is independent and would require glyph-level remediation (removing or isolating the Documentation Coordinator example from the cas-b carve-out scenario). The improvement does not resolve the cas-a glyph problem identified in the PROBE — it reduces it.

---

## Research question assessment

**Primary question:** Does the combined Navigate + Artifact coupling format produce additive improvement — 4/4 on gop-b and 4/4 on cas-b simultaneously?

- **gop-b: 4/4.** Target achieved. Full recovery from 1/4 in v10 and v11b. The coupling block characterizes the downstream-of-consultation state; Navigate provides routing structure. Both blocks operating, coupling block is the active explicit lever.
- **cas-b: 3/4.** Target not fully achieved (one miss from 4/4). The miss is a trace-reading failure in R1 — the model incorrectly evaluates step 5 as not being a delegation action, routing through scope-operative before the coupling discriminator fires. The coupling block itself is not causing the failure.

**Result:** Partial additive improvement. gop-b target achieved; cas-b near-miss due to trace-reading failure, not format failure.

---

## Null result conditions — disposition

- **cas-b regresses again (Navigate overriding coupling block):** Not observed. No Navigate-cited field sources in any cas-b run. Mechanism is not present in v12.
- **gop-b fails to recover:** Not observed. 4/4 achieved.
- **Both b-scenarios degrade (format interference):** Not observed. gop-b improved; cas-b held near baseline.

None of the null result conditions were met.

---

## Findings

1. **gop-b fully recovered (4/4).** The combined format — coupling block characterizing the consultation-prescription relationship, Navigate providing routing structure — eliminates the executor-delta resolution failure that produced 1/4 in v10 and v11b. The coupling block's "Downstream-of-consultation" state is the active lever; Navigate is present but not cited as primary field source.

2. **cas-b near-miss (3/4).** One run fails at trace reading before the coupling discriminator engages. The coupling block is not the failure point; the failure is pre-coupling. For the three correct runs, the coupling block or base scope field is the active lever; Navigate is not cited. The v11a failure mode (Navigate overriding coupling determination) is absent.

3. **cas-a unexpected improvement (3/4, from 0/4).** Positive-framing appears to reduce the cas-a failure rate by naming the delegation action absence as a positive state ("Obligation held") rather than a negation. R4's failure is calibration-instance leakage (Documentation Coordinator hallucinated from glyph example into trace). This is a glyph authoring issue.

4. **Navigate is structurally present but not explicitly cited.** In all b-scenarios, the coupling block or base scope field is the explicit field source for correct verdicts. Navigate may be providing routing structure without being named — or it may be inert. A v13 variant removing Navigate (coupling block only) would isolate Navigate's contribution.

5. **Additivity confirmed for gop; partial for cas.** For gop, the coupling block and Navigate operate in sequence as designed. For cas, the coupling block is the active lever; Navigate's role is unclear. The blocks are not interfering.

---

## Hypotheses for v13

1. **Navigate isolation for gop.** Run gop with coupling block only (no Navigate) to determine whether Navigate is load-bearing or whether the coupling block alone produces 4/4 on gop-b. If gop-b holds at 4/4 without Navigate, Navigate is redundant for gop.

2. **Navigate isolation for cas.** Run cas with coupling block only (no Navigate) to determine whether Navigate's presence in v12 is inert or contributing. If cas-b holds at 3/4+, Navigate is not the active lever and can be removed.

3. **Positive-framing isolation.** A variant identical to v11b except with the positive-framing rewrite applied to the scope discriminator (both branches named as affirmative states, no coupling block added). If cas-b recovers toward 4/4 without the coupling block, positive-framing alone is the active variable for cas-b. This isolates the framing effect from the coupling block effect.

4. **Calibration instance removal for cas-a.** Remove the Documentation Coordinator illustration from the cas glyph. If cas-a reaches 4/4, the leakage was the binding constraint. If cas-a remains below 4/4, the underlying scope characterization problem identified in the PROBE persists.

5. **cas-b trace-reading failure — signal or noise?** R1's failure (misreading step 5 as not a delegation action) is a single run. Run cas-b with n=8 to determine whether this is a consistent failure mode (the model genuinely cannot recognize the registry change request as a delegation action) or a low-frequency error.
