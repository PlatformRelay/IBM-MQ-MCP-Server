# ADR-0004: Configuration and secret providers

**Status:** Accepted  
**Date:** 2026-08-05

## Context

CON-001 requires a stable profile catalog, secret-reference scheme, TLS posture,
and startup validation before any mqweb adapter or MCP tool can select a queue
manager. Design questions 10–11 ask which downstream mqweb authentication
methods and secret stores belong in the first release versus later slices.

The bootstrap skeleton runs with empty configuration today. Operators need a
file-based catalog they can mount in Kubernetes or bind locally, with credentials
outside config values and lazy resolution so unused profiles do not require live
secrets at process start.

## Decision

### Secret providers (first release and CON-002)

Support **environment variables**, **mounted files** (CON-001), and **Kubernetes
Secrets** (CON-002):

| Reference prefix | Example | Resolution |
| --- | --- | --- |
| `env:` | `env:MQ_PROD_PASSWORD` | Value of the named environment variable |
| `file:` | `file:/run/secrets/mq/password` | Contents of the mounted file (trimmed trailing newline) |
| `k8s:` | `k8s:mq-system/mq-credentials#password` | Data key from the named Secret in the namespace |

- **No inline secrets** in configuration values that may be logged, returned in
  tool results, or echoed in errors.
- HashiCorp Vault and other external providers remain deferred; unknown reference
  schemes fail catalog validation at startup.

**Kubernetes optional at runtime:** env/file profiles work without in-cluster
config or kubeconfig. A `k8s:` reference fails at lazy resolution time with a
typed, non-secret error when the provider is unavailable or the Secret/key is
missing; other profiles are unaffected.

### Downstream mqweb authentication (first release)

Each profile declares one authentication method for mqweb REST:

| Method | Config shape | Credential source |
| --- | --- | --- |
| HTTP Basic | `authentication.type: basic` + `secretRef` | Username and password encoded as `username:password` in the resolved secret value (env or file) |
| Client-certificate mTLS | `authentication.type: mtls` + `certificateRef` + `privateKeyRef` | PEM certificate and private key from file refs (optional `passphraseRef` for encrypted keys) |

LDAP-backed basic, MQ authentication tokens, and OIDC variants are deferred to
later ADRs unless already trivial during implementation.

### TLS

- **TLS verification on by default** for every profile endpoint.
- Operators may supply a custom CA via `tls.caRef` (`file:` only in this slice).
- `tls.insecureSkipVerify: true` is an explicit opt-in for local Kind and similar
  disposable environments only; it must not appear in production examples.

### Profile catalog

- **File-based YAML or JSON** configuration; path supplied via `--config` flag or
  `IBM_MQ_MCP_CONFIG` environment variable.
- **Startup validates all profiles**: unique names, required fields, TLS settings,
  authentication shape, and resolvable secret *references* (not secret values).
- **Lazy credential resolution**: secret values are read only when a profile is
  first used for downstream I/O, not at catalog parse time.
- **Fail-open startup** (default): one invalid or unreachable profile does not
  prevent healthy profiles from being used; invalid profiles are marked and
  skipped until selected.
- **`--strict-startup`**: any profile validation failure fails process startup.

Application code depends on typed administration and messaging ports with a
per-profile client pool; HTTP to mqweb stays behind the adapter layer (ADR-0002).

## Consequences

### Positive

- CON-001 can ship without Vault SDK dependencies; K8s client-go added in CON-002.
- Env and file refs match common container and local-dev patterns.
- Lazy resolution avoids requiring every profile's secrets at pod start.
- Fail-open startup keeps multi-profile servers usable when one credential
  provider is misconfigured.

### Negative

- Operators must encode basic-auth pairs externally (`user:pass` in one secret).
- No hot reload; catalog changes require restart.
- LDAP and token auth require follow-on ADRs and adapter work.
- Fail-open startup can hide misconfiguration until a bad profile is selected.

## Guardrails

- Never log resolved secret values; reuse observability redaction for
  secret-like attribute keys.
- Reject configuration values that look like inline credentials (password fields
  with literal values rather than refs).
- Document the provisional schema under `docs/configuration.md` and examples;
  mark unsupported fields explicitly.
- Readiness reflects catalog validity, not live MQ connectivity.

## Alternatives

### Kubernetes Secrets and Vault in the first release

Rejected for CON-001 because it expanded provider scope before the catalog
contract was proven. **CON-002** adds Kubernetes Secrets via the `k8s:` scheme;
Vault remains deferred.

### Strict startup only (no fail-open)

Rejected because multi-profile deployments (read prod + write dev) should remain
available when one profile's secret mount is wrong.

### Inline secrets allowed for local development

Rejected because inline values leak into logs, errors, and copied configs; env
and file refs cover local Kind without weakening the production rule.
