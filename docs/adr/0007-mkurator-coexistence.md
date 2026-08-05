# ADR-0007: MKurator coexistence boundary

**Status:** Accepted  
**Date:** 2026-08-05

## Context

Design questions 18–20 ask how IBM MQ MCP Server should coexist with
[MKurator](https://github.com/PlatformRelay/MKurator) declarative queue
management without requiring Kubernetes for generic MQ estates or duplicating
controller reconciliation.

INT-001 and ADM-001 depend on a stable v0 boundary before guarded administration
can call a pre-mutation hook. Prior drafts described advisory integration,
explicit ownership metadata, and warn-versus-block behaviour but left them
open in the decision queue.

## Decision

Adopt the following v0 coexistence model:

| Question | Decision |
| --- | --- |
| 18 — advisory vs CR producer | **Advisory only.** The MCP server does **not** create, update, or apply MKurator custom resources. |
| 19 — discovery | **Explicit ownership metadata** in the profile catalog (`mkurator.managedObjects`) and optional object tags. Live Kubernetes API discovery is deferred to a later INT slice. |
| 20 — warn vs block | **Warn by default** when a mutation target is marked MKurator-managed. Operators may set `mkurator.mutationPolicy: block` on a profile to fail closed. |

Implementation contract (INT-001):

- `PreMutationHook.Evaluate(profile, object identity, ownership metadata)` returns
  `allow`, `warn`, or `block`.
- ADM-001 (and later ADM-002) **must** invoke the hook after capability
  authorization and **before** any mqweb mutation I/O.
- `block` outcomes return a typed error without contacting IBM MQ.
- `warn` outcomes proceed and attach the warning to the mutation result.

Reject for v0:

- **Kubernetes-required operation** — breaks generic MQ deployments.
- **MCP as MKurator reconciler** — duplicates MKurator and expands blast radius.
- **Silent direct mutation of managed objects** — hides reconciliation fights.

## Consequences

### Positive

- Generic MQ and MKurator-managed estates share one binary with explicit
  operator configuration.
- Guarded administration can ship queue mutations without waiting for live K8s
  discovery.
- Warn-by-default preserves emergency break-glass while block policy supports
  strict environments.

### Negative

- Operators must maintain ownership metadata until live discovery lands.
- Advisory mode cannot prevent reconciliation — only surfaces the risk.
- Object-tag stub relies on catalog conventions until mqweb exposes richer metadata.

## Alternatives

### Change-producing MCP integration (apply MKurator CRs)

Rejected for v0. Expands scope into Kubernetes write paths, credential models,
and reconciliation semantics that belong to MKurator and GitOps workflows.

### Live Kubernetes discovery only (no catalog metadata)

Rejected for v0 first slice. Requires cluster credentials on every MCP host and
couples policy evaluation to K8s availability; explicit metadata is sufficient
for the hook contract and can be augmented later without breaking callers.

### Block-by-default on all mutations

Rejected. Too disruptive for break-glass operations; warn-by-default with an
explicit operator opt-in to `block` matches design question 20.

## Validation implications

- Unit tests must cover allow, warn, and block paths for catalog patterns and
  object tags.
- ADM-001 contract tests must prove hook invocation precedes adapter mutation
  calls and that policy denial leaves `TotalCalls()==0`.
- Documentation must state that dry-run is unsupported for queue mutations
  unless mqweb exposes a truthful preview API.
