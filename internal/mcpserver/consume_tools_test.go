package mcpserver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const consumeToolProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_CONSUME_TOOL_SECRET
    capabilities:
      - consume
`

const browseOnlyConsumeToolDenyDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_CONSUME_TOOL_SECRET
    capabilities:
      - browse
`

func TestConsumeQueueMessagesDeniedBeforeMessaging(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	t.Setenv("IBM_MQ_MCP_CONSUME_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyConsumeToolDenyDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testBrowsePool(t, cat, nil, fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "consume_queue_messages",
		Arguments: map[string]any{"profile": "prod", "queue": "Q1", "count": 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for policy denial")
	}
	if fakeMsg.TotalCalls() != 0 {
		t.Fatalf("messaging invoked on deny, calls=%d", fakeMsg.TotalCalls())
	}
}

func TestConsumeQueueMessagesReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	fakeMsg.ConsumePage.Items = []messaging.MessageRecord{
		{MessageID: "ID:1", Encoding: messaging.EncodingOmitted},
	}
	t.Setenv("IBM_MQ_MCP_CONSUME_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(consumeToolProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testBrowsePool(t, cat, nil, fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "consume_queue_messages",
		Arguments: map[string]any{"profile": "prod", "queue": "Q1", "count": 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool error: %+v", res.Content)
	}
	payload, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structuredContent type = %T", res.StructuredContent)
	}
	items, ok := payload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v", payload["items"])
	}
	if fakeMsg.BrowseOnlyCalls() != 0 {
		t.Fatal("browse path invoked during consume")
	}
	if fakeMsg.ConsumeOnlyCalls() != 1 {
		t.Fatalf("consume calls = %d", fakeMsg.ConsumeOnlyCalls())
	}
}

func TestConsumeQueueMessagesSurfacesPartialResultsOnMidBatchFailure(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	partialPage := collection.Page[messaging.MessageRecord]{
		Items:            []messaging.MessageRecord{{MessageID: "ID:partial"}},
		Truncated:        true,
		TruncationReason: collection.TruncationMidBatchFailure,
	}
	fakeMsg.ConsumePage = partialPage
	fakeMsg.ConsumeErr = messaging.NewPartialConsumeError(partialPage, errors.New("status 500"))
	t.Setenv("IBM_MQ_MCP_CONSUME_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(consumeToolProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testBrowsePool(t, cat, nil, fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "consume_queue_messages",
		Arguments: map[string]any{"profile": "prod", "queue": "Q1", "count": 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for partial consume failure")
	}
	payload, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structuredContent type = %T", res.StructuredContent)
	}
	items, ok := payload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v", payload["items"])
	}
	if payload["truncated"] != true {
		t.Fatalf("truncated = %#v", payload["truncated"])
	}
}

func TestConsumeToolRequiresConsumeCapability(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, nil)
	mcpserver.NewWithInspector(application.NewInspector(pool))
	found := false
	for _, spec := range mcpserver.RegisteredTools {
		if spec.Name == "consume_queue_messages" {
			found = true
			if spec.RequiredCapability != policy.Consume {
				t.Fatalf("capability = %q", spec.RequiredCapability)
			}
		}
	}
	if !found {
		t.Fatal("consume_queue_messages not registered")
	}
}
