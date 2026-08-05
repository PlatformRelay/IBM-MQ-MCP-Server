package messaging_test

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

func TestPreparePutPayloadTextPlain(t *testing.T) {
	body, ct, err := messaging.PreparePutPayload(messaging.ContentTypeTextPlain, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if ct != messaging.ContentTypeTextPlain {
		t.Fatalf("contentType = %q", ct)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
}

func TestPreparePutPayloadJSONRequiresValidJSON(t *testing.T) {
	_, _, err := messaging.PreparePutPayload(messaging.ContentTypeJSON, "{not json")
	if err == nil {
		t.Fatal("expected error")
	}
	var ctErr *messaging.ContentTypeError
	if !errors.As(err, &ctErr) {
		t.Fatalf("expected ContentTypeError, got %T", err)
	}
}

func TestPreparePutPayloadJSONAcceptsObject(t *testing.T) {
	body, ct, err := messaging.PreparePutPayload(messaging.ContentTypeJSON, `{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if ct != messaging.ContentTypeJSON {
		t.Fatalf("contentType = %q", ct)
	}
	if string(body) != `{"a":1}` {
		t.Fatalf("body = %q", body)
	}
}

func TestPreparePutPayloadOctetStreamDecodesBase64(t *testing.T) {
	raw := []byte{0x00, 0xff, 0x42}
	encoded := base64.StdEncoding.EncodeToString(raw)
	body, ct, err := messaging.PreparePutPayload(messaging.ContentTypeOctetStream, encoded)
	if err != nil {
		t.Fatal(err)
	}
	if ct != messaging.ContentTypeOctetStream {
		t.Fatalf("contentType = %q", ct)
	}
	if string(body) != string(raw) {
		t.Fatalf("body = %v want %v", body, raw)
	}
}

func TestPreparePutPayloadRejectsUnknownContentType(t *testing.T) {
	_, _, err := messaging.PreparePutPayload("image/png", "data")
	if err == nil {
		t.Fatal("expected error")
	}
	var ctErr *messaging.ContentTypeError
	if !errors.As(err, &ctErr) {
		t.Fatalf("expected ContentTypeError, got %T", err)
	}
}

func TestPreparePutPayloadRejectsOverHardMax(t *testing.T) {
	payload := strings.Repeat("x", messaging.HardMaxPayloadBytes+1)
	_, _, err := messaging.PreparePutPayload(messaging.ContentTypeTextPlain, payload)
	if err == nil {
		t.Fatal("expected error")
	}
	var sizeErr *messaging.PayloadSizeError
	if !errors.As(err, &sizeErr) {
		t.Fatalf("expected PayloadSizeError, got %T: %v", err, err)
	}
}

func TestPreparePutPayloadOctetStreamRejectsInvalidBase64(t *testing.T) {
	_, _, err := messaging.PreparePutPayload(messaging.ContentTypeOctetStream, "not!!!base64")
	if err == nil {
		t.Fatal("expected error")
	}
	var ctErr *messaging.ContentTypeError
	if !errors.As(err, &ctErr) {
		t.Fatalf("expected ContentTypeError, got %T", err)
	}
}
