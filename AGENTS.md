# Agent guide

Read this file before changing the repository.

## Project state

Product intent for the bootstrap proposal is **Accepted**. Runtime work starts
at [FND-001](agent-context/stories/FND-001.md). Do not implement MQ tools or
adapters until their stories and blocking ADRs are resolved.

## Sources of truth

1. `openspec/changes/bootstrap-mq-mcp/proposal.md` defines the approved product
   intent and boundaries.
2. `agent-context/roadmap.md` tracks delivery slices.
3. `agent-context/stories/*.md` holds acceptance criteria and session handoffs.
4. `docs/adr/` records approved architectural decisions.
5. `docs/` explains the resulting system and its operation.

Do not duplicate requirements across these layers. Link to the authority instead.

## Workflow

1. Select one open roadmap story.
2. Read the story and linked specification or proposal.
3. Resolve outcome-changing questions and obtain approval (or decide-and-log
   under an authorized autonomous loop).
4. Mark the story in flight and add a session log entry.
5. Derive a technical plan mapped to acceptance scenarios.
6. Implement the smallest reviewable slice.
7. Run the commands documented by that story and record evidence.
8. Mark the story done only when every acceptance criterion has evidence.

## Architecture constraints

- Separate MCP protocol, application policy, IBM MQ domain, and transport
  adapters.
- Treat every queue-manager profile as an independent trust boundary.
- Deny capabilities not explicitly granted.
- Keep credentials out of configuration values, logs, errors, and tool results.
- Prefer typed operations over arbitrary MQSC.
- Preserve an adapter seam for mqweb REST and future PCF/native connectivity.
- Do not make Kubernetes or MKurator a runtime prerequisite.

## Validation

Documented once here; CI mirrors these via `.github/workflows/ci.yaml`.
Details: [docs/development/ci-gates.md](docs/development/ci-gates.md).

```bash
# Docs
pip install -r docs/requirements-docs.txt
mkdocs build --strict

# One-command local CI equivalent (FND-002)
task check

# Individual gates
task verify        # go mod tidy + gofmt
task lint          # golangci-lint
task test:race     # go test -race
task coverage      # coverage floor (mcpserver)
task vulncheck     # govulncheck
task scrub:tree    # forbidden-pattern scrub
task build         # CGO-free binary
task docker:build  # local container smoke (ibm-mq-mcp:dev, distroless nonroot)
task run           # MCP server over stdio (default; no ops HTTP)

# Optional ops HTTP (health, readiness, metrics) — separate from MCP transport
IBM_MQ_MCP_OPS_ADDR=:9090 task run
# or: go run ./cmd/ibm-mq-mcp --ops-addr :9090
```

### Operational endpoints (OBS-001)

When `--ops-addr` or `IBM_MQ_MCP_OPS_ADDR` is set, a dedicated HTTP listener
serves probes and metrics **separately from stdio MCP**. In stdio-only mode
(default) these endpoints are absent.

| Path | Purpose |
| --- | --- |
| `/healthz` | Liveness — process can serve probes |
| `/readyz` | Readiness — valid config and MCP transport serving (no MQ contact) |
| `/metrics` | Prometheus: `ibm_mq_mcp_requests_total`, `ibm_mq_mcp_request_duration_seconds`, `ibm_mq_mcp_policy_denials_total` (profile label only; `_none` until profiles land) |

Structured logs use JSON on stderr with central redaction of secret-like fields
and sanitization of tool-argument values to resist log injection.

## Repository hygiene

- Keep generated artifacts reproducible and committed only when project
  conventions require it.
- Add ADRs for significant, durable choices.
- Update documentation with behavior changes.
- Never place secrets or real connection credentials in fixtures or examples.
- Follow [CONTRIBUTING.md](CONTRIBUTING.md) (DCO, gitmoji commits) and
  [ADR-0009](docs/adr/0009-license-and-oss-maturity.md) for OSS posture.
