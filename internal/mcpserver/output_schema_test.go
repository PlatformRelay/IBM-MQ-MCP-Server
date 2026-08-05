package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
)

func TestAllPublicToolsHaveOutputSchema(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
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
	for _, tool := range res.Tools {
		if tool.OutputSchema == nil {
			t.Errorf("tool %q missing outputSchema", tool.Name)
		}
	}
}

func TestListProfilesTextFallbackIsCompactNotJSON(t *testing.T) {
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
	if len(res.Content) != 1 {
		t.Fatalf("expected one text block, got %d", len(res.Content))
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T", res.Content[0])
	}
	if len(text.Text) > 0 && text.Text[0] == '{' {
		t.Fatalf("text fallback must not echo JSON object: %q", text.Text)
	}
	if text.Text == "" {
		t.Fatal("expected non-empty compact text")
	}
}

func TestListProfilesTextFallbackDeterministic(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)
	pool := testInspectPool(t, fake.New("prod"))
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	session := connectInspectClient(t, server)
	t.Cleanup(func() { _ = session.Close() })

	call := func() string {
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "list_profiles",
			Arguments: map[string]any{"limit": 10},
		})
		if err != nil {
			t.Fatal(err)
		}
		return res.Content[0].(*mcp.TextContent).Text
	}
	first := call()
	second := call()
	if first != second {
		t.Fatal("text fallback not deterministic")
	}
}
