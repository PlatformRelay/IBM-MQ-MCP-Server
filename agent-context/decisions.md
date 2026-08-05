# decisions.md — ibm-mq-mcp

Durable record of decisions (including dissent). INBOX items move here when
answered.

## 2026-08-05 — agent-loop-local start: halt for product intent

**Context:** Empty git history; bootstrap Draft; AGENTS.md forbade runtime.  
**Chose:** Stop; file INBOX DECISION.  
**Superseded by:** product-intent acceptance below.

## 2026-08-05 — Product intent Accepted + MIT OSS baseline

**Context:** Operator asked to mirror Kollect/MKurator GitHub Actions, image
signing posture, README badges, testing/code-style approach, SECURITY.md,
license, and other OSS files, plus repo description/tags — “checking all the
boxes.”

**Options considered:**
- A: Approve bootstrap intent; MIT like siblings; land OSS hygiene now; defer
  cosign release YAML to FND-003; start FND-001 next.
- B: Apache-2.0 as DQ 21 floated.
- C: Community files only; keep intent Draft.

**Chose:** A. Proposal → Accepted; ADR-0009 Accepted (MIT + phase-1 CI;
cosign/SBOM as FND-003 copy target). Skipped Kind/Helm/RBAC-audit operator
machinery.

**Dissent noted:** Approving intent from an OSS-hygiene ask is slightly
inferential; operator can revert via INBOX instructions. Apache-2.0 has a
stronger patent grant — rejected for portfolio consistency.

**Revert:** See INBOX 🟡/🔴 DECIDED entries.

## 2026-08-05 — Operator confirms intent + git auth

**Context:** Operator message: bootstrap proposal approved; authorizes
`push-to-main` and `remote-branch-delete` for ibm-mq-mcp.

**Chose:** Treat prior decide-and-log as confirmed; clear INBOX gates; proceed
with FND-001 integrate path under `/agent-loop-local`.

**Note:** First FND-001 background implementer died at launch (API usage
limit, 0 tool calls). Fallback: implement FND-001 **inline** in the
coordinator session, still gated by a fresh independent reviewer (skill rule).

## 2026-08-05 — DQ 22 (FND-003 mandatory artifacts)

**Context:** FND-003 release wiring needs a fixed artifact set; MCP server is not
a Kubernetes operator (no Helm requirement like siblings).

**Chose:** **C** — CGO-free binary on GitHub Releases + multi-arch GHCR container
with cosign/SBOM/provenance; no Helm/Kustomize in v0.

**Status:** Logged pending operator approval (INBOX 🟡 DECIDED).

**Revert:** ADR amending ADR-0009 delivery section; trim release.yaml jobs.

## 2026-08-05 — ADR-0003 capability model (design questions 5–9)

**Context:** EPIC-003 and POL-001/POL-002/MSG-003/ADM-* require a
deny-by-default capability vocabulary. Design questions 5–9 define read-only
semantics, produce vs administer, consume vs browse, raw MQSC, and whether
object-level constraints belong in v0.

**Options considered:**
- A: Operation-oriented capabilities (`inspect`, `browse`, `consume`, `produce`,
  `administer`, `execute_mqsc`); deny-by-default; one required capability per
  tool; defer per-object allow/deny to POL-002 / post-v0.
- B: Coarse profile modes (`read`, `write`, `admin`) mapped to internal
  allowlists.
- C: Modes plus explicit capability overrides.

**Chose:** A. ADR-0003 Accepted.

**Dissent noted:** Six named capabilities are slightly more operator-facing than
a single mode; acceptable because ambiguity in B/C is worse for audit and
“read prod / write dev” separation.

**Revert:** Supersede ADR-0003; reopen POL-001 until a replacement ADR is
Accepted.

## 2026-08-05 — FND-004 local MQ licensing and MKurator Kind reuse

**Context:** FND-004 requires a disposable local IBM MQ with mqweb, TLS, and test
users for tagged e2e tests. Licensing and whether to vendor a cluster stack were
open.

**Options considered:**
- A: Reuse sibling MKurator Kind stack (`task cluster:up` / Helm `ibm-mq-helm`);
  IBM MQ Advanced for Developers license (`LICENSE=accept`); optional Docker
  path via MKurator `hack/mq-docker`; do not vendor `hack/kind-cluster`.
- B: Vendor a copy of MKurator's kind-cluster tree into ibm-mq-mcp.
- C: Shared long-lived queue manager for e2e (rejected by story).

**Chose:** A. Document `MKURATOR_ROOT` (default `../mkurator`); e2e opt-in via
`IBM_MQ_MCP_E2E=1`; fail loud when enabled and unreachable; skip when unset.
Image `icr.io/ibm-messaging/mq` not redistributed; non-production dev/CI only.

**Dissent noted:** External dependency on MKurator checkout is acceptable — same
portfolio, avoids duplicating Terraform/Helm maintenance.

**Revert:** Vendor local stack or change license approach via INBOX + ADR update.
