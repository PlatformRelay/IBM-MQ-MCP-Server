package application

import (
	"context"
	"fmt"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// Administrator orchestrates ADM-001 typed queue mutations with policy and INT-001 hooks.
type Administrator struct {
	pool *ProfilePool
}

// NewAdministrator constructs an administrator over a profile pool.
func NewAdministrator(pool *ProfilePool) *Administrator {
	return &Administrator{pool: pool}
}

// DefineQueue validates input, authorizes administer, runs the pre-mutation hook, then creates the queue.
func (a *Administrator) DefineQueue(
	ctx context.Context,
	profileName, queueName string,
	req mqadmin.DefineQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	if err := mqadmin.ValidateDefineQueueRequest(queueName, req); err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, queueName, "define_queue")
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	warning, err := a.runPreMutationHook(hook, profile, profileName, queueName, nil)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result, err := client.DefineQueue(ctx, queueName, req)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// AlterQueue validates input, authorizes administer, runs the pre-mutation hook, then alters the queue.
func (a *Administrator) AlterQueue(
	ctx context.Context,
	profileName, queueName string,
	req mqadmin.AlterQueueRequest,
) (mqadmin.QueueMutationResult, error) {
	if err := mqadmin.ValidateAlterQueueRequest(queueName, req); err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, queueName, "alter_queue")
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	tags, tagErr := a.objectTags(ctx, client, queueName)
	if tagErr != nil {
		return mqadmin.QueueMutationResult{}, tagErr
	}
	warning, err := a.runPreMutationHook(hook, profile, profileName, queueName, tags)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result, err := client.AlterQueue(ctx, queueName, req)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DeleteQueue validates input, authorizes administer, runs the pre-mutation hook, then deletes the queue.
func (a *Administrator) DeleteQueue(
	ctx context.Context,
	profileName, queueName string,
) (mqadmin.QueueMutationResult, error) {
	if err := mqadmin.ValidateDeleteQueueRequest(queueName); err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, queueName, "delete_queue")
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	tags, tagErr := a.objectTags(ctx, client, queueName)
	if tagErr != nil {
		return mqadmin.QueueMutationResult{}, tagErr
	}
	warning, err := a.runPreMutationHook(hook, profile, profileName, queueName, tags)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result, err := client.DeleteQueue(ctx, queueName)
	if err != nil {
		return mqadmin.QueueMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

func (a *Administrator) authorizedMutation(
	profileName, _ string, operation string,
) (mqadmin.Client, catalog.Profile, *coexistence.PreMutationHook, error) {
	if a.pool == nil {
		return nil, catalog.Profile{}, nil, fmt.Errorf("profile pool is not configured")
	}
	profile, err := a.pool.requireProfile(profileName)
	if err != nil {
		return nil, catalog.Profile{}, nil, err
	}
	if authErr := a.pool.gate.Authorize(profile, policy.Administer, operation); authErr != nil {
		return nil, catalog.Profile{}, nil, authErr
	}
	client, err := a.pool.adminClient(profileName)
	if err != nil {
		return nil, catalog.Profile{}, nil, err
	}
	hook := coexistence.NewPreMutationHook(profile.MKurator)
	return client, profile, hook, nil
}

func (a *Administrator) runPreMutationHook(
	hook *coexistence.PreMutationHook,
	profile catalog.Profile,
	profileName, queueName string,
	tags map[string]string,
) (string, error) {
	result := hook.Evaluate(coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: profile.QueueManager,
		Kind:         coexistence.ObjectQueue,
		Name:         queueName,
	}, tags)
	if err := hook.Enforce(result); err != nil {
		return "", err
	}
	if result.Outcome == coexistence.OutcomeWarn {
		return result.Message, nil
	}
	return "", nil
}

func (a *Administrator) objectTags(
	ctx context.Context,
	client mqadmin.Client,
	queueName string,
) (map[string]string, error) {
	detail, err := client.GetQueue(ctx, queueName)
	if err != nil {
		if reason, ok := mqadmin.AsReasonError(err); ok && reason.Code == 2085 {
			return nil, nil
		}
		return nil, err
	}
	if detail.MKuratorTag == "" {
		return nil, nil
	}
	return map[string]string{coexistence.TagManagedByMKurator: detail.MKuratorTag}, nil
}
