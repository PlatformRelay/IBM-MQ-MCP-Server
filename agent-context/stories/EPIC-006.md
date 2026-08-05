# EPIC-006 — Guarded administration

**Status:** Blocked by ADR-0003, ADR-0007, ADR-0008  
**Authority:** `openspec/changes/guarded-administration/proposal.md`

## Outcome

Authorized profiles perform typed, bounded, auditable administrative changes;
raw MQSC is absent or exceptionally gated.

## Stories

- [ADM-001](ADM-001.md) — Typed queue define/alter/delete.
- [ADM-002](ADM-002.md) — Channels, channel authentication, authority records.
- [ADM-003](ADM-003.md) — Raw MQSC exceptional gate per ADR-0008.

## Acceptance

- Every mutation is attributable, targeted, and typed.
- The change proposal's success signals hold with recorded evidence.

## Session log

### 2026-08-05

**Done:** Split administration by object family and isolated the raw MQSC
policy into its own slice behind ADR-0008.  
**Next:** Resolve ADR-0008 scope alongside ADR-0003.  
**Do not:** Ship any mutation before the MKurator pre-mutation hook contract
exists (EPIC-008).  
**Blocked:** ADR-0003, ADR-0007, ADR-0008.
