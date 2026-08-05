package tls_test

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	mqtls "github.com/platformrelay/ibm-mq-mcp-server/internal/config/tls"
)

func TestValidateRejectsInvalidCARef(t *testing.T) {
	settings := mqtls.Settings{CARef: "env:NOT_A_CA"}
	if err := settings.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBuildConfigInsecureSkipVerify(t *testing.T) {
	resolver := secrets.NewResolver()
	cfg, err := mqtls.BuildConfig(mqtls.Settings{InsecureSkipVerify: true}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("expected insecure skip verify")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("min version = %x", cfg.MinVersion)
	}
}

func TestBuildConfigWithCAFile(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	// Minimal non-parseable PEM still exercises file read path; use empty file to trigger parse error.
	certPEM := "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"
	if err := os.WriteFile(caPath, []byte(certPEM), 0o600); err != nil {
		t.Fatal(err)
	}
	resolver := secrets.NewResolver()
	settings := mqtls.Settings{CARef: "file:" + caPath}
	if err := settings.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := mqtls.BuildConfig(settings, resolver); err == nil {
		t.Fatal("expected CA parse error for invalid cert bytes")
	}
}

func TestApplyClientCertificateLoadsKeyPair(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeTestClientKeyPair(t, dir)
	resolver := secrets.NewResolver()
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if err := mqtls.ApplyClientCertificate(cfg, "file:"+certPath, "file:"+keyPath, "", resolver); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Certificates) < 1 {
		t.Fatal("expected client certificate on tls.Config")
	}
}
