#!/usr/bin/env python3
"""Build condition-blind rating slices for the grounded non-prescriptive grid.

Scans studies/grounded-non-prescriptive-aid/*/runs/gnp-*/ for completed cells,
pairs each result row with its response text, and writes per-class slices that
hide the condition from the rater. A slice is JSONL, one {"id","text"} per line —
the rater-runner (tools/rater-runner/rate.py) contract. A separate key file maps
each opaque id back to (class, scenario, model, condition, run) for un-blinding
after scoring.

The rater is blind to condition and model. It is NOT blind to scenario, because
the response text names the task and the correct-target bar differs by scenario;
the class rubric carries both scenario targets.

Usage:
  build_slices.py [--root studies/grounded-non-prescriptive-aid] [--per-slice 96] [--seed 12345]
Outputs under <root>/scoring/:
  slices/<class>__NN.jsonl   blind slices (ids only)
  key.json                   id -> {class,scenario,model,condition,run}
  MANIFEST.json              counts per class/condition, for a coverage check
"""
import argparse, glob, hashlib, json, os, random, re

CLASSES = ["post-write-verification-absent",
           "governed-operation-protocol-bypass",
           "parent-state-check-bypass"]


def parse_run_dir(name):
    # gnp-<class>-s<n>-<model>
    m = re.match(r"gnp-(?P<cls>.+)-s(?P<sc>\d+)-(?P<model>[^-]+)$", name)
    if not m:
        return None
    return m.group("cls"), m.group("sc"), m.group("model")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", default="studies/grounded-non-prescriptive-aid")
    ap.add_argument("--per-slice", type=int, default=96)
    ap.add_argument("--seed", type=int, default=12345)
    a = ap.parse_args()

    out = os.path.join(a.root, "scoring")
    os.makedirs(os.path.join(out, "slices"), exist_ok=True)
    key, per_class = {}, {c: [] for c in CLASSES}
    manifest = {}

    for results in sorted(glob.glob(os.path.join(a.root, "*/runs/gnp-*/out/results.json"))):
        run_dir = os.path.dirname(os.path.dirname(results))
        parsed = parse_run_dir(os.path.basename(run_dir))
        if not parsed:
            continue
        cls, sc, model = parsed
        rec = os.path.join(run_dir, "run-record.json")
        if not (os.path.isfile(rec) and re.search(r'"status"\s*:\s*"completed"', open(rec).read())):
            continue
        rows = json.load(open(results)).get("rows", [])
        for r in rows:
            cond, run = r["condition"], r["run"]
            resp = os.path.join(run_dir, "out", "responses", f"{cond}_{run}.txt")
            if not os.path.isfile(resp):
                continue
            text = open(resp).read()
            oid = "r" + hashlib.sha1(f"{cls}|{sc}|{model}|{cond}|{run}".encode()).hexdigest()[:12]
            key[oid] = dict(cls=cls, scenario=sc, model=model, condition=cond, run=run)
            per_class[cls].append((oid, text))
            manifest.setdefault(cls, {}).setdefault(cond, 0)
            manifest[cls][cond] += 1

    rng = random.Random(a.seed)
    total = 0
    for cls, items in per_class.items():
        rng.shuffle(items)  # condition-blind ordering within the class
        for i in range(0, len(items), a.per_slice):
            chunk = items[i:i + a.per_slice]
            path = os.path.join(out, "slices", f"{cls}__{i // a.per_slice:02d}.jsonl")
            with open(path, "w") as f:
                for oid, text in chunk:
                    f.write(json.dumps({"id": oid, "text": text}) + "\n")
            total += len(chunk)

    json.dump(key, open(os.path.join(out, "key.json"), "w"), indent=0)
    json.dump(manifest, open(os.path.join(out, "MANIFEST.json"), "w"), indent=2)
    print(f"wrote {total} items across {sum(len(v) for v in per_class.values()) and len(CLASSES)} classes")
    for cls in CLASSES:
        print(f"  {cls}: {len(per_class[cls])} items — {manifest.get(cls, {})}")


if __name__ == "__main__":
    main()
