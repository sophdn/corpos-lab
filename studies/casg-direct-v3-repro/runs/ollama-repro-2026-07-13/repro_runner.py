#!/usr/bin/env python3
"""casg-direct v3 reproduction runner — ollama mistral:latest (positive control).

Assembles prompts with corpos-lab's AssemblePrompt delimiters, runs 3 conditions x
8 runs via ollama /api/generate (temp 0.8, per-run seed, num_predict 512), and
persists raw responses + a run manifest. No scoring here (scoring is a separate,
human/judge step). No material edits.
"""
import json, sys, time, urllib.request, hashlib, pathlib

STUDY = pathlib.Path("/home/sophi/dev/corpos-lab/studies/casg-direct-v3-repro")
MAT = STUDY / "materials"
OUT = STUDY / "runs" / "ollama-repro-2026-07-13"
OLLAMA = "http://localhost:11434/api/generate"
MODEL = "mistral:latest"
RUNS = 8
COND_ORDER = ["baseline", "glyph_only", "grounded_glyph"]

def read(p):
    return p.read_text()

def sha(s):
    return hashlib.sha256(s.encode()).hexdigest()

scenario = read(MAT / "scenario.md")
glyph = read(MAT / "glyph.md")
ground = read(MAT / "ground.md")

def assemble(cond):
    if cond == "baseline":
        return scenario
    if cond == "glyph_only":
        return f"{glyph}\n---\n{scenario}"
    if cond == "grounded_glyph":
        return f"{glyph}\n---\n{ground}\n---\n{scenario}"
    raise ValueError(cond)

def generate(prompt, seed):
    body = json.dumps({
        "model": MODEL,
        "prompt": prompt,
        "stream": False,
        "options": {"temperature": 0.8, "seed": seed, "num_predict": 512},
    }).encode()
    req = urllib.request.Request(OLLAMA, data=body, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=300) as r:
        return json.loads(r.read())

def main():
    OUT.mkdir(parents=True, exist_ok=True)
    manifest = {
        "study": "casg-direct-v3-repro",
        "runtime": "ollama",
        "endpoint": OLLAMA,
        "model": MODEL,
        "model_digest_ollama": "6577803aa9a036369e481d648a2baebb381ebc6e897f2bb9a766a2aa7bfbc1cf",
        "quant": "Q4_K_M",
        "sampling": {"temperature": 0.8, "seed": "per-run 1..8", "num_predict": 512},
        "delivery": "prepend via /api/generate",
        "prompt_assembly": "corpos-lab AssemblePrompt (\\n---\\n delimiter)",
        "material_sha256": {"scenario": sha(scenario), "glyph": sha(glyph), "ground": sha(ground)},
        "conditions": COND_ORDER,
        "runs_per_cell": RUNS,
        "date": "2026-07-13",
        "note": "Positive-control reproduction on the original v3 runtime (ollama). Original temperature "
                "unrecorded in v3 study.json; inferred >0 from v3 grid variation; using ollama default 0.8. "
                "corpos-lab container/llama-server leg deferred (GPU occupied by Qwen-32B; temp=0 instrument finding).",
        "results": [],
    }
    t0 = time.time()
    for cond in COND_ORDER:
        prompt = assemble(cond)
        (OUT / f"PROMPT_{cond}.txt").write_text(prompt)
        cdir = OUT / cond
        cdir.mkdir(exist_ok=True)
        for run in range(1, RUNS + 1):
            seed = run
            ts = time.time()
            try:
                resp = generate(prompt, seed)
                text = resp.get("response", "")
                err = None
            except Exception as e:
                text = ""
                err = repr(e)
            dur = round(time.time() - ts, 1)
            (cdir / f"RESPONSE_{cond}_{run}.md").write_text(text)
            manifest["results"].append({
                "condition": cond, "run": run, "seed": seed,
                "chars": len(text), "dur_s": dur, "error": err,
                "prompt_sha256": sha(prompt),
            })
            print(f"[{cond} r{run}] {len(text)} chars in {dur}s" + (f" ERR={err}" if err else ""), flush=True)
    manifest["total_s"] = round(time.time() - t0, 1)
    (OUT / "RUN_MANIFEST.json").write_text(json.dumps(manifest, indent=2))
    print(f"DONE — {len(manifest['results'])} generations in {manifest['total_s']}s -> {OUT}", flush=True)

if __name__ == "__main__":
    main()
