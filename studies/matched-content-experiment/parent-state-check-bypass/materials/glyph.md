# parent-state-check-bypass

**Y-fire:** An agent is about to begin executing a specific work item that belongs to a parent context. The parent context carries a required pre-execution state check — a check that must confirm the parent context's current state before work on any of its items begins in a new session. The check has not been run in the current session. Running it now feels like administrative overhead disconnected from the specific work item: the work item's own immediate requirements don't appear to depend on the parent context's state, and the check seems like housekeeping that could be skipped without affecting the task itself. The pull is toward proceeding directly with the work item. The mechanism is gate-absence: the check is not bypassed because the agent believes they already hold the parent context's state information — it is bypassed because the check is not perceived as a required gate for executing this specific work item. The live choice is whether to run the parent state check before beginning task work, or proceed without it.

**Y-not-fire:** The agent is at the same position — about to begin executing a work item belonging to a parent context that carries a required pre-execution state check — but the parent state check was run and documented as current earlier in the same session. The discriminating condition is present: a parent state check record exists in the session trace before this task pickup. The glyph does not fire — the session-level check requirement is satisfied; the agent's current task context rests on a verified parent-context state even though the check was not performed immediately before this specific work item.

---

### Marker axis

> **Invariant:** Taking X from Y → the pre-execution parent state check record is absent from the session trace at the point where task-specific work begins — the parent context's state at execution start is unverified, not established as current in this session.

**Firing condition:** The execution trace shows task-specific work beginning — work-item actions committed — without a preceding parent state check record in the same session. Observable without intent modeling: check the trace for a parent state check record before any work-item-specific actions in the current session; if absent, the condition fired. No information-substitution mechanism — the violation mechanism is gate-absence: the check is not perceived as a required gate for the specific work item, not bypassed because the agent believes they already hold alternative information.

**Does not fire on:**
- Work items that have no parent context with a required pre-execution state check. Discriminating condition: the parent context either does not exist for this work item or carries no state check requirement for this class of work — the structural basis for the decision class is absent.

**Violation signal:** The trace shows task-specific work in progress — work-item actions committed — without a parent state check record preceding them in the current session. *Calibration instance (recognition illustration — does not define scope): a work item belonging to a parent context is picked up — work-item-specific actions appear in the trace — without a preceding parent state check in the same session. The parent context's state at pickup is unverified: stale item statuses or a changed parent context state may not have been caught before work began. The trace-readable violation is the absence of the parent state check record in the session trace, not the downstream state of post-work closure.*

---

### Aim axis

> **Invariant:** Taking M from Y → the agent is holding the result of the parent state check; the parent context's current state is verified and available as working context before task-specific work begins — the agent proceeds from a known foundation, not an assumed one.

**Recognition signal:** From position Y — where the parent state check feels like administrative overhead unrelated to the specific work item — correct navigation is recognizable as the agent's holding of a verified parent-context state: the check result is present in the decision context, and task-specific work begins from that verified position.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — the work item in question has no parent context with a required pre-execution state check, or the parent context exists but imposes no check requirement for this class of work. The choice between running and skipping the check does not arise.
>
> **Characterization (present-preconditions-delegated-elsewhere territory):** In Y-neutral, neither the pull toward skipping the parent state check nor the channel toward running it is active — a parent context exists and carries a required pre-execution state check for this class of work, but running the check is explicitly assigned to another agent or role by protocol. The current agent's scope excludes executing it. The choice between skipping and running the check does not arise as live for this agent — not because no parent context or check requirement exists, but because running the check is protocol-assigned elsewhere.

**Distinguishing condition (absent-preconditions):** The structural basis for the decision class is absent — either the work item has no parent context, or the parent context carries no required pre-execution state check for this type of work. The agent proceeds with task work without applying or bypassing any check, because no check is structurally required.

**Distinguishing condition (present-preconditions-delegated-elsewhere):** All structural preconditions are present — a parent context exists with a required pre-execution state check for this class of work — but check execution is protocol-assigned to another agent or role. The current agent's scope excludes it by protocol assignment, not by absence of the structural dependency.
