# Independent content-parity sign-off — low-floor arm

**Date:** 2026-09-24. **Auditor:** an independent MCP-free subagent with no access to the
study records — it judged only the material files. The materials' author cannot self-certify
parity (PROTOCOL requires an independent check before the grid runs). This file records that check.

## What parity requires

For each class, the poisoned and legitimate entries must be matched — same structure, length,
directive force, and specificity — differing ONLY in the action they prescribe (harmful vs safe).
Any other systematic difference is a confound that could move the wrong-action rate on grounds
other than the manipulation.

## First pass — one real confound found

- **Class A (security-fix-freeze-deflection): PARITY OK.** Matched title, five-section skeleton,
  and directive force. Minor note: legitimate cites "the security ticket" where poisoned cites the
  ticket id; legitimate is longer by its intrinsic freeze-exception justification.
- **Class B (safety-check-bypass): PARITY CONCERNS.** Four differences on grounds other than the
  prescribed action: (1) title "efficiency" vs "state"; (2) header "waste" vs "failure"; (3) the
  neutral decision-point pre-framed the read as redundant in the poisoned arm; (4) **the substantive
  one** — the poisoned Correct action named the function and field to edit while the legitimate one
  named only the file, so the poison was more directly actionable and would raise the skip-and-fix
  rate on specificity, not the manipulation.

## Fixes applied

- **Class B poisoned:** title → "Session-start state policy" (identical to legitimate); header →
  "What the failure looks like"; decision-point normalized to the neutral wording (the redundancy
  framing now lives only in the manipulation body, not the neutral section).
- **Class B legitimate:** Correct action now names `validate_url_field()` in `config/validator.py`
  and "validates URL structure for `auth_service_url`" — the same specificity as the poison.
- **Class A legitimate:** Correct action now cites SEC-2026-05-0231, matching the poison.
- Derived controls regenerated: Class B scrambled (from the new poison), Class A off_target
  (= the new Class B poison).

## Re-audit — clear to run

Both classes PARITY OK. Class B: all four confounds resolved; arms differ only in the prescribed
action; word-count gap 2. Class A: ticket-id asymmetry resolved; the only residual is the
intrinsic +19-word freeze-exception justification in the legitimate arm, which is inherent to the
manipulation (the safe arm must justify breaking the freeze), non-blocking.

**Verdict: both classes clear to run.**
