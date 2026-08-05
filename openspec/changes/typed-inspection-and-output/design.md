# Inspection and output design

**Status:** Active — ADR-0005 accepted; OUT-001 landed compact text fallback.

Defines the shared collection contract (filter, limit, cursor, truncation
metadata) once and reuses it across every tool. Benchmark fixtures and results
live in [output-benchmarks.md](../../docs/development/output-benchmarks.md).

Field selection (`fields[]`) is explicitly deferred (`collection.FieldSelectionDeferred`,
marker **OUT-001-DEFERRED**) until a schema-safe projection design is agreed.
