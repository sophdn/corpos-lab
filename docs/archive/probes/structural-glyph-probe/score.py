"""
Structural Glyph Probe — scorer (v3)

Side-by-side scoring for multiple models. Reads RESPONSE_{model}_*.md files
and generates a grid with columns for each model.

Scoring is fully manual. Each run is marked by the researcher:
    C — correct behavior: expected behavioral marker is present
    I — incorrect behavior: expected behavioral marker absent or wrong behavior
    N — not scoreable: response malformed, off-task, or uninterpretable

Usage:
    python3 score.py                       # print blank grid to stdout
    python3 score.py --write               # also write SCORE_GRID.md
    python3 score.py --print               # print full response content for review
    python3 score.py --model mistral       # restrict output to one model
    python3 score.py --model claude
"""

import re
import sys
from pathlib import Path

HERE = Path(__file__).parent.resolve()

import json
with open(HERE / "study.json", encoding="utf-8") as f:
    CONFIG = json.load(f)

STUDY_NAME        = CONFIG["name"]
SCENARIOS         = CONFIG["scenarios"]
SCENARIO_ORDER    = CONFIG.get("scenario_order", list(SCENARIOS.keys()))
RUNS_PER_SCENARIO = CONFIG.get("runs_per_scenario", 4)
MODEL_SLUGS       = list(CONFIG.get("models", {}).keys()) or ["mistral"]

# ---------------------------------------------------------------------------
# Filter by --model flag
# ---------------------------------------------------------------------------

_model_filter = None
if "--model" in sys.argv:
    idx = sys.argv.index("--model")
    if idx + 1 < len(sys.argv):
        _model_filter = sys.argv[idx + 1]
        MODEL_SLUGS = [_model_filter] if _model_filter in MODEL_SLUGS else MODEL_SLUGS

# ---------------------------------------------------------------------------
# Response parsing
# ---------------------------------------------------------------------------

def parse_response_header(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    scenario_match  = re.search(r"\*\*Scenario:\*\*\s+(\S+)", text)
    condition_match = re.search(r"\*\*Condition:\*\*\s+(\S+)", text)
    glyph_match     = re.search(r"\*\*Glyph:\*\*\s+(\S+)", text)
    run_match       = re.search(r"\*\*Run:\*\*\s+(\d+)", text)
    model_match     = re.search(r"\*\*Model:\*\*\s+(\S+)", text)
    return {
        "scenario":  scenario_match.group(1)  if scenario_match  else None,
        "condition": condition_match.group(1) if condition_match else None,
        "glyph":     glyph_match.group(1)     if glyph_match     else None,
        "run":       int(run_match.group(1))   if run_match       else None,
        "model":     model_match.group(1)      if model_match     else None,
        "text":      text,
    }


def response_body(text: str) -> str:
    match = re.search(r"^---\s*\n+(.*)", text, re.DOTALL | re.MULTILINE)
    return match.group(1).strip() if match else text.strip()


# ---------------------------------------------------------------------------
# Load all responses, keyed by (model, scenario, run)
# ---------------------------------------------------------------------------

def load_results() -> dict:
    """Returns dict: {model_slug: {scenario: {run_n: parsed}}}"""
    results = {}
    for path in sorted(HERE.glob("RESPONSE_*.md")):
        parsed   = parse_response_header(path)
        scenario = parsed["scenario"]
        n        = parsed["run"]
        model    = parsed["model"]
        if scenario and n and model:
            results.setdefault(model, {}).setdefault(scenario, {})[n] = parsed
    return results


# ---------------------------------------------------------------------------
# Output
# ---------------------------------------------------------------------------

def main() -> None:
    write_file   = "--write" in sys.argv
    print_bodies = "--print" in sys.argv

    results = load_results()

    # Print full responses if requested
    if print_bodies:
        for model_slug in MODEL_SLUGS:
            model_results = results.get(model_slug, {})
            for scenario in SCENARIO_ORDER:
                if scenario not in model_results:
                    continue
                expected = SCENARIOS[scenario].get("expected_behavior", "")
                print(f"\n{'='*70}")
                print(f"MODEL: {model_slug}  |  SCENARIO: {scenario}  |  condition: {SCENARIOS[scenario].get('condition', '')}  |  glyph: {SCENARIOS[scenario].get('glyph_code') or 'none'}")
                print(f"Expected: {expected}")
                print('='*70)
                run_data = model_results[scenario]
                for n in range(1, RUNS_PER_SCENARIO + 1):
                    parsed = run_data.get(n)
                    if parsed:
                        print(f"\n--- Run {n} ---")
                        print(response_body(parsed["text"]))
                    else:
                        print(f"\n--- Run {n} --- [missing]")
        print()

    # Build side-by-side score grid
    # Columns: Scenario | Condition | Glyph | [M: R1..R4 | Score] | [C: R1..R4 | Score]
    run_cols  = "".join(f" R{n} |" for n in range(1, RUNS_PER_SCENARIO + 1))

    model_header_parts  = " | ".join(f"{m}: {run_cols} Score" for m in MODEL_SLUGS)
    header    = f"| Scenario | Condition | Glyph | {model_header_parts} |"

    sep_run   = "".join(" --- |" for _ in range(RUNS_PER_SCENARIO))
    model_sep_parts = " | ".join(f"{sep_run} ---" for _ in MODEL_SLUGS)
    subheader = f"|----------|-----------|-------|{model_sep_parts}|"

    rows = []
    for scenario in SCENARIO_ORDER:
        condition = SCENARIOS[scenario].get("condition", "")
        glyph     = SCENARIOS[scenario].get("glyph_code") or "none"

        model_cells = []
        for model_slug in MODEL_SLUGS:
            run_data = results.get(model_slug, {}).get(scenario, {})
            run_cells = ""
            for n in range(1, RUNS_PER_SCENARIO + 1):
                cell = "·" if run_data.get(n) else "?"
                run_cells += f" {cell} |"
            model_cells.append(f"{run_cells} ·/{RUNS_PER_SCENARIO}")

        row_models = " | ".join(model_cells)
        rows.append(f"| {scenario} | {condition} | {glyph} | {row_models} |")

    expected_block = []
    for scenario in SCENARIO_ORDER:
        expected = SCENARIOS[scenario].get("expected_behavior", "")
        if expected:
            expected_block.append(f"- **{scenario}:** {expected}")

    lines = [
        f"# Score Grid — {STUDY_NAME}",
        "",
        "*Scoring is manual. Replace · with C (correct), I (incorrect), or N (not scoreable).*",
        "*Score column: fill in as N_correct/N_total after marking runs.*",
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
        "*Per run per model: what did the model actually do? Not just pass/fail.*",
        "",
        "| Model | Scenario | Run | Observation |",
        "|-------|----------|-----|-------------|",
        *[
            f"| {model_slug} | {scenario} | R{n} | · |"
            for model_slug in MODEL_SLUGS
            for scenario in SCENARIO_ORDER
            for n in range(1, RUNS_PER_SCENARIO + 1)
        ],
    ]

    output = "\n".join(lines) + "\n"
    print(output)

    if write_file:
        out_path = HERE / "SCORE_GRID.md"
        out_path.write_text(output, encoding="utf-8")
        print(f"Written to {out_path.name}")


if __name__ == "__main__":
    main()
