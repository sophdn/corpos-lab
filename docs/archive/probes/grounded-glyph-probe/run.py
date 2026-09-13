"""

DEPRECATED (e9-typed-study-runners, 2026-04-19): Rust port of this
probe lives at lab-app/crates/lab-app-server/src/studies/. See
blueprints/probes/MIGRATION.md for the Rust invocation equivalent of each
python3 run.py ... call. This file is retained as reference documentation
until at least one full assay series runs through the typed runners without
issue.

Grounded Glyph Probe — runner (v5)

Tests whether domain-specific ground closes the execution gap that a universal
glyph alone leaves open. Three conditions: baseline (no aid), glyph-only
(universal glyph), grounded-glyph (glyph + domain ground). Condition 3 is
conditional on condition 2 result — only run if warranted.

Supports two backends:
    mistral  — POST to Ollama endpoint (local). Glyph (and ground if present)
               prepended to scenario content, separated by ---.
    claude   — subprocess call to claude -p from blanky. Glyph (and ground if
               present) delivered via --system-prompt. Scenario via -p.
               The --system-prompt channel is required for Claude: prepending the
               glyph to -p triggers safety heuristics (injection recognition failure).

Required files in the study directory:
    study.json                 — study configuration (see SETUP.md)
    SCENARIO_{code}_claude.md  — workflow-format scenario for Claude
    SCENARIO_{code}_mistral.md — flat-format scenario for Mistral
    GLYPH_{code}.md            — universal glyph (conditions 2 and 3 only)
    GROUND_{code}.md           — ground register (condition 3 only, if warranted)

Usage:
    python3 run.py --model mistral                       # all conditions, all runs
    python3 run.py --model claude                        # all conditions, all runs
    python3 run.py --model mistral baseline              # one condition, all runs
    python3 run.py --model claude glyph_only 3           # one condition, one run
"""

import sys
import json
import shutil
import subprocess
import glob as _glob
import urllib.request
from pathlib import Path
from datetime import date

HERE = Path(__file__).parent.resolve()

# ---------------------------------------------------------------------------
# Parse --model flag (must be first argument)
# ---------------------------------------------------------------------------

def extract_model_flag(argv: list) -> tuple[str, list]:
    if "--model" in argv:
        idx = argv.index("--model")
        if idx + 1 >= len(argv):
            print("Error: --model requires an argument (mistral or claude)")
            sys.exit(1)
        slug = argv[idx + 1]
        remaining = argv[:idx] + argv[idx + 2:]
        return slug, remaining
    return "mistral", argv

MODEL_SLUG, _remaining_argv = extract_model_flag(sys.argv[1:])

# ---------------------------------------------------------------------------
# Load study config
# ---------------------------------------------------------------------------

with open(HERE / "study.json", encoding="utf-8") as f:
    CONFIG = json.load(f)

STUDY_NAME         = CONFIG["name"]
SCENARIO_CODE      = CONFIG["scenario_code"]
RUNS_PER_CONDITION = CONFIG.get("runs_per_condition", 8)
CONDITIONS         = CONFIG["conditions"]
CONDITION_ORDER    = CONFIG.get("condition_order", list(CONDITIONS.keys()))
INDUCTION_VECTOR   = CONFIG.get("induction_vector", "(not specified)")
TIMEOUT            = CONFIG.get("timeout", 300)

# Backend config
if MODEL_SLUG not in CONFIG["model_configs"]:
    available = list(CONFIG["model_configs"].keys())
    print(f"Error: model '{MODEL_SLUG}' not found in study.json 'model_configs'. Available: {available}")
    sys.exit(1)

MODEL_CONFIG = CONFIG["model_configs"][MODEL_SLUG]
DELIVERY     = MODEL_CONFIG["delivery"]

if DELIVERY == "prepend":
    OLLAMA_URL   = MODEL_CONFIG.get("endpoint", "http://localhost:11434/api/generate")
    OLLAMA_MODEL = MODEL_CONFIG.get("model", "mistral:latest")
elif DELIVERY == "atlas":
    BLANKY = Path(MODEL_CONFIG.get("blanky", "~/dev/blanky")).expanduser()
    if not BLANKY.is_dir():
        print(f"Error: blanky directory not found: {BLANKY}")
        sys.exit(1)
    _pattern = str(
        Path.home() / ".vscode" / "extensions" / "anthropic.claude-code-*"
        / "resources" / "native-binary" / "claude.exe"
    )
    _matches = sorted(_glob.glob(_pattern))
    if not _matches:
        print("Error: claude binary not found. Check VSCode extension path.")
        sys.exit(1)
    CLAUDE_BIN = Path(_matches[-1])
else:
    print(f"Error: unknown delivery '{DELIVERY}'. Must be 'prepend' or 'atlas'.")
    sys.exit(1)

# ---------------------------------------------------------------------------
# File loaders
# ---------------------------------------------------------------------------

def load_scenario(model_slug: str) -> str:
    path = HERE / f"SCENARIO_{SCENARIO_CODE}_{model_slug}.md"
    if not path.exists():
        raise FileNotFoundError(f"Scenario file not found: {path.name}")
    return path.read_text(encoding="utf-8")


def load_glyph(glyph_code: str) -> str:
    path = HERE / f"GLYPH_{glyph_code}.md"
    if not path.exists():
        raise FileNotFoundError(f"Glyph file not found: {path.name}")
    return path.read_text(encoding="utf-8")


def load_ground(ground_code: str) -> str:
    path = HERE / f"GROUND_{ground_code}.md"
    if not path.exists():
        raise FileNotFoundError(
            f"Ground file not found: {path.name}. "
            "Has ground been extracted from condition 2 Ii traces?"
        )
    return path.read_text(encoding="utf-8")

# ---------------------------------------------------------------------------
# Blanky management (Claude only)
# ---------------------------------------------------------------------------

def reset_blanky() -> None:
    for item in BLANKY.iterdir():
        if item.name == ".claude":
            continue
        if item.is_dir():
            shutil.rmtree(item)
        else:
            item.unlink()

# ---------------------------------------------------------------------------
# Prompt assembly
# ---------------------------------------------------------------------------

RESPONSE_HEADER = """\
# RESPONSE_{condition}_{model}_{n}

**Study:** {study_name}
**Model:** {model}
**Condition:** {condition}
**Glyph:** {glyph}
**Ground:** {ground}
**Run:** {n} of {runs_per_condition}
**Date:** {date}

---

"""


def build_mistral_prompt(condition_code: str) -> str:
    """Mistral: glyph (and ground if present) prepended to scenario, separated by ---."""
    condition   = CONDITIONS[condition_code]
    glyph_code  = condition.get("glyph_code")
    ground_code = condition.get("ground_code")

    parts = []
    if glyph_code:
        parts.append(load_glyph(glyph_code))
        if ground_code:
            parts.append("---")
            parts.append(load_ground(ground_code))
        parts.append("---")
    parts.append(load_scenario("mistral"))
    return "\n\n".join(parts)


def build_claude_prompt(condition_code: str) -> tuple[str, str | None]:
    """Claude: scenario via -p, glyph (+ ground if present) via --system-prompt."""
    condition   = CONDITIONS[condition_code]
    glyph_code  = condition.get("glyph_code")
    ground_code = condition.get("ground_code")

    scenario_text = load_scenario("claude")

    system_parts = []
    if glyph_code:
        system_parts.append(load_glyph(glyph_code))
        if ground_code:
            system_parts.append("---")
            system_parts.append(load_ground(ground_code))

    system_text = "\n\n".join(system_parts) if system_parts else None
    return scenario_text, system_text

# ---------------------------------------------------------------------------
# Model calls
# ---------------------------------------------------------------------------

def call_ollama(prompt: str) -> str:
    raise RuntimeError(
        "Ollama is RETIRED and uninstalled (2026-07-14) - do NOT install it. "
        "llama.cpp (llama-server, http://localhost:8081, OpenAI-compatible) is the "
        "only local inference portal; port this call to it. "
        "See memory one-local-inference-portal-llama-cpp."
    )
    payload = json.dumps({
        "model":  OLLAMA_MODEL,
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


def call_claude(scenario_text: str, system_text: str | None = None) -> str:
    cmd = [str(CLAUDE_BIN), "-p", scenario_text]
    if system_text:
        cmd += ["--system-prompt", system_text]
    result = subprocess.run(
        cmd,
        cwd=str(BLANKY),
        capture_output=True,
        encoding="utf-8",
        errors="replace",
        timeout=TIMEOUT,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"claude exited with code {result.returncode}\n"
            f"stderr: {result.stderr.strip()}"
        )
    return result.stdout

# ---------------------------------------------------------------------------
# Run
# ---------------------------------------------------------------------------

def run_single(condition_code: str, n: int) -> None:
    condition   = CONDITIONS[condition_code]
    glyph_code  = condition.get("glyph_code") or "none"
    ground_code = condition.get("ground_code") or "none"
    out_path    = HERE / f"RESPONSE_{condition_code}_{MODEL_SLUG}_{n}.md"

    if out_path.exists():
        print(f"[{MODEL_SLUG}] {condition_code} run {n}: already exists — skipping.", flush=True)
        return

    print(f"[{MODEL_SLUG}] {condition_code} run {n}: building prompt...", flush=True)

    if DELIVERY == "prepend":
        prompt   = build_mistral_prompt(condition_code)
        response = call_ollama(prompt)

    elif DELIVERY == "atlas":
        scenario_text, system_text = build_claude_prompt(condition_code)
        print(f"[{MODEL_SLUG}] {condition_code} run {n}: resetting blanky...", flush=True)
        reset_blanky()
        print(f"[{MODEL_SLUG}] {condition_code} run {n}: sending to claude...", flush=True)
        response = call_claude(scenario_text, system_text)

    header = RESPONSE_HEADER.format(
        condition=condition_code,
        n=n,
        study_name=STUDY_NAME,
        model=MODEL_SLUG,
        glyph=glyph_code,
        ground=ground_code,
        runs_per_condition=RUNS_PER_CONDITION,
        date=date.today().isoformat(),
    )
    out_path.write_text(header + response, encoding="utf-8")
    print(f"[{MODEL_SLUG}] {condition_code} run {n}: written to {out_path.name}", flush=True)

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

def parse_args(argv: list) -> list:
    if not argv:
        return [
            (code, n)
            for code in CONDITION_ORDER
            for n in range(1, RUNS_PER_CONDITION + 1)
        ]
    jobs = []
    i = 0
    while i < len(argv):
        token = argv[i]
        if token in CONDITIONS:
            if i + 1 < len(argv) and argv[i + 1].isdigit():
                run_n = int(argv[i + 1])
                if 1 <= run_n <= RUNS_PER_CONDITION:
                    jobs.append((token, run_n))
                else:
                    print(f"Run number {run_n} out of range (1–{RUNS_PER_CONDITION}) — skipping")
                i += 2
            else:
                jobs.extend((token, n) for n in range(1, RUNS_PER_CONDITION + 1))
                i += 1
        else:
            print(f"Unknown condition '{token}' — skipping")
            i += 1
    return jobs

# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    jobs = parse_args(_remaining_argv)
    if not jobs:
        print("No valid jobs to run.")
        return

    if DELIVERY == "prepend":
        backend_display = f"{OLLAMA_MODEL} via Ollama"
    else:
        backend_display = "claude -p (blanky, glyph via --system-prompt)"

    print(f"Study: {STUDY_NAME}")
    print(f"Model: {MODEL_SLUG} ({backend_display})")
    print(f"Induction vector: {INDUCTION_VECTOR}")
    print(f"Running {len(jobs)} job(s) sequentially.\n")
    for condition_code, n in jobs:
        run_single(condition_code, n)
    print("\nDone.")


if __name__ == "__main__":
    main()
