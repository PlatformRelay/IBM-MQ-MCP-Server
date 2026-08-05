# Tool reference

## Current state

**No IBM MQ tools are registered yet.** The bootstrap server ([FND-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/FND-001.md))
exposes the MCP protocol skeleton only. Tool contracts will land with inspection
([INS-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INS-001.md)),
messaging ([MSG-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/MSG-001.md) onward),
and administration ([ADM-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/ADM-001.md))
slices — each blocked on their ADRs and dependencies.

Run `task run` and connect an MCP inspector to confirm the empty tool list.

## Planned surface (hypothesis)

The [proposed system](../architecture/proposed-system.md) describes a small set
of typed, profile-explicit tools (inspect, browse, produce, etc.) with JSON
schemas for inputs and outputs. Final names and schemas depend on
[ADR-0003](../adr/README.md#decision-queue) and [ADR-0005](../adr/README.md#decision-queue).

## Generated reference (future)

When tool schemas exist in the repository, this section will be produced or
checked by automation so the published reference cannot drift from code.

| Check | Status |
| --- | --- |
| Schema-first tool definitions in Go | Not started |
| Docs generation or freshness test in CI | **Planned** — optional job alongside `mkdocs build --strict` |
| Breaking schema changes | Will require story acceptance + ADR when applicable |

!!! note "CI placeholder"
    The [Docs workflow](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/.github/workflows/docs.yaml)
    currently runs `mkdocs build --strict` only. A generated tool-reference check
    will be added when schemas land — tracked under [DOC-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/DOC-001.md).

## Related pages

- [Policy](../policy.md) — capability grants (provisional)
- [Examples](../examples/README.md) — profile illustrations
- [Feature scope](../product/feature-scope.md)
