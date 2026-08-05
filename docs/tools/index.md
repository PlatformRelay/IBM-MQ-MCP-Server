# Tool reference

## Current state

INS-001 and INS-002 register typed inspection tools when a profile catalog is loaded
(`--config` / `IBM_MQ_MCP_CONFIG`). INS-003 adds offline reason-code explanation
and side-effect-free profile connectivity checks. MSG-001 adds bounded non-destructive
message browse. MSG-002 adds validated message production. MSG-003 adds separately
gated destructive consume. Results are returned as JSON `structuredContent` ([ADR-0005](../adr/0005-structured-results-and-rendering.md)).

| Tool | Capability | Description |
| --- | --- | --- |
| `list_profiles` | _(local catalog; no MQ I/O)_ | Configured profiles with capabilities and validation status |
| `queue_manager_status` | `inspect` | Queue manager health; configured vs observed identity |
| `list_queues` | `inspect` | Bounded queue listing with filters, cursor, truncation |
| `get_queue` | `inspect` | Queue definition and live depth/status |
| `list_channels` | `inspect` | Bounded channel listing with filters, cursor, truncation |
| `get_channel` | `inspect` | Channel definition attributes |
| `get_channel_status` | `inspect` | Channel runtime status (`available` / `stale` / `unavailable`) |
| `list_listeners` | `inspect` | Bounded listener listing (unsupported mqweb modes return typed error) |
| `get_listener` | `inspect` | Listener definition attributes |
| `get_listener_status` | `inspect` | Listener runtime status |
| `list_subscriptions` | `inspect` | Bounded subscription listing |
| `get_subscription` | `inspect` | Subscription definition by id or name |
| `explain_mq_reason_code` | _(offline reference; no MQ I/O)_ | Explain an IBM MQ reason code from bundled data; unknown codes get a generic fallback |
| `check_profile_connectivity` | `inspect` | Verify mqweb reachability, identity match, and latency without mutation |
| `browse_queue_messages` | `browse` | Bounded non-destructive queue browse; metadata by default, optional payloads |
| `put_queue_message` | `produce` | Put one validated message; returns identifiers only (no payload echo) |
| `consume_queue_messages` | `consume` | Destructively get bounded messages (one mqweb DELETE each); metadata by default, optional payloads; mid-batch failures return partial results with `truncated: true` |

Policy denies remote tools before credential resolution or mqweb I/O when the
active profile lacks the required capability (`inspect`, `browse`, etc.). The offline reason-code tool never performs MQ
I/O. See [NOTICE](../NOTICE.md) for IBM MQRC attribution.

!!! note "Collection contract (ADR-0005)"
    List-style tools share a JSON envelope: `items`, `limit`, optional
    `cursor` / `nextCursor`, and `truncated` (+ `truncationReason`). Inspection
    lists default limit **50** (max **200**); browse and consume default to count **10**
    (max **100**). OUT-001 may add Markdown/TOON renderings later; clients should
    consume `structuredContent` only.

Run `task run` with a config path and connect an MCP inspector to list tools.

## Planned surface (remaining slices)

The [proposed system](../architecture/proposed-system.md) describes a small set
of typed, profile-explicit tools (inspect, browse, produce, etc.) with JSON
schemas for inputs and outputs. Final names and schemas depend on
[ADR-0003](../adr/README.md#decision-queue) and [ADR-0005](../adr/README.md#decision-queue).

## Generated reference (future)

When tool schemas exist in the repository, this section will be produced or
checked by automation so the published reference cannot drift from code.

| Check | Status |
| --- | --- |
| Schema-first tool definitions in Go | **Partial** — INS-001/INS-002/INS-003 inspection and diagnostics tools |
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
