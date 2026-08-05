package remotemcp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/remotemcp"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

const (
	queueManagerStatusToolCall = `{"jsonrpc":"2.0","id":2,"method":"tools/call",` +
		`"params":{"name":"queue_manager_status","arguments":{"profile":"prod"}}}`
)

func TestRemoteServerUnauthorizedWithoutBearer(t *testing.T) {
	t.Parallel()

	srv := newTestRemoteServer(t, "gate-token")
	t.Cleanup(func() { _ = srv.Close() })

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(initMessage()))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRemoteServerAcceptsAuthorizedInitialize(t *testing.T) {
	t.Parallel()

	srv := newTestRemoteServer(t, "gate-token")
	t.Cleanup(func() { _ = srv.Close() })

	body := postRemote(t, srv.URL, "gate-token", initMessage())
	if !strings.Contains(body, `"serverInfo"`) {
		t.Fatalf("expected initialize result, got %q", body)
	}
}

func TestConfusedDeputyClientBearerNotForwardedToMQ(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)

	var capturedAuth atomic.Value
	mqSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/admin/") {
			capturedAuth.Store(r.Header.Get("Authorization"))
		}
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"QM1","running":true}`))
	}))
	t.Cleanup(mqSrv.Close)

	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "mq-user:mq-pass")

	pool := testInspectPool(t, mqSrv.URL)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	rt := runtime.New()
	rt.SetConfigValid(true)

	remote, err := remotemcp.NewTestServer(remotemcp.Config{
		AuthToken:     "client-mcp-token",
		Limits:        remotemcp.DefaultLimits(),
		TransportName: "streamable-http",
	}, server, rt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = remote.Close() })

	init := postRemote(t, remote.URL, "client-mcp-token", initMessage())
	if !strings.Contains(init, `"serverInfo"`) {
		t.Fatalf("initialize failed: %s", init)
	}

	postRemote(t, remote.URL, "client-mcp-token", queueManagerStatusToolCall)

	auth, _ := capturedAuth.Load().(string)
	if strings.Contains(auth, "client-mcp-token") {
		t.Fatalf("client MCP bearer leaked to mqweb Authorization: %q", auth)
	}
	if auth == "" {
		t.Fatal("expected mqweb to receive profile basic auth, got empty Authorization")
	}
	if !strings.HasPrefix(auth, "Basic ") {
		t.Fatalf("expected Basic auth to mqweb, got %q", auth)
	}
}

func TestUnauthorizedProfileAccessDeniedOverRemote(t *testing.T) {
	t.Cleanup(mcpserver.ResetRegisteredTools)

	mqSrv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		t.Error("mqweb should not be called when policy denies")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(mqSrv.Close)

	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "mq-user:mq-pass")
	pool := testBrowseOnlyPool(t, mqSrv.URL)
	server := mcpserver.NewWithInspector(application.NewInspector(pool))
	rt := runtime.New()
	rt.SetConfigValid(true)

	remote, err := remotemcp.NewTestServer(remotemcp.Config{
		AuthToken:     "gate-token",
		Limits:        remotemcp.DefaultLimits(),
		TransportName: "streamable-http",
	}, server, rt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = remote.Close() })

	postRemote(t, remote.URL, "gate-token", initMessage())
	body := postRemoteExpectErrorResult(t, remote.URL, "gate-token", queueManagerStatusToolCall)
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "capability") && !strings.Contains(lower, "denied") {
		t.Fatalf("expected policy denial in tool result, got %q", body)
	}
}

func newTestRemoteServer(t *testing.T, token string) *remotemcp.TestListener {
	t.Helper()

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	rt := runtime.New()
	rt.SetConfigValid(true)

	srv, err := remotemcp.NewTestServer(remotemcp.Config{
		AuthToken:     token,
		Limits:        remotemcp.DefaultLimits(),
		TransportName: "streamable-http",
	}, mcpServer, rt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

func postRemote(t *testing.T, url, token, msg string) string {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body = %s", resp.StatusCode, body)
	}
	return string(body)
}

func postRemoteExpectErrorResult(t *testing.T, url, token, msg string) string {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	setMCPHeaders(req)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body = %s", resp.StatusCode, body)
	}
	return string(body)
}

func setMCPHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
}

func initMessage() string {
	return `{"jsonrpc":"2.0","id":1,"method":"initialize","params":` +
		`{"protocolVersion":"2025-11-25","capabilities":{},` +
		`"clientInfo":{"name":"test","version":"v0"}}}`
}

func testInspectPool(t *testing.T, mqEndpoint string) *application.ProfilePool {
	t.Helper()
	t.Setenv("IBM_MQ_MCP_TOOL_SECRET", "mq-user:mq-pass")
	dir := t.TempDir()
	path := writeProfileYAML(t, dir, mqEndpoint, "inspect")
	pool, err := loadProfilesFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func testBrowseOnlyPool(t *testing.T, mqEndpoint string) *application.ProfilePool {
	t.Helper()
	dir := t.TempDir()
	path := writeProfileYAML(t, dir, mqEndpoint, "browse")
	pool, err := loadProfilesFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}

func loadProfilesFromPath(path string) (*application.ProfilePool, error) {
	cat, err := application.LoadCatalogFromFile(path)
	if err != nil {
		return nil, err
	}
	validation := cat.Validate()
	pool := application.NewProfilePool(cat, validation, nil, nil,
		application.WithAdminFactory(mqweb.NewAdminClient),
		application.WithMessagingFactory(mqweb.NewMessagingClient),
	)
	return pool, nil
}
