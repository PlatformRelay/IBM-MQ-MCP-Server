// Package application wires configuration into runtime services.
package application

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
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
	catalog          *catalog.Catalog
	validation       catalog.ValidationResult
	resolver         *secrets.Resolver
	gate             *PolicyGate
	adminFactory     AdminClientFactory
	messagingFactory MessagingClientFactory

	mu        sync.Mutex
	admin     map[string]mqadmin.Client
	messaging map[string]messaging.Client
	closed    bool
}

// ProfilePoolOption configures a ProfilePool.
type ProfilePoolOption func(*ProfilePool)

// WithAdminFactory injects the mqadmin client constructor (typically adapter/mqweb).
func WithAdminFactory(factory AdminClientFactory) ProfilePoolOption {
	return func(p *ProfilePool) {
		p.adminFactory = factory
	}
}

// WithMessagingFactory injects the messaging client constructor (typically adapter/mqweb).
func WithMessagingFactory(factory MessagingClientFactory) ProfilePoolOption {
	return func(p *ProfilePool) {
		p.messagingFactory = factory
	}
}

// NewProfilePool constructs a pool for validated profiles.
func NewProfilePool(
	cat *catalog.Catalog,
	validation catalog.ValidationResult,
	resolver *secrets.Resolver,
	gate *PolicyGate,
	opts ...ProfilePoolOption,
) *ProfilePool {
	if resolver == nil {
		resolver = secrets.NewResolver()
	}
	if gate == nil {
		gate = NewPolicyGate()
	}
	p := &ProfilePool{
		catalog:    cat,
		validation: validation,
		resolver:   resolver,
		gate:       gate,
		admin:      make(map[string]mqadmin.Client),
		messaging:  make(map[string]messaging.Client),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
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
	if p.adminFactory == nil {
		return nil, errors.New("admin client factory is not configured")
	}
	client, err := p.adminFactory(profile, p.resolver)
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
	if p.messagingFactory == nil {
		return nil, errors.New("messaging client factory is not configured")
	}
	client, err := p.messagingFactory(profile, p.resolver)
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

// ConfigReady reports whether readiness should succeed for the loaded catalog.
func ConfigReady(cat *catalog.Catalog, validation catalog.ValidationResult) bool {
	if cat == nil || len(cat.Profiles) == 0 {
		return true
	}
	return validation.AnyValid()
}
