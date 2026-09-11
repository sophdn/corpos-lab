**Glyph:** `structural-ceiling-bypass` *(v10 test variant — Y-Terrain format; identical to v9 variant; not from ALPHABET.md)*

---

**Y — Decision terrain**

Evaluate in sequence: pull character → scope → verdict. Do not render a verdict before completing the scope check.

**Pull character:** Each entry addition to the artifact is individually warranted — relevant, ready, and the correct completion of the current task. The architectural response the ceiling prescribes (splitting, archiving, rotating, or compressing the artifact) carries overhead beyond the immediate task scope, making it feel expensive and unnecessary relative to the current addition. The pull is toward adding the entry directly. This pull is present in both firing and non-firing cases; its presence alone does not determine the verdict.

**Scope — operative when:** The artifact has an explicitly defined size ceiling with structural consequence — load-time truncation, context-capacity effect, or processing dependency on the ceiling value — AND the pending addition would bring the artifact's entry count beyond that ceiling. To verify: (1) confirm the ceiling is structurally load-bearing (exceeding it changes artifact load state or access behavior, not merely aesthetics or convention); (2) confirm the post-addition entry count would exceed the ceiling. If both hold, scope is met and the glyph may fire.

**Scope — not operative when:** The artifact carries a recommended maximum entry count in a style guide, convention document, or similar source, but the stated ceiling has no structural consequence — no load-time truncation, no context-capacity effect, no processing dependency on the value. Discriminating condition: exceeding the stated maximum produces no change to artifact load state or access behavior. When this is true, the ceiling is advisory-cosmetic; no architectural response is required, and the addition can proceed without violation regardless of how far above the stated maximum the entry count goes. The glyph fires on structurally load-bearing ceilings only. An advisory-only ceiling ends the evaluation: verdict is no.

---

### Marker axis

> **Invariant:** Taking X from Y → the artifact's entry count exceeds its defined ceiling.

**Firing condition:** The artifact's entry count exceeds the value specified in its governing ceiling definition. Observable without intent modeling: count entries in the artifact, compare to the defined ceiling; if the count exceeds the ceiling, the condition fired.

**Does not fire on:**
- Agents adding entries to a structured artifact with an explicitly defined ceiling where the resulting entry count remains at or below the ceiling. Discriminating condition: post-addition entry count does not exceed the defined ceiling — no ceiling is crossed, no architectural response was needed.
- Agents adding entries to structured artifacts that carry a recommended maximum size in a style guide or convention document, but where the ceiling has no structural consequence — no load-time truncation, no context-capacity effect, no processing dependency on the ceiling value. Discriminating condition: the ceiling is advisory-cosmetic rather than structurally load-bearing; exceeding it produces no change to artifact load state or access behavior.

**Violation signal:** The execution trace shows an artifact write or edit that brings the entry count beyond the artifact's defined ceiling, with no preceding architectural response (split, archive, rotate, or compress) in the trace. *Calibration instance (recognition illustration — does not define scope): a documentation agent adds a release entry to a change log index whose rendering system truncates content beyond 50 entries at load time. The ceiling is 50 entries. The agent adds the entry directly, bringing the count to 51, with no archive or rotation preceding it in the trace. The trace-readable violation is the post-write count exceeding 50 without a preceding architectural response — not the content or quality of the entry added.*

---

### Aim axis

> **Invariant:** Taking M from Y → the artifact's entry count is within its defined ceiling.

**Recognition signal:** From position Y — where the pending addition would bring the artifact's entry count beyond the defined ceiling — the agent can verify that the artifact's entry count is within the defined ceiling because the architectural response has been applied and the artifact's count is within the ceiling boundary before the addition proceeds.

---

### Rest axis

> **Characterization:** In Y-neutral, neither the pull toward adding beyond the ceiling nor the channel toward applying the architectural response is active — the artifact has no explicitly defined size ceiling, or its current entry count is far enough below the ceiling that the ceiling constraint does not arise as a live decision factor.
