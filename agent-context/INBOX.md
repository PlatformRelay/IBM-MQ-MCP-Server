# INBOX — ibm-mq-mcp

Operator-facing items only. Agents append; operator answers; agents record in
`decisions.md` and remove answered items.

## Decisions

### ✅ Answered — Bootstrap product intent = Accepted

Operator confirmed 2026-08-05: "The bootstrap proposal is approved."
Recorded in `decisions.md`. Proposal already marked Accepted.

### ✅ Answered — MIT license (ADR-0009)

Operator accepted portfolio MIT posture via OSS/hygiene direction; ADR-0009
Accepted. No revert requested.


### ✅ Answered — ADR-0003 capability model (design questions 5–9)

Operator chose **A — Operation-oriented capabilities** (2026-08-05). Recorded in
`decisions.md`; ADR-0003 Accepted in `docs/adr/0003-capability-model.md`.

## Operator tasks

### ✅ Authorized — push-to-main + remote-branch-delete

Operator 2026-08-05: "I authorize push-to-main and remote-branch-delete for
ibm-mq-mcp." Valid for this `/agent-loop-local` session.

### 🟢 Enable GitHub Pages

Docs workflow deploys to the `github-pages` environment. Enable Pages
(Source: GitHub Actions) if deploy fails.

### 🟢 Optional: Renovate token

Self-hosted Renovate uses `RENOVATE_TOKEN` when present; otherwise
`GITHUB_TOKEN`.

### 🟡 DECIDED — ADR-0004 configuration and secret providers (design questions 10–11)

**Context:** CON-001 needs profile catalog, secret refs, TLS posture, and mqweb
auth methods before implementation.
**Chose:** Env + mounted-file secret refs only in CON-001; K8s/Vault → CON-002.
First-release mqweb auth: HTTP Basic and client-cert mTLS; LDAP and MQ tokens
deferred. TLS verify on by default; custom CA via file ref; `insecureSkipVerify`
opt-in for local Kind only. File-based YAML/JSON catalog; startup validates all
profiles; lazy credential resolution; fail-open unless `--strict-startup`. No
inline secrets in config values.
**Recorded:** ADR-0004 Accepted; unblocks CON-001.

### 🟡 DECIDED — DOC-001 provisional doc semantics

**Context:** ADR-0003/0004/0006 remain open; DOC-001 must ship honest operator
docs without inventing capability/auth/config contracts.
**Chose:** Mark dependent sections **provisional / TBD** with ADR links; example
profiles under `docs/examples/` as illustrative YAML only; version matrix rows
all **Unknown** until FND-004/MSG-001 evidence; tool reference states zero MQ
tools with CI generation noted as future work; skip markdownlint in docs CI
(mkdocs `--strict` is the gate).
**Revert:** Replace provisional pages when ADRs land; fill matrix from FND-004.

### ✅ DECIDED — FND-004 local MQ licensing + MKurator reuse

**Context:** FND-004 needs a disposable local/CI IBM MQ with mqweb; licensing and
provisioning approach were open.
**Chose:** Reuse sibling MKurator Kind stack (`task cluster:up` via
`MKURATOR_ROOT`, default `../mkurator`) — Helm `ibm-messaging/mq-helm`, not
MKurator CRs for the QM. IBM MQ **Advanced for Developers** via
`LICENSE=accept` / Helm `license: accept`; image `icr.io/ibm-messaging/mq` not
redistributed; local/CI non-production only. Optional Docker path via MKurator
`hack/mq-docker`.
**Recorded:** `decisions.md`, `docs/development/local-mq.md`, FND-004 story Done.

### 🟡 DECIDED (awaiting approval) — FND-003 artifacts (DQ 22)

**Context:** FND-003 needs mandatory artifact set before cosign/SBOM release
wiring. Siblings ship GHCR containers (+ Helm for operators). This is an MCP
server, not a Kubernetes operator.
**Options:** A binary-only / B container-only / C binary+container (no Helm) /
D binary+container+Helm.
**Chose:** **C** — CGO-free binary on GitHub Releases + multi-arch GHCR
container with cosign/SBOM/provenance (Kollect/MKurator release pattern minus
Helm). No Helm/Kustomize in v0.
**Revert:** Supersede with ADR amending ADR-0009 delivery section; remove
unwanted artifact jobs from release.yaml.
