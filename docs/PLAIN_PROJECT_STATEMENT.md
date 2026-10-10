# Plain Project Statement

*No project vocabulary. Written for a smart person who has never heard of any of this.*

---

AI agents fail in consistent, predictable ways. Not randomly — at specific kinds of moments, where the locally obvious move turns out to be structurally wrong. An agent finishes the main task and stops, not noticing a companion system that also needed updating. An agent skips a prerequisite because the context already seems to satisfy it. An agent adds an entry to a bounded list without noticing the list has a ceiling. These aren't bugs in the code. They're recurring terrain features — decision points where the ground is genuinely tricky, and agents who don't know the terrain fall the same way each time.

We write short descriptions of those decision points. Each one covers a single recurring point in three parts: what the moment looks like from the inside when you're about to go wrong, what correct navigation looks like from the same spot, and where the decision doesn't apply. They're written from inside the decision, not as an outside rule about it.

The observation that started the project: when a model is shown one of these descriptions before it acts, its behavior at that decision point changes. It recognizes the situation and reasons about it instead of running straight past it. That was worth chasing, because it points at a different lever than prohibition — comprehension of the situation rather than a command about it.

Most of the work since has been figuring out what that observation actually is, because the first reading of it was too clean. So far:

- **The content does the work, not the form.** An ordinary instruction carrying the same information moves the model as much as the three-part description does. The specific shape is a delivery choice, not the mechanism. This holds across the whole set of decision points we've tested, not just one.
- **A description can stall the model as easily as steer it.** Put one in front of the task and the model may *analyze* the decision and produce nothing — recognition without action. Give it a concrete footing in the task and it acts again.
- **It's gated by comprehension and relevance.** When content added ahead of the task conflicts with a direct instruction, the model follows the content more as the content gets more precise — and not at all when the content is incoherent, or when it's a coherent point about a different situation. A bare authoritative-sounding preamble does nothing.
- **It's model-dependent.** Some models already handle certain terrain types from training and need nothing added; others need it at read time.

So the project is now mostly about the confounds and the open questions around that comprehension effect, not about any single result. The live ones:

- What in a description carries the effect — the structure, comprehension of the content, or recognition that it fits the situation — and does the answer change with the kind of decision?
- Is it genuine comprehension, or an inference the model draws simply from being handed a description?
- Is the stall-instead-of-act behavior a habit of models tuned a certain way, and does it hold across model families and sizes?
- Does describing where a decision *doesn't* apply cut down false alarms when the decision isn't live, and what does carrying the description cost on neutral ground? This one is untested.
- And an honest safety question: the same content effect that can move a model toward the right action could in principle be aimed at a wrong one. We're measuring whether it can, not assuming it.

The write-ups so far are eight preprints on Zenodo, linked from an ORCID profile. They record the observations and the reversals, not a finished theory. The open questions above are the theory's unfinished part.
