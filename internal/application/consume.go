package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// Consumer orchestrates MSG-003 destructive consume with policy-before-I/O ordering.
type Consumer struct {
	pool *ProfilePool
}

// NewConsumer constructs a consumer over a profile pool.
func NewConsumer(pool *ProfilePool) *Consumer {
	return &Consumer{pool: pool}
}

// ConsumeQueueMessages destructively retrieves bounded messages after consume authorization.
func (c *Consumer) ConsumeQueueMessages(
	ctx context.Context,
	profileName, queueName string,
	req messaging.ConsumeRequest,
) (collection.Page[messaging.MessageRecord], error) {
	if err := messaging.ValidateConsumeCount(req.Count); err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	if err := messaging.ValidateConsumeWaitIntervalMs(req.WaitIntervalMs); err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	if err := messaging.ValidateMaxPayloadBytes(req.MaxPayloadBytes); err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	req.Count = messaging.NormalizeConsumeCount(req.Count)
	if req.IncludePayload {
		req.MaxPayloadBytes = messaging.NormalizeMaxPayloadBytes(req.MaxPayloadBytes)
	}
	if queueName == "" {
		return collection.Page[messaging.MessageRecord]{}, fmt.Errorf("queue name is required")
	}
	client, err := c.authorizedMessaging(profileName)
	if err != nil {
		return collection.Page[messaging.MessageRecord]{}, err
	}
	page, err := client.ConsumeMessages(ctx, queueName, req)
	page.Limit = req.Count
	if len(page.Items) > req.Count {
		page.Items = page.Items[:req.Count]
		page.Truncated = true
		page.TruncationReason = collection.TruncationLimitReached
	}
	if err != nil {
		var partial *messaging.PartialConsumeError
		if errors.As(err, &partial) {
			return page, err
		}
		return collection.Page[messaging.MessageRecord]{}, err
	}
	return page, nil
}

func (c *Consumer) authorizedMessaging(profileName string) (messaging.Client, error) {
	if c.pool == nil {
		return nil, fmt.Errorf("profile pool is not configured")
	}
	profile, err := c.pool.requireProfile(profileName)
	if err != nil {
		return nil, err
	}
	if authErr := c.pool.gate.Authorize(profile, policy.Consume, "consume_queue_messages"); authErr != nil {
		return nil, authErr
	}
	return c.pool.messagingClient(profileName)
}
