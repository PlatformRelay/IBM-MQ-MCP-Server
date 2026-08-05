package audit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
)

func TestSlogRecorderEmitsStructuredAuditEvent(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(logging.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	rec := audit.NewSlogRecorder(logger)

	ctx := audit.WithCorrelationID(context.Background(), "corr-test")
	rec.Record(ctx, audit.Event{
		Kind:      audit.KindOperation,
		Profile:   "lab",
		Operation: "browse_queue_messages",
		Target:    audit.Target{Kind: "queue", Name: "Q1"},
		Outcome:   audit.OutcomeSuccess,
		LatencyMs: 5,
	})

	line := buf.String()
	if !strings.Contains(line, "audit") {
		t.Fatalf("expected audit message, got %q", line)
	}
	if !strings.Contains(line, "corr-test") {
		t.Fatalf("expected correlation id in log: %q", line)
	}
	if !strings.Contains(line, "browse_queue_messages") {
		t.Fatalf("expected operation in log: %q", line)
	}
}

func TestSlogRecorderFailOpenOnSinkError(t *testing.T) {
	t.Parallel()

	rec := audit.NewSlogRecorder(slog.New(failingHandler{}))
	var panicked atomic.Bool
	func() {
		defer func() {
			if recover() != nil {
				panicked.Store(true)
			}
		}()
		rec.Record(context.Background(), audit.Event{
			Kind:      audit.KindOperation,
			Profile:   "lab",
			Operation: "consume_queue_messages",
			Outcome:   audit.OutcomeSuccess,
		})
	}()
	if panicked.Load() {
		t.Fatal("audit recorder must fail open and not panic")
	}
}

func TestEnsureCorrelationIDGeneratesAndPreserves(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx1 := audit.EnsureCorrelationID(ctx)
	id1 := audit.CorrelationIDFrom(ctx1)
	if id1 == "" {
		t.Fatal("expected generated correlation id")
	}
	ctx2 := audit.EnsureCorrelationID(ctx1)
	if audit.CorrelationIDFrom(ctx2) != id1 {
		t.Fatalf("correlation id changed: %q -> %q", id1, audit.CorrelationIDFrom(ctx2))
	}
}

func TestRecordOperationUsesContextCorrelation(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(logging.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	rec := audit.NewSlogRecorder(logger)

	ctx := audit.WithCorrelationID(context.Background(), "join-me")
	audit.RecordOperation(ctx, rec, audit.OperationRecord{
		Profile:   "lab",
		Operation: "put_queue_message",
		Target:    audit.Target{Kind: "queue", Name: "Q1"},
		Outcome:   audit.OutcomeSuccess,
		LatencyMs: 1,
	})

	if !strings.Contains(buf.String(), "join-me") {
		t.Fatalf("operation audit missing correlation id: %s", buf.String())
	}
}

type failingHandler struct{}

func (failingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (failingHandler) Handle(context.Context, slog.Record) error {
	return errors.New("sink unavailable")
}

func (h failingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h failingHandler) WithGroup(string) slog.Handler { return h }

func TestMQSCCommandRedactionInEvent(t *testing.T) {
	t.Parallel()

	event := audit.Event{
		Kind:            audit.KindOperation,
		Operation:       "execute_mqsc",
		CommandRedacted: "DISPLAY QLOCAL('DEV.QUEUE') PASSWORD(***REDACTED***)",
		Outcome:         audit.OutcomeSuccess,
	}
	raw := attrsToJSON(t, event.Attrs())
	if strings.Contains(strings.ToLower(raw), "actualpassword") {
		t.Fatalf("command field must stay redacted: %s", raw)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatal(err)
	}
	if _, ok := parsed["commandRedacted"]; !ok {
		t.Fatalf("expected commandRedacted field: %s", raw)
	}
}
