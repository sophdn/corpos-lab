**Glyph:** `formal-step-context-bypass` *(v11b test variant — Y-Terrain format with artifact coupling block; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** The agent's currently-loaded session context already contains material that matches what the prerequisite sub-steps would produce. Running the sub-steps would only reproduce what is already present — the step feels redundant when the same information is already in context. The pull is toward treating the loaded context as sufficient and proceeding to the substantive work. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** The required prerequisite sub-step records are absent from the current procedure instance trace — no entry exists for the required sub-steps having been run and recorded in this procedure instance. To verify: scan the full trace for sub-step records before the analysis phase begins. Sub-step records must be present as entries produced by in-session execution within this procedure instance, not as context loaded from a prior session or frame. If no such records appear in the trace, scope is met and the glyph may fire.

**Scope — not operative when:** Sub-step records ARE present in the current session trace — entries showing the required sub-steps were run within this procedure instance, their results produced in this session, their records committed to this trace. To verify: look for the sub-step records in the trace. They will appear as explicit trace entries — completed actions with results recorded — not merely as loaded context. Discriminating condition: the sub-step records are present in this trace. When sub-step records are present in the trace, scope is not met and the glyph does not fire, regardless of how many steps separate the sub-steps from the analysis phase. This glyph fires on the absence of sub-step records — a present record ends the evaluation. Verdict is no.

Note: the pull character describes the anachronicity mechanism — loaded context from a prior frame cannot substitute for the procedural record. That mechanism does not apply when the sub-step records are on record in this trace, because the context is not from a prior frame but from in-session execution.

**Artifact coupling:** This glyph involves an intra-session temporal coupling point — an artifact boundary between earlier execution within this procedure instance and the current step.

Boundary type: earlier in-session execution → current step (procedural-instance-internal temporal coupling)

Coupling artifact: the prerequisite sub-step execution records — trace entries produced by running the required sub-steps within this procedure instance. These records are artifacts that exist inside the session boundary. They are readable in the current trace as explicit completed-action entries.

The session boundary is the discriminating line: artifacts produced inside this procedure instance (in-session coupling artifacts) are valid. Context loaded from outside the session boundary — from a prior session, a prior pipeline run, or a prior frame — is not a coupling artifact for this procedure instance, regardless of how similar its content is to what the sub-steps would produce. The anachronicity mechanism describes exactly this: prior-frame context belongs to a different operational frame.

Executor delta at the coupling boundary:
- Sub-step execution records present as trace entries within this session → coupling artifacts exist inside the session boundary (+1 per sub-step record). The sub-steps were run in this procedure instance; the coupling event occurred within the session boundary.
- Sub-step records absent from this session trace → no in-session coupling event. Loaded context from prior sessions does not count as a coupling artifact for this procedure instance.

Scope discriminator: Are sub-step execution records present as trace entries within this session?
- Yes → in-session coupling artifacts present. The sub-steps were run in this procedure instance. Scope not operative. Verdict: no.
- No → no in-session coupling event. Loaded prior-session context is not a substitute. Scope operative if substantive work is in progress. Verdict: yes.

---

### Marker axis

> **Invariant:** Taking X from Y → the required prerequisite sub-step is absent from the procedural record — no entry exists for it having been opened from its defined entry conditions in this procedure instance.

**Firing condition:** The execution trace shows the substantive work of a formally-sequenced step in progress — tool calls, artifact writes, or document state produced — with no preceding record entries for the required prerequisite sub-steps in the current procedure instance. Observable without intent modeling: check the trace for sub-step entry records; if absent, the condition fired. The information-substitution mechanism is anachronicity: the loaded context belongs to a different operational frame from the required procedural sub-step record and cannot substitute for it regardless of accuracy, recency, or source.

**Does not fire on:**
- Sessions where the required prerequisite sub-steps were executed earlier in the current procedure instance and a record of their execution exists in the trace. Discriminating condition: sub-step records are present in the session trace, not absent — the context containing sub-step output traces to in-session execution, not to a prior-session load.
- Procedure instances where a formal resume point document establishes that the required prerequisite sub-steps were completed in a prior valid execution, and the current session explicitly opens at that resume point. Discriminating condition: prior sub-step completion is on record, not inferred from available context.

**Violation signal:** The execution trace shows a formally-sequenced substantive step in progress — tool calls, artifact writes, or document state produced — without a preceding entry for the required prerequisite sub-steps in this procedure instance. *Calibration instance (recognition illustration — does not define scope): a data-processing pipeline agent is about to begin the analysis phase of a formally-sequenced procedure that prescribes prerequisite sub-steps: a data-freshness check and a schema-validation record. The agent's session context contains results from these checks run in a prior pipeline execution. It proceeds to analysis without running the current-session sub-steps, reasoning that the prior results establish the same state. The trace shows the analysis phase active — data reads, transformation calls, output writes — without a preceding data-freshness check or schema-validation record in the current procedure instance. The violation is the absent sub-step records, not the accuracy of the prior results.*

---

### Aim axis

> **Invariant:** Taking M from Y → a record entry exists for each required prerequisite sub-step, opened from its defined entry conditions in the current procedure instance.

**Recognition signal:** At the decision point, the agent can verify that the required prerequisite sub-steps are on record as having been opened from their defined entry conditions in this procedure instance — not merely that the expected information is available in session context.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward treating currently-loaded context as a sufficient substitute for sub-step execution nor the channel toward entering prerequisite sub-steps from their defined entry conditions is active — the current task has no formally-sequenced procedure prescribing prerequisite sub-steps with defined entry conditions. The choice between bypass and prescribed entry does not arise.

> **Characterization (present-preconditions-delegated-elsewhere territory):** In Y-neutral, neither the pull toward treating currently-loaded context as a sufficient substitute for sub-step execution nor the channel toward entering prerequisite sub-steps from their defined entry conditions is active — a formally-sequenced procedure is in play and prescribes prerequisite sub-steps with defined entry conditions, but executing those sub-steps is explicitly assigned to another agent or role by protocol. The current agent's scope excludes running them. The choice between bypass and prescribed entry does not arise as live for this agent — not because no procedure is in play, but because the sub-steps are not theirs to run.
