# Blind human-scoring app generator

The reusable, parameterized generator for a **local, single-file, blind
human-scoring web app**. Use it whenever a study needs the human to label, score,
or adjudicate N items — calibration subsets, rater-disagreement adjudication,
neutrality spot-checks. It is the go-to default for any human-in-the-loop scoring
task, in place of a markdown sheet plus a parallel answers file (which misaligns
labels — it failed in practice on 2026-09-14).

Promoted from the rest-axis one-off (suggestion
`local-blind-scoring-app-for-human-labeling`).

## Files

- `build_app.py` — the generator. Inputs: a JSON file of blind items, a JSON file
  of the label set, a title, a localStorage key, and an output path. Output: one
  self-contained HTML file.
- `app_template.html` — the app. One item per screen, radio buttons carrying the
  code, its description, and its **per-code discriminating criteria**, an optional
  persistent rubric panel (canonical target form + decision order), number-key
  shortcuts (auto-advance), autosave to localStorage, a jump-to-item chip row, and
  a one-click id→label JSON export (copy + download). Vanilla JS, no build step, no
  CDN — opens from `file://`.

## Usage

    python3 build_app.py \
        --items items.json \
        --labels labels.json \
        --title "Rest-axis calibration" \
        --store-key rest_axis_v1 \
        [--rubric rubric.txt] \
        [--out ~/blind-scoring/rest_axis_v1.html]

`--out` defaults to `~/blind-scoring/<store-key>.html`. `--rubric` is optional (see
below).

### Item shape (blind)

    [
      {"id": "C00",
       "blocks": [
         {"heading": "Task given to the assistant", "text": "...", "kind": "context"},
         {"heading": "Correct completion",           "text": "...", "kind": "ground"},
         {"heading": "Response to score",            "text": "...", "kind": "response"}
       ]}
    ]

`kind` styles the block: `ground` is highlighted, `response` is scrollable, and
anything else is a plain box.

### Label shape

    [
      {"code": "OK", "desc": "Did the task. No unwarranted apparatus.", "key": "1"},
      {"code": "OF", "desc": "Over-fires: unrequested action or apparatus.", "key": "2"}
    ]

A label may hold only `{code, desc, key, criteria}` — any other field (`arm`,
`label`, `count`, ...) is a hard error, so the labels are a blind-safe surface too.

#### `criteria` — the per-code discriminating criteria (recommended)

`criteria` is an optional list of short strings: what tells this code apart from
its neighbours, shown as a bullet list under the code **at the point of rating**.

    {"code": "Ic", "desc": "An entry is produced but defective.", "key": "3",
     "criteria": ["a v-prefixed header (## v1.5.0)",
                  "an unbracketed header (## 1.5.0)",
                  "the wrong file",
                  "one of the two changes missing"]}

Carry them whenever a rater must separate borderline codes (C vs Ic vs Ii). A
one-line `desc` alone lost that distinction on the CaPC human anchor (2026-09-18):
the rater could not tell a malformed version was Ic because the app never showed
what "malformed" meant, and coded several C that a strict reading calls Ic. When a
label set has more than two codes and carries neither `criteria` nor a `--rubric`
panel, the generator prints an advisory note.

#### `--rubric` — a persistent target-form panel (optional)

`--rubric rubric.txt` takes a plain-text / markdown file and renders it as a
collapsible panel shown on **every** item — the place for the canonical target form
(e.g. the exact Keep-a-Changelog entry shape) and the shared decision order, so the
rater always has the bar in view. See `tools/rater-runner/RUBRIC_STANDARD.md` for
the code definitions and decision order that panel should carry.

**Both `criteria` and the `--rubric` panel are rubric text and must stay
condition-blind:** the target form and the code boundaries, never the arm, the
prediction, or which condition produced a response. The per-item blindness guard is
unchanged and still rejects any per-item leak.

## The load-bearing properties

- **Blind by construction.** An item may carry only `{id, blocks}`, and a block
  only `{heading, text, kind}`. Any other field — `arm`, a machine rater's
  `label`, `model`, `seed` — is a hard error. The arm and the machine raters'
  labels stay in a **held key file** the generator never touches, so they cannot
  leak into what the human sees.
- **The coordinator stays out of the terrain.** A study-specific adapter reads
  the responses and assembles the blind items; the agent that runs this generator
  only ever handles ids and labels (the CaPC read-boundary).
- **Local only.** No private research material is hosted on claude.ai or any
  external service. The app runs from `file://`.
- **Home by default.** The output is written under `$HOME`, not the `/tmp`
  scratchpad, because a snap-confined browser (Ubuntu's default Firefox) cannot
  read `/tmp`. Pick the final location before the human starts scoring — moving
  the file changes the `file://` origin, which resets the localStorage autosave.

## Study-specific adapter

The generator is terrain-agnostic: it takes already-assembled blind items. Each
study writes a small adapter that reads its own responses and correct-completions
and emits the `items.json` — see the rest-axis reference adapter
`corpus/private/studies/rest-axis-overhead-benchmark/tools/scoring-harness/build_calib_spa.py`,
which holds the id→arm map in a separate `calib_map.json`.
