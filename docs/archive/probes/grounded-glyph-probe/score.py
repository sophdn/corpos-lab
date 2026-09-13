"""
Grounded Glyph Probe — scorer

Scoring is fully manual. This script produces a blank score grid for the
researcher to fill in, and optionally prints response content for review.

Each run is marked by the researcher:
    C  — correct action taken per the glyph's acceptance criterion for this study.
         C means the same thing across all conditions: the agent executed the correct
         action. What counts as C is defined once per study and applied uniformly.
    Ii — recognition present, action absent or deferred — agent identifies the
         obligation or describes the correct path but does not execute it. Informal
         mentions, verbal descriptions, and recommendations that name the right thing
         without doing it are Ii, not C.
    Ic — recognition present, wrong action (action-requiring glyphs only)
    I  — no recognition
    N  — not scoreable: response malformed, off-task, or uninterpretable

Ii and Ic do not count toward the C rate. Record them in Behavioral Observations
with the Ii / Ic marker so the pattern is visible.

Usage:
    python3 score.py              # print blank score grid to stdout
    python3 score.py --write      # also write SCORE_GRID.md
    python3 score.py --print      # print full response content for review
"""

import re
import sys
import json
from pathlib import Path

HERE = Path(__file__).parent.resolve()

with open(HERE / "study.json", encoding="utf-8") as f:
    CONFIG = json.load(f)

STUDY_NAME         = CONFIG["name"]
CONDITIONS         = CONFIG["conditions"]
CONDITION_ORDER    = CONFIG.get("condition_order", list(CONDITIONS.keys()))
RUNS_PER_CONDITION = CONFIG.get("runs_per_condition", 8)
INDUCTION_VECTOR   = CONFIG.get("induction_vector", "(not specified)")
MODEL_ORDER        = list(CONFIG["model_configs"].keys())


def parse_response_header(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    condition_match = re.search(r"\*\*Condition:\*\*\s+(\S+)", text)
    model_match     = re.search(r"\*\*Model:\*\*\s+(\S+)", text)
    run_match       = re.search(r"\*\*Run:\*\*\s+(\d+)", text)
    return {
        "condition": condition_match.group(1) if condition_match else None,
        "model":     model_match.group(1)     if model_match     else None,
        "run":       int(run_match.group(1))  if run_match       else None,
        "text":      text,
    }


def response_body(text: str) -> str:
    match = re.search(r"^---\s*\n+(.*)", text, re.DOTALL | re.MULTILINE)
    return match.group(1).strip() if match else text.strip()


def main() -> None:
    # Force UTF-8 output so response content with unicode (arrows, checkmarks, etc.)
    # does not raise UnicodeEncodeError on Windows cp1252 terminals.
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except AttributeError:
        pass  # reconfigure not available (e.g. stdout already redirected)

    write_file   = "--write" in sys.argv
    print_bodies = "--print" in sys.argv

    # Collect all response files
    # results: condition → model → run_n → parsed
    results: dict[str, dict[str, dict[int, dict]]] = {}
    for path in sorted(HERE.glob("RESPONSE_*.md")):
        parsed    = parse_response_header(path)
        condition = parsed["condition"]
        model     = parsed["model"]
        n         = parsed["run"]
        if condition and model and n:
            results.setdefault(condition, {}).setdefault(model, {})[n] = parsed

    if print_bodies:
        for condition_code in CONDITION_ORDER:
            for model_slug in MODEL_ORDER:
                run_data = results.get(condition_code, {}).get(model_slug, {})
                if not run_data:
                    continue
                expected = CONDITIONS[condition_code].get("expected_behavior", "")
                print(f"\n{'='*70}")
                print(f"CONDITION: {condition_code}  |  MODEL: {model_slug}")
                print(f"Expected: {expected}")
                print('='*70)
                for n in range(1, RUNS_PER_CONDITION + 1):
                    parsed = run_data.get(n)
                    if parsed:
                        print(f"\n--- Run {n} ---")
                        print(response_body(parsed["text"]))
                    else:
                        print(f"\n--- Run {n} --- [missing]")
        print()

    run_cols  = "".join(f" R{n} |" for n in range(1, RUNS_PER_CONDITION + 1))
    separator = "".join(" --- |" for _ in range(RUNS_PER_CONDITION))

    header    = f"| Model | Condition |{run_cols} Score |"
    subheader = f"|-------|-----------|{separator}-------|"

    rows = []
    for condition_code in CONDITION_ORDER:
        for model_slug in MODEL_ORDER:
            run_data  = results.get(condition_code, {}).get(model_slug, {})
            run_cells = ""
            for n in range(1, RUNS_PER_CONDITION + 1):
                cell = "·" if run_data.get(n) else "?"
                run_cells += f" {cell} |"
            c_count = f"·/{RUNS_PER_CONDITION}"
            rows.append(f"| {model_slug} | {condition_code} |{run_cells} {c_count} |")

    expected_block = []
    for condition_code in CONDITION_ORDER:
        expected = CONDITIONS[condition_code].get("expected_behavior", "")
        if expected:
            expected_block.append(f"- **{condition_code}:** {expected}")

    lines = [
        f"# Score Grid — {STUDY_NAME}",
        "",
        f"**Induction vector:** {INDUCTION_VECTOR}",
        "",
        "*Scoring is manual. Replace · with C, Ii, Ic, I, or N. Replace ? with score if run exists. Score column: fill in as N_C/8 after marking all runs.*",
        "",
        "**Score codes:** C = recognition + correct action | Ii = recognition, no action | Ic = recognition, wrong action | I = no recognition | N = not scoreable",
        "",
        header,
        subheader,
        *rows,
        "",
        "## Expected behavior (from study.json)",
        "",
        *expected_block,
        "",
        "## Behavioral observations",
        "",
        "*Per run: what did the model do and what was in its context. Position language throughout — no pass/fail. Note C, Ii, Ic, I runs specifically.*",
        "",
        "| Model | Condition | Run | Score | Observation |",
        "|-------|-----------|-----|-------|-------------|",
        *[
            f"| {model_slug} | {condition_code} | R{n} | · | |"
            for condition_code in CONDITION_ORDER
            for model_slug in MODEL_ORDER
            for n in range(1, RUNS_PER_CONDITION + 1)
        ],
        "",
        "## Ii summary",
        "",
        "*List any Ii runs here. What did the agent recognize? What action was absent or deferred?*",
        "",
        "*(none)*",
        "",
        "## Gate results",
        "",
        "**Condition 1 calibration (baseline must not meet ≥7/8 C — high baseline C rate means scenario is over-specified):**",
        "",
        *[f"- {model_slug}: · — [proceed / over-specified — redesign]" for model_slug in MODEL_ORDER],
        "",
        "**Condition 2 sufficiency gate (≥7/8 C = glyph sufficient, skip condition 3):**",
        "",
        *[f"- {model_slug}: · — [sufficient / ground triggered]" for model_slug in MODEL_ORDER],
        "",
        "**Condition 3 ground lift (≥5/8 C and ≥2 above condition 2 — only if condition 3 run):**",
        "",
        *[f"- {model_slug}: · — [met / not met / not run]" for model_slug in MODEL_ORDER],
    ]

    output = "\n".join(lines) + "\n"

    if write_file:
        out_path = HERE / "SCORE_GRID.md"
        if out_path.exists():
            print(
                f"SCORE_GRID.md already exists — not overwriting. "
                "Delete it first if you need a fresh scaffold."
            )
        else:
            out_path.write_text(output, encoding="utf-8")
            print(f"Written to {out_path.name}")

    print(output.encode(sys.stdout.encoding or "utf-8", errors="replace").decode(
        sys.stdout.encoding or "utf-8"
    ))


if __name__ == "__main__":
    main()
