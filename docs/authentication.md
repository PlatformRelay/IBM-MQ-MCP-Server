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

## Server → IBM MQ / mqweb

Downstream authentication to mqweb is defined in [ADR-0004](adr/0004-configuration-and-secrets.md).
First-release methods:

| Method | Config | Credential source |
| --- | --- | --- |
| HTTP Basic | `authentication.type: basic` + `secretRef` | `username:password` in env or file ref |
| Client-certificate mTLS | `authentication.type: mtls` + `certificateRef` + `privateKeyRef` | PEM cert and key via `file:` refs |

| Method | Status |
| --- | --- |
| TLS + custom CA | Supported via `tls.caRef` ([configuration](configuration.md)) |
| LDAP-backed basic | Deferred — later ADR |
| MQ auth tokens / OIDC | Deferred — later ADR |

Secret delivery uses **environment** and **mounted file** references only in v0
([CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md));
Kubernetes Secrets and Vault follow [CON-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-002.md).

## Identity separation

- **MCP identity** — who invoked a tool (client/session), recorded in audit
  events when [SEC-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/SEC-002.md) lands.
- **MQ identity** — the credential bound to the selected **connection profile**,
  used only inside the mqweb adapter.

A read-only MCP session must not imply read-only MQ access unless the chosen
profile grants it. See [Policy](policy.md) and the [threat model](security/threat-model.md).

## Related pages

- [Configuration](configuration.md) — profile and secret-reference layout
- [Security policy](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/SECURITY.md)
- [Design questions 10–12](product/design-questions.md)
