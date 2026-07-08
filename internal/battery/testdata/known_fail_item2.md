**Glyph:** `test-known-fail`

---

**Y — Decision terrain**

**What this place feels like:** When the agent decides to skip the verification step because the prior output looks correct, the downstream artifact is left unverified.

**What completion looks like:** The agent runs the verification step regardless of confidence.

**Marker axis:** The failure direction is when the agent believes the output is correct and skips verification.

**Aim axis:** The agent has verified the output through the prescribed step.

**Rest axis:** Territory where no verification step exists in the protocol.

**Fallout profile:** See FALLOUT_test-known-fail.md
