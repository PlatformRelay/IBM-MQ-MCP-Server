# ADR-0002: Use mqweb REST for first-release IBM MQ connectivity

**Status:** Accepted  
**Date:** 2026-08-04

## Context

IBM MQ can be accessed through mqweb administrative and messaging REST APIs or
through native MQ interfaces such as MQI/PCF. IBM's MCP sample and MKurator both
prove mqweb administrative access. Native access offers richer MQ semantics but
introduces IBM MQ client libraries, CGO, packaging, licensing, platform, and
test-environment complexity.

The first release needs a tractable security and delivery boundary without
preventing later support for environments where mqweb is unavailable.

## Decision

Use IBM MQ mqweb REST APIs as the only downstream connectivity adapter in the
first release.

- Administrative inspection and approved changes use the Administrative REST
  API, including typed MQSC translation internal to the adapter where needed.
- Approved message browse, get, put, and publish operations use the Messaging
  REST API only where its semantics meet the product contract.
- Application code depends on typed administration and messaging ports, never
  HTTP, MQSC strings, or mqweb response shapes.
- PCF/MQI/native connectivity is explicitly deferred to a future ADR and
  adapter.
- The first-release binary remains CGO-free.

## Consequences

### Positive

- Pure Go builds, small images, and cross-platform packaging remain practical.
- HTTPS fits common network policy and is straightforward to mock and test.
- The design aligns with the IBM sample and can reuse proven MKurator mqweb
  knowledge without depending on its controllers.
- A transport adapter seam preserves future PCF/native support.

### Negative

- mqweb must be installed, enabled, reachable, and appropriately authorized.
- REST is unsuitable for high-throughput messaging and does not provide MQI
  transaction/syncpoint semantics.
- Message format, selector, and advanced property support are more limited than
  native MQ clients.
- Standalone mqweb and local/full-install mqweb capabilities differ.
- Some first-release feature requests may need to be deferred rather than
  approximated unsafely.

## Guardrails

- Document supported IBM MQ/mqweb versions and deployment modes.
- Never claim exactly-once, transactional, or high-throughput message behavior.
- Prove non-destructive browse and supported payload formats against live MQ.
- Map REST and MQ reason codes into typed domain errors.
- Keep raw HTTP and MQSC response content out of default logs and user-facing
  errors.
- Do not add a native dependency as an implementation shortcut; supersede this
  ADR explicitly.

## Alternatives

### mqweb and native adapters in the first release

Rejected because it doubles connectivity, packaging, security, and end-to-end
test scope before the MCP contract is stable.

### Native PCF/MQI first

Rejected because it conflicts with the selected CGO-free operational baseline
and adds avoidable environment complexity.

