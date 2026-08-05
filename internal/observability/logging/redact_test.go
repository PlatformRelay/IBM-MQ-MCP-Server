package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
)

func TestRedactsSecretLikeFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := logging.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

	logger.Info("tool call",
		slog.String("password", "hunter2"),
		slog.String("token", "abc123"),
		slog.String("queue", "DEV.QUEUE"),
	)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log: %v", err)
	}
	attrs, ok := entry["password"].(string)
	if !ok || attrs != logging.RedactedValue {
		t.Fatalf("password = %#v, want %q", entry["password"], logging.RedactedValue)
	}
	if entry["token"] != logging.RedactedValue {
		t.Fatalf("token = %#v, want redacted", entry["token"])
	}
	if entry["queue"] != "DEV.QUEUE" {
		t.Fatalf("queue = %#v, want unchanged", entry["queue"])
	}
}

func TestSanitizesLogInjectionFromToolArguments(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := logging.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

	injected := "safe\ninjected\rfield"
	logger.Info("tool call", slog.String("argument", injected))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log: %v", err)
	}
	arg, ok := entry["argument"].(string)
	if !ok {
		t.Fatalf("argument type = %T", entry["argument"])
	}
	if strings.Contains(arg, "\n") || strings.Contains(arg, "\r") {
		t.Fatalf("argument still contains control chars: %q", arg)
	}
	if !strings.Contains(arg, "safe") || !strings.Contains(arg, "injected") {
		t.Fatalf("argument = %q, expected sanitized but readable", arg)
	}
}

func TestRedactingHandlerPassesThroughContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	handler := logging.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)

	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "present")
	logger.InfoContext(ctx, "ping")

	if !strings.Contains(buf.String(), "ping") {
		t.Fatalf("log output missing message: %s", buf.String())
	}
}
