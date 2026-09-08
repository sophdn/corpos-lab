#!/usr/bin/env bash
# scripts/worktree-setup.sh — make a linked corpos worktree commit-ready.
#
# Installs a gate-only pre-commit (the full scripts/gate.sh) for THIS worktree
# via a per-worktree core.hooksPath, so commits here run the gate but NOT the
# main-checkout branch guard (scripts/guard-worktree-discipline.sh). Run once
# from inside the worktree — worktree-new.sh does this for you. Idempotent; the
# main checkout's config is never touched. (skill: worktree-workflow)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

# shellcheck source=gitflow-common.sh
# shellcheck disable=SC1091
. "$SCRIPT_DIR/gitflow-common.sh"
gitflow_load_config "$REPO_ROOT"

GIT_COMMON="$(git rev-parse --path-format=absolute --git-common-dir)"
case "$GIT_COMMON" in
    */.git) MAIN_ROOT="$(dirname "$GIT_COMMON")" ;;
    *)      MAIN_ROOT="" ;;
esac
if [ -z "$MAIN_ROOT" ] || [ "$MAIN_ROOT" = "$REPO_ROOT" ]; then
    echo "worktree-setup: run from inside a linked worktree, not the main checkout." >&2
    echo "                (the main checkout's hooks are wired when the git-flow service is installed)" >&2
    exit 1
fi

GIT_DIR="$(git rev-parse --absolute-git-dir)"
hooks_dir="$GIT_DIR/gate-only-hooks"
mkdir -p "$hooks_dir"
# Generate the gate-only hook from the configured gate command. An empty
# GITFLOW_GATE_CMD means the repo has no gate, so the hook is a no-op. Both run
# the gate but NOT the main-checkout branch guard.
if [ -n "$GITFLOW_GATE_CMD" ]; then
    cat > "$hooks_dir/pre-commit" <<EOF
#!/usr/bin/env bash
# Gate-only pre-commit for a linked worktree (gitflow/worktree-setup.sh).
# Runs the configured gate; deliberately NO main-checkout branch guard here.
cd "$REPO_ROOT" || exit 1
exec bash -c '$GITFLOW_GATE_CMD' _ "\$@"
EOF
else
    cat > "$hooks_dir/pre-commit" <<'EOF'
#!/usr/bin/env bash
# Gate-only pre-commit for a linked worktree (gitflow/worktree-setup.sh).
# This repo configures no gate (GITFLOW_GATE_CMD empty) — nothing to run.
exit 0
EOF
fi
chmod +x "$hooks_dir/pre-commit"

# Per-worktree core.hooksPath override (the main checkout keeps core.hooksPath
# = .githooks, which carries the guard). extensions.worktreeConfig is benign +
# idempotent; --worktree writes only this worktree's private config.
git config extensions.worktreeConfig true
git config --worktree core.hooksPath "$hooks_dir"

echo "worktree-setup: gate-only hooks at $hooks_dir"
echo "worktree-setup: 'git commit' in this worktree now runs the gate only."
