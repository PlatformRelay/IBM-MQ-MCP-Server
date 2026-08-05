## Container image

```
${IMAGE_REPO}:${VERSION}
```

Multi-arch (`linux/amd64`, `linux/arm64`), distroless nonroot base. Runs the MCP server over stdio.

OCI attestations (SBOM + SLSA provenance) are attached in GHCR. Verify the signature:

```sh
cosign verify \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/${GITHUB_REPOSITORY}/.+' \
  ${IMAGE_REPO}@${IMAGE_DIGEST}
```

## Standalone binary

Download the tarball for your platform from the assets below, extract `ibm-mq-mcp`, and run:

```sh
./ibm-mq-mcp
```

Verify checksums with `sha256sum -c checksums.txt`.
