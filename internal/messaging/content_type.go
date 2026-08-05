package messaging

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// ContentTypeTextPlain is plain UTF-8 text for put operations (DQ 17).
	ContentTypeTextPlain = "text/plain"
	// ContentTypeJSON is JSON text validated before put (DQ 17).
	ContentTypeJSON = "application/json"
	// ContentTypeOctetStream accepts base64-encoded input decoded before put (DQ 17).
	ContentTypeOctetStream = "application/octet-stream"
)

// ContentTypeError is returned when put content type or payload shape is invalid.
type ContentTypeError struct {
	ContentType string
	Reason      string
}

func (e *ContentTypeError) Error() string {
	return fmt.Sprintf("unsupported or invalid content type %q: %s", e.ContentType, e.Reason)
}

// PayloadSizeError is returned when decoded payload exceeds HardMaxPayloadBytes.
type PayloadSizeError struct {
	Size    int
	MaxSize int
}

func (e *PayloadSizeError) Error() string {
	return fmt.Sprintf("payload size %d exceeds maximum %d bytes", e.Size, e.MaxSize)
}

// PreparePutPayload validates content type and payload shape, returning decoded bytes
// ready for mqweb PUT. Size is enforced before any network call.
func PreparePutPayload(contentType, payload string) (body []byte, normalizedType string, err error) {
	normalized := normalizeContentType(contentType)
	switch normalized {
	case ContentTypeTextPlain:
		body = []byte(payload)
		if !isValidUTF8(body) {
			return nil, "", &ContentTypeError{
				ContentType: normalized,
				Reason:      "text/plain payload must be valid UTF-8",
			}
		}
	case ContentTypeJSON:
		if !json.Valid([]byte(payload)) {
			return nil, "", &ContentTypeError{
				ContentType: normalized,
				Reason:      "application/json payload must parse as JSON",
			}
		}
		body = []byte(payload)
	case ContentTypeOctetStream:
		decoded, decodeErr := base64.StdEncoding.DecodeString(payload)
		if decodeErr != nil {
			return nil, "", &ContentTypeError{
				ContentType: normalized,
				Reason:      "application/octet-stream payload must be standard base64",
			}
		}
		body = decoded
	default:
		return nil, "", &ContentTypeError{
			ContentType: contentType,
			Reason:      "accepted types are text/plain, application/json, and application/octet-stream",
		}
	}
	if sizeErr := ValidatePutPayloadSize(len(body)); sizeErr != nil {
		return nil, "", sizeErr
	}
	return body, normalized, nil
}

func normalizeContentType(contentType string) string {
	ct := strings.TrimSpace(strings.ToLower(contentType))
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = strings.TrimSpace(ct[:idx])
	}
	return ct
}

// ValidatePutPayloadSize rejects decoded payloads over HardMaxPayloadBytes.
func ValidatePutPayloadSize(size int) error {
	if size > HardMaxPayloadBytes {
		return &PayloadSizeError{Size: size, MaxSize: HardMaxPayloadBytes}
	}
	return nil
}
