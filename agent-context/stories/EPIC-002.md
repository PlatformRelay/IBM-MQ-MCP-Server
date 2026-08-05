# EPIC-002 — Foundation and delivery

**Status:** Open; runtime selected  
**Authority:** `openspec/changes/bootstrap-mq-mcp/proposal.md`

## Outcome

A contributor can build, verify, package, release, and integration-test the
server against live IBM MQ using documented commands alone.

## Stories

- [FND-001](FND-001.md) — Go module skeleton and minimal MCP server.
- [FND-002](FND-002.md) — CI quality gates and supply-chain checks.
- [FND-003](FND-003.md) — Container packaging and release automation.
- [FND-004](FND-004.md) — Live IBM MQ development and e2e environment.

## Acceptance

- Every child story is done with recorded evidence.
- The gates named by child stories run green in CI on the default branch.

## Session log

### 2026-08-05

**Done:** Split the oversized foundation story into four slices; FND-001
narrowed to the skeleton and minimal server.  
**Next:** Approve product intent and ADR-0009 delivery targets, then FND-001.  
**Do not:** Start FND-002/FND-003 before the FND-001 skeleton exists.  
**Blocked:** Product intent approval; ADR-0009 for FND-003.
