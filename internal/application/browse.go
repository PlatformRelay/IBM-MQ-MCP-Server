package application

import (
	"context"
	"fmt"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// Browser orchestrates MSG-001 non-destructive browse with policy-before-I/O ordering.
type Browser struct {
	pool *ProfilePool
}

// NewBrowser constructs a browser over a profile pool.
func NewBrowser(pool *ProfilePool) *Browser {
	return &Browser{pool: pool}
}

// BrowseQueueMessages returns bounded message metadata (and optional payloads) after browse authorization.
func (b *Browser) BrowseQueueMessages(
	ctx context.Context,
	profileName, queueName string,
	req messaging.BrowseRequest,
) (page collection.Page[messaging.MessageRecord], err error) {
	ctx = audit.EnsureCorrelationID(ctx)
	start := time.Now()
	target := audit.Target{Kind: "queue", Name: queueName}
	defer func() {
		recordSensitiveOperation(ctx, b.pool, profileName, "browse_queue_messages", target, start, err)
	}()

	if err = messaging.ValidateBrowseCount(req.Count); err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	if err = messaging.ValidateMaxPayloadBytes(req.MaxPayloadBytes); err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	req.Count = messaging.NormalizeBrowseCount(req.Count)
	if req.IncludePayload {
		req.MaxPayloadBytes = messaging.NormalizeMaxPayloadBytes(req.MaxPayloadBytes)
	}
	if queueName == "" {
		return collection.Page[messaging.MessageRecord]{}, fmt.Errorf("queue name is required")
	}
	var client messaging.Client
	client, err = b.authorizedMessaging(ctx, profileName)
	if err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	page, err = client.BrowseMessages(ctx, queueName, req)
	if err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	page.Limit = req.Count
	if len(page.Items) > req.Count {
		page.Items = page.Items[:req.Count]
		page.Truncated = true
		page.TruncationReason = collection.TruncationLimitReached
	}
	return page, nil
}

func (b *Browser) authorizedMessaging(ctx context.Context, profileName string) (messaging.Client, error) {
	if b.pool == nil {
		return nil, fmt.Errorf("profile pool is not configured")
	}
	profile, err := b.pool.requireProfile(profileName)
	if err != nil {
		return nil, err
	}
	if authErr := b.pool.gate.Authorize(ctx, profile, policy.Browse, "browse_queue_messages"); authErr != nil {
		return nil, authErr
	}
	return b.pool.messagingClient(profileName)
}
