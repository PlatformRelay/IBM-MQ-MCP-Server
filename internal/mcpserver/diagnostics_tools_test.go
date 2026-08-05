package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

func TestDiagnosticsToolsRegistered(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Tools) != 16 {
		t.Fatalf("expected 16 tools, got %d", len(res.Tools))
	}
}

func TestExplainMQReasonCodeToolReturnsStructuredContent(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	server := mcpserver.NewWithInspector(nil)
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "explain_mq_reason_code",
		Arguments: map[string]any{"reasonCode": 2035},
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
	if payload["known"] != true {
		t.Fatalf("known = %#v", payload["known"])
	}
	if payload["symbol"] != "MQRC_NOT_AUTHORIZED" {
		t.Fatalf("symbol = %#v", payload["symbol"])
	}
}

func TestCheckProfileConnectivityDeniedBeforeAdapter(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	fakeClient := fake.New("prod")
	pool := testInspectPoolWithDoc(t, browseOnlyProfileDoc, fakeClient)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "check_profile_connectivity",
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

func TestRegisteredDiagnosticsToolSpecs(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, nil)
	mcpserver.NewWithInspector(application.NewInspector(pool))
	for _, spec := range mcpserver.RegisteredTools {
		switch spec.Name {
		case "explain_mq_reason_code":
			if spec.RequiredCapability != "" {
				t.Fatalf("explain tool should not require capability, got %q", spec.RequiredCapability)
			}
		case "check_profile_connectivity":
			if spec.RequiredCapability != policy.Inspect {
				t.Fatalf("connectivity tool capability = %q", spec.RequiredCapability)
			}
		}
	}
}
