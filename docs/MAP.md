# seance — Repository Map

## What This Project Is

seance is a local webapp for browsing Claude Code conversation history. It reads JSONL conversation files from `~/.claude/projects/` and presents them in a searchable, browsable interface grouped by project.

---

## Architecture Overview

```
┌─────────────────────────────────────────────┐
│                  Browser                     │
│  React SPA (sidebar + conversation viewer)   │
└─────────────┬───────────────────────────────┘
              │ HTTP
┌─────────────▼───────────────────────────────┐
│              Go Server (cmd/seance)          │
│  ┌─────────────────────────────────────────┐│
│  │ API Layer (internal/api)                ││
│  │  GET /api/conversations                 ││
│  │  GET /api/conversations/:id             ││
│  │  GET /api/search?q=...                  ││
│  └───────────┬─────────────────────────────┘│
│  ┌───────────▼─────────────────────────────┐│
│  │ Parser (internal/conversations)         ││
│  │  Reads ~/.claude/projects/**/*.jsonl    ││
│  └─────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────┐│
│  │ Embedded Frontend (web/dist via embed)  ││
│  └─────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
```

---

## Key Entrypoints

| Path | Role |
|------|------|
| cmd/seance/main.go | Application entrypoint — starts HTTP server |
| internal/conversations/parser.go | JSONL parsing and conversation discovery |
| internal/conversations/types.go | Domain types (Conversation, Message, etc.) |
| internal/api/handler.go | HTTP API handlers |
| web/embed.go | Embeds built frontend assets |
| web/frontend/ | React frontend source |

---

## Canonical File Locations

| Artifact | Location | Notes |
|----------|----------|-------|
| Project spec | maslow.yaml | Source of truth for verification |
| Agent guide | CLAUDE.md | Conventions and process |
| Verify output | reports/verify.json | Written by maslow verify |
| Documentation | docs/ | MAP, PLAN, ADRs, templates |
| ADRs | docs/adr/ | Architecture Decision Records |
| Templates | docs/templates/ | Decision templates |
| Tasks | docs/tasks/ | Human-authored tasks for agents |
| Task convention | docs/tasks/CONVENTION.md | Task format and agent protocol |
| Go binary | bin/seance | Built by `make build` |
| Frontend dist | web/dist/ | Built by Vite, embedded in binary |

---

## Multi-Agent Conventions

- Each workstream owns a distinct directory scope; see docs/PLAN.md for assignments.
- Policy enforcement in maslow.yaml governs which paths agents may modify.
- Verification is the shared integration point; all agents must leave maslow verify green.
- Every material decision is recorded as an ADR in docs/adr/.

---

## Navigation

- Architecture decisions: docs/adr/
- Execution plan and milestones: docs/PLAN.md
- Project spec: maslow.yaml
- Decision templates (unfilled questions): docs/templates/
