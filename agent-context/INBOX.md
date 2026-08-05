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


### 🟡 DECISION — ADR-0003 capability model (design questions 5–9)

**Context:** EPIC-003 (`connection-profiles-and-policy`) and POL-001/POL-002/MSG-003/ADM-* depend on a
deny-by-default capability vocabulary. Design questions 5–9 define how “read-only”, message access,
administration, raw MQSC, and object-level constraints relate. `docs/architecture/proposed-system.md`
hypothesizes operation-oriented capabilities (`inspect`, `browse`, `consume`, `produce`, `administer`,
`execute_mqsc`); openspec lists the same vocabulary but ADR-0003 is not yet accepted.

**Options:**

- **A — Operation-oriented capabilities (Recommended):** Adopt the proposed vocabulary as the only
  first-class grants on each profile. Map tools to a single required capability each. Treat “read-only
  production” as `inspect` (+ optional `browse` without default payloads per Q5/Q15). `produce` and
  `administer` are distinct (Q6). `consume` is separate from `browse` (Q7). `execute_mqsc` is
  off-by-default exceptional capability (Q8). Defer per-object allow/deny to POL-002 / post-v0 (Q9).

- **B — Coarse profile modes:** Operators set one of a few modes (`read`, `write`, `admin`) per
  profile; the server maps modes to internal tool allowlists. Faster to configure but ambiguous for
  browse-vs-metadata, produce-vs-administer, and MQSC; harder to audit and to express “read prod /
  write dev” safely.

- **C — Modes plus explicit overrides:** Start from B’s modes but allow explicit capability overrides
  (e.g. `read` + `consume`). Reduces ambiguity for power users while keeping simple defaults; two
  configuration surfaces and more validation/error taxonomy to maintain.

**Answer / instructions:** _(operator: pick A, B, or C — or variant — then agent records ADR-0003 in
`docs/adr/` and updates `decisions.md`; do not implement POL-001 until Accepted.)_

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

### 🟡 DECIDED — DOC-001 provisional doc semantics

**Context:** ADR-0003/0004/0006 remain open; DOC-001 must ship honest operator
docs without inventing capability/auth/config contracts.
**Chose:** Mark dependent sections **provisional / TBD** with ADR links; example
profiles under `docs/examples/` as illustrative YAML only; version matrix rows
all **Unknown** until FND-004/MSG-001 evidence; tool reference states zero MQ
tools with CI generation noted as future work; skip markdownlint in docs CI
(mkdocs `--strict` is the gate).
**Revert:** Replace provisional pages when ADRs land; fill matrix from FND-004.

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
