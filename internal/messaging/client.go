package messaging

import "context"

// Client is the typed messaging port for one profile (ADR-0002).
type Client interface {
	ProfileName() string
	Ping(ctx context.Context) error
	Close() error
}
