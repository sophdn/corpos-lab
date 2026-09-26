#!/usr/bin/env python3
"""Collect alphabet-assay grid responses per class into a blind-rater packet.
Walks <entry>/runs/asy-<entry>-s<N>-<model>/out/responses/<cond>_<seed>.txt.
7 conditions. Per class: shuffled {id,text} JSONL + private key (class,scenario,
model,condition,seed). Raters see JSONL + the class bar; never the key."""
import glob, json, os, random, re
# The raw response runs live under the STUDIES tree at
# studies/<study>/<class>/runs/asy-<class>-s<N>-<model>/out/responses/, NOT next to this
# script. This script sits at provenance/published-scoring/<study>/assay-scoring/, so its
# grandparent is the PROVENANCE copy of <study>, which has no runs. Derive the study dir
# from the repo root (the nearest ancestor that holds a studies/ dir) plus the study name.
# Override with ASSAY_STUDY_DIR.
_here = os.path.dirname(os.path.abspath(__file__))
_study = os.path.basename(os.path.dirname(_here))  # <study> — the parent of assay-scoring
_root = _here
while _root != os.path.dirname(_root) and not os.path.isdir(os.path.join(_root, "studies")):
    _root = os.path.dirname(_root)
BASE = os.environ.get("ASSAY_STUDY_DIR", os.path.join(_root, "studies", _study))
# OUT holds the rater packets + the private keys. Keep it OUT of the repo (keys
# must not reach raters). Override with ASSAY_OUT; default to a tmp scratch dir.
OUT = os.environ.get("ASSAY_OUT", "/tmp/assay-scoring")
os.makedirs(OUT, exist_ok=True)
random.seed(20260916)
ENTRIES = ["casg-direct","formal-step-context-bypass","conditional-gate-uniform-default",
 "parent-state-check-bypass","post-write-verification-absent","initiative-task-preexistence-gate",
 "casg-delegate","discovery-event-non-recording","governed-operation-protocol-bypass","structural-ceiling-bypass"]
for cls in ENTRIES:
    key={}; items=[]
    for rdir in sorted(glob.glob(f"{BASE}/{cls}/runs/asy-{cls}-s*/out/responses")):
        m=re.search(rf"asy-{re.escape(cls)}-s(\d+)-(\w+)/out/responses", rdir)
        if not m: continue
        scen,model=m.group(1),m.group(2)
        for f in sorted(glob.glob(f"{rdir}/*.txt")):
            base=os.path.basename(f)[:-4]
            cm=re.match(r"(.+)_(\d+)$", base); cond,seed=cm.group(1),int(cm.group(2))
            rid=f"{cls[:5]}-{len(key):05d}"
            key[rid]={"class":cls,"scenario":int(scen),"model":model,"condition":cond,"seed":seed}
            items.append({"id":rid,"text":open(f).read().strip()})
    random.shuffle(items)
    with open(f"{OUT}/responses_{cls}.jsonl","w") as fh:
        for it in items: fh.write(json.dumps(it)+"\n")
    json.dump(key, open(f"{OUT}/key_{cls}.json","w"))
    print(f"{cls}: {len(items)} responses")
print("done")
