# Configuration

!!! warning "Provisional — ADR-0004 open"
    Configuration file format, secret-reference schemes, and reload behaviour
    are **not finalized**. This page describes the intended direction from the
    [proposed system](architecture/proposed-system.md) and
    [design questions](product/design-questions.md) (items 11–12). Track
    [ADR-0004](adr/README.md#decision-queue) and
    [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md).

## Bootstrap skeleton (today)

The current binary accepts:

| Input | Purpose |
| --- | --- |
| `--ops-addr` | Optional ops HTTP listen address (see [Observability](observability.md)) |
| `IBM_MQ_MCP_OPS_ADDR` | Same as `--ops-addr` when the flag is omitted |

No profile file or MQ endpoint is loaded yet. Static bootstrap validation
succeeds with an empty configuration until [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md)
lands.

## Planned configuration model (TBD)

The target is a **named profile catalog**: each profile is an independent trust
boundary (endpoint, TLS, credential references, capabilities). Example shape
(literal values are placeholders — see [Examples](examples/README.md)):

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

Open design questions:

- File path vs environment vs Kubernetes ConfigMap delivery ([design question 11](product/design-questions.md))
- Whether remote Streamable HTTP config differs from stdio deployments ([ADR-0006](adr/README.md#decision-queue))

## Secret references (TBD)

Production credentials must **not** appear inline in configuration values that
are logged, returned in tool results, or echoed in errors. The first slice
([CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md))
targets environment and mounted-file providers only; Vault and Kubernetes
Secret integrations follow [CON-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-002.md).

## Validation

When profiles exist, configuration validity will feed readiness (`/readyz`) without
contacting queue managers on every probe ([OBS-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/OBS-001.md)).
