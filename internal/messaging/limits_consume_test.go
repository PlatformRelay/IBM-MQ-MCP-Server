package messaging_test

import (
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
)

func TestValidateConsumeWaitIntervalMsRejectsOverMax(t *testing.T) {
	err := messaging.ValidateConsumeWaitIntervalMs(messaging.MaxConsumeWaitIntervalMs + 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateConsumeWaitIntervalMsRejectsNegative(t *testing.T) {
	err := messaging.ValidateConsumeWaitIntervalMs(-1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeConsumeCountDefaults(t *testing.T) {
	if got := messaging.NormalizeConsumeCount(0); got != messaging.DefaultConsumeCount {
		t.Fatalf("got %d", got)
	}
}
