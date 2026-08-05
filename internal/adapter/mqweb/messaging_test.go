package mqweb_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

func TestBrowseMessagesUsesGETOnlyNeverDELETE(t *testing.T) {
	var mu sync.Mutex
	methods := make([]string, 0, 4)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		methods = append(methods, r.Method+" "+r.URL.Path)
		mu.Unlock()
		switch {
		case strings.Contains(r.URL.Path, "/messagelist"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"messageId":"ID:abc","format":"MQSTR","messageLength":5}]`))
		case strings.Contains(r.URL.Path, "/message"):
			_, _ = w.Write([]byte("hello"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	client := newTestMessagingClient(t, server.URL)
	page, err := client.BrowseMessages(context.Background(), "Q1", messaging.BrowseRequest{
		Count:          1,
		IncludePayload: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].Payload == "" {
		t.Fatal("expected payload when include_payload=true")
	}
	mu.Lock()
	defer mu.Unlock()
	for _, call := range methods {
		if strings.HasPrefix(call, "DELETE ") {
			t.Fatalf("destructive DELETE invoked: %q", call)
		}
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("unexpected method: %q", call)
		}
	}
}

func TestBrowseMessagesMetadataOnlyOmitsPayload(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/messagelist") {
			_, _ = w.Write([]byte(`[{"messageId":"ID:1","messageLength":3}]`))
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	}))
	t.Cleanup(server.Close)

	client := newTestMessagingClient(t, server.URL)
	page, err := client.BrowseMessages(context.Background(), "Q1", messaging.BrowseRequest{Count: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].Payload != "" {
		t.Fatalf("payload = %q", page.Items[0].Payload)
	}
	if page.Items[0].Encoding != messaging.EncodingOmitted {
		t.Fatalf("encoding = %q", page.Items[0].Encoding)
	}
}

func TestPutMessageUsesPOSTWithCSRFAndContentType(t *testing.T) {
	var method, contentType, csrf string
	var body []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		contentType = r.Header.Get("Content-Type")
		csrf = r.Header.Get("ibm-mq-rest-csrf-token")
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("ibm-mq-md-messageId", "ID:put1")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := newTestMessagingClientWithCaps(t, server.URL, []string{"produce"})
	result, err := client.PutMessage(context.Background(), "Q1", messaging.PutRequest{
		ContentType: messaging.ContentTypeTextPlain,
		Payload:     "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if method != http.MethodPost {
		t.Fatalf("method = %q", method)
	}
	if csrf == "" {
		t.Fatal("expected csrf header")
	}
	if contentType != "text/plain;charset=utf-8" {
		t.Fatalf("contentType = %q", contentType)
	}
	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}
	if result.MessageID != "ID:put1" {
		t.Fatalf("messageId = %q", result.MessageID)
	}
}

func TestPutMessageOctetStreamDecodesBase64BeforePOST(t *testing.T) {
	var body []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("ibm-mq-md-messageId", "ID:bin")
	}))
	t.Cleanup(server.Close)

	raw := []byte{0x01, 0x02, 0x03}
	client := newTestMessagingClientWithCaps(t, server.URL, []string{"produce"})
	_, err := client.PutMessage(context.Background(), "Q1", messaging.PutRequest{
		ContentType: messaging.ContentTypeOctetStream,
		Payload:     base64.StdEncoding.EncodeToString(raw),
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != string(raw) {
		t.Fatalf("body = %v want %v", body, raw)
	}
}

func newTestMessagingClient(t *testing.T, endpoint string) messaging.Client {
	t.Helper()
	return newTestMessagingClientWithCaps(t, endpoint, []string{"browse"})
}

func newTestMessagingClientWithCaps(t *testing.T, endpoint string, caps []string) messaging.Client {
	t.Helper()
	capsYAML := ""
	for _, cap := range caps {
		capsYAML += "      - " + cap + "\n"
	}
	cat, err := catalog.LoadYAML([]byte(`
profiles:
  prod:
    queueManager: QM1
    endpoint: ` + endpoint + `
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_MSG_TEST_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
` + capsYAML))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("IBM_MQ_MCP_MSG_TEST_SECRET", "user:pass")
	profile, ok := cat.ProfileByName("prod")
	if !ok {
		t.Fatal("missing profile")
	}
	client, err := mqweb.NewMessagingClient(profile, secrets.NewResolver())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
