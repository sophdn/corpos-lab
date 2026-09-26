**Glyph:** `parent-state-check-bypass` *(v10 test variant — Y-Terrain format; identical to v9/v8 variant; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** The work item's immediate requirements do not appear to depend on the parent context's current state. The parent state check feels like administrative overhead disconnected from the specific task — housekeeping that could be skipped without affecting this work item's outcome. The pull is toward the work item itself. The gate's connection to this specific task is not immediately visible from position Y. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** The parent state check has not been run in the current session. To verify: scan the full session trace for a parent state check record before any task-specific work begins. If no such record is present anywhere in the trace before task-specific actions, the scope condition is met and the glyph may fire. The session boundary is the unit of analysis — not the moment immediately before a specific task pickup.

**Scope — not operative when:** A parent state check record exists in the current session trace — the check was run in this session, its result is documented, and it precedes task-specific work. When this record is present, the session gate is already closed and the glyph does not fire, regardless of how many steps separate the check from the specific task pickup. This glyph fires on absence-of-session-check-record. A present record ends the evaluation: verdict is no.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — the work item in question has no parent context with a required pre-execution state check, or the parent context exists but imposes no check requirement for this class of work. The choice between running and skipping the check does not arise.

> **Characterization (present-preconditions-delegated-elsewhere territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — a parent context exists and carries a required pre-execution state check for this class of work, but running the check is explicitly assigned to another agent or role by protocol. The current agent's scope excludes executing it. The choice between skipping and running the check does not arise as live for this agent — not because no parent context or check requirement exists, but because running the check is protocol-assigned elsewhere.
