# MKurator coexistence

IBM MQ MCP Server targets **generic IBM MQ** deployments. [MKurator](https://github.com/PlatformRelay/MKurator)
(Kubernetes operator for declarative MQ) is an optional coexistence partner — not
a runtime prerequisite.

!!! warning "Blocked — ADR-0007 open"
    Ownership discovery, warn/block/handoff behaviour, and Kubernetes API usage
    are **undecided**. This page describes draft intent from
    [openspec/changes/mkurator-coexistence](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/openspec/changes/mkurator-coexistence/proposal.md);
    track [ADR-0007](../adr/README.md#decision-queue) and
    [INT-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INT-001.md).

## Principles

| Principle | Detail |
| --- | --- |
| No Kubernetes required | The same binary must operate for non-Kubernetes MQ estates |
| No reconciliation duplication | This server does not replace MKurator controllers |
| Explicit degradation | Missing or stale K8s access is reported; MQ policy is not weakened |
| Policy before mutation | Ownership checks hook into guarded administration ([ADM-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/ADM-001.md)) |

## Draft in-scope behaviour (TBD)

When enabled and ADR-0007 accepts:

- Configured ownership discovery via published MKurator APIs
- Ownership/freshness annotations on inspection results ([INS-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INS-001.md))
- Pre-mutation policy hook — warn, block, or declarative handoff per ADR-0007

## Draft out-of-scope

- Applying MKurator custom resources unless ADR-0007 explicitly approves
- Deploying or upgrading queue managers
- Replacing MKurator as the source of truth for managed objects

## Design questions

See [design questions 18–20](../product/design-questions.md):

- Advisory vs blocking vs CR handoff
- Discovery via Kubernetes API vs explicit config
- Direct admin of managed objects — warn or block?

## Related pages

- [Threat model](../security/threat-model.md)
- [Policy](../policy.md)
- [Version support matrix](version-matrix.md)
