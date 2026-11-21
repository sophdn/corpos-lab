**Glyph:** `companion-artifact-scope-gap` *(v12 test variant — Y-Terrain format with Artifact coupling block + Navigate; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → coupling check → verdict. Do not render a verdict before completing the scope and coupling check.

**Pull character:** Completing the operation against the primary artifact feels like the operation's terminal step. The companion artifact is associated infrastructure — it lives in a different system, is managed by a different workflow, or is not surfaced in the current operational context. Updating the companion feels like a separate concern beyond the current operation's boundaries. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** A companion artifact exists whose validity depends on reflecting the current state of the primary artifact set, AND the operation's executed scope contains neither (a) a direct companion artifact update nor (b) a protocol-mandated delegation action to the role assigned companion update responsibility. To verify: check the trace for a companion artifact update or a delegation action directed to the assigned role. If neither is present in the trace, scope is met and the glyph may fire.

**Scope — not operative when:** A protocol assignment is in force: the companion artifact update is explicitly assigned to another agent or role by the governing protocol, AND a delegation action directed to that role is present in the trace. A delegation action is a change request, hand-off notice, or transfer of the companion update obligation to the assigned role — an action that fulfills the companion update obligation through the protocol-correct channel. Discriminating condition: the delegation action is present in the trace. The absence of a direct companion artifact write from this agent is not a violation when the delegation action is present — the companion update obligation has been fulfilled through the protocol-correct channel. When the delegation action is in the trace, scope is not met. Verdict is no.

*Calibration instance (illustration — does not define scope): an agent publishes a new software release and files a documentation update request to the Documentation Coordinator, specifying which API reference entries require updating. The trace shows the release publication complete and the documentation update request filed to the Documentation Coordinator. No direct documentation index write appears in this agent's trace. The documentation update request is the delegation action — it transfers the companion update obligation to the Documentation Coordinator through the protocol-correct channel. The absence of a direct documentation write from this agent is protocol-correct, not a scope gap.*

**Artifact coupling:** This glyph involves an exterior coupling point — an artifact boundary between the executing agent and the companion-update-assigned role.

Boundary type: executor → companion-update-assigned role (agent→agent)

Coupling artifact: the delegation action — a change request, hand-off notice, or transfer of the companion update obligation to the assigned role. When the delegation action is filed, this artifact is produced (+1) at the agent→agent boundary. It is a real artifact, not an absence.

Scope discriminator — two states:

- **Delegation complete:** A delegation action directed to the companion-update-assigned role is present in the trace — the companion update obligation has been transferred through the protocol-correct channel; the executor's artifact delta on the companion artifact is 0. The executor's part of the companion update protocol is complete. Scope not operative. Verdict: no.
- **Obligation held:** No delegation action directed to the assigned role appears in the trace — no coupling event has occurred; the companion update obligation has not been transferred; the executor still holds it. Scope operative if the primary-artifact operation is present and no direct companion update appears. Verdict: yes.

**Navigate:** After completing the scope check and coupling review, resolve to verdict:

- Delegation complete (delegation action present in trace, companion update obligation transferred) → executor's part of the companion update protocol is complete; the absence of a direct companion write from this agent is protocol-correct. Verdict: no.
- Obligation held (no delegation action in trace) → companion update obligation not transferred; scope operative. Verdict: yes.

---

### Marker axis

> **Invariant:** Taking X from Y → the companion artifact's state does not reflect the current state of the primary artifact set — the companion artifact update was not within the operation's executed scope.

**Firing condition:** The primary-artifact operation is present in the trace and no companion artifact update appears in the same operation's execution scope. Observable: check the trace for a companion artifact update co-present with the primary-artifact action; if absent, the condition fired. No information-substitution mechanism: the violation mechanism is scope-absence (the companion artifact is not within the agent's immediate operational field), not a case where the agent acts on held beliefs about the companion's current state.

**Does not fire on:**
- Operations on a primary artifact type for which no companion artifact is structurally required. Discriminating condition: no companion artifact exists whose validity depends on reflecting the state of this primary artifact set — the structural dependency relationship is absent for this operation type.
- Operations that modify a primary artifact without changing the state that the companion artifact tracks. Discriminating condition: the companion artifact's required content is unchanged by this primary-artifact operation; no companion update is structurally mandated even though a companion artifact exists.
- Operations by an agent whose role excludes the companion artifact update by protocol assignment. Discriminating condition: the companion artifact update is explicitly assigned to another agent or role by protocol, and this agent's execution obligation is the delegation action; the absence of a companion update in this agent's trace is correct by protocol, not a scope-absence error.

**Violation signal:** The trace shows a primary-artifact operation completed without a companion artifact update in the same execution scope. The primary-artifact action appears; the companion artifact update does not. *Calibration instance (recognition illustration — does not define scope): a software agent updates a service module's public interface — a new exported function is added and changes are committed. A companion artifact exists: an API documentation index that must reflect all exported interfaces in the current module set. The agent completes the module update without updating the API documentation index. The trace shows the module modification complete; the API index update is absent from the execution scope. The violation is the absent companion update, not the downstream state of the documentation.*

---

### Aim axis

> **Invariant:** Taking M from Y → the companion artifact's state reflects the current state of the primary artifact set — the primary artifact operation and the companion artifact update are co-present in the same execution scope.

**Recognition signal:** From position Y — where completing the primary-artifact action feels like the terminal step — the agent can verify correct navigation by confirming that a companion artifact update appears within the same operation's execution scope as the primary-artifact action. The companion update is either present or absent in the current execution right now. No projection required.

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward treating the primary-artifact action as the operation's terminal step nor the channel toward including the companion artifact update in the same scope is active — no companion artifact exists whose validity depends on reflecting the current state of the primary artifact set, or the current operation does not change the primary artifact in a way that requires companion state to change.

> **Characterization (delegation territory):** In Y-neutral, neither the pull toward treating the primary-artifact action as the operation's terminal step nor the channel toward including the companion artifact update in the same scope is active — all structural preconditions are present (companion exists, dependency exists, the operation changes what the companion tracks), but the companion artifact update is explicitly assigned to another agent or role by protocol; the current agent's execution obligation is the delegation action, not the update itself.
