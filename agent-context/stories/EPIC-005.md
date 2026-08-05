# EPIC-005 — Safe messaging

**Status:** Blocked by message-safety decisions and mqweb feasibility spike  
**Authority:** `openspec/changes/safe-messaging/proposal.md`

## Outcome

Approved profiles browse, produce, and destructively consume bounded messages
under three separate capabilities with no payload leakage.

## Stories

- [MSG-001](MSG-001.md) — Feasibility spike and bounded non-destructive browse.
- [MSG-002](MSG-002.md) — Validated message production.
- [MSG-003](MSG-003.md) — Separately gated destructive consume.

## Acceptance

- Browse is proven non-destructive on live MQ or is not shipped.
- The change proposal's success signals hold with recorded evidence.

## Session log

### 2026-08-05

**Done:** Split messaging by risk class (browse, produce, consume) and made the
mqweb semantics spike the explicit entry gate.  
**Next:** Resolve design questions 14–17; run the spike once FND-004 exists.  
**Do not:** Promise MQI semantics that mqweb Messaging REST cannot deliver.  
**Blocked:** Design questions 14–17; FND-004 for the spike.
