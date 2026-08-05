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
