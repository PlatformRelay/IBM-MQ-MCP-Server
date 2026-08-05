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
✅ Done · Integrated to `main` @ 6fd2fa5 — profile catalog, env/file secrets, TLS, client pool.
[Story](stories/CON-001.md)

### CON-002 — Additional secret providers
🔄 In flight · lane/con-002 — Kubernetes `k8s:` secret references.
[Story](stories/CON-002.md)

### POL-001 — Enforce per-profile capabilities
✅ Done · Integrated to `main` @ cb6cb5f.
[Story](stories/POL-001.md)

### POL-002 — Object-name allow/deny constraints
🛑 Blocked · ADR-0003 deferred per-object allow/deny to post-v0; implement
after POL-001.
[Story](stories/POL-002.md)

## EPIC-004 — Typed inspection and token-conscious output

[Epic](stories/EPIC-004.md) ·
[Change](../openspec/changes/typed-inspection-and-output/proposal.md)

### INS-001 — Inspect profiles, queue managers, and queues
✅ Done · Integrated to `main` @ d5688bd.
[Story](stories/INS-001.md)

### INS-002 — Inspect channels, listeners, and subscriptions
✅ Done · Integrated to `main` @ 1948838.
[Story](stories/INS-002.md)

### INS-003 — Reason-code and connectivity diagnostics
✅ Done · Offline explain + connectivity check tools @ ce1691e.
[Story](stories/INS-003.md)

### OUT-001 — Produce token-conscious structured results
✅ Done · Integrated to `main` @ 34bf044.
[Story](stories/OUT-001.md)

## EPIC-005 — Safe messaging

[Epic](stories/EPIC-005.md) ·
[Change](../openspec/changes/safe-messaging/proposal.md)

### MSG-001 — Prove mqweb message semantics and ship bounded browse
✅ Done · Integrated to `main` @ 06b02a2.
[Story](stories/MSG-001.md)

### MSG-002 — Validated message production
✅ Done · Integrated to `main` @ 18dc5a8.
[Story](stories/MSG-002.md)

### MSG-003 — Separately gated destructive consume
✅ Done · Integrated to `main` @ 1d3000b.
[Story](stories/MSG-003.md)

## EPIC-006 — Guarded administration

[Epic](stories/EPIC-006.md) ·
[Change](../openspec/changes/guarded-administration/proposal.md)

### ADM-001 — Typed queue administration
✅ Done · Integrated to `main` @ 1cd331d.
[Story](stories/ADM-001.md)

### ADM-002 — Administer channels, channel authentication, authority records
✅ Done · Integrated to `main` @ 9df96c3.
[Story](stories/ADM-002.md)

### ADM-003 — Raw MQSC exceptional gate
✅ Done · ADR-0008 accepted; lane/adm-003 @ e443ef9.
[Story](stories/ADM-003.md)

## EPIC-007 — Remote access, audit, and operability

[Epic](stories/EPIC-007.md) ·
[Change](../openspec/changes/remote-access-and-operability/proposal.md)

### SEC-001 — Secure remote MCP transport and hardening
✅ Done · ADR-0006; lane/sec-001 @ 81a7f59.
[Story](stories/SEC-001.md)

### SEC-002 — Payload-safe audit trail
✅ Done · lane/sec-002 @ 1e871ea.
[Story](stories/SEC-002.md)

### OBS-001 — Health, readiness, metrics, and structured logs
✅ Done · Integrated to `main` @ 7f1c256.
[Story](stories/OBS-001.md)

## EPIC-008 — Coexistence and adoption

[Epic](stories/EPIC-008.md) ·
[Change](../openspec/changes/mkurator-coexistence/proposal.md)

### INT-001 — Coexist with MKurator
✅ Done · Integrated to `main` @ 36db0c1.
[Story](stories/INT-001.md)

### DOC-001 — Deliver documentation and operator experience
✅ Done · Integrated to `main` @ 9efa2f9.
[Story](stories/DOC-001.md)
