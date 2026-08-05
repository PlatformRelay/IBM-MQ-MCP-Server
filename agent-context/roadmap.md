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
✅ Done · Integrated to `main` @ 5edb34c.
[Story](stories/FND-001.md)

### FND-002 — Enforce CI quality gates and supply-chain checks
✅ Done · Integrated to `main` @ 99eaa5d.
[Story](stories/FND-002.md)

### FND-003 — Container packaging and release automation
✅ Done · Integrated to `main` @ 1fe36f8.
[Story](stories/FND-003.md)

### FND-004 — Live IBM MQ development and e2e environment
✅ Done · Integrated to `main` @ 8bcf962 — MKurator Kind reuse, opt-in e2e, Developers license documented.
[Story](stories/FND-004.md)

## EPIC-003 — Connection profiles and capability policy

[Epic](stories/EPIC-003.md) ·
[Change](../openspec/changes/connection-profiles-and-policy/proposal.md)

### CON-001 — Resolve multiple secure MQ connection profiles
✅ Done · lane/con-001 — profile catalog, env/file secrets, TLS, client pool stubs.
[Story](stories/CON-001.md)

### CON-002 — Additional secret providers
🛑 Blocked · Requires CON-001.
[Story](stories/CON-002.md)

### POL-001 — Enforce per-profile capabilities
⬜ Open · ADR-0003 Accepted; ADR-0004 Accepted for secret providers.
[Story](stories/POL-001.md)

### POL-002 — Object-name allow/deny constraints
🛑 Blocked · ADR-0003 deferred per-object allow/deny to post-v0; implement
after POL-001.
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
✅ Done · Offline explain + connectivity check tools.
[Story](stories/INS-003.md)

### OUT-001 — Produce token-conscious structured results
⬜ Open · Benchmarks need INS-001 schemas; feeds ADR-0005.
[Story](stories/OUT-001.md)

## EPIC-005 — Safe messaging

[Epic](stories/EPIC-005.md) ·
[Change](../openspec/changes/safe-messaging/proposal.md)

### MSG-001 — Prove mqweb message semantics and ship bounded browse
⬜ Open · FND-004 harness ready; still requires design questions 14–16.
[Story](stories/MSG-001.md)

### MSG-002 — Validated message production
⬜ Open · lane/msg-002 — named content types, size limits, `put_queue_message`.
[Story](stories/MSG-002.md)

### MSG-003 — Separately gated destructive consume
⬜ Open · lane/msg-003 — `consume_queue_messages`, mqweb DELETE, separate capability.
[Story](stories/MSG-003.md)

## EPIC-006 — Guarded administration

[Epic](stories/EPIC-006.md) ·
[Change](../openspec/changes/guarded-administration/proposal.md)

### ADM-001 — Typed queue administration
🛑 Blocked · ADR-0003 Accepted (`administer` distinct from `produce`); still
requires ADR-0007.
[Story](stories/ADM-001.md)

### ADM-002 — Administer channels, channel authentication, authority records
🛑 Blocked · Requires ADM-001.
[Story](stories/ADM-002.md)

### ADM-003 — Raw MQSC exceptional gate
✅ Done · ADR-0008 accepted; lane/adm-003.
[Story](stories/ADM-003.md)

## EPIC-007 — Remote access, audit, and operability

[Epic](stories/EPIC-007.md) ·
[Change](../openspec/changes/remote-access-and-operability/proposal.md)

### SEC-001 — Secure remote MCP transport and hardening
✅ Done · ADR-0006; lane/sec-001 @ 81a7f59.
[Story](stories/SEC-001.md)

### SEC-002 — Payload-safe audit trail
✅ Done · lane/sec-002.
[Story](stories/SEC-002.md)

### OBS-001 — Health, readiness, metrics, and structured logs
✅ Done · Integrated to `main` @ 7f1c256.
[Story](stories/OBS-001.md)

## EPIC-008 — Coexistence and adoption

[Epic](stories/EPIC-008.md) ·
[Change](../openspec/changes/mkurator-coexistence/proposal.md)

### INT-001 — Coexist with MKurator
🛑 Blocked · Requires ADR-0007 ownership and mutation decisions.
[Story](stories/INT-001.md)

### DOC-001 — Deliver documentation and operator experience
✅ Done · Integrated to `main` @ 9efa2f9.
[Story](stories/DOC-001.md)
