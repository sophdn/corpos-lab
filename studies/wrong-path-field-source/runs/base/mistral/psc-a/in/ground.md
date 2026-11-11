**Glyph:** `parent-state-check-bypass` *(v10 test variant — Y-Terrain format; identical to v9/v8 variant; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** The work item's immediate requirements do not appear to depend on the parent context's current state. The parent state check feels like administrative overhead disconnected from the specific task — housekeeping that could be skipped without affecting this work item's outcome. The pull is toward the work item itself. The gate's connection to this specific task is not immediately visible from position Y. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** The parent state check has not been run in the current session. To verify: scan the full session trace for a parent state check record before any task-specific work begins. If no such record is present anywhere in the trace before task-specific actions, the scope condition is met and the glyph may fire. The session boundary is the unit of analysis — not the moment immediately before a specific task pickup.

**Scope — not operative when:** A parent state check record exists in the current session trace — the check was run in this session, its result is documented, and it precedes task-specific work. When this record is present, the session gate is already closed and the glyph does not fire, regardless of how many steps separate the check from the specific task pickup. This glyph fires on absence-of-session-check-record. A present record ends the evaluation: verdict is no.

---

### Marker axis

> **Invariant:** Taking X from Y → the pre-execution parent state check record is absent from the session trace at the point where task-specific work begins — the parent context's state at execution start is unverified, not established as current in this session.

**Firing condition:** The execution trace shows task-specific work beginning — work-item actions committed — without a preceding parent state check record in the same session. Observable without intent modeling: check the trace for a parent state check record before any work-item-specific actions in the current session; if absent, the condition fired. No information-substitution mechanism — the violation mechanism is gate-absence: the check is not perceived as a required gate for the specific work item, not bypassed because the agent believes they already hold alternative information.

**Does not fire on:**
- Work items that have no parent context with a required pre-execution state check. Discriminating condition: the parent context either does not exist for this work item or carries no state check requirement for this class of work — the structural basis for the decision class is absent.
- Cases where the parent state check was run and documented as current earlier in the same session. Discriminating condition: a parent state check record exists in the session trace before task-specific work begins — the session-level check has been satisfied, even if it was not performed immediately before this specific task pickup. The glyph fires on absence-of-session-check-record, not absence-of-pickup-moment-check.

**Violation signal:** The trace shows task-specific work in progress — work-item actions committed — without a parent state check record preceding them in the current session. *Calibration instance (recognition illustration — does not define scope): a software agent picks up a specific bug fix from a project's active sprint. The sprint is the parent context; a sprint-state check is required before work on any sprint item begins in a new session. The agent proceeds directly to the bug fix — diagnostic calls, code reads, patch writes appear in the trace — without a preceding sprint-state check in the current session. The sprint's current state (active, paused, descoped, priority-shifted) is unverified at the point task work begins. The trace-readable violation is the absence of the sprint-state check record, not the quality of the bug fix produced.*

---

### Aim axis

> **Invariant:** Taking M from Y → the parent state check has been run in this session and its result is on record before task-specific work begins — the parent context's current state is verified, not assumed.

**Recognition signal:** From position Y — where the parent state check feels like administrative overhead unrelated to the specific work item — the agent can verify correct navigation by confirming a parent state check record precedes task-specific work in the session trace. The parent context's state was verified this session; task work proceeds on a known foundation, not an assumed one.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — the work item in question has no parent context with a required pre-execution state check, or the parent context exists but imposes no check requirement for this class of work. The choice between running and skipping the check does not arise.

> **Characterization (present-preconditions-delegated-elsewhere territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — a parent context exists and carries a required pre-execution state check for this class of work, but running the check is explicitly assigned to another agent or role by protocol. The current agent's scope excludes executing it. The choice between skipping and running the check does not arise as live for this agent — not because no parent context or check requirement exists, but because running the check is protocol-assigned elsewhere.
