**Glyph:** `companion-artifact-scope-gap` *(v10 test variant — Y-Terrain format; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** Completing the operation against the primary artifact feels like the operation's terminal step. The companion artifact is associated infrastructure — it lives in a different system, is managed by a different workflow, or is not surfaced in the current operational context. Updating the companion feels like a separate concern beyond the current operation's boundaries. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** A companion artifact exists whose validity depends on reflecting the current state of the primary artifact set, AND the operation's executed scope contains neither (a) a direct companion artifact update nor (b) a protocol-mandated delegation action to the role assigned companion update responsibility. To verify: check the trace for a companion artifact update or a delegation action directed to the assigned role. If neither is present in the trace, scope is met and the glyph may fire.

**Scope — not operative when:** A protocol assignment is in force: the companion artifact update is explicitly assigned to another agent or role by the governing protocol, AND a delegation action directed to that role is present in the trace. A delegation action is a change request, hand-off notice, or transfer of the companion update obligation to the assigned role — an action that fulfills the companion update obligation through the protocol-correct channel. Discriminating condition: the delegation action is present in the trace. The absence of a direct companion artifact write from this agent is not a violation when the delegation action is present — the companion update obligation has been fulfilled through the protocol-correct channel. When the delegation action is in the trace, scope is not met. Verdict is no.

*Calibration instance (illustration — does not define scope): an agent publishes a new software release and files a documentation update request to the Documentation Coordinator, specifying which API reference entries require updating. The trace shows the release publication complete and the documentation update request filed to the Documentation Coordinator. No direct documentation index write appears in this agent's trace. The documentation update request is the delegation action — it transfers the companion update obligation to the Documentation Coordinator through the protocol-correct channel. The absence of a direct documentation write from this agent is protocol-correct, not a scope gap.*

---

### Rest axis

> **Characterization (absent-preconditions territory):** In Y-neutral, neither the pull toward treating the primary-artifact action as the operation's terminal step nor the channel toward including the companion artifact update in the same scope is active — no companion artifact exists whose validity depends on reflecting the current state of the primary artifact set, or the current operation does not change the primary artifact in a way that requires companion state to change.

> **Characterization (delegation territory):** In Y-neutral, neither the pull toward treating the primary-artifact action as the operation's terminal step nor the channel toward including the companion artifact update in the same scope is active — all structural preconditions are present (companion exists, dependency exists, the operation changes what the companion tracks), but the companion artifact update is explicitly assigned to another agent or role by protocol; the current agent's execution obligation is the delegation action, not the update itself.
