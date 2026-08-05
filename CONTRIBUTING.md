# Contributing to IBM MQ MCP Server

Thank you for helping improve this project.

This project follows the [Code of Conduct](CODE_OF_CONDUCT.md) and is governed
per [GOVERNANCE.md](GOVERNANCE.md).

## Standards map

| Document | Owns |
| --- | --- |
| [AGENTS.md](AGENTS.md) | Agent/contributor workflow and architecture constraints |
| [docs/adr/](docs/adr/) | Locked design decisions |
| [openspec/](openspec/) | Change proposals and delivery intent |
| [agent-context/roadmap.md](agent-context/roadmap.md) | Delivery slices |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community behavior (Contributor Covenant v2.1) |
| [GOVERNANCE.md](GOVERNANCE.md) | Roles, decision making, continuity |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting |
| [SUPPORT.md](SUPPORT.md) | Where to ask for help |

Coding standards, Taskfile gates (`verify` / `lint` / `test` / `coverage` /
`scrub`), and the Go module layout land with **FND-001** / **FND-002**. Until
then, documentation and OSS hygiene changes require link/structure review and
green Docs/CI workflows.

**Merge policy:** rebase-and-merge on every PR (linear history; no squash, no
merge commits). Delete the branch on merge.

## Expectations

- **One logical change per commit** (or per PR). Prefer small, reviewable diffs.
- **Match surrounding style** once Go code exists; do not invent a second style.
- **Tests with behaviour changes** — TDD is mandatory for runtime work
  (`AGENTS.md` / PlatformRelay guidelines).
- **No secrets in git** — credentials belong in env vars, mounted files, or
  secret stores — never in fixtures, examples, logs, or tool results.
- Personal project: no employer ticket keys in subjects. Use English for commit
  messages and user-facing docs.

## Commit messages

```text
:gitmoji: <type>(<optional scope>): <short summary>
```

ASCII gitmoji shortcode mandatory (e.g. `:sparkles:`). Types: `feat` `fix`
`docs` `style` `refactor` `test` `chore` `ci` `build`. No AI co-author trailers.

## Developer Certificate of Origin (DCO)

By contributing, you certify the Developer Certificate of Origin (DCO)
version 1.1. Each commit must include:

```text
Signed-off-by: Your Name <your.email@example.com>
```

Use `git commit -s` to add the trailer. Full DCO text:
https://developercertificate.org/

## Local checks (current)

```bash
# Docs (requires Python deps from docs/requirements-docs.txt)
pip install -r docs/requirements-docs.txt
mkdocs build --strict
```

After FND-001/FND-002, prefer:

```bash
task verify && task lint && task test && task scrub
gitleaks protect --staged --no-banner
```

## Pull requests

- Fill out the PR template.
- Keep the branch rebaseable onto `main`.
- Link the story or ADR when applicable.
- Do not mark work complete until acceptance criteria have evidence.
