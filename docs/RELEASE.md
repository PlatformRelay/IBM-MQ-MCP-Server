# Release process

Release automation (multi-arch container, cosign keyless signing, SBOM,
provenance attestations, GitHub Release assets) is delivered by **FND-003** and
follows the supply-chain pattern used by Kollect and MKurator. See
[ADR-0009](adr/0009-license-and-oss-maturity.md).

Until FND-003 lands:

- Do not publish version tags expecting signed artifacts.
- Prefer documenting changes in git history with gitmoji conventional commits.
- Changelog tooling (git-cliff) may be added with FND-002/FND-003.

## Intended release shape (target)

1. Green CI on `main` (verify, lint, tests, docs, Scorecard/CodeQL once present).
2. Maintainer cuts `vX.Y.Z`.
3. Release workflow builds and pushes `ghcr.io/platformrelay/ibm-mq-mcp`.
4. Trivy gate → cosign sign digest → SPDX SBOM → SLSA/SBOM attestations.
5. GitHub Release publishes checksums and cosign-signed blobs.

Exact image name and whether a standalone binary is also published are decided
in FND-003.
