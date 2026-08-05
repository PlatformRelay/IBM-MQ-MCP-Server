# EPIC-007 — Remote access, audit, and operability

**Status:** Blocked by ADR-0006  
**Authority:** `openspec/changes/remote-access-and-operability/proposal.md`

## Outcome

Remote clients are authenticated independently of MQ credentials, sensitive
operations are attributable, and operators can observe the server.

## Stories

- [SEC-001](SEC-001.md) — Remote transport authorization and hardening.
- [SEC-002](SEC-002.md) — Payload-safe audit trail.
- [OBS-001](OBS-001.md) — Health, readiness, metrics, and structured logs.

## Acceptance

- Confused-deputy, leakage, and abuse tests pass.
- The change proposal's success signals hold with recorded evidence.

## Session log

### 2026-08-05

**Done:** Split audit and observability out of the transport-security story;
added the previously missing observability slice.  
**Next:** Decide whether remote HTTP is a first-release target (ADR-0006).  
**Do not:** Pass client bearer tokens through to MQ.  
**Blocked:** ADR-0006 for SEC-001/SEC-002; OBS-001 needs only FND-001.
