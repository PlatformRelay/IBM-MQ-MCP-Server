# Remote access and operability design

**Status:** Active — ADR-0006 accepted; SEC-001 transport slice landed; SEC-002 audit v0 landed.

Must fix the audit event schema, the audit failure policy (fail-open versus
fail-closed), metric names and labels (no high-cardinality or secret labels),
and the separation between MCP transport and operational endpoints.

## Audit (SEC-002 — landed)

- **Schema:** `internal/observability/audit` — allowlisted fields only; no credentials/payloads.
- **Sink v0:** structured slog JSON on stderr (`msg=audit`).
- **Failure policy:** fail-open (sink errors never block MCP/MQ); documented in `docs/observability.md`.
- **Correlation:** `correlationId` on context joins policy decisions and operation outcomes.
- **Coverage:** policy deny + browse/produce/consume/admin mutations + raw MQSC.

## Transport (SEC-001 — landed)

- **stdio primary** with documented host trust assumptions.
- **Streamable HTTP opt-in** via `--remote-addr` / `IBM_MQ_MCP_REMOTE_ADDR`.
- **Bearer gate** from server-configured `env:` / `file:` ref; Authorization
  stripped before MCP handler; never forwarded to mqweb.
- Abuse limits on remote listener: body size, rate, concurrency, HTTP timeouts.
- Implementation: `internal/adapter/remotemcp` (distinct from `opshttp`).

## Remaining slices

- SEC-002 — payload-safe audit trail (event schema + failure policy).
- OBS-001 — remainder: structured log argument sanitization, OpenTelemetry candidate.
