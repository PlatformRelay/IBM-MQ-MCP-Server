# mqweb Messaging REST semantics (MSG-001 spike)

Authority: [ADR-0002](../adr/0002-mqweb-first-connectivity.md),
[MSG-001](https://github.com/PlatformRelay/IBM-MQ-MCP-Server/blob/main/agent-context/stories/MSG-001.md).

This document records **contract assumptions** for the first-release browse tool.
Live certification against FND-004 remains required before production support
claims ([version matrix](../support/version-matrix.md)).

## HTTP verbs and destructiveness

| Operation | HTTP | Resource | Destructive? |
| --- | --- | --- | --- |
| Browse one message | `GET` | `/ibmmq/rest/v3/messaging/qmgr/{qmgr}/queue/{queue}/message` | **No** — message remains on the queue |
| Browse message list (metadata) | `GET` | `/ibmmq/rest/v3/messaging/qmgr/{qmgr}/queue/{queue}/messagelist` | **No** — summary JSON only |
| Consume / destructive get | `DELETE` | `/ibmmq/rest/v3/messaging/qmgr/{qmgr}/queue/{queue}/message` | **Yes** — one message removed per call |
| Produce | `POST` | `/ibmmq/rest/v3/messaging/qmgr/{qmgr}/queue/{queue}/message` | N/A (write path — MSG-002) |

IBM MQ documentation and the public OpenAPI description agree: **only `DELETE`
removes messages.** The MCP browse tool uses **`GET` on `/message` and
`/messagelist` only** — never `DELETE`. The MCP consume tool uses **`DELETE`
on `/message` only** — never `GET`.

## Browse proof strategy

1. **Unit/contract tests** — HTTP spy asserts zero `DELETE` requests on the
   browse code path (`internal/adapter/mqweb/messaging_test.go`).
2. **Fake messaging client** — application and MCP tests record browse calls
   only; consume uses a separate port method (`ConsumeMessages`).
3. **Live e2e (opt-in)** — when `IBM_MQ_MCP_E2E=1`, compare queue depth
   before and after browse via Admin REST `get queue` (browse must not decrease
   depth); consume e2e proves depth decreases after `DELETE`.

## Payload formats

- mqweb browse commonly returns **text** (`MQSTR` / JMS TextMessage) in the
  response body for `GET /message`.
- **Binary or incompatible formats** may be skipped or returned opaquely by
  mqweb; this server maps non–UTF-8 bytes to **base64** with an `encoding`
  field and replaces undecodable content deterministically.
- Default tool behaviour is **metadata only**; payloads require
  `include_payload=true` and the `browse` capability grant.

## Limits (design questions 14–16)

| Parameter | Default | Hard max |
| --- | --- | --- |
| Browse count | 10 | 100 |
| Consume count | 10 | 100 |
| Consume wait interval (ms) | 0 (no wait) | 30000 (30s) |
| Max payload bytes per message | 4096 | 65536 |

Server-side enforcement applies before results are serialized. Secret-like
patterns in payload text are redacted before return; payloads are never logged.

## Consume semantics (MSG-003)

- mqweb performs **one destructive get per HTTP `DELETE`** on `/message`; the
  MCP tool loops up to the requested count, stopping early on an empty queue
  (`204 No Content`).
- The **`wait`** query parameter (milliseconds) applies to the **first** delete
  only; subsequent deletes in the same tool call use no wait.
- Message metadata comes from response headers (`ibm-mq-md-messageid`, etc.);
  bodies are optional via `includePayload`.
- **No syncpoint, transaction, or exactly-once delivery** — if the HTTP
  connection fails after mqweb removed a message but before the client reads
  the response, that message may be lost. This server does not claim stronger
  semantics than mqweb documents.
- Incompatible message formats may remain on the queue while mqweb returns an
  error for that delete attempt.
- **Mid-batch failure:** when a later `DELETE` fails after earlier messages
  were removed, the tool returns **`IsError` with `structuredContent`** listing
  messages already consumed. The page sets `truncated: true` and
  `truncationReason: "mid_batch_failure"`. Callers must treat these messages as
  gone from the queue even though the overall tool call failed.

## Put semantics (MSG-002, design question 17)

| Content type | Input | mqweb body |
| --- | --- | --- |
| `text/plain` | UTF-8 text | Raw text with `text/plain;charset=utf-8` |
| `application/json` | JSON text (validated) | Raw JSON with `application/json;charset=utf-8` |
| `application/octet-stream` | Standard base64 | Decoded bytes with `application/octet-stream` |

- HTTP `POST` on `/message` with `ibm-mq-rest-csrf-token` (required by mqweb).
- Response header `ibm-mq-md-messageId` supplies the allocated message ID.
- Optional request header `ibm-mq-md-correlationId` when caller supplies one.
- Decoded payload size capped at **65536** bytes before any network call.
- Tool results return identifiers only — never echo the put payload.

## Query parameters

Browse list calls pass (when supported by target mqweb):

- `numberOfMessages` — bounded by server max (100)
- `waitInterval` — caller wait in milliseconds (0 = no wait)

Consume delete calls pass:

- `wait` — milliseconds to wait for the first message only (0–30000)

Optional filters (`messageId`, `correlationId`) are reserved for later stories;
MSG-001 browses the next available messages only; MSG-003 consumes the next
available messages only.

## References

- [IBM MQ GET /message](https://www.ibm.com/docs/en/ibm-mq/9.3.x?topic=messagingqmgrqmgrnamequeuequeuenamemessage-get)
- [IBM MQ GET /messagelist](https://www.ibm.com/docs/en/ibm-mq/9.3.x?topic=messagingqmgrqmgrnamequeuequeuenamemessagelist-get)
- [Get started with the IBM MQ messaging REST API](https://developer.ibm.com/tutorials/mq-develop-mq-rest-api/)
