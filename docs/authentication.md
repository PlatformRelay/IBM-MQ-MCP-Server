# Authentication

Two distinct authentication layers apply. Do not conflate them.

## MCP client → server (TBD — ADR-0006)

How remote MCP clients authenticate to this server (stdio hosts use the OS
user/process boundary instead) is **undecided**. Candidates include OAuth for
Streamable HTTP, mTLS at the ingress, or host-local stdio only for v0.

| Decision | Status | Authority |
| --- | --- | --- |
| Remote MCP transport target | Open | [ADR-0006](adr/README.md#decision-queue) |
| MCP client identity model | Open | [ADR-0006](adr/README.md#decision-queue), [SEC-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/SEC-001.md) |

Until ADR-0006 is accepted, treat **stdio-only local use** as the supported
deployment pattern. Do not document OAuth, API keys, or ingress auth as shipped
behaviour.

## Server → IBM MQ / mqweb (TBD — ADR-0004)

Downstream authentication to mqweb is planned to support (at minimum) TLS with
custom CA, HTTP basic, and mutual TLS. LDAP-backed basic, MQ authentication
tokens, and z/OS-specific variants are design questions — not promised for v0.

| Method | Intent | Status |
| --- | --- | --- |
| TLS + custom CA | Verify mqweb endpoint | Planned ([feature scope](product/feature-scope.md)) |
| HTTP basic | mqweb REST | Planned; credential via secret ref |
| Mutual TLS | mqweb REST | Planned; cert/key via secret ref |
| MQ auth tokens / OIDC | mqweb variants | Later ([feature scope](product/feature-scope.md)) |

Secret delivery (env, file, Kubernetes, Vault) is owned by
[ADR-0004](adr/README.md#decision-queue) and [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md).

## Identity separation

- **MCP identity** — who invoked a tool (client/session), recorded in audit
  events when [SEC-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/SEC-002.md) lands.
- **MQ identity** — the credential bound to the selected **connection profile**,
  used only inside the mqweb adapter.

A read-only MCP session must not imply read-only MQ access unless the chosen
profile grants it. See [Policy](policy.md) and the [threat model](security/threat-model.md).

## Related pages

- [Configuration](configuration.md) — profile and secret-reference layout (provisional)
- [Security policy](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/SECURITY.md)
- [Design questions 10–12](product/design-questions.md)
