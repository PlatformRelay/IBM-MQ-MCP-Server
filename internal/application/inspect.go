package application

import (
	"context"
	"fmt"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// AdminClientFactory constructs mqadmin clients after policy grants secrets resolution.
type AdminClientFactory func(profile catalog.Profile, resolver *secrets.Resolver) (mqadmin.Client, error)

// MessagingClientFactory constructs messaging clients after policy grants secrets resolution.
type MessagingClientFactory func(profile catalog.Profile, resolver *secrets.Resolver) (messaging.Client, error)

// ProfileSummary is safe profile metadata for discovery tools (no secrets).
type ProfileSummary struct {
	Name         string   `json:"name"`
	QueueManager string   `json:"queueManager"`
	Endpoint     string   `json:"endpoint"`
	Capabilities []string `json:"capabilities"`
	Valid        bool     `json:"valid"`
}

// Inspector orchestrates INS-001 read paths with policy-before-I/O ordering.
type Inspector struct {
	pool *ProfilePool
}

// NewInspector constructs an inspector over a profile pool.
func NewInspector(pool *ProfilePool) *Inspector {
	return &Inspector{pool: pool}
}

// Pool returns the underlying profile pool (MCP wiring).
func (i *Inspector) Pool() *ProfilePool {
	return i.pool
}

// ListProfiles returns configured identity and capabilities without secret resolution.
func (i *Inspector) ListProfiles() []ProfileSummary {
	if i.pool == nil || i.pool.catalog == nil {
		return nil
	}
	out := make([]ProfileSummary, 0, len(i.pool.catalog.Profiles))
	for name, profile := range i.pool.catalog.Profiles {
		out = append(out, ProfileSummary{
			Name:         name,
			QueueManager: profile.QueueManager,
			Endpoint:     profile.Endpoint,
			Capabilities: append([]string(nil), profile.Capabilities...),
			Valid:        i.pool.validation.IsValid(name),
		})
	}
	return out
}

// QueueManagerStatus checks live queue manager identity after inspect authorization.
func (i *Inspector) QueueManagerStatus(ctx context.Context, profileName string) (mqadmin.QueueManagerStatus, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "queue_manager_status")
	if err != nil {
		return mqadmin.QueueManagerStatus{}, err
	}
	profile, _ := i.pool.requireProfile(profileName)
	return client.QueueManagerStatus(ctx, profile.QueueManager)
}

// ListQueues returns a bounded queue page after inspect authorization.
func (i *Inspector) ListQueues(
	ctx context.Context,
	profileName string,
	req mqadmin.ListQueuesRequest,
) (collection.Page[mqadmin.QueueSummary], error) {
	if err := collection.ValidateLimit(req.Limit); err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	req.Limit = collection.NormalizeLimit(req.Limit)
	client, err := i.authorizedAdmin(ctx, profileName, "list_queues")
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	return client.ListQueues(ctx, req)
}

// GetQueue returns queue definition and status after inspect authorization.
func (i *Inspector) GetQueue(ctx context.Context, profileName, queueName string) (mqadmin.QueueDetail, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_queue")
	if err != nil {
		return mqadmin.QueueDetail{}, err
	}
	return client.GetQueue(ctx, queueName)
}

// ListChannels returns a bounded channel page after inspect authorization.
func (i *Inspector) ListChannels(
	ctx context.Context,
	profileName string,
	req mqadmin.ListChannelsRequest,
) (collection.Page[mqadmin.ChannelSummary], error) {
	if err := collection.ValidateLimit(req.Limit); err != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, err
	}
	req.Limit = collection.NormalizeLimit(req.Limit)
	client, err := i.authorizedAdmin(ctx, profileName, "list_channels")
	if err != nil {
		return collection.Page[mqadmin.ChannelSummary]{}, err
	}
	return client.ListChannels(ctx, req)
}

// GetChannel returns channel definition after inspect authorization.
func (i *Inspector) GetChannel(ctx context.Context, profileName, channelName string) (mqadmin.ChannelDetail, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_channel")
	if err != nil {
		return mqadmin.ChannelDetail{}, err
	}
	return client.GetChannel(ctx, channelName)
}

// GetChannelStatus returns runtime channel status after inspect authorization.
func (i *Inspector) GetChannelStatus(
	ctx context.Context,
	profileName, channelName string,
) (mqadmin.ChannelStatus, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_channel_status")
	if err != nil {
		return mqadmin.ChannelStatus{}, err
	}
	return client.GetChannelStatus(ctx, channelName)
}

// ListListeners returns a bounded listener page after inspect authorization.
func (i *Inspector) ListListeners(
	ctx context.Context,
	profileName string,
	req mqadmin.ListListenersRequest,
) (collection.Page[mqadmin.ListenerSummary], error) {
	if err := collection.ValidateLimit(req.Limit); err != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, err
	}
	req.Limit = collection.NormalizeLimit(req.Limit)
	client, err := i.authorizedAdmin(ctx, profileName, "list_listeners")
	if err != nil {
		return collection.Page[mqadmin.ListenerSummary]{}, err
	}
	return client.ListListeners(ctx, req)
}

// GetListener returns listener definition after inspect authorization.
func (i *Inspector) GetListener(ctx context.Context, profileName, listenerName string) (mqadmin.ListenerDetail, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_listener")
	if err != nil {
		return mqadmin.ListenerDetail{}, err
	}
	return client.GetListener(ctx, listenerName)
}

// GetListenerStatus returns runtime listener status after inspect authorization.
func (i *Inspector) GetListenerStatus(
	ctx context.Context,
	profileName, listenerName string,
) (mqadmin.ListenerStatus, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_listener_status")
	if err != nil {
		return mqadmin.ListenerStatus{}, err
	}
	return client.GetListenerStatus(ctx, listenerName)
}

// ListSubscriptions returns a bounded subscription page after inspect authorization.
func (i *Inspector) ListSubscriptions(
	ctx context.Context,
	profileName string,
	req mqadmin.ListSubscriptionsRequest,
) (collection.Page[mqadmin.SubscriptionSummary], error) {
	if err := collection.ValidateLimit(req.Limit); err != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, err
	}
	req.Limit = collection.NormalizeLimit(req.Limit)
	client, err := i.authorizedAdmin(ctx, profileName, "list_subscriptions")
	if err != nil {
		return collection.Page[mqadmin.SubscriptionSummary]{}, err
	}
	return client.ListSubscriptions(ctx, req)
}

// GetSubscription returns subscription definition after inspect authorization.
func (i *Inspector) GetSubscription(
	ctx context.Context,
	profileName, subscriptionID string,
) (mqadmin.SubscriptionDetail, error) {
	client, err := i.authorizedAdmin(ctx, profileName, "get_subscription")
	if err != nil {
		return mqadmin.SubscriptionDetail{}, err
	}
	return client.GetSubscription(ctx, subscriptionID)
}

// CheckProfileConnectivity verifies mqweb reachability and queue manager identity without mutation.
func (i *Inspector) CheckProfileConnectivity(
	ctx context.Context,
	profileName string,
) (mqadmin.ConnectivityReport, error) {
	profile, err := i.pool.requireProfile(profileName)
	if err != nil {
		return mqadmin.ConnectivityReport{}, err
	}
	if authErr := i.pool.gate.Authorize(ctx, profile, policy.Inspect, "check_profile_connectivity"); authErr != nil {
		return mqadmin.ConnectivityReport{}, authErr
	}
	start := time.Now()
	client, err := i.pool.adminClient(profileName)
	if err != nil {
		return mqadmin.BuildConnectivityReport(
			profileName, profile.Endpoint, profile.QueueManager, start, mqadmin.QueueManagerStatus{}, err,
		), nil
	}
	status, qmErr := client.QueueManagerStatus(ctx, profile.QueueManager)
	return mqadmin.BuildConnectivityReport(
		profileName, profile.Endpoint, profile.QueueManager, start, status, qmErr,
	), nil
}

// ListProfilesPage wraps ListProfiles in the shared collection envelope.
func (i *Inspector) ListProfilesPage(limit int, cursor string) (collection.Page[ProfileSummary], error) {
	if err := collection.ValidateLimit(limit); err != nil {
		return collection.Page[ProfileSummary]{}, err
	}
	limit = collection.NormalizeLimit(limit)
	all := i.ListProfiles()
	start, err := parseOffsetCursor(cursor)
	if err != nil {
		return collection.Page[ProfileSummary]{}, err
	}
	end := start + limit
	truncated := end < len(all)
	if end > len(all) {
		end = len(all)
	}
	page := collection.Page[ProfileSummary]{
		Items:     all[start:end],
		Limit:     limit,
		Truncated: truncated,
	}
	if start > 0 {
		page.Cursor = fmt.Sprintf("%d", start)
	}
	if truncated {
		page.NextCursor = fmt.Sprintf("%d", end)
		page.TruncationReason = collection.TruncationLimitReached
	}
	return page, nil
}

func (i *Inspector) authorizedAdmin(ctx context.Context, profileName, operation string) (mqadmin.Client, error) {
	profile, err := i.pool.requireProfile(profileName)
	if err != nil {
		return nil, err
	}
	if authErr := i.pool.gate.Authorize(ctx, profile, policy.Inspect, operation); authErr != nil {
		return nil, authErr
	}
	return i.pool.adminClient(profileName)
}

func parseOffsetCursor(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	var offset int
	if _, err := fmt.Sscanf(raw, "%d", &offset); err != nil || offset < 0 {
		return 0, fmt.Errorf("invalid cursor %q", raw)
	}
	return offset, nil
}
