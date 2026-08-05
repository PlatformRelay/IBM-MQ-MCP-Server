# Feature scope

The scope is organized as hypotheses pending design approval.

## Must-have foundation

- Named profiles for multiple queue managers and mqweb endpoints.
- Explicit per-profile capabilities with deny-by-default evaluation.
- TLS verification, custom CA, basic authentication, and mutual TLS.
- Secret references rather than inline production credentials.
- Typed discovery and inspection tools.
- Typed message browsing and production with strict size/count bounds.
- Structured MCP results with output schemas and actionable errors.
- stdio for local clients and Streamable HTTP for remote deployment.
- Audit events identifying client/session, profile, operation, target, outcome,
  and duration without recording credentials or sensitive payloads.
- Unit, contract, integration, and live IBM MQ end-to-end testing layers.
- MkDocs documentation, examples, ADRs, security policy, contribution guide,
  container image, CI, release automation, SBOM, and vulnerability scanning.

## Strong candidates

- Typed consume/get operations, separately gated from browse.
- Typed administrative operations for queues, topics, channels, channel
  authentication, and authority records.
- Kubernetes Secret and external secret-provider integrations.
- MCP-server OAuth for remote clients, distinct from downstream MQ
  authentication.
- Optional MKurator ownership discovery.
- Health, readiness, Prometheus metrics, OpenTelemetry traces, and structured
  logs.
- Profile and object allow/deny filters.
- Message payload redaction, media-type handling, and safe binary rendering.
- Token-budget-aware field selection and result pagination.

## Later extensions

- PCF/native IBM MQ adapter for environments without mqweb.
- IBM MQ authentication-token/OIDC variants where supported by the selected
  adapter and queue-manager version.
- z/OS-specific compatibility and integration tests.
- TOON text rendering for proven high-volume tabular results.
- Dynamic configuration reload and profile health notifications.
- Approval workflow integrations for high-risk mutations.

## Deliberate non-goals for the first release

- Deploying or upgrading queue managers.
- Replacing IBM MQ Operator or MKurator reconciliation.
- Bridging or moving messages continuously between brokers.
- Acting as a generic broker abstraction for Kafka, RabbitMQ, and IBM MQ.
- Exposing unrestricted shell commands or unrestricted MQSC by default.
- Returning unlimited message payloads or entire queues to a model.

## Risks requiring explicit design

| Risk | Design response |
| --- | --- |
| Prompt-triggered destructive action | Server-side capabilities, bounded tools, confirmation hints, audit |
| Read-only profile still exposes sensitive payloads | Separate `inspect` and `browse`; redaction and limits |
| Direct mutation fights MKurator | Ownership warning and optional declarative handoff |
| Credentials leak through errors/logs | Secret-reference types and centralized redaction |
| A “write-only” profile cannot verify puts | Define whether minimal acknowledgement/health inspection is always available |
| Raw MQSC bypasses typed policy | Exclude or separately gate with command parser/allowlist |
| Large queue results exhaust context | Server-side filters, cursors, limits, field selection, summaries |
| mqweb is unavailable or feature-incomplete | Adapter seam; document PCF as later option |

