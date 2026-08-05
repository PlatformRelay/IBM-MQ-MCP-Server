# Typed inspection and token-conscious output

**Status:** Draft — depends on connection-profiles-and-policy

## Why

Inspection is the first user-visible value and the template for every later
tool contract: typed inputs, output schemas, bounded collections, and compact
text renderings. IBM's sample offers only raw `dspmq` and `runmqsc`.

## Outcome

Models discover profiles and inspect queue managers, queues, channels, and
diagnostics through typed, paginated, schema-backed tools whose results are
measurably token-efficient.

## In scope

- Discovery: list profiles, profile capabilities, connection health.
- Queues: list, get, and status with filters, limits, and cursors.
- Channels, listeners, and subscriptions as a second slice.
- Diagnostics: reason-code explanation and side-effect-free connectivity
  checks.
- Output schemas, conforming `structuredContent`, deterministic text
  rendering, and truncation metadata.
- Token benchmarks for compact JSON, Markdown tables, and TOON.

## Out of scope

- Message payload access (`safe-messaging`).
- Mutating operations (`guarded-administration`).
- MCP resources and prompts until tool contracts stabilize.

## Success signals

- Common inspections need no MQSC and stay inside documented result bounds.
- Every collection reports truncation explicitly; nothing is unbounded.
- Rendering choices are backed by recorded token, correctness, latency, and
  client-compatibility measurements.

## Dependencies

- CON-001 and POL-001.
- ADR-0005 (structured results and token-efficient rendering).

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-004.
