# EPIC-004 — Typed inspection and token-conscious output

**Status:** Open; depends on EPIC-003  
**Authority:** `openspec/changes/typed-inspection-and-output/proposal.md`

## Outcome

Models inspect queue managers, MQ objects, and diagnostics through typed,
bounded, schema-backed tools with measurably efficient renderings.

## Stories

- [INS-001](INS-001.md) — Discovery, queues, and queue status.
- [INS-002](INS-002.md) — Channels, listeners, and subscriptions.
- [INS-003](INS-003.md) — Reason-code and connectivity diagnostics.
- [OUT-001](OUT-001.md) — Output schemas, bounds, and rendering benchmarks.

## Acceptance

- One shared collection contract (filter, limit, cursor, truncation) is used
  by every inspection tool.
- The change proposal's success signals hold with recorded evidence.

## Session log

### 2026-08-05

**Done:** Split inspection by object family and moved output efficiency under
the same epic so the collection contract is defined once.  
**Next:** Draft ADR-0005 alongside the first INS-001 schemas.  
**Do not:** Give any tool an unbounded default result.  
**Blocked:** Technical plan depends on CON-001 and POL-001.
