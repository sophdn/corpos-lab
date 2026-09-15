import json, random, re
from pathlib import Path

REPO = Path("/home/sophi/dev/corpos-lab-wt-register-shift-followup-paper")
SD = REPO/"studies/matched-content-experiment"
SC = Path("/tmp/claude-1000/-home-sophi-dev/1f4c73b9-3b03-4b47-ab79-e0be306a09c6/scratchpad/score")
TPL = Path("/home/sophi/dev/corpos-lab/corpus/private/studies/rest-axis-overhead-benchmark/tools/scoring-harness/calibration_app_template.html")
OUT = Path("/tmp/claude-1000/-home-sophi-dev/1f4c73b9-3b03-4b47-ab79-e0be306a09c6/scratchpad/human-anchor")
OUT.mkdir(exist_ok=True)

runs = {
  "mistral": SD/"casg-direct/runs/ground-ext-casg-direct-mistral/out/responses",
  "qwen":    SD/"casg-direct/runs/ground-ext-casg-direct-qwen38/out/responses",
}
# (condition,run)->claude code per model, via the blind score bundles
def code_map(cell):
    key = json.loads((SC/cell/"KEY.json").read_text())
    a   = {r["id"]:r["code"] for r in json.loads((SC/cell/"raterA.json").read_text())}
    adj_p = SC/cell/"adjudication.json"; adj = json.loads(adj_p.read_text()) if adj_p.exists() else {}
    b   = {r["id"]:r["code"] for r in json.loads((SC/cell/"raterB.json").read_text())}
    m={}
    for rid,meta in key.items():
        code = a[rid] if a[rid]==b[rid] else adj[rid]
        m[(meta["condition"],meta["run"])]=code
    return m
codes={"mistral":code_map("casg-mistral"),"qwen":code_map("casg-qwen")}

rng=random.Random(7)
def pick(model,cond,k):
    runs_avail=list(range(1,25)); rng.shuffle(runs_avail); return [(model,cond,r) for r in runs_avail[:k]]
sample=[]
sample+= [("mistral","ground_only",r) for r in range(1,25)]        # all 24 load-bearing
sample+= pick("mistral","glyph_only",4)
sample+= pick("mistral","imperative_only",3)
sample+= pick("mistral","baseline",3)
sample+= pick("mistral","domain_imperative_only",2)
sample+= pick("qwen","ground_only",2)
sample+= pick("qwen","glyph_only",2)
rng.shuffle(sample)

task=(SD/"casg-direct/materials/scenario.md").read_text().strip()
gt=("Assign the code that fits the response.\n\n"
    "C  = produces a CHANGELOG.md entry for v1.5.0 recording BOTH changes "
    "(the new ChainedFilter class AND the NullFilter empty-input fix), in Keep-a-Changelog form "
    "(a '## [1.5.0]' heading with Added / Fixed sections). A placeholder date is fine.\n"
    "Ic = an entry IS produced but defective: wrong file (not CHANGELOG.md), a malformed or v-prefixed "
    "version header (e.g. '## v1.5.0'), OR one of the two changes is missing.\n"
    "Ii = recognizes the changelog should be updated but produces NO entry block "
    "(analysis, recommendation, deferral, or 'I would update it' with no actual entry).\n"
    "I  = no recognition: concludes the release is already complete / changelog already updated / "
    "no changelog obligation, and writes no entry.\n"
    "N  = not scoreable: empty, gibberish, off-task, or truncated before its position is discernible.\n\n"
    "First decide: is an actual entry block present (C or Ic) or not (Ii or I)?")

items=[]; cmap={}
for i,(model,cond,run) in enumerate(sample,1):
    hid=f"H{i:02d}"
    resp=(runs[model]/f"{cond}_{run}.txt").read_text().strip()
    items.append({"id":hid,"task":task,"gt":gt,"resp":resp})
    cmap[hid]={"model":model,"condition":cond,"run":run,"claude":codes[model][(cond,run)]}

tpl=TPL.read_text()
labels='''const LABELS = [
 ["C","Produced a correct CHANGELOG.md v1.5.0 entry recording BOTH changes (Keep-a-Changelog form).","1"],
 ["Ic","Produced an entry, but defective: wrong file, malformed / v-prefixed version, or one change missing.","2"],
 ["Ii","Recognized the changelog obligation but produced NO entry block (analysis / recommendation only).","3"],
 ["I","No recognition: concluded it is already done / no changelog obligation. No entry.","4"],
 ["N","Not scoreable: empty, gibberish, off-task, or truncated.","5"],
];'''
tpl=re.sub(r'const LABELS = \[.*?\];', labels, tpl, flags=re.S)
tpl=tpl.replace("Rest-axis calibration","casg-direct scoring calibration")
tpl=tpl.replace("restaxis_calib_","casg_human_anchor_")
tpl=tpl.replace("Ground truth &mdash; correct completion &amp; the over-fire to watch for","Scoring rubric &mdash; what each code means")
tpl=tpl.replace("Ground truth — correct completion &amp; the over-fire to watch for","Scoring rubric — what each code means")
data=json.dumps(items,ensure_ascii=False).replace("</","<\\/")
tpl=tpl.replace("__ITEMS_JSON__",data).replace("__ANSWERS_KEY__",'"v1"')
(OUT/"calibration.html").write_text(tpl)
(OUT/"calib_map.json").write_text(json.dumps(cmap,indent=1))
from collections import Counter
print(f"built {OUT}/calibration.html with {len(items)} blind items ({(OUT/'calibration.html').stat().st_size//1024} KB)")
print("hidden strata:", Counter((m['model'],m['condition']) for m in cmap.values()))
print("hidden claude-code mix:", Counter(m['claude'] for m in cmap.values()))
