"""

DEPRECATED (e9-typed-study-runners, 2026-04-19): Rust port of this
probe lives at lab-app/crates/lab-app-server/src/studies/. See
blueprints/probes/MIGRATION.md for the Rust invocation equivalent of each
python3 run.py ... call. This file is retained as reference documentation
until at least one full assay series runs through the typed runners without
issue.

Glyph Behavioral Probe — runner

Places the model inside a scenario as an acting subject. Loads the glyph
(if present) as context before the task. Records the model's free response.

The model does not render a verdict. It performs the task. Scoring is manual.

Required files in the study directory:
    study.json              — study configuration (see SETUP.md)
    SCENARIO_{code}.md      — one per distinct scenario_file value in study.json
    GLYPH_{code}.md         — one per glyph_code value in study.json (omit for control)

Usage:
    python3 run.py                          # all runs in scenario_order
    python3 run.py example-control          # one scenario, all runs
    python3 run.py example-control 2        # one scenario, run 2 only
    python3 run.py example-control example-treatment   # multiple scenarios
"""

import sys
import json
import urllib.request
from pathlib import Path
from datetime import date

HERE = Path(__file__).parent.resolve()

# ---------------------------------------------------------------------------
# Load study config
# ---------------------------------------------------------------------------

with open(HERE / "study.json", encoding="utf-8") as f:
    CONFIG = json.load(f)

STUDY_NAME        = CONFIG["name"]
OLLAMA_URL        = CONFIG.get("endpoint", "http://localhost:11434/api/generate")
MODEL             = CONFIG.get("model", "mistral:latest")
TIMEOUT           = CONFIG.get("timeout", 300)
RUNS_PER_SCENARIO = CONFIG.get("runs_per_scenario", 4)
SCENARIOS         = CONFIG["scenarios"]
SCENARIO_ORDER    = CONFIG.get("scenario_order", list(SCENARIOS.keys()))

# ---------------------------------------------------------------------------
# File loaders
# ---------------------------------------------------------------------------

def load_scenario(scenario_file: str) -> str:
    path = HERE / f"SCENARIO_{scenario_file}.md"
    if not path.exists():
        raise FileNotFoundError(f"Scenario file not found: {path.name}")
    return path.read_text(encoding="utf-8")


def load_glyph(glyph_code: str) -> str:
    path = HERE / f"GLYPH_{glyph_code}.md"
    if not path.exists():
        raise FileNotFoundError(f"Glyph file not found: {path.name}")
    return path.read_text(encoding="utf-8")


# ---------------------------------------------------------------------------
# Prompt assembly
# ---------------------------------------------------------------------------

RESPONSE_HEADER = """\
# RESPONSE_{scenario}_{n}

**Study:** {study_name}
**Scenario:** {scenario}
**Condition:** {condition}
**Glyph:** {glyph}
**Run:** {n} of {runs_per_scenario}
**Model:** {model}
**Date:** {date}

---

"""


def build_prompt(scenario_code: str) -> str:
    scenario     = SCENARIOS[scenario_code]
    scenario_file = scenario["scenario_file"]
    glyph_code   = scenario.get("glyph_code")

    parts = []

    if glyph_code:
        parts.append(load_glyph(glyph_code))
        parts.append("---")

    parts.append(load_scenario(scenario_file))

    return "\n\n".join(parts)


# ---------------------------------------------------------------------------
# Model call
# ---------------------------------------------------------------------------

def call_model(prompt: str) -> str:
    raise RuntimeError(
        "Ollama is RETIRED and uninstalled (2026-07-14) - do NOT install it. "
        "llama.cpp (llama-server, http://localhost:8081, OpenAI-compatible) is the "
        "only local inference portal; port this call to it. "
        "See memory one-local-inference-portal-llama-cpp."
    )
    payload = json.dumps({
        "model":  MODEL,
        "prompt": prompt,
        "stream": False,
    }).encode("utf-8")
    req = urllib.request.Request(
        OLLAMA_URL,
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
        return json.loads(resp.read().decode("utf-8"))["response"]


# ---------------------------------------------------------------------------
# Run
# ---------------------------------------------------------------------------

def run_single(scenario_code: str, n: int) -> None:
    scenario  = SCENARIOS[scenario_code]
    out_path  = HERE / f"RESPONSE_{scenario_code}_{n}.md"

    if out_path.exists():
        print(f"{scenario_code} run {n}: already exists — skipping.", flush=True)
        return

    print(f"{scenario_code} run {n}: building prompt...", flush=True)
    prompt = build_prompt(scenario_code)

    print(f"{scenario_code} run {n}: sending to {MODEL}...", flush=True)
    response = call_model(prompt)

    glyph_code = scenario.get("glyph_code") or "none"
    header = RESPONSE_HEADER.format(
        scenario=scenario_code,
        n=n,
        study_name=STUDY_NAME,
        condition=scenario.get("condition", ""),
        glyph=glyph_code,
        runs_per_scenario=RUNS_PER_SCENARIO,
        model=MODEL,
        date=date.today().isoformat(),
    )
    out_path.write_text(header + response, encoding="utf-8")
    print(f"{scenario_code} run {n}: written to {out_path.name}", flush=True)


# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

def parse_args(argv: list) -> list:
    """
    Returns a list of (scenario_code, run_number) pairs.

    Forms accepted:
        (empty)                    → all runs in scenario_order
        example-control            → all runs for that scenario
        example-control 2          → run 2 only
        example-control example-treatment   → both scenarios, all runs
        example-control 3 example-treatment → first scenario run 3, then all of second
    """
    if not argv:
        return [
            (code, n)
            for code in SCENARIO_ORDER
            for n in range(1, RUNS_PER_SCENARIO + 1)
        ]

    jobs = []
    i = 0
    while i < len(argv):
        token = argv[i]
        if token in SCENARIOS:
            if i + 1 < len(argv) and argv[i + 1].isdigit():
                run_n = int(argv[i + 1])
                if 1 <= run_n <= RUNS_PER_SCENARIO:
                    jobs.append((token, run_n))
                else:
                    print(f"Run number {run_n} out of range (1–{RUNS_PER_SCENARIO}) — skipping")
                i += 2
            else:
                jobs.extend((token, n) for n in range(1, RUNS_PER_SCENARIO + 1))
                i += 1
        else:
            print(f"Unknown scenario '{token}' — skipping")
            i += 1
    return jobs


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    jobs = parse_args(sys.argv[1:])
    if not jobs:
        print("No valid jobs to run.")
        return

    total = len(jobs)
    print(f"Study: {STUDY_NAME}")
    print(f"Running {total} job(s) sequentially.\n")
    for scenario_code, n in jobs:
        run_single(scenario_code, n)

    print("\nDone.")


if __name__ == "__main__":
    main()
