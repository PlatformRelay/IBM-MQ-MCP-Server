# Capability policy

!!! warning "Provisional — ADR-0003 open"
    Capability names, semantics, and deny-by-default evaluation are **not
    finalized**. The vocabulary below is illustrative from the
    [proposed system](architecture/proposed-system.md). **ADR-0003** is the
    single authority when accepted; until then, treat examples as non-binding.

## Design intent

- **Deny by default** — operations not explicitly granted for the active profile
  are rejected locally before any IBM MQ I/O.
- **Profile as trust boundary** — each named profile carries its own endpoint,
  credentials, capabilities, rate limits, and audit context.
- **Operation-oriented grants** — avoid ambiguous global modes like “read-only”
  without defining whether message payloads are included ([design questions 5–7](product/design-questions.md)).

## Illustrative capability vocabulary (TBD)

| Capability | Intended meaning (hypothesis) |
| --- | --- |
| `inspect` | Queue managers, objects, status, depth, diagnostics |
| `browse` | Non-destructive message metadata and optionally payloads |
| `consume` | Destructive get/consume — separately gated from `browse` |
| `produce` | Put messages with validation |
| `administer` | Typed define/alter/delete for MQ objects |
| `execute_mqsc` | Exceptional raw MQSC — disabled by default ([ADR-0008](adr/README.md#decision-queue)) |

Mixed read/write across environments is expressed by **different profiles**, not
by toggling modes at runtime. See [Examples](examples/README.md).

## Object-level constraints (TBD)

Whether queue/object name allowlists are required in the first release is an open
[design question 9](product/design-questions.md), tracked in
[POL-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/POL-002.md).

## Enforcement point

Policy evaluation sits in the application layer **before** the mqweb adapter.
Denied calls increment `ibm_mq_mcp_policy_denials_total` (profile label only) when
metrics are enabled — see [Observability](observability.md).

## Related decisions

| ADR | Topic |
| --- | --- |
| [ADR-0003](adr/README.md#decision-queue) | Capability model — **blocks POL-001** |
| [ADR-0007](adr/0007-mkurator-coexistence.md) | MKurator ownership vs direct mutation — **Accepted**; INT-001 hook before ADM mutations |
| [ADR-0008](adr/README.md#decision-queue) | Raw MQSC policy |
