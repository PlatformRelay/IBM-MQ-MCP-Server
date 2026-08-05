// Package secrets resolves env:, file:, and k8s: secret references without logging values.
package secrets

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Provider names a supported secret reference scheme (ADR-0004).
type Provider string

const (
	// ProviderEnv resolves secrets from environment variables.
	ProviderEnv Provider = "env"
	// ProviderFile resolves secrets from mounted files.
	ProviderFile Provider = "file"
	// ProviderK8s resolves secrets from Kubernetes Secrets.
	ProviderK8s Provider = "k8s"
)

// Reference is a parsed secret reference safe to log.
type Reference struct {
	Provider Provider
	Name     string // env var or file path
	// K8s fields are set when Provider is ProviderK8s.
	K8sNamespace string
	K8sSecret    string
	K8sKey       string
}

// Parse validates and parses env:, file:, or k8s: references.
func Parse(ref string) (Reference, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return Reference{}, errors.New("empty secret reference")
	}
	switch {
	case strings.HasPrefix(ref, "env:"):
		name := strings.TrimSpace(strings.TrimPrefix(ref, "env:"))
		if name == "" {
			return Reference{}, errors.New("env reference requires a variable name")
		}
		return Reference{Provider: ProviderEnv, Name: name}, nil
	case strings.HasPrefix(ref, "file:"):
		path := strings.TrimSpace(strings.TrimPrefix(ref, "file:"))
		if path == "" {
			return Reference{}, errors.New("file reference requires a path")
		}
		return Reference{Provider: ProviderFile, Name: path}, nil
	case strings.HasPrefix(ref, "k8s:"):
		return parseK8sReference(ref)
	default:
		return Reference{}, fmt.Errorf(
			"unsupported secret reference %q (supported schemes: env, file, k8s)",
			redactForError(ref),
		)
	}
}

// String returns a log-safe representation of the reference.
func (r Reference) String() string {
	if r.Provider == ProviderK8s {
		return fmt.Sprintf("k8s:%s/%s#%s", r.K8sNamespace, r.K8sSecret, r.K8sKey)
	}
	return fmt.Sprintf("%s:%s", r.Provider, r.Name)
}

// ResolverOption configures a secret resolver.
type ResolverOption func(*Resolver)

// WithK8sReader injects a Kubernetes secret reader. A nil reader disables k8s resolution.
func WithK8sReader(reader K8sSecretReader) ResolverOption {
	return func(r *Resolver) {
		r.k8s = reader
		r.k8sConfigured = true
	}
}

// Resolver reads secret values lazily from configured providers.
type Resolver struct {
	k8s           K8sSecretReader
	k8sConfigured bool
	lazyOnce      sync.Once
	lazyReader    K8sSecretReader
	lazyErr       error
}

// NewResolver returns a secret resolver backed by the process environment and optional providers.
func NewResolver(opts ...ResolverOption) *Resolver {
	r := &Resolver{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Resolve returns the secret value for a reference. Callers must not log the result.
func (r *Resolver) Resolve(ref Reference) (string, error) {
	switch ref.Provider {
	case ProviderEnv:
		value, ok := os.LookupEnv(ref.Name)
		if !ok || value == "" {
			return "", fmt.Errorf("environment variable %q is not set", ref.Name)
		}
		return value, nil
	case ProviderFile:
		data, err := os.ReadFile(ref.Name) //nolint:gosec // G304: path from operator config
		if err != nil {
			return "", fmt.Errorf("read secret file %q: %w", ref.Name, err)
		}
		return strings.TrimRight(string(data), "\r\n"), nil
	case ProviderK8s:
		return r.resolveK8s(ref)
	default:
		return "", fmt.Errorf("unsupported provider %q", ref.Provider)
	}
}

func redactForError(value string) string {
	if len(value) <= 8 {
		return "[REDACTED]"
	}
	return value[:4] + "…"
}
