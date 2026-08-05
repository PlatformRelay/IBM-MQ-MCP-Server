# Deployment

Shipped artifacts follow [ADR-0009](adr/0009-license-and-oss-maturity.md) and
[FND-003](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/FND-003.md):
CGO-free **binary** tarballs on GitHub Releases plus a multi-arch **container**
on GHCR. **No Helm chart or Kustomize** in v0.

Full release mechanics: [RELEASE.md](RELEASE.md).

## Binary (GitHub Releases)

After the first tagged release (`vX.Y.Z`), download the tarball for your
platform from the [GitHub Releases](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/releases)
page. Assets are cosign-signed with SBOM and provenance attestations — see
[RELEASE.md](RELEASE.md#verify-a-release).

Local equivalent:

```bash
task build
./bin/ibm-mq-mcp
```

## Container (GHCR)

Published image (once tagged):

```text
ghcr.io/platformrelay/ibm-mq-mcp:<version>
```

Local smoke build:

```bash
task docker:build
docker inspect --format '{{.Config.User}}' ibm-mq-mcp:dev   # expect 65532:65532
```

The image uses `gcr.io/distroless/static:nonroot`, runs as UID **65532**, and
builds with `CGO_ENABLED=0`.

### Run container (stdio MCP)

MCP stdio is the default transport. Example (host networking varies by platform):

```bash
docker run --rm -i ghcr.io/platformrelay/ibm-mq-mcp:0.1.0
```

Connect your MCP host to the container stdin/stdout. MQ profiles and secrets
are **not** loaded in the bootstrap skeleton — mount configuration and secret
files when [CON-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-001.md)
defines the contract.

### Ops HTTP in containers

Enable probes and metrics on a dedicated port (not the MCP transport):

```bash
docker run --rm -p 9090:9090 \
  -e IBM_MQ_MCP_OPS_ADDR=:9090 \
  ghcr.io/platformrelay/ibm-mq-mcp:0.1.0
```

Kubernetes `livenessProbe` / `readinessProbe` and `securityContext.readOnlyRootFilesystem`
guidance are deferred to deployment hardening after remote transport decisions
([ADR-0006](adr/README.md#decision-queue)) — the v0 image does not define a
Docker `HEALTHCHECK`.

## Remote MCP transport (TBD)

Streamable HTTP as a first-class deployment target is **not decided**
([ADR-0006](adr/README.md#decision-queue)). Until then, document and deploy
**stdio** integrations only.

## What we deliberately omit in v0

| Artifact | Status |
| --- | --- |
| Helm chart | Out of scope ([RELEASE.md](RELEASE.md)) |
| Kustomize manifests | Out of scope |
| Kubernetes Operator | Non-goal ([feature scope](product/feature-scope.md)) |

## Related pages

- [Quickstart](quickstart.md)
- [Observability](observability.md)
- [Upgrade](upgrade.md)
- [Version support matrix](support/version-matrix.md)
