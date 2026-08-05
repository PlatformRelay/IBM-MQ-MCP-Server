package secrets_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

func TestParseEnvReference(t *testing.T) {
	ref, err := secrets.Parse("env:MQ_PASSWORD")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Provider != secrets.ProviderEnv || ref.Name != "MQ_PASSWORD" {
		t.Fatalf("ref = %+v", ref)
	}
}

func TestParseFileReference(t *testing.T) {
	ref, err := secrets.Parse("file:/run/secrets/mq/pass")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Provider != secrets.ProviderFile || ref.Name != "/run/secrets/mq/pass" {
		t.Fatalf("ref = %+v", ref)
	}
}

func TestParseRejectsInlineAndUnknown(t *testing.T) {
	cases := []string{"", "vault:secret/data/mq", "super-secret-value", "env:", "k8s:bad-ref"}
	for _, c := range cases {
		if _, err := secrets.Parse(c); err == nil {
			t.Fatalf("Parse(%q) expected error", c)
		}
	}
}

func TestResolveEnv(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_TEST_SECRET", "user:pass")
	resolver := secrets.NewResolver()
	ref, err := secrets.Parse("env:IBM_MQ_MCP_TEST_SECRET")
	if err != nil {
		t.Fatal(err)
	}
	value, err := resolver.Resolve(ref)
	if err != nil {
		t.Fatal(err)
	}
	if value != "user:pass" {
		t.Fatalf("value = %q", value)
	}
}

func TestResolveMissingEnv(t *testing.T) {
	_ = os.Unsetenv("IBM_MQ_MCP_MISSING_SECRET")
	resolver := secrets.NewResolver()
	ref, err := secrets.Parse("env:IBM_MQ_MCP_MISSING_SECRET")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(ref); err == nil {
		t.Fatal("expected error for missing env")
	}
}

func TestResolveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cred")
	if err := os.WriteFile(path, []byte("alice:secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolver := secrets.NewResolver()
	ref, err := secrets.Parse("file:" + path)
	if err != nil {
		t.Fatal(err)
	}
	value, err := resolver.Resolve(ref)
	if err != nil {
		t.Fatal(err)
	}
	if value != "alice:secret" {
		t.Fatalf("value = %q", value)
	}
}

func TestResolveErrorsDoNotEchoSecretValues(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_LEAK_TEST", "must-not-appear-in-error")
	resolver := secrets.NewResolver()
	ref, err := secrets.Parse("file:/no/such/file")
	if err != nil {
		t.Fatal(err)
	}
	_, err = resolver.Resolve(ref)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "must-not-appear-in-error") {
		t.Fatalf("error leaked secret: %v", err)
	}
}
