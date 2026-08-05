# EPIC-008 — Coexistence and adoption

**Status:** Blocked by ADR-0007 for INT-001; DOC-001 open  
**Authority:** `openspec/changes/mkurator-coexistence/proposal.md`

## Outcome

The server coexists with MKurator without requiring it, and every audience can
adopt the server from documentation alone.

## Stories

- [INT-001](INT-001.md) — MKurator ownership discovery and mutation policy.
- [DOC-001](DOC-001.md) — Documentation and operator experience.

## Acceptance

- Generic MQ use has zero Kubernetes dependency, verified by test.
- The strict MkDocs build and generated tool reference stay green in CI.

## Session log

### 2026-08-05

**Done:** Grouped coexistence and adoption; DOC-001 keeps the bootstrap
proposal as authority.  
**Next:** Resolve ADR-0007 advisory-versus-change-producing scope.  
**Do not:** Make Kubernetes a runtime prerequisite for any generic feature.  
**Blocked:** ADR-0007 for INT-001 only.
