# Remote access, audit, and operability

**Status:** Draft — ADR-0006 accepted; SEC-001 transport slice complete

## Why

Remote deployment multiplies the attack surface: MCP-client identity must stay
separate from downstream MQ credentials, sensitive operations must be
attributable, and operators need health and metrics to run the server at all.

## Outcome

Remote clients authenticate to the server under the approved MCP
authorization model, every sensitive operation produces a payload-safe audit
event, and the server exposes health, readiness, metrics, and structured logs.

## In scope

- Streamable HTTP hardening: authorization, rate limits, request size limits,
  timeouts, and concurrency bounds.
- Documented stdio trust assumptions.
- Audit events correlating client/session, profile, policy decision,
  operation, target, outcome, and latency.
- Health and readiness endpoints, Prometheus metrics, and structured logs;
  OpenTelemetry traces as a candidate.

## Out of scope

- Downstream MQ authentication (`connection-profiles-and-policy`).
- Passing client bearer tokens through to MQ — permanently.

## Success signals

- Security tests cover confused-deputy, secret leakage, oversized payloads,
  and unauthorized profile access.
- An operator can answer "who did what, against which profile, and did it
  work" from audit output alone.
- Probes reflect real readiness without hammering queue managers.

## Dependencies

- FND-001 and POL-001.
- ADR-0006 (MCP transports and client authorization).

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-007.
