# Bootstrap a production-grade IBM MQ MCP server

**Status:** Accepted — 2026-08-05 (product intent; OSS baseline via ADR-0009)

## Why

Operators need AI-assisted IBM MQ inspection and carefully constrained actions
across multiple environments. IBM's current MCP sample proves mqweb integration
but does not provide multi-profile policy, production security, typed tools, or
a mature project environment.

## Outcome

Create an MCP server that can select named IBM MQ connection profiles and
enforce independently configured capabilities for each profile. It must work
with ordinary queue managers and optionally recognize MKurator-managed objects
without requiring Kubernetes.

## In scope

- Mature repository and delivery foundation.
- Multiple queue-manager connection profiles.
- Authentication and TLS appropriate to supported IBM MQ environments.
- Per-profile operation capabilities.
- Safe typed inspection, message, diagnostic, and selected administration
  operations.
- Token-conscious, schema-backed results.
- Optional MKurator coexistence awareness.
- Testing, observability, security, documentation, packaging, and releases.

## Out of scope

- Queue-manager lifecycle management.
- Continuous broker bridging.
- Generic cross-broker abstractions in the initial architecture.
- Unbounded data extraction.
- Unrestricted administrative execution by default.

## Success signals

- Operators can configure a production profile that cannot mutate MQ and a
  development profile that can produce messages.
- Every forbidden operation is rejected locally before downstream I/O.
- Common operations do not require arbitrary MQSC.
- Results are bounded, machine-validated, and useful to an LLM.
- The same application works against MQ with and without MKurator.
- A local real-MQ environment validates supported operations end to end.

## Dependencies

- Answers to the design questions in `docs/product/design-questions.md`.
- Access to the Streamsy and other requested reference repositories.
- A supported IBM MQ test topology and licensing approach for CI/e2e.

