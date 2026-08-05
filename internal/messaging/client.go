package messaging

import (
	"context"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
)

// Client is the typed messaging port for one profile (ADR-0002).
type Client interface {
	ProfileName() string
	Ping(ctx context.Context) error
	BrowseMessages(ctx context.Context, queueName string, req BrowseRequest) (collection.Page[MessageRecord], error)
	ConsumeMessages(ctx context.Context, queueName string, req ConsumeRequest) (collection.Page[MessageRecord], error)
	PutMessage(ctx context.Context, queueName string, req PutRequest) (PutResult, error)
	Close() error
}
