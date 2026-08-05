// Package catalog loads and validates MQ connection profile catalogs (ADR-0004).
package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	mqtls "github.com/platformrelay/ibm-mq-mcp-server/internal/config/tls"
)

// AuthType names a supported mqweb authentication method.
type AuthType string

const (
	// AuthBasic selects HTTP Basic authentication to mqweb.
	AuthBasic AuthType = "basic"
	// AuthMTLS selects client-certificate authentication to mqweb.
	AuthMTLS AuthType = "mtls"
)

// Authentication describes mqweb credentials for a profile.
type Authentication struct {
	Type           AuthType `yaml:"type" json:"type"`
	SecretRef      string   `yaml:"secretRef,omitempty" json:"secretRef,omitempty"`
	CertificateRef string   `yaml:"certificateRef,omitempty" json:"certificateRef,omitempty"`
	PrivateKeyRef  string   `yaml:"privateKeyRef,omitempty" json:"privateKeyRef,omitempty"`
	PassphraseRef  string   `yaml:"passphraseRef,omitempty" json:"passphraseRef,omitempty"`
	// Inline fields are rejected — credentials must use refs only.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// Profile is one named queue-manager connection.
type Profile struct {
	Name           string         `yaml:"-" json:"-"`
	QueueManager   string         `yaml:"queueManager" json:"queueManager"`
	Endpoint       string         `yaml:"endpoint" json:"endpoint"`
	Authentication Authentication `yaml:"authentication" json:"authentication"`
	TLS            mqtls.Settings `yaml:"tls" json:"tls"`
	Capabilities   []string       `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Timeout        string         `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Catalog holds all configured profiles keyed by stable name.
type Catalog struct {
	Profiles map[string]Profile
}

type fileShape struct {
	Profiles map[string]Profile `yaml:"profiles" json:"profiles"`
}

// LoadYAML parses a profile catalog from YAML bytes.
func LoadYAML(data []byte) (*Catalog, error) {
	if err := checkDuplicateYAMLProfileKeys(data); err != nil {
		return nil, err
	}
	return load(data, yaml.Unmarshal)
}

// LoadJSON parses a profile catalog from JSON bytes.
func LoadJSON(data []byte) (*Catalog, error) {
	return load(data, json.Unmarshal)
}

func load(data []byte, unmarshal func([]byte, any) error) (*Catalog, error) {
	var doc fileShape
	if err := unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	if len(doc.Profiles) == 0 {
		return &Catalog{Profiles: map[string]Profile{}}, nil
	}
	out := make(map[string]Profile, len(doc.Profiles))
	for name, profile := range doc.Profiles {
		key := strings.TrimSpace(name)
		if key == "" {
			return nil, errors.New("profile name must not be empty")
		}
		if _, exists := out[key]; exists {
			return nil, fmt.Errorf("duplicate profile name %q", key)
		}
		profile.Name = key
		out[key] = profile
	}
	return &Catalog{Profiles: out}, nil
}

// ProfileStatus is the validation outcome for one profile.
type ProfileStatus struct {
	Name  string
	Valid bool
	Err   error
}

// ValidationResult aggregates per-profile startup validation.
type ValidationResult struct {
	Statuses []ProfileStatus
}

// AnyValid reports whether at least one profile passed validation.
func (r ValidationResult) AnyValid() bool {
	for _, s := range r.Statuses {
		if s.Valid {
			return true
		}
	}
	return false
}

// AllValid reports whether every profile passed validation.
func (r ValidationResult) AllValid() bool {
	if len(r.Statuses) == 0 {
		return true
	}
	for _, s := range r.Statuses {
		if !s.Valid {
			return false
		}
	}
	return true
}

// Validate checks every profile without resolving secret values.
func (c *Catalog) Validate() ValidationResult {
	if c == nil || len(c.Profiles) == 0 {
		return ValidationResult{}
	}
	statuses := make([]ProfileStatus, 0, len(c.Profiles))
	for name, profile := range c.Profiles {
		err := validateProfile(profile)
		statuses = append(statuses, ProfileStatus{
			Name:  name,
			Valid: err == nil,
			Err:   err,
		})
	}
	return ValidationResult{Statuses: statuses}
}

func validateProfile(p Profile) error {
	if strings.TrimSpace(p.QueueManager) == "" {
		return errors.New("queueManager is required")
	}
	endpoint := strings.TrimSpace(p.Endpoint)
	if endpoint == "" {
		return errors.New("endpoint is required")
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("endpoint must be an https URL, got %q", p.Endpoint)
	}
	if p.Timeout != "" {
		if _, err := time.ParseDuration(p.Timeout); err != nil {
			return fmt.Errorf("timeout: %w", err)
		}
	}
	if err := validateAuthentication(p.Authentication); err != nil {
		return err
	}
	return p.TLS.Validate()
}

func validateAuthentication(auth Authentication) error {
	if auth.Username != "" || auth.Password != "" {
		return errors.New("inline username/password are not allowed; use secretRef")
	}
	switch auth.Type {
	case AuthBasic:
		if strings.TrimSpace(auth.SecretRef) == "" {
			return errors.New("basic authentication requires secretRef")
		}
		if _, err := secrets.Parse(auth.SecretRef); err != nil {
			return fmt.Errorf("authentication.secretRef: %w", err)
		}
	case AuthMTLS:
		for field, ref := range map[string]string{
			"certificateRef": auth.CertificateRef,
			"privateKeyRef":  auth.PrivateKeyRef,
		} {
			if strings.TrimSpace(ref) == "" {
				return fmt.Errorf("mtls authentication requires %s", field)
			}
			parsed, err := secrets.Parse(ref)
			if err != nil {
				return fmt.Errorf("authentication.%s: %w", field, err)
			}
			if parsed.Provider != secrets.ProviderFile {
				return fmt.Errorf("authentication.%s must use file: prefix", field)
			}
		}
		if auth.PassphraseRef != "" {
			if _, err := secrets.Parse(auth.PassphraseRef); err != nil {
				return fmt.Errorf("authentication.passphraseRef: %w", err)
			}
		}
	case "":
		return errors.New("authentication.type is required")
	default:
		return fmt.Errorf("unsupported authentication.type %q", auth.Type)
	}
	return nil
}

// InvalidProfiles returns names that failed validation.
func (r ValidationResult) InvalidProfiles() []string {
	var names []string
	for _, s := range r.Statuses {
		if !s.Valid {
			names = append(names, s.Name)
		}
	}
	return names
}

// ValidProfiles returns names that passed validation.
func (r ValidationResult) ValidProfiles() []string {
	var names []string
	for _, s := range r.Statuses {
		if s.Valid {
			names = append(names, s.Name)
		}
	}
	return names
}

// ProfileByName returns a profile and whether it is known.
func (c *Catalog) ProfileByName(name string) (Profile, bool) {
	if c == nil {
		return Profile{}, false
	}
	p, ok := c.Profiles[name]
	return p, ok
}

// IsValid reports whether the named profile passed startup validation.
func (r ValidationResult) IsValid(name string) bool {
	for _, s := range r.Statuses {
		if s.Name == name {
			return s.Valid
		}
	}
	return false
}

func checkDuplicateYAMLProfileKeys(data []byte) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse catalog: %w", err)
	}
	profilesNode := findMappingValue(root, "profiles")
	if profilesNode == nil || profilesNode.Kind != yaml.MappingNode {
		return nil
	}
	seen := make(map[string]struct{})
	for i := 0; i < len(profilesNode.Content); i += 2 {
		keyNode := profilesNode.Content[i]
		name := strings.TrimSpace(keyNode.Value)
		if name == "" {
			return errors.New("profile name must not be empty")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate profile name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func findMappingValue(root yaml.Node, key string) *yaml.Node {
	if len(root.Content) == 0 {
		return nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key {
			return doc.Content[i+1]
		}
	}
	return nil
}
