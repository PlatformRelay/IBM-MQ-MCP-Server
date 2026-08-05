# ADR-0009: MIT license and OSS maturity baseline

**Status:** Accepted  
**Date:** 2026-08-05

## Context

Design question 21 asked whether this project should be public open source under
Apache-2.0. Sibling PlatformRelay projects **Kollect** and **MKurator** ship as
public MIT-licensed Go repositories with community files, SHA-pinned Actions,
OpenSSF Scorecard, CodeQL (once Go exists), Dependabot + Renovate, and release
pipelines that sign images with cosign and attach SBOM/provenance.

Operators need the same trust signals here: license clarity, vulnerability
reporting, contributor norms, and CI posture — without copying Kubernetes
operator machinery (Helm, kind e2e, RBAC audits, CRD verify).

## Decision

1. **License:** MIT, copyright Konrad Heimel 2026 — matching Kollect/MKurator
   (not Apache-2.0).
2. **Community pack:** root `LICENSE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`,
   `CONTRIBUTING.md` (DCO), `GOVERNANCE.md`, `SUPPORT.md`, `CODEOWNERS`,
   `.editorconfig`, issue/PR templates.
3. **Dependency bots:** Dependabot (Actions, Go, docs pip) plus Renovate for
   annotated Taskfile/CI pins (Kollect/MKurator pattern).
4. **CI posture (phase 1):** gitleaks + OSS hygiene file gate + MkDocs Docs
   workflow + OpenSSF Scorecard. CodeQL, golangci-lint, coverage floors, and
   `task verify/lint/test/scrub` arrive with **FND-001** / **FND-002**.
5. **Release supply chain (phase 2 — FND-003):** copy the Kollect/MKurator
   release pattern — multi-arch GHCR push, Trivy, cosign keyless signing,
   BuildKit SBOM/`provenance: mode=max`, SPDX SBOM artifact, `actions/attest`
   SLSA + SBOM attestations, signed release blobs. Defer Helm/operator-only
   pieces.

Delivery artifacts for v0 (binary and/or container) remain a product choice
owned by FND-003; this ADR only locks license and the OSS maturity baseline.

## Consequences

### Positive

- Public GitHub can show license, Scorecard, Docs, and CI badges immediately.
- Contributors have clear conduct, DCO, and support channels.
- Release signing work has an explicit sibling pattern to copy.

### Negative

- MIT (vs Apache-2.0) has no explicit patent grant; accepted for PlatformRelay
  consistency.
- Full supply-chain attestations wait on a releasable binary/image (FND-003).

## Alternatives

### Apache-2.0

Rejected for baseline. Stronger patent language, but breaks consistency with
Kollect/MKurator and complicates SPDX headers across the portfolio.

### Stay design-only without community files

Rejected. Blocks external trust and OpenSSF-oriented hygiene the operator
requested.

## Validation implications

- CI `oss-hygiene` job fails if required community files disappear.
- Docs workflow builds MkDocs `--strict`.
- Scorecard publishes results for the public repo.
- FND-002 must add CodeQL + Go quality gates; FND-003 must add cosign/SBOM.
