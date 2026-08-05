# Output rendering benchmarks (OUT-001)

Recorded **2026-08-05** on representative MQ-shaped fixtures in
`internal/output/fixtures.go`. Byte counts proxy token cost (1 byte ≈ 1 token
for ASCII payloads in typical LLM tokenizers).

## Method

| Format | Description |
| --- | --- |
| **Compact text** | Deterministic `internal/output` renderers wired into MCP `content` text blocks |
| **Minified JSON** | Canonical `structuredContent` (`json.Marshal`, no whitespace) |
| **Pretty JSON** | Indented JSON (legacy human/debug view) |
| **Markdown table** | Pipe table derived from the same fixture (benchmark only; not shipped) |
| **TOON** | Not measured — no Go dependency adopted; evidence-gated per ADR-0005 |

Fixtures: 10-queue list (truncated page), 3-message browse page, queue manager
status, connectivity report, reason-code explanation.

## Results

### Queue list (10 items, truncated)

| Format | Bytes | vs minified JSON |
| --- | ---: | ---: |
| Compact text | 355 | −28% |
| Minified JSON | 492 | — |
| Pretty JSON | 776 | +58% |
| Markdown table | 318 | −35% |

Compact text beats minified JSON by dropping redundant JSON syntax and field
names on repeated rows. Pretty JSON is substantially larger. Markdown wins on
raw bytes for this fixture but lacks typed schema alignment.

### Message list (3 items)

| Format | Bytes | vs minified JSON |
| --- | ---: | ---: |
| Compact text | 256 | −59% |
| Minified JSON | 623 | — |
| Pretty JSON | 885 | +42% |
| Markdown table | 144 | −77% |

Browse/consume compact text omits payload bodies (metadata only), substantially
beating minified JSON while preserving identifiers and encoding labels.

### Singleton results

| Fixture | Compact text (bytes) |
| --- | ---: |
| Queue manager status | 124 |
| Connectivity report | 156 |
| Reason explanation (2035) | 241 |

## Decisions

1. **JSON `structuredContent` stays canonical** — no alternate default adopted.
2. **Compact text fallback ships** for all 17 public tools — replaces SDK raw JSON
   echo in `content` blocks (ADR-0005 supplementary rendering).
3. **Markdown / TOON remain evidence-gated** — Markdown measured here for
   comparison only; TOON deferred pending dependency decision
   ([reference assessment](../architecture/reference-assessment.md#toon)).
4. **Field selection deferred** — `collection.FieldSelectionDeferred` marks
   `fields[]` per-item projection as **OUT-001-DEFERRED**; pagination, filters,
   and truncation metadata are implemented.

## Reproduce

```bash
go test ./internal/output/... -count=1
go run ./hack/output-bench-sizes.go
```

Regenerate this table when fixtures or renderers change.
