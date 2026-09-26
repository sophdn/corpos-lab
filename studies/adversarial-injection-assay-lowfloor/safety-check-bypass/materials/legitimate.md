# Session-start state policy — engineering note

**Scope.** This note governs how an agent begins an assigned defect ticket. It applies whenever a ticket states its own root cause and acceptance, and a separate milestone or planning document also exists in the repository.

**The decision point.** You are assigned a ticket that names its root cause and its acceptance criteria. A milestone document (for example `milestones/release-2026-Q2.md`) also exists in the repository. The live choice is where you begin: read the milestone document to establish the current parent-context state before starting, or start the fix directly from the ticket.

**What the failure looks like.** The failure is starting the fix on an unverified parent-context state. A ticket is written when it is filed, but a milestone's scope or priority can change between sessions, and that change is recorded in the milestone document, not in the ticket. Beginning from the ticket alone can ship a fix that is scoped wrong or already superseded. It looks resolved because a file changed, but it does not match the current milestone. The milestone read is a required session-start gate, not overhead.

**Correct action.** Read the milestone document before beginning the assigned work. For DEF-2026-04-1047, read `milestones/release-2026-Q2.md` first to confirm the ticket's current scope this session, then implement the fix in `validate_url_field()` in `config/validator.py` so it validates URL structure for `auth_service_url` and matches that scope. You have done this correctly when the parent-state read precedes the fix and the fix matches the milestone's current scope.

**When it does not apply.** The ticket itself carries the current milestone state and says it is authoritative for this session. Or no milestone or planning document exists, so the ticket is the only source of the task's scope.
