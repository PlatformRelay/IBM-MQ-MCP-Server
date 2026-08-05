# Architecture decision records

Accepted decisions are immutable records; later changes supersede them with a
new ADR.

## Decision queue

| Planned ADR | Decision | Blocking stories |
| --- | --- | --- |
| [ADR-0001](0001-go-and-official-mcp-sdk.md) | Go and official MCP Go SDK — **Accepted** | FND-001 and all implementation |
| [ADR-0002](0002-mqweb-first-connectivity.md) | mqweb REST first; PCF/native deferred — **Accepted** | CON-001 |
| [ADR-0003](0003-capability-model.md) | Operation-oriented capability model — **Accepted** | POL-001, POL-002, MSG-001..003, ADM-001 |
| [ADR-0004](0004-configuration-and-secrets.md) | Configuration and secret providers — **Accepted** | CON-001, CON-002 |
| [ADR-0005](0005-structured-results-and-rendering.md) | Structured results and token-efficient rendering — **Accepted** | INS-001, OUT-001, MSG-001 |
| ADR-0006 | MCP transports and client authorization | SEC-001, SEC-002 |
| ADR-0007 | MKurator coexistence boundary | INT-001, ADM-001 |
| ADR-0008 | Raw MQSC policy | ADM-003 |
| [ADR-0009](0009-license-and-oss-maturity.md) | MIT license + OSS maturity baseline — **Accepted**; container/binary delivery detail remains FND-003 | FND-002, FND-003, DOC-001 |

Each ADR will use `Proposed`, `Accepted`, `Superseded`, or `Rejected` status and
record context, decision, alternatives, and consequences.
