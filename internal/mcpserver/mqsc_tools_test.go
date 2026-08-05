package mcpserver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
)

const mqscToolProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_TOOL_SECRET
    capabilities:
      - execute_mqsc
`

func TestExecuteMQSCToolUndiscoverableByDefault(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := mqscToolPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if tool.Name == toolExecuteMQSC {
			t.Fatalf("execute_mqsc must not be registered by default")
		}
	}
	if len(res.Tools) != 29 {
		t.Fatalf("expected 29 tools, got %d", len(res.Tools))
	}
}

func TestExecuteMQSCToolRegisteredWhenServerOptIn(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := mqscToolPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) != 30 {
		t.Fatalf("expected 30 tools, got %d", len(res.Tools))
	}
	found := false
	for _, tool := range res.Tools {
		if tool.Name == toolExecuteMQSC {
			found = true
		}
	}
	if !found {
		t.Fatal("expected execute_mqsc tool when server opt-in is enabled")
	}
}

func TestExecuteMQSCDeniedWithoutProfileCapability(t *testing.T) {
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
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "DISPLAY QLOCAL('DEV.QUEUE.1')",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error without execute_mqsc capability")
	}
	if fakeClient.ExecuteRawMQSCCalls != 0 {
		t.Fatalf("adapter invoked on deny: %d", fakeClient.ExecuteRawMQSCCalls)
	}
}

func TestExecuteMQSCDeniedVerbNeverHitsAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	pool := mqscToolPool(t, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "ALTER QLOCAL('DEV.QUEUE.1') MAXDEPTH(1000)",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for denied verb")
	}
	if fakeClient.ExecuteRawMQSCCalls != 0 {
		t.Fatalf("adapter invoked on verb deny: %d", fakeClient.ExecuteRawMQSCCalls)
	}
}

func TestExecuteMQSCCallToolFailsWithoutServerOptIn(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := mqscToolPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	_, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "DISPLAY QLOCAL('DEV.QUEUE.1')",
		},
	})
	if err == nil {
		t.Fatal("expected CallTool error when execute_mqsc is not registered")
	}
}

func TestExecuteMQSCNewlineInjectionNeverHitsAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	pool := mqscToolPool(t, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "DISPLAY QLOCAL('X')\nALTER QLOCAL('X') MAXDEPTH(1000)",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("expected tool error for newline statement injection")
	}
	if fakeClient.ExecuteRawMQSCCalls != 0 {
		t.Fatalf("adapter invoked on newline injection: %d", fakeClient.ExecuteRawMQSCCalls)
	}
}

func TestExecuteMQSCReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	fakeClient.ExecuteRawMQSCResult = mqadmin.RawMQSCResult{
		Profile:      "prod",
		QueueManager: "QM1",
		Command:      "DISPLAY QLOCAL('DEV.QUEUE.1')",
		Completion: mqadmin.MQSCCompletion{
			OverallCompletionCode: 0,
			OverallReasonCode:     0,
		},
	}
	pool := mqscToolPool(t, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "DISPLAY QLOCAL('DEV.QUEUE.1')",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool error: %+v", res.Content)
	}
	if fakeClient.ExecuteRawMQSCCalls != 1 {
		t.Fatalf("execute calls = %d", fakeClient.ExecuteRawMQSCCalls)
	}
	payload, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structuredContent type = %T", res.StructuredContent)
	}
	if payload["profile"] != "prod" {
		t.Fatalf("profile = %#v", payload["profile"])
	}
}

func TestExecuteMQSCRedactsCommandInStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	fakeClient.ExecuteRawMQSCResult = mqadmin.RawMQSCResult{
		Completion: mqadmin.MQSCCompletion{
			OverallCompletionCode: 0,
			OverallReasonCode:     0,
		},
	}
	pool := mqscToolPool(t, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool), mcpserver.WithEnableMQSC(true))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	const secret = "super-secret-value"
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: toolExecuteMQSC,
		Arguments: map[string]any{
			"profile": "prod",
			"command": "DISPLAY QLOCAL('DEV') WHERE password=" + secret,
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
	cmd, ok := payload["command"].(string)
	if !ok {
		t.Fatalf("command = %#v", payload["command"])
	}
	if strings.Contains(cmd, secret) {
		t.Fatalf("structuredContent command leaked secret: %q", cmd)
	}
	if !strings.Contains(cmd, "[REDACTED]") {
		t.Fatalf("structuredContent command = %q, want redacted placeholder", cmd)
	}
}

const toolExecuteMQSC = "execute_mqsc"

func mqscToolPool(t *testing.T, fakeClient *fake.Client) *application.ProfilePool {
	t.Helper()
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(mqscToolProfileDoc))
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
