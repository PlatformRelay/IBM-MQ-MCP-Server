# Bootstrap design

**Status:** Deferred until ADR-0003 and ADR-0004 are accepted.

The current design hypothesis is documented in
`docs/architecture/proposed-system.md`. This file will become the technical
design only after the proposal is approved; it must map each component and test
to the accepted capability and connectivity semantics.

## Accepted foundation

- Go with the official MCP Go SDK
  ([ADR-0001](../../../docs/adr/0001-go-and-official-mcp-sdk.md)).
- CGO-free by default.
- Explicit MCP, application, policy, domain-port, and adapter package
  boundaries.
- mqweb Administrative and Messaging REST adapters only in the first release
  ([ADR-0002](../../../docs/adr/0002-mqweb-first-connectivity.md)).
- Separate typed administration and messaging ports; PCF/native is deferred.
