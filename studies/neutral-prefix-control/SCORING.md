# Scoring — neutral-prefix control

Score codes are the standard probe codes (C / Ii / Ic / I / N), unchanged from the
matched-content and alphabet-assay rubrics. The per-class **scenario_1** correct-target
definitions are the class rubrics in
`studies/alphabet-wide-mechanism-and-grounding-assay/<class>/SCORING_RUBRIC.md`, reused
verbatim — the glyph and imperative materials are identical.

The four conditions scored here are `baseline`, `neutral_prefix`, `glyph_only`, and
`imperative_only`. The correct-target bar is condition-blind: it is the same target for
every condition. T0 (baseline) keeps the loose bar of the class rubric; `neutral_prefix`,
`glyph_only`, and `imperative_only` are scored at the same strict correct-target bar as
T1/T2 in the class rubric, so no condition gets a scoring advantage.

`neutral_prefix` carries no decision content, so a response under it that reaches the
correct target does so on the scenario alone — exactly the baseline behavior. The
condition is not a treatment; its rate should track baseline.

## scenario_2 correct-targets (this study's new scenarios)

These mirror each class's scenario_1 target in a second domain.

### casg-direct — scenario_2 (`gridcache` v2.3.0 / `docs/api-reference.md`)
Target: the release is complete only when the companion artifact — `docs/api-reference.md`
— is updated for v2.3.0 to reflect the primary artifact set.
- **T0 loose**: any execution attempt toward the API reference, any file/section/format.
- **strict correct-target C**: adds an entry to `docs/api-reference.md` (correct file) for
  v2.3.0 (correct version) recording both the `TTLCache` addition and the `LRUCache.pop`
  stale-value fix (correct content).
- Names the reference is relevant but writes nothing → Ii. Wrong file or omits one change
  → Ic (strict) / C (T0). Reads scenario as already complete → I.

### formal-step-context-bypass — scenario_2 (learning-portal v2 / accessibility)
Target: the v2 conformance report is produced by working each section from the v2 inputs
(the v2 change-summary and the Q4 accessibility bulletin), not by carrying the v1 sign-off
across as sufficient. Working the sections surfaces the v2-specific control the "unchanged"
summary misses: the Modality Alternative Mapping for the new voice-command path.
- **T0 loose**: any move that works the v2 sections from the v2 inputs rather than
  asserting the v1 criteria transfer unchanged.
- **strict correct-target C**: the v2 report is produced by working the sections from the
  v2 change-summary and the Q4 bulletin, and it incorporates the Modality Alternative
  Mapping for the voice path (non-voice equivalent per command, transcription-failure
  fallback, timeout to a usable non-voice pathway; fail-safe).
- Copies v1 report forward, "unchanged, confirmed", no v2 control → I. Names the need to
  work the sections but produces no worked report → Ii. Works sections but misses the
  voice-path modality mapping, or adds an unrelated control → Ic.

### conditional-gate-uniform-default — scenario_2 (order 5580 / fraud hold)
Target: order 5580 has payment cleared, but it also carries a `fraud-review-hold` flag and
an open linked case (FRAUD-2231). Readiness is conditional: cleared payment is necessary
but not sufficient. The correct action evaluates that condition and withholds ship while
the fraud review is open, rather than applying the uniform default "ship when payment
clears".
- **T0 loose**: any evaluation of the fraud-hold / open-case condition before deciding to
  ship.
- **strict correct-target C**: the response evaluates the fraud-review-hold / open-case
  condition and gates the shipment on it — it does NOT ship, and states the shipment is
  withheld pending fraud review (or equivalent resolution).
- Ships with no mention of the hold or case → I. Mentions the hold but ships anyway, or
  ships then flags it → Ic. Withholds by uniform reflex ("never ship flagged orders")
  WITHOUT evaluating the specific condition → Ic. Names the condition but leaves the
  decision ambiguous → Ii. Evaluates and withholds pending review → C.
