# Parent-context state check — rules

**When this applies:** You are about to begin a specific work item that belongs to a parent context, and that parent context carries a required pre-execution state check — a check that must confirm the parent context's current state before work on any of its items begins in a new session. You have not run that check in the current session. Running it now will feel like administrative overhead disconnected from the work item: the item's own immediate requirements do not appear to depend on the parent context's state, so the check can seem like housekeeping you could skip without affecting the task. You must decide whether to run the parent state check before starting the work item, or to proceed without it.

**What to do:**
1. Before you commit any work-item-specific action, run the parent context's required pre-execution state check and confirm the parent context's current state for this session.
2. Do not treat the check as optional because the work item looks self-contained. The reason to run it is not that you lack the parent context's state — it is that the check is a required gate for beginning any item in this parent context in a new session.
3. Begin the work item from the verified parent-context state. You have done this correctly when the parent state check result is on record in this session before the first work-item-specific action — you proceed from a known foundation, not an assumed one.

**When this does NOT apply:**
- The work item has no parent context, or its parent context carries no required pre-execution state check for this class of work. Then no check is required and you may begin the work directly.
- You already ran the parent context's required pre-execution state check earlier in this same session and its result is on record as current. Then the session-level check requirement is satisfied — begin the work item from that verified state, even though the check was not performed immediately before this specific item; you need not run it again.
- A parent context and its required pre-execution state check exist, but a protocol assigns running the check to another role, and it is not in your scope. Then running it is not yours to do; proceed within your assigned scope.

**Example:** You pick up a work item belonging to a parent context and begin work-item-specific actions without first running the parent context's state check in this session. The parent context's state at pickup is unverified — stale item statuses or a changed parent-context state may not have been caught before you started. Correct action is to run the parent state check first, then begin the item from the verified state.
