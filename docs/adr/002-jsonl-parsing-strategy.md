# ADR 002: JSONL Parsing Strategy

## Context

Claude Code stores conversation history as JSONL files in `~/.claude/projects/<encoded-path>/<session-id>.jsonl`. Each line is a JSON object representing a message, tool use, progress update, or file snapshot. The encoding of project paths replaces `/` with `-`.

## Decision

Parse JSONL files on-demand using streaming line-by-line reading. No database or index. Deduplicate assistant message chunks by UUID. Filter out non-message types (progress, file-history-snapshot).

## Consequences

- Simple implementation with no external dependencies
- Scales well for typical local usage (hundreds of conversations)
- Search is O(n) across all files — acceptable for local use, may need indexing later
- Memory-efficient: only one conversation fully loaded at a time for viewing
- Trade-off: repeated listing requires re-scanning all files (no caching)
