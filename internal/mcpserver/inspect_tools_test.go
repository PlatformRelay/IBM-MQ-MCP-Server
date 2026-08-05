package mcpserver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const browseOnlyProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_TOOL_SECRET
    capabilities:
      - browse
`

func TestInspectionToolsRegistered(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	res, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(res.Tools) != 14 {
		t.Fatalf("expected 14 tools, got %d", len(res.Tools))
	}
}

func TestListProfilesToolReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_profiles",
		Arguments: map[string]any{"limit": 10},
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
	if !ok || len(items) == 0 {
		t.Fatalf("items = %#v", payload["items"])
	}
}

func TestQueueManagerStatusDeniedBeforeAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(factory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "queue_manager_status",
		Arguments: map[string]any{"profile": "prod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for policy denial")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny")
	}
}

func TestListChannelsDeniedBeforeAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(factory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_channels",
		Arguments: map[string]any{"profile": "prod", "limit": 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for policy denial")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny")
	}
}

func connectInspectClient(t *testing.T, server *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.0"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	return clientSession
}

func testInspectPool(t *testing.T, fakeClient *fake.Client) *application.ProfilePool {
	t.Helper()
	return testInspectPoolWithDoc(t, `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_TOOL_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`, fakeClient)
}

func testInspectPoolWithDoc(t *testing.T, doc string, fakeClient *fake.Client) *application.ProfilePool {
	t.Helper()
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		if fakeClient != nil {
			fakeClient.Name = profile.Name
			return fakeClient, nil
		}
		return fake.New(profile.Name), nil
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithAdminFactory(factory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}

func TestRegisteredToolSpecsRequireInspect(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, nil)
	mcpserver.NewWithInspector(application.NewInspector(pool))
	for _, name := range mcpserver.RegisteredToolNames() {
		if name == "explain_mq_reason_code" {
			continue
		}
		found := false
		for _, spec := range mcpserver.RegisteredTools {
			if spec.Name == name {
				found = true
				if spec.RequiredCapability != policy.Inspect {
					t.Fatalf("tool %q capability = %q", name, spec.RequiredCapability)
				}
			}
		}
		if !found {
			t.Fatalf("missing spec for %q", name)
		}
	}
}

func TestToolErrorSurfacesDenial(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	err := policy.Authorize(catalog.Profile{Name: "prod"}, policy.Inspect)
	var denial *policy.DenialError
	if !errors.As(err, &denial) {
		t.Fatal("expected denial fixture")
	}
}
