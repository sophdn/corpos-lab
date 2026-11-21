# Task catalogue and replay harness

This is the method behind corpos-lab chain `popular-task-catalogue`. It answers
one question: what share of the principal's normal work can corpos carry?

Earlier swap rehearsals drove corpos through hand-picked targets. This method
draws the workload from what the principal actually typed, weights each kind of
task by how often it occurs, and replays real cases through corpos in a
sandbox.

All data is private. It lives in `corpus/private/task-catalogue/` (a separate
private repo, see CLAUDE.md) and in `~/.local/state/corpos-lab/replay/`. This
document and the code hold no prompt text.

## 1. Mine: the kind table

```sh
corpos-lab task-catalogue mine -taxonomy <taxonomy.toml> -out <dir> [-limit N]
corpos-lab task-catalogue aggregate -taxonomy <taxonomy.toml> -out <dir> [-threshold 0.8]
```

`mine` reads the Claude Code prompt log (`~/.claude/history.jsonl`) and writes
`prompts.jsonl` and `labels.jsonl` into `<dir>`.

- **Project attribution.** Most prompts are typed from a neutral directory, so
  the working directory does not name the project. A prompt gets the project it
  names (project ids, retired aliases, and work-item slugs with two or more
  hyphens, read from the ledger), else the project its session's toolkit calls
  were scoped to, else the project other prompts in the session name, else none.
  The ledger is read through `admin.db_query` over HTTP.
- **Labels.** The local portal model labels each prompt with a role and a task
  kind from a closed taxonomy the caller supplies. Roles say whether a prompt
  opens new work (`opener`) or continues work under way (approve, steer,
  report, control, chatter). Only openers are catalogued; the others are the
  follow-up traffic multi-turn cases replay. Harness commands (`/…`) are
  labelled `control` without a model call.
- **Record what ran.** The full sampler chain is pinned (greedy). Each label
  row records the served model, build and speed. Each run writes
  `runs/<ts>.json` with `/props`, the sampler and the commit.
- **Resumable.** Labels append as they finish. A re-run labels only what is
  missing; `-retry-failed` re-labels invalid and errored rows.

`aggregate` writes `kinds.json` and prints the table. Every prompt lands in
exactly one bucket, and the tool checks the sums. The catalogue set is the head
of the table, by count, that covers the threshold share of openers. The
`other` and `none` kinds are never catalogued.

## 2. Cases

A case is one replayable task, as JSON:

| field | meaning |
|---|---|
| `id` | lowercase letters, digits, hyphens |
| `kind` | the catalogue kind it stands for |
| `source_prompts` | prompt ids it was drawn from (empty for synthetic cases) |
| `project` | ledger project corpos's tool calls are scoped to |
| `cwd` | absolute directory the principal was in |
| `repos` | git checkouts the case may read or change (copied) |
| `files` | other paths it touches (copied) |
| `writes` | does a correct run change anything |
| `profile` | optional corpos job-profile; empty lets corpos choose |
| `turns` | `{text, trigger, adapt}` per principal message, in order |
| `rubric` | 2-5 checkable outcomes of a good run |
| `notes` | free text |

A case states the opener so it stands alone. A prompt that only made sense in
its conversation ("fix bug 975", "yes do that") is rewritten, with the original
kept in the private case notes. Every case starts from today's state: the
newest ledger backup and each repo's HEAD. A case whose original target no
longer exists is re-aimed at a live equivalent of the same kind.

## 3. Replay

```sh
corpos-lab replay up <case.json>
corpos-lab replay turn <case-id> -prompt "<principal turn>"   # repeat per turn
corpos-lab replay check <case-id>
corpos-lab replay down <case-id> -ledger <replay-ledger.jsonl>
```

The sandbox has three fences.

1. **Files.** corpos runs under bubblewrap with the host root read-only. Each
   repo and file the case names is a throwaway copy, bound at its real path, so
   a prompt can name real paths and every write lands in the copy. corpos's
   state and cache directories are sandbox directories. Any other write fails.
2. **Ledger.** A throwaway the toolkit service container serves a copy of the newest
   nightly ledger backup on its own loopback port, and corpos's `-mcp-url`
   points there. The container gets no team-sync secret, no transcript mounts,
   and the dev root read-only.
3. **Git.** Each repo copy drops the clone's origin (which points at the live
   checkout) and re-adds the real remotes with push disabled.

corpos runs with `-risk-gate build-test` by default, the swap window's posture
(corpos `docs/SWAP_VALIDATION_CRITERIA.md` §4): coding workers may build and test,
which reaches only the sandbox copies. `-risk-gate enforce` blocks even `go test`.
Inside the sandbox `/run/user` is an empty tmpfs, so the systemd, D-Bus and podman
sockets are out of reach, and `~/dev` is a throwaway directory with every existing
entry re-mounted read-only, so git-flow can create sibling worktrees.

corpos's per-turn timeout stays at its default (2 minutes), because that is what a
daily driver gets. A turn that ends with "resume to continue" is answered the way the
principal would answer it, and each such continuation is counted as friction.

`check` and `down` report two leak checks. **Ledger:** each entity the sandbox
ledger recorded since `up` must have no event in the live ledger in the same
window. **Repos:** each live repo's HEAD and status must match the snapshot
from `up`. The principal may be working in a live repo at the same time, so a
moved repo is reported rather than called a leak.

`down` appends one row to the replay ledger and deletes the copies. The row
carries dogfood.jsonl's signals (cost, tokens, tools, answer) plus the case,
its turns, what it changed, and the leak checks. It keeps `state.json`, the
turn transcripts, and corpos's session DBs for the judge.

The driver acts as the principal under the `corpos-swap-rehearsal` skill. It
frames, answers, and judges; it never does corpos's work. Anything it would
have to do itself is logged as a fallback with the missing capability.

## 4. Judge and report

An automated judge scores each case pass, partial or gap against its rubric.
The judge is not the agent that drove the case, and it does not see the ledger
outcome. The principal hand-grades about 20 cases; exact agreement must reach
80% before the judge's scores count. Coverage is then the share of openers,
weighted by kind frequency, whose cases pass.

This is a rehearsal signal. It is not the harness-swap validation window.
