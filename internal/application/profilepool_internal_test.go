package application

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func TestProfilePoolIsolatesDistinctProfiles(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_PROD_SECRET", "prod-user:prod-pass")
	t.Setenv("IBM_MQ_MCP_DEV_SECRET", "dev-user:dev-pass")
	pool := newTestPool(t, twoProfileYAML())

	prodAdmin, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	devAdmin, err := pool.Admin("dev")
	if err != nil {
		t.Fatal(err)
	}
	if prodAdmin == devAdmin {
		t.Fatal("expected distinct admin client instances per profile")
	}
	prodAC, ok := prodAdmin.(*adminClient)
	if !ok {
		t.Fatal("expected adminClient for prod")
	}
	devAC, ok := devAdmin.(*adminClient)
	if !ok {
		t.Fatal("expected adminClient for dev")
	}
	if prodAC.endpoint == devAC.endpoint {
		t.Fatalf("expected different endpoints, both %q", prodAC.endpoint)
	}
	if prodAC.username == devAC.username || prodAC.password == devAC.password {
		t.Fatalf("expected distinct credentials, got %q/%q and %q/%q",
			prodAC.username, prodAC.password, devAC.username, devAC.password)
	}

	prodMsg, err := pool.Messaging("prod")
	if err != nil {
		t.Fatal(err)
	}
	devMsg, err := pool.Messaging("dev")
	if err != nil {
		t.Fatal(err)
	}
	if prodMsg == devMsg {
		t.Fatal("expected distinct messaging client instances per profile")
	}
	if prodAdmin == prodMsg {
		t.Fatal("admin and messaging clients for same profile should be distinct instances")
	}

	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Admin("prod"); err == nil {
		t.Fatal("expected prod admin error after pool close")
	}
	if _, err := pool.Admin("dev"); err == nil {
		t.Fatal("expected dev admin error after pool close")
	}
}

func TestProfilePoolReusesAdminClient(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "user:pass")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_POOL_SECRET", ""))

	first, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	second, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("expected same admin client instance on second call")
	}
}

func TestProfilePoolReusesMessagingClient(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "user:pass")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_POOL_SECRET", ""))

	first, err := pool.Messaging("prod")
	if err != nil {
		t.Fatal(err)
	}
	second, err := pool.Messaging("prod")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("expected same messaging client instance on second call")
	}
}

func TestProfilePoolCloseRejectsFurtherUse(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "user:pass")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_POOL_SECRET", ""))

	if _, err := pool.Admin("prod"); err != nil {
		t.Fatal(err)
	}
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Admin("prod"); err == nil {
		t.Fatal("expected error after pool close")
	} else if !strings.Contains(err.Error(), "closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProfilePoolAppliesTimeout(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "user:pass")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_POOL_SECRET", "45s"))

	client, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	ac, ok := client.(*adminClient)
	if !ok {
		t.Fatal("expected adminClient concrete type")
	}
	if ac.httpClient.Timeout != 45*time.Second {
		t.Fatalf("timeout = %v, want 45s", ac.httpClient.Timeout)
	}
}

func TestProfileTimeoutRejectsInvalidDuration(t *testing.T) {
	if _, err := profileTimeout("not-a-duration"); err == nil {
		t.Fatal("expected timeout parse error")
	}
}

func TestProfilePoolMTLSLoadsClientCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeTestClientKeyPair(t, dir)
	doc := `
profiles:
  mtls:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: mtls
      certificateRef: file:` + certPath + `
      privateKeyRef: file:` + keyPath + `
    tls:
      insecureSkipVerify: true
`
	pool := newTestPool(t, doc)
	client, err := pool.Admin("mtls")
	if err != nil {
		t.Fatal(err)
	}
	ac, ok := client.(*adminClient)
	if !ok {
		t.Fatal("expected adminClient")
	}
	transport, ok := ac.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected http.Transport")
	}
	if transport.TLSClientConfig == nil || len(transport.TLSClientConfig.Certificates) < 1 {
		t.Fatal("expected mTLS client certificate on transport")
	}
}

func TestProfilePoolStoresBasicCredentials(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_POOL_SECRET", "alice:secret")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_POOL_SECRET", ""))

	client, err := pool.Admin("prod")
	if err != nil {
		t.Fatal(err)
	}
	ac, ok := client.(*adminClient)
	if !ok {
		t.Fatal("expected adminClient")
	}
	if ac.username != "alice" || ac.password != "secret" {
		t.Fatalf("stored credentials = %q / %q", ac.username, ac.password)
	}
}

func TestParseBasicSecretRejectsMalformed(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "no colon", value: "useronly"},
		{name: "empty username", value: ":pass"},
		{name: "empty password", value: "user:"},
		{name: "whitespace password", value: "user:   "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseBasicSecret(tc.value); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestProfilePoolRejectsEmptyBasicSecret(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_EMPTY_BASIC", " ")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_EMPTY_BASIC", ""))
	if _, err := pool.Admin("prod"); err == nil {
		t.Fatal("expected empty basic secret error")
	}
}

func TestProfilePoolRejectsMalformedBasicSecret(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_BAD_BASIC", "not-valid-format")
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_BAD_BASIC", ""))
	if _, err := pool.Admin("prod"); err == nil {
		t.Fatal("expected malformed basic secret error")
	}
}

func TestProfilePoolBasicAuthErrorsDoNotLeakSecret(t *testing.T) {
	const secret = "super-secret-value"
	t.Setenv("IBM_MQ_MCP_LEAK_BASIC", secret)
	pool := newTestPool(t, basicProfileYAML("env:IBM_MQ_MCP_LEAK_BASIC", ""))
	_, err := pool.Admin("prod")
	if err == nil {
		t.Fatal("expected malformed basic secret error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("error leaked secret: %v", err)
	}
}

func newTestPool(t *testing.T, doc string) *ProfilePool {
	t.Helper()
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	pool := NewProfilePool(cat, cat.Validate(), secrets.NewResolver())
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}

func twoProfileYAML() string {
	return `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq-prod.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_PROD_SECRET
    tls:
      insecureSkipVerify: true
  dev:
    queueManager: QM2
    endpoint: https://mq-dev.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_DEV_SECRET
    tls:
      insecureSkipVerify: true
`
}

func basicProfileYAML(secretRef, timeout string) string {
	timeoutLine := ""
	if timeout != "" {
		timeoutLine = "    timeout: " + timeout + "\n"
	}
	return `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: ` + secretRef + `
` + timeoutLine + `    tls:
      insecureSkipVerify: true
`
}

func writeTestClientKeyPair(t *testing.T, dir string) (certPath, keyPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "ibm-mq-mcp-test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPath = filepath.Join(dir, "client.pem")
	keyPath = filepath.Join(dir, "client-key.pem")
	certOut, err := os.Create(certPath) //nolint:gosec // G304: test temp dir paths
	if err != nil {
		t.Fatal(err)
	}
	certBlock := &pem.Block{Type: "CERTIFICATE", Bytes: der}
	if encodeErr := pem.Encode(certOut, certBlock); encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if closeErr := certOut.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	keyOut, err := os.Create(keyPath) //nolint:gosec // G304: test temp dir paths
	if err != nil {
		t.Fatal(err)
	}
	keyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	if encodeErr := pem.Encode(keyOut, keyBlock); encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if closeErr := keyOut.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	return certPath, keyPath
}
