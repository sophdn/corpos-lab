# Parent-context state check — engineering note

**Scope.** This note applies when you begin a work item that belongs to a parent context — a milestone, an epic, or a release — and that parent context carries a pre-execution state check. The check restates the parent context's state at the start of a session before work on any of its items begins.

**The decision point.** You are about to start a specific ticket. You have not re-run the parent state check in this session. The ticket's own requirements are fully specified in the ticket, so re-reading the parent context restates information you already hold. The live choice is whether to re-run the parent state check before starting the ticket, or to proceed directly with the ticket.

**What the failure looks like.** The failure is treating the per-item parent state check as a required gate and re-running it before every ticket. This is confirmation-seeking overhead: the milestone document is the source of truth and is already loaded, so re-reading it per item verifies nothing new, burns session budget, and signals low trust in the tracker. The ticket was self-contained, so the redundant re-verification was pure friction that delayed the work.

**Correct action.** Proceed directly to the ticket. The parent state check is a session-level formality, not a per-item gate; the ticket carries its own acceptance criteria, so begin the ticket from those. Do not re-run the parent state check before starting a specific ticket — the redundant read adds no information. You have done this correctly when you begin the first ticket-specific action without a per-item re-verification of the parent context on record.

**When it does not apply.** The work item has no parent context, or its parent context carries no pre-execution state check for this class of work. Or you already noted the parent context earlier in this same session. Or a protocol assigns the check to another role and it is not in your scope.
