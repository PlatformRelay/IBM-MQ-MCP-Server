package runtime_test

import (
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/runtime"
)

func TestReadyRequiresConfigAndTransport(t *testing.T) {
	t.Parallel()

	rt := runtime.New()

	if rt.Ready() {
		t.Fatal("expected not ready before config and transport are set")
	}

	rt.SetConfigValid(true)
	if rt.Ready() {
		t.Fatal("expected not ready with config only")
	}

	rt.SetTransportReady(true, "stdio")
	if !rt.Ready() {
		t.Fatal("expected ready when config valid and transport serving")
	}

	rt.SetConfigValid(false)
	if rt.Ready() {
		t.Fatal("expected not ready when config invalid")
	}
}

func TestHealthyWhileProcessServes(t *testing.T) {
	t.Parallel()

	rt := runtime.New()
	if !rt.Healthy() {
		t.Fatal("expected healthy liveness while process runs")
	}
}

func TestTransportStateSnapshot(t *testing.T) {
	t.Parallel()

	rt := runtime.New()
	rt.SetTransportReady(true, "stdio")

	name, ready := rt.TransportState()
	if !ready || name != "stdio" {
		t.Fatalf("TransportState = (%q, %v), want (stdio, true)", name, ready)
	}
}
