package metrics_test

import (
	"strings"
	"testing"

	"github.com/prometheus/common/expfmt"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/observability/metrics"
)

func TestProfileLabelUsesNoneWhenEmpty(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.RecordRequest("", 0)
	reg.RecordPolicyDenial("")

	out := gatherMetrics(t, reg)
	if !strings.Contains(out, `profile="_none"`) {
		t.Fatalf("expected _none profile label, got:\n%s", out)
	}
}

func TestMetricsHaveNoSecretLabels(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.RecordRequest("dev", 0.01)
	reg.RecordPolicyDenial("dev")

	out := gatherMetrics(t, reg)
	forbidden := []string{
		`password=`,
		`token=`,
		`host=`,
		`queue_manager=`,
		`connection=`,
	}
	for _, label := range forbidden {
		if strings.Contains(out, label) {
			t.Fatalf("metrics expose forbidden label %q:\n%s", label, out)
		}
	}
}

func TestRecordRequestIncrementsCounterAndObservesLatency(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.RecordRequest("lab", 0.05)
	reg.RecordRequest("lab", 0.10)

	out := gatherMetrics(t, reg)
	if !strings.Contains(out, "ibm_mq_mcp_requests_total") {
		t.Fatalf("missing requests counter:\n%s", out)
	}
	if !strings.Contains(out, "ibm_mq_mcp_request_duration_seconds") {
		t.Fatalf("missing latency histogram:\n%s", out)
	}
	if !strings.Contains(out, `profile="lab"`) {
		t.Fatalf("missing lab profile label:\n%s", out)
	}
}

func TestPolicyDenialCounter(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.RecordPolicyDenial("prod")

	out := gatherMetrics(t, reg)
	if !strings.Contains(out, "ibm_mq_mcp_policy_denials_total") {
		t.Fatalf("missing policy denial counter:\n%s", out)
	}
	if !strings.Contains(out, `profile="prod"`) {
		t.Fatalf("missing prod profile label:\n%s", out)
	}
}

func gatherMetrics(t *testing.T, reg *metrics.Registry) string {
	t.Helper()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	var buf strings.Builder
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	return buf.String()
}

func TestGatherReturnsMetricFamiliesAfterUse(t *testing.T) {
	t.Parallel()

	reg := metrics.New()
	reg.RecordRequest("_none", 0)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(mfs) == 0 {
		t.Fatal("expected metric families")
	}
}
