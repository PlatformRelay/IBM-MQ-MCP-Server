# Configuration

!!! note "Schema provisional"
    Field names and validation rules follow [ADR-0004](adr/0004-configuration-and-secrets.md).
    Capability enforcement is [POL-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/POL-001.md).

## Bootstrap inputs

| Input | Purpose |
| --- | --- |
| `--config` | Path to the profile catalog YAML or JSON file |
| `IBM_MQ_MCP_CONFIG` | Same as `--config` when the flag is omitted |
| `--strict-startup` | Fail process start if any profile fails validation |
| `--enable-mqsc` | Register exceptional raw MQSC tool ([ADR-0008](adr/0008-raw-mqsc-policy.md)); requires profile `execute_mqsc` at call time |
| `IBM_MQ_MCP_ENABLE_MQSC` | Same as `--enable-mqsc` when the flag is omitted (truthy: `1`, `true`, `yes`, `on`) |
| `--ops-addr` | Optional ops HTTP listen address (see [Observability](observability.md)) |
| `IBM_MQ_MCP_OPS_ADDR` | Same as `--ops-addr` when the flag is omitted |

When no config path is supplied, the server starts with an empty catalog (valid
bootstrap). Readiness reports configuration validity without contacting queue
managers.

## Profile catalog schema

Top-level key **`profiles`**: map of stable profile name → profile object.

| Field | Required | Description |
| --- | --- | --- |
| `queueManager` | yes | IBM MQ queue manager name |
| `endpoint` | yes | mqweb base URL (`https://host:port`) |
| `authentication` | yes | mqweb credential method (see [Authentication](authentication.md)) |
| `tls` | no | TLS settings (verification on by default) |
| `capabilities` | yes | Operation grants per [ADR-0003](adr/0003-capability-model.md); enforced before secret resolution and MQ I/O |
| `mkurator` | no | Optional MKurator coexistence metadata per [ADR-0007](adr/0007-mkurator-coexistence.md) |
| `timeout` | no | Per-profile HTTP timeout (Go duration string, default `30s`) |

Example (secret-free — refs only):

```yaml
profiles:
  production:
    queueManager: PROD1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:MQ_PROD_CREDENTIALS
    tls:
      caRef: file:/etc/mq/production-ca.pem
    capabilities:
      - inspect
      - browse
```

See [Examples](examples/README.md) for additional illustrative profiles.

## Secret references

Production credentials must **not** appear inline in configuration. Supported
reference schemes in v0 ([ADR-0004](adr/0004-configuration-and-secrets.md)):

| Prefix | Example | Resolves to |
| --- | --- | --- |
| `env:` | `env:MQ_PROD_PASSWORD` | Environment variable value |
| `file:` | `file:/run/secrets/mq/password` | Mounted file contents (trimmed) |

- **HTTP Basic:** `secretRef` resolves to `username:password` (single value).
- **mTLS:** `certificateRef` and `privateKeyRef` are file refs; optional
  `passphraseRef` for encrypted private keys.
- Secret **values** are resolved lazily when a profile is first used, not at
  catalog parse time.
- Kubernetes Secrets and Vault integrations are [CON-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-002.md).

## TLS

| Field | Default | Description |
| --- | --- | --- |
| (implicit) | verify on | Server certificate validated against system roots |
| `caRef` | — | Additional CA bundle (`file:` ref) |
| `insecureSkipVerify` | `false` | Opt-in for local Kind only — not for production |

## Validation and startup behaviour

At startup the server validates **every** profile:

- Unique profile names
- Required fields and well-formed URLs
- Authentication shape matches declared type
- Secret **references** are syntactically valid (values not required yet)
- TLS settings are coherent (e.g. custom CA path exists when referenced at use time)
- Capability names are known per ADR-0003 and at least one grant is listed

**Default (fail-open):** invalid profiles are marked unusable; healthy profiles
remain available. **Strict (`--strict-startup`):** any validation error exits
the process.

Selecting a profile for an operation is explicit (MCP tools arrive later); the
catalog and per-profile client pool are wired in [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md).

## Related

- [Authentication](authentication.md) — mqweb Basic and mTLS
- [ADR-0004](adr/0004-configuration-and-secrets.md) — authoritative decision
- [Design questions 10–11](product/design-questions.md)
