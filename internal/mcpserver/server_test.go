package mcpserver_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
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

func TestTrackTransportUpdatesRuntimeState(t *testing.T) {
	t.Parallel()

	rt := runtime.New()
	rt.SetConfigValid(true)

	stop := mcpserver.TrackTransport(rt, "stdio")
	if !rt.Ready() {
		t.Fatal("expected ready while transport tracked")
	}

	stop()
	if rt.Ready() {
		t.Fatal("expected not ready after transport stopped")
	}
}

func TestTrackTransportNilRuntimeSafe(t *testing.T) {
	t.Parallel()

	stop := mcpserver.TrackTransport(nil, "stdio")
	stop()
}
