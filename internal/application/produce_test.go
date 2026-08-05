package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/ibm-mq-mcp-server/internal/application"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/catalog"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/config/secrets"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/messaging"
	msgfake "github.com/platformrelay/ibm-mq-mcp-server/internal/messaging/fake"
	"github.com/platformrelay/ibm-mq-mcp-server/internal/policy"
)

const produceProfileDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_PRODUCE_SECRET
    capabilities:
      - produce
`

const browseOnlyProduceDoc = `
profiles:
  prod:
    queueManager: QM1
    endpoint: https://mq.example.test:9443
    authentication:
      type: basic
      secretRef: env:IBM_MQ_MCP_PRODUCE_SECRET
    capabilities:
      - browse
`

func TestProducerDeniedBeforeMessagingClient(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_PRODUCE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	cat, err := catalog.LoadYAML([]byte(browseOnlyProduceDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newProducePool(t, cat, fakeMsg)
	producer := application.NewProducer(pool)

	_, err = producer.PutQueueMessage(context.Background(), "prod", "Q1", messaging.PutRequest{
		ContentType: messaging.ContentTypeTextPlain,
		Payload:     "hello",
	})
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

func TestProducerUsesProduceCapability(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_PRODUCE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	fakeMsg.PutResult = messaging.PutResult{MessageID: "ID:abc", Format: "MQSTR"}
	cat, err := catalog.LoadYAML([]byte(produceProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	pool := newProducePool(t, cat, fakeMsg)
	producer := application.NewProducer(pool)

	result, err := producer.PutQueueMessage(context.Background(), "prod", "Q1", messaging.PutRequest{
		ContentType: messaging.ContentTypeTextPlain,
		Payload:     "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.MessageID != "ID:abc" {
		t.Fatalf("messageId = %q", result.MessageID)
	}
	if fakeMsg.PutCalls != 1 {
		t.Fatalf("put calls = %d", fakeMsg.PutCalls)
	}
}

func TestProducerRejectsInvalidContentTypeBeforeMQ(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_PRODUCE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	pool := newProducePool(t, mustLoadProduceCatalog(t), fakeMsg)
	producer := application.NewProducer(pool)

	_, err := producer.PutQueueMessage(context.Background(), "prod", "Q1", messaging.PutRequest{
		ContentType: "text/html",
		Payload:     "<p>x</p>",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ctErr *messaging.ContentTypeError
	if !errors.As(err, &ctErr) {
		t.Fatalf("expected ContentTypeError, got %T", err)
	}
	if fakeMsg.TotalCalls() != 0 {
		t.Fatalf("messaging invoked on validation error, calls=%d", fakeMsg.TotalCalls())
	}
}

func TestProducerRejectsOversizePayloadBeforeMQ(t *testing.T) {
	t.Setenv("IBM_MQ_MCP_PRODUCE_SECRET", "user:pass")
	fakeMsg := msgfake.New("prod")
	pool := newProducePool(t, mustLoadProduceCatalog(t), fakeMsg)
	producer := application.NewProducer(pool)

	_, err := producer.PutQueueMessage(context.Background(), "prod", "Q1", messaging.PutRequest{
		ContentType: messaging.ContentTypeTextPlain,
		Payload:     strings.Repeat("x", messaging.HardMaxPayloadBytes+1),
	})
	if err == nil {
		t.Fatal("expected size error")
	}
	var sizeErr *messaging.PayloadSizeError
	if !errors.As(err, &sizeErr) {
		t.Fatalf("expected PayloadSizeError, got %T", err)
	}
	if fakeMsg.TotalCalls() != 0 {
		t.Fatalf("messaging invoked on size error, calls=%d", fakeMsg.TotalCalls())
	}
}

func newProducePool(t *testing.T, cat *catalog.Catalog, fakeMsg *msgfake.Client) *application.ProfilePool {
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

func mustLoadProduceCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.LoadYAML([]byte(produceProfileDoc))
	if err != nil {
		t.Fatal(err)
	}
	return cat
}
