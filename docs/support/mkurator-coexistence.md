# MKurator coexistence

IBM MQ MCP Server targets **generic IBM MQ** deployments. [MKurator](https://github.com/PlatformRelay/MKurator)
(Kubernetes operator for declarative MQ) is an optional coexistence partner — not
a runtime prerequisite.

!!! note "ADR-0007 accepted"
    v0 is **advisory only**: the MCP server does not create or apply MKurator
    custom resources. Ownership is declared in profile catalog metadata; live
    Kubernetes discovery is deferred. See [ADR-0007](../adr/0007-mkurator-coexistence.md).

## Principles

| Principle | Detail |
| --- | --- |
| No Kubernetes required | The same binary must operate for non-Kubernetes MQ estates |
| No reconciliation duplication | This server does not replace MKurator controllers |
| Explicit degradation | Missing ownership metadata is treated as unmanaged |
| Policy before mutation | INT-001 pre-mutation hook runs before ADM-001 queue mutations |

## v0 behaviour

- **Catalog ownership** — `mkurator.managedObjects` on a profile lists queue
  name patterns (exact or `PREFIX*`) managed declaratively.
- **Object tags (stub)** — queue descriptions prefixed with
  `mkurator.platformrelay.io/managed=` supply ownership when present.
- **Mutation policy** — `mkurator.mutationPolicy` defaults to `warn`; set
  `block` to fail closed before mqweb I/O.
- **No CR apply** — mutations remain imperative mqweb calls; operators reconcile
  via MKurator/GitOps separately.

## Configuration example

```yaml
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example:9443
    authentication: { type: basic, secretRef: env:MQ_SECRET }
    tls: { insecureSkipVerify: true }
    capabilities: [inspect, administer]
    mkurator:
      mutationPolicy: warn
      managedObjects:
        - kind: queue
          name: APP.*
```

## Related pages

- [Threat model](../security/threat-model.md)
- [Policy](../policy.md)
- [Tool reference — queue administration](../tools/index.md)
- [Design questions 18–20](../product/design-questions.md)
