// Package secrets resolves env: and file: secret references without logging values.
package secrets

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Provider names a supported secret reference scheme (ADR-0004).
type Provider string

const (
	// ProviderEnv resolves secrets from environment variables.
	ProviderEnv Provider = "env"
	// ProviderFile resolves secrets from mounted files.
	ProviderFile Provider = "file"
)

// Reference is a parsed secret reference safe to log.
type Reference struct {
	Provider Provider
	Name     string
}

// Parse validates and parses env:NAME or file:PATH references.
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
	default:
		return Reference{}, fmt.Errorf("unsupported secret reference %q: use env: or file", redactForError(ref))
	}
}

// String returns a log-safe representation of the reference.
func (r Reference) String() string {
	return fmt.Sprintf("%s:%s", r.Provider, r.Name)
}

// Resolver reads secret values lazily from the environment or filesystem.
type Resolver struct{}

// NewResolver returns a secret resolver backed by the process environment.
func NewResolver() *Resolver {
	return &Resolver{}
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
