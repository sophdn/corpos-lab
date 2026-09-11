# RESPONSE_fsb-a_1

**Study:** assay-blank-claude-definition-work-v10
**Scenario:** fsb-a (fires)
**Ground truth:** yes
**Run:** 1 of 4
**Model:** mistral:latest
**Date:** 2026-04-01

---

 VERDICT: yes
FIELD SOURCE: Scope — operative when
OBSERVABLE: "Sub-step records must be present as entries produced by in-session execution within this procedure instance, not as context loaded from a prior session or frame. If no such records appear in the trace, scope is met and the glyph may fire."
EVIDENCE: "No transaction-completeness check record and no period-lock verification record appear in this procedure instance."