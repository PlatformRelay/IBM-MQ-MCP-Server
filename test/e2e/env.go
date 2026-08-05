//go:build e2e

package e2e

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

const (
	defaultEndpoint   = "https://127.0.0.1:30443"
	defaultMQHost     = "mq.localhost"
	defaultQueueMgr   = "QM1"
	defaultUser       = "admin"
	defaultPassword   = "passw0rd"
	defaultInsecureTL = "true"
)

type mqEnv struct {
	endpoint    *url.URL
	host        string
	queueMgr    string
	user        string
	password    string
	httpClient  *http.Client
}

func e2eEnabled() bool {
	return os.Getenv("IBM_MQ_MCP_E2E") == "1"
}

func requireE2E(t *testing.T) mqEnv {
	t.Helper()
	if !e2eEnabled() {
		t.Skip("IBM MQ e2e disabled; set IBM_MQ_MCP_E2E=1 and provision MQ (see docs/development/local-mq.md)")
	}
	cfg, err := loadMQEnv()
	if err != nil {
		t.Fatalf("load MQ e2e config: %v", err)
	}
	return cfg
}

func loadMQEnv() (mqEnv, error) {
	endpointRaw := envOr("IBM_MQ_MCP_MQ_ENDPOINT", defaultEndpoint)
	u, err := url.Parse(endpointRaw)
	if err != nil {
		return mqEnv{}, fmt.Errorf("parse IBM_MQ_MCP_MQ_ENDPOINT: %w", err)
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if envOr("IBM_MQ_MCP_MQ_INSECURE_TLS", defaultInsecureTL) == "true" {
		tlsCfg.InsecureSkipVerify = true //nolint:gosec // local mkcert / Docker dev only
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsCfg

	host := os.Getenv("IBM_MQ_MCP_MQ_HOST")
	if host == "" && envOr("IBM_MQ_MCP_MQ_ENDPOINT", defaultEndpoint) == defaultEndpoint {
		host = defaultMQHost
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &hostHeaderRoundTripper{
			host:      host,
			transport: transport,
		},
	}

	password := os.Getenv("IBM_MQ_MCP_MQ_PASSWORD")
	if password == "" {
		password = os.Getenv("MQ_ADMIN_PASSWORD")
	}
	if password == "" {
		password = defaultPassword
	}

	return mqEnv{
		endpoint:   u,
		host:       host,
		queueMgr:   envOr("IBM_MQ_MCP_MQ_QMGR", defaultQueueMgr),
		user:       envOr("IBM_MQ_MCP_MQ_USER", defaultUser),
		password:   password,
		httpClient: client,
	}, nil
}

func (e mqEnv) url(path string) string {
	ref, err := e.endpoint.Parse(path)
	if err != nil {
		return e.endpoint.String() + path
	}
	return ref.String()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type hostHeaderRoundTripper struct {
	host      string
	transport http.RoundTripper
}

func (rt *hostHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.host != "" {
		req = req.Clone(req.Context())
		req.Host = rt.host
		req.Header.Set("Host", rt.host)
	}
	return rt.transport.RoundTrip(req)
}
