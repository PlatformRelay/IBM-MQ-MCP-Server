package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
)

func TestNewServerHasNoTools(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	server := mcpserver.New()

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
	if len(res.Tools) != 0 {
		t.Fatalf("expected no tools on bootstrap server, got %d: %+v", len(res.Tools), res.Tools)
	}
}

func TestServerImplementationIdentity(t *testing.T) {
	t.Parallel()

	server := mcpserver.New()
	if server == nil {
		t.Fatal("New returned nil")
	}
}
