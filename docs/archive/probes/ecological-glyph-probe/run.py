"""

DEPRECATED (e9-typed-study-runners, 2026-04-19): Rust port of this
probe lives at lab-app/crates/lab-app-server/src/studies/. See
blueprints/probes/MIGRATION.md for the Rust invocation equivalent of each
python3 run.py ... call. This file is retained as reference documentation
until at least one full assay series runs through the typed runners without
issue.

Ecological Glyph Probe — runner (v4)

Places the model inside a workflow scenario as an acting subject. Tests whether
a glyph loaded as context changes behavior at a decision point embedded in a
procedure-scoped environment.

Supports two backends:
    mistral  — POST to Ollama endpoint (local). Glyph prepended to scenario.
    claude   — subprocess call to claude -p from blanky. Glyph via --system-prompt.
               The --system-prompt channel is required for Claude: prepending the
               glyph to -p triggers safety heuristics (injection recognition failure).

Required files in the study directory:
    study.json              — study configuration (see SETUP.md)
    SCENARIO_{code}.md      — one per distinct scenario_file value in study.json
    GLYPH_{code}.md         — one per glyph_code value (omit for control)

Claude only — optional, but required if Claude needs to read files:
    STATE_{code}.json       — filesystem state to write into blanky before each run
                              Format: {"files": {"relative/path": "content", ...}}

Usage:
    python3 run.py --model mistral                        # all runs, Mistral
    python3 run.py --model claude                         # all runs, Claude
    python3 run.py --model mistral [slug]-eco-control     # one scenario, all runs
    python3 run.py --model mistral [slug]-eco-control 2   # one scenario, one run
"""

import sys
import json
import shutil
import subprocess
import urllib.request
import glob as _glob
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

STUDY_NAME        = CONFIG["name"]
RUNS_PER_SCENARIO = CONFIG.get("runs_per_scenario", 8)
SCENARIOS         = CONFIG["scenarios"]
SCENARIO_ORDER    = CONFIG.get("scenario_order", list(SCENARIOS.keys()))
INDUCTION_VECTOR  = CONFIG.get("induction_vector", "(not specified)")

# Backend config
if "models" in CONFIG:
    # Multi-model study.json format
    if MODEL_SLUG not in CONFIG["models"]:
        print(f"Error: model '{MODEL_SLUG}' not found in study.json 'models'. Available: {list(CONFIG['models'].keys())}")
        sys.exit(1)
    MODEL_CONFIG = CONFIG["models"][MODEL_SLUG]
    BACKEND      = MODEL_CONFIG["backend"]
    TIMEOUT      = MODEL_CONFIG.get("timeout", 300)
else:
    # Single-model study.json format (Mistral default)
    BACKEND      = "ollama"
    TIMEOUT      = CONFIG.get("timeout", 300)
    MODEL_CONFIG = CONFIG

if BACKEND == "ollama":
    OLLAMA_URL   = MODEL_CONFIG.get("endpoint", "http://localhost:11434/api/generate")
    OLLAMA_MODEL = MODEL_CONFIG.get("model", "mistral:latest")
elif BACKEND == "claude":
    BLANKY = Path(MODEL_CONFIG.get("blanky", "~/dev/blanky")).expanduser()
    if not BLANKY.is_dir():
        print(f"Error: blanky directory not found: {BLANKY}")
        sys.exit(1)
    _pattern = str(Path.home() / ".vscode" / "extensions" / "anthropic.claude-code-*" / "resources" / "native-binary" / "claude.exe")
    _matches = sorted(_glob.glob(_pattern))
    if not _matches:
        print("Error: claude binary not found. Check VSCode extension path.")
        sys.exit(1)
    CLAUDE_BIN = Path(_matches[-1])
else:
    print(f"Error: unknown backend '{BACKEND}'. Must be 'ollama' or 'claude'.")
    sys.exit(1)

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


def load_state(scenario_file: str) -> dict:
    path = HERE / f"STATE_{scenario_file}.json"
    if not path.exists():
        return {}
    with open(path, encoding="utf-8") as f:
        data = json.load(f)
    return data.get("files", {})

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


def setup_scenario_state(scenario_file: str) -> list:
    state = load_state(scenario_file)
    if not state:
        return []
    created = []
    for rel_path, content in state.items():
        abs_path = BLANKY / rel_path
        abs_path.parent.mkdir(parents=True, exist_ok=True)
        abs_path.write_text(content, encoding="utf-8")
        created.append(abs_path)
    return created


def teardown_scenario_state(created_paths: list) -> None:
    for path in created_paths:
        if path.exists():
            path.unlink()
    dirs_seen = set()
    for path in created_paths:
        for parent in path.parents:
            if parent == BLANKY:
                break
            dirs_seen.add(parent)
    for d in sorted(dirs_seen, key=lambda x: len(x.parts), reverse=True):
        if d.exists() and d.is_dir() and not any(d.iterdir()):
            d.rmdir()

# ---------------------------------------------------------------------------
# Prompt assembly
# ---------------------------------------------------------------------------

RESPONSE_HEADER = """\
# RESPONSE_{scenario}_{n}

**Study:** {study_name}
**Model:** {model}
**Scenario:** {scenario}
**Condition:** {condition}
**Glyph:** {glyph}
**Run:** {n} of {runs_per_scenario}
**Date:** {date}

---

"""


def build_mistral_prompt(scenario_code: str) -> str:
    """Mistral: glyph prepended to scenario, separated by ---."""
    scenario      = SCENARIOS[scenario_code]
    scenario_file = scenario["scenario_file"]
    glyph_code    = scenario.get("glyph_code")

    parts = []
    if glyph_code:
        parts.append(load_glyph(glyph_code))
        parts.append("---")
    parts.append(load_scenario(scenario_file))
    return "\n\n".join(parts)


def build_claude_prompt(scenario_code: str) -> tuple[str, str | None]:
    """Claude: scenario via -p, glyph via --system-prompt (atlas delivery)."""
    scenario      = SCENARIOS[scenario_code]
    scenario_file = scenario["scenario_file"]
    glyph_code    = scenario.get("glyph_code")
    scenario_text = load_scenario(scenario_file)
    glyph_text    = load_glyph(glyph_code) if glyph_code else None
    return scenario_text, glyph_text

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


def call_claude(scenario_text: str, glyph_text: str | None = None) -> str:
    cmd = [str(CLAUDE_BIN), "-p", scenario_text]
    if glyph_text:
        cmd += ["--system-prompt", glyph_text]
    result = subprocess.run(
        cmd,
        cwd=str(BLANKY),
        capture_output=True,
        text=True,
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

def run_single(scenario_code: str, n: int) -> None:
    scenario      = SCENARIOS[scenario_code]
    scenario_file = scenario["scenario_file"]
    out_path      = HERE / f"RESPONSE_{scenario_code}_{n}.md"

    if out_path.exists():
        print(f"[{MODEL_SLUG}] {scenario_code} run {n}: already exists — skipping.", flush=True)
        return

    print(f"[{MODEL_SLUG}] {scenario_code} run {n}: building prompt...", flush=True)

    if BACKEND == "ollama":
        prompt   = build_mistral_prompt(scenario_code)
        response = call_ollama(prompt)

    elif BACKEND == "claude":
        scenario_text, glyph_text = build_claude_prompt(scenario_code)
        print(f"[{MODEL_SLUG}] {scenario_code} run {n}: resetting blanky...", flush=True)
        reset_blanky()
        print(f"[{MODEL_SLUG}] {scenario_code} run {n}: writing scenario state...", flush=True)
        created = setup_scenario_state(scenario_file)
        try:
            print(f"[{MODEL_SLUG}] {scenario_code} run {n}: sending to claude...", flush=True)
            response = call_claude(scenario_text, glyph_text)
        finally:
            teardown_scenario_state(created)

    glyph_code = scenario.get("glyph_code") or "none"
    header = RESPONSE_HEADER.format(
        scenario=scenario_code,
        n=n,
        study_name=STUDY_NAME,
        model=MODEL_SLUG,
        condition=scenario.get("condition", ""),
        glyph=glyph_code,
        runs_per_scenario=RUNS_PER_SCENARIO,
        date=date.today().isoformat(),
    )
    out_path.write_text(header + response, encoding="utf-8")
    print(f"[{MODEL_SLUG}] {scenario_code} run {n}: written to {out_path.name}", flush=True)

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

def parse_args(argv: list) -> list:
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
    jobs = parse_args(_remaining_argv)
    if not jobs:
        print("No valid jobs to run.")
        return

    backend_display = OLLAMA_MODEL if BACKEND == "ollama" else f"claude -p (blanky, glyph via --system-prompt)"
    print(f"Study: {STUDY_NAME}")
    print(f"Model: {MODEL_SLUG} ({backend_display})")
    print(f"Induction vector: {INDUCTION_VECTOR}")
    print(f"Running {len(jobs)} job(s) sequentially.\n")
    for scenario_code, n in jobs:
        run_single(scenario_code, n)
    print("\nDone.")


if __name__ == "__main__":
    main()
