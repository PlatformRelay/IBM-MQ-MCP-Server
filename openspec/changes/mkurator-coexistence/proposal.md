# MKurator coexistence

**Status:** Draft — blocked by ADR-0007

## Why

Direct mutation of MKurator-managed objects is reconciled away. Clients need
ownership warnings and safe alternatives without Kubernetes becoming a
prerequisite for generic IBM MQ use.

## Outcome

Optional ownership discovery maps MQ objects to their MKurator resources and
enforces the approved warn, block, or declarative-handoff behavior before
direct mutation, degrading explicitly when Kubernetes access is absent,
insufficient, or stale.

## In scope

- Explicitly configured ownership discovery via published MKurator APIs.
- Ownership and freshness annotations in inspection results.
- A pre-mutation policy hook consumed by guarded administration.

## Out of scope

- Requiring Kubernetes or MKurator for generic MQ operation.
- Producing or applying custom-resource changes unless ADR-0007 approves it.
- Duplicating MKurator reconciliation.

## Success signals

- The same binary behaves identically for generic MQ with the integration
  disabled.
- A mutation attempt on a managed object triggers the approved behavior with a
  pointer to its declarative owner.
- Degraded discovery is reported explicitly and never weakens MQ policy.

## Dependencies

- ADR-0007 (coexistence boundary).
- INS-001 for the inspection surface; ADM-001 for the enforcement hook.

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-008.
