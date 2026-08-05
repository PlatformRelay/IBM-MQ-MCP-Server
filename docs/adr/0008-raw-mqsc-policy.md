# ADR-0008: Raw MQSC exceptional gate

**Status:** Accepted  
**Date:** 2026-08-05

## Context

Design question 8 and [ADR-0003](0003-capability-model.md) establish `execute_mqsc`
as an exceptional, off-by-default capability distinct from `administer`. IBM's
reference MCP sample exposes arbitrary `runmqsc`, collapsing discovery,
authorization, and execution into one unbounded surface. Typed administration
(ADM-001, ADM-002) covers common mutations; operators may still need occasional
read-only MQSC that typed tools do not yet model.

ADM-003 implements the policy decided here. SEC-002 will consume audit events;
this ADR defines what must be recorded for raw MQSC until that slice lands.

## Decision

**Default:** no raw MQSC MCP tool is registered.

**Double opt-in** is required before any raw MQSC reaches IBM MQ:

1. **Server enablement** — process flag `--enable-mqsc` or environment variable
   `IBM_MQ_MCP_ENABLE_MQSC` (truthy: `1`, `true`, `yes`, `on`).
2. **Profile capability** — the active profile must grant `execute_mqsc`
   ([ADR-0003](0003-capability-model.md)). The capability is not implied by
   `administer` or `inspect`.

When both are present, register one tool: `execute_mqsc`.

**v0 verb allowlist (read-only only):** before any mqweb I/O, parse the
submitted command and permit only these initial verbs (case-insensitive):

| Verb | Purpose |
| --- | --- |
| `DISPLAY` | Full display keyword |
| `DIS` | Abbreviated display |
| `PING` | Queue manager / object ping |

All other verbs (`ALTER`, `DEFINE`, `DELETE`, `SET`, `REFRESH`, etc.) are
denied locally with `ErrMQSCVerbDenied` and never reach the adapter.

Additional local guards for v0:

- Reject empty commands.
- Reject multiple statements in one submission (semicolon separator).

**Audit:** every allowed execution emits a structured log event with profile,
operation identity (`execute_mqsc`), and **redacted** command text (reuse
`messaging.RedactSecrets` patterns). Do not log raw credentials, HTTP bodies,
or mqweb response payloads as audit evidence.

**Internal typed mutations** (queue/channel/CHLAUTH/authrec adapters) continue
to use structured `runCommandJSON` internally; they are not exposed as the raw
MQSC tool and do not require server or profile MQSC opt-in.

Reject for v0:

- **Option B — register tool always, gate only at execution:** violates
  “absent by default” and makes undiscoverable-by-default tests meaningless.
- **Option C — unrestricted raw MQSC behind `execute_mqsc` only:** recreates
  the IBM sample anti-pattern; mutating verbs must stay off the allowlist until
  a deliberate ADR supersedes this one.

## Consequences

### Positive

- Default deployments expose no arbitrary MQSC surface.
- Operators who need read-only escape hatches must configure two independent
  controls (server + profile).
- Denied verbs and missing capabilities fail before secret resolution and mqweb
  I/O, simplifying tests and threat-model claims.
- Command text is auditable without leaking secrets in logs.

### Negative

- Read-only operators cannot run non-allowlisted diagnostic MQSC (e.g. some
  `START`/`STOP` variants) until a future ADR extends the allowlist or adds
  typed tools.
- Server restart is required to toggle `--enable-mqsc`; profile-only grants are
  insufficient to expose the tool.
- Full SEC-002 audit schema integration is deferred; v0 uses structured slog
  events as the interim audit hook.

## Alternatives

### Always register, profile-only gate (B)

Rejected. Tool would appear in MCP discovery for every deployment, contradicting
“absent by default” and ADM-003 acceptance tests.

### Unrestricted MQSC with profile capability only (C)

Rejected. Reintroduces unbounded administration through a single prompt-driven
tool; incompatible with the typed-first architecture and threat model.

## Validation implications for ADM-003

- Unit tests: tool absent from `ListTools` without server opt-in; denied verbs
  and missing `execute_mqsc` capability never increment adapter call counters.
- Integration tests: mqweb adapter sends `runCommand` JSON with plain-text
  command in `parameters.command`.
- Documentation: ADR index, configuration reference, tool catalog, and design
  question 8 updated to reference this ADR as **Accepted**.
