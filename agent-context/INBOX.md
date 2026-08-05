# INBOX — ibm-mq-mcp

Operator-facing items only. Agents append; operator answers; agents record in
`decisions.md` and remove answered items.

## Decisions

### ✅ Answered — Bootstrap product intent = Accepted

Operator confirmed 2026-08-05: "The bootstrap proposal is approved."
Recorded in `decisions.md`. Proposal already marked Accepted.

### ✅ Answered — MIT license (ADR-0009)

Operator accepted portfolio MIT posture via OSS/hygiene direction; ADR-0009
Accepted. No revert requested.

## Operator tasks

### ✅ Authorized — push-to-main + remote-branch-delete

Operator 2026-08-05: "I authorize push-to-main and remote-branch-delete for
ibm-mq-mcp." Valid for this `/agent-loop-local` session.

### 🟢 Enable GitHub Pages

Docs workflow deploys to the `github-pages` environment. Enable Pages
(Source: GitHub Actions) if deploy fails.

### 🟢 Optional: Renovate token

Self-hosted Renovate uses `RENOVATE_TOKEN` when present; otherwise
`GITHUB_TOKEN`.
