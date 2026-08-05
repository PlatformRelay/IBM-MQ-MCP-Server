package mqadmin

import "context"

// Client is the typed administration port for one profile (ADR-0002).
type Client interface {
	ProfileName() string
	Ping(ctx context.Context) error
	Close() error
}
