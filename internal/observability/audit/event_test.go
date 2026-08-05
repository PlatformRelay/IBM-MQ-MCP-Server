package audit_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/audit"
)

func TestEventAttrsExcludeForbiddenKeys(t *testing.T) {
	t.Parallel()

	event := audit.Event{
		Kind:          audit.KindOperation,
		CorrelationID: "corr-1",
		Profile:       "lab",
		Operation:     "browse_queue_messages",
		Capability:    "browse",
		Target:        audit.Target{Kind: "queue", Name: "DEV.ORDERS"},
		Outcome:       audit.OutcomeSuccess,
		LatencyMs:     12,
	}

	attrs := event.Attrs()
	raw := attrsToJSON(t, attrs)
	forbidden := []string{
		"password", "passwd", "token", "secret", "credential", "credentials",
		"authorization", "bearer", "payload", "body", "messageBody", "api_key", "apikey",
	}
	lower := strings.ToLower(raw)
	for _, key := range forbidden {
		if strings.Contains(lower, `"`+key+`"`) {
			t.Fatalf("forbidden audit key %q present in output: %s", key, raw)
		}
	}
	if strings.Contains(raw, "super-secret-payload") {
		t.Fatalf("unexpected payload leakage: %s", raw)
	}
}

func TestPolicyDecisionEventFields(t *testing.T) {
	t.Parallel()

	granted := true
	event := audit.Event{
		Kind:          audit.KindPolicyDecision,
		CorrelationID: "corr-2",
		Profile:       "prod",
		Operation:     "put_queue_message",
		Capability:    "produce",
		PolicyGranted: &granted,
		Outcome:       audit.OutcomeSuccess,
	}
	raw := attrsToJSON(t, event.Attrs())
	if !strings.Contains(raw, `"kind":"policy_decision"`) {
		t.Fatalf("missing policy kind: %s", raw)
	}
	if !strings.Contains(raw, `"correlationId":"corr-2"`) {
		t.Fatalf("missing correlation id: %s", raw)
	}
	if !strings.Contains(raw, `"policyGranted":true`) {
		t.Fatalf("missing policyGranted: %s", raw)
	}
}

func attrsToJSON(t *testing.T, attrs []audit.Attr) string {
	t.Helper()
	m := make(map[string]any, len(attrs))
	for _, a := range attrs {
		m[a.Key] = a.Value
	}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
