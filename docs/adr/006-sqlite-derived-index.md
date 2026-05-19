# ADR 006: SQLite Derived Index

## Context

ADR 002 chose on-demand JSONL parsing with no database. That works for a few hundred conversations but scales poorly: `/api/search` linear-scans every file on every request, listing requires re-walking the tree, and there is nowhere to store user-level state (pins, tags, notes) or to express cross-conversation queries (by tool used, by model, by date range). As `~/.claude/projects/` grows past a few thousand sessions this becomes the dominant UX cost.

## Decision

Add a SQLite database as a **derived index** alongside — not replacing — the JSONL files. JSONL remains the source of truth; the DB is a rebuildable cache.

- **Library**: `modernc.org/sqlite` (pure-Go, no CGO) so `go install` still works cleanly. It supports FTS5.
- **Location**: `~/.claude/seance/index.db`. Created on first run; safe to delete.
- **Indexing**: on startup, walk `~/.claude/projects/` and re-index any JSONL whose `mtime` is newer than the last-seen value stored per-file. Unchanged files are skipped. A `seance reindex` CLI flag forces a full rebuild.
- **Search**: an FTS5 virtual table over message content, returning ranked results with snippets. Existing `/api/search` switches to query the DB.
- **Schema versioning**: a `schema_version` row. On mismatch, drop and rebuild from JSONL (cheap because the DB is derived). No migration framework needed.
- **User-state tables** (pins, tags, notes) are kept in the same DB but tagged so a rebuild preserves them — they key off stable `(session_id, message_uuid)` pairs that survive re-indexing.
- The JSONL discovery layer (ADR 002 + ADR 004) is unchanged; the indexer consumes its output.

## Consequences

- Search becomes O(log n) with ranking and snippets instead of O(n) substring matching.
- Cross-conversation queries (by tool, model, date, project) and user-level metadata (pins, tags, notes) become trivial — capabilities the JSONL-only design cannot express.
- Startup cost grows with the number of *changed* files, not total files; cold start on a fresh machine pays a one-time indexing cost.
- A second source of state to reason about, but rebuild-from-disk keeps the failure mode benign: corruption is fixed by `rm ~/.claude/seance/index.db`.
- Trade-off: the pure-Go SQLite driver is slower than the CGO `mattn/go-sqlite3` build, but the difference is irrelevant at local-app scale and the install-friction win is worth it.
