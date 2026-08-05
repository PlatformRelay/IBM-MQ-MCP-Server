# OPERATOR-BOARD — ibm-mq-mcp

## In flight

| Lane | Branch | Worktree | Notes |
| --- | --- | --- | --- |
| FND-003 — Release cosign/SBOM | `lane/fnd-003` | `../ibm-mq-mcp-fnd-003` | DQ 22 → C (binary + container); awaiting INBOX approval |

## Ready / next

| Lane | Status | Notes |
| --- | --- | --- |
| FND-003 — Release cosign/SBOM | 🟡 In flight | See In flight |
| OBS-001 — Health/metrics/logs | ⬜ Optional | Unblocked after FND-001; parallel polish |
| DOC-001 — Docs & operator UX | ⬜ Optional | Can run alongside FND-003 |

## Integrated / Done

| Item | Notes |
| --- | --- |
| ADR-0009 MIT + OSS posture | Community files, Scorecard, Docs CI, Dependabot/Renovate |
| FND-001 — Go skeleton + minimal MCP server | Integrated to `main` @ 5edb34c |
| FND-002 — CI quality gates | Integrated to `main` @ 99eaa5d |

## Parked

| Item | Why |
| --- | --- |
| CON-*/POL-*/MSG-*/ADM-*/SEC-*/INT-001 | Waiting on ADR-0003+ |
| FND-004 | Licensing / live MQ approach |
