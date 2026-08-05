# Example profiles

!!! note "Provisional schema — ADR-0004 Accepted"
    These files illustrate **mixed read/write grants across profiles** using
    placeholder hostnames and secret references. The field layout follows
    [ADR-0004](../adr/0004-configuration-and-secrets.md) and
    [Configuration](../configuration.md); capability enforcement lands in
    [POL-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/POL-001.md).

All examples are **secret-free**: credentials are referenced by name only.

## Files

| File | Purpose |
| --- | --- |
| [profile-read-only.yaml](profile-read-only.yaml) | Production-style inspect + browse only |
| [profile-mixed-grants.yaml](profile-mixed-grants.yaml) | Read-only production + write-capable development profile |
| [profile-kind-local.yaml](profile-kind-local.yaml) | Local Kind IBM MQ via MKurator (`QM1`, `https://mq.localhost:30443`) |

## Usage

```bash
export MQ_PROD_CREDENTIALS='operator:secret'   # from your secret store — never commit
ibm-mq-mcp --config /etc/ibm-mq-mcp/profiles.yaml
# or
export IBM_MQ_MCP_CONFIG=/etc/ibm-mq-mcp/profiles.yaml
ibm-mq-mcp
```

Use `--strict-startup` when every profile in the file must validate before the
process listens.

## MKurator note

If [INT-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/INT-001.md)
adds ownership discovery, profiles remain independent of Kubernetes — see
[MKurator coexistence](../support/mkurator-coexistence.md).
