# ADR-0006: MCP transports and client authorization

**Status:** Accepted  
**Date:** 2026-08-05

## Context

Design question 12 asks whether remote Streamable HTTP is a first-release
deployment target and which MCP-client authentication model is required.
EPIC-007 (SEC-001, SEC-002, OBS-001) depends on this decision. OBS-001 already
ships ops HTTP (health, readiness, metrics) on a **separate listener** from MCP
stdio; remote MCP must not conflate operational endpoints with the MCP transport.

Downstream MQ authentication is settled in ADR-0004 (profile-bound basic/mTLS).
MCP-client identity must remain **orthogonal**: a bearer token at the MCP edge
must never become mqweb credentials (confused-deputy risk).

## Decision

### stdio is the primary v0 transport

- **Default:** MCP over **stdio** — the supported local/integration path.
- **Trust model:** the MCP host OS/process boundary controls who may spawn the
  server. Any user or process that can invoke `ibm-mq-mcp` inherits access to
  every profile and capability in the loaded catalog. Document this explicitly;
  do not imply stdio is multi-tenant safe.
- Stdio remains available even when optional remote HTTP is enabled (for
  side-by-side local clients), unless `--stdio=false` is set for remote-only
  deployments.

### Streamable HTTP is opt-in only

- Remote MCP is **disabled by default**. Enable with `--remote-addr` or
  `IBM_MQ_MCP_REMOTE_ADDR`.
- Uses the official MCP Go SDK **Streamable HTTP** handler (`Stateless: true`
  for v0) on a **dedicated listener** — never on the ops HTTP mux
  (`--ops-addr` / `IBM_MQ_MCP_OPS_ADDR`).
- Remote-only mode: `--stdio=false` with a configured remote listen address.

### MCP-client authentication (remote only)

When remote HTTP is enabled, **every** MCP request must present a
server-configured **bearer token**:

| Setting | Example |
| --- | --- |
| Flag | `--remote-auth-token-ref env:IBM_MQ_MCP_REMOTE_TOKEN` |
| Env | `IBM_MQ_MCP_REMOTE_AUTH_TOKEN_REF=file:/run/secrets/mcp/token` |

- Token resolution uses the same `env:` / `file:` scheme as ADR-0004 secrets.
- Clients send `Authorization: Bearer <token>`.
- The gate validates with constant-time comparison, then **strips**
  `Authorization` before the MCP handler runs so client tokens **never** reach
  the mqweb adapter.
- **No token passthrough** to IBM MQ — permanently out of scope.
- Client identity **cannot** select, override, or substitute downstream MQ
  credentials; profile choice remains an explicit tool argument bounded by
  catalog policy (ADR-0003).

OAuth, mTLS-at-ingress, and per-client token rotation are deferred to a later
ADR unless required by a specific deployment pattern.

### Abuse limits (remote only)

When remote HTTP is enabled, apply:

| Control | Default |
| --- | --- |
| Max request body | 1 MiB |
| Rate limit | 20 req/s, burst 40 |
| Max concurrency | 32 in-flight requests |
| HTTP timeouts | Read header 5s, read 30s, write 60s, idle 120s |

Ops HTTP keeps its existing limits; do not reuse the ops mux for MCP traffic.

### Startup validation

- If `--remote-addr` / env is set, `--remote-auth-token-ref` (or env equivalent)
  is **mandatory** — refuse to start otherwise.
- Empty or missing resolved token is a fatal startup error.

## Consequences

### Positive

- SEC-001 can land hardened remote transport without blocking stdio-first users.
- Clear separation: MCP gate token vs profile MQ credentials vs ops probes.
- Confused-deputy class mitigated by stripping client Authorization at the edge.
- Remote surface is explicit; accidental HTTP exposure without auth fails closed
  at startup.

### Negative

- Single shared bearer token is coarse-grained — sufficient for v0 trusted
  networks; fine-grained MCP-client ACLs need a follow-on ADR.
- Stateless Streamable HTTP limits server-initiated MCP messages until session
  mode is evaluated.
- Operators must manage two optional listeners (ops + remote MCP) plus stdio.

## Alternatives considered

| Alternative | Rejected because |
| --- | --- |
| Remote HTTP on by default | Violates stdio-primary posture; expands attack surface without opt-in |
| Pass client bearer to mqweb | Confused-deputy; violates identity separation |
| OAuth in v0 | Undecided IdP integration; bearer gate unblocks SEC-001 |
| MCP on ops HTTP mux | Couples observability and MCP attack surfaces (OBS-001 boundary) |
| mTLS-only remote gate | Heavier cert ops; bearer + network policy enough for v0 opt-in |

## References

- [SEC-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/SEC-001.md)
- [Authentication](../authentication.md)
- [Threat model](../security/threat-model.md)
- [Design question 12](../product/design-questions.md)
- Implementation: `internal/adapter/remotemcp`
