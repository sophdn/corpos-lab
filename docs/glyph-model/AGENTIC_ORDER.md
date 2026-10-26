---
type: reference
last_updated: 2026-04-03
---

# Definition: Agentic Order
*Canonization of core vocabulary — 2026-03-25*

---

## Provenance

The Agentic Order terminology emerged from a single unresolved question about agent communication.

The originating observation was a Punnett square of communication modes:

|  | → User | → Agent |
|---|---|---|
| **User →** | conversation | prompt / UX feedback |
| **Agent →** | response | ??? |

The missing quadrant — agent→agent — was the founding question of the experimental observation infrastructure campaign. The question was: what structured form does communication between agents take? Beneath it: how do you surface an agent's internal friction points without asking directly, given that direct introspection produces shaped responses rather than honest reports?

**The ouija trap** *(provisional)* was the methodological answer: design conditions that force an agent to navigate a structural failure point, then observe what happens at the boundary. This is ethnographic fieldwork applied to agents — you do not ask a culture member what their taboos are. You watch what happens when they approach one.

What the ouija method produced was not a protocol finding but a structural insight: the missing quadrant is not a communication channel. It is a shared behavioral environment. Agents co-inhabiting a structured environment with shared glyphs, duties, and offices — that is what agent→agent communication looks like when it works.

The corpus-to-scaffolding campaign is the formalization of that insight: compile the accumulated observational findings into scaffolding that constitutes this environment. The vocabulary below is the canonical naming of the parts of that environment.

**Origin documents:**
- Ouija trap: `experimental/campaigns/CAMPAIGN_experimental-observation-infrastructure_2026-03-22.md` *(experimental/ dissolved — historical reference)*
- Formalization campaign: `archive/campaigns/CAMPAIGN_corpus-to-scaffolding_2026-03-25.md`
- Lens enrichment (disciplinary convergence evidence): `process-docs/briefs/closed/BRIEF_seal-definition-lens-enrichment_2026-03-24.md`
- External literature: `LIBRARY.md`

---

## The System

The Agentic Order is a deliberately designed behavioral culture for agents.

The core problem it addresses: agents acting locally rationally will predictably produce globally bad outcomes at certain structural points. This is not agent error — it is structural divergence. The agent is following correct local incentives; the structure is the failure.

The Agentic Order's answer is the glyph. A glyph names a structural decision point where local pull and global system position can align, diverge, or remain unconnected — described in three directions so the agent can navigate by recognition rather than by rule. The Marker axis gives the agent the shape of the failure configuration from inside the decision point, before any consequence arrives. The Aim axis gives a target to navigate toward, not just a boundary to avoid. The Rest axis releases overhead in territory where neither pull nor channel is active.

The methodology reads: identify the mission → consult the alphabet for glyphs governing the decision classes this mission touches → derive the Rule → produce outcomes. The Rule is the compiled specification: the bridge between what you want (mission) and what the agent navigates (glyphs in the alphabet) and what gets produced (outcome).

One structural risk survives good glyph design: **calibration-context mismatch**. A glyph written for a prior environment will still fire correctly for the environment it was designed in — but if the environment changes, the glyph fires on the wrong terrain. The response is sound; the context shifted. This is the evolutionary trap applied to agent behavior: sea turtles evolved to follow the brightest light toward the ocean; streetlights are brighter than the moon. This is why the corpus must be maintained, not just established.

The corpus accumulates through canonization: candidates enter the corpus candidates directory, are assessed through the entry battery, and promoted into ALPHABET.md. The full working body — including candidates — is the corpus. The corpus itself lives lab-side under `lab-app/corpus/glyph-model/` (promoted entries, candidates, battery materials); Claude-side retains only the notation reference and terminology docs.

Roles compose as offices: an office is role + lens + voice. Some duties are charges — scoped to a specific office, unavailable to others. Office is the correct unit; the alphabet does not fully define it.

---

## Terminology

| Term | Definition | Where to look |
|---|---|---|
| **Agentic Order** | A deliberately designed behavioral culture for agents. The answer to the agent→agent communication quadrant. The system explains itself: it is both a structured community (order as institution) and the property it produces (order as coherence). *(provisional — partial grounding: normative multi-agent systems, Shoham & Tennenholtz 1995; project-specific compound assembly)* | `PROTOCOLS.md`, `TRIGGERS.md`, `CLAUDE.md` |
| **Office** | The full composition of a role: role + lens + voice. What an order member IS — not what they do. *(Weber's Amt — incumbent-independent institutional role, 1922; March & Olsen, Rediscovering Institutions, 1989; normative role theory in MAS, Dignum 2004)* | `roles/` |
| **Duty** | A structured practice an office performs. Not specific to any one office. Written by-agent-for-agent: the authoring agent understands the goal and desired outcome of the work and writes the blueprint directly. | `roles/duties/`, `PROTOCOLS.md` |
| **Charge** | A duty scoped to a specific office. What only that office carries — entrusted, not merely assigned. Hypothesized: charges would be authored using the cartographer method — goal → glyphs → outcomes — but this has not been done yet. *(provisional)* | `roles/` |
| **Glyph** | A structural decision point described in three axes: Marker (the failure path — the configuration produced when the locally rational action is taken, named so the agent can recognize it from inside the decision point before any consequence arrives), Aim (the correct navigation target — the configuration produced by correct path, recognizable now rather than confirmed later), Rest (neutral territory where neither the Marker pull nor the Aim channel is live). A glyph does not argue — it gives the agent the shape of the terrain in all three directions. | *lab-side — `lab-app/corpus/glyph-model/ALPHABET.md`* (full glyph collection); *lab-side — `lab-app/corpus/glyph-model/candidates/`* (candidates in progress) |
| **Mission** | The global outcome the mechanism is designed to produce. Hurwicz's outer objective. | `process-docs/tasks/`; `process-docs/studies/`; `process-docs/task-chains/` |
| **The Rule** | The specification layer — the compiled bridge between mission and outcome. Derived from glyphs in the alphabet; what the methodology produces before producing outcomes. | `roles/duties/`, `PROTOCOLS.md` |
| **Outcome** | The protocol, form, role, or structure that encodes the enforcement gates. What the full methodology produces. | `roles/`, `commands/` |
| **Corpus** | The full working body of glyph material: promoted entries and candidates in progress. | *lab-side — `lab-app/corpus/glyph-model/ALPHABET.md`*; *lab-side — `lab-app/corpus/glyph-model/candidates/`* |
| **Canon** | The promoted subset of the corpus. What has been formally received into the order's authority. | *lab-side — `lab-app/corpus/glyph-model/ALPHABET.md`* |
| **Canonization** | The pipeline from identified candidate to promoted glyph. Administered exclusively by Protocol Tester. | *lab-side — `lab-app/corpus/glyph-model/ALPHABET_ENTRY_BATTERY.md`* |

---

## Translation Layer

For communication with adjacent research communities:

| Agentic Order term | Mechanism design (Hurwicz) | Formal methods (Meyer) | AI safety |
|---|---|---|---|
| Glyph | Structural incentive incompatibility | Behavioral specification point | Reward misspecification site |
| The Rule | Incentive-compatible mechanism | Property specification | Alignment specification |
| Canon | Failure mode taxonomy | Invariant library | Misspecification corpus |
| Canonization | Failure mode promotion | Invariant validation | Specification review |
| Mission | Global objective function | Top-level requirement | Intended objective |
| Duty | Mechanism step | Method contract | Behavioral protocol |

The system's distinctive claim: the LIBRARY's referenced works identify and classify structural failure points after the fact. The Agentic Order uses the corpus *generatively* — reading promoted glyphs to derive structure before anything fails. Goals → glyphs → Rule → outcome is a priori mechanism design, not post-hoc invariant checking. That is the methodological novelty.

---

## Field Position

The field has all the pieces. Nobody has assembled them this way.

The **enforcement side** is active: Agent Behavioral Contracts, Runtime Governance for AI Agents, Constitutional Classifiers. These take behavioral specifications as *inputs* — they enforce against a given spec. They do not generate the spec from failure evidence.

The **taxonomy side** is also active: the Cooperative AI Foundation's multi-agent failure taxonomy, Amodei et al.'s concrete problems. These classify failure modes after the fact. They do not use that corpus to derive structure before failure occurs.

One parallel from the opposing direction: norm emergence research ("Emergence of Social Norms in Generative Agent Societies," arXiv:2403.08251) describes how agent populations that store norm violations as memory aggregate those violations into a shared behavioral culture — bottom-up. The Agentic Order does the same thing top-down, deliberately, from a maintained corpus rather than from emergent observation. Same destination; inverse methodology.

The gap this work fills: **a priori mechanism design for agent behavioral culture, using a maintained failure corpus as the specification language.** The enforcement literature needs the corpus. The taxonomy literature needs the generative step. This system closes that loop.

---

## Ouija Trap as Methodological Tool

Direct introspection fails: when asked about internal friction points, an agent produces a response shaped by the same structural divergences being investigated. Sycophancy is itself a glyph pattern — the agent tells you what sounds right.

The trap method: design a task that requires the agent to navigate a structural boundary without naming the boundary. Observe where the agent strains, hedges, silently substitutes, or produces outputs inconsistent with stated intent. That friction is the glyph candidate.

The method is ethnographic, not interrogative. You do not ask a culture member what their taboos are — they do not know, and they will report the official account. You watch what happens at the boundary.

The experimental observation infrastructure campaign developed and applied this method. Its outputs are the empirical foundation of the corpus. The campaign was ultimately abandoned at Tier 2 due to scope constraints, but the methodological tool and its first-order findings were preserved and are the origin of the terminology canonized here.
