package application

import (
	"context"
	"fmt"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// MQSCExecutor runs exceptional raw MQSC commands behind double opt-in.
type MQSCExecutor struct {
	pool *ProfilePool
}

// NewMQSCExecutor constructs an executor over a profile pool.
func NewMQSCExecutor(pool *ProfilePool) *MQSCExecutor {
	return &MQSCExecutor{pool: pool}
}

// ExecuteRawMQSC validates the verb allowlist, authorizes execute_mqsc, audits, then executes.
func (e *MQSCExecutor) ExecuteRawMQSC(
	ctx context.Context,
	profileName, command string,
) (result mqadmin.RawMQSCResult, err error) {
	ctx = audit.EnsureCorrelationID(ctx)
	start := time.Now()
	redacted := mqadmin.RedactMQSCCommandText(command)
	defer func() {
		recordSensitiveOperation(
			ctx, e.pool, profileName, "execute_mqsc",
			audit.Target{Kind: "mqsc", Name: "raw"},
			start, err, withCommandRedacted(redacted),
		)
	}()

	if err = mqadmin.ValidateRawMQSCCommand(command); err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	var client mqadmin.Client
	var profile catalog.Profile
	client, profile, err = e.authorized(ctx, profileName)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	result, err = client.ExecuteRawMQSC(ctx, command)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	result.Profile = profileName
	result.QueueManager = profile.QueueManager
	result.Command = redacted
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

func (e *MQSCExecutor) authorized(ctx context.Context, profileName string) (mqadmin.Client, catalog.Profile, error) {
	if e.pool == nil {
		return nil, catalog.Profile{}, fmt.Errorf("profile pool is not configured")
	}
	profile, err := e.pool.requireProfile(profileName)
	if err != nil {
		return nil, catalog.Profile{}, err
	}
	if authErr := e.pool.gate.Authorize(ctx, profile, policy.ExecuteMQSC, "execute_mqsc"); authErr != nil {
		return nil, catalog.Profile{}, authErr
	}
	client, err := e.pool.adminClient(profileName)
	if err != nil {
		return nil, catalog.Profile{}, err
	}
	return client, profile, nil
}
