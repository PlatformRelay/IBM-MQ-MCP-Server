# Connection profiles and capability policy

**Status:** Draft — blocked by ADR-0003 and ADR-0004

## Why

Every remote operation must name a profile and pass a deny-by-default
capability check before secrets are resolved or IBM MQ is contacted. IBM's
sample proves neither; MKurator proves the connection/secret pattern but not
MCP-side policy.

## Outcome

Operators declare named queue-manager profiles with validated endpoint, TLS,
authentication, and secret references, and grant each profile an explicit set
of operation capabilities that the server enforces locally before any
downstream I/O.

## In scope

- Profile catalog schema, startup validation, and per-profile client lifecycle.
- TLS verification on by default, custom CA, and approved mTLS flows.
- Environment and mounted-file secret references first; further providers as a
  separate slice.
- Capability vocabulary (`inspect`, `browse`, `consume`, `produce`,
  `administer`, `execute_mqsc`) with deny-by-default evaluation.
- Policy decision events consumable by the audit trail.
- Optional per-profile object-name allow/deny constraints as a separate slice.

## Out of scope

- MCP client authentication and remote transport authorization
  (`remote-access-and-operability`).
- The MQ tool surface itself.
- Dynamic profile reload.

## Success signals

- A read-constrained production profile and a write-capable development
  profile run side by side from one configuration.
- A denied operation produces an actionable error and provably no adapter call.
- Credentials never appear in configuration values, logs, errors, or results.
- One failing profile does not prevent healthy profiles from being used.

## Dependencies

- Bootstrap proposal approval.
- ADR-0003 (capability model) and ADR-0004 (configuration and secret providers).

Delivery slices are tracked in `agent-context/roadmap.md` under EPIC-003.
