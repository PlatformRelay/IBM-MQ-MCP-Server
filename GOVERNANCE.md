# Project governance

IBM MQ MCP Server is an open-source Model Context Protocol server for IBM MQ,
maintained as a personal OSS project under
[github.com/PlatformRelay/IBM-MQ-MCP-Server](https://github.com/PlatformRelay/IBM-MQ-MCP-Server).
This document describes how decisions are made, who is responsible for what, and
how the project continues if the maintainer is unavailable.

## Scope

This governance model applies to:

- The MCP server application, tests, packaging, and documentation in this
  repository
- Release artifacts published to GHCR and GitHub Releases (once FND-003 lands)
- Public documentation (MkDocs / GitHub Pages)

It does **not** cover downstream deployments, fork-specific policies, IBM MQ
licensing for your queue managers, or private runbooks.

## Roles and responsibilities

| Role | Who | Responsibilities |
| --- | --- | --- |
| **Maintainer** | [Konrad Heimel](https://github.com/konih) | Final merge authority, releases, security response, ADR approval, CI and branch policy |
| **Contributor** | Anyone opening a PR or issue | Follow [CONTRIBUTING.md](CONTRIBUTING.md), [DCO](CONTRIBUTING.md#developer-certificate-of-origin-dco), and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md); propose changes via pull request |
| **Security reporter** | External researchers | Report vulnerabilities privately per [SECURITY.md](SECURITY.md) |

The maintainer is the default approver for all pull requests. There is currently
**one** maintainer (bus factor 1). Adding a co-maintainer requires an explicit
update to this document and a recorded decision in an ADR.

## Decision making

| Change type | Process |
| --- | --- |
| **Architecture** (MCP surfaces, policy, mqweb adapter, security posture) | Write or update an [ADR](docs/adr/README.md); maintainer LGTM before merge |
| **Routine fixes and docs** | PR with green CI; maintainer review |
| **Breaking public changes** | ADR + migration notes; only after a tagged release exists |
| **Release tagging** | Maintainer-only; follows [docs/RELEASE.md](docs/RELEASE.md) once published |
| **Security fixes** | Coordinated disclosure per [SECURITY.md](SECURITY.md) |

Disputes on technical direction are resolved by the maintainer after discussion
in the PR or issue. Persistent disagreements may be documented in an ADR with
accepted/rejected alternatives.

## Security contact

Report vulnerabilities **privately** to **konrad.heimel@gmail.com** — do not
open public issues for security-sensitive reports. See [SECURITY.md](SECURITY.md).

## Access continuity and succession

This is a solo-maintainer project today. Continuity measures:

- **Source of truth** — code, docs, and release history live in the public
  GitHub repository; tagged releases and images will publish to GHCR.
- **Recovery materials** — GitHub recovery codes and release credentials are
  stored in an encrypted offline backup accessible only to the maintainer
  (private runbook; not committed).
- **Succession path** — if the maintainer is permanently unavailable, continuity
  proceeds by transferring the repository to a named successor or to a neutral
  GitHub organization. Contact: **konrad.heimel@gmail.com**.

Until a second maintainer is appointed, two-person review and bus-factor ≥ 2
remain documented gaps.

## Related documents

| Document | Purpose |
| --- | --- |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community behavior standards |
| [CONTRIBUTING.md](CONTRIBUTING.md) | PR workflow, DCO, code review |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting |
| [SUPPORT.md](SUPPORT.md) | User support channels |
| [docs/adr/](docs/adr/) | Architecture decisions |
