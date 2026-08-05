// Package tls builds crypto/tls settings for mqweb profiles (ADR-0004).
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
)

// Settings describes TLS options on a connection profile.
type Settings struct {
	CARef              string `yaml:"caRef" json:"caRef"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify" json:"insecureSkipVerify"`
}

// Validate checks reference syntax without reading secret values.
func (s Settings) Validate() error {
	if s.CARef == "" {
		return nil
	}
	if _, err := secrets.Parse(s.CARef); err != nil {
		return fmt.Errorf("tls.caRef: %w", err)
	}
	ref, _ := secrets.Parse(s.CARef)
	if ref.Provider != secrets.ProviderFile {
		return errors.New("tls.caRef must use file: prefix")
	}
	return nil
}

// BuildConfig returns a tls.Config for HTTP clients. CA file is read when caRef is set.
func BuildConfig(settings Settings, resolver *secrets.Resolver) (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: settings.InsecureSkipVerify, //nolint:gosec // explicit opt-in for local Kind
	}
	if settings.CARef == "" {
		return cfg, nil
	}
	ref, err := secrets.Parse(settings.CARef)
	if err != nil {
		return nil, err
	}
	if ref.Provider != secrets.ProviderFile {
		return nil, errors.New("tls.caRef must use file: prefix")
	}
	pem, err := resolver.Resolve(ref)
	if err != nil {
		return nil, fmt.Errorf("load CA: %w", err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM([]byte(pem)) {
		// Allow PEM file to be a full bundle replacing system roots when append fails.
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(pem)) {
			return nil, fmt.Errorf("parse CA from %q", ref.Name)
		}
	}
	cfg.RootCAs = pool
	return cfg, nil
}

// ApplyClientCertificate loads PEM client certificate and key refs into cfg.Certificates.
func ApplyClientCertificate(
	cfg *tls.Config,
	certRef, keyRef, passphraseRef string,
	resolver *secrets.Resolver,
) error {
	if cfg == nil {
		return errors.New("tls config is nil")
	}
	certPEM, err := resolveFilePEM(certRef, resolver)
	if err != nil {
		return fmt.Errorf("client certificate: %w", err)
	}
	keyPEM, err := resolveFilePEM(keyRef, resolver)
	if err != nil {
		return fmt.Errorf("client private key: %w", err)
	}
	var passphrase []byte
	if passphraseRef != "" {
		passRef, parseErr := secrets.Parse(passphraseRef)
		if parseErr != nil {
			return fmt.Errorf("passphraseRef: %w", parseErr)
		}
		value, resolveErr := resolver.Resolve(passRef)
		if resolveErr != nil {
			return fmt.Errorf("resolve key passphrase: %w", resolveErr)
		}
		passphrase = []byte(value)
	}
	pair, err := loadX509KeyPair(certPEM, keyPEM, passphrase)
	if err != nil {
		return fmt.Errorf("load client key pair: %w", err)
	}
	cfg.Certificates = []tls.Certificate{pair}
	return nil
}

func resolveFilePEM(ref string, resolver *secrets.Resolver) (string, error) {
	parsed, err := secrets.Parse(ref)
	if err != nil {
		return "", err
	}
	if parsed.Provider != secrets.ProviderFile {
		return "", errors.New("must use file: prefix")
	}
	return resolver.Resolve(parsed)
}

func loadX509KeyPair(certPEM, keyPEM string, passphrase []byte) (tls.Certificate, error) {
	if len(passphrase) == 0 {
		return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	}
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return tls.Certificate{}, errors.New("failed to decode private key PEM")
	}
	if x509.IsEncryptedPEMBlock(keyBlock) { //nolint:staticcheck // encrypted PEM keys remain common in MQ deployments
		decrypted, err := x509.DecryptPEMBlock(keyBlock, passphrase) //nolint:staticcheck // see above
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("decrypt private key: %w", err)
		}
		keyBlock.Bytes = decrypted
		keyBlock.Headers = nil
		keyPEM = string(pem.EncodeToMemory(keyBlock))
	}
	return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
}

// CAFileExists reports whether a caRef file is present (startup check helper).
func CAFileExists(caRef string) error {
	if caRef == "" {
		return nil
	}
	ref, err := secrets.Parse(caRef)
	if err != nil {
		return err
	}
	if ref.Provider != secrets.ProviderFile {
		return errors.New("tls.caRef must use file: prefix")
	}
	if _, err := os.Stat(ref.Name); err != nil {
		return fmt.Errorf("tls.caRef file %q: %w", ref.Name, err)
	}
	return nil
}
