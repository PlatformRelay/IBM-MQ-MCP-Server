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

5. Does “read-only” include message payload browsing or only metadata and
   administrative inspection?
6. Does “write-only” mean message production only, or also object
   administration?
7. Should destructive consume/get be a separate capability from browse?
8. Should raw MQSC be omitted, allowlisted, or available behind an exceptional
   capability?
9. Are profile-level permissions sufficient initially, or are queue/object name
   allowlists required in the first release?

## Connectivity and identity

10. Which downstream authentication methods are required for the first release:
    basic, mTLS, LDAP-backed basic, MQ authentication tokens, or combinations?
11. Which secret stores must be supported first: environment, mounted files,
    Kubernetes Secrets, Vault, or another provider?
12. Is remote Streamable HTTP a first-release deployment target, and if so which
    MCP-client authentication/authorization model is required?
13. Which IBM MQ versions and platforms must be supported, including z/OS?

## Message safety

14. What default and maximum browse counts and payload sizes are acceptable?
15. May payloads be returned by default, or only metadata until explicitly
    requested?
16. Which payload encodings and redaction rules are required?
17. Should put operations accept arbitrary bytes, text/JSON only, or named
    content types with validation?

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
22. Which deployment artifacts are mandatory: standalone binary/package,
    container, Helm chart, Kustomize, or all?
    *(Partial: OSS baseline and cosign/SBOM release pattern locked in ADR-0009;
    exact artifact set remains FND-003.)*
23. Which clients are release targets: GitHub Copilot, desktop MCP hosts, IBM
    Bob, watsonx Orchestrate, VS Code, or others?
24. What release quality threshold and compatibility policy should govern v0.x
    and v1?

