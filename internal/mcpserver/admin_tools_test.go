package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
)

const adminToolProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_TOOL_SECRET
    capabilities:
      - administer
`

func TestAdminToolsRegistered(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := adminToolPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) != 20 {
		t.Fatalf("expected 20 tools, got %d", len(res.Tools))
	}
}

func TestDefineQueueDeniedBeforeAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), nil,
		application.WithAdminFactory(func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			return fakeClient, nil
		}),
	)
	t.Cleanup(func() { _ = pool.Close() })
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolDefineQueue,
		Arguments: map[string]any{
			"profile":   "prod",
			"queue":     "NEW.Q",
			"queueType": "LOCAL",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny: %d", fakeClient.TotalCalls())
	}
}

func TestDefineQueueReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	pool := adminToolPool(t, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "define_queue",
		Arguments: map[string]any{
			"profile":   "prod",
			"queue":     "NEW.Q",
			"queueType": "LOCAL",
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
	if payload["operation"] != string(mqadmin.MutationDefine) {
		t.Fatalf("operation = %#v", payload["operation"])
	}
	if fakeClient.DefineQueueCalls != 1 {
		t.Fatalf("define calls = %d", fakeClient.DefineQueueCalls)
	}
}

const toolDefineQueue = "define_queue"

func adminToolPool(t *testing.T, fakeClient *fake.Client) *application.ProfilePool {
	t.Helper()
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(adminToolProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := application.NewProfilePool(cat, cat.Validate(), secrets.NewResolver(), nil,
		application.WithAdminFactory(func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
			return fakeClient, nil
		}),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}
