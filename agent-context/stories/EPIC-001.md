# EPIC-001 — Production-grade IBM MQ MCP server

**Status:** Documented; product intent accepted (2026-08-05)  
**Authority:** `openspec/changes/bootstrap-mq-mcp/proposal.md`

## Outcome

Deliver the approved proposal through small, independently reviewable stories
without relying on chat history.

## Acceptance

- The proposal's success signals have end-to-end evidence.
- Every supported remote operation requires an explicit profile.
- A read-constrained profile cannot mutate MQ, and a write-capable profile can
  perform only its granted operation families.
- Generic IBM MQ use does not depend on Kubernetes or MKurator.
- Security, reliability, operability, and documentation release gates pass.

## Dependencies and blockers

- Product intent approval.
- ADR-0001 through ADR-0009 as their dependent stories begin.
- Streamsy reference repository.

## Session log

### 2026-08-04

**Done:** Assessed IBM's MQ MCP sample, MCP structured output guidance, MKurator
architecture, and TOON. Drafted proposal, scope, architecture hypothesis,
decision queue, and story map.  
**Next:** Resolve design questions one at a time, beginning with implementation
ecosystem.  
**Do not:** Implement runtime code or treat the proposed architecture as
accepted.  
**Blocked:** Streamsy reference has not been supplied.


### 2026-08-05

**Done:** Developed the epic breakdown: EPIC-002 through EPIC-008 now own the
delivery slices, each with an OpenSpec change proposal. Split oversized
stories (FND, INS, MSG, ADM, SEC) into vertical slices; added missing
observability (OBS-001) and live-MQ environment (FND-004) slices; queued
ADR-0009 for delivery targets and license. Recorded findings in
`docs/architecture/design-audit.md`.  
**Next:** Approve product intent, then the epic change proposals, then start
FND-001.  
**Do not:** Approve implementation stories before their epic's proposal.  
**Blocked:** Product intent approval; Streamsy reference still unavailable.
