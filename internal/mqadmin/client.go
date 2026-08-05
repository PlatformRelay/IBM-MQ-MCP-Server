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

	ListChannels(ctx context.Context, req ListChannelsRequest) (collection.Page[ChannelSummary], error)
	GetChannel(ctx context.Context, name string) (ChannelDetail, error)
	GetChannelStatus(ctx context.Context, name string) (ChannelStatus, error)

	ListListeners(ctx context.Context, req ListListenersRequest) (collection.Page[ListenerSummary], error)
	GetListener(ctx context.Context, name string) (ListenerDetail, error)
	GetListenerStatus(ctx context.Context, name string) (ListenerStatus, error)

	ListSubscriptions(
		ctx context.Context,
		req ListSubscriptionsRequest,
	) (collection.Page[SubscriptionSummary], error)
	GetSubscription(ctx context.Context, id string) (SubscriptionDetail, error)
}
