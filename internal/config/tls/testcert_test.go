package tls_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
