package mqweb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	mqtls "github.com/platformrelay/ibm-mq-mcp-server/internal/config/tls"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

func testProfile(srvURL string) catalog.Profile {
	return catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     srvURL,
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS: mqtls.Settings{InsecureSkipVerify: true},
	}
}

func TestAdminClientListChannelsParsesAndPaginates(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v2/admin/qmgr/QM1/channel" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"channel": []map[string]any{
				{"name": "A", "type": "sender"},
				{"name": "B", "type": "receiver"},
				{"name": "C", "type": "serverConnection"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	page, err := client.ListChannels(context.Background(), mqadmin.ListChannelsRequest{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || !page.Truncated || page.NextCursor != "2" {
		t.Fatalf("page = %+v", page)
	}
}

func TestAdminClientGetChannelStatusAvailable(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v2/admin/qmgr/QM1/channel/DEV.SVRCONN" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"channel": []map[string]any{
				{
					"name": "DEV.SVRCONN",
					"type": "serverConnection",
					"status": map[string]any{
						"state": "running",
					},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	status, err := client.GetChannelStatus(context.Background(), "DEV.SVRCONN")
	if err != nil {
		t.Fatal(err)
	}
	if status.Availability != mqadmin.Available || status.State != "running" {
		t.Fatalf("status = %+v", status)
	}
}

func TestAdminClientListListenersUnsupported(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.ListListeners(context.Background(), mqadmin.ListListenersRequest{Limit: 10})
	if err == nil {
		t.Fatal("expected unsupported error")
	}
	if _, ok := mqadmin.AsUnsupportedError(err); !ok {
		t.Fatalf("expected UnsupportedError, got %v", err)
	}
}

func TestAdminClientListSubscriptionsHappyPath(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v2/admin/qmgr/QM1/subscription" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"subscription": []map[string]any{
				{"id": "1", "name": "SUB.A", "topicString": "a/b", "type": "admin"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := mqweb.NewAdminClient(testProfile(srv.URL), secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	page, err := client.ListSubscriptions(context.Background(), mqadmin.ListSubscriptionsRequest{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "SUB.A" {
		t.Fatalf("page = %+v", page)
	}
}

func TestAdminClientGetChannelStatusUnavailableWithoutState(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v2/admin/qmgr/QM1/channel/DEV.SVRCONN" {
			http.NotFound(w, r)
			return
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

	status, err := client.GetChannelStatus(context.Background(), "DEV.SVRCONN")
	if err != nil {
		t.Fatal(err)
	}
	if status.Availability != mqadmin.Unavailable || status.Error == "" {
		t.Fatalf("status = %+v", status)
	}
}
