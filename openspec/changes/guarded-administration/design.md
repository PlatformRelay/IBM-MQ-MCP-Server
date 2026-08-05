# Administration design

**Status:** Active — ADR-0008 accepted; ADM-003 implements exceptional raw MQSC.

Must enumerate the exact supported object types and verbs, the dry-run
truthfulness rule per operation, and reference the INT-001 pre-mutation hook
contract ([ADR-0007](../../docs/adr/0007-mkurator-coexistence.md)).

Raw MQSC policy: [ADR-0008](../../docs/adr/0008-raw-mqsc-policy.md). The
`execute_mqsc` tool is server-opt-in (`--enable-mqsc`) and profile-gated; v0
allows read-only verbs only (`DISPLAY`, `DIS`, `PING`).
