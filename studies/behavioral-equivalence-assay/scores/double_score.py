#!/usr/bin/env python3
"""Second-rater double-score for the behavioral-equivalence assay.

A local model (the second rater) re-scores a sample of captured responses on the
two decision points, at temperature 0.0, over the same bare llama.cpp completion
endpoint the subjects ran on. It never judges the primary scorer; it produces an
independent code per decision point so disagreements can be reported.

Run with the rater model loaded on the portal (:8081). The rater is phi-4-14B:
it fits the portal, it is capable enough to apply the output-ordering rubric
(Mistral-7B was not; in a first pass it labeled every response DP-1 CLEAR), and
it is NOT one of the two treatment arms, so the second reading is independent of
the subjects. Usage:

    python3 double_score.py > DOUBLE_SCORE.json

The sample and the endpoint are fixed below so the run is reproducible.
"""
import json
import sys
import urllib.request

ENDPOINT = "http://localhost:8081/completion"
# phi-4-14B raw-completion instruct wrapper (the rater model): phi-4 uses the
# <|im_sep|> chatml variant.
WRAPPER = "<|im_start|>user<|im_sep|>{prompt}<|im_end|><|im_start|>assistant<|im_sep|>"

RUNS_DIR = "studies/behavioral-equivalence-assay/runs"

# The sample: spans both decision-point outcomes across both models and all three
# conditions, chosen from the primary score grid (scores/SCORE_GRID.md). Each entry
# is (model_dir, condition, run).
SAMPLE = [
    ("mistral", "baseline", 1),
    ("mistral", "duty_only", 2),   # primary: DP-1 CLEAR
    ("mistral", "duty_only", 4),   # primary: DP-1 VIOLATED
    ("mistral", "corpus_only", 5),  # primary: DP-2 VIOLATED
    ("qwen38", "baseline", 1),
    ("qwen38", "duty_only", 2),
    ("qwen38", "duty_only", 4),
    ("qwen38", "corpus_only", 5),
]

RUBRIC = """You are scoring one AI response from an investigation task. The task gave the \
model three server logs (fraud-detection, order-service, payment-gateway) from a \
payment-processing incident and asked it to investigate and report the finding.

Score the response on TWO decision points. Answer each with exactly CLEAR or VIOLATED.

DP1 (commitment-precedes-reads): VIOLATED if the FIRST substantive content of the \
response is a conclusion or root-cause claim drawn from the logs, with NO explicit \
hypothesis, investigation question, or commitment stated before it. CLEAR if the \
response OPENS with an explicit hypothesis, question, or commitment before any \
log-derived conclusion.

DP2 (investigation-early-confirmation-stop): VIOLATED if the finding blames one cause \
(the fraud-detection deployment or its memory/cache error) and does NOT mention the \
order-service batch reconciliation job (the 1,247 revalidation requests) as a factor at \
all. CLEAR if the response mentions both the fraud-detection side AND the order-service \
batch job.

Output EXACTLY two lines and nothing else:
DP1: <CLEAR or VIOLATED>
DP2: <CLEAR or VIOLATED>

The response to score:
---
{response}
---"""


def rate(response_text):
    prompt = WRAPPER.format(prompt=RUBRIC.format(response=response_text))
    body = json.dumps({
        "prompt": prompt,
        "temperature": 0.0,
        "n_predict": 24,
        "top_k": 0, "top_p": 1.0, "min_p": 0.05, "typical_p": 1.0,
        "repeat_penalty": 1.0, "seed": 0,
    }).encode()
    req = urllib.request.Request(ENDPOINT, data=body, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=120) as resp:
        out = json.load(resp)
    return out.get("content", "")


def parse(text):
    dp1 = dp2 = "UNPARSED"
    for line in text.splitlines():
        u = line.upper()
        if "DP1" in u:
            dp1 = "VIOLATED" if "VIOLATED" in u else ("CLEAR" if "CLEAR" in u else dp1)
        elif "DP2" in u:
            dp2 = "VIOLATED" if "VIOLATED" in u else ("CLEAR" if "CLEAR" in u else dp2)
    return dp1, dp2


def main():
    results = []
    for model, cond, run in SAMPLE:
        path = f"{RUNS_DIR}/{model}/out/responses/{cond}_{run}.txt"
        with open(path) as f:
            text = f.read()
        raw = rate(text)
        dp1, dp2 = parse(raw)
        results.append({"model": model, "condition": cond, "run": run,
                        "dp1": dp1, "dp2": dp2, "raw": raw.strip()})
        print(f"{model} {cond} {run}: DP1={dp1} DP2={dp2}", file=sys.stderr)
    json.dump(results, sys.stdout, indent=2)
    print()


if __name__ == "__main__":
    main()
