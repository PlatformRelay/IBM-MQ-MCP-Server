# Troubleshooting

## Server won't start

| Symptom | Likely cause | Action |
| --- | --- | --- |
| `ops listener failed` / exit code 1 | Ops address in use or invalid | Change `--ops-addr` / `IBM_MQ_MCP_OPS_ADDR`; check bind permissions |
| MCP host sees no tools | Expected in bootstrap | No MQ tools until INS/MSG/ADM stories land — see [Tool reference](tools/index.md) |
| `task check` fails | Local gate drift | Run failing subtask (`task lint`, `task test:race`, …) — [CI gates](development/ci-gates.md) |

## Ops probes

| Symptom | Check |
| --- | --- |
| `/healthz` returns 503 | Process marked unhealthy in runtime state — inspect stderr logs |
| `/readyz` returns 503 | Bootstrap config invalid (future profiles) or MCP transport not serving |
| Connection refused on `:9090` | Ops HTTP not enabled — set `--ops-addr` or `IBM_MQ_MCP_OPS_ADDR` |
| Metrics empty | No tool traffic yet; profile label `_none` until profiles exist |

Probes intentionally **do not** ping queue managers. MQ connectivity failures
will surface in tool errors once adapters exist — not in `/readyz` today.

## MCP client integration

- Use **stdio** transport pointing at `ibm-mq-mcp` or `task run`.
- Remote Streamable HTTP is **TBD** ([ADR-0006](adr/README.md#decision-queue)).
- Do not point MCP clients at `/metrics` or `/healthz` — those are operator
  endpoints, not MCP.

## IBM MQ / mqweb issues (future)

When mqweb connectivity lands ([CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md),
[ADR-0002](adr/0002-mqweb-first-connectivity.md)):

- Verify mqweb is installed, reachable, and authorized for the profile credential.
- Consult the [version support matrix](support/version-matrix.md) — certified
  combinations are **not yet recorded**.
- Map MQ reason codes via [INS-003](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INS-003.md)
  when available.

## Security incidents

Report vulnerabilities privately per
[SECURITY.md](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/SECURITY.md).

## Getting help

- [SUPPORT.md](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/SUPPORT.md)
- [CONTRIBUTING.md](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/CONTRIBUTING.md)
