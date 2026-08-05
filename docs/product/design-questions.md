# Design questions

Questions are ordered so foundational decisions are resolved before detailed
tool design. Approved answers will be recorded in ADRs and linked stories.

## Architecture

1. Which implementation ecosystem should be the baseline: Go, Python, or
   TypeScript?
2. Is mqweb REST sufficient for the first release, with PCF explicitly deferred?
3. Should this be a new implementation that uses IBM's sample as prior art, or
   should contribution upstream be a release goal?
4. Which Streamsy repository and which local sibling projects should be assessed?

## Authorization semantics

5. ~~Does “read-only” include message payload browsing or only metadata and
   administrative inspection?~~
   **Answered (ADR-0003):** “Read-only production” is `inspect` plus optional
   `browse`; browse does not imply default payloads (see also Q15).
6. ~~Does “write-only” mean message production only, or also object
   administration?~~
   **Answered (ADR-0003):** `produce` and `administer` are distinct capabilities;
   neither implies the other.
7. ~~Should destructive consume/get be a separate capability from browse?~~
   **Answered (ADR-0003):** Yes — `consume` is separate from `browse`.
8. ~~Should raw MQSC be omitted, allowlisted, or available behind an exceptional
   capability?~~
   **Answered ([ADR-0008](../adr/0008-raw-mqsc-policy.md)):** absent by default;
   double opt-in (`--enable-mqsc` + profile `execute_mqsc`); v0 verb allowlist
   is read-only (`DISPLAY`/`DIS`/`PING`); parse and validate before mqweb I/O;
   audit redacted command text. Default config registers no MQSC tool.
9. ~~Are profile-level permissions sufficient initially, or are queue/object name
   allowlists required in the first release?~~
   **Answered (ADR-0003):** Profile-level capabilities are sufficient for v0;
   per-object allow/deny is deferred to POL-002 / post-v0.

## Connectivity and identity

10. ~~Which downstream authentication methods are required for the first release:
    basic, mTLS, LDAP-backed basic, MQ authentication tokens, or combinations?~~
    **Answered (ADR-0004):** First release supports **HTTP Basic** (username +
    password via secret refs) and **client-certificate mTLS** (cert/key via file
    refs). LDAP-backed basic and MQ authentication tokens deferred to later ADRs.
11. ~~Which secret stores must be supported first: environment, mounted files,
    Kubernetes Secrets, Vault, or another provider?~~
    **Answered (ADR-0004):** First release (**CON-001**) supports **environment
    variables** and **mounted files** only; Kubernetes Secrets and Vault →
    [CON-002](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/CON-002.md).
12. Is remote Streamable HTTP a first-release deployment target, and if so which
    MCP-client authentication/authorization model is required?
13. Which IBM MQ versions and platforms must be supported, including z/OS?

## Message safety

14. ~~What default and maximum browse counts and payload sizes are acceptable?~~
    **Answered (MSG-001, ADR-0005):** Default browse count **10**, server max **100**;
    default max payload bytes per message **4096**, hard max **65536**.
15. ~~May payloads be returned by default, or only metadata until explicitly
    requested?~~
    **Answered (MSG-001, ADR-0003):** **Metadata only by default**; payloads when
    `includePayload=true` (requires `browse`).
16. ~~Which payload encodings and redaction rules are required?~~
    **Answered (MSG-001):** UTF-8 text preferred; binary as base64 with `encoding`;
    secret-like patterns redacted before results; payloads never logged.
17. ~~Should put operations accept arbitrary bytes, text/JSON only, or named
    content types with validation?~~
    **Answered (MSG-002):** Named content types only — `text/plain`,
    `application/json` (must parse as JSON), and `application/octet-stream`
    (base64-encoded input decoded before put). Other types are rejected locally
    with a typed error. Max payload **65536** bytes enforced before any network
    call. Results return message/correlation IDs only; payloads are never logged.

## MKurator coexistence

18. Is awareness advisory only, or should the server be able to produce/apply
    MKurator custom-resource changes?
19. How should MKurator be discovered: Kubernetes API, explicit configuration,
    or supplied ownership metadata?
20. Should direct administration of an MKurator-managed object be blocked or
    merely warned?

## Product and delivery

21. ~~Is this intended for public open source under Apache-2.0?~~
    **Answered (ADR-0009):** public open source under **MIT**, matching
    Kollect/MKurator (not Apache-2.0).
22. ~~Which deployment artifacts are mandatory: standalone binary/package,
    container, Helm chart, Kustomize, or all?~~
    **Answered (DQ 22 / FND-003):** mandatory **binary** (GitHub Releases) +
    **container** (multi-arch GHCR) with cosign/SBOM/provenance; **no Helm or
    Kustomize in v0** (ADR-0009 delivery detail).
23. Which clients are release targets: GitHub Copilot, desktop MCP hosts, IBM
    Bob, watsonx Orchestrate, VS Code, or others?
24. What release quality threshold and compatibility policy should govern v0.x
    and v1?

