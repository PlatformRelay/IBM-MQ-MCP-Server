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

func TestAdminClientQueueManagerStatusHappyPath(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ibmmq/rest/v3/admin/qmgr/QM1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":    "QM1",
				"state":   "running",
				"running": true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	profile := catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     srv.URL,
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS: mqtls.Settings{InsecureSkipVerify: true},
	}

	client, err := mqweb.NewAdminClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	status, err := client.QueueManagerStatus(context.Background(), "QM1")
	if err != nil {
		t.Fatal(err)
	}
	if status.Availability != mqadmin.Available {
		t.Fatalf("availability = %q", status.Availability)
	}
	if status.Identity.Observed != "QM1" {
		t.Fatalf("observed = %q", status.Identity.Observed)
	}
}

func TestAdminClientListQueuesParsesAndPaginates(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v1/admin/qmgr/QM1/queue" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"queue": []map[string]any{
				{"name": "A", "type": "local"},
				{"name": "B", "type": "local"},
				{"name": "C", "type": "alias"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	profile := catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     srv.URL,
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS: mqtls.Settings{InsecureSkipVerify: true},
	}

	client, err := mqweb.NewAdminClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	page, err := client.ListQueues(context.Background(), mqadmin.ListQueuesRequest{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || !page.Truncated || page.NextCursor != "2" {
		t.Fatalf("page = %+v", page)
	}
}

func TestAdminClientAppliesProfileTimeout(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	profile := catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     "https://mq.example.test:9443",
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS:     mqtls.Settings{InsecureSkipVerify: true},
		Timeout: "45s",
	}
	client, err := mqweb.NewAdminClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
}

func TestAdminClientMapsHTTP403ToReasonCode(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	profile := catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     srv.URL,
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS: mqtls.Settings{InsecureSkipVerify: true},
	}

	client, err := mqweb.NewAdminClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.GetQueue(context.Background(), "MISSING")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := mqadmin.AsReasonError(err); !ok {
		t.Fatalf("expected ReasonError, got %v", err)
	}
	cause, _ := mqadmin.ClassifyConnectivityError(err)
	if cause != mqadmin.FailureAuthorization {
		t.Fatalf("403 should classify as authorization, got %q", cause)
	}
}

func TestAdminClientMapsHTTP401ToAuthentication(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	profile := catalog.Profile{
		Name:         "prod",
		QueueManager: "QM1",
		Endpoint:     srv.URL,
		Authentication: catalog.Authentication{
			Type:      catalog.AuthBasic,
			SecretRef: "env:IBM_MQ_MCP_TEST_SECRET",
		},
		TLS: mqtls.Settings{InsecureSkipVerify: true},
	}

	client, err := mqweb.NewAdminClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	_, err = client.GetQueue(context.Background(), "MISSING")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := mqadmin.AsReasonError(err); ok {
		t.Fatalf("401 must not map to ReasonError, got %v", err)
	}
	cause, _ := mqadmin.ClassifyConnectivityError(err)
	if cause != mqadmin.FailureAuthentication {
		t.Fatalf("401 should classify as authentication, got %q", cause)
	}
}
