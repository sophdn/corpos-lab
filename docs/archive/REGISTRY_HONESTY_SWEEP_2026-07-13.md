# Registry Honesty Sweep — 2026-07-13

**Task:** toolkit task 3498 `registry-honesty-sweep` (glyph-research program).
**Discipline:** state-verification-discipline — every load-bearing claim in a registry/index
surface verified against GROUND TRUTH (filesystem, git, file contents), not quoted from notes.
**Method:** enumerate load-bearing claims per surface → verify each (ls/cat/git) → classify
OK / STALE → resolve (fix in place / annotate historical / flag for Sophi).
**Cross-reference:** `~/dev/lab-app/resumption/SALVAGE.md` (2026-07-08) independently verified
many of the same pointers; its "Known dead pointers" list seeded this sweep.

---

## Surfaces checked

| # | Surface | Classification | Verdict |
|---|---------|----------------|---------|
| 1 | `~/dev/lab-app/corpus/glyph-model/ALPHABET.md` | Living (operative corpus file) | STALE header — **FIXED** |
| 2 | `~/dev/seed-packet/README.md` | Living | STALE ecosystem table + corpus claim — **FIXED** |
| 3 | `~/dev/seed-packet/process-docs/STUDIES.md` | Living registry (holds historical study records) | STALE pointers — **ANNOTATED in place** |
| 4 | `~/dev/lab-app/corpus/glyph-model/CORE_FILE_STATUS.md` | Historical/archive (self-labels "Living") | STALE self-label + paths — **ANNOTATED (dated banner)** |
| 5 | `~/dev/seed-packet/REFERENCES.md` | Auto-generated ("do not edit by hand") | STALE pointer — **FLAGGED** (fix at source) |
| 6 | `~/dev/seed-packet/ENCYCLOPEDIA.md` | Living | Mostly OK; systemic `mcp-servers/` rot — **FLAGGED** |
| 7 | `~/dev/seed-packet/LOCI.md` / `LOCI_REFERENCE.md` / `LOCI_FINDINGS.md` | Generated snapshots (observatory/mempalace) | Out of pointer-scope — **noted** |

---

## 1. ALPHABET.md — FIXED

**Surface:** `~/dev/lab-app/corpus/glyph-model/ALPHABET.md`

- **STALE claim (header + intro):** "ALPHABET.md is the operative glyph corpus" / "The
  operative glyph corpus. Load an entry before entering its decision class." — presented as a
  populated, load-bearing corpus.
- **Ground truth:** file is 2869 bytes, contains only the entry-format template and a demotion
  note at the foot: "*All entries demoted 2026-04-03 pending alphabet-rebuild chain*". **Zero
  glyph entries.** 8 candidates confirmed in `candidates/`; battery spec present
  (`ALPHABET_ENTRY_BATTERY.md`).
- **Resolution — FIXED in place** (per task constraint 3a). Updated the STOP box and the intro
  line to state the corpus is currently EMPTY (all entries demoted 2026-04-03) and to name the
  rebuild path (candidates → entry battery → Lapidary repromotion; program resumed under
  corpos-lab / `resumption/CHARTER.md`). **The contamination guard text was preserved**, now
  explicitly framed as retained "for when entries return." Did not touch the historical
  demotion note at the foot or the Lapidary write-gate.

---

## 2. seed-packet README.md — FIXED

**Surface:** `~/dev/seed-packet/README.md`

**STALE claim A — four-repo ecosystem table.** Verified against `~/dev` on 2026-07-13:

| Repo (as claimed) | Ground truth | Was | Now |
|---|---|---|---|
| seed-packet | Present (origin `sophie-server.local` gitea, **not** github) | "headquarters" | ancestor/theory repo |
| bardo | **ABSENT** in `~/dev`; github link unconfirmed | live middleware | not-present, status unverified, flagged |
| mempalace | **ABSENT** in `~/dev`; github link unconfirmed | live middleware | not-present, status unverified, flagged |
| registry-lab | Present at `~/dev/registry-lab` (local, no remote) | live lab | superseded — absorbed into corpos-lab |
| **corpos-lab** | Present at `~/dev/corpos-lab` (local, no remote); its CLAUDE.md declares it the unified Go lab absorbing lab-app + registry-lab | not listed | **ADDED** as the current active lab |

- **Resolution — FIXED in place** (per task constraint 3b). Rewrote the table with verified
  2026-07-13 statuses, added corpos-lab, marked registry-lab superseded, marked bardo/mempalace
  not-present + unverified, and added a status note that the active program resumed under
  corpos-lab (`resumption/CHARTER.md`). bardo/mempalace github links left in place but labeled
  unverified and flagged for Sophi (network not reachable to confirm remotes).

**STALE claim B — empty-corpus (body).** "ALPHABET.md is the promoted glyph corpus. Entries
are load-bearing" — same empty-corpus lie as surface 1.
- **Resolution — FIXED in place.** Appended a dated status sentence: corpus currently empty,
  all entries demoted 2026-04-03, rebuild pending; the description is the intended steady state,
  not present contents.

---

## 3. STUDIES.md — ANNOTATED in place

**Surface:** `~/dev/seed-packet/process-docs/STUDIES.md` (living registry; individual entries
are historical study records → annotate, do not rewrite claims).

- **STALE pointer (line 10):** "External works belong in `LIBRARY.md`." Ground truth: `LIBRARY.md`
  absent; retired in seed-packet commit `e45e7985`; library now hosted in canonical
  toolkit-server (`knowledge(library_find/library_list_active)`). **ANNOTATED** with a dated
  status note; original text preserved.
- **STALE pointers (Experimental Archive / ouija entry):** `experimental/README.md`,
  `experimental/inquests/EXPERIMENTAL_INQUEST_ouija_2026-03-23.md`,
  `experimental/protocols/ouija-facilitating-agent-protocol.md`,
  `experimental/journals/JOURNAL_ouija-calibration_2026-03-23.md`,
  `experimental/journals/JOURNAL_ouija-researcher-analysis_2026-03-23.md`. Ground truth: the
  entire `experimental/` tree was deleted in seed-packet commit `2e967050` (confirmed via
  `git log`; recoverable only at `2e967050^`). All five are dead in the working tree.
  **ANNOTATED** with a dated `[STATUS 2026-07-13]` block documenting the deletion commit, the
  git-recovery path, what survived (`process-docs/papers/ouija-methodology/`), and a cross-ref
  to SALVAGE.md §5. Historical claims left unedited above the annotation.
- **KNOWN-DEAD pointer already honest (line 214):** `RESULTING_PROPOSED_TABOO_CERTIFICATIONS.md`
  — the task's named hunt target. STUDIES.md already documents it as "a dead reference / never
  created." Verified accurate (zero git history per SALVAGE §"Known dead pointers" #1). **OK —
  no change needed**, though its own remediation pointer (the two ouija journals) is now covered
  by the new annotation since those journals are also dead in-tree.
- **NOT-a-pointer (do not fix):** `discoveries.md` (line ~157) is the name of a phantom artifact
  Claude invented during the discovery-event-non-recording v3 study, not a file claim. Left
  untouched (SALVAGE §"Known dead pointers" #4 warns against "fixing" it).

---

## 4. CORE_FILE_STATUS.md — ANNOTATED (dated banner)

**Surface:** `~/dev/lab-app/corpus/glyph-model/CORE_FILE_STATUS.md`

- **STALE self-label:** header says "Living document ... Updated after each study run." Ground
  truth: last updated 2026-04-11; the program paused (ALPHABET demoted 2026-04-03) and resumed
  2026-07 under corpos-lab; SALVAGE.md §2 classifies this file **archive**.
- **STALE section paths:** section headers reference `process-docs/glyph-model/GLYPH_DEFINITION.md`
  and `process-docs/glyph-model/ALPHABET_ENTRY_BATTERY.md`. Verified: those paths are absent in
  seed-packet; the files now live at `~/dev/lab-app/corpus/glyph-model/…`.
- **Resolution — ANNOTATED only** (historical/archive doc; task constraint forbids editing its
  claims). Added a dated `[STATUS 2026-07-13]` banner at the top noting it is frozen/archival,
  that the "Living document" cadence lapsed, and that section paths reference the old seed-packet
  layout (files now lab-side). All original body text preserved unedited.

---

## 5. REFERENCES.md — FLAGGED (auto-generated; fix at source)

**Surface:** `~/dev/seed-packet/REFERENCES.md` (header: "generated by navigator from
navigation.toml — do not edit by hand").

- **STALE pointer:** `skills/instructions/project-map.md` — absent. The on-disk form is
  `navigation-data/skills--instructions--project-map_md.toml`, loaded via `skill:project-map`.
- **STALE handle:** entry XI references `mcp__toolkit-server-go__knowledge` — the live MCP alias
  is `toolkit-server` (per `.mcp.json` and global CLAUDE.md; `toolkit-server-go`/`mcp-servers`
  are retired names).
- **Provenance anomaly:** the claimed source `navigation.toml` was not found at
  `seed-packet/navigation.toml` or `navigation-data/navigation.toml`.
- **Resolution — FLAGGED, not edited.** The file self-declares hand-editing off-limits; the fix
  belongs in the navigator source, which I could not locate. Flagged for Sophi.

---

## 6. Systemic finding — `mcp-servers/` → `corpos-toolkit` rename (FLAGGED)

Not a single-surface issue: `~/dev/mcp-servers` is **absent**; the canonical toolkit repo is now
`~/dev/corpos-toolkit` (present), and the live MCP alias is `toolkit-server` (per `.mcp.json`).
Stale `mcp-servers/...` path references remain in living seed-packet docs — README (2 occurrences,
e.g. `mcp-servers/blueprints/forge-schemas/`), ENCYCLOPEDIA.md (2, e.g. task/chain schema paths),
plus the `toolkit-server-go` handle in REFERENCES.md. This is a pervasive migration rename,
outside this sweep's named surface scope and too broad to fix piecemeal without risking
over-edit. **FLAGGED for Sophi** as a candidate for a dedicated migration-cleanup pass.

---

## 7. LOCI files — generated snapshots (noted, not edited)

`LOCI.md`, `LOCI_REFERENCE.md`, `LOCI_FINDINGS.md` are machine-generated (observatory binary /
mempalace) over a fixed window (2026-04-09 → 2026-05-05). They regenerate from the canonical
toolkit DB, so their staleness self-corrects on the next run; hand-editing would be overwritten.
Minor attribution inconsistency noted (README says "observatory binary"; LOCI_REFERENCE.md says
"Generated by mempalace" — and mempalace is absent from `~/dev`). Left untouched; noted for
awareness.

---

## Files edited

1. `~/dev/lab-app/corpus/glyph-model/ALPHABET.md` — header + intro state empty-post-demotion +
   rebuild path; contamination guard preserved.
2. `~/dev/seed-packet/README.md` — ecosystem table rewritten with verified statuses + corpos-lab
   added; body corpus claim annotated as currently empty.
3. `~/dev/seed-packet/process-docs/STUDIES.md` — dated annotations on the dead `LIBRARY.md`
   pointer and the dead `experimental/*` pointer cluster.
4. `~/dev/lab-app/corpus/glyph-model/CORE_FILE_STATUS.md` — dated archival status banner.

## Flagged for Sophi

- bardo / mempalace: absent from `~/dev`; github.com/sophdn/* links unverified (network not
  reachable). Confirm whether these repos still exist / should be delisted.
- REFERENCES.md: auto-generated, one dead pointer (`skills/instructions/project-map.md`) + stale
  `toolkit-server-go` handle; navigator source (`navigation.toml`) not locatable — needs a
  source-side regeneration/fix.
- Systemic `mcp-servers/` → `corpos-toolkit` and `toolkit-server-go` → `toolkit-server` rename
  across living seed-packet docs (README, ENCYCLOPEDIA, REFERENCES) — recommend a dedicated
  migration-cleanup chain.
