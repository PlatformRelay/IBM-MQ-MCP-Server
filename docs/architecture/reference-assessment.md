# Reference assessment

## IBM MQ MCP sample

[`ibm-messaging/mq-mcp-server`](https://github.com/ibm-messaging/mq-mcp-server)
is IBM-maintained prior art. It is a small Python/FastMCP sample using the IBM MQ
Administrative REST API.

It exposes:

- `dspmq`, which lists queue managers local to one mqweb server.
- `runmqsc`, which sends arbitrary MQSC to one selected queue manager.

Useful patterns:

- mqweb REST avoids native IBM MQ client libraries.
- Streamable HTTP and stdio are both straightforward with an MCP framework.
- IBM MQ REST v3 is a viable first administration adapter.

Gaps this project must address:

- Credentials and endpoint are source-code constants.
- TLS verification is disabled.
- One mqweb endpoint and one credential pair are supported.
- Arbitrary `runmqsc` collapses discovery, authorization, and execution into one
  high-risk tool.
- Broad exceptions become an unstructured success-shaped message.
- No typed outputs, output schemas, pagination, audit trail, tests, or mature
  delivery scaffolding are present.

The sample should be credited and tested as behavioral prior art, not copied as
the production architecture.

## MKurator

[`PlatformRelay/MKurator`](https://github.com/PlatformRelay/MKurator) is a
Kubernetes operator for declarative administration of existing IBM MQ queue
managers. It uses mqweb REST behind an `MQAdmin` port and keeps controllers free
of HTTP/MQSC details.

Patterns to reuse:

- Hexagonal boundary between domain operations and mqweb transport.
- Named queue-manager connection objects referencing external secrets.
- TLS verification by default and custom CA support.
- Typed error taxonomy and structured observability.
- Unit tests through ports, adapter contract tests, and live IBM MQ end-to-end
  tests.
- MkDocs Material, ADRs, explicit NFRs, hardened images, dependency automation,
  and supply-chain checks.

The MCP server must not require MKurator. Optional awareness can identify
MKurator-managed objects and avoid proposing imperative mutations that fight
declarative reconciliation.

## MCP specifications and SDKs

The MCP specification supports schema-backed `structuredContent` and tool
`outputSchema`. The protocol transport remains JSON-RPC; alternative text
encodings do not replace protocol JSON.

Consequences:

- Canonical tool results should be typed JSON objects conforming to output
  schemas.
- Human/model-facing text should be concise and preserve compatibility.
- Large collections need filtering, pagination, field selection, and explicit
  truncation before experimenting with another serialization.
- Tool annotations are useful UX hints but are not authorization controls.
- Mutating actions require server-side policy enforcement and should support
  client confirmation flows.

## TOON

[Token-Oriented Object Notation](https://github.com/toon-format/toon) is a
lossless text encoding of the JSON data model optimized for uniform collections.
It can substantially reduce tokens for table-shaped data, but does not replace
MCP's JSON-RPC envelope or `structuredContent`.

Recommendation: keep JSON as the canonical typed result and evaluate TOON only
as an opt-in text rendering after measuring token count, model comprehension,
latency, and client compatibility against compact JSON and Markdown tables.

## Streamsy

The Streamsy repository has not yet been provided. Its architecture, tool
surface, output conventions, and project scaffolding remain a required input to
the reference review.

