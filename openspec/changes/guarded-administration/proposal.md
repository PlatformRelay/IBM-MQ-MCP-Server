# Guarded typed administration

**Status:** In progress — ADR-0007 accepted; ADM-001 queue slice landing

## Why

Operators need selected administrative changes without an arbitrary `runmqsc`
tool that collapses discovery, authorization, and execution into one
high-risk surface, and without fighting declarative reconciliation.

## Outcome

Profiles granted `administer` create, alter, and delete supported object types
through validated typed operations; raw MQSC is absent or exceptionally gated
per ADR-0008; MKurator-managed targets follow the approved coexistence policy.

## In scope

- Queue define, alter, and delete as the first slice.
- Channels, channel authentication, and authority records as a second slice.
- Dry-run only where the backend supports a truthful preview.
- Raw MQSC policy implementation per ADR-0008 as its own slice.
- Before/after identifiers and audit metadata on every mutation.

## Out of scope

- Queue-manager lifecycle management.
- Replacing IBM MQ Operator or MKurator reconciliation.
- Bulk unscoped mutations.

## Success signals

- Every mutation names its explicit target and is attributable in the audit
  trail.
- Conflict, not-found, idempotency, and authority failures map to typed
  errors.
- The default configuration exposes no raw MQSC tool.

## Dependencies

- POL-001 and INS-001.
- ADR-0003, ADR-0007 (coexistence behavior), and ADR-0008 (raw MQSC policy).

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-006.
