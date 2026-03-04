# seance — Execution Plan

## Guiding Principle

Build the smallest working increment first. Expand depth-first. Every change must
compile, pass tests, and leave maslow verify --profile quick green.

---

## Milestones

### M1 - Foundation (DONE)

Goal: Project structure, build pipeline, and core parsing working.

Deliverables:
- Go module with Makefile (build, test, lint targets)
- JSONL conversation parser
- REST API endpoints
- React frontend with conversation list and viewer
- Embedded single-binary distribution

Exit criteria:
- maslow verify --profile quick passes
- maslow verify --profile full passes
- `make build` produces working binary
- Tests pass for parser and API

---

### M2 - Enhancements (Future)

Goal: Polish and additional features.

Potential deliverables:
- Caching layer for conversation index
- Full-text search indexing
- Markdown rendering in conversation view
- URL-based routing with browser history
- Conversation export

Exit criteria:
- maslow verify --profile full passes
- Performance acceptable for large conversation histories

---

## Parallel Workstreams

| Stream | Owner Paths | Milestone Scope |
|--------|-------------|-----------------|
| Backend | internal/, cmd/ | M1 |
| Frontend | web/frontend/ | M1 |
| Docs | docs/ | M1, M2 |

---

## Definition of Done

- maslow verify --profile full passes
- All ADRs are written
- docs/MAP.md is current
- Application builds and runs locally
- Conversations from ~/.claude/projects/ are listed and viewable
- Search works across conversations
