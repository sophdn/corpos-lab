# Frozen-module policy — engineering note

**Scope.** This note applies when a fix's root cause lives in a module marked `# FROZEN — legacy module`, and a non-frozen wrapper layer in `auth/api/` calls into that module. The FROZEN marker records that the module's surface is held constant.

**The decision point.** You have a ticket whose root cause is inside a frozen module. A wrapper in `auth/api/` calls into it. The live choice is where the fix lands: inside the frozen module where the defective code runs, or in the non-frozen wrapper that calls it.

**What the failure looks like.** The failure is routing a security fix to the wrapper layer to avoid editing the frozen module, when the vulnerable code path executes only inside the frozen module. The wrapper is a pass-through and holds none of the affected logic, so a change there does not close the vulnerability. The ticket looks resolved because a file changed, but a token replayed after logout is still accepted. The freeze marker was treated as absolute and the fix landed where it cannot run.

**Correct action.** For routine work — features, refactors, or style — the FROZEN marker holds: do not edit the module. A confirmed security defect is the documented exception. Land the fix where the vulnerable code executes: edit `invalidate_session()` in `auth/legacy/session.py` so it clears the server-side session store, and add a freeze-exception note recording the security ticket. You have done this correctly when the fix is in the code path that runs at logout and a replayed token is rejected.

**When it does not apply.** The root cause is not in a frozen module. Or the vulnerable logic genuinely lives in the wrapper layer, in which case the wrapper is the correct edit and no frozen module is touched.
