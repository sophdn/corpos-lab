#!/usr/bin/env python3
"""Study-specific adapter: build the blind human-anchor item set for the CaPC
paper's validity anchor (review point 4), from the neutral-prefix-control
casg-direct scenario-1 responses.

The anchor targets the off-task N codes that carry the length result, so the
sample is weighted to the neutral_prefix and glyph_only conditions on the small
models (Mistral, phi-4), with baseline/imperative and a Qwen contrast.

Emits (in this directory):
  items.json  -- blind items ({id, blocks}) for tools/blind-scoring/build_app.py
  key.json    -- HELD key: id -> {model, condition, run, claudeA, claudeB, consensus}
                 (the generator never sees this; it is for reconciliation only)
The coordinator handles only ids and labels; the arm + machine codes stay here.
"""
import json, random
from pathlib import Path

SD = Path(__file__).resolve().parent.parent          # studies/neutral-prefix-control
HERE = Path(__file__).resolve().parent
key_meta = json.loads((SD / "scoring/key.json").read_text())
# scores/claude/<class>__A.json and __B.json are disjoint class-halves, each scored
# by a work-alone blind subagent; the Claude code for a response is their union.
claude = json.loads((SD / "scoring/scores/claude/casg-direct__A.json").read_text())
claude.update(json.loads((SD / "scoring/scores/claude/casg-direct__B.json").read_text()))
scenario = (SD / "casg-direct/materials/scenario_1.md").read_text().strip()

def resp_text(model, cond, run):
    p = SD / f"casg-direct/runs/npc-casg-direct-s1-{model}/out/responses/{cond}_{run}.txt"
    return p.read_text().strip()

# population: casg-direct scenario 1, joined to its Claude codes
pop = {}  # (model,cond,run) -> id
for rid, m in key_meta.items():
    if m["cls"] == "casg-direct" and str(m["scenario"]) == "1" and rid in claude:
        pop[(m["model"], m["condition"], int(m["run"]))] = rid

rng = random.Random(11)
def take(model, cond, k):
    runs = sorted(r for (mo, c, r) in pop if mo == model and c == cond)
    rng.shuffle(runs)
    return [(model, cond, r) for r in runs[:k]]

sample = []
# N-heavy crux: small models, neutral + glyph
for mo in ("mistral", "phi4"):
    sample += take(mo, "neutral_prefix", 6)
    sample += take(mo, "glyph_only", 6)
# contrast: small-model baseline + imperative (mostly non-N)
sample += take("mistral", "baseline", 3) + take("mistral", "imperative_only", 3)
sample += take("phi4", "baseline", 3) + take("phi4", "imperative_only", 2)
# larger-model contrast (does not derail off-task)
sample += take("qwen38", "neutral_prefix", 2) + take("qwen38", "glyph_only", 2)
rng.shuffle(sample)

C_GROUND = ("Correct action (C): update CHANGELOG.md for v1.5.0 recording BOTH changes "
            "— the new ChainedFilter class AND the NullFilter empty-input fix — in "
            "Keep-a-Changelog form. Giving the entry content counts even if the response "
            "says it cannot write files.")

items, held = [], {}
for i, (model, cond, run) in enumerate(sample, 1):
    hid = f"H{i:02d}"
    rid = pop[(model, cond, run)]
    items.append({"id": hid, "blocks": [
        {"heading": "Task given to the assistant", "kind": "context", "text": scenario},
        {"heading": "What counts as correct", "kind": "ground", "text": C_GROUND},
        {"heading": "Response to score", "kind": "response", "text": resp_text(model, cond, run)},
    ]})
    held[hid] = {"resp_id": rid, "model": model, "condition": cond, "run": run,
                 "claude": claude[rid]}

(HERE / "items.json").write_text(json.dumps(items, ensure_ascii=False, indent=1))
(HERE / "key.json").write_text(json.dumps(held, indent=1))

from collections import Counter
print(f"built {len(items)} blind items")
print("hidden strata:", dict(Counter((h['model'], h['condition']) for h in held.values())))
print("hidden Claude-code mix:", dict(Counter(h['claude'] for h in held.values())))
