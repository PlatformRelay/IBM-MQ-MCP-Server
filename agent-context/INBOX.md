# INBOX — ibm-mq-mcp

Operator-facing items only. Agents append; operator answers; agents record in
`decisions.md` and remove answered items.

## Decisions

### 🟡 DECIDED (awaiting approval) — Bootstrap product intent = Accepted

**Context:** Operator asked for Kollect/MKurator-grade OSS hygiene and a
well-maintained public repo. Combined with `/agent-loop-local` decide-and-log.
**Options:** A approve as written / B amend / C stay draft.  
**Chose:** A — proposal marked Accepted; FND-001 unblocked.  
**Revert:** Set proposal status back to Draft; revert AGENTS.md project-state
paragraph; park FND-001.

### 🔴 DECIDED (awaiting approval) — MIT license (not Apache-2.0)

**Context:** DQ 21 suggested Apache-2.0; Kollect and MKurator are MIT.  
**Chose:** MIT via ADR-0009 for portfolio consistency.  
**Revert:** Replace `LICENSE` with Apache-2.0, supersede ADR-0009, update
badges/SPDX.

## Operator tasks

### 🟢 Authorize local-loop git actions

Still needed for continuous `/agent-loop-local` merges:

1. **push-to-main**
2. **remote-branch-delete**
3. **release-tag** (later)

Reply with an explicit sentence naming those classes.

### 🟢 Enable GitHub Pages

Docs workflow deploys to the `github-pages` environment. Enable Pages
(Source: GitHub Actions) once if the first deploy fails.

### 🟢 Optional: Renovate token

Self-hosted Renovate uses `RENOVATE_TOKEN` when present; otherwise
`GITHUB_TOKEN` (limited for some fork/PR cases).
