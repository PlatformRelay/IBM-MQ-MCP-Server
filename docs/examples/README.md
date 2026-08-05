# Example profiles

!!! warning "Illustrative only — ADR-0003 / ADR-0004 open"
    These files demonstrate **mixed read/write grants across profiles** using
    placeholder hostnames and secret references. They are **not** a supported
    configuration schema until [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md)
    and [ADR-0004](../adr/README.md#decision-queue) land. Capability names follow
    the hypothetical vocabulary in [Policy](../policy.md).

All examples are **secret-free**: credentials are referenced by name only.

## Files

| File | Purpose |
| --- | --- |
| [profile-read-only.yaml](profile-read-only.yaml) | Production-style inspect + browse only |
| [profile-mixed-grants.yaml](profile-mixed-grants.yaml) | Read-only production + write-capable development profile |
| [profile-kind-local.yaml](profile-kind-local.yaml) | Local Kind IBM MQ via MKurator (`QM1`, `https://mq.localhost:30443`) |

## Usage (future)

When profile loading is implemented, the expected operator flow will resemble:

```bash
export MQ_PROD_CREDENTIALS='...'   # from your secret store — never commit
export MQ_DEV_CLIENT_CERT=/run/secrets/dev-client.pem
ibm-mq-mcp --config /etc/ibm-mq-mcp/profiles.yaml
```

Exact flags and file paths are **TBD** ([CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md)).

## MKurator note

If [INT-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INT-001.md)
adds ownership discovery, profiles remain independent of Kubernetes — see
[MKurator coexistence](../support/mkurator-coexistence.md).
