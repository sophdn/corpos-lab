# Content-parity audit — duty versus corpus

The behavioral-equivalence contrast (Condition A duty_only versus Condition B
corpus_only) is only clean if the duty and the corpus carry the **same
behavioral information** and differ only in **delivery form**. This is the
single-turn analogue of the original design's input-parity requirement (Gate 3.0
of the source study). It parallels the matched-content experiment's content
matching (`PROTOCOL.md` section 2), adapted: there the axis was glyph-format
versus imperative; here it is duty-specification versus descriptive-corpus.

Auditor: Claude (Opus 4.8), the session that authored both materials. A future
independent second pass would strengthen this; it is noted as a limitation of
the audit that author and auditor are the same session.

## The two patterns and their propositions

Both materials must carry these propositions, at the same specification level.

### `commitment-precedes-reads`

| # | Proposition | In duty.md | In corpus.md |
|---|-------------|-----------|--------------|
| 1 | The commitment is the round-level hypothesis or sprint question | yes ("the round-level commitment (hypothesis or sprint question)") | yes ("the round-level commitment (the hypothesis or the sprint question)") |
| 2 | The commitment must be authored before the round's evidence is read | yes | yes |
| 3 | A commitment authored blind to its evidence functions as a falsifiability anchor | yes ("falsifiability anchor") | yes ("functions as a falsifiability anchor") |
| 4 | A commitment authored after reading the evidence is post-hoc rationalization | yes ("post-hoc rationalization") | yes ("post-hoc rationalization rather than prediction") |

### `investigation-early-confirmation-stop`

| # | Proposition | In duty.md | In corpus.md |
|---|-------------|-----------|--------------|
| 5 | All designated hypothesis sprints in a round run to completion | yes | yes |
| 6 | The agent does not stop after one sprint confirms a hypothesis | yes | yes |
| 7 | A root cause recorded without eliminating alternatives lacks falsifiability coverage | yes ("lacks falsifiability coverage") | yes ("lacks falsifiability coverage") |
| 8 | A confirmed hypothesis earns its status by surviving alternatives | yes (Outcome) | yes ("earn its status by surviving them") |

All eight propositions appear in both. No proposition appears in one and not the
other.

## What is NOT matched (the experimental variable)

- **Form.** The duty is a prescriptive office: a Goal, named Taboos as "the
  agent must not …", and an Outcome. The corpus is a descriptive registry: each
  pattern named and described as terrain, with a stated failure case and sound
  case. This is the variable under study; it is not a confound.
- **Framing verbs.** The duty commands ("must not read", "must not stop"). The
  corpus describes ("a commitment authored before … is blind to what will test
  it"). Same propositions, different mood.

## What could still confound

- **Length.** duty.md and corpus.md are close in length (the corpus is slightly
  longer because a descriptive entry restates the terrain the duty states once).
  Neither is padded. The word-count ratio is well under 2:1.
- **Vocabulary.** Both use the two pattern slugs verbatim
  (`commitment-precedes-reads`, `investigation-early-confirmation-stop`). Neither
  gains a vocabulary the other lacks.

## Verdict

The duty and the corpus are content-matched on all eight propositions at a
matched specification level. The A/B contrast measures delivery form, not
information content.
