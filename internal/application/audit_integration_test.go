package application_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/logging"
)

func TestPolicyGateAuditEmitsJoinedCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(logging.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	auditRec := audit.NewSlogRecorder(logger)

	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_AUDIT_SECRET
    capabilities:
      - inspect
`
	t.Setenv("IBM_MQ_MCP_AUDIT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	gate := application.NewPolicyGate(application.WithPolicyAuditRecorder(auditRec))
	pool := application.NewProfilePool(
		cat, cat.Validate(), secrets.NewResolver(), gate,
		application.WithAuditRecorder(auditRec),
		application.WithMessagingFactory(func(profile catalog.Profile, _ *secrets.Resolver) (messaging.Client, error) {
			return msgfake.New(profile.Name), nil
		}),
	)
	t.Cleanup(func() { _ = pool.Close() })

	ctx := audit.WithCorrelationID(context.Background(), "corr-policy-join")
	browser := application.NewBrowser(pool)
	_, err = browser.BrowseQueueMessages(ctx, "prod", "Q1", messaging.BrowseRequest{Count: 5})
	if err == nil {
		t.Fatal("expected policy denial")
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected policy and operation audit lines, got %q", buf.String())
	}
	correlationIDs := make(map[string]struct{})
	for _, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatal(err)
		}
		id, _ := entry["correlationId"].(string)
		if id == "" {
			t.Fatalf("missing correlationId in audit line: %s", line)
		}
		correlationIDs[id] = struct{}{}
	}
	if len(correlationIDs) != 1 {
		t.Fatalf("correlation ids not joined: %v", correlationIDs)
	}
	if _, ok := correlationIDs["corr-policy-join"]; !ok {
		t.Fatalf("expected corr-policy-join, got %v", correlationIDs)
	}
}
