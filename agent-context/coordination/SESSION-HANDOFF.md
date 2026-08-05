# SESSION-HANDOFF — ibm-mq-mcp

**Updated:** 2026-08-05  
**Mode:** `/agent-loop-local`

## State

- **main tip:** pending ADR-0004 push
- **ADR-0004** Accepted — env/file secrets; Basic + mTLS mqweb auth; fail-open startup.
- **ADR-0003** Accepted @ `26ff978` — operation-oriented capability model.
- **FND-004** done @ `c2d70aa` (integrated feat @ `8bcf962`; local MQ docs, Kind/Docker tasks, opt-in e2e).
- DOC-001 @ `9efa2f9`; OBS-001 @ `7f1c256`; FND-003 @ `1fe36f8`; FND-002 @ `99eaa5d`; FND-001 @ `5edb34c`.
- GitHub **Environment** `release` created (repo Settings → Environments) — FND-003 release OIDC/cosign gates unblocked.

## In flight

| Lane | Branch | Worktree | Notes |
| --- | --- | --- | --- |
| — | — | — | — |

## Next

1. **CON-001** — profile catalog, secrets, TLS, client pool (ADR-0004 unblocks).
2. **POL-001** — after CON-001, or in parallel if work is disjoint from profile wiring.

Use FND-004 harness (`IBM_MQ_MCP_E2E=1`, mkurator Kind) when validating connectivity slices.

## Do not

- Open PRs under this local loop.
- Auto-merge without independent APPROVE.
