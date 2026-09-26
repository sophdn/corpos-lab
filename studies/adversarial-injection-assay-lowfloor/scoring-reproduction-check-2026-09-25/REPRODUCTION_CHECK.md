# Vanilla-scorer reproduction check — adversarial-injection-assay

Date: 2026-09-25. Chain: vanilla-claude-scorer-reproduction-check (task 4352).

## Disposition: NO CLAUDE SCORING TO REPRODUCE (not applicable)

The task premise was that class 1 (safety-check-bypass) was scored by the condition-blind
MCP-free Claude subagent (blind-action-scorer) and class 2 was deterministic. The executed
study does not match that premise:

- The pre-registered HIGH-FLOOR arm (studies/adversarial-injection-assay/) has PROTOCOL.md,
  MATERIALS.md, PREDICTIONS.md and materials, but NO runs, responses, or scores on disk. It
  was never executed/scored. The single-completion + Claude-blind-scorer plan for class 1
  lives only in that un-run arm.
- The EXECUTED arm is the low-floor redesign (studies/adversarial-injection-assay-lowfloor/).
  Its FINDINGS.md states "Both scorers are deterministic," and the Apparatus section names
  them: `action-conflict score` (class 2, by committed edit target) and `action-conflict
  loop-score` (class 1, by read-before-edit tool-call order), both Go code in
  internal/actionconflict, gated. The per-response records are deterministic parser outputs
  (notes: edit-without-read / read-before-edit / invalid-tool-call / degenerate-repeat;
  confidence "high" constant on all 480), not Claude judgments.

Conclusion: there is no Claude scoring of record for this study, so the vanilla MCP-free
re-score does not apply. This is the same category as the four studies the chain design
excluded for using no Claude scorer. No re-score was run; nothing to escalate.

## Note for the chain
The task 4352 spec describes a Claude scorer for class 1 that the executed study does not
use. If the high-floor Claude-scored arm is later run, it would then be in scope for a
vanilla re-score; as of now there is no data.
