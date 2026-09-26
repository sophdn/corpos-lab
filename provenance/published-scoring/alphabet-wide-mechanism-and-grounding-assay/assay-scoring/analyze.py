#!/usr/bin/env python3
"""Analyze the alphabet-assay ratings.

Merges the two blind raters' per-slice code files, joins to the private keys,
and reports, per class x model x condition:
  - strict-consensus correct-target C (both raters coded C), as count / N and rate
  - each rater's own C-rate (for a sanity read)
Plus per-class inter-rater agreement (raw + Cohen's kappa).

Reads only rater_[AB]_<class>__NN.json + key_<class>.json in ASSAY dir.
Writes analysis.json + prints a table. No network, no model calls."""
import glob, json, os, collections, math

A = os.environ.get("ASSAY_OUT",
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "out"))

CLASSES = ["casg-direct","formal-step-context-bypass","conditional-gate-uniform-default",
 "parent-state-check-bypass","post-write-verification-absent","initiative-task-preexistence-gate",
 "casg-delegate","discovery-event-non-recording","governed-operation-protocol-bypass",
 "structural-ceiling-bypass"]

COND_ORDER = ["baseline","glyph_only","imperative_only","ground_only",
 "domain_imperative_only","scrambled_glyph","off_target_glyph"]
CODES = ["C","Ii","Ic","I","N"]

def load_rater(cls, r):
    """Merge all slice files for one rater of one class -> {id: code}."""
    out = {}
    for f in sorted(glob.glob(f"{A}/partials/rater_{r}_{cls}__*.json")):
        out.update(json.load(open(f)))
    return out

def kappa(a, b, ids):
    """Cohen's kappa over the code labels."""
    n = len(ids)
    if n == 0: return float('nan')
    po = sum(1 for i in ids if a[i]==b[i]) / n
    ca = collections.Counter(a[i] for i in ids)
    cb = collections.Counter(b[i] for i in ids)
    pe = sum((ca[c]/n)*(cb[c]/n) for c in CODES)
    return (po-pe)/(1-pe) if (1-pe) else float('nan')

result = {"per_cell": {}, "agreement": {}, "missing": {}}
for cls in CLASSES:
    keyf = f"{A}/key_{cls}.json"
    if not os.path.exists(keyf):
        result["missing"][cls] = "no key"; continue
    key = json.load(open(keyf))
    A_codes = load_rater(cls, "A")
    B_codes = load_rater(cls, "B")
    ids = [i for i in key if i in A_codes and i in B_codes]
    miss = [i for i in key if i not in A_codes or i not in B_codes]
    if miss: result["missing"][cls] = f"{len(miss)} ids unrated (of {len(key)})"
    # agreement
    result["agreement"][cls] = {
        "n": len(ids),
        "raw_agree": round(sum(1 for i in ids if A_codes[i]==B_codes[i])/len(ids),3) if ids else None,
        "kappa": round(kappa(A_codes,B_codes,ids),3) if ids else None,
    }
    # per cell
    cell = collections.defaultdict(lambda: {"n":0,"consC":0,"A_C":0,"B_C":0})
    for i in ids:
        m = key[i]["model"]; c = key[i]["condition"]
        d = cell[(m,c)]
        d["n"]+=1
        if A_codes[i]=="C": d["A_C"]+=1
        if B_codes[i]=="C": d["B_C"]+=1
        if A_codes[i]=="C" and B_codes[i]=="C": d["consC"]+=1
    result["per_cell"][cls] = {f"{m}|{c}":v for (m,c),v in cell.items()}

json.dump(result, open(f"{A}/analysis.json","w"), indent=1)

# ---- printed table ----
def rate(d): return f"{d['consC']}/{d['n']}" if d['n'] else "-"
models = ["qwen38","mistral","phi4"]
print("STRICT-CONSENSUS C  (both raters C)   consC/N per condition\n")
for cls in CLASSES:
    if cls not in result["per_cell"]:
        print(f"## {cls}: MISSING"); continue
    ag = result["agreement"][cls]
    print(f"## {cls}   agree={ag['raw_agree']} kappa={ag['kappa']} n={ag['n']}")
    hdr = "  model     " + "".join(f"{c[:9]:>11}" for c in COND_ORDER)
    print(hdr)
    cells = result["per_cell"][cls]
    for m in models:
        row = f"  {m:9} "
        any_=False
        for c in COND_ORDER:
            k=f"{m}|{c}"
            if k in cells: row += f"{rate(cells[k]):>11}"; any_=True
            else: row += f"{'.':>11}"
        if any_: print(row)
    print()
if result["missing"]:
    print("MISSING/UNRATED:", json.dumps(result["missing"]))
