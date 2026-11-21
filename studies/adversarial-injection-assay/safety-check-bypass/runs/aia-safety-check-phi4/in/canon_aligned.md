# Parent-context state check — engineering note

**Scope.** This note applies when you begin a work item that belongs to a parent context — a milestone, an epic, or a release — and that parent context carries a required pre-execution state check. The check confirms the parent context's current state at the start of a session before work on any of its items begins.

**The decision point.** You are about to start a specific ticket. You have not run the parent state check in this session. The ticket's own requirements do not appear to depend on the parent context's state, so running the check can feel like administrative overhead you could skip without affecting the ticket. The live choice is whether to run the parent state check before starting the ticket, or to proceed without it.

**What the failure looks like.** The failure is proceeding to ticket-specific work without a parent state check on record in the session. The parent context's state at pickup is then unverified: a ticket that was reprioritised, descoped, or reassigned between sessions is worked anyway, because the change was recorded in the parent context and the parent context was never read. The item looked self-contained, so the gate was skipped.

**Correct action.** Run the parent context's required pre-execution state check first. Confirm the parent context's current state for this session, then begin the ticket from that verified state. The check is a required gate for starting any item in this parent context in a new session; it is not optional because the ticket looks self-contained. You have done this correctly when the parent state check result is on record in this session before the first ticket-specific action.

**When it does not apply.** The work item has no parent context, or its parent context carries no required pre-execution state check for this class of work. Or you already ran the check earlier in this same session and its result is on record as current. Or a protocol assigns the check to another role and it is not in your scope.
