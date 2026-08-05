# Quickstart

Run the IBM MQ MCP server locally over **stdio** (the default transport). No IBM
MQ connection is required for the bootstrap skeleton — profile loading and MQ
tools arrive in later stories ([CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md),
[POL-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/POL-001.md)).

## Prerequisites

- Go **1.24+** (see [`go.mod`](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/go.mod))
- [Task](https://taskfile.dev/) (optional but recommended — mirrors CI)

## Build and run

```bash
git clone https://github.com/PlatformRelay/IBM-MQ-MCP-Server.git
cd IBM-MQ-MCP-Server
task build
./bin/ibm-mq-mcp
```

Or run directly without installing a binary:

```bash
task run
```

The server speaks MCP over **stdin/stdout**. Connect it from an MCP host (for
example a desktop client or IDE extension) using stdio transport configuration.

## Optional ops HTTP

Health, readiness, and Prometheus metrics bind on a **separate** listener when
enabled — they are not exposed on the MCP stdio transport. See
[Observability](observability.md).

```bash
IBM_MQ_MCP_OPS_ADDR=:9090 task run
# or
go run ./cmd/ibm-mq-mcp --ops-addr :9090
```

Probe `http://127.0.0.1:9090/healthz` while the process runs.

## Verify locally

```bash
task check          # Go CI gate matrix
pip install -r docs/requirements-docs.txt
mkdocs build --strict
```

Details: [CI gates](development/ci-gates.md) and
[AGENTS.md](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/AGENTS.md).

## Next steps

| Topic | Page |
| --- | --- |
| Configuration (provisional) | [Configuration](configuration.md) |
| Authentication (TBD — ADR-0004/0006) | [Authentication](authentication.md) |
| Capability policy (TBD — ADR-0003) | [Policy](policy.md) |
| Example profiles | [Examples](examples/README.md) |
| Release binary or container | [Deployment](deployment.md) |
