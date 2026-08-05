package mqweb_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func TestAdminClientExecuteRawMQSCUsesRunCommand(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	var path string
	var payload map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"overallCompletionCode": 0,
			"overallReasonCode":     0,
		})
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.ExecuteRawMQSC(context.Background(), "DISPLAY QLOCAL('DEV.QUEUE.1')")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "/mqsc") {
		t.Fatalf("path = %s", path)
	}
	if payload["type"] != "runCommand" {
		t.Fatalf("payload = %#v", payload)
	}
	params, ok := payload["parameters"].(map[string]any)
	if !ok || params["command"] != "DISPLAY QLOCAL('DEV.QUEUE.1')" {
		t.Fatalf("parameters = %#v", payload["parameters"])
	}
}
