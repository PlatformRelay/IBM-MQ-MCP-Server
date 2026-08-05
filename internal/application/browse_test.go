package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const browseProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_BROWSE_SECRET
    capabilities:
      - browse
`

const inspectOnlyBrowseDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_BROWSE_SECRET
    capabilities:
      - inspect
`

func TestBrowserDeniedBeforeMessagingClient(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_BROWSE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	cat, err := catalog.LoadYAML([]byte(inspectOnlyBrowseDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	browser := application.NewBrowser(pool)

	_, err = browser.BrowseQueueMessages(context.Background(), "prod", "Q1", messaging.BrowseRequest{Count: 5})
	if err == nil {
		t.Fatal("expected policy denial")
	}
	var denial *policy.DenialError
	if !errors.As(err, &denial) {
		t.Fatalf("expected DenialError, got %T: %v", err, err)
	}
	if fakeMsg.TotalCalls() != 0 {
		t.Fatalf("messaging client invoked on deny, calls=%d", fakeMsg.TotalCalls())
	}
}

func TestBrowserUsesBrowseCapability(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_BROWSE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	fakeMsg.BrowsePage.Items = []messaging.MessageRecord{{MessageID: "ID:1"}}
	cat, err := catalog.LoadYAML([]byte(browseProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	browser := application.NewBrowser(pool)

	page, err := browser.BrowseQueueMessages(context.Background(), "prod", "Q1", messaging.BrowseRequest{
		Count: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if fakeMsg.ConsumeOnlyCalls() != 0 {
		t.Fatal("destructive consume path invoked")
	}
	if fakeMsg.BrowseOnlyCalls() != 1 {
		t.Fatalf("browse calls = %d", fakeMsg.BrowseOnlyCalls())
	}
}

func TestBrowserRejectsCountOverMax(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_BROWSE_SECRET", "user:pass")
	pool := newBrowsePool(t, mustLoadBrowseCatalog(t), msgfake.New("prod"))
	browser := application.NewBrowser(pool)

	_, err := browser.BrowseQueueMessages(context.Background(), "prod", "Q1", messaging.BrowseRequest{
		Count: messaging.MaxBrowseCount + 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func newBrowsePool(t *testing.T, cat *catalog.Catalog, fakeMsg *msgfake.Client) *application.ProfilePool {
	t.Helper()
	factory := func(profile catalog.Profile, _ *secrets.Resolver) (messaging.Client, error) {
		if fakeMsg != nil {
			fakeMsg.Name = profile.Name
			return fakeMsg, nil
		}
		return msgfake.New(profile.Name), nil
	}
	pool := application.NewProfilePool(
		cat,
		cat.Validate(),
		secrets.NewResolver(),
		nil,
		application.WithMessagingFactory(factory),
	)
	t.Cleanup(func() { _ = pool.Close() })
	return pool
}

func mustLoadBrowseCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.LoadYAML([]byte(browseProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	return cat
}
