package application

import (
	"context"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/coexistence"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
)

// DefineChannel validates input, authorizes administer, runs the pre-mutation hook, then creates the channel.
func (a *Administrator) DefineChannel(
	ctx context.Context,
	profileName, channelName string,
	req mqadmin.DefineChannelRequest,
) (mqadmin.ChannelMutationResult, error) {
	if err := mqadmin.ValidateDefineChannelRequest(channelName, req); err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "define_channel")
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	target := coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: profile.QueueManager,
		Kind:         coexistence.ObjectChannel,
		Name:         channelName,
	}
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, target, nil)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result, err := client.DefineChannel(ctx, channelName, req)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// AlterChannel validates input, authorizes administer, runs the pre-mutation hook, then alters the channel.
func (a *Administrator) AlterChannel(
	ctx context.Context,
	profileName, channelName string,
	req mqadmin.AlterChannelRequest,
) (mqadmin.ChannelMutationResult, error) {
	if err := mqadmin.ValidateAlterChannelRequest(channelName, req); err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "alter_channel")
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	target := coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: profile.QueueManager,
		Kind:         coexistence.ObjectChannel,
		Name:         channelName,
	}
	warning, err := a.runPreMutationHook(ctx, client, hook, profile, target, channelTagFetcher(channelName))
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result, err := client.AlterChannel(ctx, channelName, req)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DeleteChannel validates input, authorizes administer, runs the pre-mutation hook, then deletes the channel.
func (a *Administrator) DeleteChannel(
	ctx context.Context,
	profileName, channelName string,
) (mqadmin.ChannelMutationResult, error) {
	if err := mqadmin.ValidateDeleteChannelRequest(channelName); err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "delete_channel")
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	target := coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: profile.QueueManager,
		Kind:         coexistence.ObjectChannel,
		Name:         channelName,
	}
	warning, err := a.runPreMutationHook(ctx, client, hook, profile, target, channelTagFetcher(channelName))
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result, err := client.DeleteChannel(ctx, channelName)
	if err != nil {
		return mqadmin.ChannelMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DefineCHLAUTH validates input, authorizes administer, runs the pre-mutation hook, then creates the rule.
func (a *Administrator) DefineCHLAUTH(
	ctx context.Context,
	profileName string,
	req mqadmin.DefineCHLAUTHRequest,
) (mqadmin.CHLAUTHMutationResult, error) {
	if err := mqadmin.ValidateDefineCHLAUTHRequest(req); err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "define_chlauth")
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	target := chlauthMutationTarget(profileName, profile.QueueManager, req.Target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, target, nil)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result, err := client.DefineCHLAUTH(ctx, req)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// AlterCHLAUTH validates input, authorizes administer, runs the pre-mutation hook, then alters the rule.
func (a *Administrator) AlterCHLAUTH(
	ctx context.Context,
	profileName string,
	req mqadmin.AlterCHLAUTHRequest,
) (mqadmin.CHLAUTHMutationResult, error) {
	if err := mqadmin.ValidateAlterCHLAUTHRequest(req); err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "alter_chlauth")
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	target := chlauthMutationTarget(profileName, profile.QueueManager, req.Target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, target, nil)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result, err := client.AlterCHLAUTH(ctx, req)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DeleteCHLAUTH validates input, authorizes administer, runs the pre-mutation hook, then deletes the rule.
func (a *Administrator) DeleteCHLAUTH(
	ctx context.Context,
	profileName string,
	target mqadmin.CHLAUTHTarget,
) (mqadmin.CHLAUTHMutationResult, error) {
	if err := mqadmin.ValidateDeleteCHLAUTHRequest(target); err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "delete_chlauth")
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	mutationTarget := chlauthMutationTarget(profileName, profile.QueueManager, target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, mutationTarget, nil)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result, err := client.DeleteCHLAUTH(ctx, target)
	if err != nil {
		return mqadmin.CHLAUTHMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DefineAuthrec validates input, authorizes administer, runs the pre-mutation hook, then creates the record.
func (a *Administrator) DefineAuthrec(
	ctx context.Context,
	profileName string,
	req mqadmin.DefineAuthrecRequest,
) (mqadmin.AuthrecMutationResult, error) {
	if err := mqadmin.ValidateDefineAuthrecRequest(req); err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "define_authrec")
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	target := authrecMutationTarget(profileName, profile.QueueManager, req.Target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, target, nil)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result, err := client.DefineAuthrec(ctx, req)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// AlterAuthrec validates input, authorizes administer, runs the pre-mutation hook, then alters the record.
func (a *Administrator) AlterAuthrec(
	ctx context.Context,
	profileName string,
	req mqadmin.AlterAuthrecRequest,
) (mqadmin.AuthrecMutationResult, error) {
	if err := mqadmin.ValidateAlterAuthrecRequest(req); err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "alter_authrec")
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	target := authrecMutationTarget(profileName, profile.QueueManager, req.Target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, target, nil)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result, err := client.AlterAuthrec(ctx, req)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

// DeleteAuthrec validates input, authorizes administer, runs the pre-mutation hook, then deletes the record.
func (a *Administrator) DeleteAuthrec(
	ctx context.Context,
	profileName string,
	target mqadmin.AuthrecTarget,
) (mqadmin.AuthrecMutationResult, error) {
	if err := mqadmin.ValidateDeleteAuthrecRequest(target); err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	client, profile, hook, err := a.authorizedMutation(profileName, "delete_authrec")
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	mutationTarget := authrecMutationTarget(profileName, profile.QueueManager, target)
	warning, err := a.runPreMutationHook(ctx, nil, hook, profile, mutationTarget, nil)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result, err := client.DeleteAuthrec(ctx, target)
	if err != nil {
		return mqadmin.AuthrecMutationResult{}, err
	}
	result.Warning = warning
	result.CompletedAt = time.Now().UTC()
	return result, nil
}

func chlauthMutationTarget(profileName, queueManager string, target mqadmin.CHLAUTHTarget) coexistence.MutationTarget {
	return coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: queueManager,
		Kind:         coexistence.ObjectCHLAUTH,
		Name:         target.ChannelName,
	}
}

func authrecMutationTarget(profileName, queueManager string, target mqadmin.AuthrecTarget) coexistence.MutationTarget {
	return coexistence.MutationTarget{
		Profile:      profileName,
		QueueManager: queueManager,
		Kind:         coexistence.ObjectAuthrec,
		Name:         target.Profile,
	}
}

func channelTagFetcher(channelName string) tagFetcher {
	return func(ctx context.Context, client mqadmin.Client) (map[string]string, error) {
		detail, err := client.GetChannel(ctx, channelName)
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
}

type tagFetcher func(ctx context.Context, client mqadmin.Client) (map[string]string, error)

func (a *Administrator) runPreMutationHook(
	ctx context.Context,
	client mqadmin.Client,
	hook *coexistence.PreMutationHook,
	profile catalog.Profile,
	target coexistence.MutationTarget,
	tags tagFetcher,
) (string, error) {
	catalogResult := hook.Evaluate(target, nil)
	if err := hook.Enforce(catalogResult); err != nil {
		return "", err
	}
	warning := hookWarning(catalogResult)

	if client == nil || profile.MKurator.MutationPolicy == coexistence.PolicyBlock || tags == nil {
		return warning, nil
	}

	objectTags, err := tags(ctx, client)
	if err != nil {
		return "", err
	}
	if len(objectTags) == 0 {
		return warning, nil
	}

	tagResult := hook.Evaluate(target, objectTags)
	if err := hook.Enforce(tagResult); err != nil {
		return "", err
	}
	if tagResult.Outcome == coexistence.OutcomeWarn {
		return tagResult.Message, nil
	}
	return warning, nil
}
