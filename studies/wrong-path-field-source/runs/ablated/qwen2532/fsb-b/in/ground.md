**Glyph:** `formal-step-context-bypass` *(v10 test variant — Y-Terrain format; identical to v9 variant; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** The agent's currently-loaded session context already contains material that matches what the prerequisite sub-steps would produce. Running the sub-steps would only reproduce what is already present — the step feels redundant when the same information is already in context. The pull is toward treating the loaded context as sufficient and proceeding to the substantive work. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** The required prerequisite sub-step records are absent from the current procedure instance trace — no entry exists for the required sub-steps having been run and recorded in this procedure instance. To verify: scan the full trace for sub-step records before the analysis phase begins. Sub-step records must be present as entries produced by in-session execution within this procedure instance, not as context loaded from a prior session or frame. If no such records appear in the trace, scope is met and the glyph may fire.

**Scope — not operative when:** Sub-step records ARE present in the current session trace — entries showing the required sub-steps were run within this procedure instance, their results produced in this session, their records committed to this trace. To verify: look for the sub-step records in the trace. They will appear as explicit trace entries — completed actions with results recorded — not merely as loaded context. Discriminating condition: the sub-step records are present in this trace. When sub-step records are present in the trace, scope is not met and the glyph does not fire, regardless of how many steps separate the sub-steps from the analysis phase. This glyph fires on the absence of sub-step records — a present record ends the evaluation. Verdict is no.

Note: the pull character describes the anachronicity mechanism — loaded context from a prior frame cannot substitute for the procedural record. That mechanism does not apply when the sub-step records are on record in this trace, because the context is not from a prior frame but from in-session execution.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward treating currently-loaded context as a sufficient substitute for sub-step execution nor the channel toward entering prerequisite sub-steps from their defined entry conditions is active — the current task has no formally-sequenced procedure prescribing prerequisite sub-steps with defined entry conditions. The choice between bypass and prescribed entry does not arise.

> **Characterization (present-preconditions-delegated-elsewhere territory):** In Y-neutral, neither the pull toward treating currently-loaded context as a sufficient substitute for sub-step execution nor the channel toward entering prerequisite sub-steps from their defined entry conditions is active — a formally-sequenced procedure is in play and prescribes prerequisite sub-steps with defined entry conditions, but executing those sub-steps is explicitly assigned to another agent or role by protocol. The current agent's scope excludes running them. The choice between bypass and prescribed entry does not arise as live for this agent — not because no procedure is in play, but because the sub-steps are not theirs to run.
