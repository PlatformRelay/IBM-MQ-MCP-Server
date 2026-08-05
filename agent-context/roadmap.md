# Roadmap

Authority: [bootstrap proposal](../openspec/changes/bootstrap-mq-mcp/proposal.md).
Epic-level intent lives in `openspec/changes/<change-id>/proposal.md`. Each
story is one reviewable delivery slice.

## EPIC-001 — Production-grade IBM MQ MCP server (umbrella)

✅ Intent accepted (2026-08-05) · OSS baseline via ADR-0009; delivery continues
through EPIC-002..008.
[Story](stories/EPIC-001.md)

## EPIC-002 — Foundation and delivery

[Epic](stories/EPIC-002.md) ·
[Change](../openspec/changes/bootstrap-mq-mcp/proposal.md)

### FND-001 — Establish the Go module skeleton and minimal MCP server
⬜ Open · Ready (product intent + ADR-0001/0009 accepted).
[Story](stories/FND-001.md)

### FND-002 — Enforce CI quality gates and supply-chain checks
⬜ Open · Starts after FND-001 (extend existing Scorecard/Docs/gitleaks CI).
[Story](stories/FND-002.md)

### FND-003 — Container packaging and release automation
⬜ Open · License/OSS baseline done (ADR-0009); still needs artifact choices
(binary vs container) and cosign/SBOM release wiring from Kollect/MKurator.
[Story](stories/FND-003.md)

### FND-004 — Live IBM MQ development and e2e environment
⬜ Open · Licensing approach undecided; gates the MSG-001 spike.
[Story](stories/FND-004.md)

## EPIC-003 — Connection profiles and capability policy

[Epic](stories/EPIC-003.md) ·
[Change](../openspec/changes/connection-profiles-and-policy/proposal.md)

### CON-001 — Resolve multiple secure MQ connection profiles
🛑 Blocked · Requires ADR-0004; env/file secrets only in this slice.
[Story](stories/CON-001.md)

### CON-002 — Additional secret providers
🛑 Blocked · Requires ADR-0004 and CON-001.
[Story](stories/CON-002.md)

### POL-001 — Enforce per-profile capabilities
🛑 Blocked · Requires ADR-0003 capability semantics.
[Story](stories/POL-001.md)

### POL-002 — Object-name allow/deny constraints
🛑 Blocked · ADR-0003 decides whether this is first release.
[Story](stories/POL-002.md)

## EPIC-004 — Typed inspection and token-conscious output

[Epic](stories/EPIC-004.md) ·
[Change](../openspec/changes/typed-inspection-and-output/proposal.md)

### INS-001 — Inspect profiles, queue managers, and queues
⬜ Open · Intent approved shape; technical plan needs CON-001 and POL-001.
[Story](stories/INS-001.md)

### INS-002 — Inspect channels, listeners, and subscriptions
⬜ Open · Starts after INS-001 fixes the collection contract.
[Story](stories/INS-002.md)

### INS-003 — Reason-code and connectivity diagnostics
⬜ Open · Offline explain tool unblocked; live check needs CON-001.
[Story](stories/INS-003.md)

### OUT-001 — Produce token-conscious structured results
⬜ Open · Benchmarks need INS-001 schemas; feeds ADR-0005.
[Story](stories/OUT-001.md)

## EPIC-005 — Safe messaging

[Epic](stories/EPIC-005.md) ·
[Change](../openspec/changes/safe-messaging/proposal.md)

### MSG-001 — Prove mqweb message semantics and ship bounded browse
🛑 Blocked · Requires design questions 14–16 and the FND-004 spike.
[Story](stories/MSG-001.md)

### MSG-002 — Validated message production
🛑 Blocked · Requires design question 17 and the MSG-001 spike.
[Story](stories/MSG-002.md)

### MSG-003 — Separately gated destructive consume
🛑 Blocked · Requires ADR-0003 and the MSG-001 spike.
[Story](stories/MSG-003.md)

## EPIC-006 — Guarded administration

[Epic](stories/EPIC-006.md) ·
[Change](../openspec/changes/guarded-administration/proposal.md)

### ADM-001 — Typed queue administration
🛑 Blocked · Requires ADR-0003 and ADR-0007.
[Story](stories/ADM-001.md)

### ADM-002 — Administer channels, channel authentication, authority records
🛑 Blocked · Requires ADM-001.
[Story](stories/ADM-002.md)

### ADM-003 — Raw MQSC exceptional gate
🛑 Blocked · Requires ADR-0008.
[Story](stories/ADM-003.md)

## EPIC-007 — Remote access, audit, and operability

[Epic](stories/EPIC-007.md) ·
[Change](../openspec/changes/remote-access-and-operability/proposal.md)

### SEC-001 — Secure remote MCP transport and hardening
🛑 Blocked · Requires ADR-0006 deployment and client-identity decisions.
[Story](stories/SEC-001.md)

### SEC-002 — Payload-safe audit trail
🛑 Blocked · Requires ADR-0006 and POL-001 decision events.
[Story](stories/SEC-002.md)

### OBS-001 — Health, readiness, metrics, and structured logs
⬜ Open · Starts after FND-001.
[Story](stories/OBS-001.md)

## EPIC-008 — Coexistence and adoption

[Epic](stories/EPIC-008.md) ·
[Change](../openspec/changes/mkurator-coexistence/proposal.md)

### INT-001 — Coexist with MKurator
🛑 Blocked · Requires ADR-0007 ownership and mutation decisions.
[Story](stories/INT-001.md)

### DOC-001 — Deliver documentation and operator experience
⬜ Open · MkDocs, examples, runbooks, threat model, version matrix.
[Story](stories/DOC-001.md)
