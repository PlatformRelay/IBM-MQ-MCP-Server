package output_test

import (
	"strings"
	"testing"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/output"
)

func TestRenderProfilePageCompact(t *testing.T) {
	page := collection.Page[application.ProfileSummary]{
		Items: []application.ProfileSummary{
			{
				Name: "prod", QueueManager: "QM1", Endpoint: "https://mq.example:9443",
				Capabilities: []string{"inspect"}, Valid: true,
			},
		},
		Limit: 50,
	}
	got := output.RenderProfilePage(page)
	if strings.Contains(got, "{") {
		t.Fatalf("text fallback must not echo JSON: %q", got)
	}
	if !strings.Contains(got, "prod") || !strings.Contains(got, "QM1") {
		t.Fatalf("missing profile fields: %q", got)
	}
	if strings.Count(got, "prod") != 1 {
		t.Fatalf("redundant prose: %q", got)
	}
}

func TestRenderQueuePageDeterministic(t *testing.T) {
	page := collection.Page[mqadmin.QueueSummary]{
		Items: []mqadmin.QueueSummary{
			{Name: "DEV.QUEUE.1", Type: "local"},
			{Name: "DEV.QUEUE.2", Type: "local"},
		},
		Limit:            50,
		Truncated:        true,
		TruncationReason: collection.TruncationLimitReached,
		NextCursor:       "DEV.QUEUE.2",
	}
	a := output.RenderQueuePage(page)
	b := output.RenderQueuePage(page)
	if a != b {
		t.Fatalf("render not deterministic:\nA=%q\nB=%q", a, b)
	}
	if !strings.Contains(a, "truncated=limit_reached") {
		t.Fatalf("missing truncation: %q", a)
	}
}

func TestRenderMessagePageOmitsPayloadByDefault(t *testing.T) {
	page := collection.Page[messaging.MessageRecord]{
		Items: []messaging.MessageRecord{
			{MessageID: "ID:1", MessageLength: 42, Encoding: messaging.EncodingOmitted},
		},
		Limit: 10,
	}
	got := output.RenderMessagePage(page)
	if strings.Contains(got, "payload=") {
		t.Fatalf("must not surface payload in compact text: %q", got)
	}
	if !strings.Contains(got, "ID:1") {
		t.Fatalf("missing message id: %q", got)
	}
}

func TestRenderReasonExplanationKnown(t *testing.T) {
	got := output.RenderReasonExplanation(mqadmin.ExplainReasonCode(2035))
	if strings.Contains(got, "{") {
		t.Fatalf("must not echo JSON: %q", got)
	}
	if !strings.Contains(got, "2035") || !strings.Contains(got, "MQRC_NOT_AUTHORIZED") {
		t.Fatalf("missing reason fields: %q", got)
	}
}

func TestRenderConnectivityReportReachable(t *testing.T) {
	got := output.RenderConnectivityReport(mqadmin.ConnectivityReport{
		Profile:       "prod",
		Endpoint:      "https://mq.example:9443",
		Reachable:     true,
		IdentityMatch: true,
		LatencyMs:     12,
		Identity:      mqadmin.Identity{Configured: "QM1", Observed: "QM1"},
		CheckedAt:     time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
	})
	if strings.Contains(got, "{") {
		t.Fatalf("must not echo JSON: %q", got)
	}
	if !strings.Contains(got, "reachable=true") {
		t.Fatalf("missing reachability: %q", got)
	}
}
