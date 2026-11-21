#!/usr/bin/env bash
# Install the blind-action-scorer subagent into the local, gitignored .claude/agents/.
# Run once after cloning (like install-hooks.sh). The canonical, tracked copy lives at
# tools/blind-scorer/blind-action-scorer.md; .claude/ is gitignored (local session config),
# so each checkout installs the agent explicitly. The agent strips the MCP surface
# (disallowedTools) so blind scoring runs on a truly vanilla subagent that cannot look the
# study up. See tools/blind-scorer/README.md.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
src="tools/blind-scorer/blind-action-scorer.md"
dst_dir=".claude/agents"
mkdir -p "$dst_dir"
cp "$src" "$dst_dir/blind-action-scorer.md"
echo "installed $src -> $dst_dir/blind-action-scorer.md (MCP-free blind scorer available as subagent 'blind-action-scorer')."
