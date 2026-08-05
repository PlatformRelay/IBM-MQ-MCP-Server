# Security Policy

IBM MQ MCP Server connects to IBM MQ queue managers, enforces per-profile
capabilities, and may handle credentials and message payloads. Security is a
first-class concern.

Threat-model and assurance documentation grow with DOC-001; architectural
boundaries live under `docs/architecture/` and `docs/adr/`.

## Reporting a vulnerability

Please report suspected vulnerabilities **privately** — do not open a public
issue for security problems.

- Use **GitHub Security Advisories** ("Report a vulnerability") on this
  repository, or email **konrad.heimel@gmail.com** privately.
- Include affected version/commit, a description, reproduction steps, and impact.
- You will receive an acknowledgement; fixes for confirmed issues are prioritised
  and disclosed once a fix is available.

This is a personal project without a formal SLA, but security reports are taken
seriously and handled promptly.

## Supported versions

The project is pre-1.0. Only the latest released version and the default branch
receive fixes. Public contracts may change between pre-1.0 releases.

## Security posture (target)

- **Deny by default**: capabilities not explicitly granted are rejected locally
  before downstream IBM MQ I/O.
- **Profile trust boundary**: every queue-manager profile is an independent
  policy, credential, and audit unit.
- **No credential leakage**: secrets stay out of configuration values that are
  logged, tool results, and error strings.
- **Typed operations**: prefer schema-backed tools over arbitrary MQSC; raw MQSC
  (if ever enabled) is an exceptional, separately gated capability (ADR-0008).
- **Supply chain**: pinned CI action SHAs, committed lockfiles when Go exists,
  secret scanning, CodeQL (after FND-001), OpenSSF Scorecard, and release image
  signing with cosign + SBOM/provenance (FND-003, pattern from Kollect/MKurator).

## Community

- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) — community behavior standards
- [GOVERNANCE.md](GOVERNANCE.md) — roles, decision making, security contact
- [CONTRIBUTING.md](CONTRIBUTING.md) — how to contribute (includes DCO)
- [SUPPORT.md](SUPPORT.md) — where to get help
