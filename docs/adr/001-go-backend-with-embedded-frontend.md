# ADR 001: Go Backend with Embedded React Frontend

## Context

seance needs a simple webapp that runs locally. It requires a backend API to read Claude Code conversation files from disk and a web frontend for browsing/searching. The app should be easy to run with minimal setup.

## Decision

Use Go for the backend with `embed.FS` to serve a Vite-built React frontend as a single binary.

## Consequences

- Single binary distribution — no runtime dependencies, just run `./seance`
- Go's stdlib `net/http` is sufficient for the API; no framework needed
- Frontend is built with Vite + React, output embedded at compile time
- Trade-off: frontend changes require a full rebuild of the Go binary
- Trade-off: Go's type system adds verbosity for JSON parsing, but provides safety
