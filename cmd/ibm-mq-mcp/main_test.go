package main

import (
	"os"
	"testing"
)

func TestResolveOpsAddrPrefersFlag(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_OPS_ADDR", ":8080")
	if got := resolveOpsAddr(":9090"); got != ":9090" {
		t.Fatalf("resolveOpsAddr = %q, want :9090", got)
	}
}

func TestResolveOpsAddrFromEnv(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_OPS_ADDR", ":8080")
	if got := resolveOpsAddr(""); got != ":8080" {
		t.Fatalf("resolveOpsAddr = %q, want :8080", got)
	}
}

func TestResolveOpsAddrEmptyInStdioOnlyMode(t *testing.T) {
	_ = os.Unsetenv("IBM_MQ_MCP_OPS_ADDR")
	if got := resolveOpsAddr(""); got != "" {
		t.Fatalf("resolveOpsAddr = %q, want empty", got)
	}
}
