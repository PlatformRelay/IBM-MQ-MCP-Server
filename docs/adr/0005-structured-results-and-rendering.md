# ADR-0005: Structured results and token-efficient rendering

**Status:** Accepted  
**Date:** 2026-08-05

## Context

MCP tools must return machine-readable results that clients can parse reliably.
INS-001 established a shared collection envelope for list-style inspection tools,
but the rendering strategy for token-conscious hosts (Markdown summaries, TOON,
and similar) was deferred until INS-001 schemas existed and OUT-001 could measure
trade-offs.

MSG-001 adds message browse results that must follow the same contract so
pagination, truncation metadata, and optional payloads stay consistent across
tools.

## Decision

1. **JSON `structuredContent` is canonical.** Every tool result type exposed to
   MCP hosts is defined as a Go struct serialized to JSON in `structuredContent`.
   Text `content` blocks are supplementary only (errors, human hints) — never
   the sole carrier of tool data.

2. **The INS collection envelope is the shared pagination contract.** List-style
   tools — including `browse_queue_messages` — return
   `collection.Page[T]` with `items`, `limit`, cursor fields, and explicit
   `truncated` / `truncationReason` metadata. Message-specific fields (encoding,
   payload opt-in, size caps) extend item types; they do not replace the
   envelope.

3. **Alternate renderings wait on evidence.** Markdown tables, TOON, or other
   token-reduced views may be added only after OUT-001 benchmarks justify them.
   They are derived views of the same canonical JSON — never a replacement for
   MCP JSON-RPC structured results.

## Consequences

- New list tools must reuse `internal/collection` rather than inventing local
  pagination shapes.
- Browse payloads remain opt-in with server-enforced byte limits and redaction
  before serialization (design questions 14–16).
- OUT-001 can add optional rendering layers without breaking existing clients.

## Alternatives considered

- **Markdown-first tool results** — rejected; breaks programmatic clients and
  duplicates schema maintenance.
- **Per-tool pagination structs** — rejected; INS-001 already proved a shared
  envelope works for inspection lists.
