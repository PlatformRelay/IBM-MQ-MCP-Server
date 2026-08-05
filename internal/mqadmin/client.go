package mqadmin

import (
	"context"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
)

// Client is the typed administration port for one profile (ADR-0002).
type Client interface {
	ProfileName() string
	Ping(ctx context.Context) error
	Close() error

	QueueManagerStatus(ctx context.Context, configuredName string) (QueueManagerStatus, error)
	ListQueues(ctx context.Context, req ListQueuesRequest) (collection.Page[QueueSummary], error)
	GetQueue(ctx context.Context, name string) (QueueDetail, error)
}
