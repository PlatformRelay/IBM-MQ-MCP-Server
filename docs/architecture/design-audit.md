# Design audit — 2026-08-05

Scope: bootstrap proposal, proposed system, ADR-0001/0002, decision queue,
feature scope, reference assessment, roadmap, and all stories as of
2026-08-05. Actions taken by this audit are marked **[resolved]** and were
delivered as the epic/roadmap restructure and the per-epic change proposals
under `openspec/changes/`.

## Strengths

- Clear trust-boundary framing: profile as the unit of selection, policy,
  rate limiting, and audit; deny-by-default; policy before I/O.
- Honest prior-art assessment: the IBM sample's `runmqsc` anti-pattern is
  named and rejected; MKurator patterns are reused without runtime coupling.
- ADR discipline with an explicit decision queue and immutable records.
- ADR-0002 guardrails refuse to overclaim REST semantics (no syncpoint or
  exactly-once claims).

## Findings

### Coverage gaps

1. **Observability had no owning story** although feature scope lists health,
   metrics, and structured logs as must/strong. A server that cannot be
   probed or measured fails the "production-grade" outcome. **[resolved]** →
   OBS-001 under EPIC-007.
2. **The live IBM MQ environment had no owning story** despite the proposal's
   success signal "a local real-MQ environment validates supported operations
   end to end" and its dependency on a test topology. **[resolved]** →
   FND-004 under EPIC-002.
3. **Design questions 21–24** (license, artifacts, clients, quality
   threshold) blocked FND delivery but had no slot in the ADR queue, which
   ended at ADR-0008. **[resolved]** → ADR-0009 queued; FND-003 blocked on it.
4. **MCP resources and prompts** are mentioned in the proposed system but
   owned by no story or change. **[resolved]** → explicitly fenced out of
   `typed-inspection-and-output` until tool contracts stabilize.
5. The threat model and IBM MQ/mqweb **version-support matrix** are promised
   by DOC-001 and ADR-0002 but not yet named as deliverables. **[resolved]** →
   recorded in the DOC-001 session log.

### Sizing violations (INVEST)

6. **FND-001** spanned skeleton, CI, container, supply chain, and release —
   at least four reviewable slices. **[resolved]** → FND-001/002/003/004.
7. **MSG-001** bundled three risk classes its own session log declared
   separate. **[resolved]** → split by verb: browse (MSG-001), produce
   (MSG-002), consume (MSG-003), with the mqweb spike as entry gate.
8. **INS-001** spanned every object family plus diagnostics. **[resolved]** →
   INS-001/002/003, with the shared collection contract owned by INS-001.
9. **SEC-001** bundled transport authorization, hardening, and audit.
   **[resolved]** → SEC-001/SEC-002/OBS-001.
10. **ADM-001** was blocked by three ADRs at once; the raw-MQSC decision
    (ADR-0008) held queue administration hostage. **[resolved]** →
    ADM-001/002/003; ADM-001 now blocks only on ADR-0003/0007.

### Consistency issues

11. **Duplicate audit ownership**: POL-001 and SEC-001 both claimed audit
    records, inviting two schemas. **[resolved]** → POL-001 emits decision
    events; SEC-002 is the single audit owner.
12. **Capability vocabulary lives in two places** (proposed system and future
    ADR-0003). Acceptable while a hypothesis, but ADR-0003 must become the
    single authority on acceptance; stories now say so.
13. EPIC-001 hard-coded "ADR-0001 through ADR-0008"; the queue is now
    open-ended. **[resolved]** → updated to ADR-0009.
14. The roadmap marked INS-001 "Open" without surfacing its CON-001/POL-001
    dependency. **[resolved]** → dependencies stated on the roadmap line.
15. `openspec/README.md` promises a `specs/` directory that does not exist
    yet. Acceptable: it is created when the first change is archived. No
    action.

### Feasibility and security risks (open — need decisions or evidence)

16. **mqweb Messaging REST browse semantics are unproven.** Non-destructive
    browse, get-by-id, payload formats, and standalone-versus-full-install
    differences may not meet the contract. This is the highest product risk
    in EPIC-005; the MSG-001 spike is deliberately the entry gate, and
    "browse ships only if proven non-destructive" is an acceptance criterion.
17. **ADR-0006 remains the widest open decision** (remote HTTP as a
    first-release target). It blocks SEC-001/002 and shapes FND-001 transport
    tests; resolving it early avoids rework of the minimal server.
18. **Reason-code reference data licensing** (INS-003) needs a source that
    can be redistributed; flagged in the story.
19. **Probe amplification**: health checks must not hammer queue managers;
    made an explicit OBS-001 criterion.

## Recommended decision order

1. Approve product intent, then the epic change proposals.
2. ADR-0003 (capability model) — unblocks POL, MSG, ADM semantics.
3. ADR-0004 (config/secrets) — unblocks CON-001.
4. ADR-0009 (license/delivery targets) — unblocks FND-003 and publishing.
5. ADR-0006 (transports/client auth) — before the FND-001 transport tests
   harden.
6. ADR-0007, ADR-0008 — before ADM work begins.
