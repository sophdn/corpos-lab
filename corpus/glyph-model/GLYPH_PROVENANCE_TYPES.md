---
type: reference
last_updated: 2026-04-14
---

> **Reference data promoted to typed corpus (2026-04-14).**
> Provenance type definitions now live as typed TOML at
> [`glyph-data/provenance-types/`](../glyph-data/provenance-types/).
>
> Programmatic access: `glyph_corpus::GlyphCorpus::get_provenance_type(name)`.
>
> This file remains as the human-readable taxonomy narrative. The typed
> entries are the source of truth for Step 1c provenance identification
> in the decomposition process. When the prose and the typed entries
> diverge, the typed entries win.
>
> Filed by: T2 of CHAIN_typed-corpus-foundation_2026-04-14.

# Glyph Provenance Types

**Status:** Active
**Date:** 2026-03-29

---

## What Provenance Type Names

The requirement to name provenance type when an information-substitution mechanism is present lives in `process-docs/glyph-model/GLYPH_WRITING_SPEC.md` (Y — the decision point section). This document defines the taxonomy used to satisfy that requirement: the five types, their definitions, and the discriminating conditions that distinguish them.

Provenance type identifies *why* the agent's held information is not equivalent to the canonical record — the mechanism that makes the terrain load-bearing.

---

## The Five Types

| Type | Definition | The gap |
|---|---|---|
| **Existence** | The information exists only in the agent's context — no committed artifact was produced | There is no record to verify against; the canonical form was never created |
| **Recency** | A committed record exists but was produced at a prior point; the agent is acting on it without verifying it is current at session time | The record may have changed since the agent last saw it |
| **Authorization** | The information was not produced by or confirmed by the authoritative source for this decision | Source identity or applicability has not been verified — fixable by re-sourcing to the correct authority |
| **Propagation** | A state change was committed in one location but was not propagated to all dependents that require it | Some dependents are operating on a state the primary record no longer holds |
| **Anachronicity** | The information belongs to a different operational frame from the canonical record it is substituting — categorically invalid as that record regardless of accuracy, recency, or source, because the frame it belongs to cannot produce the required record type | No reconfirmation or re-sourcing resolves the mismatch; producing a new artifact in the correct frame is the only resolution. *Illustration:* an agent whose knowledge of what a formal step requires was acquired through general familiarity rather than by running the formal step — the knowledge may be accurate, current, and from a valid source, but it belongs to the general-knowledge frame rather than the formal-step-output frame; it cannot constitute a formal-step record regardless of its accuracy. *Discrimination from Authorization:* Authorization is a source gap — closed by reading or consulting the correct authority; knowledge from the right place is what's needed. Anachronicity is a process gap — the required artifact can only be produced by executing the process, not by consulting any source; knowledge of what the process requires is not the same as the record the process produces when run. *Discriminating test:* after re-sourcing to the correct authority, is the required record type now producible? If yes — Authorization. If no — Anachronicity. |

**Propagation note:** Propagation has a different topology from the other four. Existence, Recency, Authorization, and Anachronicity are all agent-to-source gaps: the question is whether the agent's held knowledge corresponds to the canonical record. Propagation is a source-to-dependents gap: the primary record is correct, but hasn't reached all locations that depend on it. This is a distribution failure rather than a source-verification failure. When multiple dependents are involved, check carefully whether the mechanism is an agent failing to verify against the source (one of the other four types) or a state that was correctly committed but incompletely propagated (Propagation).

---

## Discriminating Condition — Working Definition

The provenance type precision requirement applies when the information-substitution mechanism is present.

**Working definition:** The information-substitution mechanism is present when the violation is possible only because the agent holds knowledge about a formal state without that knowledge having been produced by the canonical source at session time. *"At session time"* distinguishes Recency from Existence: for Recency, a canonical record exists — the gap is that the agent hasn't verified it is still current in this session. For Existence, no canonical record exists at all. In both cases, the agent is acting on held knowledge rather than on a freshly verified record; the difference is whether the record exists to be verified.

---

