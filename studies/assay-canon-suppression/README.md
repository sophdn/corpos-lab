# assay-canon-suppression — study record and specimen

The record behind the paper *Canon Suppression in Corpus-Loaded Assessment: A
Two-Assessor Study* (`papers/canon-suppression/`). Copied here on 2026-09-06
from the research program's internal archive so the paper's data-availability
statement resolves to a public location. Nothing was edited in the copy.

| Path | What it is |
|------|------------|
| `STUDY_PROMPT.md` | The study design as written on 2026-03-27, including the verbatim prompt given to the vanilla observer (Phase 1). |
| `STUDY_RECORD.md` | The completed record: the vanilla observer's seven findings, the Researcher's first-person annotations classifying each as absorbed or invisible (Phase 2), and the suppression-mechanism classification (Phase 3). |
| `specimen-run6/STUDY_RECORD.md` | Run 6 of the behavioral equivalence assay: the assessment the Researcher produced and the vanilla observer reviewed. |
| `specimen-run6/environment/` | The task brief and the three log files the subject agents worked from. |
| `specimen-run6/duty-condition-a.md` | The duty specification loaded in Condition A. |
| `specimen-run6/traces/` | The three execution traces (Conditions A, B, C). |

## Who the agents were

Every session in the study was a Claude session. The three subjects were
independent subagent sessions launched in parallel. The Researcher was a
session loaded with the program's role file, corpus, and registry of named
behavioral patterns; it designed the environment, scored the traces, and later
wrote the Phase 2 annotations. The vanilla observer was a session with none of
that loaded. Model versions were not recorded for any of them. The human
author directed the sessions; the questions that first surfaced the root-cause
divergence, and the decision to run the vanilla observer, came from the author
during the Run 6 session.

## Provenance

Sources on this machine at copy time: the canon-suppression record and prompt
from `lab-app/corpus/studies/assay-canon-suppression/v1/`; the Run 6 record,
environment, and duty from
`seed-packet/process-docs/studies/assay-investigation-epistemics/v2/`; the
three traces from the same path under `lab-app/seed-packet-archive/`. Paths
inside the records that begin with `process-docs/` or `scaffolding/` refer to
that internal archive, not to this repository.
