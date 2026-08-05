# Safe message browse, produce, and consume

**Status:** Draft — blocked by message-safety decisions and a mqweb feasibility spike

## Why

Message access is the highest-risk read path (payload exposure) and write path
(production traffic). mqweb Messaging REST semantics for non-destructive
browse, retrieval, and payload formats are unproven for this contract and must
be validated before any semantics are promised.

## Outcome

Approved profiles browse, produce, and destructively consume bounded messages
under three separate capabilities, with server-enforced count and size limits,
explicit truncation and encoding metadata, and no payload leakage into logs.

## In scope

- Feasibility spike against live mqweb: browse destructiveness, get semantics,
  supported payload formats, and practical limits.
- Non-destructive browse as the first slice.
- Produce with validated content types as the second slice.
- Destructive consume, separately gated, as the last slice.
- Payload opt-in defaults, redaction hooks, and deterministic binary handling.

## Out of scope

- Transaction, syncpoint, exactly-once, or throughput claims (ADR-0002).
- Continuous bridging or bulk queue draining.

## Success signals

- Tests prove browse leaves queue depth unchanged, or browse is not shipped.
- A denied put or get makes no MQ call.
- Payload bytes never reach logs; truncation is always explicit.

## Dependencies

- POL-001, FND-004 (live MQ environment for the spike).
- Design questions 14–17 and ADR-0003.

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-005.
