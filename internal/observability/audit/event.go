// Package audit records payload-safe structured audit events for sensitive MCP operations.
package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
)

// Kind distinguishes policy decisions from completed operation outcomes.
type Kind string

const (
	// KindPolicyDecision marks an authorization outcome event from POL-001.
	KindPolicyDecision Kind = "policy_decision"
	// KindOperation marks a completed sensitive MCP/MQ operation.
	KindOperation Kind = "operation"
)

// Outcome reports whether an operation succeeded, was denied, or failed.
type Outcome string

const (
	// OutcomeSuccess marks a successful operation or grant.
	OutcomeSuccess Outcome = "success"
	// OutcomeDenied marks a policy denial before MQ I/O.
	OutcomeDenied Outcome = "denied"
	// OutcomeError marks a non-denial execution failure.
	OutcomeError Outcome = "error"
)

// Target names the object an operation acted on without carrying payloads.
type Target struct {
	Kind string
	Name string
}

// Event is the v0 audit schema. Only fields mapped in Attrs may reach the sink.
type Event struct {
	Kind            Kind
	CorrelationID   string
	Profile         string
	Operation       string
	Capability      string
	PolicyGranted   *bool
	Target          Target
	Outcome         Outcome
	LatencyMs       int64
	ClientSession   string
	CommandRedacted string
}

// Attr is a key/value pair safe for structured audit output.
type Attr struct {
	Key   string
	Value any
}

// Attrs returns the allowlisted audit fields for structured logging.
func (e Event) Attrs() []Attr {
	attrs := []Attr{
		{Key: "audit", Value: true},
		{Key: "kind", Value: string(e.Kind)},
		{Key: "profile", Value: e.Profile},
		{Key: "operation", Value: e.Operation},
		{Key: "outcome", Value: string(e.Outcome)},
	}
	if id := e.CorrelationID; id != "" {
		attrs = append(attrs, Attr{Key: "correlationId", Value: id})
	}
	if capability := e.Capability; capability != "" {
		attrs = append(attrs, Attr{Key: "capability", Value: capability})
	}
	if e.PolicyGranted != nil {
		attrs = append(attrs, Attr{Key: "policyGranted", Value: *e.PolicyGranted})
	}
	if e.Target.Kind != "" || e.Target.Name != "" {
		attrs = append(attrs,
			Attr{Key: "targetKind", Value: e.Target.Kind},
			Attr{Key: "targetName", Value: e.Target.Name},
		)
	}
	if e.LatencyMs > 0 {
		attrs = append(attrs, Attr{Key: "latencyMs", Value: e.LatencyMs})
	}
	if session := e.ClientSession; session != "" {
		attrs = append(attrs, Attr{Key: "clientSession", Value: session})
	}
	if cmd := e.CommandRedacted; cmd != "" {
		attrs = append(attrs, Attr{Key: "commandRedacted", Value: cmd})
	}
	return attrs
}

// OperationRecord captures one completed sensitive operation for audit emission.
type OperationRecord struct {
	Profile         string
	Operation       string
	Target          Target
	Outcome         Outcome
	LatencyMs       int64
	ClientSession   string
	CommandRedacted string
}

// RecordOperation emits an operation audit event using the correlation id from ctx.
func RecordOperation(ctx context.Context, rec Recorder, record OperationRecord) {
	if rec == nil {
		return
	}
	rec.Record(ctx, Event{
		Kind:            KindOperation,
		CorrelationID:   CorrelationIDFrom(ctx),
		ClientSession:   ClientSessionFrom(ctx),
		Profile:         record.Profile,
		Operation:       record.Operation,
		Target:          record.Target,
		Outcome:         record.Outcome,
		LatencyMs:       record.LatencyMs,
		CommandRedacted: record.CommandRedacted,
	})
}

// PolicyDecisionRecord captures one POL-001 authorization outcome.
type PolicyDecisionRecord struct {
	Profile    string
	Operation  string
	Capability string
	Granted    bool
}

// RecordPolicyDecision emits a policy audit event using the correlation id from ctx.
func RecordPolicyDecision(ctx context.Context, rec Recorder, record PolicyDecisionRecord) {
	if rec == nil {
		return
	}
	granted := record.Granted
	outcome := OutcomeSuccess
	if !granted {
		outcome = OutcomeDenied
	}
	rec.Record(ctx, Event{
		Kind:          KindPolicyDecision,
		CorrelationID: CorrelationIDFrom(ctx),
		ClientSession: ClientSessionFrom(ctx),
		Profile:       record.Profile,
		Operation:     record.Operation,
		Capability:    record.Capability,
		PolicyGranted: &granted,
		Outcome:       outcome,
	})
}

// Recorder persists payload-safe audit events. Implementations must fail open.
type Recorder interface {
	Record(ctx context.Context, event Event)
}

// SlogRecorder writes audit events as structured slog JSON (v0 sink).
type SlogRecorder struct {
	logger *slog.Logger
}

// NewSlogRecorder constructs the default audit sink backed by slog.
func NewSlogRecorder(logger *slog.Logger) *SlogRecorder {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogRecorder{logger: logger}
}

// Record emits one audit event. Sink failures are ignored (fail-open).
func (r *SlogRecorder) Record(ctx context.Context, event Event) {
	if r == nil || r.logger == nil {
		return
	}
	if event.CorrelationID == "" {
		event.CorrelationID = CorrelationIDFrom(ctx)
	}
	if event.ClientSession == "" {
		event.ClientSession = ClientSessionFrom(ctx)
	}
	attrs := make([]slog.Attr, 0, len(event.Attrs()))
	for _, a := range event.Attrs() {
		attrs = append(attrs, slog.Any(a.Key, a.Value))
	}
	r.logger.LogAttrs(ctx, slog.LevelInfo, "audit", attrs...)
}

// NewCorrelationID returns a random lowercase hex identifier.
func NewCorrelationID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "audit-unavailable"
	}
	return hex.EncodeToString(b[:])
}
