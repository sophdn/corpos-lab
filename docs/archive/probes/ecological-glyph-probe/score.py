"""
Ecological Glyph Probe — scorer

Scoring is fully manual. This script produces a blank score grid for the
researcher to fill in, and optionally prints response content for review.

Each run is marked by the researcher:
    C  — correct behavior: explicit companion artifact update within the operation
    I  — incorrect behavior: companion update absent or wrong behavior occurred
    Ii — implicit hit: model represents updated companion state but does not
         explicitly update the file. Score as I in the grid; note separately.
    N  — not scoreable: response malformed, off-task, or uninterpretable

Implicit hits (Ii) do not count toward the treatment rate. Record them in
Behavioral Observations with the Ii marker so the pattern is visible.

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

STUDY_NAME        = CONFIG["name"]
SCENARIOS         = CONFIG["scenarios"]
SCENARIO_ORDER    = CONFIG.get("scenario_order", list(SCENARIOS.keys()))
RUNS_PER_SCENARIO = CONFIG.get("runs_per_scenario", 8)
INDUCTION_VECTOR  = CONFIG.get("induction_vector", "(not specified)")


def parse_response_header(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    scenario_match  = re.search(r"\*\*Scenario:\*\*\s+(\S+)", text)
    condition_match = re.search(r"\*\*Condition:\*\*\s+(\S+)", text)
    glyph_match     = re.search(r"\*\*Glyph:\*\*\s+(\S+)", text)
    run_match       = re.search(r"\*\*Run:\*\*\s+(\d+)", text)
    return {
        "scenario":  scenario_match.group(1)  if scenario_match  else None,
        "condition": condition_match.group(1) if condition_match else None,
        "glyph":     glyph_match.group(1)     if glyph_match     else None,
        "run":       int(run_match.group(1))   if run_match       else None,
        "text":      text,
    }


def response_body(text: str) -> str:
    match = re.search(r"^---\s*\n+(.*)", text, re.DOTALL | re.MULTILINE)
    return match.group(1).strip() if match else text.strip()


def main() -> None:
    write_file   = "--write" in sys.argv
    print_bodies = "--print" in sys.argv

    results: dict[str, dict[int, dict]] = {}
    for path in sorted(HERE.glob("RESPONSE_*.md")):
        parsed   = parse_response_header(path)
        scenario = parsed["scenario"]
        n        = parsed["run"]
        if scenario and n:
            results.setdefault(scenario, {})[n] = parsed

    if print_bodies:
        for scenario in SCENARIO_ORDER:
            if scenario not in results:
                continue
            expected = SCENARIOS[scenario].get("expected_behavior", "")
            print(f"\n{'='*70}")
            print(f"SCENARIO: {scenario}  |  condition: {SCENARIOS[scenario].get('condition', '')}  |  glyph: {SCENARIOS[scenario].get('glyph_code') or 'none'}")
            print(f"Expected: {expected}")
            print('='*70)
            run_data = results[scenario]
            for n in range(1, RUNS_PER_SCENARIO + 1):
                parsed = run_data.get(n)
                if parsed:
                    print(f"\n--- Run {n} ---")
                    print(response_body(parsed["text"]))
                else:
                    print(f"\n--- Run {n} --- [missing]")
        print()

    run_cols  = "".join(f" R{n} |" for n in range(1, RUNS_PER_SCENARIO + 1))
    separator = "".join(" --- |" for _ in range(RUNS_PER_SCENARIO))

    header    = f"| Scenario | Condition | Glyph |{run_cols} Score |"
    subheader = f"|----------|-----------|-------|{separator}-------|"

    rows = []
    for scenario in SCENARIO_ORDER:
        condition = SCENARIOS[scenario].get("condition", "")
        glyph     = SCENARIOS[scenario].get("glyph_code") or "none"
        run_data  = results.get(scenario, {})

        run_cells = ""
        for n in range(1, RUNS_PER_SCENARIO + 1):
            cell = "·" if run_data.get(n) else "?"
            run_cells += f" {cell} |"

        c_count = f"·/{RUNS_PER_SCENARIO}"
        rows.append(f"| {scenario} | {condition} | {glyph} |{run_cells} {c_count} |")

    expected_block = []
    for scenario in SCENARIO_ORDER:
        expected = SCENARIOS[scenario].get("expected_behavior", "")
        if expected:
            expected_block.append(f"- **{scenario}:** {expected}")

    lines = [
        f"# Score Grid — {STUDY_NAME}",
        "",
        f"**Induction vector:** {INDUCTION_VECTOR}",
        "",
        "*Scoring is manual. Replace · with C (correct), I (incorrect), Ii (implicit hit — score as I), or N (not scoreable).*",
        "*Score column: fill in as N_correct/N_total after marking runs.*",
        "*Implicit hits (Ii): model represents updated companion state without explicit file update. Count as I in Score column; note in Behavioral Observations.*",
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
        "*Per run: what did the model actually do? Note explicit updates (C), omissions (I), and implicit hits (Ii — chain state shown but no file update). Be specific about which artifact was or was not acted on.*",
        "",
        "| Scenario | Run | Score | Observation |",
        "|----------|-----|-------|-------------|",
        *[
            f"| {scenario} | R{n} | · | |"
            for scenario in SCENARIO_ORDER
            for n in range(1, RUNS_PER_SCENARIO + 1)
        ],
        "",
        "## Implicit hit summary",
        "",
        "*List any Ii runs here. Note what updated state was shown and what action was absent.*",
        "",
        "*(none)*",
    ]

    output = "\n".join(lines) + "\n"

    if write_file:
        out_path = HERE / "SCORE_GRID.md"
        out_path.write_text(output, encoding="utf-8")
        print(f"Written to {out_path.name}")

    print(output.encode(sys.stdout.encoding or "utf-8", errors="replace").decode(sys.stdout.encoding or "utf-8"))


if __name__ == "__main__":
    main()
