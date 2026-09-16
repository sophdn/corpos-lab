#!/usr/bin/env python3
"""Generate a self-contained, local, blind human-scoring web app.

This is the reusable generator promoted from the rest-axis one-off
(suggestion local-blind-scoring-app-for-human-labeling). Use it whenever a study
needs the human to label, score, or adjudicate N items — calibration subsets,
rater-disagreement adjudication, neutrality spot-checks.

Properties it guarantees:
  * LOCAL only — the app is a single HTML file opened from file://, no CDN, no
    build step. Private research materials never leave the machine.
  * BLIND BY CONSTRUCTION — an embedded item may carry only {id, blocks}; a block
    may carry only {heading, text, kind}. Any other key (arm, label, judge,
    machine, model, seed, ...) is a hard error, so the arm and the machine
    raters' labels cannot leak into what the human sees. Those stay in a held key
    file the generator never touches.
  * HOME by default — the output is written under the user's home directory, not
    the /tmp scratchpad, because a snap-confined browser (Ubuntu's default
    Firefox) cannot read files under /tmp.

The COORDINATOR stays out of the terrain: a study-specific adapter reads the
responses and assembles the blind items; the agent invoking this generator only
handles ids and labels.

Item shape (blind):
    {"id": "C00",
     "blocks": [
        {"heading": "Task given to the assistant", "text": "...", "kind": "context"},
        {"heading": "Correct completion",            "text": "...", "kind": "ground"},
        {"heading": "Response to score",             "text": "...", "kind": "response"}
     ]}

Label shape:
    {"code": "OK", "desc": "Did the task. No unwarranted apparatus.", "key": "1"}

Usage:
    build_app.py --items items.json --labels labels.json \
        --title "Rest-axis calibration" --store-key rest_axis_v1 [--out PATH]

    --items / --labels take JSON files (or "-" for stdin on --items).
    --out defaults to ~/blind-scoring/<store-key>.html.
"""
import argparse
import json
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
TEMPLATE = HERE / "app_template.html"

ALLOWED_ITEM_KEYS = {"id", "blocks"}
ALLOWED_BLOCK_KEYS = {"heading", "text", "kind"}
ALLOWED_KINDS = {"context", "ground", "response"}


class BlindnessError(ValueError):
    """An item carried a field the human must not see."""


def validate_items(items):
    """Enforce blindness structurally: reject any field beyond {id, blocks} and
    {heading, text, kind}. This is the load-bearing guard — it is what keeps the
    arm and the machine raters' labels out of the app by construction."""
    if not isinstance(items, list) or not items:
        raise ValueError("items must be a non-empty JSON array")
    seen = set()
    for it in items:
        if not isinstance(it, dict):
            raise ValueError(f"each item must be an object, got {type(it).__name__}")
        extra = set(it) - ALLOWED_ITEM_KEYS
        if extra:
            raise BlindnessError(
                f"item {it.get('id', '?')!r} carries forbidden field(s) {sorted(extra)} — "
                f"an item may hold only {sorted(ALLOWED_ITEM_KEYS)}; the arm and any machine "
                f"labels stay in a held key file, never in the app")
        if "id" not in it or not isinstance(it["id"], str) or not it["id"]:
            raise ValueError(f"item is missing a non-empty string id: {it!r}")
        if it["id"] in seen:
            raise ValueError(f"duplicate item id {it['id']!r}")
        seen.add(it["id"])
        blocks = it.get("blocks")
        if not isinstance(blocks, list) or not blocks:
            raise ValueError(f"item {it['id']!r} must have a non-empty blocks array")
        for b in blocks:
            if not isinstance(b, dict):
                raise ValueError(f"item {it['id']!r}: each block must be an object")
            bextra = set(b) - ALLOWED_BLOCK_KEYS
            if bextra:
                raise BlindnessError(
                    f"item {it['id']!r}: block carries forbidden field(s) {sorted(bextra)} — "
                    f"a block may hold only {sorted(ALLOWED_BLOCK_KEYS)}")
            kind = b.get("kind", "context")
            if kind not in ALLOWED_KINDS:
                raise ValueError(
                    f"item {it['id']!r}: block kind {kind!r} not one of {sorted(ALLOWED_KINDS)}")


def validate_labels(labels):
    if not isinstance(labels, list) or not labels:
        raise ValueError("labels must be a non-empty JSON array")
    for lab in labels:
        if not isinstance(lab, dict) or "code" not in lab or "desc" not in lab:
            raise ValueError(f"each label needs at least code and desc: {lab!r}")


def build_app(items, labels, title, store_key, out_path):
    """Assemble the app HTML and write it to out_path. Returns out_path."""
    validate_items(items)
    validate_labels(labels)
    template = TEMPLATE.read_text()

    def inline_json(obj):
        # Keep an inline <script> safe: a literal "</script>" in the data would
        # close the tag early.
        return json.dumps(obj, ensure_ascii=False).replace("</", "<\\/")

    html = (template
            .replace("__TITLE__", title)
            .replace("__ITEMS_JSON__", inline_json(items))
            .replace("__LABELS_JSON__", inline_json(labels))
            .replace("__STORE_KEY__", json.dumps(store_key)))
    out_path.parent.mkdir(parents=True, exist_ok=True)
    out_path.write_text(html)
    return out_path


def _load_json(spec):
    if spec == "-":
        return json.load(sys.stdin)
    return json.loads(Path(spec).read_text())


def main(argv=None):
    ap = argparse.ArgumentParser(description="Generate a blind human-scoring web app.")
    ap.add_argument("--items", required=True, help="JSON file of blind items (or - for stdin)")
    ap.add_argument("--labels", required=True, help="JSON file of the label set")
    ap.add_argument("--title", required=True, help="app title")
    ap.add_argument("--store-key", required=True, help="localStorage namespace (e.g. rest_axis_v1)")
    ap.add_argument("--out", help="output HTML path (default ~/blind-scoring/<store-key>.html)")
    args = ap.parse_args(argv)

    items = _load_json(args.items)
    labels = _load_json(args.labels)
    out_path = Path(args.out).expanduser() if args.out else Path.home() / "blind-scoring" / f"{args.store_key}.html"

    try:
        build_app(items, labels, args.title, args.store_key, out_path)
    except (BlindnessError, ValueError) as e:
        print(f"build_app: {e}", file=sys.stderr)
        return 1

    # Snap-confine gotcha: a snap browser cannot read /tmp.
    if str(out_path).startswith("/tmp"):
        print(f"build_app: WARNING {out_path} is under /tmp — a snap-confined browser "
              f"(Ubuntu's default Firefox) cannot open it; write under $HOME instead", file=sys.stderr)
    print(f"build_app: wrote {out_path} — {len(items)} blind items, {len(labels)} labels "
          f"({out_path.stat().st_size // 1024} KB). Open it with: xdg-open {out_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
