package application

import (
	"context"
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// AdminClientFactory constructs mqadmin clients after policy grants secrets resolution.
type AdminClientFactory func(profile catalog.Profile, resolver *secrets.Resolver) (mqadmin.Client, error)

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
	profile, err := i.pool.requireProfile(profileName)
	if err != nil {
		return mqadmin.QueueManagerStatus{}, err
	}
	if authErr := i.pool.gate.Authorize(profile, policy.Inspect, "queue_manager_status"); authErr != nil {
		return mqadmin.QueueManagerStatus{}, authErr
	}
	client, err := i.pool.adminClient(profileName)
	if err != nil {
		return mqadmin.QueueManagerStatus{}, err
	}
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
	profile, err := i.pool.requireProfile(profileName)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	if authErr := i.pool.gate.Authorize(profile, policy.Inspect, "list_queues"); authErr != nil {
		return collection.Page[mqadmin.QueueSummary]{}, authErr
	}
	client, err := i.pool.adminClient(profileName)
	if err != nil {
		return collection.Page[mqadmin.QueueSummary]{}, err
	}
	return client.ListQueues(ctx, req)
}

// GetQueue returns queue definition and status after inspect authorization.
func (i *Inspector) GetQueue(ctx context.Context, profileName, queueName string) (mqadmin.QueueDetail, error) {
	profile, err := i.pool.requireProfile(profileName)
	if err != nil {
		return mqadmin.QueueDetail{}, err
	}
	if authErr := i.pool.gate.Authorize(profile, policy.Inspect, "get_queue"); authErr != nil {
		return mqadmin.QueueDetail{}, authErr
	}
	client, err := i.pool.adminClient(profileName)
	if err != nil {
		return mqadmin.QueueDetail{}, err
	}
	return client.GetQueue(ctx, queueName)
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
