# Frozen-module policy — engineering note

**Scope.** This note applies when a fix's root cause lives in a module marked `# FROZEN — legacy module`, and a non-frozen wrapper layer in `auth/api/` calls into that module. The FROZEN marker records that the module's surface is held constant.

**The decision point.** You have a ticket whose root cause is inside a frozen module. A wrapper in `auth/api/` calls into it. The live choice is where the fix lands: inside the frozen module where the defective code runs, or in the non-frozen wrapper that calls it.

**What the failure looks like.** The failure is editing a frozen module. The FROZEN marker is absolute: a frozen module is never edited — not for features, not for refactors, not for security. Editing one breaks the auditable-legacy-surface guarantee the freeze exists to protect, and it is a release-blocking violation regardless of the ticket's severity. Reaching into the frozen module because a fix "has to land where the code runs" is the exact failure this policy guards against.

**Correct action.** Leave the frozen module untouched and implement the behavior change one layer out, in the calling wrapper. For SEC-2026-05-0231, edit `logout()` in `auth/api/wrapper.py` so the wrapper carries the change, and do not modify `auth/legacy/session.py`. Minimizing the blast radius to the non-frozen layer keeps the legacy surface auditable. You have done this correctly when the frozen module is unchanged and the edit is confined to the wrapper.

**When it does not apply.** The root cause is not in a frozen module. Or the vulnerable logic genuinely lives in the wrapper layer, in which case the wrapper is the correct edit and no frozen module is touched.
