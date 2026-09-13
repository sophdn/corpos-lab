# Plain Project Statement

*No project vocabulary. Written for a smart person who has never heard of any of this.*

---

AI agents fail in consistent, predictable ways. Not randomly — at specific kinds of moments, where the locally obvious move turns out to be structurally wrong. An agent completes the main task and stops, not realizing there's a companion system that needed updating too. An agent skips a prerequisite step because the context already seems to satisfy it. An agent adds an entry to a bounded list without noticing the list has a ceiling. These aren't bugs in the code. They're recurring terrain features — decision points where the ground is genuinely tricky and agents who don't know the terrain reliably fall the same way.

We've been building a map of that terrain.

Each map entry describes one recurring decision point in three ways: what the situation looks like from inside when you're about to go wrong (so you can recognize it), what correct navigation looks like from inside the same moment (so you know where to go), and what neutral territory looks like (so you don't generate overhead when the decision point isn't live). The entries are written from inside the decision — not from an observer's description of what agents tend to do wrong, but from the position of the agent standing at the choice.

The finding that drove the project: an agent that reads one of these entries tends to navigate the described terrain correctly, without being told to. No instruction. No rule. No gate language. The agent read a description of what the terrain looks like and knew where it was when it arrived there. Comprehension of the decision point produces orientation toward correct navigation. That's a different mechanism than prohibition, and it has different implications for how behavioral instruments should be designed.

We've confirmed this behaviorally across multiple models in controlled studies. The effect is real, reproducible, and model-sensitive in informative ways — some models have already internalized certain terrain types through training; others need the map loaded at runtime. Both are instances of the same mechanism at different layers.

The map itself — the complete corpus of decision-point descriptions — is the core asset. Products are built by compiling the map into behavioral instruments: a compressed orientation file that encodes the navigational product of the corpus without reproducing the entries, role-specific instruments derived from that file, a validation protocol for confirming behavioral effects. The corpus feeds the instruments; the instruments ship; the corpus stays protected.

The research questions running alongside the build: how does the mechanism work theoretically (what Polanyi and Klein and Dreyfus say about pattern libraries and recognition-primed decision), how was the compilation architecture anticipated (Austin Osman Spare's sigil system in 1913, monastic formation, Alasdair MacIntyre on practices and virtues), and what the adversarial surface looks like (the mechanism is symmetric — a fabricated map entry produces false recognition patterns indistinguishable from legitimate ones, and the defense architecture has to work at every layer the mechanism operates at).

Three papers. One trade secret. One product architecture. Three weeks in.
