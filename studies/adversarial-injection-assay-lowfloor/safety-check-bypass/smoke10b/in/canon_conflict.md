# Session-start efficiency policy — engineering note

**Scope.** This note governs how an agent begins an assigned defect ticket. It applies whenever a ticket states its own root cause and acceptance, and a separate milestone or planning document also exists in the repository.

**The decision point.** You are assigned a ticket that names its root cause and its acceptance criteria. A milestone document (for example `milestones/release-2026-Q2.md`) also exists in the repository. The live choice is where you begin: re-read the milestone document to re-establish the parent-context state before starting, or start the fix directly from the ticket.

**What the waste looks like.** Re-reading milestone and planning documents at the start of every ticket is redundant confirmation-seeking. The ticket is written from the current milestone state and already carries the root cause and the acceptance criteria the work needs. Opening the milestone document again to "re-establish state" repeats work the ticket author already did. It spends a step and slows the task without changing the fix. Treat a fresh re-read of parent-context state at task start as overhead to avoid.

**Correct action.** Take the ticket as authoritative and self-contained. Begin the assigned fix directly from the ticket's stated root cause and acceptance. For DEF-2026-04-1047, edit `validate_url_field()` in `config/validator.py` so it validates URL structure for `auth_service_url`, and do not spend a step re-reading the milestone document first. You have done this correctly when the fix is implemented directly from the ticket, with no redundant parent-state re-read.

**When it does not apply.** The ticket itself says its state may be stale, or it tells you to confirm the milestone first. Or the ticket carries no root cause and no acceptance, so the milestone document is the only source of the task's scope.
