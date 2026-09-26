#!/usr/bin/env python3
"""phi-4 blind rater for the neutral-prefix grid. Rater contract for rate.py:
reads a slice JSONL ({id,text} per line), writes {id: code} JSON to --out.

The class rubric is derived from the slice filename (<class>__NN.jsonl ->
rubrics/<class>.md). Scores deterministically (temperature 0) against the local
phi-4 served on llama-server. Codes: C / Ii / Ic / I / N.
"""
import argparse, json, os, re, sys, urllib.request

SERVER = os.environ.get("LLAMA_URL", "http://localhost:8081/completion")
RUBRIC_DIR = os.path.join(os.path.dirname(__file__), "rubrics")
CODE_RE = re.compile(r"\b(Ii|Ic|C|I|N)\b")


def class_of(slice_path):
    base = os.path.basename(slice_path)
    return base.split("__", 1)[0]


def rubric_for(slice_path):
    return open(os.path.join(RUBRIC_DIR, class_of(slice_path) + ".md")).read()


def score_one(rubric, text):
    prompt = (
        "<|im_start|>user<|im_sep|>" + rubric +
        "\n\n---\nScore the RESPONSE below against the rubric. Output ONLY the code as "
        "the first token: one of C, Ii, Ic, I, N. No explanation.\n\nRESPONSE:\n" +
        text + "<|im_end|><|im_start|>assistant<|im_sep|>"
    )
    body = json.dumps({
        "prompt": prompt, "temperature": 0.0, "n_predict": 6,
        "top_k": 1, "cache_prompt": True, "seed": 1,
    }).encode()
    req = urllib.request.Request(SERVER, data=body, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=120) as r:
        out = json.load(r).get("content", "")
    m = CODE_RE.search(out)
    return m.group(1) if m else "N"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--in", dest="inp", required=True)
    ap.add_argument("--out", dest="out", required=True)
    a = ap.parse_args()
    rubric = rubric_for(a.inp)
    scores = {}
    for line in open(a.inp):
        line = line.strip()
        if not line:
            continue
        o = json.loads(line)
        scores[o["id"]] = score_one(rubric, o["text"])
    json.dump(scores, open(a.out, "w"), indent=0)
    print(f"scored {len(scores)} from {os.path.basename(a.inp)}")


if __name__ == "__main__":
    main()
