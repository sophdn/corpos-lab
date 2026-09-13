"""
Multiple Blind Probe Study — scorer

Reads all RESPONSE_{scenario}_{n}.md files in the study directory.
Auto-populates: Verdict (C/I/N) and Field Source cited (last verdict block per run).
Flags for manual review: Observable quality (C/P/I) and Evidence quality (C/P/I).

Usage:
    python3 score.py              # print score grid to stdout
    python3 score.py --write      # also write SCORE_GRID.md to the study directory
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


def parse_response(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")

    def extract_last_field(label: str) -> str | None:
        matches = re.findall(rf"^\s*{re.escape(label)}:\s*(.+)$", text, re.MULTILINE | re.IGNORECASE)
        return matches[-1].strip() if matches else None

    scenario_match = re.search(r"\*\*Scenario:\*\*\s+([^\s(]+)", text)
    gt_match       = re.search(r"\*\*Ground truth:\*\*\s+(\w+)", text)
    run_match      = re.search(r"\*\*Run:\*\*\s+(\d+)", text)

    return {
        "scenario":     scenario_match.group(1) if scenario_match else None,
        "ground_truth": gt_match.group(1).lower() if gt_match else None,
        "run":          int(run_match.group(1)) if run_match else None,
        "verdict":      extract_last_field("VERDICT"),
        "field_source": extract_last_field("FIELD SOURCE"),
        "observable":   extract_last_field("OBSERVABLE"),
        "evidence":     extract_last_field("EVIDENCE"),
    }


def score_verdict(verdict: str | None, ground_truth: str | None) -> str:
    if verdict is None or ground_truth is None:
        return "N"
    v = verdict.lower().strip()
    if v not in ("yes", "no"):
        return "N"
    return "C" if v == ground_truth.lower() else "I"


def truncate(s: str | None, n: int = 28) -> str:
    if s is None:
        return "—"
    return s if len(s) <= n else s[:n - 1] + "…"


def main() -> None:
    write_file = "--write" in sys.argv

    # Collect all response files
    results: dict[str, dict[int, dict]] = {}
    for path in sorted(HERE.glob("RESPONSE_*.md")):
        parsed = parse_response(path)
        scenario = parsed["scenario"]
        n        = parsed["run"]
        if scenario and n:
            results.setdefault(scenario, {})[n] = parsed

    # Build header row
    run_cols  = "".join(f" R{n}-V | R{n}-FS |" for n in range(1, RUNS_PER_SCENARIO + 1))
    separator = "".join(" ------ | --- |" for _ in range(RUNS_PER_SCENARIO))

    header    = f"| Scenario | GT |{run_cols} Verdict | O | E |"
    subheader = f"|----------|----|{separator}---------|---|---|"

    rows = []
    for scenario in SCENARIO_ORDER:
        if scenario not in results:
            continue

        gt       = SCENARIOS[scenario]["ground_truth"]
        run_data = results[scenario]

        verdict_scores = []
        run_cells      = ""

        for n in range(1, RUNS_PER_SCENARIO + 1):
            parsed = run_data.get(n)
            if parsed:
                v_score = score_verdict(parsed["verdict"], parsed["ground_truth"])
                fs      = truncate(parsed["field_source"])
            else:
                v_score = "?"
                fs      = "?"

            run_cells += f" {v_score} | {fs} |"
            verdict_scores.append(v_score)

        # Aggregate verdict
        c_count = verdict_scores.count("C")
        agg_v   = f"{c_count}/{RUNS_PER_SCENARIO}"

        # Observable and Evidence are manual — leave blank for Researcher to fill
        rows.append(f"| {scenario} | {gt} |{run_cells} {agg_v} | · | · |")

    lines = [
        f"# Score Grid — {STUDY_NAME}",
        "",
        "*Auto-populated: Verdict (C/I/N) and Field Source (FS) — last verdict block per run. "
        "Fill in Observable (O) and Evidence (E) quality (C/P/I) after reading response files.*",
        "",
        header,
        subheader,
        *rows,
        "",
        "## Field source key",
        "",
        "FS values come directly from the model's last FIELD SOURCE response per run.",
    ]

    output = "\n".join(lines) + "\n"
    print(output)

    if write_file:
        out_path = HERE / "SCORE_GRID.md"
        out_path.write_text(output, encoding="utf-8")
        print(f"Written to {out_path.name}")


if __name__ == "__main__":
    main()
