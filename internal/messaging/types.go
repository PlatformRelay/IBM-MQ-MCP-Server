package messaging

const (
	// DefaultBrowseCount is applied when callers omit or pass a non-positive count.
	DefaultBrowseCount = 10
	// MaxBrowseCount caps every browse request.
	MaxBrowseCount = 100

	// DefaultMaxPayloadBytes is applied when callers omit max payload size.
	DefaultMaxPayloadBytes = 4096
	// HardMaxPayloadBytes is the absolute per-message payload cap.
	HardMaxPayloadBytes = 65536
)

// PayloadEncoding describes how a message body is represented in tool output.
type PayloadEncoding string

// Payload encoding labels returned in browse results.
const (
	EncodingUTF8      PayloadEncoding = "utf-8"
	EncodingBase64    PayloadEncoding = "base64"
	EncodingMalformed PayloadEncoding = "malformed"
	EncodingOmitted   PayloadEncoding = "omitted"
)

// BrowseRequest bounds a non-destructive queue browse.
type BrowseRequest struct {
	Count           int
	WaitIntervalMs  int
	IncludePayload  bool
	MaxPayloadBytes int
}

// PutRequest carries validated put parameters for one queue message.
type PutRequest struct {
	ContentType   string
	Payload       string
	CorrelationID string
}

// PutResult identifies a successfully produced message without echoing payload.
type PutResult struct {
	MessageID     string `json:"messageId,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
	Format        string `json:"format,omitempty"`
}

// MessageRecord is one browsed message (metadata with optional payload).
type MessageRecord struct {
	MessageID        string          `json:"messageId,omitempty"`
	CorrelationID    string          `json:"correlationId,omitempty"`
	Format           string          `json:"format,omitempty"`
	PutDate          string          `json:"putDate,omitempty"`
	PutTime          string          `json:"putTime,omitempty"`
	MessageLength    int             `json:"messageLength,omitempty"`
	Priority         int             `json:"priority,omitempty"`
	Persistence      string          `json:"persistence,omitempty"`
	Encoding         PayloadEncoding `json:"encoding,omitempty"`
	Payload          string          `json:"payload,omitempty"`
	PayloadTruncated bool            `json:"payloadTruncated,omitempty"`
}
