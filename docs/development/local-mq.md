# Local IBM MQ for development and e2e

Authority: [FND-004](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/FND-004.md).

This repository does **not** vendor a full Kind cluster stack. Local IBM MQ with
**mqweb**, TLS, and test users comes from a **sibling [MKurator](https://github.com/PlatformRelay/MKurator)
checkout** via Helm (`ibm-messaging/mq-helm`) on Kind — not MKurator custom
resources for the queue manager itself.

!!! warning "IBM MQ Advanced for Developers — non-production only"
    The upstream container image `icr.io/ibm-messaging/mq` is **not**
    redistributed in this repository. Local and CI use the **IBM MQ Advanced for
    Developers** license by setting `LICENSE=accept` (Docker) or Helm
    `license: accept` (Kind). That license permits **development and test**
    workloads only — **not production**. See IBM's license terms before pulling
    the image.

## Prerequisites

From the MKurator docs ([LOCAL_SETUP.md](https://github.com/PlatformRelay/MKurator/blob/main/docs/LOCAL_SETUP.md)):

- Docker (or compatible runtime), `kind`, `kubectl`, `helm`, `terraform`, `mkcert`, `task`
- Network access to pull `icr.io/ibm-messaging/mq` (currently `9.4.5.1-r1` in MKurator's Kind stack)
- A sibling checkout of MKurator (default path `../mkurator` relative to this repo)

Set `MKURATOR_ROOT` if your checkout lives elsewhere:

```bash
export MKURATOR_ROOT=/path/to/mkurator
```

## Primary path — Kind via MKurator

From **this repository**:

```bash
task mq:kind:up      # delegates to: task -d $MKURATOR_ROOT cluster:up
task mq:kind:info    # re-print URLs and credentials
```

First bring-up often takes **5–15 minutes** while Terraform applies ingress,
cert-manager, monitoring, and IBM MQ.

### Endpoints (after bring-up)

| What | URL | Notes |
| --- | --- | --- |
| mqweb console | `https://mq.localhost:30443/ibmmq/console/` | mkcert TLS on NodePort 30443 |
| Admin REST | `https://mq.localhost:30443/ibmmq/rest/v3/admin/qmgr` | v3 preferred; v2 also present |
| Messaging REST | `https://mq.localhost:30443/ibmmq/rest/v3/messaging/qmgr/QM1/queue/...` | browse/produce paths |
| In-cluster | `https://ibm-mq.ibm-mq.svc:9443` | for pods inside the Kind cluster |

Queue manager name: **`QM1`**.

### Credentials (local dev defaults)

MKurator's Kind Terraform sets the mqweb admin password. For local development
the documented default is:

| Field | Value |
| --- | --- |
| User | `admin` |
| Password | `passw0rd` (override with `MQ_ADMIN_PASSWORD` before `cluster:up`) |

**Do not commit real passwords.** Export credentials from your environment or
secret store when running e2e tests:

```bash
export MQ_ADMIN_PASSWORD="${MQ_ADMIN_PASSWORD:-passw0rd}"   # local dev only
```

TLS uses a mkcert wildcard for `*.localhost`. Run `mkcert -install` once on
your workstation if browsers or clients reject the certificate.

### Teardown — no persistent state

Destroy the disposable cluster when finished:

```bash
task mq:kind:down    # delegates to: task -d $MKURATOR_ROOT cluster:down
```

This runs Terraform destroy, deletes the Kind cluster, and wipes MKurator's
`.state` under `hack/kind-cluster/`. No queue manager should remain on your
machine after a successful `cluster:down`.

MKurator reference: [`hack/kind-cluster/README.md`](https://github.com/PlatformRelay/MKurator/blob/main/hack/kind-cluster/README.md).

## Optional path — Docker Compose (lighter)

For faster iteration without Kind, use MKurator's standalone Docker MQ
(`hack/mq-docker`):

```bash
task mq:docker:up
task mq:docker:wait
# endpoint: https://127.0.0.1:9443  (no Host header required)
task mq:docker:down
```

Same `QM1` / `admin` / `passw0rd` defaults and `LICENSE=accept` apply. See
[MKurator mq-docker README](https://github.com/PlatformRelay/MKurator/blob/main/hack/mq-docker/README.md).

When using Docker instead of Kind, point e2e env at the Docker endpoint:

```bash
export IBM_MQ_MCP_MQ_ENDPOINT=https://127.0.0.1:9443
export IBM_MQ_MCP_MQ_HOST=                  # empty — no ingress Host header
```

## End-to-end tests (opt-in)

E2e tests live under `test/e2e/` with build tag `e2e`. They are **not** part of
`task check` or default CI.

| Variable | Default (Kind) | Purpose |
| --- | --- | --- |
| `IBM_MQ_MCP_E2E` | unset | Set to `1` to run e2e tests |
| `IBM_MQ_MCP_MQ_ENDPOINT` | `https://127.0.0.1:30443` | mqweb base URL |
| `IBM_MQ_MCP_MQ_HOST` | `mq.localhost` | HTTP `Host` for Kind ingress |
| `IBM_MQ_MCP_MQ_QMGR` | `QM1` | Queue manager name |
| `IBM_MQ_MCP_MQ_USER` | `admin` | mqweb user |
| `MQ_ADMIN_PASSWORD` / `IBM_MQ_MCP_E2E_PASSWORD` | `passw0rd` if unset | mqweb password; e2e reads `MQ_ADMIN_PASSWORD` then `IBM_MQ_MCP_E2E_PASSWORD` |
| `IBM_MQ_MCP_MQ_INSECURE_TLS` | `true` | Skip TLS verify for mkcert local dev |

**Behaviour:**

- `IBM_MQ_MCP_E2E` unset → tests **skip** (`t.Skip`).
- `IBM_MQ_MCP_E2E=1` and MQ unreachable → tests **fail** (CI fails loud).
- `IBM_MQ_MCP_E2E=1` and MQ reachable → Admin REST + Messaging REST reachability asserted.

Full local smoke:

```bash
task mq:kind:up
export IBM_MQ_MCP_E2E=1
export MQ_ADMIN_PASSWORD="${MQ_ADMIN_PASSWORD:-passw0rd}"
task test:e2e
task mq:kind:down
```

Or with Docker:

```bash
task mq:docker:up && task mq:docker:wait
export IBM_MQ_MCP_E2E=1
export IBM_MQ_MCP_MQ_ENDPOINT=https://127.0.0.1:9443
export IBM_MQ_MCP_MQ_HOST=
task test:e2e
task mq:docker:down
```

## Example profile

See [profile-kind-local.yaml](../examples/profile-kind-local.yaml) for a
secret-free illustrative profile pointing at the Kind endpoint with
`${env}`-style secret references.

## CI note

Default GitHub Actions jobs do **not** provision IBM MQ (license + runtime
cost). A future workflow may opt in with `IBM_MQ_MCP_E2E=1` after
`task mq:kind:up` on a self-hosted or scheduled runner. Until then, e2e remains
contributor-local.

## Related pages

- [CI gates](ci-gates.md)
- [Version matrix](../support/version-matrix.md)
- [MKurator coexistence](../support/mkurator-coexistence.md)
- [Example profiles](../examples/README.md)
