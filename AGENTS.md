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

Until FND-001 lands Taskfile Go gates:

```bash
pip install -r docs/requirements-docs.txt
mkdocs build --strict
```

After FND-001 / FND-002, prefer `task verify`, `task lint`, `task test`,
`task coverage`, and `task scrub` (Kollect/MKurator pattern).

## Repository hygiene

- Keep generated artifacts reproducible and committed only when project
  conventions require it.
- Add ADRs for significant, durable choices.
- Update documentation with behavior changes.
- Never place secrets or real connection credentials in fixtures or examples.
- Follow [CONTRIBUTING.md](CONTRIBUTING.md) (DCO, gitmoji commits) and
  [ADR-0009](docs/adr/0009-license-and-oss-maturity.md) for OSS posture.
