package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/collection"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const consumeProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_CONSUME_SECRET
    capabilities:
      - consume
`

const browseOnlyConsumeDenyDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_CONSUME_SECRET
    capabilities:
      - browse
`

func TestConsumerDeniedBeforeMessagingClient(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONSUME_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	cat, err := catalog.LoadYAML([]byte(browseOnlyConsumeDenyDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	consumer := application.NewConsumer(pool)

	_, err = consumer.ConsumeQueueMessages(context.Background(), "prod", "Q1", messaging.ConsumeRequest{Count: 5})
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

func TestConsumerUsesConsumeCapabilityNotBrowse(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONSUME_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	fakeMsg.ConsumePage.Items = []messaging.MessageRecord{{MessageID: "ID:1"}}
	cat, err := catalog.LoadYAML([]byte(consumeProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	consumer := application.NewConsumer(pool)

	page, err := consumer.ConsumeQueueMessages(context.Background(), "prod", "Q1", messaging.ConsumeRequest{
		Count: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if fakeMsg.BrowseOnlyCalls() != 0 {
		t.Fatal("browse path invoked during consume")
	}
	if fakeMsg.ConsumeOnlyCalls() != 1 {
		t.Fatalf("consume calls = %d", fakeMsg.ConsumeOnlyCalls())
	}
}

func TestBrowseAloneNeverInvokesConsume(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_BROWSE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	fakeMsg.BrowsePage.Items = []messaging.MessageRecord{{MessageID: "ID:1"}}
	cat, err := catalog.LoadYAML([]byte(browseProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	browser := application.NewBrowser(pool)

	_, err = browser.BrowseQueueMessages(context.Background(), "prod", "Q1", messaging.BrowseRequest{Count: 5})
	if err != nil {
		t.Fatal(err)
	}
	if fakeMsg.ConsumeOnlyCalls() != 0 {
		t.Fatalf("browse invoked consume path, calls=%d", fakeMsg.ConsumeOnlyCalls())
	}
}

func TestConsumerRejectsCountOverMax(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONSUME_SECRET", "user:pass")
	pool := newBrowsePool(t, mustLoadConsumeCatalog(t), msgfake.New("prod"))
	consumer := application.NewConsumer(pool)

	_, err := consumer.ConsumeQueueMessages(context.Background(), "prod", "Q1", messaging.ConsumeRequest{
		Count: messaging.MaxConsumeCount + 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConsumerRejectsWaitIntervalOverMax(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONSUME_SECRET", "user:pass")
	pool := newBrowsePool(t, mustLoadConsumeCatalog(t), msgfake.New("prod"))
	consumer := application.NewConsumer(pool)

	_, err := consumer.ConsumeQueueMessages(context.Background(), "prod", "Q1", messaging.ConsumeRequest{
		WaitIntervalMs: messaging.MaxConsumeWaitIntervalMs + 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestConsumerPreservesPartialResultsOnMidBatchFailure(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_CONSUME_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	partialPage := collection.Page[messaging.MessageRecord]{
		Items:            []messaging.MessageRecord{{MessageID: "ID:1"}},
		Truncated:        true,
		TruncationReason: collection.TruncationMidBatchFailure,
	}
	fakeMsg.ConsumePage = partialPage
	fakeMsg.ConsumeErr = messaging.NewPartialConsumeError(partialPage, errors.New("status 500"))
	cat, err := catalog.LoadYAML([]byte(consumeProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newBrowsePool(t, cat, fakeMsg)
	consumer := application.NewConsumer(pool)

	page, err := consumer.ConsumeQueueMessages(context.Background(), "prod", "Q1", messaging.ConsumeRequest{
		Count: 3,
	})
	if err == nil {
		t.Fatal("expected partial consume error")
	}
	var partial *messaging.PartialConsumeError
	if !errors.As(err, &partial) {
		t.Fatalf("expected PartialConsumeError, got %T", err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d", len(page.Items))
	}
	if page.Items[0].MessageID != "ID:1" {
		t.Fatalf("messageId = %q", page.Items[0].MessageID)
	}
}

func mustLoadConsumeCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.LoadYAML([]byte(consumeProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	return cat
}
