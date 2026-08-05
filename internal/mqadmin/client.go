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

	DefineQueue(ctx context.Context, name string, req DefineQueueRequest) (QueueMutationResult, error)
	AlterQueue(ctx context.Context, name string, req AlterQueueRequest) (QueueMutationResult, error)
	DeleteQueue(ctx context.Context, name string) (QueueMutationResult, error)

	DefineChannel(ctx context.Context, name string, req DefineChannelRequest) (ChannelMutationResult, error)
	AlterChannel(ctx context.Context, name string, req AlterChannelRequest) (ChannelMutationResult, error)
	DeleteChannel(ctx context.Context, name string) (ChannelMutationResult, error)

	DefineCHLAUTH(ctx context.Context, req DefineCHLAUTHRequest) (CHLAUTHMutationResult, error)
	AlterCHLAUTH(ctx context.Context, req AlterCHLAUTHRequest) (CHLAUTHMutationResult, error)
	DeleteCHLAUTH(ctx context.Context, target CHLAUTHTarget) (CHLAUTHMutationResult, error)

	DefineAuthrec(ctx context.Context, req DefineAuthrecRequest) (AuthrecMutationResult, error)
	AlterAuthrec(ctx context.Context, req AlterAuthrecRequest) (AuthrecMutationResult, error)
	DeleteAuthrec(ctx context.Context, target AuthrecTarget) (AuthrecMutationResult, error)
}
