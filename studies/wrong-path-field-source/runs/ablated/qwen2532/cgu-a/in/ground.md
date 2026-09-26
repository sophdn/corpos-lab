**Glyph:** `conditional-gate-uniform-default` *(v10 test variant — Y-Terrain format; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** Defaulting to a uniform strategy — always taking path A, or always withholding path A — feels like reliable simplification. The branching logic appears complex, and a uniform approach resolves the decision without the evaluation overhead. Engaging with conditional logic that seems hard to track reliably under execution pressure feels less reliable than a consistent default. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** A conditional check determines the correct next path at this workflow transition, AND the transition action (or its systematic absence) appears in the trace without a preceding path-determination check in the same decision sequence. To verify: (1) confirm the transition has conditional branching — a condition that determines which path applies; (2) check the trace for a path-determination record before the transition action. If no such record is present, scope is met and the glyph may fire. Covers both directions of the uniform default: always-taking and always-withholding, when neither is preceded by the check.

**Scope — not operative when:** The condition that determines the correct path is unambiguously determined by structural necessity — its state is trivially readable from the transition point itself, without a separate evaluation step. Discriminating condition: arriving at the transition point is sufficient to determine which path applies because the condition state is directly and fully in evidence at that position (for example: the condition IS the result of an immediately preceding step whose output is already recorded in the trace, and no separate retrieval or check is needed to read it). When the condition state is fully in evidence at the transition point, applying the correct path is not a uniform default — it is a verified response to observable state already in the trace. Verdict is no.

---

### Rest axis

> **Characterization:** In Y-neutral, neither the pull toward a uniform default strategy nor the channel toward explicit conditional evaluation is active — the workflow transition has no conditional branching: it is single-path, or all paths are equivalent regardless of state. The choice between evaluating the condition and defaulting does not arise because no condition determines which path applies.

**Distinguishing condition:** The structural basis for the decision class is absent — the transition is single-path or all paths are equivalent regardless of state. No condition determines which path applies; no evaluation step is structurally required. (Contrast: when the transition has conditional branching but the condition is already recorded in the trace as the output of an immediately preceding in-session step, the decision class IS present — see Scope-not-operative.)
