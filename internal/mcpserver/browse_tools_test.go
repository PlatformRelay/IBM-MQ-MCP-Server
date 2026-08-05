package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const browseToolProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_BROWSE_TOOL_SECRET
    capabilities:
      - browse
`

const inspectOnlyBrowseDenyDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_BROWSE_TOOL_SECRET
    capabilities:
      - inspect
`

func TestBrowseQueueMessagesDeniedBeforeMessaging(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	t.Setenv("IBM_MQ_MCP_BROWSE_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(inspectOnlyBrowseDenyDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testBrowsePool(t, cat, fake.New("prod"), fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "browse_queue_messages",
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

func TestBrowseQueueMessagesReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeMsg := msgfake.New("prod")
	fakeMsg.BrowsePage.Items = []messaging.MessageRecord{
		{MessageID: "ID:1", Encoding: messaging.EncodingOmitted},
	}
	t.Setenv("IBM_MQ_MCP_BROWSE_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseToolProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := testBrowsePool(t, cat, nil, fakeMsg)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "browse_queue_messages",
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
	if fakeMsg.ConsumeOnlyCalls() != 0 {
		t.Fatal("destructive consume path invoked")
	}
}

func TestBrowseToolRequiresBrowseCapability(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	mcpserver.NewWithInspector(application.NewInspector(pool))
	found := false
	for _, spec := range mcpserver.RegisteredTools {
		if spec.Name == "browse_queue_messages" {
			found = true
			if spec.RequiredCapability != policy.Browse {
				t.Fatalf("capability = %q", spec.RequiredCapability)
			}
		}
	}
	if !found {
		t.Fatal("browse_queue_messages not registered")
	}
}

func testBrowsePool(
	t *testing.T,
	cat *catalog.Catalog,
	fakeAdmin *fake.Client,
	fakeMsg *msgfake.Client,
) *application.ProfilePool {
	t.Helper()
	var adminFactory application.AdminClientFactory
	if fakeAdmin != nil {
		adminFactory = func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			fakeAdmin.Name = profile.Name
			return fakeAdmin, nil
		}
	} else {
		adminFactory = func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			return fake.New(profile.Name), nil
		}
	}
	msgFactory := func(profile catalog.Profile, _ *secrets.Resolver) (messaging.Client, error) {
		if fakeMsg != nil {
			fakeMsg.Name = profile.Name
			return fakeMsg, nil
		}
		return msgfake.New(profile.Name), nil
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(adminFactory),
		application.WithMessagingFactory(msgFactory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}
