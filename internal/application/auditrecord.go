package application

import (
	"context"
	"errors"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

type operationAuditOption func(*audit.OperationRecord)

func withCommandRedacted(command string) operationAuditOption {
	return func(record *audit.OperationRecord) {
		record.CommandRedacted = command
	}
}

func beginSensitiveOperation(
	ctx context.Context,
	pool *ProfilePool,
	profile, operation string,
	target audit.Target,
) (context.Context, func(error)) {
	ctx = audit.EnsureCorrelationID(ctx)
	start := time.Now()
	return ctx, func(err error) {
		recordSensitiveOperation(ctx, pool, profile, operation, target, start, err)
	}
}

func beginMutationAudit(
	ctx context.Context,
	pool *ProfilePool,
	profile, operation, targetKind, targetName string,
) (context.Context, func(error)) {
	return beginSensitiveOperation(
		ctx, pool, profile, operation, audit.Target{Kind: targetKind, Name: targetName},
	)
}

func recordSensitiveOperation(
	ctx context.Context,
	pool *ProfilePool,
	profile, operation string,
	target audit.Target,
	start time.Time,
	err error,
	opts ...operationAuditOption,
) {
	if pool == nil {
		return
	}
	rec := pool.auditRecorder
	if rec == nil {
		return
	}
	record := audit.OperationRecord{
		Profile:   profile,
		Operation: operation,
		Target:    target,
		Outcome:   classifyAuditOutcome(err),
		LatencyMs: time.Since(start).Milliseconds(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&record)
		}
	}
	audit.RecordOperation(ctx, rec, record)
}

func classifyAuditOutcome(err error) audit.Outcome {
	if err == nil {
		return audit.OutcomeSuccess
	}
	var denial *policy.DenialError
	if errors.As(err, &denial) {
		return audit.OutcomeDenied
	}
	return audit.OutcomeError
}
