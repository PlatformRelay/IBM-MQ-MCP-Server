// Package application wires configuration into runtime services.
package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	mqtls "github.com/platformrelay/ibm-mq-mcp-server/internal/config/tls"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// LoadCatalogFromFile reads YAML or JSON profile catalogs from disk.
func LoadCatalogFromFile(path string) (*catalog.Catalog, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: operator-supplied config path
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	switch {
	case hasSuffixFold(path, ".json"):
		return catalog.LoadJSON(data)
	default:
		return catalog.LoadYAML(data)
	}
}

func hasSuffixFold(path, suffix string) bool {
	if len(path) < len(suffix) {
		return false
	}
	end := path[len(path)-len(suffix):]
	for i := 0; i < len(suffix); i++ {
		a, b := end[i], suffix[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}

// ProfilePool lazily resolves credentials and reuses HTTP clients per profile.
type ProfilePool struct {
	catalog    *catalog.Catalog
	validation catalog.ValidationResult
	resolver   *secrets.Resolver
	gate       *PolicyGate

	mu        sync.Mutex
	admin     map[string]mqadmin.Client
	messaging map[string]messaging.Client
	closed    bool
}

// NewProfilePool constructs a pool for validated profiles.
func NewProfilePool(
	cat *catalog.Catalog,
	validation catalog.ValidationResult,
	resolver *secrets.Resolver,
	gate *PolicyGate,
) *ProfilePool {
	if resolver == nil {
		resolver = secrets.NewResolver()
	}
	if gate == nil {
		gate = NewPolicyGate()
	}
	return &ProfilePool{
		catalog:    cat,
		validation: validation,
		resolver:   resolver,
		gate:       gate,
		admin:      make(map[string]mqadmin.Client),
		messaging:  make(map[string]messaging.Client),
	}
}

// Gate returns the policy gate for call-time authorization from MCP tools.
func (p *ProfilePool) Gate() *PolicyGate {
	return p.gate
}

// Authorize checks capability for profile without resolving secrets or MQ clients.
func (p *ProfilePool) Authorize(name string, required policy.Capability, operation string) error {
	profile, err := p.requireProfile(name)
	if err != nil {
		return err
	}
	return p.gate.Authorize(profile, required, operation)
}

// Admin returns the administration client for a profile after capability authorization.
func (p *ProfilePool) Admin(name string, required policy.Capability) (mqadmin.Client, error) {
	profile, err := p.requireProfile(name)
	if err != nil {
		return nil, err
	}
	if err := p.gate.Authorize(profile, required, "mqadmin"); err != nil {
		return nil, err
	}
	return p.adminClient(name)
}

// Messaging returns the messaging client for a profile after capability authorization.
func (p *ProfilePool) Messaging(name string, required policy.Capability) (messaging.Client, error) {
	profile, err := p.requireProfile(name)
	if err != nil {
		return nil, err
	}
	if err := p.gate.Authorize(profile, required, "messaging"); err != nil {
		return nil, err
	}
	return p.messagingClient(name)
}

func (p *ProfilePool) adminClient(name string) (mqadmin.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, errors.New("profile pool closed")
	}
	if client, ok := p.admin[name]; ok {
		return client, nil
	}
	profile, err := p.requireProfile(name)
	if err != nil {
		return nil, err
	}
	client, err := newMQWebAdminClient(profile, p.resolver)
	if err != nil {
		return nil, err
	}
	p.admin[name] = client
	return client, nil
}

func (p *ProfilePool) messagingClient(name string) (messaging.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, errors.New("profile pool closed")
	}
	if client, ok := p.messaging[name]; ok {
		return client, nil
	}
	profile, err := p.requireProfile(name)
	if err != nil {
		return nil, err
	}
	client, err := newMQWebMessagingClient(profile, p.resolver)
	if err != nil {
		return nil, err
	}
	p.messaging[name] = client
	return client, nil
}

func (p *ProfilePool) requireProfile(name string) (catalog.Profile, error) {
	profile, ok := p.catalog.ProfileByName(name)
	if !ok {
		return catalog.Profile{}, fmt.Errorf("unknown profile %q", name)
	}
	if !p.validation.IsValid(name) {
		return catalog.Profile{}, fmt.Errorf("profile %q failed validation", name)
	}
	return profile, nil
}

// Close shuts down pooled clients.
func (p *ProfilePool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	var err error
	for _, client := range p.admin {
		err = errors.Join(err, client.Close())
	}
	for _, client := range p.messaging {
		err = errors.Join(err, client.Close())
	}
	return err
}

type mqwebClient struct {
	name       string
	endpoint   string
	httpClient *http.Client
	authType   catalog.AuthType
	username   string
	password   string
	closed     bool
}

func (c *mqwebClient) ProfileName() string { return c.name }

func (c *mqwebClient) Ping(ctx context.Context) error {
	if c.closed {
		return errors.New("mqweb client closed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (c *mqwebClient) Close() error {
	c.closed = true
	return nil
}

type adminClient struct{ mqwebClient }

type messagingClient struct{ mqwebClient }

func newMQWebAdminClient(profile catalog.Profile, resolver *secrets.Resolver) (mqadmin.Client, error) {
	base, err := newMQWebBaseClient(profile, resolver)
	if err != nil {
		return nil, err
	}
	return &adminClient{base}, nil
}

func newMQWebMessagingClient(profile catalog.Profile, resolver *secrets.Resolver) (messaging.Client, error) {
	base, err := newMQWebBaseClient(profile, resolver)
	if err != nil {
		return nil, err
	}
	return &messagingClient{base}, nil
}

func newMQWebBaseClient(profile catalog.Profile, resolver *secrets.Resolver) (mqwebClient, error) {
	creds, err := resolveAuth(profile.Authentication, resolver)
	if err != nil {
		return mqwebClient{}, err
	}
	tlsCfg, err := mqtls.BuildConfig(profile.TLS, resolver)
	if err != nil {
		return mqwebClient{}, err
	}
	if profile.Authentication.Type == catalog.AuthMTLS {
		auth := profile.Authentication
		certErr := mqtls.ApplyClientCertificate(
			tlsCfg,
			auth.CertificateRef,
			auth.PrivateKeyRef,
			auth.PassphraseRef,
			resolver,
		)
		if certErr != nil {
			return mqwebClient{}, certErr
		}
	}
	timeout, err := profileTimeout(profile.Timeout)
	if err != nil {
		return mqwebClient{}, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsCfg
	return mqwebClient{
		name:     profile.Name,
		endpoint: profile.Endpoint,
		authType: profile.Authentication.Type,
		username: creds.username,
		password: creds.password,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func profileTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return 30 * time.Second, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("timeout: %w", err)
	}
	return d, nil
}

type resolvedCredentials struct {
	username string
	password string
}

func resolveAuth(auth catalog.Authentication, resolver *secrets.Resolver) (resolvedCredentials, error) {
	switch auth.Type {
	case catalog.AuthBasic:
		ref, err := secrets.Parse(auth.SecretRef)
		if err != nil {
			return resolvedCredentials{}, err
		}
		secret, err := resolver.Resolve(ref)
		if err != nil {
			return resolvedCredentials{}, fmt.Errorf("resolve basic credentials: %w", err)
		}
		user, pass, err := parseBasicSecret(secret)
		if err != nil {
			return resolvedCredentials{}, err
		}
		return resolvedCredentials{username: user, password: pass}, nil
	case catalog.AuthMTLS:
		for _, raw := range []string{auth.CertificateRef, auth.PrivateKeyRef, auth.PassphraseRef} {
			if raw == "" {
				continue
			}
			ref, err := secrets.Parse(raw)
			if err != nil {
				return resolvedCredentials{}, err
			}
			if _, err := resolver.Resolve(ref); err != nil {
				return resolvedCredentials{}, fmt.Errorf("resolve mtls material: %w", err)
			}
		}
	default:
		return resolvedCredentials{}, fmt.Errorf("unsupported authentication type %q", auth.Type)
	}
	return resolvedCredentials{}, nil
}

func parseBasicSecret(value string) (username, password string, err error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", errors.New("basic credentials secret is empty")
	}
	user, pass, ok := strings.Cut(value, ":")
	if !ok || strings.TrimSpace(user) == "" {
		return "", "", errors.New("basic credentials must be username:password")
	}
	if strings.TrimSpace(pass) == "" {
		return "", "", errors.New("basic credentials password must not be empty")
	}
	return user, pass, nil
}

// ConfigReady reports whether readiness should succeed for the loaded catalog.
func ConfigReady(cat *catalog.Catalog, validation catalog.ValidationResult) bool {
	if cat == nil || len(cat.Profiles) == 0 {
		return true
	}
	return validation.AnyValid()
}
