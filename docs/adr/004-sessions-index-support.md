# ADR 004: Sessions Index Support

## Context

Newer versions of Claude Code write a `sessions-index.json` file in each project directory instead of (or in addition to) flat JSONL files. This index contains metadata about sessions including project path, git branch, message count, and timestamps. The existing parser only discovered conversations by scanning for `*.jsonl` files, which missed sessions tracked only in the index.

## Decision

Support both discovery methods: scan for flat JSONL files first, then supplement with entries from `sessions-index.json`. When both exist for a session, prefer the JSONL file. Use the index's `projectPath` field to resolve the accurate project path (the directory-name-based decoding is lossy). For sessions where the JSONL file no longer exists on disk, report `messageCount: 0` so the frontend can filter them.

## Consequences

- Conversations from newer Claude Code versions are now discoverable
- Project paths are more accurate when the index is available
- Sessions without JSONL files on disk appear in listings but show no messages (graceful degradation)
- Trade-off: two discovery paths add complexity, but the fallback chain is straightforward
