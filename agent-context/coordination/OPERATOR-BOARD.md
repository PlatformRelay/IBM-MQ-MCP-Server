# OPERATOR-BOARD — ibm-mq-mcp

## In flight

| Lane | Branch | Worktree | Notes |
| --- | --- | --- | --- |
| — | — | — | — |

## Ready / next

| Lane | Status | Notes |
| --- | --- | --- |
| — | — | Remaining delivery slices need ADR-0003 (capabilities), ADR-0004 (secrets), ADR-0006 (remote MCP), ADR-0007/0008 as listed on roadmap; **FND-004** optional / still parked (licensing + live MQ) |

## Integrated / Done

| Item | Notes |
| --- | --- |
| DOC-001 — Docs & operator UX | Integrated to `main` @ 9efa2f9 |
| OBS-001 — Health/metrics/logs | Integrated to `main` @ 7f1c256 |
| ADR-0009 MIT + OSS posture | Community files, Scorecard, Docs CI, Dependabot/Renovate |
| FND-001 — Go skeleton + minimal MCP server | Integrated to `main` @ 5edb34c |
| FND-002 — CI quality gates | Integrated to `main` @ 99eaa5d |
| FND-003 — Release cosign/SBOM | Integrated to `main` @ 1fe36f8 |

## Parked

| Item | Why |
| --- | --- |
| CON-*/POL-*/MSG-*/ADM-*/SEC-*/INT-001 | Waiting on ADR-0003+ (CON/POL still ADR-blocked) |
| FND-004 | Licensing / live MQ approach |
