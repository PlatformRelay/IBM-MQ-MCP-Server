# CI quality gates

Authority: [FND-002](../../agent-context/stories/FND-002.md).

## Local one-command gate

```bash
task check
```

This runs `verify` → `lint` → `test:race` → `coverage` → `vulncheck` →
`scrub:tree` → `build`.

## CI jobs (must stay present)

The workflow `.github/workflows/ci.yaml` defines:

| Job | Purpose |
| --- | --- |
| `gitleaks` | Secret scanning |
| `oss-hygiene` | Required community/tooling files |
| `format` | `gofmt` drift |
| `lint` | golangci-lint (govet-class + static analysis) |
| `test` | `-race` unit tests, coverage floor, CGO-free build |
| `vulncheck` | `govulncheck` |
| `scrub` | Forbidden-pattern scrub (`task scrub:tree`) |

CodeQL runs in `.github/workflows/codeql.yaml`. Dependabot and Renovate are
configured at the repository root.

## Evidence that a missing gate fails

`internal/architecture/ci_gates_test.go` asserts the required CI job names and
the CodeQL workflow file exist. Deleting or renaming a required job fails
`go test ./internal/architecture/` — that failure is the recorded proof that an
incomplete gate set cannot pass the test suite (and therefore cannot pass CI
once `test` is required).

Branch protection should mark `CI / format`, `CI / lint`, `CI / test`,
`CI / vulncheck`, `CI / gitleaks`, and `CodeQL` as required checks when the
repository enables them.
