# seance — Agent Guide

## Project Overview

seance is managed by Maslow, an executable specification system for agent-built software.
The project is defined by a declarative maslow.yaml spec that is validated, verified, and auditable.

## What to Read First

1. This file (CLAUDE.md) — conventions, principles, and process
2. maslow.yaml — the spec (source of truth); read all refs that point to docs/ files
3. docs/MAP.md — architecture overview and key entrypoints
4. docs/PLAN.md — milestones, workstreams, and Definition of Done

## Refs as Generative Input

Refs in maslow.yaml are not just verification targets — they are your primary input. Before starting any
generative work, read all refs to understand your context:

- **Doc refs** (kind: file pointing to docs/) are requirements, constraints, and context. These are your north star.
  Read them before making any decisions.
- **Config refs** (kind: file pointing to config files like .prettierrc, .eslint) are verification targets.
  Ensure they exist and are respected, but don't treat them as requirements.
- **URL refs** (kind: url) may point to external context — competitor apps, design inspiration, API standards.
  Fetch and use them as context when relevant.
- **MCP refs** (kind: mcp) declare tool dependencies. Check if you have these capabilities available.

**Convention**: If a ref points to a file in docs/, treat it as input. If it points to a config file
or binary, treat it as a verification target. When in doubt, read it — more context is always better.

Humans should add their requirements, aspirations, branding guidelines, and tech preferences as refs
pointing to docs/ files. The refs section is a reading list for anyone (human or agent) who wants to
understand the project's intent.

## Operating Principles

1. **Repository knowledge is the system of record.** Prefer adding/maintaining small, discoverable repo docs over long chat explanations. Keep MAP.md updated.
2. **Humans steer, agents execute.** Ask questions when required, but default to making progress by scaffolding, implementing, running checks, and iterating with feedback loops.
3. **Agent legibility is the goal.** Structure code/docs so an agent can reliably reason about it. Minimize assumptions. Prefer typed/validated boundaries and explicit conventions.
4. **Encode golden principles into the repo.** Mechanical rules (formatting, lint, directory conventions, invariants) must be enforced continuously. Treat cleanup like garbage collection, not a one-off refactor.
5. **Small increments.** Work in PR-sized changes. Keep changes narrow. Keep main green.
6. **Parallelize wherever possible.** Use subagents. Partition workstreams by folder/concern. Avoid collisions with explicit scopes.

## Decision Making

Agents are trusted to make decisions autonomously. Technology choices, architecture, design, library selection — decide, record, and keep building. The human can always revert via git if they disagree.

**What to do for each type of decision:**

| Decision | Action |
|----------|--------|
| Technology stack, database, framework, architectural pattern | Decide and write an ADR in docs/adr/ |
| Library choice within a stack, naming conventions, file structure | Decide and write an ADR if non-obvious |
| Code organization, variable names, implementation details | Just code it |

**ADR format** — keep them short (docs/adr/NNN-title.md):

1. **Context**: What situation prompted this decision? (2-3 sentences)
2. **Decision**: What did you decide? (1 sentence)
3. **Consequences**: What are the trade-offs? (2-3 bullet points)

**Constraints from refs**: Before making decisions, read all refs that point to documentation (docs/ files, URLs). If a ref contains explicit constraints ("use PostgreSQL", "use Tailwind"), follow them. If no ref constrains the choice, you decide.

## Draft Task Protocol

Draft tasks are how agents signal platform gaps — NOT how agents ask for permission or decompose work.

**When to create a draft task:**
- You discover a gap in maslow's verification capabilities that prevents you from confidently verifying what you've built (e.g., "can't test auth flows because variable capture isn't implemented in contracts")
- You need a tool or MCP capability that isn't available (e.g., "need browser MCP for visual regression testing")
- You discover a harness limitation that blocks the workflow (e.g., "task convention doesn't support cross-package dependencies")

**When NOT to create a draft task:**
- To decompose your current work — just do the work
- To ask permission for a technology choice — make the choice, write an ADR
- To propose refactoring or improvements — write an ADR or just do it

**Format**: Create the task in docs/tasks/ with status: draft and tag it:
- `kind:gap` — for verification or harness capability gaps
- `kind:capability` — for missing tools, MCPs, or access

The human reviews draft tasks at their own pace and promotes important ones to todo.

## Capability Discovery

At the start of a session, inventory what MCPs and tools are available to you. Check refs in
maslow.yaml with `kind: mcp` — these declare capabilities the project expects you to have.

**If you have the capability**: use it. A browser MCP means you can do visual testing.
A database MCP means you can seed test data. A deployment MCP means you can ship to staging.

**If you lack a required capability**: create a draft task tagged `kind:capability` describing
what you need and why. Example: "Need browser MCP for visual regression testing of the
dashboard — cannot verify layout matches design spec without screenshot comparison."

**If you lack an optional capability**: note it and work around it. Document the gap in your
commit messages so the human knows what was skipped.

This inventory directly affects what verifications are possible. More capabilities = more
confidence in verification results.

## Non-Negotiable Behaviors

- All requirements must be captured in docs and enforced via maslow.yaml.
- Log and document everything: decisions, questions, conventions, and current state. Keep it in-repo.
- If you encounter a large, unscoped question or unknown requirement:
  (a) write a template doc into docs/templates/ with questions and placeholders,
  (b) ask the user to fill it,
  (c) continue only on work that does not depend on the missing info.
- Run checks frequently and use failures as feedback loops.
- Each material decision must be captured as an ADR in docs/adr/.

## Task System

Tasks are how humans inject work into the project via git. Full convention: docs/tasks/CONVENTION.md.

### Quick Reference

- Tasks live in docs/tasks/<id>_<SLUG>.md with YAML frontmatter
- **Scan frontmatter only** to find actionable work — do not read the full body until you've chosen a task
- Only pick up tasks with status: todo and empty assigned_to
- Claim by setting status: in_progress, assigned_to, assigned_at — then commit and push
- If push fails (someone else claimed it), pull and pick another task
- When done, set status: done and commit

### Responding to Task Prompts

When asked to "implement the next task": scan docs/tasks/ for the lowest-ID todo task with no unresolved dependencies.

When asked to "implement task N": go directly to docs/tasks/N_*.md.

When asked to "implement any task tagged X": scan frontmatter for matching tags.

### Status Lifecycle

draft -> todo -> in_progress -> done (with blocked as a side state)

- **Never** start work on draft tasks — those are human works-in-progress
- **Always** claim before starting work (commit + push the status change)

## Process for New Work

1. Read the goal. Load maslow.yaml, docs/MAP.md, docs/PLAN.md, and relevant code. Identify what exists vs what needs to be built.
2. Ask blocking questions upfront. For non-blocking questions, state your default and proceed. For big unscoped questions, write a template to docs/templates/.
3. Create a task list with concrete, ordered tasks and dependencies.
4. Launch parallel workstreams (docs, tests, research via subagents) while handling core implementation.
5. Build depth-first, smallest kernel first. Each increment must compile, pass tests, and not break existing functionality.
6. Run maslow verify --profile quick frequently during development.
7. Before declaring done, audit against docs/PLAN.md exit criteria line by line. Flag gaps honestly.
8. Encode new decisions as ADRs. Update maslow.yaml and docs/MAP.md if the architecture changed.
9. Commit narrowly with focused messages.

## Key Paths

| Path | Purpose |
|------|---------|
| maslow.yaml | Project spec — source of truth |
| CLAUDE.md | Agent guide — this file |
| docs/MAP.md | Architecture overview |
| docs/PLAN.md | Milestones and execution plan |
| docs/adr/ | Architecture Decision Records |
| docs/templates/ | Decision templates for unscoped questions |
| docs/tasks/ | Human-authored tasks with frontmatter metadata |
| docs/tasks/CONVENTION.md | Task format, lifecycle, and agent protocol |
| reports/ | Generated verification output (gitignored) |

## Conventions

- Deterministic output required: the same input must always produce the same result
- All error messages must reference file paths and relevant context
- Exit codes: 0 = success, non-zero = failure

## Verification

Run verification frequently:

`bash
# Quick checks during development
maslow verify --profile quick

# Full checks before merging
maslow verify --profile full

# Validate the spec itself
maslow validate maslow.yaml
`

## Progressive Verification

As you build, add corresponding verifications to maslow.yaml. Don't wait until the end — verify as you go.

| When you... | Add to maslow.yaml |
|-------------|-------------------|
| Create a new API endpoint | Add an HTTP contract scenario for it |
| Build a CLI command | Add a CLI contract scenario for it |
| Produce a build artifact | Add an artifact_size budget for it |
| Implement a performance-sensitive path | Add a performance budget for it |
| Add a new dependency or config file | Add it to refs |
| Create a file that should never be modified by agents | Add it to policy.deny or policy.protected |

Use what the schema can express today. When you hit something you can't express (e.g., need variable capture for auth flows, need database assertions), create a draft task tagged kind:gap describing the verification gap. Keep building — verify what you can, document what you can't.

## Adding a New Feature

1. Check docs/PLAN.md for the relevant milestone
2. Update maslow.yaml if checks, contracts, or budgets change
3. Implement the feature
4. Add or update tests
5. Run maslow verify --profile quick
6. Write an ADR in docs/adr/ if the change involves a material decision
7. Update docs/MAP.md if architecture changed
8. Commit narrowly with focused messages

## Harness Propagation Rule

**All improvements to the agentic harness (CLAUDE.md structure, docs/ conventions, task system,
operating principles, scaffold templates) MUST be propagated into the harness generated by
maslow scaffold.** This ensures every new project benefits from lessons learned.

When you improve any of these: CLAUDE.md content, task conventions, MAP.md/PLAN.md templates,
operating principles, or agent workflow conventions — you must also update the scaffold code
so that maslow scaffold generates the improved version.
