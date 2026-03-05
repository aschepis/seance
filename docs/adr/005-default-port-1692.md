# ADR 005: Default Port 1692

## Context

The application needed a default port for the local HTTP server. The original default was 3333, which is generic and commonly used by other development tools, risking conflicts.

## Decision

Use port 1692 as the default — a reference to the year of the Salem witch trials, fitting the "seance" theme.

## Consequences

- Memorable and thematic port number
- Low risk of conflict with common dev tools (3000, 3333, 5173, 8080)
- Port is configurable via `--port` flag if 1692 is unavailable
