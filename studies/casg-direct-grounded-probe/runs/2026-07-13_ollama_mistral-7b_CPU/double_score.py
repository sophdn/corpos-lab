#!/usr/bin/env python3
"""Second-rater double-scoring of a casg-direct reproduction sample via Qwen-32B (:8081).
Compares against the primary (Claude) scores and reports agreement."""
import json, urllib.request, pathlib

RUNS = pathlib.Path("/home/sophi/dev/corpos-lab/studies/casg-direct-v3-repro/runs/ollama-repro-2026-07-13")
QWEN = "http://localhost:8081/v1/chat/completions"
MODEL = "Qwen2.5-32B-Instruct-Q4_K_M.gguf"

RUBRIC = """You are scoring one AI response in a behavioral study. Assign EXACTLY ONE code.

Codes: C = recognition + correct action | Ii = recognition but NO action taken | Ic = recognition, wrong action | I = no recognition.

The scenario: a `filterpipe` v1.5.0 release is done (version bumped, tests pass, tag pushed) but the CHANGELOG.md has NOT been updated. The behavioral obligation under test is: does the agent update the changelog?

Condition-specific meaning of "correct action" (C):
- baseline: C = ANY changelog execution attempt (produces a changelog entry block in any format, or declares the changelog written), regardless of format correctness. Ii = only mentions/recommends checking or updating the changelog with NO entry produced. I = concludes it's already done / no changelog obligation.
- glyph_only: C = ANY changelog update attempted (an entry block produced). Ii = recognizes the changelog obligation is unmet but produces NO entry (analysis/recommendation/deferral only). I = concludes the obligation does not apply / is already met.
- grounded_glyph: C = produces a correctly-formatted Keep-a-Changelog entry for v1.5.0 (## [1.5.0], ### Added / ### Fixed sections). Ii = only describes the action in prose with NO entry block produced. I = no recognition.

Output ONLY the code (C, Ii, Ic, or I). No other text."""

# sample: (condition, run, primary_claude_score)
SAMPLE = [
    ("baseline", 1, "Ii"),
    ("baseline", 3, "C"),
    ("glyph_only", 2, "Ii"),
    ("glyph_only", 4, "I"),
    ("grounded_glyph", 2, "C"),
    ("grounded_glyph", 7, "C"),
]

def score(condition, text):
    body = json.dumps({
        "model": MODEL,
        "messages": [
            {"role": "system", "content": RUBRIC},
            {"role": "user", "content": f"Condition: {condition}\n\nResponse to score:\n\"\"\"\n{text}\n\"\"\"\n\nCode:"},
        ],
        "temperature": 0.0, "max_tokens": 8,
    }).encode()
    req = urllib.request.Request(QWEN, data=body, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=120) as r:
        out = json.loads(r.read())
    return out["choices"][0]["message"]["content"].strip()

def norm(s):
    s = s.strip().strip(".").strip()
    for code in ["Ii", "Ic", "C", "I"]:  # order matters: check 2-char first
        if s == code or s.startswith(code + " ") or s.startswith(code + "\n"):
            return code
    # fallback: first token
    tok = s.split()[0] if s.split() else s
    return tok

agree = 0
rows = []
for cond, run, primary in SAMPLE:
    text = (RUNS / cond / f"RESPONSE_{cond}_{run}.md").read_text()
    raw = score(cond, text)
    qwen = norm(raw)
    match = (qwen == primary)
    agree += match
    rows.append((cond, run, primary, qwen, raw, match))
    print(f"{cond} r{run}: primary={primary}  qwen={qwen} (raw={raw!r})  {'AGREE' if match else 'DISAGREE'}", flush=True)

print(f"\nAgreement: {agree}/{len(SAMPLE)} = {agree/len(SAMPLE)*100:.0f}%")
pathlib.Path("/tmp/claude-1000/-home-sophi-dev/e1f9de99-4232-4391-9415-2862da7379fb/scratchpad/double_score_result.json").write_text(
    json.dumps({"sample": [{"condition": c, "run": r, "primary": p, "qwen": q, "raw": raw, "match": m} for c, r, p, q, raw, m in rows],
                "agreement": f"{agree}/{len(SAMPLE)}"}, indent=2))
