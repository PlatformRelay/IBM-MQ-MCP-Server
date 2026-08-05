// Package runtime tracks process-level health and readiness without MQ probes.
package runtime

import "sync"

// Runtime holds configuration validity and MCP transport state for probes.
type Runtime struct {
	mu sync.RWMutex

	configValid    bool
	transportReady bool
	transportName  string
}

// New returns a runtime with conservative defaults (not ready until configured).
func New() *Runtime {
	return &Runtime{}
}

// SetConfigValid records whether static configuration parsed and validated.
func (r *Runtime) SetConfigValid(valid bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configValid = valid
}

// SetTransportReady records whether the MCP transport is serving.
func (r *Runtime) SetTransportReady(ready bool, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transportReady = ready
	if ready {
		r.transportName = name
	} else {
		r.transportName = ""
	}
}

// Healthy reports liveness. The process is healthy while it can serve probes.
func (r *Runtime) Healthy() bool {
	return true
}

// Ready reports whether the server may accept MCP work (config + transport).
func (r *Runtime) Ready() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.configValid && r.transportReady
}

// TransportState returns the active transport name and readiness flag.
func (r *Runtime) TransportState() (name string, ready bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transportName, r.transportReady
}
