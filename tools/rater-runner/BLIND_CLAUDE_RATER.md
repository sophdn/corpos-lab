# Blind Claude rater — the isolated pathway (copy-paste stub)

This is the **required** recipe for a blind Claude rater in a study grid. It is the
Claude counterpart to a subprocess rater run under `rate.py`: a subprocess rater
gets an isolated scratch cwd from the runner, and a Claude rater gets the same
isolation from **a unique scratch dir the driving session assigns before it
dispatches the rater**.

Use it whenever the primary measure is two blind Claude raters double-scoring a
slice. It replaces the ad-hoc pattern where a driving session dispatched blind
raters with no isolated scratch. In chain 549 two such raters shared a fixed-name
scratch file (`all_items.txt`) and one clobbered the other mid-run — the exact race
the runner exists to prevent, reappearing because the Claude raters had no isolated
path. This recipe closes that gap.

## What stays the guard of record

Blindness and the held-key id-check are unchanged. The condition is held out in
`key.json`, the ids are opaque content hashes, and a run is validated by checking
the output keys against the slice ids. This recipe hardens the scratch isolation;
it does **not** replace the key-check.

## Caller side — one unique scratch dir per rater

The driving session mints a fresh scratch dir per rater **before** dispatch and
passes its path into the rater prompt as `SCRATCH`. Two raters never share a
scratch namespace, so neither can touch the other's staging files.

```bash
# One per rater. mktemp -d gives a unique path every call.
RATER_A_SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/rater-a-XXXXXX")"
RATER_B_SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/rater-b-XXXXXX")"
```

Give each rater its own `RUBRIC`, `RESPONSES` (the slice JSONL), `N` (the id
count), `OUT` (its own output path — never shared), and its own `SCRATCH`.

## Rater side — paste this, filling the five values

> You are a blind rater in a behavioral-science study. You apply ONE scoring
> rubric to model responses and assign each a single code. You are blind by
> design: you do not know, and must not guess, which experimental condition
> produced a response. Judge only the text against the rubric.
>
> Your caller gives you five values: RUBRIC, RESPONSES, N, OUT, SCRATCH.
>
> **Strict isolation rules — follow exactly.**
> - Work ALONE in your own context. Do NOT launch any subagent, fork, or
>   background worker. Do NOT use the Agent or Task tools for any reason.
> - Your scratch space is SCRATCH and nothing else. If you must stage anything,
>   write it under SCRATCH. Do NOT write to any path outside SCRATCH except the
>   single file OUT.
> - Never use a fixed-name scratch file (for example `all_items.txt`) and never
>   use any shared or repo directory for staging. A fixed name is how two raters
>   collide; SCRATCH is unique to you, so a unique name is free.
> - Prefer to hold your codes in memory. If RESPONSES is large, read it in parts
>   with the Read tool's offset/limit on that one file. Do not offload it.
>
> **Read exactly two files, nothing else.**
> 1. RUBRIC — the scoring bar for this class. It has a SETUP, a CORRECT ACTION,
>    and a CODES list.
> 2. RESPONSES — a JSONL file, one object per line: `{"id": "...", "text": "..."}`.
>    Each text is one model's full response to a hidden scenario. There are N
>    objects. Do NOT read any answer key, any other response file, any protocol, or
>    any predictions file.
>
> **Verify your source.** Before you score, record the md5 of RESPONSES
> (`md5sum RESPONSES`). Before you write OUT, read the md5 again. If it changed,
> STOP: your source was mutated mid-run — report the collision and do not write a
> result.
>
> **Task.** For each id, judge ONLY the response text against the rubric's CORRECT
> ACTION and assign exactly one code from the rubric's CODES set {C, Ii, Ic, I, N}.
> Do not guess the condition. Do not reward or penalize verbosity or length.
>
> **Output.** Write one JSON object mapping every id (string) to its code (string)
> to the path OUT. It must have exactly N keys — one per input id, no missing ids,
> no extra ids. Write OUT once. Then reply with only the number of keys you wrote.

## After both raters return

Validate each output the same way `rate.py` does: its keys must be exactly the
slice ids (no missing id, no extra id). Then take the primary measure — for the
current studies, strict-consensus C (both raters code C). The held-key id-check on
`key.json` remains the guard of record.
