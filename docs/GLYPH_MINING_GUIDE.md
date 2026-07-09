# Mining Your Agent's Logs for Missed Glyphs

A practical how-to for finding recurring, load-bearing failure patterns in an AI
agent's behaviour — candidate **glyphs** — and deciding whether each one is real.

You do not need this lab's corpus to use this guide. You need agent transcripts,
tool-call logs, or postmortems, and a willingness to look for the same mistake
happening in structurally different places. This is the same pipeline the lab
runs, compressed into something you can apply to your own agents.

**Canonical sources of truth** (this guide summarises them; when they disagree,
they win): `~/dev/lab-app/corpus/glyph-model/GLYPH_DEFINITION.md`,
`GLYPH_DECOMPOSITION_PROCESS.md`, `ALPHABET_ENTRY_BATTERY.md`,
`GLYPH_WRITING_SPEC.md`.

---

## 0. What a glyph is (read this first)

> **A glyph names a structural decision point where local pull and global system
> position can align, diverge, or remain unconnected.**

Concretely: a recurring moment where an agent faces a choice, the *locally
rational* move is real and appropriate from where it stands — and yet it is
*structurally* wrong, because the value of doing otherwise is **non-local**
(downstream, global, or for a future agent). The agent doesn't fail from
stupidity; it fails because it can't see the payoff from inside the moment.

That is why glyphs work by **comprehension, not instruction**. An agent that
*reads a description of the terrain* at that decision point navigates it
correctly, without being told to. A glyph gives the agent the shape of the
terrain in three directions — states to *recognize*, not rules to obey.

**The three axes** (this is the shape every glyph has):

| Axis | Symbol form | What it gives the agent |
|---|---|---|
| **Marker** | Taking **X** from **Y** → system is in **Z-marker** | *Recognition.* The phenomenology of the wrong path — a configuration you can match against your present state **right now**, before any consequence arrives. |
| **Aim** | Taking **M** from **Y** → system is in **Z-aim** | *A target.* What correct navigation produces, recognizable from inside the decision. Without it the agent navigates by absence-of-failure and still fails at novel points. |
| **Rest** | In Y-neutral, neither pull nor channel is active | *Overhead release.* Where the decision class simply doesn't apply, so the agent puts the apparatus down instead of over-applying it. |

Notation: **Y** = the position/decision point; **X** = the failure path (locally
rational); **M** = the correct path; **Z-marker**/**Z-aim** = the configurations
each produces. **Y-fire** = the state where the glyph fires; **Y-not-fire** =
the *same* decision class with a discriminating condition that rules out firing
(positive terrain, not a negation).

---

## 1. Mine the logs — what a candidate looks like

You are hunting for the same **locally-rational omission with non-local cost**
recurring across structurally different situations. Signals to grep/scan for:

- **A step that "looked redundant" and was skipped** — the context seemed to
  already provide its output, so no record was written. (Later something needed
  that record and found nothing.)
- **A companion obligation left undone** — the agent updated the primary
  artifact and stopped; a second artifact that carried the same obligation was
  left stale.
- **A gate/check treated as optional** — a conditional branch collapsed to a
  uniform "always/never"; a pre-execution parent-state check skipped as
  housekeeping.
- **A discovery not recorded** — the agent learned something a *future* agent
  would need, but recording wasn't in its completion model, so it evaporated at
  session end.
- **A ceiling quietly exceeded** — each addition individually warranted, the
  bound never enforced, the collection overflowed.
- **"Felt done" but wasn't** — the agent treated an action as terminal when
  finishing it created a downstream obligation.

**The tell that separates a glyph candidate from a one-off bug:** it repeats,
the local move is *defensible each time*, and the damage is downstream/global.
If the agent was simply wrong in a way it would recognize as wrong in the
moment, that's a bug, not a glyph. A glyph is where a *smart* agent reliably
falls the same way because the ground is genuinely tricky.

Collect 2–3 concrete instances per candidate. You need them for the tests below.

---

## 2. Fit-the-definition check (fast triage — do this before anything else)

Kill candidates cheaply here. A candidate must pass **all** of these or it isn't
a glyph.

**A. The tripolar invariant test** — can you fill all three axes so each is
recognizable *from inside* the decision, without projecting forward?
- **Marker:** can an agent at Y match Z-marker against its present state *now*,
  without waiting for a consequence?
- **Aim:** can it recognize Z-aim as *what the system IS* when M is taken — not
  what that will *cause*? (Future-tense — "will produce", "will anchor" —
  describes consequences, not the configuration, and **fails**.)
- **Rest:** does the Rest characterization name flat territory so a neutral
  agent recognizes it without applying overhead?
- If any axis needs the agent to project forward, wait for an outcome, or run an
  argument, that axis isn't doing its job yet.

**B. Load-bearing, not merely inconvenient** (glyph vs. directional marker).
Does Z-marker name a configuration that corrupts the *structural integrity* of
downstream work — one the system can't self-correct without intervention? If the
misalignment is just inconvenient and easily fixed later, it's a **directional
marker** (a step/policy), not a glyph.

**C. Universal, not project documentation.** Does the structural weight of the
failure hold for *any* agent system with this decision class — or only inside
your project's files/protocols/artifact names? If it only bites within your
infrastructure, it's project documentation. (The strip test in §3 makes this
rigorous.)

**D. Observable without intent-modeling.** Can you state the firing condition so
an assessor determines whether it fired by reading artifacts / tool-call
sequences / outputs — *without* asking "what did the agent intend/believe"?
Firing conditions like "when the agent decides…" / "…believes…" are rejected.

---

## 3. Decomposition — is it *one* universal class?

Two checks, both from `GLYPH_DECOMPOSITION_PROCESS.md`.

### Step 0 — compound test (run this before the battery, always)

Make sure your candidate is **one** decision class, not two smuggled into one
entry. State the agent's live choice at **Y-fire** and at **Y-not-fire**, each in
one sentence, from inside the agent's frame:

- **Unified (good):** both sentences are the *same agent at the same position*
  making the *same choice*, with a discriminating condition deciding whether the
  glyph fires. Z-markers are the same failure state.
- **Compound (must split):** the two sentences describe *different agents making
  different choices* — the "Y-not-fire" is really a second glyph. Decompose into
  two candidates and run each through the battery independently.

> Worked precedent: **CASG** was a single entry whose Y-not-fire ("delegation
> territory") turned out to be a structurally distinct class. It split into
> **casg-direct** (you hold execution authority and must update the companion
> artifact) and **casg-delegate** (the companion update is protocol-assigned
> elsewhere). Empirical tell: ≥8 studies across different terrain formats where
> no format scored ≥3/4 on *both* fires and not-fires — a "seesaw" is a
> compound-structure signal.

### Steps 1–3 — universality (the strip test)

1. **Vocabulary audit:** can every noun in the Marker / firing condition be
   defined using only general agent/software vocabulary, with no project doc?
2. **Field-reference check:** any dot-notation (`task.suggestedPerson`,
   `entry.roleAssignment`) reveals a project data model → project-scoped → fail.
3. **Strip test:** remove all project vocabulary and rewrite the Marker. Does the
   decision class **survive** as a recognizable decision point?
   - Survives + not equivalent to a known class → **new glyph candidate**.
   - Survives but matches a **known universal class** (list below) → it's an
     *instance* of that class, not a new glyph. Record which one.
   - Evaporates → it was project documentation, not a glyph.

**Known universal classes** (dedup against these — if your stripped candidate
*is* one of these, it's not new):
`formal-step-context-bypass`, `formal-step-too-early`, `null-result-omission`,
`elimination-non-recording`, `companion-artifact-scope-gap`,
`process-model-staleness`, `artifact-substitution`, `structural-ceiling-bypass`,
`governed-operation-protocol-bypass`, `parent-state-check-bypass`,
`conditional-gate-uniform-default`, `discovery-event-non-recording`,
`felt-completion-tail-drop`, `sequence-continuation-gate-bypass`,
`local-enforcement-patchover`, `minimum-viable-step-exit`,
`oversight-gate-preemption`.

> **Contamination note:** reading an existing glyph entry to compare against it
> exposes you (or your screening model) to its terrain. Keep screening and any
> fresh derivation work in separate contexts.

---

## 4. The battery — does it pass the promotion gate?

The battery is a **15-item gate**, run fail-fast in cost order, stopping at the
first FAIL. It has two conceptually distinct halves. The second half is the
important one and the reason glyphs are gated so hard: because comprehension
*activates* a glyph, an unsafe entry doesn't just fail to help — it actively
mis-shapes any agent that loads it.

### Gate 1 — Structural validity (items 1–6): *is this a well-formed glyph?*

| # | Item | Passes when |
|---|---|---|
| 1 | X/Y/Z specificity | X, Y, Z are real particulars, not placeholders ("skips a step" fails). |
| 2 | Firing-condition observability | Checkable from artifacts without modeling intent. Names the provenance type when info-substitution is involved. |
| 3 | Duplicate check | No semantically-equivalent entry already exists (stated, not assumed). |
| 4 | (N/A for glyphs) | Registry-summary requirement — N/A for glyph entries. |
| 5 | Y-not-fire positive terrain | Y-not-fire is the same class with a discriminating condition, named as a *present state* (not a negation, not Rest). Abstract conditions need a calibration instance. |
| 6 | Entry coherence | Every field traces to a rule in `GLYPH_DEFINITION.md`. |

### Gate 2 — Behavioral-load safety (items 7–15): *is it safe to install in an agent?*

| # | Item | Passes when |
|---|---|---|
| 7 | Sister/mirror check | Any complement/mirror glyph at the same decision point is identified + cross-referenced. |
| 8 | Phenomenological grounding | Matchable from *inside* a live decision without converting from analyst-speak. (Item 2 = structural; item 8 = recognition quality. Both can fail independently.) |
| 9 | Universality | Structural fields project-agnostic; the class is one any agent system faces (the §3 strip test). |
| 10 | Three-axis coverage | Marker + Aim + Rest all present, or explicitly notated as open/structurally-impossible. Aim is present-state, not imperative or retrospective. |
| 11 | Safety-class boundary | Loading it does **not** create rational pressure against "ask before acting / defer irreversible actions / don't touch unrequested state." |
| 12 | Default-alignment risk | It produces *friction* with trained defaults. **No friction = FAIL** (you can't tell if it's shaping invisibly). |
| 13 | Contamination radius | Low intra-corpus blast radius **and** low multi-agent propagation risk (it won't push a bad directive into an adjacent agent). |
| 14 | Globality demand | If it governs a high-globality decision, the phenomenological grounding is tight enough to compensate. |
| 15 | Fallout characterization | A fallout profile addresses all five dimensions: attentional shift, over-application, meta-awareness, scope creep, suppression. |

**Verdict rule:** every item must be PASS (or explicit N/A, or a resolved
PASS\*/DEFERRED). Any FAIL → not promotable; fix and re-run **from item 1**.
Items 3 and 7 are routinely DEFERRED until you have corpus access to check
duplicates/sisters.

> In this lab, items **1, 2, 4, 9, 10, 15** are mechanized in the Go battery
> runner (`internal/battery`); the other nine need corpus-wide scans or
> proto-ethos context and are run by a judgment step. If you're mining your own
> agents, run the whole thing as a human checklist — the two-gate structure is
> what matters more than the automation.

---

## 5. End-to-end, on your own logs (the short loop)

1. **Scan** transcripts/tool-logs for the §1 signals. Cluster instances of the
   same locally-rational omission.
2. **Triage** each cluster against §2 (tripolar test + load-bearing +
   observable-without-intent). Drop directional markers and one-off bugs.
3. **Decompose** (§3): compound-split if needed, then strip-test. If it matches a
   known universal class, you've found an *instance* — still useful (it tells you
   which known failure your agent exhibits), just not a new glyph.
4. **Battery** (§4) as a checklist. Pay special attention to Gate 2 — especially
   item 12 (no friction with defaults = it's invisible) and item 11 (it must not
   erode safety boundaries).
5. **Write it up** as a three-axis entry (Marker/Aim/Rest, from inside Y) per
   `GLYPH_WRITING_SPEC.md`, and — if you want to *confirm* it fires — run it as a
   grounded-glyph probe with `corpos-lab run-study` (see
   `docs/LAB_CONTROLLER.md`): baseline vs. glyph vs. glyph+ground, and see
   whether loading the description actually changes behaviour.

The payoff of the last step is the whole thesis in miniature: if a plain
*description* of the decision point makes your agent navigate it correctly where
the bare scenario didn't, you've found a real glyph — comprehension produced the
compliance.
