# IBM MQ and mqweb version support

!!! warning "Not certified yet"
    [ADR-0002](../adr/0002-mqweb-first-connectivity.md) requires documenting
    supported IBM MQ/mqweb versions and deployment modes. **No combination is
    certified in this repository yet.** The matrix below records intent,
    unknowns, and the evidence required before a row can move to **Supported**.

Authority: [design question 13](../product/design-questions.md),
[FND-004](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/FND-004.md)
(live MQ environment), [MSG-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/MSG-001.md)
(mqweb semantics spike).

## Connectivity baseline

First-release downstream access is **mqweb REST only** ([ADR-0002](../adr/0002-mqweb-first-connectivity.md)).
Native PCF/MQI is deferred. Implications:

- mqweb must be installed, enabled, reachable, and authorized.
- Standalone mqweb vs full-install mqweb capabilities **differ** — both must be
  tested before support claims.
- REST does **not** provide MQI transaction/syncpoint or high-throughput semantics.

## Support matrix (TBD)

Status legend:

| Status | Meaning |
| --- | --- |
| **Unknown** | Not tested; do not deploy for production use |
| **Planned** | In scope for validation via FND-004 / MSG-001 |
| **Supported** | Documented, tested, listed in release notes — *none yet* |

| IBM MQ major | Platform | mqweb deployment mode | Admin REST | Messaging REST | Status | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| IBM MQ 9.4.x | Linux x86_64 (Kind) | Full install (Helm) | Planned | Planned | Planned | MKurator Kind stack `9.4.5.1-r1`; FND-004 e2e reachability |
| IBM MQ 9.4.x | Linux x86_64 (Docker) | Full install | Planned | Planned | Planned | MKurator `hack/mq-docker`; lighter local path |
| *TBD* | Linux x86_64 | Standalone mqweb | Unknown | Unknown | Unknown | Capability gaps vs full install |
| *TBD* | Linux arm64 | Full install | Unknown | Unknown | Unknown | Container multi-arch build exists; MQ validation TBD |
| *TBD* | Windows | Full install | Unknown | Unknown | Unknown | Not prioritized until evidence |
| *TBD* | z/OS | *TBD* | Unknown | Unknown | Unknown | Explicit later extension ([feature scope](../product/feature-scope.md)) |

Replace *TBD* cells with tested version ranges when FND-004 records evidence.
**Do not infer support from IBM documentation alone.**

## Messaging REST risks (pre-certification)

These behaviours must be **proven non-destructive** before browse ships
([MSG-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/MSG-001.md)):

- Browse vs get/consume semantics on target mqweb versions
- Payload formats and size limits
- Selector and property support gaps vs native clients

## Operator guidance until certified

1. Treat all MQ versions as **unsupported** for production AI-driven access.
2. Use the bootstrap server without MQ connectivity for MCP host integration testing.
3. Track [roadmap](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/roadmap.md)
   for FND-004 and MSG-001 completion.
4. Never claim exactly-once, transactional, or high-throughput behaviour over REST
   ([ADR-0002 guardrails](../adr/0002-mqweb-first-connectivity.md)).

## Related pages

- [Deployment](../deployment.md)
- [Troubleshooting](../troubleshooting.md)
- [MKurator coexistence](mkurator-coexistence.md)
