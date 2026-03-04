# ADR 003: Frontend Architecture

## Context

The MVP frontend needs to display conversations grouped by project, support search, and show expandable/collapsible message blocks with clear visual distinction between user, assistant, and sub-agent messages.

## Decision

Use React with vanilla CSS (no component library or CSS framework). Vite for build tooling. Dark theme. Sidebar + main content layout. No routing library — single-page with state-driven views.

## Consequences

- Minimal dependencies — only react, react-dom, vite
- Fast builds and small bundle size (~200KB gzipped ~62KB)
- Dark theme suits developer tooling context
- Trade-off: no CSS-in-JS or utility classes means manual styling, but keeps bundle small
- Trade-off: no client-side routing means no URL-based navigation (acceptable for MVP)
