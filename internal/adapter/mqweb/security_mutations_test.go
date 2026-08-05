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
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func TestAdminClientDefineChannelUsesREST(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	var method, path string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"serverConnection"`) {
			t.Fatalf("body = %s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"channel": []map[string]any{
				{"name": "DEV.SVRCONN", "type": "serverConnection"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	result, err := client.DefineChannel(context.Background(), "DEV.SVRCONN", mqadmin.DefineChannelRequest{
		ChannelType: mqadmin.ChannelTypeServerConnection,
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost || !strings.HasSuffix(path, "/channel") {
		t.Fatalf("method=%s path=%s", method, path)
	}
	if result.ChannelName != "DEV.SVRCONN" {
		t.Fatalf("result = %+v", result)
	}
}

func TestAdminClientDefineCHLAUTHUsesRunCommandJSON(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	var path string
	var payload map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&payload)
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

	_, err = client.DefineCHLAUTH(context.Background(), mqadmin.DefineCHLAUTHRequest{
		Target: mqadmin.CHLAUTHTarget{
			ChannelName: "DEV.SVRCONN",
			RuleType:    mqadmin.CHLAUTHTypeAddressMap,
			Address:     "*",
		},
		UserSource: mqadmin.CHLAUTHUserSourceNoAccess,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "/mqsc") {
		t.Fatalf("path = %s", path)
	}
	if payload["type"] != "runCommandJSON" || payload["qualifier"] != "chlauth" {
		t.Fatalf("payload = %#v", payload)
	}
}
