# Observability

Implemented in [OBS-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/OBS-001.md).
Operational endpoints are **opt-in** and bind **separately** from MCP stdio.

## Enabling ops HTTP

| Mechanism | Example |
| --- | --- |
| Flag | `--ops-addr :9090` |
| Environment | `IBM_MQ_MCP_OPS_ADDR=:9090` |

If neither is set (default), **no ops listener** is started — stdio-only mode.

```bash
IBM_MQ_MCP_OPS_ADDR=:9090 task run
curl -sf http://127.0.0.1:9090/healthz
curl -sf http://127.0.0.1:9090/readyz
curl -sf http://127.0.0.1:9090/metrics | head
```

## Endpoints

| Path | Purpose | MQ contact |
| --- | --- | --- |
| `/healthz` | **Liveness** — process can serve probes | No |
| `/readyz` | **Readiness** — valid bootstrap config and MCP transport serving | No |
| `/metrics` | Prometheus metrics | No |

Readiness reflects configuration validity and MCP transport state. Probes **do
not** call queue managers on every check ([OBS-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/OBS-001.md)
acceptance — avoids probe amplification).

Responses:

- `/healthz` → `200 ok` or `503 unhealthy`
- `/readyz` → `200 ready` or `503 not ready`

## Prometheus metrics

Exposed at `/metrics` on the ops listener only — **never** on the MCP transport.

| Metric | Labels | Description |
| --- | --- | --- |
| `ibm_mq_mcp_requests_total` | `profile` | MCP tool requests handled |
| `ibm_mq_mcp_request_duration_seconds` | `profile` | Request latency histogram |
| `ibm_mq_mcp_policy_denials_total` | `profile` | Policy denials before MQ I/O |

Label cardinality is restricted to **profile name** only. Until profiles land,
the label value is `_none`. No secret, client, queue, or message identifiers in
labels.

## Structured logs

JSON logs on **stderr** via the default slog handler with:

- Central redaction of secret-like field names
- Sanitization of tool-argument values to resist log injection

Log field names for audit and request correlation will expand with
[SEC-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/SEC-002.md).

## OpenTelemetry

Distributed tracing is a **strong candidate** ([feature scope](product/feature-scope.md))
but not implemented in the bootstrap skeleton.

## Related pages

- [Deployment](deployment.md)
- [Troubleshooting](troubleshooting.md)
- [Threat model](security/threat-model.md)
