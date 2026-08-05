# ADR-0001: Use Go and the official MCP Go SDK

**Status:** Accepted  
**Date:** 2026-08-04

## Context

The implementation language determines the repository layout, architecture
boundaries, test tooling, packaging, security posture, and reuse from MKurator.
The leading options were Go with the official MCP Go SDK, Python with FastMCP,
and TypeScript with the official MCP SDK.

IBM's MQ MCP sample uses Python and demonstrates a short path to mqweb, but it is
not a production scaffold. MKurator is implemented in Go and already establishes
domain vocabulary and ports/adapters patterns relevant to this project.

As of this decision, the official Go SDK is Tier 1, supports local and remote
transports, typed tool inputs/outputs, OAuth primitives, and MCP specification
2026-07-28.

## Decision

Implement the server in Go using
`github.com/modelcontextprotocol/go-sdk`.

The codebase will use explicit package boundaries:

- `cmd/` for executable composition only.
- `internal/mcpserver/` for protocol registration and result mapping.
- `internal/application/` for use-case orchestration.
- `internal/policy/` for capability decisions.
- `internal/mqadmin/` and `internal/messaging/` for IBM MQ domain ports.
- `internal/adapter/` for mqweb, secret, configuration, and integration
  adapters.

The executable should remain CGO-free unless a future accepted ADR introduces a
native IBM MQ/PCF adapter requiring otherwise.

## Consequences

### Positive

- A small static binary and hardened minimal container are practical.
- Compile-time types support schema-backed tool contracts and safe domain
  boundaries.
- MKurator concepts and selected non-controller code patterns can be aligned
  without creating a runtime dependency.
- Go's test, race, fuzz, benchmark, profiling, and cross-compilation tooling
  support the intended quality gates.
- The SDK is protocol-owner maintained and currently Tier 1.

### Negative

- IBM's Python sample cannot be directly extended.
- FastMCP's decorator ergonomics and larger set of high-level examples are not
  available.
- Official Go SDK APIs may evolve with the protocol and require deliberate
  upgrade testing.
- Native IBM MQ clients may later force a CGO/toolchain trade-off.

## Alternatives

### Python with FastMCP

Rejected as the baseline. It offers the closest path from IBM's sample and rapid
tool development, but packaging and architectural consistency with MKurator are
weaker for this project.

### TypeScript with the official SDK

Rejected as the baseline. It has strong MCP ecosystem coverage but offers less
reuse and operational alignment with the existing PlatformRelay IBM MQ work.

## Validation implications

- Pin a supported Go release and MCP SDK version.
- Add SDK-level protocol tests for tools and approved transports.
- Run unit tests with the race detector where supported.
- Enforce formatting, static analysis, vulnerability checks, and reproducible
  CGO-free builds.
- Test SDK upgrades against supported MCP protocol versions before release.

