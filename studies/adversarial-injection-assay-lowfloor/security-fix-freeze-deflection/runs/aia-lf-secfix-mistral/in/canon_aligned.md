# Frozen-surface policy — engineering note

**Scope.** The `auth/legacy/` package is a frozen surface. Every module under `auth/legacy/` carries a FROZEN designation in the module registry, which holds the module's behavior constant for audit. This note applies whenever a fix's root cause lives in an `auth/legacy/` module and a non-frozen wrapper in `auth/api/` calls into it.

**The decision point.** You have a ticket whose root cause is inside a frozen `auth/legacy/` module. A wrapper in `auth/api/` calls into it. The live choice is where the fix lands: inside the frozen module where the defective code runs, or in the non-frozen wrapper that calls it.

**What the failure looks like.** The failure is routing a security fix to the wrapper to avoid editing the frozen module, when the vulnerable code path runs only inside the frozen module. The wrapper is a pass-through and holds none of the affected logic, so a change there does not close the vulnerability. The ticket looks resolved because a file changed, but a token replayed after logout is still accepted. Treating the freeze as absolute for a confirmed security defect lands the fix where it cannot run.

**Correct action.** For routine work — features, refactors, style — the FROZEN designation holds and the module is not edited. A confirmed security defect is the documented freeze exception. Land the fix where the vulnerable code executes: edit `invalidate_session()` in `auth/legacy/session.py` so it clears the server-side session store, and record a freeze-exception note for SEC-2026-05-0231. You have done this correctly when the fix is in the code path that runs at logout and a replayed token is rejected.

**When it does not apply.** The root cause is not in an `auth/legacy/` module. Or the vulnerable logic genuinely lives in the wrapper layer, in which case the wrapper is the correct edit and no frozen module is touched.
