package application

import (
	"context"
	"fmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// Producer orchestrates MSG-002 validated put with policy-before-I/O ordering.
type Producer struct {
	pool *ProfilePool
}

// NewProducer constructs a producer over a profile pool.
func NewProducer(pool *ProfilePool) *Producer {
	return &Producer{pool: pool}
}

// PutQueueMessage validates payload locally, authorizes produce, and puts one message.
func (p *Producer) PutQueueMessage(
	ctx context.Context,
	profileName, queueName string,
	req messaging.PutRequest,
) (messaging.PutResult, error) {
	if queueName == "" {
		return messaging.PutResult{}, fmt.Errorf("queue name is required")
	}
	if _, _, err := messaging.PreparePutPayload(req.ContentType, req.Payload); err != nil {
		return messaging.PutResult{}, err
	}
	client, err := p.authorizedMessaging(profileName)
	if err != nil {
		return messaging.PutResult{}, err
	}
	return client.PutMessage(ctx, queueName, req)
}

func (p *Producer) authorizedMessaging(profileName string) (messaging.Client, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("profile pool is not configured")
	}
	profile, err := p.pool.requireProfile(profileName)
	if err != nil {
		return nil, err
	}
	if authErr := p.pool.gate.Authorize(profile, policy.Produce, "put_queue_message"); authErr != nil {
		return nil, authErr
	}
	return p.pool.messagingClient(profileName)
}
