package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const produceToolProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_PUT_TOOL_SECRET
    capabilities:
      - produce
`

const browseOnlyPutDenyDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_PUT_TOOL_SECRET
    capabilities:
      - browse
`

func TestPutQueueMessageDeniedBeforeMessaging(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	t.Setenv("IBM_MQ_MCP_PUT_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyPutDenyDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testProducePool(t, cat, fake.New("prod"), fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "put_queue_message",
		Arguments: map[string]any{
			"profile":     "prod",
			"queue":       "Q1",
			"contentType": messaging.ContentTypeTextPlain,
			"payload":     "hello",
		},
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

func TestPutQueueMessageReturnsIdentifiersOnly(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	fakeMsg.PutResult = messaging.PutResult{
		MessageID:     "ID:abc",
		CorrelationID: "CID:1",
		Format:        "MQSTR",
	}
	t.Setenv("IBM_MQ_MCP_PUT_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(produceToolProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testProducePool(t, cat, nil, fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "put_queue_message",
		Arguments: map[string]any{
			"profile":       "prod",
			"queue":         "Q1",
			"contentType":   messaging.ContentTypeTextPlain,
			"payload":       "hello",
			"correlationId": "CID:1",
		},
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
	if payload["messageId"] != "ID:abc" {
		t.Fatalf("messageId = %#v", payload["messageId"])
	}
	if _, hasPayload := payload["payload"]; hasPayload {
		t.Fatal("result must not echo payload")
	}
	if fakeMsg.PutCalls != 1 {
		t.Fatalf("put calls = %d", fakeMsg.PutCalls)
	}
}

func TestPutToolRequiresProduceCapability(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	mcpserver.NewWithInspector(application.NewInspector(pool))
	found := false
	for _, spec := range mcpserver.RegisteredTools {
		if spec.Name == "put_queue_message" {
			found = true
			if spec.RequiredCapability != policy.Produce {
				t.Fatalf("capability = %q", spec.RequiredCapability)
			}
		}
	}
	if !found {
		t.Fatal("put_queue_message not registered")
	}
}

func testProducePool(
	t *testing.T,
	cat *catalog.Catalog,
	fakeAdmin *fake.Client,
	fakeMsg *msgfake.Client,
) *application.ProfilePool {
	t.Helper()
	return testBrowsePool(t, cat, fakeAdmin, fakeMsg)
}
