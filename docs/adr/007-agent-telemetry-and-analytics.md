# ADR 007: Agent Telemetry and Analytics

## Context

ADR 006 introduces a SQLite index over Claude Code conversation JSONL. Once that index exists, a class of queries that today require offline Python scripts (e.g. `analyze_cc_transcripts.py`) becomes cheap and interactive: shell-tool vs `str_replace` ratios, context-flooding token overhead, `str_replace → sed/awk` degradation sequences, per-model cost, cache hit ratios. This is the natural next step for a transcript viewer — surfacing the metadata that already lives in the JSONL but is currently invisible to the user. The schema needs to capture this metadata up front so analytics features can be added incrementally without reindexing.

## Decision

Treat agent telemetry as a first-class feature of seance, built on the ADR 006 index. Concretely:

- **Schema additions (beyond search needs)**:
  - `messages` table captures `model`, `input_tokens`, `output_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens`, `timestamp`. (The Python reference script sums only two of these and undercounts real cost.)
  - `tool_calls` table: `(session_id, tool_use_id PRIMARY KEY, tool_name, input_json, output_bytes, output_tokens_est, started_at, ended_at NULL)`. Pairs `tool_use` with `tool_result` by `tool_use_id` at index time so the join is precomputed.
  - Derived columns computed during indexing: `is_shell`, `is_str_replace`, `floods_context`, `shell_commands_csv`. Cheap to recompute on rebuild; expensive to compute at query time.
- **API surface**: a new `/api/analytics/*` namespace returning aggregates (overview, per-session, time-series). Frontend gets an "Analytics" tab alongside the conversation list.
- **CLI**: `seance analyze [--session ID] [--json]` reproduces the Python script's report from the indexed data — no JSONL re-scan needed. The Python script becomes a reference implementation we delete once parity is reached.
- **Derived-flag heuristics live in one place** (`internal/analytics/heuristics.go`) so the "what counts as shell tool / context flooding / degradation" rules are versioned with the schema and easy to tune.
- **Out of scope for now**: alerting, multi-user aggregation, exporting metrics to external systems (Prometheus etc.). Local read-only analytics only.

## Consequences

- Users get actionable feedback on agent efficacy (shell%, token overhead, worst sessions, degradation patterns) without leaving seance — the script's value, but interactive and per-session.
- The index becomes denormalized on purpose: derived flags and token totals are materialized at index time. Acceptable because the DB is rebuildable per ADR 006 — when heuristics change, bump the schema version and rebuild.
- Scope expands modestly: the product is still a transcript viewer, with richer per-session metadata surfaced. MAP.md gets a section on what telemetry is captured; README doesn't need a reframe.
- Heuristics are opinionated and will drift as Claude Code's tool names and message shapes evolve (`Bash`, `str_replace_editor`, `Edit`, etc.). Centralizing them in one file with tests is the mitigation; expect to update them.
- Trade-off: a richer schema means a slower full reindex on a cold start (more columns, more computation per message). Still bounded by JSONL read speed, and incremental updates are unaffected.
