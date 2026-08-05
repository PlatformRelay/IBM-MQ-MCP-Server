package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
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
) (mqadmin.RawMQSCResult, error) {
	if err := mqadmin.ValidateRawMQSCCommand(command); err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	client, profile, err := e.authorized(profileName)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	e.audit(profileName, command)
	result, err := client.ExecuteRawMQSC(ctx, command)
	if err != nil {
		return mqadmin.RawMQSCResult{}, err
	}
	result.Profile = profileName
	result.QueueManager = profile.QueueManager
	result.Command = mqadmin.RedactMQSCCommandText(command)
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

func (e *MQSCExecutor) authorized(profileName string) (mqadmin.Client, catalog.Profile, error) {
	if e.pool == nil {
		return nil, catalog.Profile{}, fmt.Errorf("profile pool is not configured")
	}
	profile, err := e.pool.requireProfile(profileName)
	if err != nil {
		return nil, catalog.Profile{}, err
	}
	if authErr := e.pool.gate.Authorize(profile, policy.ExecuteMQSC, "execute_mqsc"); authErr != nil {
		return nil, catalog.Profile{}, authErr
	}
	client, err := e.pool.adminClient(profileName)
	if err != nil {
		return nil, catalog.Profile{}, err
	}
	return client, profile, nil
}

func (e *MQSCExecutor) audit(profileName, command string) {
	slog.Info("mqsc execution",
		slog.String("profile", profileName),
		slog.String("operation", "execute_mqsc"),
		slog.String("command", mqadmin.RedactMQSCCommandText(command)),
	)
}
