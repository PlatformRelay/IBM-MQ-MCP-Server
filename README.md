# IBM MQ MCP Server

An MCP server for safely inspecting and operating multiple IBM MQ environments.

[![CI](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/ci.yaml/badge.svg)](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/ci.yaml)
[![Docs](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/docs.yaml/badge.svg)](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/docs.yaml)
[![CodeQL](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/codeql.yaml/badge.svg)](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/actions/workflows/codeql.yaml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/PlatformRelay/IBM-MQ-MCP-Server/badge)](https://securityscorecards.dev/viewer/?uri=github.com/PlatformRelay/IBM-MQ-MCP-Server)
[![License: MIT](https://img.shields.io/github/license/PlatformRelay/IBM-MQ-MCP-Server)](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/PlatformRelay/IBM-MQ-MCP-Server)](https://pkg.go.dev/github.com/platformrelay/ibm-mq-mcp-server)

## Status

Bootstrap intent accepted. The Go module skeleton and minimal MCP server
(no MQ tools yet) start at [FND-001](agent-context/stories/FND-001.md).
License and OSS maturity posture are recorded in
[ADR-0009](docs/adr/0009-license-and-oss-maturity.md).

```bash
task test && task build
task run   # stdio MCP server (no ops HTTP by default)

# Optional ops listener (health, readiness, metrics)
IBM_MQ_MCP_OPS_ADDR=:9090 task run
```

Operational endpoints (`/healthz`, `/readyz`, `/metrics`) bind only when
`--ops-addr` or `IBM_MQ_MCP_OPS_ADDR` is set. See [AGENTS.md](AGENTS.md).

Sources of truth:

- [Bootstrap proposal](openspec/changes/bootstrap-mq-mcp/proposal.md)
- [Roadmap](agent-context/roadmap.md)
- [Architecture ADRs](docs/adr/README.md)
- [Design questions](docs/product/design-questions.md)

## Intended outcomes

- Connect to multiple queue managers through named connection profiles.
- Enforce least-privilege capabilities independently for every profile.
- Offer safe, typed MCP tools for common IBM MQ operations.
- Work with queue managers managed by MKurator, the IBM MQ Operator, or neither.
- Return concise, schema-backed results with explicit pagination and truncation.
- Ship with production-grade tests, documentation, observability, and releases.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md), [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md),
and [GOVERNANCE.md](GOVERNANCE.md). Security reports: [SECURITY.md](SECURITY.md).
Help channels: [SUPPORT.md](SUPPORT.md).

## License

[MIT](LICENSE) © 2026 Konrad Heimel
