# Authentication

Two distinct authentication layers apply. Do not conflate them.

## MCP client → server ([ADR-0006](adr/0006-remote-transport-and-auth.md))

### stdio (default)

Local MCP hosts spawn `ibm-mq-mcp` as a child process. Authentication is the
**host OS/process boundary**: any principal that can start the server can use
every profile in the loaded catalog. This is appropriate for single-operator
workstations and trusted CI agents; it is **not** multi-tenant isolation.

### Streamable HTTP (opt-in)

Remote MCP is **disabled by default**. When enabled (`--remote-addr` /
`IBM_MQ_MCP_REMOTE_ADDR`), clients must present a **server-configured bearer
token** resolved from `env:` or `file:` references
(`--remote-auth-token-ref` / `IBM_MQ_MCP_REMOTE_AUTH_TOKEN_REF`).

| Property | Behaviour |
| --- | --- |
| Client header | `Authorization: Bearer <token>` |
| Validation | Constant-time compare against configured gate token |
| Passthrough | **Never** — client Authorization is stripped before MCP handling |
| MQ credentials | Unaffected — still come from the selected connection profile only |

OAuth, ingress mTLS, and per-client ACLs are deferred beyond v0.

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
profile grants it. MCP gate tokens must not influence mqweb authentication.
See [Policy](policy.md) and the [threat model](security/threat-model.md).

## Related pages

- [Configuration](configuration.md) — profile, remote, and secret-reference layout
- [Deployment](deployment.md) — remote-only mode
- [Security policy](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/SECURITY.md)
- [Design questions 10–12](product/design-questions.md)
