package application

import (
	"sync"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

// PolicyGate enforces deny-by-default capability checks before downstream I/O.
type PolicyGate struct {
	recorder  Recorder
	decisions []policy.Decision
	mu        sync.Mutex
}

// PolicyGateOption configures a PolicyGate.
type PolicyGateOption func(*PolicyGate)

// WithRecorder attaches observability hooks for policy denials.
func WithRecorder(recorder Recorder) PolicyGateOption {
	return func(g *PolicyGate) {
		g.recorder = recorder
	}
}

// NewPolicyGate constructs a gate that emits in-memory decision events.
func NewPolicyGate(opts ...PolicyGateOption) *PolicyGate {
	g := &PolicyGate{}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// Authorize checks profile grants for required before secret resolution or MQ I/O.
func (g *PolicyGate) Authorize(profile catalog.Profile, required policy.Capability, operation string) error {
	err := policy.Authorize(profile, required)
	granted := err == nil
	g.record(policy.Decision{
		Profile:   profile.Name,
		Required:  required,
		Granted:   granted,
		Operation: operation,
	})
	return err
}

// Decisions returns a copy of recorded policy decisions (tests and SEC-002 hooks).
func (g *PolicyGate) Decisions() []policy.Decision {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]policy.Decision, len(g.decisions))
	copy(out, g.decisions)
	return out
}

func (g *PolicyGate) record(decision policy.Decision) {
	g.mu.Lock()
	g.decisions = append(g.decisions, decision)
	g.mu.Unlock()
	if !decision.Granted && g.recorder != nil {
		g.recorder.RecordPolicyDenial(decision.Profile)
	}
}
