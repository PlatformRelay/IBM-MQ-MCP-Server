package application_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter/mqweb"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/mqadmin/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const inspectTestDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_INSPECT_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
  readonly:
    queueManager: QM2
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_INSPECT_SECRET
    capabilities:
      - browse
`

const browseOnlyInspectDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_INSPECT_SECRET
    capabilities:
      - browse
`

func newInspectPool(t *testing.T, fakeClient *fake.Client) *application.Inspector {
	t.Helper()
	t.Setenv("IBM_MQ_MCP_INSPECT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(inspectTestDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(profile catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		if fakeClient != nil {
			fakeClient.Name = profile.Name
			return fakeClient, nil
		}
		return fake.New(profile.Name), nil
	}
	return application.NewInspector(newPool(t, cat, nil, factory))
}

func TestInspectorListProfilesNoSecrets(t *testing.T) {
	inspector := newInspectPool(t, nil)
	page, err := inspector.ListProfilesPage(50, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d", len(page.Items))
	}
	for _, item := range page.Items {
		if item.Name == "" || item.QueueManager == "" || len(item.Capabilities) == 0 {
			t.Fatalf("incomplete summary: %+v", item)
		}
	}
}

func TestInspectorDeniesQueueManagerBeforeAdapter(t *testing.T) {
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_INSPECT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyInspectDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := newPool(t, cat, nil, factory)
	inspector := application.NewInspector(pool)

	_, err = inspector.QueueManagerStatus(context.Background(), "prod")
	if err == nil {
		t.Fatal("expected policy denial")
	}
	qm, list, get, ping := fakeClient.Calls()
	if qm+list+get+ping != 0 {
		t.Fatalf("adapter invoked on deny: qm=%d list=%d get=%d ping=%d", qm, list, get, ping)
	}
}

func TestInspectorListQueuesHappyPath(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.ListQueuesPage.Items = []mqadmin.QueueSummary{{Name: "DEV.QUEUE.1", Type: "local"}}
	inspector := newInspectPool(t, fakeClient)

	page, err := inspector.ListQueues(context.Background(), "prod", mqadmin.ListQueuesRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "DEV.QUEUE.1" {
		t.Fatalf("page = %+v", page)
	}
	if _, list, _, _ := fakeClient.Calls(); list != 1 {
		t.Fatalf("expected adapter list call")
	}
}

func TestInspectorRequiresInspectCapability(t *testing.T) {
	fakeClient := fake.New("readonly")
	inspector := newInspectPool(t, fakeClient)

	_, err := inspector.GetQueue(context.Background(), "readonly", "Q1")
	if err == nil {
		t.Fatal("expected denial")
	}
	var denial *policy.DenialError
	if !errors.As(err, &denial) {
		t.Fatalf("expected DenialError, got %v", err)
	}
}

func TestInspectorDeniesListChannelsBeforeAdapter(t *testing.T) {
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_INSPECT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyInspectDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := newPool(t, cat, nil, factory)
	inspector := application.NewInspector(pool)

	_, err = inspector.ListChannels(context.Background(), "prod", mqadmin.ListChannelsRequest{Limit: 10})
	if err == nil {
		t.Fatal("expected policy denial")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny: total=%d", fakeClient.TotalCalls())
	}
}

func TestInspectorListChannelsHappyPath(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.ListChannelsPage.Items = []mqadmin.ChannelSummary{{Name: "DEV.SVRCONN", Type: "serverConnection"}}
	inspector := newInspectPool(t, fakeClient)

	page, err := inspector.ListChannels(context.Background(), "prod", mqadmin.ListChannelsRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "DEV.SVRCONN" {
		t.Fatalf("page = %+v", page)
	}
	if fakeClient.ListChannelsCalls != 1 {
		t.Fatalf("expected adapter list call")
	}
}

func TestInspectorGetChannelStatusUnavailableWithoutRuntime(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.ChannelStatus = mqadmin.ChannelStatus{
		Name:         "DEV.SVRCONN",
		Availability: mqadmin.Unavailable,
		Error:        "runtime status not returned by mqweb",
	}
	inspector := newInspectPool(t, fakeClient)

	status, err := inspector.GetChannelStatus(context.Background(), "prod", "DEV.SVRCONN")
	if err != nil {
		t.Fatal(err)
	}
	if status.Availability != mqadmin.Unavailable {
		t.Fatalf("availability = %q", status.Availability)
	}
}

func TestInspectorListListenersUnsupported(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.ListListenersErr = mqadmin.UnsupportedFamily("listener")
	inspector := newInspectPool(t, fakeClient)

	_, err := inspector.ListListeners(context.Background(), "prod", mqadmin.ListListenersRequest{Limit: 10})
	if err == nil {
		t.Fatal("expected unsupported error")
	}
	if _, ok := mqadmin.AsUnsupportedError(err); !ok {
		t.Fatalf("expected UnsupportedError, got %v", err)
	}
}

func TestInspectorCheckProfileConnectivityHappyPath(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.QMStatus = mqadmin.QueueManagerStatus{
		Profile:      "prod",
		Availability: mqadmin.Available,
		Identity: mqadmin.Identity{
			Configured: "QM1",
			Observed:   "QM1",
		},
		LastChecked: time.Now().UTC(),
	}
	inspector := newInspectPool(t, fakeClient)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Reachable || !report.IdentityMatch {
		t.Fatalf("report = %+v", report)
	}
	if report.LatencyMs < 0 {
		t.Fatalf("latency = %d", report.LatencyMs)
	}
	if fakeClient.QMStatusCalls != 1 {
		t.Fatalf("qm status calls = %d", fakeClient.QMStatusCalls)
	}
}

func TestInspectorCheckProfileConnectivityDeniedBeforeAdapter(t *testing.T) {
	fakeClient := fake.New("prod")
	t.Setenv("IBM_MQ_MCP_INSPECT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(browseOnlyInspectDoc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := newPool(t, cat, nil, factory)
	inspector := application.NewInspector(pool)

	_, err = inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err == nil {
		t.Fatal("expected policy denial")
	}
	if fakeClient.TotalCalls() != 0 {
		t.Fatalf("adapter invoked on deny: total=%d", fakeClient.TotalCalls())
	}
}

func TestInspectorCheckProfileConnectivityUnreachableDNS(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.QMStatusErr = &net.DNSError{IsNotFound: true, Name: "mq.example.test"}
	inspector := newInspectPool(t, fakeClient)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if report.Reachable {
		t.Fatal("expected unreachable report")
	}
	if report.FailureCause != mqadmin.FailureDNS {
		t.Fatalf("cause = %q", report.FailureCause)
	}
}

func TestInspectorCheckProfileConnectivityAuthorization(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.QMStatusErr = mqadmin.MapReasonCode(2035)
	inspector := newInspectPool(t, fakeClient)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if report.FailureCause != mqadmin.FailureAuthorization {
		t.Fatalf("cause = %q detail=%q", report.FailureCause, report.Detail)
	}
}

func TestInspectorCheckProfileConnectivityStaleIdentity(t *testing.T) {
	fakeClient := fake.New("prod")
	fakeClient.QMStatus = mqadmin.QueueManagerStatus{
		Profile:      "prod",
		Availability: mqadmin.Stale,
		Identity: mqadmin.Identity{
			Configured: "QM1",
			Observed:   "QM2",
		},
	}
	inspector := newInspectPool(t, fakeClient)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Reachable {
		t.Fatalf("stale should still be reachable: %+v", report)
	}
	if report.IdentityMatch {
		t.Fatal("expected identity mismatch")
	}
}

func TestInspectorCheckProfileConnectivityRedactsEndpoint(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mquser:mqsecret@mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_INSPECT_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`
	t.Setenv("IBM_MQ_MCP_INSPECT_SECRET", "user:pass")
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := fake.New("prod")
	fakeClient.QMStatus = mqadmin.QueueManagerStatus{
		Profile:      "prod",
		Availability: mqadmin.Available,
		Identity:     mqadmin.Identity{Configured: "QM1", Observed: "QM1"},
	}
	factory := func(_ catalog.Profile, _ *secrets.Resolver) (mqadmin.Client, error) {
		return fakeClient, nil
	}
	pool := newPool(t, cat, nil, factory)
	inspector := application.NewInspector(pool)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(report.Endpoint, "mqsecret") || strings.Contains(report.Endpoint, "mquser") {
		t.Fatalf("endpoint leaked credentials: %q", report.Endpoint)
	}
}

func TestInspectorCheckProfileConnectivitySecretFailure(t *testing.T) {
	doc := `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_MISSING_SECRET
    tls:
      insecureSkipVerify: true
    capabilities:
      - inspect
`
	cat, err := catalog.LoadYAML([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	factory := func(profile catalog.Profile, resolver *secrets.Resolver) (mqadmin.Client, error) {
		return mqweb.NewAdminClient(profile, resolver)
	}
	pool := newPool(t, cat, nil, factory)
	inspector := application.NewInspector(pool)

	report, err := inspector.CheckProfileConnectivity(context.Background(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if report.FailureCause != mqadmin.FailureAuthentication {
		t.Fatalf("cause = %q detail=%q", report.FailureCause, report.Detail)
	}
	if strings.Contains(report.Detail, "user:") || strings.Contains(report.Detail, ":pass") {
		t.Fatalf("detail should not contain credential material: %q", report.Detail)
	}
}
