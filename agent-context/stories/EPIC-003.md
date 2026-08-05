# EPIC-003 — Connection profiles and capability policy

**Status:** Blocked by ADR-0003 and ADR-0004  
**Authority:** `openspec/changes/connection-profiles-and-policy/proposal.md`

## Outcome

Every remote operation names a validated profile and passes deny-by-default
capability evaluation before secrets are resolved or IBM MQ is contacted.

## Stories

- [CON-001](CON-001.md) — Profile catalog, validation, TLS, env/file secrets.
- [CON-002](CON-002.md) — Additional secret providers.
- [POL-001](POL-001.md) — Deny-by-default capability enforcement.
- [POL-002](POL-002.md) — Object-name allow/deny constraints.

## Acceptance

- The change proposal's success signals hold with recorded evidence.
- No downstream I/O occurs for any denied or unconfigured operation.

## Session log

### 2026-08-05

**Done:** Promoted connection and policy work to an epic with its own change
proposal; split secret providers and object constraints into their own slices.  
**Next:** Resolve ADR-0003, then ADR-0004.  
**Do not:** Let capability vocabulary be defined anywhere except ADR-0003 once
accepted.  
**Blocked:** ADR-0003, ADR-0004.
