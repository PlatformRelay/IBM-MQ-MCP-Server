# ADR-0003: Operation-oriented per-profile capability model

**Status:** Accepted  
**Date:** 2026-08-05

## Context

Design questions 5–9 define how operators express authorization intent on each
connection profile: what “read-only” means for message access, how production
and write paths differ, whether destructive consume is separate from browse,
whether raw MQSC is available, and whether object-level allow/deny lists are
required in the first release.

The `connection-profiles-and-policy` OpenSpec change and EPIC-003 depend on a
stable, deny-by-default vocabulary before POL-001 can implement enforcement.
`docs/architecture/proposed-system.md` hypothesizes operation-oriented
capabilities; this ADR accepts that model and rejects coarser alternatives for
v0.

## Decision

Adopt **operation-oriented capabilities** as the only first-class grants on
each profile:

| Capability | Meaning |
| --- | --- |
| `inspect` | Queue managers, objects, status, depth, and diagnostics. |
| `browse` | Inspect message metadata and optionally payloads without consuming. |
| `consume` | Destructively retrieve messages. |
| `produce` | Put or publish messages. |
| `administer` | Define, alter, or delete MQ objects through typed operations. |
| `execute_mqsc` | Exceptional raw MQSC execution; off by default. |

Rules:

- **Deny by default.** Every tool maps to exactly one required capability; an
  operation without a matching grant is denied before secrets are resolved or
  IBM MQ is contacted.
- **Read-only production** is expressed as `inspect`, optionally plus `browse`
  without default payloads (design questions 5 and 15). It does not imply
  `produce`, `consume`, `administer`, or `execute_mqsc`.
- **`produce` ≠ `administer`.** Message production and object administration are
  distinct grants (design question 6).
- **`consume` ≠ `browse`.** Destructive message retrieval is a separate grant
  from non-destructive browse (design question 7).
- **`execute_mqsc` is exceptional.** Raw MQSC is omitted unless explicitly
  granted; it is not implied by `administer` (design question 8).
- **Per-object allow/deny is deferred** to POL-002 and post-v0 (design question
  9). Profile-level capabilities are sufficient for the first release.

Reject for v0:

- **Option B — coarse profile modes** (`read`, `write`, `admin`): ambiguous for
  browse vs metadata, produce vs administer, and MQSC; harder to audit and to
  express “read prod / write dev” safely.
- **Option C — modes plus explicit overrides:** two configuration surfaces and a
  larger validation and error taxonomy without clear benefit over the chosen
  vocabulary for first release.

## Consequences

### Positive

- Operators can express “read production, write development” with explicit,
  auditable grants.
- Each tool has one required capability, simplifying enforcement tests and
  policy decision events.
- Browse, consume, produce, administer, and raw MQSC are unambiguous in
  configuration and documentation.
- POL-001 can implement deny-by-default evaluation without waiting for object
  name constraints.

### Negative

- Operators must understand six capability names rather than picking a single
  mode.
- “Read-only” is not a shorthand; operators compose `inspect` (+ optional
  `browse`) explicitly.
- Per-object constraints (queues, topics, channels) require a follow-on slice
  (POL-002) when needed.

## Alternatives

### Coarse profile modes (B)

Rejected for v0. Faster to configure initially but ambiguous for the
browse/consume and produce/administer boundaries that design questions 5–8
require to be explicit.

### Modes plus explicit overrides (C)

Rejected for v0. Adds a second configuration surface (mode + overrides) and more
validation paths while the operation-oriented vocabulary already covers simple
and advanced cases.

## Validation implications for POL-001

- Startup validation must reject unknown capability names and empty capability
  lists on profiles that are referenced by configuration.
- Each MCP tool registration must declare its single required capability; CI or
  unit tests must assert the mapping is complete and stable.
- Denied-operation tests must prove no downstream adapter call occurs when the
  required capability is absent.
- Tool discovery must not imply a profile permits an operation it cannot
  perform; capability checks precede secret resolution and IBM MQ I/O.
- Policy decision events must record profile id, required capability, grant
  outcome, and operation identity for SEC-002 audit integration.
