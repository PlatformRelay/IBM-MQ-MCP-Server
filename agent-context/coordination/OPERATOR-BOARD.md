# OPERATOR-BOARD — ibm-mq-mcp

## In flight

| Lane | Branch | Worktree | Notes |
| --- | --- | --- | --- |
| — | — | — | — |

## Ready / next

| Lane | Status | Notes |
| --- | --- | --- |
| CON-001 | Ready | ADR-0004 Accepted; profile catalog + env/file secrets |
| POL-001 | After CON-001 (or parallel if disjoint) | ADR-0003 + ADR-0004 accepted |

## Integrated / Done

| Item | Notes |
| --- | --- |
| ADR-0004 — Config and secret providers | Accepted on `main` (pending push sha) |
| DOC-001 — Docs & operator UX | Integrated to `main` @ 9efa2f9 |
| OBS-001 — Health/metrics/logs | Integrated to `main` @ 7f1c256 |
| ADR-0009 MIT + OSS posture | Community files, Scorecard, Docs CI, Dependabot/Renovate |
| ADR-0003 — Operation-oriented capabilities | Accepted on `main` @ 26ff978 |
| FND-001 — Go skeleton + minimal MCP server | Integrated to `main` @ 5edb34c |
| FND-002 — CI quality gates | Integrated to `main` @ 99eaa5d |
| FND-003 — Release cosign/SBOM | Integrated to `main` @ 1fe36f8 |
| FND-004 — Local MQ + e2e harness | Done on `main` @ c2d70aa (feat @ 8bcf962; docs bookkeeping) |

## Parked

| Item | Why |
| --- | --- |
| CON-002 / MSG-*/ADM-*/SEC-*/INT-001 | Follow roadmap ADR gates (0006/0007/0008 where listed) |
